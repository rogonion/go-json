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

// NewError creates a new core.Error with the default base error ErrObjectError.
func NewError() *core.Error {
	n := core.NewError().WithDefaultBaseError(ErrObjectError)
	return n
}

// mapKeyString returns the string representation of a map key.
// If the key is already a string, it returns it directly.
// Otherwise, it uses JSON stringification to ensure a consistent string representation.
func mapKeyString(mapKey reflect.Value) string {
	if mapKey.Kind() == reflect.String {
		return mapKey.String()
	}
	return fmt.Sprintf("%v", core.JsonStringifyMust(mapKey.Interface()))
}
