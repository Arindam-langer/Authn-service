package handlers

type (
	statusCode int
	health     struct {
		Message string     `json:"Message"`
		Code    statusCode `json:"Code"`
	}
)
