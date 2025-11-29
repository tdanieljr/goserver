package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
	"github.com/tdanieljr/goserver/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}
type apiChirpResp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
type apiChirp struct {
	Body string `json:"body"`
	//UserID uuid.UUID `json:"user_id"`
}
type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) // runs on every request
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Add(-cfg.fileserverHits.Load())
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
		return
	}
	cfg.db.Reset(r.Context())
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

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	dbQueries := database.New(db)
	c := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries, platform: platform, secret: secret}
	m := http.NewServeMux()

	m.Handle("/app/", c.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	m.HandleFunc("GET /api/healthz", handlerHeathz)
	m.HandleFunc("POST /admin/reset", c.resetHits)
	m.HandleFunc("GET /admin/metrics", c.countHits)

	m.HandleFunc("POST /api/chirps", c.Chirp)
	m.HandleFunc("GET /api/chirps", c.GetChirps)
	m.HandleFunc("GET /api/chirps/{chirpID}", c.GetChirp)
	m.HandleFunc("DELETE /api/chirps/{chirpID}", c.DeletChirp)

	m.HandleFunc("POST /api/login", c.Login)
	m.HandleFunc("POST /api/users", c.createUser)
	m.HandleFunc("POST /api/refresh", c.Refresh)
	m.HandleFunc("POST /api/revoke", c.Revoke)
	m.HandleFunc("PUT /api/users", c.UpdateEmail)

	s := http.Server{Handler: m, Addr: ":8080"}

	s.ListenAndServe()

}
