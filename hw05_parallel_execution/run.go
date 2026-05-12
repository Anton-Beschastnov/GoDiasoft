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
	if len(tasks) == 0 {
		return nil
	}

	ignoreErrors := m <= 0
	taskChan := make(chan Task, n)
	var wg sync.WaitGroup
	var errorsCount atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, taskChan, &errorsCount, m, ignoreErrors, cancel)
		}()
	}

	return feeder(ctx, tasks, taskChan, &wg, &errorsCount, m, ignoreErrors)
}

func worker(
	ctx context.Context, tasks <-chan Task, errCnt *atomic.Int32,
	m int, ignore bool, cancel context.CancelFunc,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok {
				return
			}
			if err := task(); err != nil && !ignore {
				if errCnt.Add(1) > int32(m) {
					cancel()
					return
				}
			}
		}
	}
}

func feeder(
	ctx context.Context, tasks []Task, ch chan Task,
	wg *sync.WaitGroup, errCnt *atomic.Int32, m int, ignore bool,
) error {
	var sendErr error
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			sendErr = ErrErrorsLimitExceeded
		case ch <- task:
		}
		if sendErr != nil {
			break
		}
	}
	close(ch)
	wg.Wait()

	if sendErr != nil || (!ignore && errCnt.Load() > int32(m)) {
		return ErrErrorsLimitExceeded
	}
	return nil
}
