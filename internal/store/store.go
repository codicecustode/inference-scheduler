package store

import (
	"fmt"
	"sync"

	"github.com/codicecustode/inference-scheduler/internal/models"
)

type Store interface {
	Save(job *models.Job) error
	Get(id string) (*models.Job, error)
	List() []*models.Job
	ListByStatus(status models.JobStatus) []*models.Job
}

type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]*models.Job
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]*models.Job),
	}
}

func (ms *MemoryStore) Save(job *models.Job) error {

	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.jobs[job.ID] = job

	return nil
}

func (ms *MemoryStore) Get(jobID string) (*models.Job, error) {

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	job, ok := ms.jobs[jobID]

	if ok != true {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return job, nil

}

func (ms *MemoryStore) List() []*models.Job {

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]*models.Job, 0, len(ms.jobs))

	for _, job := range ms.jobs {
		result = append(result, job)
	}

	return result

}

func (ms *MemoryStore) ListByStatus(status models.JobStatus) []*models.Job {

	var result []*models.Job

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, job := range ms.jobs {
		job.RLock()
		if job.Status == status {

			result = append(result, job)

		}
		job.RUnlock()
	}

	return result

}
