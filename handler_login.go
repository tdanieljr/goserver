package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/tdanieljr/goserver/internal/auth"
	"github.com/tdanieljr/goserver/internal/database"
)

func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding user: %s", err)
		w.WriteHeader(500)
		return
	}
	dbUser, err := cfg.db.GetUserWithEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("error finding user %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}
	t, err := auth.MakeJWT(dbUser.ID, cfg.secret)
	if err != nil {
		log.Printf("error Making jwt token %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}
	rt, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("error making refresh token %s", err)
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashPassword)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	if !match {
		w.WriteHeader(401)
		w.Write([]byte("Incorrect email or password"))
		return
	}
	tokenParams := database.InsertTokenParams{
		Token:     rt,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}
	cfg.db.InsertToken(r.Context(), tokenParams)
	apiUser := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        t,
		RefreshToken: rt,
		IsRed:        dbUser.IsChirpyRed,
	}
	resp, err := json.Marshal(apiUser)
	if err != nil {
		log.Printf("error marshalling response %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(resp)

}
