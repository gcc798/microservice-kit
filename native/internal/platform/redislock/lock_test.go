package redislock

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func testLocker(t *testing.T, options ...Option) (*Locker, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return New(client, options...), server
}

func TestLockSerializesConcurrentCallers(t *testing.T) {
	locker, _ := testLocker(t, WithRetryRange(time.Millisecond, 2*time.Millisecond))
	var active atomic.Int32
	var overlap atomic.Bool
	done := make(chan error, 2)
	for range 2 {
		go func() {
			done <- locker.WithLock(context.Background(), "lock:test", func() error {
				if active.Add(1) != 1 {
					overlap.Store(true)
				}
				time.Sleep(15 * time.Millisecond)
				active.Add(-1)
				return nil
			})
		}()
	}
	for range 2 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if overlap.Load() {
		t.Fatal("critical sections overlapped")
	}
}

func TestAcquireHonorsContextCancellation(t *testing.T) {
	locker, _ := testLocker(t, WithRetryRange(time.Millisecond, time.Millisecond))
	first, err := locker.Acquire(context.Background(), "lock:test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = first.Release(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := locker.Acquire(ctx, "lock:test"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Acquire() error = %v, want deadline exceeded", err)
	}
}

func TestReleaseDoesNotDeleteAnotherOwnersLock(t *testing.T) {
	locker, server := testLocker(t, WithTTL(time.Second))
	lock, err := locker.Acquire(context.Background(), "lock:test")
	if err != nil {
		t.Fatal(err)
	}
	server.Set("lock:test", "another-token")
	if err := lock.Release(context.Background()); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("Release() error = %v, want ErrNotOwner", err)
	}
	if got, _ := server.Get("lock:test"); got != "another-token" {
		t.Fatalf("lock value = %q, want another-token", got)
	}
}

func TestLockExpiresAfterHolderCrash(t *testing.T) {
	locker, server := testLocker(t, WithTTL(time.Second), WithRetryRange(time.Millisecond, time.Millisecond))
	if _, err := locker.Acquire(context.Background(), "lock:test"); err != nil {
		t.Fatal(err)
	}
	server.FastForward(2 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	lock, err := locker.Acquire(ctx, "lock:test")
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWithLockReleasesAfterFunctionError(t *testing.T) {
	locker, server := testLocker(t)
	want := errors.New("business failure")
	if err := locker.WithLock(context.Background(), "lock:test", func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("WithLock() error = %v, want business failure", err)
	}
	if server.Exists("lock:test") {
		t.Fatal("lock remained after function error")
	}
}

func TestWithLockReleasesAfterPanic(t *testing.T) {
	locker, server := testLocker(t)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("WithLock() did not propagate panic")
			}
		}()
		_ = locker.WithLock(context.Background(), "lock:test", func() error {
			panic("boom")
		})
	}()
	if server.Exists("lock:test") {
		t.Fatal("lock remained after panic")
	}
}

func TestTryClaimAllowsOnlyOneCallerUntilTTL(t *testing.T) {
	locker, server := testLocker(t, WithTTL(time.Second))
	ctx := t.Context()

	claimed, err := locker.TryClaim(ctx, "worker:cleanup")
	if err != nil || !claimed {
		t.Fatalf("first TryClaim() = %v, %v; want true, nil", claimed, err)
	}
	claimed, err = locker.TryClaim(ctx, "worker:cleanup")
	if err != nil || claimed {
		t.Fatalf("second TryClaim() = %v, %v; want false, nil", claimed, err)
	}

	server.FastForward(2 * time.Second)
	claimed, err = locker.TryClaim(ctx, "worker:cleanup")
	if err != nil || !claimed {
		t.Fatalf("TryClaim() after TTL = %v, %v; want true, nil", claimed, err)
	}
}
