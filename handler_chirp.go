package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tdanieljr/goserver/internal/auth"
	"github.com/tdanieljr/goserver/internal/database"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (cfg *apiConfig) Chirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	c := apiChirp{}
	err := decoder.Decode(&c)
	if err != nil {
		log.Printf("Error decoding chirp: %s", err)
		w.WriteHeader(500)
		return
	}
	t, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Auth token: %s", err)
		w.WriteHeader(401)
		return
	}
	tokenID, err := auth.ValidateJWT(t, cfg.secret)
	if err != nil {
		log.Printf("Error validating Auth token: %s", err)
		w.WriteHeader(401)
		return
	}

	bad_words := []string{"kerfuffle", "sharbert", "fornax"}
	if len(c.Body) <= 140 {
		for _, v := range bad_words {
			c.Body = strings.ReplaceAll(c.Body, v, "****")
			c.Body = strings.ReplaceAll(c.Body, cases.Title(language.Und).String(v), "****")
		}

		tmp := database.CreateChirpParams{Body: c.Body, UserID: tokenID}
		chirp, err := cfg.db.CreateChirp(r.Context(), tmp)
		if err != nil {
			log.Printf("Error creating chirp: %s", err)
			w.WriteHeader(500)
			return
		}

		resp := apiChirpResp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		dat, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(201)
		w.Write(dat)

	} else {
		w.WriteHeader(400)
	}
}
func (cfg *apiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		w.WriteHeader(500)
		return
	}
	apiChirps := make([]apiChirpResp, len(dbChirps))
	for idx, v := range dbChirps {
		apiChirps[idx] = apiChirpResp{
			Id:        v.ID,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Body:      v.Body,
			UserID:    v.UserID,
		}
	}
	resp, err := json.Marshal(apiChirps)
	if err != nil {
		log.Printf("Error marshalling chirps: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(resp)

}
func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error parsing chirp id:%s", err)
		w.WriteHeader(500)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		w.WriteHeader(404)
		return
	}
	chirp := apiChirpResp{
		Id:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	resp, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling chirp: %s", err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	w.Write(resp)

}

func (cfg *apiConfig) DeletChirp(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Auth header: %s", err)
		w.WriteHeader(401)
		return
	}
	userID, err := auth.ValidateJWT(authToken, cfg.secret)
	if err != nil {
		log.Printf("Error getting validating token: %s", err)
		w.WriteHeader(403)
		return
	}
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error parsing chirp id:%s", err)
		w.WriteHeader(403)
		return
	}

	dbChirp, err := cfg.db.GetChirp(r.Context(), id)
	if dbChirp.UserID != userID {
		log.Printf("User attempting to delet a Chirp not their own")
		w.WriteHeader(403)
		return

	}
	err = cfg.db.DeleteChirp(r.Context(), dbChirp.ID)
	if err != nil {
		log.Printf("Error deleting chirp id:%s", err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)
}
