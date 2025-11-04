package object

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

var (
	// ErrObjectProcessorError is the base error for object processing.
	ErrObjectProcessorError = errors.New("object processing failed")

	//ErrPathSegmentInvalidError for when a path segment is not found or not expected.
	ErrPathSegmentInvalidError = errors.New("path segment invalid")

	//ErrValueAtPathSegmentInvalidError for when a value at a path segment is not found or not expected.
	ErrValueAtPathSegmentInvalidError = errors.New("value at path segment invalid")
)

// Error for when object processing fails.
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

// Unwrap allows for error chaining with errors.Is and errors.As.
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
		str = str + fmt.Sprintf(" \nData: %+v", internal.JsonStringifyMust(e.Data))
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

func NewError(error error, functionName string, message string) *Error {
	n := new(Error)
	if error != nil {
		n.Err = fmt.Errorf("%w: %w", ErrObjectProcessorError, error)
	} else {
		n.Err = ErrObjectProcessorError
	}
	n.FunctionName = functionName
	n.Message = message
	return n
}

func mapKeyString(mapKey reflect.Value) string {
	if mapKey.Kind() == reflect.String {
		return mapKey.String()
	}
	return fmt.Sprintf("%v", internal.JsonStringifyMust(mapKey.Interface()))
}
