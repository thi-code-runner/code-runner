package model

type ErrorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}
