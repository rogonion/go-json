package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/rogonion/go-json/path"
)

func (n *Processor) convert(source reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserialize"

	switch s := schema.(type) {
	case *DynamicSchema:
		return n.convertToDynamicSchema(source, s, pathSegments)
	case *DynamicSchemaNode:
		return n.convertToDynamicSchemaNode(source, s, pathSegments)
	default:
		return reflect.Value{}, NewError(ErrDataConversionFailed, FunctionName, "unsupported schema type", schema, source, pathSegments)
	}
}

func (n *Processor) convertToDynamicSchema(source reflect.Value, schema *DynamicSchema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
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
		return reflect.Value{}, NewError(ErrDataConversionFailed, FunctionName, "no schema nodes found", schema, source, pathSegments)
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

func (n *Processor) convertToDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
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

	if customConverter, ok := n.converters[source.Type()]; ok {
		return customConverter.Convert(source, schema, pathSegments)
	}

	// Handle nil values defensively.
	if !source.IsValid() || (source.Kind() >= reflect.Chan && source.Kind() <= reflect.Slice && source.IsNil()) {
		return reflect.New(schema.Type).Elem(), nil
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
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported schema.Kind", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToBoolWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToBoolWithDynamicSchemaNode"

	if schema.Kind != reflect.Bool {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not bool", schema, source.Interface(), pathSegments)
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
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "convert number to int for boolean conversion failed", schema, source.Interface(), pathSegments)
		}
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for bool conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToStringWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserializeToStringWithDynamicSchemaNode"

	if schema.Kind != reflect.String {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not string", schema, source.Interface(), pathSegments)
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
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "convert source to json string failed", schema, source.Interface(), pathSegments)
		}
	}
}

