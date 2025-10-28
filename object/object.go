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

type Object struct {
	// Used by ForEach.
	ifValueFoundInObject IfValueFoundInObject

	schema     schema.Schema
	valueToSet any

	noOfModifications uint64
	lastError         error

	// Root object to work with.
	//
	// Will be modified with Set and Delete.
	source     reflect.Value
	sourceType reflect.Type

	recursiveDescentSegments path.RecursiveDescentSegments
	defaultConverter         schema.DefaultConverter
}

type IfValueFoundInObject func(jsonPath path.RecursiveDescentSegment, value any) bool
