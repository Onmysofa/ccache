package ccache

import (
	"sync/atomic"
	"time"
)

type Request struct {
	Backend uint64
	Uri uint64
	Obj interface{}
}

// Get an item from the cache. Returns nil if the item wasn't found.
// This can return an expired item. Use item.Expired() to see if the item
// is expired and item.TTL() to see how long until the item expires (which
// will be negative for an already expired item).
func (c *Cache) GetPage(reqs []*Request) error {

	for _, req := range reqs {
		key := buildKey(req.Backend, req.Uri)

		item := c.Get(key)
		if item == nil {
			continue
		}

		req.Obj = item.value
	}

	return nil
}

// Set the value in the cache for the specified duration
func (c *Cache) SetPage(reqs []*Request, duration time.Duration) {

	size := int64(0)

	for _, req := range reqs {
		size += getValueSize(req.Obj)
	}

	info := &ReqInfo{time.Now(), float64(size), float64(size)}

	for _, req := range reqs {
		key := buildKey(req.Backend, req.Uri)
		value := req.Obj
		c.SetWithInfo(key, value, info, duration)
	}
}

// Set the value in the cache for the specified duration
func (c *Cache) SetPageWithMissingSize(reqs []*Request, missingSize float64, duration time.Duration) {

	if c.admissionPolicy {
		if float64(atomic.LoadInt64(&c.size)) + missingSize > float64(c.maxSize) && missingSize > float64(c.admissionThres) {
			return
		}
	}

	size := int64(0)

	for _, req := range reqs {
		size += getValueSize(req.Obj)
	}

	info := &ReqInfo{time.Now(), float64(size), missingSize}

	for _, req := range reqs {
		key := buildKey(req.Backend, req.Uri)
		value := req.Obj
		c.SetWithInfo(key, value, info, duration)
	}
}

// Set the value in the cache for the specified duration
func (c *Cache) SetWithInfo(key string, value interface{}, r *ReqInfo, duration time.Duration) {
	atomic.AddUint64(&c.counter, 1)
	c.set(key, value, r, duration)
}

// Replace the value if it exists, does not set if it doesn't.
// Returns true if the item existed an was replaced, false otherwise.
// Replace does not reset item's TTL
func (c *Cache) ReplaceWithInfo(key string, r *ReqInfo, value interface{}) bool {
	item := c.bucket(key).get(key)
	if item == nil {
		return false
	}
	c.SetWithInfo(key, value, r, item.TTL())
	return true
}