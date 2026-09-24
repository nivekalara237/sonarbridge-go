package lifecycle_test

import (
	"sonarbridge-go/internal/plugin/lifecycle"
	"testing"
)

func TestFSMHappyPath(t *testing.T) {
	f := lifecycle.NewFSM()
	steps := []lifecycle.State{
		lifecycle.StateInstalled,
		lifecycle.StateEnabled,
		lifecycle.StateStarting,
		lifecycle.StateHandshaking,
		lifecycle.StateReady,
		lifecycle.StateStopping,
		lifecycle.StateStopped,
	}
	if f.Current() != lifecycle.StateNotInstalled {
		t.Fatalf("expected NOT_INSTALLED, got %s", f.Current())
	}

	for _, s := range steps {
		if err := f.Transition(s); err != nil {
			t.Fatalf("transition to %s: %v", s, err)
		}
	}

	if f.Current() != lifecycle.StateStopped {
		t.Fatalf("expected STOPPED, got %s", f.Current())
	}
}

func TestFSMRejectsIllegalTransition(t *testing.T) {
	f := lifecycle.NewFSM()
	if err := f.Transition(lifecycle.StateReady); err == nil {
		t.Fatalf("expected error jumping straight to READY form NOT_INSTALLED")
	}
}

func TestFSMCrashAndRetry(t *testing.T) {
	f := lifecycle.NewFSM()
	for _, s := range []lifecycle.State{
		lifecycle.StateInstalled, lifecycle.StateEnabled,
		lifecycle.StateStarting, lifecycle.StateHandshaking, lifecycle.StateReady,
	} {
		if err := f.Transition(s); err != nil {
			t.Fatalf("setup transition to %s: %v", s, err)
		}
	}

	if err := f.Transition(lifecycle.StateCrashed); err != nil {
		t.Fatalf("Ready -> CRASHED: %v", err)
	}

	if err := f.Transition(lifecycle.StateStarting); err != nil {
		t.Fatalf("CRASHED -> STARTING (retry): %v", err)
	}
}
