package ccache

import "time"

type Request struct {
	Backend uint64
	Uri uint64
	obj interface{}
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

		req.obj = item.value
	}

	return nil
}

// Set the value in the cache for the specified duration
func (c *Cache) SetPage(reqs []*Request, duration time.Duration) {
	for _, req := range reqs {
		key := buildKey(req.Backend, req.Uri)
		value := req.obj
		c.Set(key, value, duration)
	}
}
