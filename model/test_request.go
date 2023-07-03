package model

type TestRequest struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
type TestData struct {
	Cmd string `json:"cmd"`

	Tests []*TestConfiguration `json:"tests"`

	Mainfilename string `json:"mainfilename"`

	Sourcefiles []*SourceFile `json:"sourcefiles"`
}

type TestConfiguration struct {
	Type  string                 `json:"type"`
	Param map[string]interface{} `json:"param"`
}
