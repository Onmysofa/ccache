package ccache

import (
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

func (_ *BucketTests) DeleteItemFromBucket() {
	bucket := testBucket()
	bucket.delete("power")
	Expect(bucket.get("power")).To.Equal(nil)
}

func (_ *BucketTests) SetsANewBucketItem() {
	bucket := testBucket()
	val := TestValue("flow")
	item, existing := bucket.set("spice", val, getDefaultReqInfo(val), time.Minute)
	assertValue(item, "flow")
	item = bucket.get("spice")
	assertValue(item, "flow")
	Expect(existing).To.Equal(nil)
}

func (_ *BucketTests) SetsAnExistingItem() {
	bucket := testBucket()
	val := TestValue("9001")
	item, existing := bucket.set("power", val, getDefaultReqInfo(val), time.Minute)
	assertValue(existing, "9000")
	item = bucket.get("power")
	assertValue(item, "9001")
	val = TestValue("9002")
	item, existing = bucket.set("power", val, getDefaultReqInfo(val), time.Minute)
	assertValue(existing, "9001")
}

func testBucket() *bucket {
	b := &bucket{lookup: make(map[string]int), arr: NewArr(10)}
	b.lookup["power"] = 0
	item := &Item{
		idx: 0,
		key:   "power",
		value: TestValue("9000"),
	}
	b.arr = append(b.arr, item)
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
