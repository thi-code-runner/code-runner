package model

type TestResponse struct {
	Test    *TestConfiguration `json:"test"`
	Message string             `json:"message"`
	Passed  bool               `json:"passed"`
}
