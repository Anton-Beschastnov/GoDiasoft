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
		maxErrorsCount := 0 // m=0 означает игнорировать ошибки

		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err, "при m=0 ошибки игнорируются, ошибка не возвращается")
		require.Equal(t, int32(tasksCount), runTasksCount, "все задачи должны выполниться")
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
		maxErrorsCount := -1 // отрицательное значение тоже игнорирует ошибки

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

		workersCount := 10 // воркеров больше чем задач
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
		// Выполнится не более n + m задач
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount))
	})

	t.Run("concurrency test with Eventually", func(t *testing.T) {
		// Тест проверяет конкурентность выполнения без использования time.Sleep
		// Используем require.Eventually для проверки что несколько задач выполняются одновременно

		tasksCount := 20
		tasks := make([]Task, 0, tasksCount)

		var completedCount atomic.Int32
		var maxConcurrent atomic.Int32
		var currentConcurrent atomic.Int32

		// Канал для контроля завершения задач
		taskDone := make(chan struct{}, tasksCount)

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				// Увеличиваем счетчик текущих выполняемых задач
				cur := currentConcurrent.Add(1)
				// Обновляем максимум одновременно выполняемых задач
				for {
					old := maxConcurrent.Load()
					if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
						break
					}
				}

				// Ждем пока не придет сигнал завершения
				<-taskDone

				currentConcurrent.Add(-1)
				completedCount.Add(1)
				return nil
			})
		}

		workersCount := 4
		maxErrorsCount := 1

		// Запускаем Run в горутине
		done := make(chan error, 1)
		go func() {
			done <- Run(tasks, workersCount, maxErrorsCount)
		}()

		// Ждем пока не увидим что несколько задач выполняются одновременно
		require.Eventually(t, func() bool {
			return maxConcurrent.Load() > 1
		}, time.Second*2, time.Millisecond*10, "задачи должны выполняться конкурентно")

		// Разрешаем задачам завершиться
		close(taskDone)

		// Ждем завершения Run
		err := <-done
		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), completedCount.Load())
	})

	t.Run("stop on error limit - detailed scenario", func(t *testing.T) {
		// Проверяем сценарий из задания: n=4, m=2
		// Должно выполниться не более n+m=6 задач

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
		// Выполнится не более n + m = 6 задач
		require.LessOrEqual(t, runTasksCount.Load(), int32(workersCount+maxErrorsCount),
			"выполнено больше задач чем n+m")
	})
}
