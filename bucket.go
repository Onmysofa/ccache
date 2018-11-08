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
	updateRatio float64
}

func NewArr(initSize int) []*Item {
	return make([]*Item, 0, initSize)
}

func NewBucket(initSize int, ur float64) *bucket {
	return &bucket{
		lookup: make(map[string]int),
		arr: NewArr(initSize),
		init: initSize,
		updateRatio: ur,
	}
}

func (b *bucket) get(key string, t *RecursionTimer) *Item {
	t.Enter("bucket:get")
	defer t.Leave()

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

func (b *bucket) set(key string, value interface{}, r *ReqInfo, duration time.Duration, t *RecursionTimer) (*Item, *Item) {
	t.Enter("bucket:set")
	defer t.Leave()

	expires := time.Now().Add(duration).UnixNano()
	item := newItem(key, value, r, expires)
	b.Lock()
	defer b.Unlock()

	existingId, ok := b.lookup[key]
	if ok {
		existing := b.arr[existingId]
		b.arr[existingId] = item
		item.idx = existingId
		item.MixReqInfo(&existing.reqInfo, b.updateRatio)
		return item, existing
	} else {
		b.arr = append(b.arr, item)
		item.idx = len(b.arr) - 1
		b.lookup[key] = item.idx
		return item, nil
	}
}

func (b *bucket) delete(key string, t *RecursionTimer) (*Item, bool) {
	t.Enter("bucket:delete")
	defer t.Leave()

	b.Lock()
	defer b.Unlock()
	return b.deleteInner(key)
}

func (b *bucket) deleteInner(key string) (*Item, bool) {

	itemId, ok := b.lookup[key]
	if ok {
		last := b.arr[len(b.arr) - 1]

		item := b.arr[itemId]
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

func (b *bucket) getCandidate(t *RecursionTimer) (*Item, int32) {
	t.Enter("bucket:getCandidate")
	defer t.Leave()

	t.Enter("bucket:getCandidate:acquireLock")
	b.RLock()
	t.Leave()
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

