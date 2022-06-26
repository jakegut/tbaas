package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var uniqKeys []string = []string{"test", "another", "key", "123"}

func toTimePointer(t time.Time) *time.Time {
	return &t
}

func TestBasicRateLimit(t *testing.T) {

	bucket := MakeBucket(10, time.Hour)

	for _, key := range uniqKeys {
		if err := bucket.Take(context.Background(), key, 1); err != nil {
			t.Fatalf("key %s failed", key)
		}
	}

	bucket = MakeBucket(10, time.Hour)

	for _, key := range uniqKeys {
		if err := bucket.Take(context.Background(), key, 11); err == nil {
			t.Fatalf("key %s failed", key)
		}
	}
}

func TestRateLimit(t *testing.T) {

	var currentTime *time.Time = toTimePointer(time.UnixMilli(0))

	testClock := newClock()
	testClock.Now = func() time.Time {
		return *currentTime
	}

	var tokens int64 = 10

	bucket := MakeBucketWithClock(tokens, time.Hour, *testClock)

	for x := 0; x < 2; x++ {
		fmt.Println(testClock.Now())
		for _, key := range uniqKeys {
			var i int64
			for i = 0; i < tokens; i++ {
				if err := bucket.Take(context.Background(), key, 1); err != nil {
					t.Fatalf("key %s failed", key)
				}
			}
			if err := bucket.Take(context.Background(), key, 1); err == nil {
				t.Fatalf("key %s failed", key)
			}
		}
		*currentTime = currentTime.Add(time.Hour)
	}
}
