package api

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	}
}

func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next(w, r)

	}
}

func Auth(apiKey string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            key := r.Header.Get("X-API-Key")
            if key != apiKey {
                writeError(w, http.StatusUnauthorized, "invalid api key")
                return
            }
            next(w, r)
        }
    }
}

func RateLimit(rps int) func(http.HandlerFunc) http.HandlerFunc {
    bucket := make(chan struct{}, rps)

    for i := 0; i < rps; i++ {
        bucket <- struct{}{}
    }

    go func() {
        ticker := time.NewTicker(time.Second / time.Duration(rps))
        defer ticker.Stop()
        for range ticker.C {
            select {
            case bucket <- struct{}{}:
            default:
            }
        }
    }()

    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            select {
            case <-bucket:
                next(w, r)
            default:
                writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
            }
        }
    }
}


