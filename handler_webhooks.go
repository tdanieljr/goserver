package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/tdanieljr/goserver/internal/auth"
)

func (cfg *apiConfig) UpgradeUser(w http.ResponseWriter, r *http.Request) {
	type upgradeEvent struct {
		Event string `json:"event"`
		Data  struct {
			User uuid.UUID `json:"user_id"`
		}
	}
	authKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("error getting api auth header: %s", err)
		w.WriteHeader(401)
		return
	}
	if authKey != cfg.polkaKey {
		log.Printf("authkey doesn't match")
		w.WriteHeader(401)
		return
	}

	req := upgradeEvent{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request: %s", err)
		w.WriteHeader(404)
		return
	}
	if req.Event != "user.upgraded" {
		log.Printf("Unkown request: %s", err)
		w.WriteHeader(204)
		return
	}
	err = cfg.db.SetUserRed(r.Context(), req.Data.User)
	if err != nil {
		log.Printf("User not found: %s", err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)

}
