package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/rogonion/go-json/path"
)

func (n *Conversion) convertToBoolWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToBoolWithDynamicSchemaNode"

	if schema.Kind != reflect.Bool {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not bool").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.Bool:
		return source, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(source.Int() != 0), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		if convertedInt, err := n.convertToIntWithDynamicSchemaNode(source, &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)}, pathSegments); err == nil {
			return reflect.ValueOf(convertedInt.Int() != 0), nil
		} else {
			return reflect.Zero(schema.Type), NewError(FunctionName, "RecursiveConvert number to int for boolean conversion failed").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for bool conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToStringWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToStringWithDynamicSchemaNode"

	if schema.Kind != reflect.String {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not string").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.String:
		return source, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(strconv.FormatInt(source.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(strconv.FormatUint(source.Uint(), 10)), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(strconv.FormatFloat(source.Float(), 'f', -1, 64)), nil
	case reflect.Bool:
		return reflect.ValueOf(strconv.FormatBool(source.Bool())), nil
	default:
		var ptrToSource any
		if source.Type().Kind() == reflect.Ptr {
			ptrToSource = source.Interface()
		} else {
			ptrToSource = reflect.New(source.Type()).Interface()
		}
		if jsonString, err := json.Marshal(ptrToSource); err == nil {
			return reflect.ValueOf(string(jsonString)), nil
		} else {
			return reflect.Zero(schema.Type), NewError(FunctionName, "RecursiveConvert source to json string failed").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
	}
}

func (n *Conversion) convertToFloatWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToFloatWithDynamicSchemaNode"

	if !slices.Contains([]reflect.Kind{reflect.Float32, reflect.Float64}, schema.Kind) {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not float or variant").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(float64(source.Int())).Convert(schema.Type), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(float64(source.Int())).Convert(schema.Type), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(source.Float()).Convert(schema.Type), nil
	case reflect.String:
		if i, err := strconv.ParseFloat(source.String(), 64); err == nil {
			return reflect.ValueOf(i).Convert(schema.Type), nil
		} else {
			return reflect.Zero(schema.Type), NewError(FunctionName, "RecursiveConvert source string to float failed").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for float conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToUintWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToUintWithDynamicSchemaNode"

	if !slices.Contains([]reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64}, schema.Kind) {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not uint or variant").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(uint64(source.Int())).Convert(schema.Type), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(source.Uint()).Convert(schema.Type), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(uint64(source.Float())).Convert(schema.Type), nil
	case reflect.String:
		if i, err := strconv.ParseUint(source.String(), 10, 64); err == nil {
			return reflect.ValueOf(i).Convert(schema.Type), nil
		} else {
			return reflect.Zero(schema.Type), NewError(FunctionName, "RecursiveConvert source string to uint failed").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for uint conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToIntWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToIntWithDynamicSchemaNode"

	if !slices.Contains([]reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64}, schema.Kind) {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not int or variant").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(source.Int()).Convert(schema.Type), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(int64(source.Uint())).Convert(schema.Type), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(int64(source.Float())).Convert(schema.Type), nil
	case reflect.String:
		if i, err := strconv.ParseInt(source.String(), 10, 64); err == nil {
			return reflect.ValueOf(i).Convert(schema.Type), nil
		} else {
			return reflect.Zero(schema.Type), NewError(FunctionName, "RecursiveConvert source string to int failed").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for int conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToStructWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToStructWithDynamicSchemaNode"

	if schema.Kind != reflect.Struct {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not struct").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.Struct:
		var newStruct reflect.Value
		if schema.DefaultValue != nil {
			newStruct = schema.DefaultValue()
		} else {
			newStruct = reflect.New(schema.Type).Elem()
		}

		for i := 0; i < schema.Type.NumField(); i++ {
			destField := schema.Type.Field(i)
			sourceField := source.FieldByName(destField.Name)

			childSchema, ok := schema.ChildNodes[destField.Name]
			if !ok {
				return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("schema for field %s has not been found for struct conversion", destField.Name)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
			}

			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: destField.Name, IsKey: true})

			if sourceField.IsValid() && newStruct.Field(i).CanSet() {
				if convertedValue, err := n.RecursiveConvert(sourceField, childSchema, currentPathSegments); err == nil {
					newStruct.Field(i).Set(convertedValue)
					continue
				} else {
					if !schema.Nilable {
						return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert value for struct field %s failed", destField.Name)).WithSchema(schema).WithData(sourceField.Interface()).WithPathSegments(currentPathSegments).WithNestedError(err)
					}
				}
			}
		}

		return newStruct, nil
	case reflect.Map:
		var newStruct reflect.Value
		if schema.DefaultValue != nil {
			newStruct = schema.DefaultValue()
		} else {
			newStruct = reflect.New(schema.Type).Elem()
		}

		iter := source.MapRange()
		for iter.Next() {
			key, val := iter.Key(), iter.Value()
			if key.Kind() != reflect.String {
				return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("map key %s is not string for struct conversion", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
			}

			// Find the corresponding field in the destination struct.
			field := newStruct.FieldByName(key.String())
			if !field.IsValid() || !field.CanSet() {
				return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("map key %s is not a valid field in struct conversion", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
			}

			childSchema, ok := schema.ChildNodes[key.String()]
			if !ok {
				return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("schema for field %s has not been found for struct conversion", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
			}

			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: key.String(), IsKey: true})

			if convertedValue, err := n.RecursiveConvert(val, childSchema, currentPathSegments); err == nil {
				field.Set(convertedValue)
				continue
			} else {
				if !schema.Nilable {
					return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert value for struct field %s failed", key)).WithSchema(schema).WithData(field.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
				}
			}
		}

		return newStruct, nil
	case reflect.String:
		// Assumes source is json string
		var deserializedData interface{}
		err := json.Unmarshal([]byte(source.String()), &deserializedData)
		if err != nil {
			return reflect.Zero(schema.Type), NewError(FunctionName, "failed to RecursiveConvert string to struct using json").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
		return n.convertToStructWithDynamicSchemaNode(reflect.ValueOf(deserializedData), schema, pathSegments)
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for struct conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToMapWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToMapWithDynamicSchemaNode"

	if schema.Kind != reflect.Map {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind is not map").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.Map:
		var newMap reflect.Value
		if schema.DefaultValue != nil {
			newMap = schema.DefaultValue()
		} else {
			newMap = reflect.MakeMap(schema.Type)
		}

		iter := source.MapRange()
		for iter.Next() {
			key, val := iter.Key(), iter.Value()
			keyString, err := n.convertToStringWithDynamicSchemaNode(key, &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")}, pathSegments)
			if err != nil {
				return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("could not RecursiveConvert key key %v to string", key.Interface())).WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
			}
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: keyString.String(), IsKey: true})

			if childSchema, ok := schema.ChildNodes[keyString.String()]; ok {
				switch cs := childSchema.(type) {
				case *DynamicSchema:
					if len(cs.Nodes) > 0 {
						for childNodeKey, childNode := range cs.Nodes {
							if convertedKey, err := n.RecursiveConvert(key, childNode.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
								if convertedValue, err := n.convertToMapWithDynamicSchemaNode(val, childNode, currentPathSegments); err == nil {
									if convertedKey.IsValid() && convertedValue.IsValid() {
										cs.ValidSchemaNodeKeys = append(cs.ValidSchemaNodeKeys, childNodeKey)
										newMap.SetMapIndex(convertedKey, convertedValue)
										break
									}
								}
							}
						}
						if len(cs.ValidSchemaNodeKeys) == 0 {
							return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("map entry with key %s not valid against any DynamicSchema nodes", key)).WithSchema(cs).WithPathSegments(currentPathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
						}
					} else {
						return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("no DynamicSchema nodes found for key %s", key)).WithSchema(cs).WithPathSegments(currentPathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
					}
				case *DynamicSchemaNode:
					if convertedKey, err := n.RecursiveConvert(key, cs.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
						if convertedValue, err := n.convertToDynamicSchemaNode(val, cs, currentPathSegments); err == nil {
							if convertedKey.IsValid() && convertedValue.IsValid() {
								newMap.SetMapIndex(convertedKey, convertedValue)
								continue
							} else {
								if !schema.Nilable {
									return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("converted map entry for key %s not valid", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataConversionFailed)
								}
							}
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert value for map key %s failed", key)).WithSchema(schema).WithData(val.Interface()).WithPathSegments(currentPathSegments).WithNestedError(err)
							}
							continue
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert key for map key %s failed", key)).WithSchema(schema).WithPathSegments(currentPathSegments).WithNestedError(err)
						}
						continue
					}
				default:
					return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("Nodes in Schema for map key %s empty", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
				}
			}

			if schema.ChildNodesAssociativeCollectionEntriesKeySchema != nil && schema.ChildNodesAssociativeCollectionEntriesValueSchema != nil {
				if convertedKey, err := n.RecursiveConvert(key, schema.ChildNodesAssociativeCollectionEntriesKeySchema, currentPathSegments); err == nil {
					if convertedValue, err := n.RecursiveConvert(val, schema.ChildNodesAssociativeCollectionEntriesValueSchema, currentPathSegments); err == nil {
						if convertedKey.IsValid() && convertedValue.IsValid() {
							newMap.SetMapIndex(convertedKey, convertedValue)
							continue
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("converted map entry for key %s not valid", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataConversionFailed)
							}
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert value for map key %s failed", key)).WithSchema(schema).WithData(val.Interface()).WithPathSegments(currentPathSegments).WithNestedError(err)
						}
						continue
					}
				} else {
					if !schema.Nilable {
						return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert key for map key %s failed", key)).WithSchema(schema).WithPathSegments(currentPathSegments).WithNestedError(err)
					}
					continue
				}
			}

			return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("Schema for map key %s not found", key)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataConversionFailed)
		}

		return newMap, nil
	case reflect.Struct:
		var newMap reflect.Value
		if schema.DefaultValue != nil {
			newMap = schema.DefaultValue()
		} else {
			newMap = reflect.MakeMap(schema.Type)
		}

		for i := 0; i < source.NumField(); i++ {
			field := source.Field(i)
			fieldName := source.Type().Field(i).Name
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: fieldName, IsKey: true})

			if childSchema, ok := schema.ChildNodes[fieldName]; ok {
				switch cs := childSchema.(type) {
				case *DynamicSchema:
					if len(cs.Nodes) > 0 {
						for childNodeKey, childNode := range cs.Nodes {
							if convertedKey, err := n.RecursiveConvert(reflect.ValueOf(fieldName), childNode.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
								if convertedValue, err := n.convertToMapWithDynamicSchemaNode(field, childNode, currentPathSegments); err == nil {
									if convertedKey.IsValid() && convertedValue.IsValid() {
										cs.ValidSchemaNodeKeys = append(cs.ValidSchemaNodeKeys, childNodeKey)
										newMap.SetMapIndex(convertedKey, convertedValue)
										break
									}
								}
							}
						}
						if len(cs.ValidSchemaNodeKeys) == 0 {
							return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("failed to RecursiveConvert struct field with name %s against any DynamicSchema nodes", fieldName)).WithSchema(schema).WithPathSegments(currentPathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
						}
						continue
					} else {
						return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("no DynamicSchema nodes found for struct field with name %s", fieldName)).WithSchema(schema).WithPathSegments(currentPathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
					}
				case *DynamicSchemaNode:
					if convertedKey, err := n.RecursiveConvert(reflect.ValueOf(fieldName), cs.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
						if convertedValue, err := n.convertToMapWithDynamicSchemaNode(field, cs, currentPathSegments); err == nil {
							if convertedKey.IsValid() && convertedValue.IsValid() {
								newMap.SetMapIndex(convertedKey, convertedValue)
								continue
							} else {
								if !cs.Nilable {
									return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("converted struct field with name %s not valid", fieldName)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataConversionFailed)
								}
							}
						} else {
							if !cs.Nilable {
								return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert value for struct field with name %s failed", fieldName)).WithSchema(schema).WithData(field.Interface()).WithPathSegments(currentPathSegments).WithNestedError(err)
							}
							continue
						}
					} else {
						if !cs.Nilable {
							return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert fieldName for struct field with name %s failed", fieldName)).WithSchema(schema).WithPathSegments(currentPathSegments).WithNestedError(err)
						}
						continue
					}
				default:
					return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("Nodes in Schema for struct field with name %s empty", fieldName)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataValidationAgainstSchemaFailed)
				}
			}

			if schema.ChildNodesAssociativeCollectionEntriesKeySchema != nil && schema.ChildNodesAssociativeCollectionEntriesValueSchema != nil {
				if convertedKey, err := n.RecursiveConvert(reflect.ValueOf(fieldName), schema.ChildNodesAssociativeCollectionEntriesKeySchema, currentPathSegments); err == nil {
					if convertedValue, err := n.RecursiveConvert(field, schema.ChildNodesAssociativeCollectionEntriesValueSchema, currentPathSegments); err == nil {
						if convertedKey.IsValid() && convertedValue.IsValid() {
							newMap.SetMapIndex(convertedKey, convertedValue)
							continue
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("converted struct field with name %s not valid", fieldName)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataConversionFailed)
							}
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert value for struct field with name %s failed", fieldName)).WithSchema(schema).WithData(field.Interface()).WithPathSegments(currentPathSegments).WithNestedError(err)
						}
						continue
					}
				} else {
					if !schema.Nilable {
						return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("RecursiveConvert fieldName for struct field with name %s failed", fieldName)).WithSchema(schema).WithData(field.Interface()).WithPathSegments(currentPathSegments).WithNestedError(err)
					}
					continue
				}
			}

			return reflect.Zero(schema.Type), NewError(FunctionName, fmt.Sprintf("Schema for struct field with name %s not found", fieldName)).WithSchema(schema).WithData(source.Interface()).WithPathSegments(currentPathSegments).WithNestedError(ErrDataConversionFailed)
		}

		return newMap, nil
	case reflect.String:
		// Assumes source is json string
		var deserializedData interface{}
		err := json.Unmarshal([]byte(source.String()), &deserializedData)
		if err != nil {
			return reflect.Zero(schema.Type), NewError(FunctionName, "failed to RecursiveConvert string to map using json").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
		return n.convertToMapWithDynamicSchemaNode(reflect.ValueOf(deserializedData), schema, pathSegments)
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for map conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToArraySliceWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToArraySliceWithDynamicSchemaNode"

	if schema.Kind != reflect.Slice && schema.Kind != reflect.Array {
		return reflect.Zero(schema.Type), NewError(FunctionName, "data.Kind is not slice or array").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	switch source.Kind() {
	case reflect.String:
		// Assumes source is json string
		var deserializedData interface{}
		err := json.Unmarshal([]byte(source.String()), &deserializedData)
		if err != nil {
			return reflect.Zero(schema.Type), NewError(FunctionName, "failed to RecursiveConvert string to array using json").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(err)
		}
		return n.convertToArraySliceWithDynamicSchemaNode(reflect.ValueOf(deserializedData), schema, pathSegments)
	case reflect.Slice, reflect.Array:
		var newArraySlice reflect.Value

		if schema.ChildNodesLinearCollectionElementsSchema == nil {
			return reflect.Zero(schema.Type), NewError(FunctionName, "no schema to RecursiveConvert element(s) in data (slice/array) found").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
		}

		if schema.Kind == reflect.Slice {
			if schema.DefaultValue != nil {
				newArraySlice = reflect.MakeSlice(schema.DefaultValue().Type(), source.Len(), source.Len())
			} else {
				newArraySlice = reflect.MakeSlice(schema.Type, source.Len(), source.Len())
			}
		} else {
			if schema.DefaultValue != nil {
				newArraySlice = schema.DefaultValue()
			} else {
				newArraySlice = reflect.New(schema.Type).Elem()
			}
		}

		for i := 0; i < source.Len(); i++ {
			if i > newArraySlice.Len() {
				break
			}

			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Index: i, IsIndex: true})

			if elementResult, err := n.RecursiveConvert(source.Index(i), schema.ChildNodesLinearCollectionElementsSchema, currentPathSegments); err == nil {
				if elementResult.IsValid() {
					newArraySlice.Index(i).Set(elementResult)
				} else {
					return elementResult, NewError(FunctionName, "elementResult not valid").WithSchema(schema).WithData(elementResult.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
				}
			} else {
				return elementResult, err
			}
		}

		return newArraySlice, nil
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported source.Kind for array/slice conversion").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToPointerWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToPointerWithDynamicSchemaNode"
	// The destination must be a pointer type.
	if schema.Kind != reflect.Pointer {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema.Kind not reflect.Pointer").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	if schema.ChildNodesPointerSchema == nil {
		return reflect.Zero(schema.Type), NewError(FunctionName, "schema for value that data (pointer) points to has not been set (schema.ChildNodesPointerSchema is nil)").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	pointerToResult, err := n.RecursiveConvert(source, schema.ChildNodesPointerSchema, pathSegments)
	if err != nil {
		return reflect.Zero(schema.Type), err
	}
	if !pointerToResult.IsValid() {
		return reflect.Zero(schema.Type), NewError(FunctionName, "pointerToResult not valid").WithSchema(schema).WithData(pointerToResult.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	var newPtr reflect.Value
	if schema.DefaultValue != nil {
		newPtr = schema.DefaultValue()
	} else {
		newPtr = reflect.New(pointerToResult.Type())
	}
	newPtr.Elem().Set(pointerToResult)
	return newPtr, nil
}

func (n *Conversion) convertToDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToDynamicSchemaNode"

	// Handle nil values defensively.
	if !source.IsValid() || (source.Kind() >= reflect.Chan && source.Kind() <= reflect.Slice && source.IsNil()) {
		return reflect.New(schema.Type).Elem(), nil
	}

	// Unwrap interface values.
	if source.Kind() == reflect.Interface {
		if !source.IsNil() {
			source = source.Elem()
		}
	}

	// If the source is already assignable to the destination type,
	// return it directly. This handles primitives and direct assignments.
	if schema.Kind == reflect.Interface || source.Type().AssignableTo(schema.Type) {
		return source, nil
	}

	if schema.Converter != nil {
		return schema.Converter.Convert(source, schema, pathSegments)
	}

	if customConverter, ok := n.customConverters[source.Type()]; ok {
		return customConverter.Convert(source, schema, pathSegments)
	}

	switch schema.Kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return n.convertToIntWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return n.convertToUintWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Float32, reflect.Float64:
		return n.convertToFloatWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.String:
		return n.convertToStringWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Bool:
		return n.convertToBoolWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Struct:
		return n.convertToStructWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Map:
		return n.convertToMapWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Slice, reflect.Array:
		return n.convertToArraySliceWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Ptr:
		return n.convertToPointerWithDynamicSchemaNode(source, schema, pathSegments)
	case reflect.Interface:
		return source, nil
	default:
		return reflect.Zero(schema.Type), NewError(FunctionName, "unsupported schema.Kind").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

func (n *Conversion) convertToDynamicSchema(source reflect.Value, schema *DynamicSchema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToDynamicSchema"

	if len(schema.DefaultSchemaNodeKey) > 0 {
		if dynamicSchemaNode, found := schema.Nodes[schema.DefaultSchemaNodeKey]; found {
			if result, err := n.convertToDynamicSchemaNode(source, dynamicSchemaNode, pathSegments); err == nil {
				schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schema.DefaultSchemaNodeKey)
				return result, nil
			}
		}
	}

	if len(schema.Nodes) == 0 {
		return reflect.Value{}, NewError(FunctionName, "no schema nodes found").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}

	var lastSchemaNodeErr error
	for schemaNodeKey, dynamicSchemaNode := range schema.Nodes {
		if schemaNodeKey == schema.DefaultSchemaNodeKey {
			continue
		}
		result, err := n.convertToDynamicSchemaNode(source, dynamicSchemaNode, pathSegments)
		if err == nil {
			schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schemaNodeKey)
			return result, nil
		}
		lastSchemaNodeErr = err
	}

	return reflect.Value{}, lastSchemaNodeErr
}

func (n *Conversion) RecursiveConvert(source reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "RecursiveConvert"

	switch s := schema.(type) {
	case *DynamicSchema:
		return n.convertToDynamicSchema(source, s, pathSegments)
	case *DynamicSchemaNode:
		return n.convertToDynamicSchemaNode(source, s, pathSegments)
	default:
		return reflect.Value{}, NewError(FunctionName, "unsupported schema type").WithSchema(schema).WithData(source.Interface()).WithPathSegments(pathSegments).WithNestedError(ErrDataConversionFailed)
	}
}

/*
Convert

Parameters:
  - source - Data to convert.
  - schema - Schema to use for conversion.
  - destination - Typed pointer to location to store converted result. Will set result using reflect if type matches.

Returns error if conversion fails.
*/
func (n *Conversion) Convert(source any, schema Schema, destination any) error {
	const FunctionName = "Convert"

	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return NewError(FunctionName, "destination is not a pointer").WithSchema(schema).WithData(source).WithNestedError(ErrDataConversionFailed)
	}

	if result, err := n.RecursiveConvert(reflect.ValueOf(source), schema, path.RecursiveDescentSegment{
		{
			Key:       "$",
			IsKeyRoot: true,
		},
	}); err != nil {
		return err
	} else {
		dest := reflect.ValueOf(destination)
		if result.Type() != dest.Elem().Type() && dest.Elem().Kind() != reflect.Interface {
			return NewError(FunctionName, "destination and result type mismatch").WithSchema(schema).WithData(source).WithNestedError(ErrDataConversionFailed)
		}
		dest.Elem().Set(result)
	}
	return nil
}

func (n *Conversion) WithCustomConverters(value Converters) *Conversion {
	n.customConverters = value
	return n
}

func (n *Conversion) SetCustomConverters(value Converters) {
	n.customConverters = value
}

func NewConversion() *Conversion {
	n := new(Conversion)
	return n
}

/*
Conversion Module for converting data against Schema.

Usage:
 1. Instantiate using NewConversion.
 2. Set required parameters.
 3. Convert data using Conversion.Convert.

Example:

	schema := &DynamicSchemaNode{
		Kind: reflect.Map,
		Type: reflect.TypeOf(map[int]int{}),
		ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
			Kind: reflect.Int,
			Type: reflect.TypeOf(0),
		},
		ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
			Kind: reflect.Int,
			Type: reflect.TypeOf(0),
		},
	}

	source := map[string]string{
		"1": "1",
		"2": "2",
		"3": "3",
	}
	var destination any
	converter := NewConversion()
	err := converter.Convert(source, schema, &destination)
*/
type Conversion struct {
	customConverters Converters
}
