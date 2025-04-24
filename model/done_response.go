package model

type DoneResponse struct {
	Type    string `json:"type"`
	Id      string `json:"id"`
	Success bool   `json:"success"`
}
