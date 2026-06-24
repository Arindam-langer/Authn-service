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

func encode(w http.ResponseWriter, response any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error in handling response %v", err)
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, req *http.Request) {
	res := health{Message: "all good", Code: 200}
	encode(w, res, http.StatusOK)
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

		encode(w, res, http.StatusNotFound)

		return
	}
	// if the above condition does not work then we are gucci so we just generate the token and return in header
	token, err := auth.CreateToken(body.Username, body.Password)
	if err != nil {
		log.Printf("error in generating Token %v", err)
	}

	res := loginResponse{Code: 201}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	encode(w, res, http.StatusCreated)
}

func (h *Handler) VerifyToken(w http.ResponseWriter, req *http.Request) {
	if req.Header["Authorization"] == nil {
		res := struct {
			Message string `json:"message"`
		}{
			"No token",
		}
		encode(w, res, http.StatusUnauthorized)
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
		encode(w, res, http.StatusForbidden)
	} else {
		res := struct {
			Message string `json:"message"`
		}{"token verified"}
		encode(w, res, http.StatusOK)

	}
}
