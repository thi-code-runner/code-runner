package errorutil

import "fmt"

func ErrorWrap(err error, message string) error {
	return fmt.Errorf("%s\n%s", message, err.Error())
}
