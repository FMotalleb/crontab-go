package concurrency

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestConcurrentPool_PanicCase(t *testing.T) {
	assert.Panics(t, func() {
		NewConcurrentPool(0)
	})
}

func TestConcurrentPool_LockUnlock(t *testing.T) {
	t.Run("Lock and Unlock with capacity 1", func(t *testing.T) {
		pool := NewConcurrentPool(1)
		pool.Lock()
		assert.Equal(t, 1, pool.access(get))
		pool.Unlock()
		assert.Equal(t, 0, pool.access(get))
	})

	t.Run("Lock and Unlock with capacity 2", func(t *testing.T) {
		pool := NewConcurrentPool(2)
		pool.Lock()
		assert.Equal(t, 1, pool.access(get))
		pool.Lock()
		assert.Equal(t, 2, pool.access(get))
		pool.Unlock()
		assert.Equal(t, 1, pool.access(get))
		pool.Unlock()
		assert.Equal(t, 0, pool.access(get))
	})

	t.Run("Unlock on a totally free pool", func(t *testing.T) {
		pool := NewConcurrentPool(1)
		assert.Panics(t, pool.Unlock)
	})
}

func TestConcurrentPool_LockUnlockGoroutine(t *testing.T) {
	t.Run("Lock and Unlock with capacity 1 (inside 2 goroutine)", func(t *testing.T) {
		pool := NewConcurrentPool(1)

		chn := make(chan int64)
		for i := 0; i < 2; i++ {
			go func() {
				pool.Lock()
				defer pool.Unlock()
				defer func() {
					chn <- time.Now().UnixMilli()
				}()
				time.Sleep(time.Second * 1)
			}()
		}
		end1 := <-chn
		end2 := <-chn
		diff := end2 - end1
		assert.NotZero(t, diff)
		assert.True(t, diff >= 1000)
	})
}
