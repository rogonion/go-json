package schemapath

import (
	"errors"
	"fmt"

	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

// SchemaPath represents a JSONPath both in its parsed and unparsed form.
//
// Ideally functions are expected to only work with a JSONPath with no recursive descent searches.
type SchemaPath interface {
	path.JSONPath | path.RecursiveDescentSegment | path.RecursiveDescentSegments
}

var (
	// ErrSchemaPathError is the base error for GetSchemaAtPath.
	ErrSchemaPathError = errors.New("schema path error")
)

type Error struct {
	Err          error
	FunctionName string
	Message      string
	PathSegments path.RecursiveDescentSegment
	Schema       schema.Schema
}

func NewError(error error, functionName string, message string, schema schema.Schema, pathSegments path.RecursiveDescentSegment) *Error {
	n := new(Error)
	if error != nil {
		n.Err = fmt.Errorf("%w: %w", ErrSchemaPathError, error)
	} else {
		n.Err = ErrSchemaPathError
	}
	n.FunctionName = functionName
	n.Message = message
	n.PathSegments = pathSegments
	n.Schema = schema
	return n
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
	if e.Schema != nil {
		str = str + " \nSchema: " + e.Schema.String()
	}
	return str
}
