package ccache

import (
	"sync/atomic"
	"time"
)

type Sized interface {
	Size() int64
}

type TrackedItem interface {
	Value() interface{}
	Release()
	Expired() bool
	TTL() time.Duration
	Expires() time.Time
	Extend(duration time.Duration)
}

type ReqInfo struct {
	TimeEntered time.Time
	ReqSize float64
}

type nilItem struct{}

func (n *nilItem) Value() interface{} { return nil }
func (n *nilItem) Release()           {}

func (i *nilItem) Expired() bool {
	return true
}

func (i *nilItem) TTL() time.Duration {
	return time.Minute
}

func (i *nilItem) Expires() time.Time {
	return time.Time{}
}

func (i *nilItem) Extend(duration time.Duration) {
}

var NilTracked = new(nilItem)

type Item struct {
	idx		   int
	key        string
	group      string
	promotions int32
	refCount   int32
	accCount   int64
	expires    int64
	size       int64
	value      interface{}
	reqInfo    ReqInfo
	createTS   time.Time
	accessTs   time.Time
}

func newItem(key string, value interface{}, r *ReqInfo, expires int64) *Item {
	size := getValueSize(value)

	return &Item{
		key:        key,
		value:      value,
		promotions: 0,
		size:       size,
		expires:    expires,
		reqInfo:    *r,
		createTS:   time.Now(),
		accessTs:   time.Now(),
	}
}

func getValueSize(value interface{}) int64 {
	size := int64(1)
	if sized, ok := value.(Sized); ok {
		size = sized.Size()
	}
	return size
}

func getDefaultReqInfo(value interface{}) *ReqInfo {

	r := &ReqInfo{time.Now(), float64(getValueSize(value))}

	return r
}

func (i *Item) shouldPromote(getsPerPromote int32) bool {
	i.promotions += 1
	return i.promotions == getsPerPromote
}

func (i *Item) Value() interface{} {
	return i.value
}

func (i *Item) track() {
	atomic.AddInt32(&i.refCount, 1)
}

func (i *Item) Release() {
	atomic.AddInt32(&i.refCount, -1)
}

func (i *Item) Expired() bool {
	expires := atomic.LoadInt64(&i.expires)
	return expires < time.Now().UnixNano()
}

func (i *Item) TTL() time.Duration {
	expires := atomic.LoadInt64(&i.expires)
	return time.Nanosecond * time.Duration(expires-time.Now().UnixNano())
}

func (i *Item) Expires() time.Time {
	expires := atomic.LoadInt64(&i.expires)
	return time.Unix(0, expires)
}

func (i *Item) Extend(duration time.Duration) {
	atomic.StoreInt64(&i.expires, time.Now().Add(duration).UnixNano())
}

func (i *Item) MixReqInfo(old *ReqInfo, updateRatio float64) {
	i.reqInfo.ReqSize = i.reqInfo.ReqSize * updateRatio + old.ReqSize * (1-updateRatio)
}