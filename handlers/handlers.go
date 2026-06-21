// Package handlers:  functions for governance come here
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Arindam-langer/governance-service/internal/auth"
	"github.com/Arindam-langer/governance-service/internal/db"
	"github.com/golang-jwt/jwt"
)

type Handler struct {
	store db.UserStore
}

func New(store db.UserStore) *Handler {
	return &Handler{store: store}
}

// maybe we can use an interface here
func response(w http.ResponseWriter, response any) {
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error in handling response %v", err)
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, req *http.Request) {
	res := health{Message: "all good", Code: 200}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response(w, res)
}

func (h *Handler) SignIn(w http.ResponseWriter, req *http.Request) {
	var body loginRequest
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		log.Printf("error in decoding %v", err)
	}
	// this would be a DB call where we call db to check if this user exists or not for now it is going to be hard coded
	if body.Username != "aru" {
		res := struct {
			Message string `json:"message"`
		}{
			"user not found",
		}

		w.WriteHeader(http.StatusNotFound)
		response(w, res)
		return
	}
	// if the above condition does not work then we are gucci so we just generate the token and return in header
	token, err := auth.CreateToken(body.Username, body.Password)
	if err != nil {
		log.Printf("error in generating Token %v", err)
	}

	res := loginResponse{Code: 201}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusCreated)
	response(w, res)
}

func (h *Handler) VerifyToken(w http.ResponseWriter, req *http.Request) {
	if req.Header["Authorization"] == nil {
		res := struct {
			Message string `json:"message"`
		}{
			"No token",
		}
		response(w, res)
	}
	headerToken := strings.TrimPrefix(req.Header["Authorization"][0], "Bearer ")
	token, err := jwt.Parse(headerToken, auth.IsValidToken)
	if err != nil {
		log.Printf("error in token parsing %v", err)
	}
	if !token.Valid {
		log.Printf("Not a valid Token")
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	username := claims["username"]
	if username != "aru" {
		res := struct {
			Message string `json:"message"`
		}{
			"unverfied",
		}
		response(w, res)
	} else {
		res := struct {
			Message string `json:"message"`
		}{"token verified"}
		response(w, res)

	}
}
