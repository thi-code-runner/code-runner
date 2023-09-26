package model

import (
	"errors"
	"fmt"
	"strings"
)

type RunRequest struct {
	Type string  `json:"type"`
	Data RunData `json:"data"`
}

type RunData struct {
	Cmd          string        `json:"cmd"`
	Mainfilename string        `json:"mainfilename"`
	Sourcefiles  []*SourceFile `json:"sourcefiles"`
	Timeout      int           `json:"timeout"`
}

func (r *RunRequest) Validate() error {
	if r == nil {
		return errors.New("request must not be empty")
	}
	var errs []string
	if r.Type == "" {
		errs = append(errs, "$.type must not be empty")
	}
	if err := r.Data.validate(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
func (r *RunData) validate() error {
	var errs []string
	if r == nil {
		return errors.New("$.data must not be empty")
	}
	if r.Cmd == "" {
		errs = append(errs, "$.data.cmd must not be empty")
	}
	if r.Mainfilename == "" {
		errs = append(errs, "$.data.mainfilename must not be empty")
	}
	if len(r.Sourcefiles) <= 0 {
		errs = append(errs, "$.data.sourcefiles must not be empty")
	} else {
		for i, sf := range r.Sourcefiles {
			if sf.Content == "" {
				errs = append(errs, fmt.Sprintf("$.data.sourcefiles[%d].content must not be empty", i))
			}
			if sf.Filename == "" {
				errs = append(errs, fmt.Sprintf("$.data.sourcefiles[%d].filename must not be empty", i))
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
