package ccache

type Configuration struct {
	maxSize        int64
	buckets        int
	candidates     int
	itemsToPrune   int
	initBucketSize int
	tracking       bool
	onDelete       func(item *Item)
}

// Creates a configuration object with sensible defaults
// Use this as the start of the fluent configuration:
// e.g.: ccache.New(ccache.Configure().MaxSize(10000))
func Configure() *Configuration {
	return &Configuration{
		buckets:        16,
		candidates:     3,
		itemsToPrune:   500,
		initBucketSize: 512,
		maxSize:        5000,
		tracking:       false,
	}
}

// The max size for the cache
// [5000]
func (c *Configuration) MaxSize(max int64) *Configuration {
	c.maxSize = max
	return c
}

// Keys are hashed into % bucket count to provide greater concurrency (every set
// requires a write lock on the bucket). Must be a power of 2 (1, 2, 4, 8, 16, ...)
// [16]
func (c *Configuration) Buckets(count uint32) *Configuration {
	if count == 0 || ((count&(^count+1)) == count) == false {
		count = 16
	}
	c.buckets = int(count)
	return c
}

// Number of eviction candidates
// [3]
func (c *Configuration) Candidates(count int) *Configuration {
	if count >= 0 && count <= c.buckets {
		c.candidates = count
	}
	return c
}

// The number of items to prune when memory is low
// [500]
func (c *Configuration) ItemsToPrune(count uint32) *Configuration {
	c.itemsToPrune = int(count)
	return c
}

// The initial size of bucket
// [512]
func (c *Configuration) InitBucketSize(size uint32) *Configuration {
	if size > 0 {
		c.initBucketSize = int(size)
	}
	return c
}

// Typically, a cache is agnostic about how cached values are use. This is fine
// for a typical cache usage, where you fetch an item from the cache, do something
// (write it out) and nothing else.

// However, if callers are going to keep a reference to a cached item for a long
// time, things get messy. Specifically, the cache can evict the item, while
// references still exist. Technically, this isn't an issue. However, if you reload
// the item back into the cache, you end up with 2 objects representing the same
// data. This is a waste of space and could lead to weird behavior (the type an
// identity map is meant to solve).

// By turning tracking on and using the cache's TrackingGet, the cache
// won't evict items which you haven't called Release() on. It's a simple reference
// counter.
func (c *Configuration) Track() *Configuration {
	c.tracking = true
	return c
}

// OnDelete allows setting a callback function to react to ideam deletion.
// This typically allows to do a cleanup of resources, such as calling a Close() on
// cached object that require some kind of tear-down.
func (c *Configuration) OnDelete(callback func(item *Item)) *Configuration {
	c.onDelete = callback
	return c
}