package handlers

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
