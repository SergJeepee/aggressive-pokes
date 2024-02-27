package worker

import (
	"aggressive-pokes/internal/stats"
	"context"
	"sync"
)

const maxWorkerPool = 10000

var tasks = make(chan func(reporter stats.Reporter), maxWorkerPool)

func Submit(task func(reporter stats.Reporter)) {
	tasks <- task
}

func Cancel() {
	close(tasks)
}

func StartWorkers(ctx context.Context, reporter stats.Reporter, n int) chan struct{} {
	if float64(n) > maxWorkerPool {
		n = maxWorkerPool
	}
	wg := &sync.WaitGroup{}
	wg.Add(n)
	poolFinished := make(chan struct{})
	go func() {
		for w := 0; w < n; w++ {
			go func(w int) {
				for {
					select {
					case <-ctx.Done():
						wg.Done()
						//fmt.Printf("Worker %v done\n", w)
						return
					case runnable, ok := <-tasks:
						if !ok {
							wg.Done()
							//fmt.Printf("Worker %v: Channel closed\n", w)
							return
						}
						runnable(reporter)
					}
				}
			}(w)
		}
		wg.Wait()
		poolFinished <- struct{}{}
		close(poolFinished)
	}()
	return poolFinished
}
