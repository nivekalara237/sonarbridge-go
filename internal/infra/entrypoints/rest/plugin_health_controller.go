package rest


import (
	"encoding/json"
	"net/http"
 
	"sonarbridge-go/internal/plugin/manager"
	"sonarbridge-go/internal/plugin/provision"
)

const (
	stateReady 			= "READY"
	stateNotInstalled 	= "NOT_INSTALLED"
)

type phealth struct {
	mgr *manager.Manager
	providers []provision.Desired
}

func (h *phealth) states() map[string]string {
	live := make(map[string]string)
	for _, s := range h.mgr.List() {
		live[s.Name] = string(s.State)
	}

	out := make(map[string]string, len(h.providers))

	for _, p := h.providers {
		if s, ok := live[p.Name]; ok {
			out[p.Name] = s
		} else {
			out[p.Name] = stateNotInstalled
		}
	}

	return out
}

func (h *phealth) PluginLive(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "OK"})
}

func (h *phealth) PluginReady(w http.ResponseWriter, _ *http.Request) error {
	states := h.states()
	ready := true
	for _, s := range states {
		if s != stateReady {
			ready = false
		}
	}

	code := http.StatusOk
	if !ready {
		code = http.StatusServiceUnavaillable
	}

	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, code, map[string]any{"ready": ready, "providers": states})
}

func NewPluginHealthHandler(m manager.Manager, ds []Desired) *phealth {
	return &phealth{ mgr: m, providers: ds}
}