func (n *Processor) convertToFloatWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToFloatWithDynamicSchemaNode"

	if !slices.Contains([]reflect.Kind{reflect.Float32, reflect.Float64}, schema.Kind) {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not float or variant", schema, source.Interface(), pathSegments)
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
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "convert source string to float failed", schema, source.Interface(), pathSegments)
		}
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for float conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToUintWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToUintWithDynamicSchemaNode"

	if !slices.Contains([]reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64}, schema.Kind) {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not uint or variant", schema, source.Interface(), pathSegments)
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
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "convert source string to uint failed", schema, source.Interface(), pathSegments)
		}
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for uint conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToIntWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToIntWithDynamicSchemaNode"

	if !slices.Contains([]reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64}, schema.Kind) {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not int or variant", schema, source.Interface(), pathSegments)
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
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "convert source string to int failed", schema, source.Interface(), pathSegments)
		}
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for int conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToStructWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToStructWithDynamicSchemaNode"

	if schema.Kind != reflect.Struct {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not struct", schema, source.Interface(), pathSegments)
	}

	switch source.Kind() {
	case reflect.Struct:
		newStruct := reflect.New(schema.Type).Elem()

		for i := 0; i < schema.Type.NumField(); i++ {
			destField := schema.Type.Field(i)
			sourceField := source.FieldByName(destField.Name)

			childSchema, ok := schema.ChildNodes[destField.Name]
			if !ok {
				return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("schema for field %s has not been found for struct conversion", destField.Name), schema, source.Interface(), pathSegments)
			}

			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: destField.Name, IsKey: true})

			if sourceField.IsValid() && newStruct.Field(i).CanSet() {
				if convertedValue, err := n.convert(sourceField, childSchema, currentPathSegments); err == nil {
					newStruct.Field(i).Set(convertedValue)
					continue
				} else {
					if !schema.Nilable {
						return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert value for struct field %s failed", destField.Name), schema, sourceField.Interface(), currentPathSegments)
					}
				}
			}
		}

		return newStruct, nil
	case reflect.Map:
		newStruct := reflect.New(schema.Type).Elem()

		iter := source.MapRange()
		for iter.Next() {
			key, val := iter.Key(), iter.Value()
			if key.Kind() != reflect.String {
				return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("map key %s is not string for struct conversion", key), schema, source.Interface(), pathSegments)
			}

			// Find the corresponding field in the destination struct.
			field := newStruct.FieldByName(key.String())
			if !field.IsValid() || !field.CanSet() {
				return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("map key %s is not a valid field in struct conversion", key), schema, source.Interface(), pathSegments)
			}

			childSchema, ok := schema.ChildNodes[key.String()]
			if !ok {
				return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("schema for field %s has not been found for struct conversion", key), schema, source.Interface(), pathSegments)
			}

			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: key.String(), IsKey: true})

			if convertedValue, err := n.convert(val, childSchema, currentPathSegments); err == nil {
				field.Set(convertedValue)
				continue
			} else {
				if !schema.Nilable {
					return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert value for struct field %s failed", key), schema, field.Interface(), currentPathSegments)
				}
			}
		}

		return newStruct, nil
	case reflect.String:
		// Assumes source is json string
		var deserializedData interface{}
		err := json.Unmarshal([]byte(source.String()), &deserializedData)
		if err != nil {
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "failed to convert string to struct using json", schema, source.Interface(), pathSegments)
		}
		return n.deserialize(reflect.ValueOf(deserializedData), schema, pathSegments)
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for struct conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToMapWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToMapWithDynamicSchemaNode"

	if schema.Kind != reflect.Map {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind is not map", schema, source.Interface(), pathSegments)
	}

	switch source.Kind() {
	case reflect.Map:
		newMap := reflect.MakeMap(schema.Type)

		iter := source.MapRange()
		for iter.Next() {
			key, val := iter.Key(), iter.Value()
			keyString, err := n.convertToStringWithDynamicSchemaNode(key, &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")}, pathSegments)
			if err != nil {
				return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("could not convert key key %v to string", key.Interface()), schema, source.Interface(), pathSegments)
			}
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: keyString.String(), IsKey: true})

			if childSchema, ok := schema.ChildNodes[keyString.String()]; ok {
				switch cs := childSchema.(type) {
				case *DynamicSchema:
					if len(cs.Nodes) > 0 {
						for childNodeKey, childNode := range cs.Nodes {
							if convertedKey, err := n.convert(key, childNode.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
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
							return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("map entry with key %s not valid against any DynamicSchema nodes", key), cs, schema, currentPathSegments)
						}
					} else {
						return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("no DynamicSchema nodes found for key %s", key), cs, schema, currentPathSegments)
					}
				case *DynamicSchemaNode:
					if convertedKey, err := n.convert(key, cs.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
						if convertedValue, err := n.convertToMapWithDynamicSchemaNode(val, cs, currentPathSegments); err == nil {
							if convertedKey.IsValid() && convertedValue.IsValid() {
								newMap.SetMapIndex(convertedKey, convertedValue)
								continue
							} else {
								if !schema.Nilable {
									return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("converted map entry for key %s not valid", key), schema, source.Interface(), currentPathSegments)
								}
							}
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert value for map key %s failed", key), schema, val.Interface(), currentPathSegments)
							}
							continue
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert key for map key %s failed", key), schema, key, currentPathSegments)
						}
						continue
					}
				default:
					return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Nodes in SchemaManip for map key %s empty", key), childSchema, source.Interface(), currentPathSegments)
				}
			}

			if schema.ChildNodesAssociativeCollectionEntriesKeySchema != nil && schema.ChildNodesAssociativeCollectionEntriesValueSchema != nil {
				if convertedKey, err := n.convert(key, schema.ChildNodesAssociativeCollectionEntriesKeySchema, currentPathSegments); err == nil {
					if convertedValue, err := n.convert(val, schema.ChildNodesAssociativeCollectionEntriesValueSchema, currentPathSegments); err == nil {
						if convertedKey.IsValid() && convertedValue.IsValid() {
							newMap.SetMapIndex(convertedKey, convertedValue)
							continue
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("converted map entry for key %s not valid", key), schema, source.Interface(), currentPathSegments)
							}
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert value for map key %s failed", key), schema, val.Interface(), currentPathSegments)
						}
						continue
					}
				} else {
					if !schema.Nilable {
						return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert key for map key %s failed", key), schema, key, currentPathSegments)
					}
					continue
				}
			}

			return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("SchemaManip for map key %s not found", key), schema, source.Interface(), currentPathSegments)
		}

		return newMap, nil
	case reflect.Struct:
		newMap := reflect.MakeMap(schema.Type)

		for i := 0; i < source.NumField(); i++ {
			field := source.Field(i)
			key := source.Type().Field(i).Name
			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Key: key, IsKey: true})

			if childSchema, ok := schema.ChildNodes[key]; ok {
				switch cs := childSchema.(type) {
				case *DynamicSchema:
					if len(cs.Nodes) > 0 {
						for childNodeKey, childNode := range cs.Nodes {
							if convertedKey, err := n.convert(reflect.ValueOf(key), childNode.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
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
							return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("map entry with key %s not valid against any DynamicSchema nodes", key), cs, schema, currentPathSegments)
						}
					} else {
						return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("no DynamicSchema nodes found for key %s", key), cs, schema, currentPathSegments)
					}
				case *DynamicSchemaNode:
					if convertedKey, err := n.convert(reflect.ValueOf(key), cs.AssociativeCollectionEntryKeySchema, currentPathSegments); err == nil {
						if convertedValue, err := n.convertToMapWithDynamicSchemaNode(field, cs, currentPathSegments); err == nil {
							if convertedKey.IsValid() && convertedValue.IsValid() {
								newMap.SetMapIndex(convertedKey, convertedValue)
								continue
							} else {
								if !schema.Nilable {
									return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("converted map entry for key %s not valid", key), schema, source.Interface(), currentPathSegments)
								}
							}
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert value for map key %s failed", key), schema, field.Interface(), currentPathSegments)
							}
							continue
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert key for map key %s failed", key), schema, key, currentPathSegments)
						}
						continue
					}
				default:
					return reflect.Zero(schema.Type), NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, fmt.Sprintf("Nodes in SchemaManip for map key %s empty", key), childSchema, source.Interface(), currentPathSegments)
				}
			}

			if schema.ChildNodesAssociativeCollectionEntriesKeySchema != nil && schema.ChildNodesAssociativeCollectionEntriesValueSchema != nil {
				if convertedKey, err := n.convert(reflect.ValueOf(key), schema.ChildNodesAssociativeCollectionEntriesKeySchema, currentPathSegments); err == nil {
					if convertedValue, err := n.convert(field, schema.ChildNodesAssociativeCollectionEntriesValueSchema, currentPathSegments); err == nil {
						if convertedKey.IsValid() && convertedValue.IsValid() {
							newMap.SetMapIndex(convertedKey, convertedValue)
							continue
						} else {
							if !schema.Nilable {
								return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("converted map entry for key %s not valid", key), schema, source.Interface(), currentPathSegments)
							}
						}
					} else {
						if !schema.Nilable {
							return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert value for map key %s failed", key), schema, field.Interface(), currentPathSegments)
						}
						continue
					}
				} else {
					if !schema.Nilable {
						return reflect.Zero(schema.Type), NewError(err, FunctionName, fmt.Sprintf("convert key for map key %s failed", key), schema, key, currentPathSegments)
					}
					continue
				}
			}

			return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("SchemaManip for map key %s not found", key), schema, source.Interface(), currentPathSegments)
		}

		return newMap, nil
	case reflect.String:
		// Assumes source is json string
		var deserializedData interface{}
		err := json.Unmarshal([]byte(source.String()), &deserializedData)
		if err != nil {
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "failed to convert string to map using json", schema, source.Interface(), pathSegments)
		}
		return n.deserialize(reflect.ValueOf(deserializedData), schema, pathSegments)
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for map conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToArraySliceWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToArraySliceWithDynamicSchemaNode"

	if schema.Kind != reflect.Slice && schema.Kind != reflect.Array {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "data.Kind is not slice or array", schema, source.Interface(), pathSegments)
	}

	switch source.Kind() {
	case reflect.String:
		// Assumes source is json string
		var deserializedData interface{}
		err := json.Unmarshal([]byte(source.String()), &deserializedData)
		if err != nil {
			return reflect.Zero(schema.Type), NewError(err, FunctionName, "failed to convert string to array using json", schema, source.Interface(), pathSegments)
		}
		return n.deserialize(reflect.ValueOf(deserializedData), schema, pathSegments)
	case reflect.Slice, reflect.Array:
		var newArraySlice reflect.Value

		if len(schema.ChildNodes) == 0 && schema.ChildNodesLinearCollectionElementsSchema == nil {
			return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "no schema to convert element(s) in data (slice/array) found", schema, source.Interface(), pathSegments)
		}

		if schema.Kind == reflect.Slice {
			newArraySlice = reflect.MakeSlice(schema.Type, source.Len(), source.Len())
		} else {
			newArraySlice = reflect.New(schema.Type).Elem()
		}

		for i := 0; i < source.Len(); i++ {
			if i > newArraySlice.Len() {
				break
			}

			currentPathSegments := append(pathSegments, &path.CollectionMemberSegment{Index: i, IsIndex: true})
			childSchema, ok := schema.ChildNodes[fmt.Sprintf("%d", i)]
			if !ok {
				if schema.ChildNodesLinearCollectionElementsSchema != nil {
					childSchema = schema.ChildNodesLinearCollectionElementsSchema
				} else {
					return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, fmt.Sprintf("SchemaManip for slice/array index %d not found", i), schema, source.Interface(), currentPathSegments)
				}
			}

			if elementResult, err := n.convert(source.Index(i), childSchema, currentPathSegments); err == nil {
				if elementResult.IsValid() {
					newArraySlice.Index(i).Set(elementResult)
				} else {
					return elementResult, NewError(ErrDataConversionFailed, FunctionName, "elementResult not valid", schema, elementResult.Interface(), pathSegments)
				}
			} else {
				return elementResult, err
			}
		}

		return newArraySlice, nil
	default:
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "unsupported source.Kind for array/slice conversion", schema, source.Interface(), pathSegments)
	}
}

func (n *Processor) convertToPointerWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "convertToPointerWithDynamicSchemaNode"
	// The destination must be a pointer type.
	if schema.Kind != reflect.Pointer {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema.Kind not reflect.Pointer", schema, source.Interface(), pathSegments)
	}

	if schema.ChildNodesPointerSchema == nil {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "schema for value that data (pointer) points to has not been set (schema.ChildNodesPointerSchema is nil)", schema, source.Interface(), pathSegments)
	}

	pointerToResult, err := n.convert(source, schema.ChildNodesPointerSchema, pathSegments)
	if err != nil {
		return reflect.Zero(schema.Type), err
	}
	if !pointerToResult.IsValid() {
		return reflect.Zero(schema.Type), NewError(ErrDataConversionFailed, FunctionName, "pointerToResult not valid", schema, pointerToResult.Interface(), pathSegments)
	}

	newPtr := reflect.New(pointerToResult.Type())
	newPtr.Elem().Set(pointerToResult)
	return newPtr, nil
}
