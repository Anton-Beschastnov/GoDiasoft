package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	ignoreErrors := m <= 0

	if len(tasks) == 0 {
		return nil
	}

	taskChan := make(chan Task, n)
	var wg sync.WaitGroup
	var errorsCount atomic.Int32

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	if !ignoreErrors && errorsCount.Load() > int32(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
