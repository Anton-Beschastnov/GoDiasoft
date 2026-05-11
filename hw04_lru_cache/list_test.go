package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestListPushFront(t *testing.T) {
	l := NewList()

	item1 := l.PushFront(1)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 1, l.Front().Value)
	require.Equal(t, 1, l.Back().Value)
	require.Nil(t, item1.Prev)
	require.Nil(t, item1.Next)

	item2 := l.PushFront(2)
	require.Equal(t, 2, l.Len())
	require.Equal(t, 2, l.Front().Value)
	require.Equal(t, 1, l.Back().Value)
	require.Equal(t, item2, l.Front())
	require.Equal(t, item1, l.Front().Next)
	require.Equal(t, item2, l.Back().Prev)
}

func TestListPushBack(t *testing.T) {
	l := NewList()

	item1 := l.PushBack(1)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 1, l.Front().Value)
	require.Equal(t, 1, l.Back().Value)
	require.Nil(t, item1.Prev)
	require.Nil(t, item1.Next)

	item2 := l.PushBack(2)
	require.Equal(t, 2, l.Len())
	require.Equal(t, 1, l.Front().Value)
	require.Equal(t, 2, l.Back().Value)
	require.Equal(t, item2, l.Back())
	require.Equal(t, item1, l.Front())
	require.Equal(t, item2, l.Front().Next)
	require.Equal(t, item1, l.Back().Prev)
}

func TestListRemove(t *testing.T) {
	t.Run("remove from middle", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)

		middle := l.Front().Next
		l.Remove(middle)

		require.Equal(t, 2, l.Len())
		require.Equal(t, 1, l.Front().Value)
		require.Equal(t, 3, l.Back().Value)
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
		require.Equal(t, l.Back(), l.Front().Next)
		require.Equal(t, l.Front(), l.Back().Prev)
	})

	t.Run("remove front", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)

		l.Remove(l.Front())

		require.Equal(t, 2, l.Len())
		require.Equal(t, 2, l.Front().Value)
		require.Equal(t, 3, l.Back().Value)
	})

	t.Run("remove back", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)

		l.Remove(l.Back())

		require.Equal(t, 2, l.Len())
		require.Equal(t, 1, l.Front().Value)
		require.Equal(t, 2, l.Back().Value)
	})

	t.Run("remove only element", func(t *testing.T) {
		l := NewList()
		item := l.PushBack(1)

		l.Remove(item)

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
}

func TestListMoveToFront(t *testing.T) {
	t.Run("move middle to front", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)

		middle := l.Front().Next
		l.MoveToFront(middle)

		require.Equal(t, 3, l.Len())
		require.Equal(t, 2, l.Front().Value)
		require.Equal(t, 1, l.Front().Next.Value)
		require.Equal(t, 3, l.Back().Value)
	})

	t.Run("move back to front", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)

		l.MoveToFront(l.Back())

		require.Equal(t, 3, l.Len())
		require.Equal(t, 3, l.Front().Value)
		require.Equal(t, 1, l.Front().Next.Value)
		require.Equal(t, 2, l.Back().Value)
	})

	t.Run("move front to front", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.PushBack(2)
		l.PushBack(3)

		l.MoveToFront(l.Front())

		require.Equal(t, 3, l.Len())
		require.Equal(t, 1, l.Front().Value)
		require.Equal(t, 2, l.Front().Next.Value)
		require.Equal(t, 3, l.Back().Value)
	})
}
