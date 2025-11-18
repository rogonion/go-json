package object

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/core"
)

var (
	// ErrObjectError is the default error.
	ErrObjectError = errors.New("object processing failed")

	//ErrPathSegmentInvalidError for when a path segment is not found or not expected.
	ErrPathSegmentInvalidError = errors.New("path segment invalid")

	//ErrValueAtPathSegmentInvalidError for when a value at a path segment is not found or not expected.
	ErrValueAtPathSegmentInvalidError = errors.New("value at path segment invalid")
)

func NewError() *core.Error {
	n := core.NewError().WithDefaultBaseError(ErrObjectError)
	return n
}

func mapKeyString(mapKey reflect.Value) string {
	if mapKey.Kind() == reflect.String {
		return mapKey.String()
	}
	return fmt.Sprintf("%v", core.JsonStringifyMust(mapKey.Interface()))
}
