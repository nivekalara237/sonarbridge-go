package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sonarbridge-go/internal/infra/entrypoints/dto"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"
	"sonarbridge-go/internal/plugin/lifecycle"
	"sonarbridge-go/internal/plugin/manager"
	"sort"
	"strings"
	"sync"
	"time"
)

// PluginLifecycleController exposes the plugin lifecycle to the Dashboard:
//
//	POST /api/plugins/{name}/restart ----> stop (if running) then start again
//	POST /api/plugins/{name}/stop -------> stop a running plugin
//	POST /api/plugins/{name}/enable -----> enable a disabled plugin (does not start it)
//	POST /api/plugins/{name}/transition -> {"to": "READY"|"STOPPED"|"ENABLED"|"DISABLED"}
//
// The controller decides nothing about lifecycle itself: every request is
// translated into one Manager operation, and the lifecycle FSM stays the
// single judge of what is legal. Its own job is the HTTP contract —
// status codes, idempotence, and not letting a dropped connection or a
// double click leave a plugin half-started.
//
// These endpoints start and stop processes: mount them behind the same
// authentication as the rest of the admin API, never on the unauthenticated
// telemetry listener.
type PluginLifecycleController struct {
	OpTimeout time.Duration
	mgr       PluginLifecycle
	mu        sync.Mutex
	inFlight  map[string]struct{}
	Logger    *slog.Logger
}

type PluginLifecycle interface {
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Restart(ctx context.Context, name string) error
	Enable(name string) error
	Disable(ctx context.Context, name string) error
	List() []manager.Status
}

var _ PluginLifecycle = (*manager.Manager)(nil)

func NewPluginLifecycleController(mgr PluginLifecycle) *PluginLifecycleController {
	return &PluginLifecycleController{mgr: mgr, inFlight: make(map[string]struct{})}
}

func (c *PluginLifecycleController) SelfRegister(mux *http.ServeMux) {
	mux.Handle("POST /plugins/{name}/restart", httpx.Handlerx(c.RestartHandler))
	mux.Handle("POST /plugins/{name}/stop", httpx.Handlerx(c.StopHandler))
	mux.Handle("POST /plugins/{name}/enable", httpx.Handlerx(c.EnableHandler))
	mux.Handle("POST /plugins/{name}/transition", httpx.Handlerx(c.TransitionHandler))
}

const (
	defaultPluginOpTimeout = 15 * time.Second
	maxPluginNameLenght    = 128
	maxTransitionBodyBytes = 1 << 10 // 1Mo = 1024Ko
)

type pluginOp string

const (
	PluginOpRestart pluginOp = "restart"
	PluginOpStop    pluginOp = "stop"
	PluginOpStart   pluginOp = "start"
	PluginOpEnable  pluginOp = "enable"
	PluginOpDisable pluginOp = "disable"
)

var pluginTransitionTargets = map[lifecycle.State]pluginOp{
	lifecycle.StateReady:    PluginOpStart,
	lifecycle.StateStopped:  PluginOpStop,
	lifecycle.StateEnabled:  PluginOpEnable,
	lifecycle.StateDisabled: PluginOpDisable,
}

// pluginAlreadyEnabled lists the states that imply the plugin is enabled:
// asking to enable it again is a no-op, not an error.
var pluginAlreadyEnabled = map[lifecycle.State]bool{
	lifecycle.StateEnabled: true, lifecycle.StateStarting: true, lifecycle.StateHandshaking: true,
	lifecycle.StateReady: true, lifecycle.StateCrashed: true, lifecycle.StateStopping: true,
	lifecycle.StateStopped: true, lifecycle.StateFailed: true,
}

// pluginVerb is how an operation reads in an error message.
var pluginVerb = map[pluginOp]string{
	PluginOpRestart: "be restarted",
	PluginOpStop:    "be stopped",
	PluginOpStart:   "be started",
	PluginOpEnable:  "be enabled",
	PluginOpDisable: "be disabled",
}

func (c *PluginLifecycleController) handle(w http.ResponseWriter, r *http.Request, op pluginOp) {
	if !httpx.RequireMethod(w, r, "POST") {
		return
	}
	c.execute(w, r, op)
}

func (c *PluginLifecycleController) execute(w http.ResponseWriter, r *http.Request, op pluginOp) {
	name := pluginOf(r)
	log := c.logger().With("plugin-lifecycle-ctrl", name, "op", string(op), "remote", r.RemoteAddr)
	if name == "" || len(name) > maxPluginNameLenght {
		pluginWriteError(w, http.StatusBadRequest, "plugin_not_valid", "plugin invalid", "")
		return
	}

	release, ok := c.acquire(name)
	if !ok {
		pluginWriteError(w, http.StatusConflict, "operation_in_progress",
			"another operation is already running on this plugin", "")
		return
	}

	defer release()

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), c.timeout())
	defer cancel()

	before, known := c.status(name)
	if !known {
		pluginWriteError(w, http.StatusNotFound, "plugin_not_found", "plugin not found", "")
		return
	}
	changed, err := c.apply(ctx, op, before.State, name)
	after, _ := c.status(name)
	if err != nil {
		c.fail(w, log, ctx, op, err, after.State)
		return
	}

	log.Info("plugin operation applied", "from", string(before.State), "to", string(after.State), "changed", changed)
	pluginWriteJson(w, http.StatusOK, dto.PluginActionResponse{
		Name:         name,
		State:        string(after.State),
		Changed:      changed,
		Version:      after.Info.Version,
		Capabilities: append([]string(nil), after.Info.Capabilities...),
	})
}

