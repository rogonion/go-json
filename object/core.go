package object

import (
	"errors"
	"fmt"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

var (
	// ErrObjectProcessorError is the base error for DataProcessor interface.
	ErrObjectProcessorError = errors.New("object processing failed")

	//ErrPathSegmentInvalidError for when a path segment is not found or not expected.
	ErrPathSegmentInvalidError = errors.New("path segment invalid")

	//ErrValueAtPathSegmentInvalidError for when a value at a path segment is not found or not expected.
	ErrValueAtPathSegmentInvalidError = errors.New("value at path segment invalid")
)

// Error for when DataProcessor validation execution fails.
type Error struct {
	Err          error
	FunctionName string
	Message      string
	PathSegments path.RecursiveDescentSegment
	Data         interface{}
}

func NewError(error error, functionName string, message string, data interface{}, pathSegments path.RecursiveDescentSegment) *Error {
	n := new(Error)
	if error != nil {
		n.Err = fmt.Errorf("%w: %w", ErrObjectProcessorError, error)
	} else {
		n.Err = ErrObjectProcessorError
	}
	n.FunctionName = functionName
	n.Message = message
	n.Data = data
	n.PathSegments = pathSegments
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

	if e.Data != nil {
		str = str + fmt.Sprintf(" \nData: %+v", internal.JsonStringifyMust(e.Data))
	}
	return str
}

// DataProcessor for performing specific tasks against data whose structure is JSON-like.
type DataProcessor interface {
	// Get Retrieves values in root based on path.
	//
	// For path with recursive descent, expect a slice of values found.
	//
	// Parameters:
	//	- root - source value.
	//	- path - path based on Path to location in root to retrieve value.
	//
	// Returns:
	//	- Value found.
	//	- Boolean to indicate if the search for the value was completely successful.
	//	- The last error encountered when searching for the value.
	Get(root any, path path.JSONPath) (any, bool, error)

	// Set Inserts value in root based on path.
	//
	// Parameters:
	//	- root - source value.
	//	- path - path based on Path to location in root to insert value.
	//	- value - new value to insert.
	//
	// Returns:
	//	- Value found.
	//	- unsigned integer indicating the number of modifications successfully made through insertion.
	//	- The last error encountered during insertion.
	Set(root any, path path.JSONPath, value any) (any, uint64, error)

	// Delete Deletes value in root based on path.
	//
	// Parameters:
	//	- root - source value.
	//	- path - path based on Path to location in root to delete value.
	//
	// Returns:
	//	- Value found.
	//	- unsigned integer indicating the number of modifications successfully made through deletion.
	//	- The last error encountered during deletion.
	Delete(root any, path path.JSONPath) (any, uint64, error)

	// ForEach Calls ifValueFoundInObject if a value is found at path.
	//
	// Parameters:
	//   - root - source value.
	//   - path - path based on Path to location in root to retrieve value.
	//   - ifValueFoundInObject - ifValueFoundInObject Called when value has been found in object. Return true to terminate loop.
	ForEach(root any, path path.JSONPath, ifValueFoundInObject IfValueFoundInObject)

	// AreEqual Recursively checks if left and right are equal
	//
	// Actively checks if elements of  slices, arrays, maps, and/or structs are equal and defaults to reflect.DeepEqual for the remaining checks.
	//
	//	May only panic if reflect functions panic though measures have been set to ensure they are called appropriately.
	//
	// Parameters:
	// 	- left - Value to check.
	//	- right - Value to check.
	//
	// Returns true if left and right are equal.
	AreEqual(left any, right any) bool
}
