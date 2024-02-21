package core

import "fmt"

var (
	ErrNotFound error = fmt.Errorf("resource not found")
)
