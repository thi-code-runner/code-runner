package model

type TestResponse struct {
	Test    TestConfiguration `json:"test"`
	Stderr  string            `json:"stderr"`
	Message string            `json:"message"`
	Passed  bool              `json:"passed"`
}
