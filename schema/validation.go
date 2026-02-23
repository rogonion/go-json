package schema

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/path"
)

// validateDataWithDynamicSchemaNodeStruct validates a struct against a DynamicSchemaNode.
// It iterates over the struct fields and validates them against the corresponding child nodes in the schema.
func (n *Validation) validateDataWithDynamicSchemaNodeStruct(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodeStruct"

	if core.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("data cannot be nil").
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
		}
		return true, nil
	}

	if data.Kind() != reflect.Struct {
		return false, NewError().WithFunctionName(FunctionName).WithMessage("data.Kind() is not a struct").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	if len(schema.ChildNodes) == 0 {
		return false, NewError().WithFunctionName(FunctionName).WithMessage("no schema for properties in in data struct found").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	childSchemaNodesValidated := make([]string, 0)
	for i := 0; i < data.NumField(); i++ {
		structFieldName := data.Type().Field(i).Name

		childSchema, ok := schema.ChildNodes[structFieldName]
		if !ok {
			continue
		}

		childSchemaNodesValidated = append(childSchemaNodesValidated, structFieldName)

		if dataValidAgainstSchema, err := n.validateData(data.Field(i), childSchema, append(pathSegments, &path.CollectionMemberSegment{Key: structFieldName, IsKey: true})); !dataValidAgainstSchema {
			return false, err
		}
	}

	if len(childSchemaNodesValidated) != len(schema.ChildNodes) && schema.ChildNodesMustBeValid {
		return false, NewError().WithFunctionName(FunctionName).WithMessage("not all child nodes are present and validated against").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	return true, nil
}

// validateDataWithDynamicSchemaNodeMap validates a map against a DynamicSchemaNode.
// It checks if the map keys and values conform to the schema defined in ChildNodes or AssociativeCollectionEntries*Schema.
func (n *Validation) validateDataWithDynamicSchemaNodeMap(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodeMap"

	if core.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("data cannot be nil").
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
		}
		return true, nil
	}

	if data.Kind() != reflect.Map {
		return false, NewError().WithFunctionName(FunctionName).WithMessage("data.Kind() is not map").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	if len(schema.ChildNodes) > 0 || (schema.ChildNodesAssociativeCollectionEntriesKeySchema != nil && schema.ChildNodesAssociativeCollectionEntriesValueSchema != nil) {
		childSchemaNodesValidated := make([]string, 0)

		for _, key := range data.MapKeys() {
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: key.String(), IsKey: true})

			if childSchema, ok := schema.ChildNodes[key.String()]; ok {
				switch cs := childSchema.(type) {
				case *DynamicSchema:
					childMapValue := data.MapIndex(key)
					if len(cs.Nodes) > 0 {
						for childNodeKey, childNode := range cs.Nodes {
							keySchema := childNode.AssociativeCollectionEntryKeySchema
							if keySchema == nil {
								keySchema = schema.ChildNodesAssociativeCollectionEntriesKeySchema
							}

							if childSchemaKeyValid, _ := n.validateData(key, keySchema, currentPathSegments); childSchemaKeyValid {
								if childValueSchemaValid, _ := n.validateDataWithDynamicSchemaNode(childMapValue, childNode, currentPathSegments); childValueSchemaValid {
									cs.ValidSchemaNodeKeys = append(cs.ValidSchemaNodeKeys, childNodeKey)
								}
							}
						}
						if len(cs.ValidSchemaNodeKeys) == 0 {
							return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("map entry with key %s not valid against any DynamicSchema nodes", key.String())).
								WithNestedError(ErrDataValidationAgainstSchemaFailed).
								WithData(core.JsonObject{"Schema": (cs), "Data": childMapValue.Interface(), "PathSegments": currentPathSegments})
						}
					} else {
						return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("no DynamicSchema nodes found for key %s", key.String())).
							WithNestedError(ErrDataValidationAgainstSchemaFailed).
							WithData(core.JsonObject{"Schema": (cs), "Data": childMapValue.Interface(), "PathSegments": currentPathSegments})
					}
				case *DynamicSchemaNode:
					keySchema := cs.AssociativeCollectionEntryKeySchema
					if keySchema == nil {
						keySchema = schema.ChildNodesAssociativeCollectionEntriesKeySchema
					}

					if childSchemaKeyValid, _ := n.validateData(key, keySchema, currentPathSegments); childSchemaKeyValid {
						childMapValue := data.MapIndex(key)
						if childValueSchemaValid, _ := n.validateDataWithDynamicSchemaNode(childMapValue, cs, currentPathSegments); childValueSchemaValid {
							break
						}
						return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("Value for map key %s not valid against schema", key.String())).
							WithNestedError(ErrDataValidationAgainstSchemaFailed).
							WithData(core.JsonObject{"Schema": (cs), "Data": data.Interface(), "PathSegments": currentPathSegments})
					}
					return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("Key for map key %s not valid against schema", key.String())).
						WithNestedError(ErrDataValidationAgainstSchemaFailed).
						WithData(core.JsonObject{"Schema": (cs), "Data": data.Interface(), "PathSegments": currentPathSegments})
				default:
					return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("Unsupported schema type for map key %s", key.String())).
						WithNestedError(ErrDataValidationAgainstSchemaFailed).
						WithData(core.JsonObject{"Schema": (childSchema), "Data": data.Interface(), "PathSegments": currentPathSegments})
				}

				childSchemaNodesValidated = append(childSchemaNodesValidated, key.String())
				continue
			}

			if schema.ChildNodesAssociativeCollectionEntriesKeySchema != nil && schema.ChildNodesAssociativeCollectionEntriesValueSchema != nil {
				if childSchemaKeyValid, _ := n.validateData(key, schema.ChildNodesAssociativeCollectionEntriesKeySchema, currentPathSegments); childSchemaKeyValid {
					childMapValue := data.MapIndex(key)
					if childValueSchemaValid, _ := n.validateData(childMapValue, schema.ChildNodesAssociativeCollectionEntriesValueSchema, currentPathSegments); childValueSchemaValid {
						continue
					}
					return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("Value for map key %s not valid against schema", key.String())).
						WithNestedError(ErrDataValidationAgainstSchemaFailed).
						WithData(core.JsonObject{"Schema": (schema.ChildNodesAssociativeCollectionEntriesValueSchema), "Data": data.Interface(), "PathSegments": currentPathSegments})
				}

				return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("Key for map key %s not valid against schema", key.String())).
					WithNestedError(ErrDataValidationAgainstSchemaFailed).
					WithData(core.JsonObject{"Schema": (schema.ChildNodesAssociativeCollectionEntriesKeySchema), "Data": data.Interface(), "PathSegments": currentPathSegments})
			}

			return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("SchemaManip for map key %s not found", key.String())).
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": currentPathSegments})
		}

		if len(childSchemaNodesValidated) != len(schema.ChildNodes) && schema.ChildNodesMustBeValid {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("not all child nodes are present and validated against").
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
		}

		return true, nil
	}

	return false, NewError().WithFunctionName(FunctionName).WithMessage("no schema to validate entries in data (map) found").
		WithNestedError(ErrDataValidationAgainstSchemaFailed).
		WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
}

