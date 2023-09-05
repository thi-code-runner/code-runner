package errorutil

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

var TimeoutErr = errors.New("timeout error")

func ErrorWrap(err error, message string) error {
	return fmt.Errorf("%s\n\t%s", message, err.Error())
}

func ErrorSlug() error {
	return fmt.Errorf("ERRCODE: " + strings.ToUpper(strings.Split(uuid.New().String(), "-")[0]))
}
