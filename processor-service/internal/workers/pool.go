package workers

import (
	"context"
	"log"
	"sync"
)

type Job func(ctx context.Context) error

type Pool struct {
	jobs chan Job
	wg   sync.WaitGroup
}

func NewPool(workerCount int, buffer int) *Pool {
	p := &Pool{
		jobs: make(chan Job, buffer),
	}

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go func(id int) {
			defer p.wg.Done()
			for job := range p.jobs {
				if err := job(context.Background()); err != nil {
					log.Printf("worker %d: job failed: %v", id, err)
				}
			}
		}(i)
	}

	return p
}

func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}