// validateDataWithDynamicSchemaNodeArraySlice validates a slice or array against a DynamicSchemaNode.
// It iterates over the elements and validates them against ChildNodesLinearCollectionElementsSchema or specific ChildNodes.
func (n *Validation) validateDataWithDynamicSchemaNodeArraySlice(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodeArraySlice"

	if core.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("data cannot be nil").
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
		}
		return true, nil
	}

	if data.Kind() != reflect.Slice && data.Kind() != reflect.Array {
		return false, NewError().WithFunctionName(FunctionName).WithMessage("data.Kind() is not slice or array").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	if schema.ChildNodesLinearCollectionElementsSchema != nil {
		for i := 0; i < data.Len(); i++ {
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Index: i, IsIndex: true})

			currentSchema := schema.ChildNodesLinearCollectionElementsSchema

			if len(schema.ChildNodes) > 0 {
				if schemaForElementAtIndex, ok := schema.ChildNodes[fmt.Sprintf("%d", i)]; ok {
					currentSchema = schemaForElementAtIndex
				}
			}

			if dataValidAgainstSchema, err := n.validateData(data.Index(i), currentSchema, currentPathSegments); !dataValidAgainstSchema {
				return false, err
			}
		}

		return true, nil
	}

	return false, NewError().WithFunctionName(FunctionName).WithMessage("schema to validate element(s) in data (slice/array) not found").
		WithNestedError(ErrDataValidationAgainstSchemaFailed).
		WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
}

