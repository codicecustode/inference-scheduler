package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/codicecustode/inference-scheduler/internal/models"
)

type WorkerPool struct {
	gpus    []*models.GPU
	jobs    chan *models.Job
	results chan *models.Job
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewWorkerPool(numGPUs, vramPerGPU int) *WorkerPool {

	ctx, cancel := context.WithCancel(context.Background())

	gpus := make([]*models.GPU, numGPUs)

	for i := 0; i < numGPUs; i++ {
		gpus[i] = &models.GPU{
			ID:        fmt.Sprintf("gpu-%d", i),
			TotalVRAM: vramPerGPU,
			AvailVRAM: vramPerGPU,
		}
	}

	return &WorkerPool{

		gpus:    gpus,
		jobs:    make(chan *models.Job, 100),
		results: make(chan *models.Job, 100),
		ctx:     ctx,
		cancel:  cancel,
	}

}

func (wp *WorkerPool) Start() {

	for _, gpu := range wp.gpus {
		go wp.worker(gpu)
	}

}

func (wp *WorkerPool) worker(gpu *models.GPU) {

	for {
		select {
		case job := <-wp.jobs:
			wp.execute(job, gpu)

		case <-wp.ctx.Done():
			return
		}
	}

}

func (wp *WorkerPool) execute(job *models.Job, gpu *models.GPU) {
	now := time.Now()
	job.Lock()
	job.Status = models.StatusRunning
	job.StartedAt = &now
	job.GPUID = gpu.ID
	job.Unlock()
	
	if !gpu.Allocate(job.VRAMReq, job.ID) {
		job.Lock()
		job.Status = models.StatusFailed
		job.Error = "insufficient VRAM"
		job.Unlock()
		wp.results <- job
		return
	}
	duration := time.Duration(job.VRAMReq/1024) * 100 * time.Millisecond
	time.Sleep(duration)

	done := time.Now()
	job.Lock()
	job.Status = models.StatusCompleted
	job.CompletedAt = &done
	job.Result = fmt.Sprintf("inference result for model %s", job.ModelName)
	job.Unlock()

	gpu.Release(job.VRAMReq)

	wp.results <- job
}


func (wp *WorkerPool) Submit(job *models.Job) {
    wp.jobs <- job
}

func (wp *WorkerPool) Results() <-chan *models.Job {
    return wp.results
}

func (wp *WorkerPool) GPUs() []*models.GPU {
    return wp.gpus
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
}

