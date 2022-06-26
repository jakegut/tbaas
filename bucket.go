package tbaas

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Bucket struct {
	maxTokens int64
	interval  time.Duration
	store     *Store
	clock     Clock
}

func MakeBucket(maxTokens int64, interval time.Duration) *Bucket {
	return MakeBucketWithClock(maxTokens, interval, *newClock())
}

func MakeBucketWithClock(maxTokens int64, interval time.Duration, clock Clock) *Bucket {
	return &Bucket{
		maxTokens: maxTokens,
		interval:  interval,
		store:     MakeStore(),
		clock:     clock,
	}
}

func (b *Bucket) Take(ctx context.Context, key string, amount int64) (int64, error) {
	bk, err := b.store.Get(key)
	if err != nil {
		b.store.Put(key, &BucketKey{tokens: b.maxTokens, lastCheck: 0, parent: b})
		bk, _ = b.store.Get(key)
	}

	return bk.Take(ctx, amount)
}

type BucketKey struct {
	tokens    int64
	lastCheck int64
	mu        sync.Mutex
	parent    *Bucket
}

var ErrorTokensExceeded = errors.New("amount exceeded tokens")

func (bk *BucketKey) Take(ctx context.Context, amount int64) (int64, error) {
	bk.mu.Lock()
	defer bk.mu.Unlock()

	now := bk.parent.clock.Now().UnixNano()

	tokensToAdd := int64(float64(now-bk.lastCheck) / float64(bk.parent.interval) * float64(bk.parent.maxTokens))

	if (tokensToAdd) > 0 {
		bk.lastCheck = now
		if bk.tokens+tokensToAdd <= bk.parent.maxTokens {
			bk.tokens += tokensToAdd
		} else {
			bk.tokens = bk.parent.maxTokens
		}
	}

	if amount > bk.tokens {
		return -1, ErrorTokensExceeded
	} else {
		bk.tokens -= amount
	}

	return bk.tokens, nil
}

type Store struct {
	m  map[string]*BucketKey
	mu sync.RWMutex
}

func MakeStore() *Store {
	return &Store{
		m: make(map[string]*BucketKey, 20),
	}
}

var ErrorNoSuchKey = errors.New("no such key")

func (s *Store) Get(key string) (*BucketKey, error) {
	s.mu.RLock()
	value, ok := s.m[key]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrorNoSuchKey
	}

	return value, nil
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()

	return nil
}

func (s *Store) Put(key string, value *BucketKey) error {
	s.mu.Lock()
	s.m[key] = value
	s.mu.Unlock()

	return nil
}

type Clock struct {
	Now func() time.Time
}

func newClock() *Clock {
	return &Clock{
		Now: time.Now,
	}
}