// validateDataWithDynamicSchemaNodePointer validates a pointer against a DynamicSchemaNode.
// It dereferences the pointer and validates the underlying value against ChildNodesPointerSchema.
func (n *Validation) validateDataWithDynamicSchemaNodePointer(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodePointer"

	if core.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("data cannot be nil").
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
		}
		return true, nil
	}

	if schema.ChildNodesPointerSchema == nil {
		return true, NewError().WithFunctionName(FunctionName).WithMessage("schema for value that data (pointer) points to has not been set (schema.ChildNodesPointerSchema is nil)").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	return n.validateData(data.Elem(), schema.ChildNodesPointerSchema, pathSegments)
}

// validateDataWithDynamicSchemaNode is the main validation logic for a single DynamicSchemaNode.
// It dispatches to specific validation methods based on the data kind (Struct, Map, Slice, Pointer, etc.)
// or uses custom validators if provided.
func (n *Validation) validateDataWithDynamicSchemaNode(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNode"

	if core.IsNilOrInvalid(data) {
		if !schema.Nilable {
			var dataInterface any
			if data.IsValid() {
				dataInterface = data.Interface()
			}
			return false, NewError().WithFunctionName(FunctionName).WithMessage("data cannot be nil").
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": dataInterface, "PathSegments": pathSegments})
		}
		return true, nil
	}

	if data.Kind() == reflect.Interface {
		data = data.Elem()
	}

	if schema.Kind == reflect.Interface {
		return true, nil
	}

	if data.Kind() != schema.Kind {
		return false, NewError().WithFunctionName(FunctionName).WithMessage("data.Kind is not valid").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	if schema.Validator != nil {
		return schema.Validator.ValidateData(data.Interface(), schema, pathSegments)
	}

	if customValidator, ok := n.customValidators[data.Type()]; ok {
		return customValidator.ValidateData(data.Interface(), schema, pathSegments)
	}

	switch data.Kind() {
	case reflect.Pointer:
		return n.validateDataWithDynamicSchemaNodePointer(data, schema, pathSegments)
	case reflect.Slice, reflect.Array:
		return n.validateDataWithDynamicSchemaNodeArraySlice(data, schema, pathSegments)
	case reflect.Map:
		return n.validateDataWithDynamicSchemaNodeMap(data, schema, pathSegments)
	case reflect.Struct:
		return n.validateDataWithDynamicSchemaNodeStruct(data, schema, pathSegments)
	default:
		if data.Type() != schema.Type {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("data.Type is not valid").
				WithNestedError(ErrDataValidationAgainstSchemaFailed).
				WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
		}
		return true, nil
	}
}

// validateDataWithDynamicSchema validates data against a DynamicSchema (which contains multiple potential schema nodes).
// It attempts to validate against the default node first, then iterates through other nodes until a match is found.
func (n *Validation) validateDataWithDynamicSchema(data reflect.Value, schema *DynamicSchema, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchema"

	if len(schema.DefaultSchemaNodeKey) > 0 {
		if dynamicSchemaNode, found := schema.Nodes[schema.DefaultSchemaNodeKey]; found {
			if dataValidAgainstSchema, _ := n.validateDataWithDynamicSchemaNode(data, dynamicSchemaNode, pathSegments); dataValidAgainstSchema {
				schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schema.DefaultSchemaNodeKey)
				return true, nil
			}
		}
	}

	if len(schema.Nodes) == 0 {
		return true, NewError().WithFunctionName(FunctionName).WithMessage("no schema nodes found").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}

	var lastSchemaNodeErr error
	for schemaNodeKey, dynamicSchemaNode := range schema.Nodes {
		if schemaNodeKey == schema.DefaultSchemaNodeKey {
			continue
		}
		dataValidAgainstSchema, err := n.validateDataWithDynamicSchemaNode(data, dynamicSchemaNode, pathSegments)
		if dataValidAgainstSchema {
			schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schemaNodeKey)
			if n.validateOnFirstMatch {
				return true, nil
			}
			continue
		}
		if err != nil {
			lastSchemaNodeErr = err
		}
	}

	if len(schema.ValidSchemaNodeKeys) == 0 {
		return false, lastSchemaNodeErr
	}
	return true, nil
}

