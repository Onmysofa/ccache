// An LRU cached aimed at high concurrency
package ccache

import "C"
import (
	"hash/fnv"
	"math/rand"
	"sync/atomic"
	"time"
)

type Cache struct {
	*Configuration
	size        int64
	buckets     []*bucket
	bucketMask  uint32
	deletables  chan *Item
	promotables chan *Item
	donec       chan struct{}
}

// Create a new cache with the specified configuration
// See ccache.Configure() for creating a configuration
func New(config *Configuration) *Cache {
	c := &Cache{
		Configuration: config,
		bucketMask:    uint32(config.buckets) - 1,
		buckets:       make([]*bucket, config.buckets),
	}
	for i := 0; i < int(config.buckets); i++ {
		c.buckets[i] = NewBucket(config.initBucketSize)
	}
	c.restart()
	return c
}

// Get an item from the cache. Returns nil if the item wasn't found.
// This can return an expired item. Use item.Expired() to see if the item
// is expired and item.TTL() to see how long until the item expires (which
// will be negative for an already expired item).
func (c *Cache) Get(key string) *Item {
	item := c.bucket(key).get(key)
	if item == nil {
		return nil
	}

	return item
}

// Used when the cache was created with the Track() configuration option.
// Avoid otherwise
func (c *Cache) TrackingGet(key string) TrackedItem {
	item := c.Get(key)
	if item == nil {
		return NilTracked
	}
	item.track()
	return item
}

// Set the value in the cache for the specified duration
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.set(key, value, duration)
}

// Replace the value if it exists, does not set if it doesn't.
// Returns true if the item existed an was replaced, false otherwise.
// Replace does not reset item's TTL
func (c *Cache) Replace(key string, value interface{}) bool {
	item := c.bucket(key).get(key)
	if item == nil {
		return false
	}
	c.Set(key, value, item.TTL())
	return true
}

// Attempts to get the value from the cache and calles fetch on a miss (missing
// or stale item). If fetch returns an error, no value is cached and the error
// is returned back to the caller.
func (c *Cache) Fetch(key string, duration time.Duration, fetch func() (interface{}, error)) (*Item, error) {
	item := c.Get(key)
	if item != nil && !item.Expired() {
		return item, nil
	}
	value, err := fetch()
	if err != nil {
		return nil, err
	}
	return c.set(key, value, duration), nil
}

// Remove the item from the cache, return true if the item was present, false otherwise.
func (c *Cache) Delete(key string) bool {
	item, _ := c.bucket(key).delete(key)
	if item != nil {
		//c.deletables <- item
		c.afterDelete(item)
		return true
	}
	return false
}

//this isn't thread safe. It's meant to be called from non-concurrent tests
func (c *Cache) Clear() {
	for _, bucket := range c.buckets {
		bucket.clear()
	}
	atomic.StoreInt64(&c.size, 0)
}

// Stops the background worker. Operations performed on the cache after Stop
// is called are likely to panic
func (c *Cache) Stop() {
	//close(c.promotables)
	//<-c.donec
}

func (c *Cache) restart() {
	//c.deletables = make(chan *Item, c.deleteBuffer)
	//c.promotables = make(chan *Item, c.promoteBuffer)
	//c.donec = make(chan struct{})
	//go c.worker()
}

func (c *Cache) deleteItem(bucket *bucket, item *Item) bool {
	_, ok := bucket.delete(item.key) //stop other GETs from getting it
	if ok {
		//c.deletables <- item
		c.afterDelete(item)
	}
	return ok
}

func (c *Cache) set(key string, value interface{}, duration time.Duration) *Item {
	item, existing := c.bucket(key).set(key, value, duration)
	if existing != nil {
		//c.deletables <- existing
		c.afterDelete(existing)
	}
	c.introduce(item)
	return item
}

func (c *Cache) bucket(key string) *bucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&c.bucketMask]
}

func (c *Cache) introduce(item *Item) {
	//c.promotables <- item

	c.atInsert(item)
	c.gc()
}

//func (c *Cache) worker() {
//	defer close(c.donec)
//
//	for {
//		select {
//		case item, ok := <-c.promotables:
//			if ok == false {
//				goto drain
//			}
//
//			c.atInsert(item)
//			if c.size > c.maxSize {
//				c.gc()
//			}
//		case item := <-c.deletables:
//			c.afterDelete(item)
//		}
//	}
//
//drain:
//	for {
//		select {
//		case item := <-c.deletables:
//			c.afterDelete(item)
//		default:
//			close(c.deletables)
//			return
//		}
//	}
//}

func (c *Cache) afterDelete(item *Item) {

	atomic.AddInt64(&c.size, -item.size)

	if c.onDelete != nil {
		c.onDelete(item)
	}
}

func (c *Cache) atInsert(item *Item) {

	atomic.AddInt64(&c.size, item.size)
}

func (c *Cache) gc() {
	s := atomic.LoadInt64(&c.size)
	if s <= c.maxSize {
		return
	}

	i := 0
	for s = atomic.LoadInt64(&c.size); s > c.maxSize || i < c.itemsToPrune; s = atomic.LoadInt64(&c.size) {

		var minBucket int
		var minItem *Item
		var minVal int32

		for j := 0; j < c.candidates; j++ {

			preSums := make([]int, c.Configuration.buckets, c.Configuration.buckets)
			sum := 0;

			for k := 0; k < c.Configuration.buckets; k++ {
				preSums[k] = sum + c.buckets[k].getNum()
				sum = preSums[k]
			}

			r := rand.Intn(preSums[c.Configuration.buckets - 1]) + 1
			left := 0;
			right := c.Configuration.buckets - 1;
			for left < right {
				mid := (right - left) / 2 + left;
				if preSums[mid] >= r {
					right = mid;
				} else {
					left = mid + 1;
				}
			}

			bucket := left
			curItem, curVal := c.buckets[bucket].getCandidate()

			if curItem != nil {
				if minItem == nil {
					minItem = curItem
					minVal = curVal
					minBucket = bucket
				} else {
					if curVal < minVal {
						minItem = curItem
						minVal = curVal
						minBucket = bucket
					}
				}
			}
			// Possible nil result, purposely left there to avoid infinite loop
		}

		if minItem != nil {
			if _, ok := c.buckets[minBucket].delete(minItem.key); ok {
				c.afterDelete(minItem)
			}
		}

		i++
	}
}