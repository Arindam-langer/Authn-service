package handlers

import (
	"errors"
	"strings"
)

type (
	statusCode int
	health     struct {
		Message string     `json:"message"`
		Code    statusCode `json:"code"`
	}
	loginResponse struct {
		Code statusCode `json:"status"`
	}
	signUpRequest struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}
	loginRequest struct {
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}
)

func (r *signUpRequest) Validate() error {
	if r.Username == "" {
		return errors.New("username is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	atIdx := strings.Index(r.Email, "@")
	if atIdx == -1 || atIdx == 0 || atIdx == len(r.Email)-1 {
		return errors.New("invalid email format: must contain @ followed by domain")
	}
	if r.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (r *loginRequest) Validate() error {
	if r.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}