// validateData is the internal entry point for validation recursion.
// It routes the validation to either validateDataWithDynamicSchema or validateDataWithDynamicSchemaNode.
func (n *Validation) validateData(data reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validationData"

	switch s := schema.(type) {
	case *DynamicSchema:
		return n.validateDataWithDynamicSchema(data, s, pathSegments)
	case *DynamicSchemaNode:
		return n.validateDataWithDynamicSchemaNode(data, s, pathSegments)
	default:
		return false, NewError().WithFunctionName(FunctionName).WithMessage("unsupported schema type").
			WithNestedError(ErrDataValidationAgainstSchemaFailed).
			WithData(core.JsonObject{"Schema": schema, "Data": data.Interface(), "PathSegments": pathSegments})
	}
}

/*
ValidateNode offers a direct entry point to Validation.validateDataWithDynamicSchemaNode.

Useful for simple type validations or custom validations as the amount of instructions and allocations are less.
*/
func (n *Validation) ValidateNode(source reflect.Value, schema *DynamicSchemaNode) (bool, error) {
	return n.validateDataWithDynamicSchemaNode(source, schema, nil)
}

// ValidateData checks if the provided data adheres to the constraints defined in the Schema.
func (n *Validation) ValidateData(data any, schema Schema) (bool, error) {
	return n.validateData(reflect.ValueOf(data), schema, path.RecursiveDescentSegment{
		{
			Key:       "$",
			IsKeyRoot: true,
		},
	})
}

// ValidateDataReflect is similar to ValidateData but accepts a reflect.Value directly.
func (n *Validation) ValidateDataReflect(data reflect.Value, schema Schema) (bool, error) {
	return n.validateData(data, schema, path.RecursiveDescentSegment{
		{
			Key:       "$",
			IsKeyRoot: true,
		},
	})
}

func (n *Validation) WithCustomValidators(value Validators) *Validation {
	n.customValidators = value
	return n
}

func (n *Validation) SetCustomValidators(value Validators) {
	n.customValidators = value
}

func (n *Validation) WithValidateOnFirstMatch(value bool) *Validation {
	n.validateOnFirstMatch = value
	return n
}

func (n *Validation) SetValidateOnFirstMatch(value bool) {
	n.validateOnFirstMatch = value
}

func NewValidation() *Validation {
	n := new(Validation)
	n.validateOnFirstMatch = true
	return n
}

/*
Validation provides methods to check if data conforms to a Schema.

Usage:
 1. Instantiate using NewValidation.
 2. (Optional) Set custom validators or flags.
 3. Call ValidateData.

Example:

	schema := &DynamicSchemaNode{
		Kind: reflect.String,
		Type: reflect.TypeOf(""),
	}

	validation := NewValidation()
	ok, err := validation.ValidateData("this is a string", schema)
*/
type Validation struct {
	validateOnFirstMatch bool
	customValidators     Validators
}
