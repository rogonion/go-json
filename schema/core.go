package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

// SchemaPath represents a JSONPath both in its parsed and unparsed form.
//
// Ideally functions are expected to only work with a JSONPath with no recursive descent searches.
type SchemaPath interface {
	path.JSONPath | path.RecursiveDescentSegment | path.RecursiveDescentSegments
}

// Deserializer For performing deserialization of data from various source formats to a destination that adheres to the Schema.
type Deserializer interface {
	// FromJSON Deserializes JSON string into destination using Schema.
	//
	// Expects destination to be a pointer to where deserialized data should be stored.
	FromJSON(data []byte, schema Schema, destination any) error

	// FromYAML deserializes YAML string into destination using Schema.
	//
	// Expects destination to be a pointer to where deserialized data should be stored.
	FromYAML(data []byte, schema Schema, destination any) error
}

type DefaultConverter interface {
	// RecursiveConvert is method called when schema of type Schema is encountered in the recursive conversion process.
	RecursiveConvert(source reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error)

	// Convert is entrypoint for conversion.
	Convert(data any, schema Schema, destination any) error
}

// Converter for defining custom conversion logic.
//
// Meant to be implemented by custom data types that need to perform specific value-based conversion beyond defaults.
type Converter interface {
	// Convert converts data based on schema.
	//
	// Parameters:
	//	- data - The data to be converted.
	//	- schema - The schema encountered to RecursiveConvert against.
	//	- pathSegments - Current Path segments where data was encountered. Useful if error is returned as Error.
	// Returns:
	//	- deserialized data.
	//	- Error of ErrDataConversionFailed if conversion was not successful
	Convert(data reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error)
}

// Converters Map of custom converters.
//
// Intended to be used for custom conversion logic of user-defined types like structs.
type Converters map[reflect.Type]Converter

type DefaultValidator interface {
	// ValidateData is entrypoint for validation.
	ValidateData(data any, schema Schema) (bool, error)
}

// Validator for defining custom data validation logic.
//
// Meant to be implemented by custom data types that need to perform specific value-based validation that goes beyond the defaults.
type Validator interface {
	// ValidateData validates data against a SchemaManip using custom rules.
	//
	// Parameters:
	//	- data - The data to be validated.
	//	- schema - The schema encountered to be validated against with custom rules.
	//	- pathSegments - Current Path segments where data was encountered. Useful if error is returned as Error.
	// Returns:
	//	- bool to indicate if schema validation was successful.
	//	- Error of ErrDataValidationAgainstSchemaFailed if schema validation was not successful.
	ValidateData(data any, schema Schema, pathSegments path.RecursiveDescentSegment) (bool, error)
}

// Validators Map of custom converters.
//
// Intended to be used for custom validation logic of user-defined types like structs.
type Validators map[reflect.Type]Validator

// Schema structs that represent a JSON-Like schema.
type Schema interface {
	// IsSchema placeholder implementor that returns `true` to indicate that they represent a JSON-Like schema.
	IsSchema() bool

	String() string
}

type DynamicSchema struct {
	// The key for the default DynamicSchemaNode in Nodes.
	DefaultSchemaNodeKey string

	// A map of DynamicSchemaNode, each representing a single valid schema.
	Nodes map[string]*DynamicSchemaNode

	// A list of valid DynamicSchemaNode keys in Nodes. Usually populated through the schema validation process.
	ValidSchemaNodeKeys []string
}

func (d *DynamicSchema) IsSchema() bool {
	return true
}

func (d *DynamicSchema) String() string {
	if j, err := json.Marshal(d); err != nil {
		return fmt.Sprintf("<DynamicSchema Error: %v>", err)
	} else {
		return string(j)
	}
}

const (
	DynamicSchemaDefaultNodeKey string = "default"
)

func NewDynamicSchema() *DynamicSchema {
	return &DynamicSchema{
		DefaultSchemaNodeKey: DynamicSchemaDefaultNodeKey,
		Nodes:                nil,
		ValidSchemaNodeKeys:  nil,
	}
}

type ChildNodes map[string]Schema

