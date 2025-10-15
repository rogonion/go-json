package schema

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

func (n *Validation) validateDataWithDynamicSchemaNodeStruct(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodeStruct"

	if internal.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError(nil, FunctionName, "data cannot be nil", schema, data.Interface(), pathSegments)
		}
		return true, nil
	}

	if data.Kind() != reflect.Struct {
		return false, NewError(nil, FunctionName, "data.Kind() is not a struct", schema, data.Interface(), pathSegments)
	}

	if len(schema.ChildNodes) == 0 {
		return false, NewError(nil, FunctionName, "no schema for properties in in data struct found", schema, data.Interface(), pathSegments)
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
		return false, NewError(nil, FunctionName, "not all child nodes are present and validated against", schema, data.Interface(), pathSegments)
	}

	return true, nil
}

func (n *Validation) validateDataWithDynamicSchemaNodeMap(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodeMap"

	if internal.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data cannot be nil", schema, data.Interface(), pathSegments)
		}
		return true, nil
	}

	if data.Kind() != reflect.Map {
		return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data.Kind() is not map", schema, data.Interface(), pathSegments)
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
							if childSchemaKeyValid, _ := n.validateData(key, childNode.AssociativeCollectionEntryKeySchema, currentPathSegments); childSchemaKeyValid {
								if childValueSchemaValid, _ := n.validateDataWithDynamicSchemaNode(childMapValue, childNode, currentPathSegments); childValueSchemaValid {
									cs.ValidSchemaNodeKeys = append(cs.ValidSchemaNodeKeys, childNodeKey)
								}
							}
						}
						if len(cs.ValidSchemaNodeKeys) == 0 {
							return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("map entry with key %s not valid against any DynamicSchema nodes", key.String()), cs, childMapValue, currentPathSegments)
						}
					} else {
						return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("no DynamicSchema nodes found for key %s", key.String()), cs, childMapValue, currentPathSegments)
					}
				case *DynamicSchemaNode:
					if childSchemaKeyValid, _ := n.validateData(key, cs.AssociativeCollectionEntryKeySchema, currentPathSegments); childSchemaKeyValid {
						childMapValue := data.MapIndex(key)
						if childValueSchemaValid, _ := n.validateDataWithDynamicSchemaNode(childMapValue, cs, currentPathSegments); childValueSchemaValid {
							break
						}
						return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Value for map key %s not valid against schema", key.String()), cs, childMapValue.Interface(), currentPathSegments)
					}
					return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Key for map key %s not valid against schema", key.String()), cs, key.Interface(), currentPathSegments)
				default:
					return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Unsupported schema type for map key %s", key.String()), childSchema, data.Interface(), currentPathSegments)
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
					return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Value for map key %s not valid against schema", key.String()), schema.ChildNodesAssociativeCollectionEntriesValueSchema, childMapValue.Interface(), currentPathSegments)
				}
				return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Key for map key %s not valid against schema", key.String()), schema.ChildNodesAssociativeCollectionEntriesKeySchema, key.Interface(), currentPathSegments)
			}

			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("SchemaManip for map key %s not found", key.String()), schema, data.Interface(), currentPathSegments)
		}

		if len(childSchemaNodesValidated) != len(schema.ChildNodes) && schema.ChildNodesMustBeValid {
			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "not all child nodes are present and validated against", schema, data.Interface(), pathSegments)
		}

		return true, nil
	}

	return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "no schema to validate entries in data (map) found", schema, data.Interface(), pathSegments)
}

func (n *Validation) validateDataWithDynamicSchemaNodeArraySlice(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodeArraySlice"

	if internal.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data cannot be nil", schema, data.Interface(), pathSegments)
		}
		return true, nil
	}

	if data.Kind() != reflect.Slice && data.Kind() != reflect.Array {
		return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data.Kind() is not slice or array", schema, data.Interface(), pathSegments)
	}

	if schema.ChildNodesLinearCollectionElementsSchema != nil {
		for i := 0; i < data.Len(); i++ {
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Index: i, IsIndex: true})

			if dataValidAgainstSchema, err := n.validateData(data.Index(i), schema.ChildNodesLinearCollectionElementsSchema, currentPathSegments); !dataValidAgainstSchema {
				return false, err
			}
		}

		return true, nil
	}

	return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "schema to validate element(s) in data (slice/array) not found", schema, data.Interface(), pathSegments)
}

func (n *Validation) validateDataWithDynamicSchemaNodePointer(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNodePointer"

	if internal.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data cannot be nil", schema, data.Interface(), pathSegments)
		}
		return true, nil
	}

	if schema.ChildNodesPointerSchema == nil {
		return true, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "schema for value that data (pointer) points to has not been set (schema.ChildNodesPointerSchema is nil)", schema, data.Interface(), pathSegments)
	}

	return n.validateData(data.Elem(), schema.ChildNodesPointerSchema, pathSegments)
}

func (n *Validation) validateDataWithDynamicSchemaNode(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validateDataWithDynamicSchemaNode"

	if internal.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data cannot be nil", schema, data.Interface(), pathSegments)
		}
	}

	if schema.Kind == reflect.Interface {
		return true, nil
	}

	if data.Kind() != schema.Kind {
		return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data.Kind is not valid", schema, data.Interface(), pathSegments)
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
			return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data.Type is not valid", schema, data.Interface(), pathSegments)
		}
		return true, nil
	}
}

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
		return true, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "no schema nodes found", schema, data.Interface(), pathSegments)
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

func (n *Validation) validateData(data reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "validationData"

	switch s := schema.(type) {
	case *DynamicSchema:
		return n.validateDataWithDynamicSchema(data, s, pathSegments)
	case *DynamicSchemaNode:
		return n.validateDataWithDynamicSchemaNode(data, s, pathSegments)
	default:
		return false, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "unsupported schema type", schema, data.Interface(), pathSegments)
	}
}

func (n *Validation) ValidateData(data any, schema Schema) (bool, error) {
	return n.validateData(reflect.ValueOf(data), schema, path.RecursiveDescentSegment{
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

type Validation struct {
	validateOnFirstMatch bool
	customValidators     Validators
}
