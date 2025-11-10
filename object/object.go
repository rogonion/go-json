package object

import (
	"reflect"

	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

func (n *Object) SetSchema(value schema.Schema) {
	n.schema = value
}

func (n *Object) WithSchema(value schema.Schema) *Object {
	n.schema = value
	return n
}

// GetSource call method when done working with the source.
func (n *Object) GetSource() any {
	return n.source.Interface()
}

func (n *Object) SetSource(value any) {
	n.source = reflect.ValueOf(value)
	n.sourceType = reflect.TypeOf(value)
}

func (n *Object) WithDefaultConverter(value schema.DefaultConverter) *Object {
	n.defaultConverter = value
	return n
}

func (n *Object) SetDefaultConverter(value schema.DefaultConverter) {
	n.defaultConverter = value
}

func NewObject(source any) *Object {
	n := new(Object)
	n.SetSource(source)
	n.defaultConverter = schema.NewConversion()
	return n
}

/*
Object is the module for manipulating a source object.

Usage:
 1. Instantiate using NewObject.
 2. Set required parameters.
 3. Manipulate source using Object.Get, Object.Set, Object.Delete, or Object.ForEach.
 4. Get modified source using Object.GetSource.

Example:

	type Address struct {
		Street  string
		City    string
		ZipCode *string
	}

	source := map[string]any{
		"data": map[string]any{
			"metadata": struct {
				Address Address
				Status  string
			}{
				Address: Address{
					Street: "123 Main St",
					City:   "Anytown",
				},
				Status: "active",
			},
		},
	}

	objManip := NewObject().WithSource(source)

	valueFound, ok, err := objManip.Get("$.data.metadata.Address.City")

	noOfModifications, err := objManip.Set("$.data.metadata.Status", "inactive")

	noOfModifications, err = objManip.Delete("$.data.metadata.Status")

	// retrieve modified source after Set/Delete
	var modifiedSource any = objManip.GetSource()
*/
type Object struct {
	// Used by ForEach.
	ifValueFoundInObject IfValueFoundInObject

	// Useful especially with the Set method for creating new nested objects when starting with an empty source.
	//
	// Initialize with SetSchema or WithSchema.
	schema schema.Schema

	// Value to set in source by the Set method.
	valueToSet any

	// Made by Set and Delete.
	noOfModifications uint64

	// Last error encountered when processing the source especially for the recursive descent pattern or union pattern in path.JSONPath.
	lastError error

	// Root object to work with.
	//
	// Will be modified with Set and Delete.
	//
	// Initialize with NewObject parameter, or SetSource.
	source reflect.Value
	// Computed when you use SetSource.
	sourceType reflect.Type

	recursiveDescentSegments path.RecursiveDescentSegments

	// Default converter to use when converting data e.g., valueToSet to the destination type at the path.JSONPath.
	//
	// Initialize with WithDefaultConverter or SetDefaultConverter.
	defaultConverter schema.DefaultConverter
}

/*
IfValueFoundInObject is called when value is found at path.JSONPath.

Parameters:
  - jsonPath - Current jsonPath where value was found.
  - value - value found.

Return `true` to terminate ForEach loop.
*/
type IfValueFoundInObject func(jsonPath path.RecursiveDescentSegment, value any) bool
