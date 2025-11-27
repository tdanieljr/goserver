package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	w.Write(fmt.Appendf(nil, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load()))

}

func handlerHeathz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		C string `json:"body"`
	}
	type valid struct {
		CleanedChirp string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(r.Body)
	c := chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		log.Printf("Error decoding chirp: %s", err)
		w.WriteHeader(500)
		return
	}
	bad_words := []string{"kerfuffle", "sharbert", "fornax"}
	if len(c.C) <= 140 {
		w.WriteHeader(200)
		for _, v := range bad_words {
			c.C = strings.ReplaceAll(c.C, v, "****")
			c.C = strings.ReplaceAll(c.C, cases.Title(language.Und).String(v), "****")
		}
		resp := valid{CleanedChirp: c.C}

		dat, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
	} else {
		w.WriteHeader(400)
	}

}

func main() {
	c := apiConfig{fileserverHits: atomic.Int32{}}
	m := http.NewServeMux()

	m.Handle("/app/", c.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	m.HandleFunc("GET /api/healthz", handlerHeathz)
	m.HandleFunc("GET /admin/metrics", c.countHits)
	m.HandleFunc("POST /admin/reset", c.resetHits)
	m.HandleFunc("POST /api/validate_chirp", validateChirp)
	s := http.Server{Handler: m, Addr: ":8080"}

	s.ListenAndServe()

}
