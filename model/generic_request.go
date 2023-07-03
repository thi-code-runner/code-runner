package model

type GenericRequest struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
