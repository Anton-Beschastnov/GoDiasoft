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

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("m <= 0 means ignore errors", func(t *testing.T) {
		tasksCount := 20
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var errorsCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				atomic.AddInt32(&errorsCount, 1)
				return fmt.Errorf("error")
			})
		}

		workersCount := 5
		maxErrorsCount := 0

		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount)
	})

	t.Run("m = -1 means ignore errors", func(t *testing.T) {
		tasksCount := 10
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				return fmt.Errorf("error")
			})
		}

		workersCount := 3
		maxErrorsCount := -1

		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount)
	})

	t.Run("empty tasks list", func(t *testing.T) {
		tasks := make([]Task, 0)
		err := Run(tasks, 5, 3)
		require.NoError(t, err)
	})

	t.Run("n greater than tasks count", func(t *testing.T) {
		tasksCount := 3
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 10
		err := Run(tasks, workersCount, 1)

		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount)
	})

	t.Run("all tasks return errors with low m limit", func(t *testing.T) {
		tasksCount := 100
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				return fmt.Errorf("error")
			})
		}

		workersCount := 5
		maxErrorsCount := 3

		err := Run(tasks, workersCount, maxErrorsCount)

		require.True(t, errors.Is(err, ErrErrorsLimitExceeded))
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount))
	})

	t.Run("concurrency test with Eventually", func(t *testing.T) {
		tasksCount := 20
		tasks := make([]Task, 0, tasksCount)

		var completedCount atomic.Int32
		var maxConcurrent atomic.Int32
		var currentConcurrent atomic.Int32

		taskDone := make(chan struct{}, tasksCount)

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

		workersCount := 4
		maxErrorsCount := 1

		done := make(chan error, 1)
		go func() {
			done <- Run(tasks, workersCount, maxErrorsCount)
		}()

		require.Eventually(t, func() bool {
			return maxConcurrent.Load() > 1
		}, time.Second*2, time.Millisecond*10)

		close(taskDone)

		err := <-done
		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), completedCount.Load())
	})

	t.Run("stop on error limit - detailed scenario", func(t *testing.T) {
		tasksCount := 20
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount atomic.Int32
		var errorsCount atomic.Int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				runTasksCount.Add(1)
				errorsCount.Add(1)
				return fmt.Errorf("error")
			})
		}

		workersCount := 4
		maxErrorsCount := 2

		err := Run(tasks, workersCount, maxErrorsCount)

		require.True(t, errors.Is(err, ErrErrorsLimitExceeded))
		require.LessOrEqual(t, runTasksCount.Load(), int32(workersCount+maxErrorsCount))
	})
}
