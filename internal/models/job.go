package models

import (
	"sync"
	"time"
)

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
)

type Job struct {
	mu          sync.RWMutex
	ID          string     `json:"id"`
	ModelName   string     `json:"modelName"`
	Input       string     `json:"input"`
	Priority    int        `json:"priority"`
	VRAMReq     int        `json:"vramReq"`
	Status      JobStatus  `json:"status"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	GPUID       string     `json:"gpuId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

func (j *Job) Lock() {

	j.mu.Lock()

}

func (j *Job) Unlock() {
	j.mu.Unlock()
}

func (j *Job) RLock() {
	j.mu.RLock()
}

func (j *Job) RUnlock() {

	j.mu.RUnlock()

}

type GPU struct {
	mu           sync.RWMutex
	ID           string `json:"id"`
	TotalVRAM    int    `json:"totalVRam"`
	AvailVRAM    int    `json:"availVRam"`
	CurrentJobID string `json:"currentJobId"`
}

func (g *GPU) Allocate(vram int, jobID string) bool {

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.AvailVRAM >= vram {
		g.AvailVRAM -= vram
		g.CurrentJobID = jobID
		return true
	}
	return false

}

func (g *GPU) Release(vram int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.AvailVRAM += vram

	g.CurrentJobID = ""

}

func (g *GPU) Lock() {

	g.mu.Lock()

}

func (g *GPU) Unlock() {

	g.mu.Unlock()

}

func (g *GPU) RLock() {

	g.mu.RLock()

}

func (g *GPU) RUnlock() {

	g.mu.RUnlock()

}

type SubmitRequest struct {
	ModelName string `json:"modelName"`
	Input     string `json:"input"`
	Priority  int    `json:"priority"`
	VRAMReq   int    `json:"vramReq"`
}

type Event struct {
	Type string `json:"type"`
	Job  *Job   `json:"job"`
}