func (c *PluginLifecycleController) apply(ctx context.Context, op pluginOp, from lifecycle.State, name string) (changed bool, err error) {
	switch op {
	case PluginOpRestart:
		return true, c.mgr.Restart(ctx, name)
	case PluginOpStop:
		if from == lifecycle.StateStopped {
			return false, nil
		}
		return true, c.mgr.Stop(ctx, name)
	case PluginOpStart:
		if from == lifecycle.StateDisabled {
			return false, nil
		}
		return true, c.mgr.Start(ctx, name)
	case PluginOpEnable:
		if pluginAlreadyEnabled[from] {
			return false, nil
		}
		return true, c.mgr.Enable(name)
	case PluginOpDisable:
		if from == lifecycle.StateDisabled {
			return false, nil
		}
		return true, c.mgr.Disable(ctx, name)
	}
	return false, fmt.Errorf("unsupported plugin operation %q", op)
}

func (c *PluginLifecycleController) fail(w http.ResponseWriter, log *slog.Logger, ctx context.Context, op pluginOp, err error, state lifecycle.State) {
	switch {
	case errors.Is(err, manager.ErrUnknownPlugin):
		log.Info("plugin operation refused: unknown plugin")
		pluginWriteError(w, http.StatusNotFound, "plugin_not_found", "plugin not found", "")
	case errors.Is(err, lifecycle.ErrIllegalTransition):
		log.Info("plugin operation refused: illegal transition", "state", string(state), "error", err)
		msg := "this operation is not allowed in the plugin's current state"
		if state != "" {
			msg = fmt.Sprintf("a plugin is state %s cannot %s", state, pluginVerb[op])
		}
		pluginWriteError(w, http.StatusConflict, "illegal_transition", msg, string(state))
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded):
		log.Error("plugin operation timed out", "state", string(state), "error", err)
		pluginWriteError(w, http.StatusGatewayTimeout, "operation_timeout", "the operation timed out", string(state))
	default:
		log.Error("plugin operation failed", "state", string(state), "error", err)
		pluginWriteError(w, http.StatusInternalServerError, "operation_failed", "the operation failed. See the server logs", string(state))
	}
}

func (c *PluginLifecycleController) status(name string) (manager.Status, bool) {
	for _, s := range c.mgr.List() {
		if s.Name == name {
			return s, true
		}
	}

	return manager.Status{}, false
}

func (c *PluginLifecycleController) acquire(name string) (release func(), ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, busy := c.inFlight[name]; busy {
		return nil, false
	}

	c.inFlight[name] = struct{}{}
	return func() {
		c.mu.Lock()
		delete(c.inFlight, name)
		c.mu.Unlock()
	}, true
}

func (c *PluginLifecycleController) timeout() time.Duration {
	if c.OpTimeout > 0 {
		return c.OpTimeout
	}
	return defaultPluginOpTimeout
}

func (c *PluginLifecycleController) logger() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

func (c *PluginLifecycleController) RestartHandler(w http.ResponseWriter, req *http.Request) error {
	c.handle(w, req, PluginOpRestart)
	return nil
}

func (c *PluginLifecycleController) StopHandler(w http.ResponseWriter, req *http.Request) error {
	c.handle(w, req, PluginOpStop)
	return nil
}

func (c *PluginLifecycleController) EnableHandler(w http.ResponseWriter, req *http.Request) error {
	c.handle(w, req, PluginOpEnable)
	return nil
}

func (c *PluginLifecycleController) TransitionHandler(w http.ResponseWriter, req *http.Request) error {
	if !httpx.RequireMethod(w, req, http.MethodPost) {
		return nil
	}

	var bodyStruct struct {
		To string `json:"to"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, maxTransitionBodyBytes))
	err := dec.Decode(&bodyStruct)
	if err == nil {
		var extra json.RawMessage
		if e := dec.Decode(&extra); !errors.Is(e, io.EOF) {
			err = errors.New("trailing data after the JSON body")
		}
	}
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			pluginWriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", "")
			return err
		}
		pluginWriteError(w, http.StatusBadRequest, "invalid_request", `body must be a JSON object like {"to":"READY"}`, "")
		return err
	}

	op, ok := pluginTransitionTargets[lifecycle.State(strings.ToUpper(strings.TrimSpace(bodyStruct.To)))]
	if !ok {
		pluginWriteError(w, http.StatusBadRequest, "invalid_target", ""+
			`"to" must be one of `+pluginAllowedTargets(), "")
		return nil
	}

	c.execute(w, req, op)
	return nil
}

func pluginOf(r *http.Request) string {
	if name := r.PathValue("name"); name != "" {
		return name
	}

	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(segments) >= 2 {
		return segments[len(segments)-2]
	}
	return ""
}

func pluginAllowedTargets() string {
	names := make([]string, 0, len(pluginTransitionTargets))
	for s := range pluginTransitionTargets {
		names = append(names, string(s))
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func pluginWriteError(w http.ResponseWriter, status int, code, message, state string) {
	pluginWriteJson(w, status, dto.PluginErrorResponse{
		Error:   code,
		Message: message,
		State:   state,
	})
}

func pluginWriteJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	httpx.WriteJSON(w, status, v)
}
