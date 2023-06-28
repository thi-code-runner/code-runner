package model

type RunRequest struct {
	Type string  `json:"type"`
	Data RunData `json:"data"`
}
type RunData struct {
	Cmd string `json:"cmd"`

	Mainfilename string `json:"mainfilename"`

	Sourcefiles []*SourceFile `json:"sourcefiles"`
}
