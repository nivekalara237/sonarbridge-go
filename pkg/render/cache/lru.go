package cache

import (
	"container/list"
	"context"
	"sonarbridge-go/pkg/render"
	"sync"
)

type LRU struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List
	items    map[string]*list.Element
}

type entry struct {
	tmpl render.Template
	key  string
}

var _ render.Cache = (*LRU)(nil)

func NewLRU(capacity int) *LRU {
	if capacity <= 0 {
		capacity = 256
	}

	return &LRU{
		// mu:       sync.Mutex{},
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		ll:       list.New(),
	}
}

func (c *LRU) Get(_ context.Context, key string) (render.Template, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.ll.MoveToFront(el)
	return el.Value.(*entry).tmpl, true
}

func (c *LRU) Set(_ context.Context, key string, t render.Template) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		el.Value.(*entry).tmpl = t
		return nil
	}

	c.items[key] = c.ll.PushFront(&entry{key: key, tmpl: t})

	if c.ll.Len() > c.capacity {
		c.evictOldest()
	}

	return nil
}

func (c *LRU) Invalidate(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.removeElement(el)
	}
	return nil
}

func (c *LRU) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *LRU) evictOldest() {
	if el := c.ll.Back(); el != nil {
		c.removeElement(el)
	}
}

func (c *LRU) removeElement(el *list.Element) {
	c.ll.Remove(el)
	delete(c.items, el.Value.(*entry).key)
}
