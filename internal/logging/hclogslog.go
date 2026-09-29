package logging

import (
	"context"
	"log/slog"
	"strings"

	"github.com/hashicorp/go-hclog"
)

// HCLogHandler est un slog.Handler qui écrit via un hclog.Logger.
// Comme hclog est configuré sur os.Stderr (obligatoire avec go-plugin),
// aucun octet ne part sur stdout : le protocole gRPC reste intact.
type HCLogHandler struct {
	logger hclog.Logger
	attrs  []slog.Attr
	groups []string
}

var _ slog.Handler = (*HCLogHandler)(nil)

// NewHCLogHandler construit le handler. Passez ici le hclog.Logger
// partagé avec go-plugin (celui de ServeConfig.Logger).
func NewHCLogHandler(l hclog.Logger) *HCLogHandler {
	return &HCLogHandler{logger: l}
}

// NewLogger est le raccourci : à utiliser partout dans le plugin.
func NewLogger(l hclog.Logger) *slog.Logger {
	return slog.New(NewHCLogHandler(l))
}

func (h *HCLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// hclog filtre lui-même via son Level ; on accepte tout.
	return true
}

func (h *HCLogHandler) Handle(_ context.Context, r slog.Record) error {
	// Préfixe de groupe (slog.WithGroup) → "group.key"
	var prefix strings.Builder
	for _, g := range h.groups {
		prefix.WriteString(g + ".")
	}

	args := make([]any, 0, 2*(len(h.attrs)+r.NumAttrs()))

	for _, a := range h.attrs {
		args = append(args, prefix.String()+a.Key, a.Value.Any())
	}
	r.Attrs(func(a slog.Attr) bool {
		args = append(args, prefix.String()+a.Key, a.Value.Any())
		return true
	})

	switch {
	case r.Level <= slog.LevelDebug-2:
		h.logger.Trace(r.Message, args...)
	case r.Level <= slog.LevelDebug:
		h.logger.Debug(r.Message, args...)
	case r.Level <= slog.LevelInfo:
		h.logger.Info(r.Message, args...)
	case r.Level <= slog.LevelWarn:
		h.logger.Warn(r.Message, args...)
	default:
		h.logger.Error(r.Message, args...)
	}
	return nil
}

func (h *HCLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &nh
}

func (h *HCLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	nh := *h
	nh.groups = append(append([]string{}, h.groups...), name)
	return &nh
}
