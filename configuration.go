package ccache

import "strings"

type Configuration struct {
	maxSize        int64
	buckets        int
	candidates     int
	itemsToPrune   int
	initBucketSize int
	tracking       bool
	countPerSampling uint64
	onDelete       func(item *Item)
	updateRatio    float64
	evalAlgorithm  func(item *Item)float64
}

// Creates a configuration object with sensible defaults
// Use this as the start of the fluent configuration:
// e.g.: ccache.New(ccache.Configure().MaxSize(10000))
func Configure() *Configuration {
	return &Configuration{
		buckets:        16,
		candidates:     10,
		itemsToPrune:   500,
		initBucketSize: 512,
		maxSize:        5000,
		countPerSampling: 1000,
		tracking:       false,
		updateRatio:    0.3,
		evalAlgorithm:  evalLFU,
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
// [10]
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

// The evaluation algorithm
// [LFU]
func (c *Configuration) EvalAlgorithm(name string) *Configuration {
	name = strings.ToLower(name)
	if strings.Compare(name, "lfu") == 0 {
		c.evalAlgorithm = evalLFU
	} else if strings.Compare(name, "lru") == 0 {
		c.evalAlgorithm = evalLRU
	} else if strings.Compare(name, "hyperbolic") == 0 {
		c.evalAlgorithm = evalHyperbolic
	} else if strings.Compare(name, "h1") == 0 {
		c.evalAlgorithm = evalOursH1
	} else if strings.Compare(name, "h2") == 0 {
		c.evalAlgorithm = evalOursH2
	} else {
		panic("Unrecognized evaluation algorithm.")
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

// The count of accesses before each recreation of sampling tables
// [1000]
func (c *Configuration) CountPerSampling(count uint64) *Configuration {
	if count > 0 {
		c.countPerSampling = count
	}
	return c
}

// OnDelete allows setting a callback function to react to ideam deletion.
// This typically allows to do a cleanup of resources, such as calling a Close() on
// cached object that require some kind of tear-down.
func (c *Configuration) OnDelete(callback func(item *Item)) *Configuration {
	c.onDelete = callback
	return c
}