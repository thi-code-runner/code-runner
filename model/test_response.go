package model

type TestResponse struct {
	Type string              `json:"type"`
	Data []*TestResponseData `json:"data"`
}

type TestResponseData struct {
	Test    *TestConfiguration `json:"test"`
	Message string             `json:"message"`
	Passed  bool               `json:"passed"`
	Detail  []*Detail          `json:"detail"`
}
type Detail struct {
	Name    string `json:"name"`
	Class   string `json:"class"`
	Message string `json:"message"`
}
