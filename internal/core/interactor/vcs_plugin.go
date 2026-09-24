package interactor

import "sonarbridge-go/internal/core/domain/plugin"

type VCSPlugin interface {
	Name() string
	Version() string
	Capabilities() []plugin.Capability
}
