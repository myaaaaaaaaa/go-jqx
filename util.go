package jqx

import "fmt"

type failError string

func (f failError) Error() string { return string(f) }

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
func failif(err error, format string, args ...any) {
	if err == nil {
		return
	}

	s := fmt.Sprintf(format, args...)
	s = fmt.Sprintf("error while %s: %v", s, err)
	panic(failError(s))
}
func catch[T error](rt *error) {
	err := recover()
	switch err := err.(type) {
	case nil:
	case T:
		*rt = err
	default:
		panic(err)
	}
}
