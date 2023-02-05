package data

import (
	"errors"
	"fmt"
)

// https://earthly.dev/blog/golang-errors/
var (
	ErrRecordNotFound = errors.New("record not found")
)

type ErrDupEmail struct {
	email string
}

func (e *ErrDupEmail) Error() string {
	return fmt.Sprintf("Duplicate Email %v", e.email)
}
