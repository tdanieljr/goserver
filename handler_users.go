package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/tdanieljr/goserver/internal/auth"
	"github.com/tdanieljr/goserver/internal/database"
)

func (cfg *apiConfig) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Auth header: %s", err)
		w.WriteHeader(401)
		return
	}
	userID, err := auth.ValidateJWT(authToken, cfg.secret)
	if err != nil {
		log.Printf("Error getting validating token: %s", err)
		w.WriteHeader(401)
		return
	}

	type emailUpdate struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	dat := emailUpdate{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&dat)
	if err != nil {
		log.Printf("Error decoding json: %s", err)
		w.WriteHeader(401)
		return
	}
	hashPassword, err := auth.HashPassword(dat.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(401)
		return
	}
	userUpdate := database.UpdateUserEmailParams{
		ID:           userID,
		Email:        dat.Email,
		HashPassword: hashPassword,
	}
	err = cfg.db.UpdateUserEmail(r.Context(), userUpdate)
	if err != nil {
		log.Printf("Error updating email: %s", err)
		w.WriteHeader(401)
		return
	}
	dbUser, err := cfg.db.GetUserWithEmail(r.Context(), dat.Email)
	apiUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	resp, err := json.Marshal(apiUser)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(401)
		return
	}
	w.WriteHeader(200)
	w.Write(resp)

}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding user: %s", err)
		w.WriteHeader(500)
		return
	}
	hashpassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("error with password: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashPassword: hashpassword})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500)
		return
	}
	apiUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	dat, err := json.Marshal(apiUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Write(dat)

}

func (cfg *apiConfig) Refresh(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Auth token: %s", err)
		w.WriteHeader(500)
		return
	}
	user, err := cfg.db.GetUserFromToken(r.Context(), rt)
	if err != nil {
		log.Printf("Error getting user: %s", err)
		w.WriteHeader(500)
		return
	}
	if user.RevokedAt.Valid {
		log.Printf("Revoked Token")
		w.WriteHeader(401)
		return
	}
	t, err := auth.MakeJWT(user.UserID, cfg.secret)
	if err != nil {
		log.Printf("Error making new token: %s", err)
		w.WriteHeader(500)
		return
	}
	type tokenResponse struct {
		T string `json:"token"`
	}
	resp := tokenResponse{T: t}

	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}
func (cfg *apiConfig) Revoke(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting Auth token: %s", err)
		w.WriteHeader(500)
		return
	}
	err = cfg.db.RevokeToken(r.Context(), rt)
	if err != nil {
		log.Printf("Error revoking token: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(204)
}
