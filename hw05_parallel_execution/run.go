package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	// Если m <= 0, считаем что ошибки игнорируются (бесконечный лимит)
	ignoreErrors := m <= 0

	if len(tasks) == 0 {
		return nil
	}

	// Канал для передачи задач воркерам (буферизированный для избежания deadlock)
	taskChan := make(chan Task, n)

	// WaitGroup для ожидания завершения всех воркеров
	var wg sync.WaitGroup

	// Счетчик ошибок
	var errorsCount atomic.Int32

	// Контекст для отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем n воркеров
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskChan:
					if !ok {
						return
					}
					if err := task(); err != nil {
						if !ignoreErrors {
							newCount := errorsCount.Add(1)
							if newCount > int32(m) {
								cancel()
								return
							}
						}
					}
				}
			}
		}()
	}

	// Отправляем задачи воркерам
	sendErr := error(nil)
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			sendErr = ErrErrorsLimitExceeded
			break
		case taskChan <- task:
		}
		if sendErr != nil {
			break
		}
	}

	close(taskChan)
	wg.Wait()

	if sendErr != nil {
		return sendErr
	}

	// Проверяем счетчик ошибок после завершения всех задач
	if !ignoreErrors && errorsCount.Load() > int32(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
