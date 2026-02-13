package kafka

import (
	"context"
	"log"
	"sync/atomic"
)

type Pool struct {
	workers  int
	jobs     chan Job
	queueLen atomic.Int64
	ctx      context.Context
	cancel   context.CancelFunc
}

type Job func(ctx context.Context) error

func NewPool(ctx context.Context, workers int, queueSize int) *Pool {
	pCtx, cancel := context.WithCancel(ctx)
	p := &Pool{
		workers: workers,
		jobs:    make(chan Job, queueSize),
		ctx:     pCtx,
		cancel:  cancel,
	}

	for i := 0; i < workers; i++ {
		go p.runWorker()
	}
	return p
}

func (p *Pool) runWorker() {
	for job := range p.jobs {
		err := job(p.ctx) //execute job
		if err != nil {
			log.Println("-------", err)
		}
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
