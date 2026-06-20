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
	loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)
