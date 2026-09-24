package runtime

import (
	"context"
	"errors"
	"sync"
)

type FakeAdapter struct {
	mu       sync.Mutex
	info     InstanceInfo
	alive    bool
	exited   chan struct{}
	failNext bool
}

func NewFakeAdapter(info InstanceInfo) *FakeAdapter {
	return &FakeAdapter{info: info, exited: make(chan struct{})}
}

func (f *FakeAdapter) FailNextStart() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failNext = true
}

func (f *FakeAdapter) Start(ctx context.Context, path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext {
		f.failNext = false
		return errors.New("fake: simulated start failure")
	}
	f.alive = true
	return nil
}

func (f *FakeAdapter) Handshake(ctx context.Context) (InstanceInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.info, nil
}

func (f *FakeAdapter) Alive() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.alive
}

func (f *FakeAdapter) Exited() <-chan struct{} {
	return f.exited
}

func (f *FakeAdapter) Stop(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.alive {
		return nil
	}
	f.alive = false
	close(f.exited)
	return nil
}

func (f *FakeAdapter) Crash() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.alive {
		f.alive = false
		close(f.exited)
	}
}

func (f *FakeAdapter) Pid() int {
	return 23
}
