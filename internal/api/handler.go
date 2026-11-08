package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/codicecustode/inference-scheduler/internal/models"
	"github.com/codicecustode/inference-scheduler/internal/scheduler"
	"github.com/codicecustode/inference-scheduler/internal/store"
	"github.com/google/uuid"
)

type Handler struct {
	scheduler *scheduler.Scheduler
	store     store.Store
	pool      *scheduler.WorkerPool
}

func NewHandler(s *scheduler.Scheduler, st store.Store, p *scheduler.WorkerPool) *Handler {

	return &Handler{

		scheduler: s,
		store:     st,
		pool:      p,
	}
}

func (h *Handler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.ModelName == "" {
		writeError(w, http.StatusBadRequest, "model_name is required")
		return
	}
	if req.VRAMReq <= 0 {
		req.VRAMReq = 1024 // default 1GB if not specified
	}

	job := &models.Job{
		ID:        uuid.NewString(),
		ModelName: req.ModelName,
		Input:     req.Input,
		Priority:  req.Priority,
		VRAMReq:   req.VRAMReq,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := h.scheduler.Submit(job); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to submit job")
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {

	data := h.store.List()
	writeJSON(w, http.StatusOK, data)

}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	jobId := urlParts[len(urlParts)-1]
	data,err := h.store.Get(jobId)
	if err!=nil{
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) CancelJob(w http.ResponseWriter, r *http.Request){

	urlParts := strings.Split(r.URL.Path,"/")

	jobId:=urlParts[len(urlParts)-1]

	job, err:=h.store.Get(jobId)
	
	if err!=nil{
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	job.Lock()
	job.Status = models.StatusCancelled
	job.Unlock()

	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) GetGPUs(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, h.pool.GPUs())
}

func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
    jobs := h.store.List()

    metrics := map[string]int{
        "pending":   0,
        "running":   0,
        "completed": 0,
        "failed":    0,
        "cancelled": 0,
        "total":     len(jobs),
    }

    for _, job := range jobs {
        job.RLock()
        metrics[string(job.Status)]++
        job.RUnlock()
    }

    writeJSON(w, http.StatusOK, metrics)
}


func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
