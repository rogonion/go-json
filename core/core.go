package core

import (
	"encoding/json"
	"errors"
	"fmt"
)

/*
Error is the default base error for the json package.
*/
type Error struct {
	Err          error
	FunctionName string
	Message      string
	Data         JsonObject

	defaultBaseError error
}

func (e *Error) SetDefaultBaseError(value error) {
	e.defaultBaseError = value
}

func (e *Error) WithDefaultBaseError(value error) *Error {
	e.SetDefaultBaseError(value)
	return e
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
	if e.Data != nil {
		str = str + " \nData: " + e.Data.String()
	}
	return str
}

func (e *Error) WithData(value JsonObject) *Error {
	e.Data = value
	return e
}

func (e *Error) WithNestedError(value error) *Error {
	e.Err = fmt.Errorf("%w: %w", e.defaultBaseError, value)
	return e
}

func (e *Error) WithFunctionName(value string) *Error {
	e.FunctionName = value
	return e
}

func (e *Error) WithMessage(value string) *Error {
	e.Message = value
	return e
}

func NewError() *Error {
	n := new(Error)
	n.defaultBaseError = errors.New("json error")
	return n
}

/*
JsonObject represents a generic JSON object structure.

This would be a map where keys are strings and values are of any type in Go.
*/
type JsonObject map[string]any

func (f JsonObject) String() string {
	if j, err := json.Marshal(f); err != nil {
		return fmt.Sprintf("<Marshalling error: %v>", err)
	} else {
		return string(j)
	}
}

/*
JsonArray represents a generic JSON array structure.

This would be a slice whose elements are of type any in Go.
*/
type JsonArray []any

func (f JsonArray) String() string {
	if j, err := json.Marshal(f); err != nil {
		return fmt.Sprintf("<Marshalling error: %v>", err)
	} else {
		return string(j)
	}
}