// DynamicSchemaNode defines a single specific schema node within a DynamicSchema.
//
// Useful when recursively setting data in a nested data structure during the creation of new nesting structure by discovering the exact type to use at each path.CollectionMemberSegment in a Path.
type DynamicSchemaNode struct {
	// The full type of the data.
	Type reflect.Type

	// The underlying type of the data.
	Kind reflect.Kind

	// Optional default value to use for new initializations.
	DefaultValue func() reflect.Value

	// Indicates if DefaultValue has been set since it can be nil.
	IsDefaultValueSet bool

	// Specifies whether the current value can be empty.
	Nilable bool

	// Specify a Validator for this specific node.
	Validator Validator

	// A recursive map defining the schema for the following:
	// 	- Specific key-value entries in an associative collection like map.
	//
	//		For each entry, it is important to specify the key type using ChildNodesAssociativeCollectionEntriesKeySchema.
	// 	- Specific elements at indexes in a linear collection like slice or array.
	// 	- All struct top level members specifically those that are User defined.
	ChildNodes ChildNodes

	// Schema for what the current DynamicSchemaNode points to. Mandatory if Kind is reflect.Pointer
	ChildNodesPointerSchema Schema

	// Schema for all keys of entries in an associative collection. Mandatory if Kind is reflect.Map.
	ChildNodesAssociativeCollectionEntriesKeySchema Schema

	// Schema for all values of entries in a map. Mandatory if Kind is reflect.Map.
	ChildNodesAssociativeCollectionEntriesValueSchema Schema

	// Schema for all elements in a slice/array. Mandatory if Kind is reflect.Slice or reflect.Array.
	ChildNodesLinearCollectionElementsSchema Schema

	// Schema for node that is a specific entry in an associative collection.
	//
	// Ideally this means that the Kind in ChildNodesAssociativeCollectionEntriesKeySchema of the parent map is reflect.Interface.
	AssociativeCollectionEntryKeySchema Schema

	// Ensure all ChildNodes are present and validated.
	ChildNodesMustBeValid bool

	// Specify Converter for this specific node.
	Converter Converter
}

func (d *DynamicSchemaNode) IsSchema() bool {
	return true
}

func (d *DynamicSchemaNode) String() string {
	if j, err := json.Marshal(d); err != nil {
		return fmt.Sprintf("<DynamicSchemaNode Error: %v>", err)
	} else {
		return string(j)
	}
}

func NewDynamicSchemaNode() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Type:              reflect.TypeOf(nil),
		DefaultValue:      nil,
		IsDefaultValueSet: false,
		Nilable:           false,
	}
}

var (
	// ErrSchemaProcessorError is the base error for DataProcessor interface.
	ErrSchemaProcessorError = errors.New("schema processing encountered an error")

	// ErrSchemaPathError is the base error for GetSchemaAtPath.
	ErrSchemaPathError = errors.New("schema path error")

	// ErrDataValidationAgainstSchemaFailed for when schema validation rule fails.
	ErrDataValidationAgainstSchemaFailed = errors.New("data validation against schema failed")

	// ErrDataDeserializationFailed for when deserialization fails.
	ErrDataDeserializationFailed = errors.New("data deserialization failed")

	// ErrDataConversionFailed for when conversion fails.
	ErrDataConversionFailed = errors.New("data conversion failed")
)

// Error for when DataProcessor execution fails.
type Error struct {
	Err          error
	FunctionName string
	Message      string
	PathSegments path.RecursiveDescentSegment
	Schema       Schema
	Data         interface{}
}

func NewError(error error, functionName string, message string, schema Schema, data interface{}, pathSegments path.RecursiveDescentSegment) *Error {
	n := new(Error)
	if error != nil {
		n.Err = fmt.Errorf("%w: %w", ErrSchemaProcessorError, error)
	} else {
		n.Err = ErrSchemaProcessorError
	}
	n.FunctionName = functionName
	n.Message = message
	n.Schema = schema
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
	if e.Schema != nil {
		str = str + " \nSchema: " + e.Schema.String()
	}
	if e.Data != nil {
		str = str + fmt.Sprintf(" \nData: %+v", internal.JsonStringifyMust(e.Data))
	}
	return str
}
