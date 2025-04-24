package model

import (
	"errors"
	"strings"
)

type Request struct {
	Id   string                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func (r *Request) Validate() error {
	if r == nil {
		return errors.New("empty request")
	}
	var errs []string
	if r.Type == "" {
		errs = append(errs, "$.type must not be empty")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
