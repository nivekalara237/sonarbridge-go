package lifecycle

import "slices"

import "fmt"

type State string

const (
	StateNotInstalled State = "NOT_INSTALLED"
	StateInstalled    State = "INSTALLED"
	StateDisabled     State = "DISABLED"
	StateEnabled      State = "ENABLED"
	StateStarting     State = "STARTING"
	StateHandshaking  State = "HANDSHAKING"
	StateReady        State = "READY"
	StateCrashed      State = "CRASHED"
	StateStopping     State = "STOPPING"
	StateStopped      State = "STOPPED"
	StateFailed       State = "FAILED"
)

// transitions defines the allowed State -> []Stage edges. Anything not
// listed here is rejected by Transition.
var transitions = map[State][]State{
	StateNotInstalled: {StateInstalled},
	StateInstalled:    {StateEnabled, StateDisabled},
	StateDisabled:     {StateEnabled, StateNotInstalled},
	StateEnabled:      {StateStarting, StateDisabled, StateNotInstalled},
	StateStarting:     {StateHandshaking, StateFailed},
	StateHandshaking:  {StateReady, StateFailed},
	StateReady:        {StateCrashed, StateStopping},
	StateCrashed:      {StateStarting, StateFailed, StateDisabled}, // retry (backoff), give upn or disable to stop retrying
	StateStopping:     {StateStopped},
	StateStopped:      {StateStarting, StateNotInstalled, StateDisabled},
	StateFailed:       {StateStarting, StateNotInstalled, StateDisabled}, // manual retry / uninstall / disable
}

// FSM is a minimal state machine for one plugin instance. It is not
// go-routine-safe on its own - the Manager is expected to serialize access
// (it already holds a mutex per instance).
type FSM struct {
	current State
}

func NewFSM() *FSM {
	return &FSM{current: StateNotInstalled}
}

func (f *FSM) Current() State {
	return f.current
}

// Transition attempts to move to next. It returns an error without mutating the
// FSM if the edge current -> next is not declared/
func (f *FSM) Transition(next State) error {
	allowed, ok := transitions[f.current]
	if !ok {
		return fmt.Errorf("lifecycle: unknown current state %q", f.current)
	}

	if slices.Contains(allowed, next) {
		f.current = next
		return nil
	}

	return fmt.Errorf("lifecycle: illegal transition %s -> %s", f.current, next)
}
