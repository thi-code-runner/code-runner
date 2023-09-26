package model

import (
	"errors"
	"fmt"
	"strings"
)

type TestRequest struct {
	Type string   `json:"type"`
	Data TestData `json:"data"`
}
type TestData struct {
	Cmd          string               `json:"cmd"`
	Tests        []*TestConfiguration `json:"tests"`
	Mainfilename string               `json:"mainfilename"`
	Sourcefiles  []*SourceFile        `json:"sourcefiles"`
	Timeout      int                  `json:"timeout"`
}

type TestConfiguration struct {
	Type  string            `json:"type"`
	Param map[string]string `json:"param"`
}

func (r *TestRequest) Validate() error {
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
func (r *TestData) validate() error {
	var errs []string
	if r == nil {
		return errors.New("$.data must not be empty")
	}
	if r.Cmd == "" {
		errs = append(errs, "$.data.cmd must not be empty")
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
	if len(r.Tests) <= 0 {
		errs = append(errs, "$.data.tests must not be empty")
	} else {
		for i, t := range r.Tests {
			if t.Type == "" {
				errs = append(errs, fmt.Sprintf("$.data.tests[%d].type must not be empty", i))
			}
			if t.Param == nil {
				errs = append(errs, fmt.Sprintf("$.data.tests[%d].param must not be empty", i))
			}
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
