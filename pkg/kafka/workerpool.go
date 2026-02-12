package kafka

import (
	"context"
	"sync/atomic"
)

type Pool struct {
	workers  int
	jobs     chan Job
	queueLen atomic.Int64
}

type Job func(ctx context.Context) error

func NewPool(workers int, queueSize int) *Pool {
	p := &Pool{
		workers: workers,
		jobs:    make(chan Job, queueSize),
	}

	for i := 0; i < workers; i++ {
		go p.runWorker()
	}
	return p
}

func (p *Pool) runWorker() {
	for job := range p.jobs {
		_ = job(context.Background()) //execute job
		p.queueLen.Add(-1)
	}
}

func (p *Pool) Submit(job Job) {
	p.queueLen.Add(1)
	p.jobs <- job
}

func (p *Pool) QueueLen() int {
	return int(p.queueLen.Load())
}

func (p *Pool) Capacity() int {
	return cap(p.jobs)
}

func (p *Pool) Shutdown() {
	close(p.jobs)
}
