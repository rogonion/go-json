package object

import (
	"reflect"

	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

// SetSchema sets the schema used for structural definitions (e.g. during Set operations).
func (n *Object) SetSchema(value schema.Schema) {
	n.schema = value
}

// WithSchema is a chainable variant of SetSchema.
func (n *Object) WithSchema(value schema.Schema) *Object {
	n.schema = value
	return n
}

/*
GetValueFoundInterface retrieve value found in Go form after Get.

If Object.valueFound.IsValid returns `false`, it will return nil.
*/
func (n *Object) GetValueFoundInterface() any {
	if n.valueFound.IsValid() {
		return n.valueFound.Interface()
	}
	return nil
}

/*
GetValueFoundReflected retrieve value found in reflect form after Get.
*/
func (n *Object) GetValueFoundReflected() reflect.Value {
	return n.valueFound
}

/*
GetSourceInterface if you want the current state of source in its interface form.
*/
func (n *Object) GetSourceInterface() any {
	if !n.source.IsValid() {
		return nil
	}
	return n.source.Interface()
}

/*
GetSourceReflected if you want the current source in its reflect form.
*/
func (n *Object) GetSourceReflected() reflect.Value {
	return n.source
}

// SetSourceInterface sets the source object to work with from an interface{}.
func (n *Object) SetSourceInterface(value any) {
	n.source = reflect.ValueOf(value)
	n.sourceType = reflect.TypeOf(value)
}

// WithSourceInterface is a chainable variant of SetSourceInterface.
func (n *Object) WithSourceInterface(value any) *Object {
	n.SetSourceInterface(value)
	return n
}

// SetSourceReflected sets the source object to work with from a reflect.Value.
func (n *Object) SetSourceReflected(value reflect.Value) {
	n.source = value
	n.sourceType = value.Type()
}

// WithSourceReflected is a chainable variant of SetSourceReflected.
func (n *Object) WithSourceReflected(value reflect.Value) *Object {
	n.SetSourceReflected(value)
	return n
}

// WithDefaultConverter sets the converter used for type coercion (e.g. during Set operations).
func (n *Object) WithDefaultConverter(value schema.DefaultConverter) *Object {
	n.defaultConverter = value
	return n
}

func (n *Object) SetDefaultConverter(value schema.DefaultConverter) {
	n.defaultConverter = value
}

func NewObject() *Object {
	n := new(Object)
	n.defaultConverter = schema.NewConversion()
	return n
}

/*
Object is the module for manipulating a source object.

Usage:
 1. Instantiate using NewObject.
 2. Set required parameters i.e., source.
 3. Manipulate source using Object.Get, Object.Set, Object.Delete, or Object.ForEach.
 4. Get modified source using Object.GetSourceInterface.

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

	noOfResults, err := objManip.Set("$.data.metadata.Status", "inactive")

	noOfResults, err = objManip.Delete("$.data.metadata.Status")

	// retrieve modified source after Set/Delete
	var modifiedSource any = objManip.GetSourceInterface()
*/
type Object struct {
	// Used by ForEach.
	ifValueFoundInObject IfValueFoundInObject

	// Useful especially with the Set method for creating new nested objects when starting with an empty source.
	//
	// Initialize with SetSchema or WithSchema.
	schema schema.Schema

	// Value to set in source by the Set method.
	valueToSet reflect.Value

	// Result from Get.
	valueFound reflect.Value

	// Made by Get, Set and Delete.
	noOfResults uint64

	// Last error encountered when processing the source especially for the recursive descent pattern or union pattern in path.JSONPath.
	lastError error

	// Root object to work with.
	//
	// Will be modified with Set and Delete.
	//
	// Initialize with NewObject parameter, or SetSourceInterface.
	source reflect.Value
	// Computed when you use SetSourceInterface.
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
  - value - value found. If you want the Go (any) value you can call `value.Interface()`

Return `true` to terminate ForEach loop.
*/
type IfValueFoundInObject func(jsonPath path.RecursiveDescentSegment, value reflect.Value) bool
