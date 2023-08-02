package model

type Request struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
