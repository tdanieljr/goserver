package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) // runs on every request
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Add(-cfg.fileserverHits.Load())
}
func (cfg *apiConfig) countHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func handlerHeathz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	c := apiConfig{fileserverHits: atomic.Int32{}}
	m := http.NewServeMux()

	m.Handle("/app", c.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	m.HandleFunc("GET /healthz", handlerHeathz)
	m.HandleFunc("GET /metrics", c.countHits)
	m.HandleFunc("POST /reset", c.resetHits)
	s := http.Server{Handler: m, Addr: ":8080"}

	s.ListenAndServe()

}
