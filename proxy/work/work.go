package work

import "context"

type WorkerPool struct {
	workers int
	c       chan Work
}

type Work struct {
	f    func()
	done func()
}

func (q *WorkerPool) Add(f func(), done func()) {
	q.c <- Work{
		f:    f,
		done: done,
	}
}

func (q *WorkerPool) AddWait(f func()) {
	ctx, cancel := context.WithCancel(context.Background())
	q.Add(f, cancel)
	<-ctx.Done()
}

func (q *WorkerPool) Start() {
	for range q.workers {
		go worker(q.c)
	}
}

func (q *WorkerPool) Stop() {
	close(q.c) // close the channel to stop workers (recieve from a closed chann)
}

func worker(c <-chan Work) {
	for work := range c {
		work.f()
		if work.done != nil {
			work.done()
		}
	}
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		c:       make(chan Work, 1000),
	}
}
