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

type Phealth struct {
	mgr       *manager.Manager
	providers []provision.Desired
}

func (h *Phealth) states() map[string]string {
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

func (h *Phealth) PluginLive(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "OK"})
	return nil
}

func (h *Phealth) PluginReady(w http.ResponseWriter, _ *http.Request) error {
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
	return nil
}

func NewPluginHealthHandler(m *manager.Manager, ds []provision.Desired) *Phealth {
	return &Phealth{mgr: m, providers: ds}
}
