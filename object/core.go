package object

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/path"
)

var (
	// ErrObjectError is the default error.
	ErrObjectError = errors.New("object processing failed")

	//ErrPathSegmentInvalidError for when a path segment is not found or not expected.
	ErrPathSegmentInvalidError = errors.New("path segment invalid")

	//ErrValueAtPathSegmentInvalidError for when a value at a path segment is not found or not expected.
	ErrValueAtPathSegmentInvalidError = errors.New("value at path segment invalid")
)

/*
Error for when object processing fails.
*/
type Error struct {
	Err          error
	FunctionName string
	Message      string
	PathSegments path.RecursiveDescentSegment
	Data         interface{}
}

func (e *Error) Error() string {
	var err error
	if e.Message != "" {
		err = errors.New(e.Message)
	}
	return fmt.Errorf("%w: %w", err, e.Err).Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) String() string {
	str := e.Error()
	if e.FunctionName != "" {
		str = str + " \nFunctionName: " + e.FunctionName
	}
	str = str + "\nMessage: " + e.Message
	if e.PathSegments != nil {
		str = str + " \nPathSegments: " + e.PathSegments.String()
	}

	if e.Data != nil {
		str = str + fmt.Sprintf(" \nData: %+v", core.JsonStringifyMust(e.Data))
	}
	return str
}

func (e *Error) WithPathSegments(pathSegments path.RecursiveDescentSegment) *Error {
	e.PathSegments = pathSegments
	return e
}

func (e *Error) WithData(data interface{}) *Error {
	e.Data = data
	return e
}

func (e *Error) WithNestedError(value error) *Error {
	e.Err = fmt.Errorf("%w: %w", ErrObjectError, value)
	return e
}

func NewError(functionName string, message string) *Error {
	n := new(Error)
	n.Err = ErrObjectError
	n.FunctionName = functionName
	n.Message = message
	return n
}

func mapKeyString(mapKey reflect.Value) string {
	if mapKey.Kind() == reflect.String {
		return mapKey.String()
	}
	return fmt.Sprintf("%v", core.JsonStringifyMust(mapKey.Interface()))
}
