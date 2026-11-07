package scheduler

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/codicecustode/inference-scheduler/internal/models"
	"github.com/codicecustode/inference-scheduler/internal/store"
)

type Scheduler struct {
	queue   *JobQueue
	pool    *WorkerPool
	store   store.Store
	eventCh chan *models.Event
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewScheduler(s store.Store, pool *WorkerPool, eventCh chan *models.Event) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	q := &JobQueue{}
	heap.Init(q)
	return &Scheduler{
		queue:   q,
		pool:    pool,
		store:   s,
		eventCh: eventCh,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *Scheduler) Submit(job *models.Job) error {

	if err := s.store.Save(job); err != nil {
		return err
	}
	s.mu.Lock()
	heap.Push(s.queue, job)
	s.mu.Unlock()

	s.eventCh <- &models.Event{
		Type: "job.submitted",
		Job:  job,
	}

	return nil

}

func (s *Scheduler) Start() {
    go s.dispatchLoop()
    go s.resultLoop()
}

func (s *Scheduler) Stop() {
    s.cancel()
}


func (s *Scheduler) dispatchLoop() {

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			s.mu.Lock()
			if s.queue.Len() > 0 {
				job := heap.Pop(s.queue).(*models.Job)
				s.mu.Unlock()
				s.pool.Submit(job)
			} else {
				s.mu.Unlock()
			}
		case <- s.ctx.Done():
			return
		}

	}

}

func (s *Scheduler) resultLoop() {
    for {
        select {
        case job := <-s.pool.Results():
            job.RLock()
            status := job.Status
            job.RUnlock()

            eventType := "job.completed"
            if status == models.StatusFailed {
                eventType = "job.failed"
            }

            s.eventCh <- &models.Event{
                Type: eventType,
                Job:  job,
            }

        case <-s.ctx.Done():
            return
        }
    }
}

