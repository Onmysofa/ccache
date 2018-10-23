package ccache

import (
	"container/heap"
	. "github.com/karlseguin/expect"
	"testing"
	"time"
)

type BucketTests struct {
}

func Test_Bucket(t *testing.T) {
	Expectify(new(BucketTests), t)
}

func (_ *BucketTests) GetMissFromBucket() {
	bucket := testBucket()
	Expect(bucket.get("invalid")).To.Equal(nil)
}

func (_ *BucketTests) GetHitFromBucket() {
	bucket := testBucket()
	item := bucket.get("power")
	assertValue(item, "9000")
}

func (_ *BucketTests) TestPQ() {
	bucket := testBucket()
	i1, i2 := bucket.set("what", TestValue("ok"), time.Minute)
	assertValue(i1, "ok")
	Expect(i2).To.Equal(nil)
	i1, i2 = bucket.set("how", TestValue("alright"), time.Minute)
	assertValue(i1, "alright")
	Expect(i2).To.Equal(nil)
	item := bucket.get("what")
	assertValue(item, "ok")
	item = bucket.get("how")
	assertValue(item, "alright")
	assertValue(bucket.pq.Peek(), "9000")
}

func (_ *BucketTests) DeleteItemFromBucket() {
	bucket := testBucket()
	bucket.delete("power")
	Expect(bucket.get("power")).To.Equal(nil)
}

func (_ *BucketTests) SetsANewBucketItem() {
	bucket := testBucket()
	item, existing := bucket.set("spice", TestValue("flow"), time.Minute)
	assertValue(item, "flow")
	item = bucket.get("spice")
	assertValue(item, "flow")
	Expect(existing).To.Equal(nil)
}

func (_ *BucketTests) SetsAnExistingItem() {
	bucket := testBucket()
	item, existing := bucket.set("power", TestValue("9001"), time.Minute)
	assertValue(existing, "9000")
	item = bucket.get("power")
	assertValue(item, "9001")
	item, existing = bucket.set("power", TestValue("9002"), time.Minute)
	assertValue(existing, "9001")
}

func testBucket() *bucket {
	b := &bucket{lookup: make(map[string]*Item), pq: NewPQ()}
	b.lookup["power"] = &Item{
		key:   "power",
		value: TestValue("9000"),
	}
	heap.Push(b.pq, b.lookup["power"])
	return b
}

func assertValue(item *Item, expected string) {
	value := item.value.(TestValue)
	Expect(value).To.Equal(TestValue(expected))
}

type TestValue string

func (v TestValue) Expires() time.Time {
	return time.Now()
}
