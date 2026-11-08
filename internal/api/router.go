package api

import (
    "net/http"

    "github.com/codicecustode/inference-scheduler/internal/stream"
)

func NewRouter(h *Handler, broker *stream.Broker, apiKey string, rateLimit int) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodPost:
            h.SubmitJob(w, r)
        case http.MethodGet:
            h.ListJobs(w, r)
        default:
            writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        }
    })

    mux.HandleFunc("/api/jobs/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            h.GetJob(w, r)
        case http.MethodDelete:
            h.CancelJob(w, r)
        default:
            writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        }
    })

    mux.HandleFunc("/api/gpus", h.GetGPUs)
    mux.HandleFunc("/api/metrics", h.GetMetrics)
    mux.Handle("/api/events", broker)

    auth := Auth(apiKey)
    rateLimiter := RateLimit(rateLimit)

    var handler http.HandlerFunc
    handler = func(w http.ResponseWriter, r *http.Request) {
        mux.ServeHTTP(w, r)
    }
    handler = Recovery(handler)
    handler = Logging(handler)
    handler = auth(handler)
    handler = rateLimiter(handler)

    return handler
}
