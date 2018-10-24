package ccache

import (
	"math/rand"
	"sync"
	"time"
)

type bucket struct {
	sync.RWMutex
	lookup map[string]int
	arr []*Item
	init int
}

func NewArr(initSize int) []*Item {
	return make([]*Item, 0, initSize)
}

func NewBucket(initSize int) *bucket {
	return &bucket{
		lookup: make(map[string]int),
		arr: NewArr(initSize),
		init: initSize,
	}
}

func (b *bucket) get(key string) *Item {
	b.RLock()
	defer b.RUnlock()
	itemId, ok := b.lookup[key]
	if ok {
		item := b.arr[itemId]
		item.accCount++
		return item
	}

	return nil
}

func (b *bucket) set(key string, value interface{}, duration time.Duration) (*Item, *Item) {
	expires := time.Now().Add(duration).UnixNano()
	item := newItem(key, value, expires)
	b.Lock()
	defer b.Unlock()

	b.arr = append(b.arr, item)
	existingId, ok := b.lookup[key]
	if ok {
		existing := b.arr[existingId]
		b.arr[existingId] = item
		item.idx = existingId
		return item, existing
	} else {
		b.arr = append(b.arr, item)
		item.idx = len(b.arr) - 1
		b.lookup[key] = item.idx
		return item, nil
	}
}

func (b *bucket) delete(key string) (*Item, bool) {
	b.Lock()
	defer b.Unlock()
	return b.deleteInner(key)
}

func (b *bucket) deleteInner(key string) (*Item, bool) {

	itemId, ok := b.lookup[key]
	if ok {
		last := b.arr[len(b.arr) - 1]
		item :=b.arr[itemId]
		last.idx, item.idx = item.idx, last.idx
		b.arr[last.idx] = last
		b.arr[item.idx] = item
		b.lookup[last.key] = last.idx

		b.arr = b.arr[:len(b.arr)-1]
		delete(b.lookup, key)

		return item, true
	}
	return nil, false
}

func (b *bucket) getNum() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.arr)
}

func (b *bucket) getCandidate() (*Item, int32) {
	b.RLock()
	defer b.RUnlock()

	l := len(b.arr)
	if l == 0 {
		return nil, 0
	}
	itemId := rand.Intn(l)
	item := b.arr[itemId]
	return item, eval(item)
}

func (b *bucket) clear() {
	b.Lock()
	defer b.Unlock()
	b.lookup = make(map[string]int)
	b.arr = NewArr(b.init)
}

