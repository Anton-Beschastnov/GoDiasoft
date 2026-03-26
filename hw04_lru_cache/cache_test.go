package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 1, val)

		c.Set("d", 4)

		_, ok = c.Get("a")
		require.True(t, ok)

		_, ok = c.Get("b")
		require.False(t, ok)

		_, ok = c.Get("c")
		require.True(t, ok)

		_, ok = c.Get("d")
		require.True(t, ok)
	})

	t.Run("eviction by capacity", func(t *testing.T) {
		c := NewCache(3)

		c.Set("key1", 100)
		c.Set("key2", 200)
		c.Set("key3", 300)
		c.Set("key4", 400)

		require.Equal(t, 3, c.(*lruCache).queue.Len())

		_, ok := c.Get("key1")
		require.False(t, ok)

		val, ok := c.Get("key2")
		require.True(t, ok)
		require.Equal(t, 200, val)

		val, ok = c.Get("key3")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("key4")
		require.True(t, ok)
		require.Equal(t, 400, val)
	})

	t.Run("eviction by lru order", func(t *testing.T) {
		c := NewCache(3)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		c.Get("a")
		c.Set("a", 10)
		c.Get("c")

		c.Set("d", 4)

		_, ok := c.Get("b")
		require.False(t, ok)

		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 10, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)

		val, ok = c.Get("d")
		require.True(t, ok)
		require.Equal(t, 4, val)
	})

	t.Run("clear cache", func(t *testing.T) {
		c := NewCache(3)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		c.Clear()

		require.Equal(t, 0, c.(*lruCache).queue.Len())
		require.Equal(t, 0, len(c.(*lruCache).items))

		_, ok := c.Get("a")
		require.False(t, ok)

		_, ok = c.Get("b")
		require.False(t, ok)

		_, ok = c.Get("c")
		require.False(t, ok)
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
