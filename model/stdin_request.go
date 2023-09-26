package model

import (
	"errors"
	"strings"
)

type StdinRequest struct {
	Stdin string `json:"stdin"`
}

func (r *StdinRequest) Validate() error {
	var errs []string
	if r == nil {
		return errors.New("request must not be empty")
	}
	if r.Stdin == "" {
		errs = append(errs, "$.stdin must not be empty")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
