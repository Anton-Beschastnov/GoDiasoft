package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun_General(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks, runTasksCount, sumTime := createSuccessTasks(tasksCount)

		workersCount, maxErrorsCount := 5, 1
		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount.Load())
		require.LessOrEqual(t, int64(time.Since(start)), int64(sumTime/2))
	})

	t.Run("empty tasks list", func(t *testing.T) {
		err := Run([]Task{}, 5, 3)
		require.NoError(t, err)
	})

	t.Run("n greater than tasks count", func(t *testing.T) {
		var runCount atomic.Int32
		tasks := []Task{func() error { runCount.Add(1); return nil }}
		err := Run(tasks, 10, 1)
		require.NoError(t, err)
		require.Equal(t, int32(1), runCount.Load())
	})
}

func TestRun_ErrorLimits(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("stop on error limit - basic", func(t *testing.T) {
		tasksCount, workers, maxErrors := 100, 5, 3
		var runCount atomic.Int32
		tasks := make([]Task, tasksCount)
		for i := range tasks {
			tasks[i] = func() error { runCount.Add(1); return errors.New("err") }
		}

		err := Run(tasks, workers, maxErrors)
		require.True(t, errors.Is(err, ErrErrorsLimitExceeded))
		require.LessOrEqual(t, runCount.Load(), int32(workers+maxErrors))
	})

	t.Run("if were errors in first M tasks", func(t *testing.T) {
		var runCount atomic.Int32
		tasks := make([]Task, 50)
		for i := range tasks {
			tasks[i] = func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
				runCount.Add(1)
				return errors.New("err")
			}
		}

		err := Run(tasks, 10, 23)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)
		require.LessOrEqual(t, runCount.Load(), int32(10+23))
	})
}

func TestRun_IgnoreErrors(t *testing.T) {
	defer goleak.VerifyNone(t)

	limits := []int{0, -1}
	for _, m := range limits {
		t.Run(fmt.Sprintf("m=%d means ignore", m), func(t *testing.T) {
			var runCount atomic.Int32
			tasks := make([]Task, 10)
			for i := range tasks {
				tasks[i] = func() error { runCount.Add(1); return errors.New("err") }
			}

			err := Run(tasks, 3, m)
			require.NoError(t, err)
			require.Equal(t, int32(10), runCount.Load())
		})
	}
}

func TestRun_Concurrency(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("concurrency test with Eventually", func(t *testing.T) {
		tasksCount, workers := 20, 4
		var completedCount, maxConcurrent, currentConcurrent atomic.Int32
		taskDone := make(chan struct{})
		tasks := make([]Task, 0, tasksCount)

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				cur := currentConcurrent.Add(1)
				for {
					old := maxConcurrent.Load()
					if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
						break
					}
				}
				<-taskDone
				currentConcurrent.Add(-1)
				completedCount.Add(1)
				return nil
			})
		}

		done := make(chan error, 1)
		go func() { done <- Run(tasks, workers, 1) }()

		require.Eventually(t, func() bool { return maxConcurrent.Load() > 1 }, time.Second*2, time.Millisecond*10)
		close(taskDone)
		require.NoError(t, <-done)
		require.Equal(t, int32(tasksCount), completedCount.Load())
	})
}

func createSuccessTasks(count int) ([]Task, *atomic.Int32, time.Duration) {
	var runCount atomic.Int32
	var totalWait time.Duration
	tasks := make([]Task, 0, count)
	for i := 0; i < count; i++ {
		wait := time.Millisecond * time.Duration(rand.Intn(20))
		totalWait += wait
		tasks = append(tasks, func() error {
			time.Sleep(wait)
			runCount.Add(1)
			return nil
		})
	}
	return tasks, &runCount, totalWait
}
