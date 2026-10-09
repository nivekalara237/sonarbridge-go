package rest

import (
	"net/http"
	"sonarbridge-go/internal/infra/entrypoints/rest/httpx"

	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/provision"
)

const (
	stateReady        = "READY"
	stateNotInstalled = "NOT_INSTALLED"
)

type PHealth struct {
	mgr       *manager.Manager
	providers []provision.Desired
}

func (h *PHealth) states() map[string]string {
	live := make(map[string]string)
	for _, s := range h.mgr.List() {
		live[s.Name] = string(s.State)
	}

	out := make(map[string]string, len(h.providers))

	for _, p := range h.providers {
		if s, ok := live[p.Name]; ok {
			out[p.Name] = s
		} else {
			out[p.Name] = stateNotInstalled
		}
	}

	return out
}

func (h *PHealth) PluginLive(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "OK"})
}

func (h *PHealth) PluginReady(w http.ResponseWriter, _ *http.Request) {
	states := h.states()
	ready := true
	for _, s := range states {
		if s != stateReady {
			ready = false
		}
	}

	code := http.StatusOK
	if !ready {
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, code, map[string]any{"ready": ready, "providers": states})
}

func NewPluginHealthCheckHandler(m *manager.Manager, ds []provision.Desired) *PHealth {
	return &PHealth{mgr: m, providers: ds}
}
