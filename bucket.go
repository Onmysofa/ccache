package ccache

import (
	"container/heap"
	"sync"
	"time"
)

type bucket struct {
	sync.RWMutex
	lookup map[string]*Item
	pq *PriorityQueue
}

func (b *bucket) get(key string) *Item {
	b.RLock()
	defer b.RUnlock()
	item, ok := b.lookup[key]
	if ok {
		heap.Remove(b.pq, item.idx)
		item.accCount++
		heap.Push(b.pq, item)
	}

	return item
}

func (b *bucket) set(key string, value interface{}, duration time.Duration) (*Item, *Item) {
	expires := time.Now().Add(duration).UnixNano()
	item := newItem(key, value, expires)
	b.Lock()
	defer b.Unlock()
	existing, ok := b.lookup[key]
	if ok {
		heap.Remove(b.pq, existing.idx)
	}
	b.lookup[key] = item
	heap.Push(b.pq, item)
	return item, existing
}

func (b *bucket) delete(key string) *Item {
	b.Lock()
	defer b.Unlock()
	item := b.lookup[key]
	heap.Remove(b.pq, item.idx)
	delete(b.lookup, key)
	return item
}

func (b *bucket) clear() {
	b.Lock()
	defer b.Unlock()
	b.lookup = make(map[string]*Item)
	b.pq = NewPQ()
}


