package object

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

/*
Set updates or creates new value(s) in `Object.source`.

If `Object.schema` is supplied, it gives the function the ability to create user defined collections such as structs at different nesting levels.
Therefore, the `Object.source` can be instantiated as a value of type any with nil and end up being an array of nested structs.

Parameters:
  - jsonPath
  - value - value to insert or replace with.

Returns the number of modifications made through setting and the last error encountered.
*/
func (n *Object) Set(jsonPath path.JSONPath, value any) (uint64, error) {
	const FunctionName = "Set"

	if jsonPath == "$" || jsonPath == "" {
		n.source = reflect.ValueOf(value)
		return 1, nil
	}

	n.noOfResults = 0
	n.lastError = nil
	n.recursiveDescentSegments = jsonPath.Parse()
	n.valueToSet = value

	currentPathSegmentIndexes := internal.PathSegmentsIndexes{
		CurrentRecursive: 0,
		LastRecursive:    len(n.recursiveDescentSegments) - 1,
	}
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive {
		return 0, NewError().WithFunctionName(FunctionName).WithMessage("recursiveDescentSegments empty").WithNestedError(ErrPathSegmentInvalidError).WithData(core.JsonObject{"Source": n.source.Interface()})
	}
	currentPathSegmentIndexes.CurrentCollection = 0
	currentPathSegmentIndexes.LastCollection = len(n.recursiveDescentSegments[0]) - 1
	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return 0, NewError().WithFunctionName(FunctionName).WithMessage("recursiveDescentSegments empty").WithNestedError(ErrPathSegmentInvalidError).WithData(core.JsonObject{"Source": n.source.Interface()})
	}

	if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
		n.source = n.recursiveSet(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]}, n.sourceType)
	} else {
		n.source = n.recursiveDescentSet(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]})
	}

	if n.noOfResults > 0 {
		return n.noOfResults, nil
	}
	return n.noOfResults, n.lastError
}

func (n *Object) recursiveSet(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment, currentValueType reflect.Type) reflect.Value {
	const FunctionName = "recursiveSet"

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("indexes empty").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return currentValue
	}

	recursiveSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveSegment == nil {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("recursiveSegment is nil").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return currentValue
	}

	if recursiveSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				return reflect.ValueOf(n.valueToSet)
			}

			recursiveDescentIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: 0,
				LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
			}

			return n.recursiveDescentSet(currentValue, recursiveDescentIndexes, currentPath)
		}

		recursiveIndexes := internal.PathSegmentsIndexes{
			CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
			LastRecursive:     currentPathSegmentIndexes.LastRecursive,
			CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
			LastCollection:    currentPathSegmentIndexes.LastCollection,
		}
		return n.recursiveSet(currentValue, recursiveIndexes, currentPath, currentValueType)
	}

	if core.IsNilOrInvalid(currentValue) {
		if newValue, err := n.getDefaultValueAtPathSegment(currentValue, currentPathSegmentIndexes, currentPath, currentValueType); err == nil {
			currentValue = newValue
		} else {
			n.lastError = err
			return currentValue
		}

		if core.IsNilOrInvalid(currentValue) {
			n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("value nil or invalid").
				WithNestedError(ErrValueAtPathSegmentInvalidError).
				WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
			return currentValue
		}
	}

	if currentValue.Kind() == reflect.Interface {
		return n.recursiveSet(currentValue.Elem(), currentPathSegmentIndexes, currentPath, currentValue.Elem().Type())
	}

	if currentValue.Kind() == reflect.Ptr {
		recursiveDescentValue := n.recursiveSet(currentValue.Elem(), currentPathSegmentIndexes, currentPath, currentValue.Elem().Type())
		currentValue.Elem().Set(recursiveDescentValue)
		return currentValue
	}

	if mapKeyType, mapValueType, ok := core.GetMapKeyValueType(currentValue); ok {
		if recursiveSegment.IsKey {
			mapEntrySchema, _ := schema.GetSchemaAtPath(append(currentPath, recursiveSegment), n.schema)
			if mapEntrySchema == nil {
				mapEntrySchema = &schema.DynamicSchemaNode{
					Kind: mapValueType.Kind(),
					Type: mapValueType,
					AssociativeCollectionEntryKeySchema: &schema.DynamicSchemaNode{
						Kind: mapKeyType.Kind(),
						Type: mapKeyType,
					},
				}
			}
			if mapKeySchema, ok := mapEntrySchema.AssociativeCollectionEntryKeySchema.(*schema.DynamicSchemaNode); ok {
				if mapKey, err := n.convertSourceToTargetType(recursiveSegment.Key, mapKeySchema, mapKeyType); err == nil {
					mapKeyR := reflect.ValueOf(mapKey)
					mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
					if !mapValue.IsValid() {
						currentValue.SetMapIndex(mapKeyR, reflect.Zero(mapValueType))
						mapValue = currentValue.MapIndex(mapKeyR)
					}

					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
							if value, err := n.convertSourceToTargetType(n.valueToSet, mapEntrySchema, mapValueType); err == nil {
								currentValue.SetMapIndex(mapKeyR, reflect.ValueOf(value))
								n.noOfResults++
							}
						} else {
							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentDeleteValue := n.recursiveDescentSet(mapValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
							currentValue.SetMapIndex(mapKeyR, recursiveDescentDeleteValue)
						}
					} else {
						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveSetValue := n.recursiveSet(mapValue, recursiveIndexes, append(currentPath, recursiveSegment), mapValueType)
						currentValue.SetMapIndex(mapKeyR, recursiveSetValue)
					}
				} else {
					n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("convert key %s failed", recursiveSegment.Key)).
						WithNestedError(err).
						WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
				}
			} else {
				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("schema for key of entry %s not found", recursiveSegment.Key)).
					WithNestedError(ErrObjectError).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
			}
		} else if recursiveSegment.IsKeyIndexAll {
			for _, mapKey := range currentValue.MapKeys() {
				mapValue := currentValue.MapIndex(mapKey)
				if !mapValue.IsValid() {
					currentValue.SetMapIndex(mapKey, reflect.Zero(mapValueType))
					mapValue = currentValue.MapIndex(mapKey)
				}

				mapKeyPathSegment := &path.CollectionMemberSegment{
					IsKey: true,
					Key:   mapKeyString(mapKey),
				}

				mapEntrySchema, _ := schema.GetSchemaAtPath(append(currentPath, mapKeyPathSegment), n.schema)
				if mapEntrySchema == nil {
					mapEntrySchema = &schema.DynamicSchemaNode{
						Kind: mapValueType.Kind(),
						Type: mapValueType,
						AssociativeCollectionEntryKeySchema: &schema.DynamicSchemaNode{
							Kind: mapKeyType.Kind(),
							Type: mapKeyType,
						},
					}
				}
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						if value, err := n.convertSourceToTargetType(n.valueToSet, mapEntrySchema, mapValueType); err == nil {
							currentValue.SetMapIndex(mapKey, reflect.ValueOf(value))
							n.noOfResults++
						}
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentDeleteValue := n.recursiveDescentSet(mapValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
					currentValue.SetMapIndex(mapKey, recursiveDescentDeleteValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveSetValue := n.recursiveSet(mapValue, recursiveIndexes, append(currentPath, recursiveSegment), mapValueType)
				currentValue.SetMapIndex(mapKey, recursiveSetValue)
			}
		} else if len(recursiveSegment.UnionSelector) > 0 {
			for _, unionKey := range recursiveSegment.UnionSelector {
				if !unionKey.IsKey {
					continue
				}

				mapEntrySchema, _ := schema.GetSchemaAtPath(append(currentPath, unionKey), n.schema)
				if mapEntrySchema == nil {
					mapEntrySchema = &schema.DynamicSchemaNode{
						Kind: mapValueType.Kind(),
						Type: mapValueType,
						AssociativeCollectionEntryKeySchema: &schema.DynamicSchemaNode{
							Kind: mapKeyType.Kind(),
							Type: mapKeyType,
						},
					}
				}

				if mapKeySchema, ok := mapEntrySchema.AssociativeCollectionEntryKeySchema.(*schema.DynamicSchemaNode); ok {
					if mapKey, err := n.convertSourceToTargetType(unionKey.Key, mapKeySchema, mapKeyType); err == nil {
						mapKeyR := reflect.ValueOf(mapKey)
						mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
						if !mapValue.IsValid() {
							currentValue.SetMapIndex(mapKeyR, reflect.Zero(mapValueType))
							mapValue = currentValue.MapIndex(mapKeyR)
						}

						if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
							if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
								if value, err := n.convertSourceToTargetType(n.valueToSet, mapEntrySchema, mapValueType); err == nil {
									currentValue.SetMapIndex(mapKeyR, reflect.ValueOf(value))
									n.noOfResults++
								}
								continue
							}

							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentDeleteValue := n.recursiveDescentSet(mapValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
							currentValue.SetMapIndex(mapKeyR, recursiveDescentDeleteValue)
							continue
						}

						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveSetValue := n.recursiveSet(mapValue, recursiveIndexes, append(currentPath, recursiveSegment), mapValueType)
						currentValue.SetMapIndex(mapKeyR, recursiveSetValue)
					} else {
						n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("convert key %s failed", recursiveSegment.Key)).
							WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
					}
					continue
				}

				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("schema for key of entry %s not found", recursiveSegment.Key)).
					WithNestedError(ErrObjectError).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
			}
		} else {
			n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in map, unsupported recursive segment %s", recursiveSegment)).
				WithNestedError(ErrPathSegmentInvalidError).
				WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		}
	} else if arraySliceType, ok := core.GetArraySliceValueType(currentValue); ok {
		if recursiveSegment.IsIndex {
			if recursiveSegment.Index > currentValue.Len()-1 && currentValue.Kind() == reflect.Slice {
				for i := currentValue.Len(); i <= recursiveSegment.Index; i++ {
					currentValue = reflect.Append(currentValue, reflect.Zero(arraySliceType))
				}
			}

			if recursiveSegment.Index >= currentValue.Len() {
				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in array/slice, index %s out of range", recursiveSegment)).
					WithNestedError(ErrValueAtPathSegmentInvalidError).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
			} else {
				arraySliceValue := currentValue.Index(recursiveSegment.Index)
				if arraySliceValue.CanSet() {
					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
							arraySliceSchema, _ := schema.GetSchemaAtPath(append(currentPath, recursiveSegment), n.schema)
							if value, err := n.convertSourceToTargetType(n.valueToSet, arraySliceSchema, arraySliceType); err == nil {
								arraySliceValue.Set(reflect.ValueOf(value))
								n.noOfResults++
							}
						} else {
							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentValue := n.recursiveDescentSet(arraySliceValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
							arraySliceValue.Set(recursiveDescentValue)
						}
					} else {
						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveValue := n.recursiveSet(arraySliceValue, recursiveIndexes, append(currentPath, recursiveSegment), arraySliceType)
						arraySliceValue.Set(recursiveValue)
					}
				}
			}
		} else if recursiveSegment.IsKeyIndexAll {
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection && currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				if currentValue.Kind() == reflect.Array {
					currentValue = reflect.New(currentValue.Type()).Elem()
				} else {
					currentValue = reflect.MakeSlice(currentValue.Type(), 0, 0)
				}
				n.noOfResults++
			} else {
				for i := 0; i < currentValue.Len(); i++ {
					arraySliceValue := currentValue.Index(i)
					if !arraySliceValue.CanSet() {
						continue
					}

					collectionMemberSegment := &path.CollectionMemberSegment{IsIndex: true, Index: i}
					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
							arraySliceSchema, _ := schema.GetSchemaAtPath(append(currentPath, collectionMemberSegment), n.schema)
							if value, err := n.convertSourceToTargetType(n.valueToSet, arraySliceSchema, arraySliceValue.Type()); err == nil {
								arraySliceValue.Set(reflect.ValueOf(value))
								n.noOfResults++
							}
						}

						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentSet(arraySliceValue, recursiveDescentIndexes, append(currentPath, collectionMemberSegment))
						arraySliceValue.Set(recursiveDescentValue)
						continue
					}

					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveSet(arraySliceValue, recursiveIndexes, append(currentPath, collectionMemberSegment), arraySliceType)
					arraySliceValue.Set(recursiveValue)
				}
			}
		} else if len(recursiveSegment.UnionSelector) > 0 {
			maxIndex := -1

			for _, unionKey := range recursiveSegment.UnionSelector {
				if !unionKey.IsIndex {
					continue
				}
				if unionKey.Index > maxIndex {
					maxIndex = unionKey.Index
				}
			}

			if maxIndex >= 0 {
				if maxIndex > currentValue.Len()-1 && currentValue.Kind() == reflect.Slice {
					for i := currentValue.Len(); i <= maxIndex; i++ {
						currentValue = reflect.Append(currentValue, reflect.Zero(arraySliceType))
					}
				}
			}

			for _, unionKey := range recursiveSegment.UnionSelector {
				if !unionKey.IsIndex || unionKey.Index >= currentValue.Len() {
					continue
				}

				arraySliceValue := currentValue.Index(unionKey.Index)
				if !arraySliceValue.CanSet() || !arraySliceValue.IsValid() {
					continue
				}

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						arraySliceSchema, _ := schema.GetSchemaAtPath(append(currentPath, unionKey), n.schema)
						if value, err := n.convertSourceToTargetType(n.valueToSet, arraySliceSchema, arraySliceValue.Type()); err == nil {
							arraySliceValue.Set(reflect.ValueOf(value))
							n.noOfResults++
						}
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentValue := n.recursiveDescentSet(arraySliceValue, recursiveDescentIndexes, append(currentPath, unionKey))
					arraySliceValue.Set(recursiveDescentValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveValue := n.recursiveSet(arraySliceValue, recursiveIndexes, append(currentPath, unionKey), arraySliceType)
				arraySliceValue.Set(recursiveValue)
			}
		} else if recursiveSegment.LinearCollectionSelector != nil {
			start := 0
			if recursiveSegment.LinearCollectionSelector.IsStart {
				start = recursiveSegment.LinearCollectionSelector.Start
			}
			step := 1
			if recursiveSegment.LinearCollectionSelector.IsStep && recursiveSegment.LinearCollectionSelector.Step > 0 {
				step = recursiveSegment.LinearCollectionSelector.Step
			}
			end := currentValue.Len()
			if recursiveSegment.LinearCollectionSelector.IsEnd {
				end = recursiveSegment.LinearCollectionSelector.End
			}

			if end > currentValue.Len() && currentValue.Kind() == reflect.Slice {
				for i := currentValue.Len(); i <= end; i++ {
					currentValue = reflect.Append(currentValue, reflect.Zero(arraySliceType))
				}
			}

			for i := start; i < end; i += step {
				if i >= currentValue.Len() {
					continue
				}

				arraySliceValue := currentValue.Index(i)
				if !arraySliceValue.CanSet() || !arraySliceValue.IsValid() {
					continue
				}

				collectionMemberSegment := &path.CollectionMemberSegment{IsIndex: true, Index: i}
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						arraySliceSchema, _ := schema.GetSchemaAtPath(append(currentPath, collectionMemberSegment), n.schema)
						if value, err := n.convertSourceToTargetType(n.valueToSet, arraySliceSchema, arraySliceValue.Type()); err == nil {
							arraySliceValue.Set(reflect.ValueOf(value))
							n.noOfResults++
						}
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentValue := n.recursiveDescentSet(arraySliceValue, recursiveDescentIndexes, append(currentPath, collectionMemberSegment))
					arraySliceValue.Set(recursiveDescentValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveValue := n.recursiveSet(arraySliceValue, recursiveIndexes, append(currentPath, collectionMemberSegment), arraySliceType)
				arraySliceValue.Set(recursiveValue)
			}
		} else {
			n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in array/slice, unsupported recursive segment %s", recursiveSegment)).
				WithNestedError(ErrPathSegmentInvalidError).
				WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		}
	} else if currentValue.Kind() == reflect.Struct {
		if recursiveSegment.IsKey {
			if !core.StartsWithCapital(recursiveSegment.Key) {
				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("key %v is not valid for struct", recursiveSegment)).
					WithNestedError(ErrPathSegmentInvalidError).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
			} else {
				structFieldValue := currentValue.FieldByName(recursiveSegment.Key)
				if structFieldValue.CanSet() {
					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
							structFieldSchema, _ := schema.GetSchemaAtPath(append(currentPath, recursiveSegment), n.schema)
							if value, err := n.convertSourceToTargetType(n.valueToSet, structFieldSchema, structFieldValue.Type()); err == nil {
								structFieldValue.Set(reflect.ValueOf(value))
								n.noOfResults++
							}
						} else {
							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentDeleteValue := n.recursiveDescentSet(structFieldValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
							structFieldValue.Set(recursiveDescentDeleteValue)
						}
					} else {
						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveSetValue := n.recursiveSet(structFieldValue, recursiveIndexes, append(currentPath, recursiveSegment), structFieldValue.Type())
						structFieldValue.Set(recursiveSetValue)
					}
				}
			}
		} else if recursiveSegment.IsKeyIndexAll {
			for i := 0; i < currentValue.NumField(); i++ {
				structField := currentValue.Field(i)

				if !structField.CanSet() {
					continue
				}

				structFieldSegment := &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						structFieldSchema, _ := schema.GetSchemaAtPath(append(currentPath, structFieldSegment), n.schema)
						if value, err := n.convertSourceToTargetType(n.valueToSet, structFieldSchema, structField.Type()); err == nil {
							structField.Set(reflect.ValueOf(value))
							n.noOfResults++
						}
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentValue := n.recursiveDescentSet(structField, recursiveDescentIndexes, append(currentPath, structFieldSegment))
					structField.Set(recursiveDescentValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveValue := n.recursiveSet(structField, recursiveIndexes, append(currentPath, structFieldSegment), structField.Type())
				structField.Set(recursiveValue)
			}
		} else if len(recursiveSegment.UnionSelector) > 0 {
			for _, unionKey := range recursiveSegment.UnionSelector {
				if !unionKey.IsKey || !core.StartsWithCapital(unionKey.Key) {
					continue
				}

				structFieldValue := currentValue.FieldByName(unionKey.Key)
				if !structFieldValue.CanSet() {
					continue
				}

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						structFieldSchema, _ := schema.GetSchemaAtPath(append(currentPath, unionKey), n.schema)
						if value, err := n.convertSourceToTargetType(n.valueToSet, structFieldSchema, structFieldValue.Type()); err == nil {
							structFieldValue.Set(reflect.ValueOf(value))
							n.noOfResults++
						}
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentDeleteValue := n.recursiveDescentSet(structFieldValue, recursiveDescentIndexes, append(currentPath, unionKey))
					structFieldValue.Set(recursiveDescentDeleteValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveDeleteValue := n.recursiveSet(structFieldValue, recursiveIndexes, append(currentPath, unionKey), structFieldValue.Type())
				structFieldValue.Set(recursiveDeleteValue)
			}
		} else {
			n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in struct, unsupported recursive segment %s", recursiveSegment)).
				WithNestedError(ErrPathSegmentInvalidError).
				WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		}
	} else {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("unsupported value at recursive segment %s", recursiveSegment)).
			WithNestedError(ErrValueAtPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
	}

	return currentValue
}

func (n *Object) getDefaultValueAtPathSegment(value reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment, valueType reflect.Type) (reflect.Value, error) {
	const FunctionName = "getDefaultValueAtPathSegment"

	valueSchema, err := schema.GetSchemaAtPath(currentPath, n.schema)
	if err == nil {
		if valueSchema.IsDefaultValueSet {
			return valueSchema.DefaultValue(), nil
		}
		if valueSchema.Kind != reflect.Interface {
			valueType = valueSchema.Type
		}
	}

	var newValue reflect.Value

	if valueType == nil {
		currentPathSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
		if currentPathSegment == nil {
			return value, NewError().WithFunctionName(FunctionName).WithMessage("indexes empty").
				WithNestedError(ErrPathSegmentInvalidError).
				WithData(core.JsonObject{"CurrentValue": value.Interface(), "CurrentPathSegment": currentPath})
		}

		if currentPathSegment.IsIndex || (len(currentPathSegment.UnionSelector) > 0 && currentPathSegment.UnionSelector[0].IsIndex) || currentPathSegment.LinearCollectionSelector != nil {
			newValue = reflect.MakeSlice(reflect.TypeOf(make([]any, 0)), 0, 0)
		} else {
			// Define the reflect.Type for the key (string)
			keyType := reflect.TypeOf("")

			// Define the reflect.Type for the value (any/interface{})
			valueType := reflect.TypeOf((*interface{})(nil)).Elem()

			// Create the map type using reflect.MapOf
			mapType := reflect.MapOf(keyType, valueType)

			newValue = reflect.MakeMap(mapType)
		}
	} else {
		switch valueType.Kind() {
		case reflect.Interface, reflect.Invalid:
			currentPathSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
			if currentPathSegment == nil {
				return value, NewError().WithFunctionName(FunctionName).WithMessage("indexes empty").
					WithNestedError(ErrPathSegmentInvalidError).
					WithData(core.JsonObject{"CurrentValue": value.Interface(), "CurrentPathSegment": currentPath})
			}

			if currentPathSegment.IsIndex || (len(currentPathSegment.UnionSelector) > 0 && currentPathSegment.UnionSelector[0].IsIndex) || currentPathSegment.LinearCollectionSelector != nil {
				newValue = reflect.MakeSlice(reflect.TypeOf(make([]any, 0)), 0, 0)
			} else {
				newValue = reflect.ValueOf(map[string]any{})
			}
		case reflect.Struct:
			// reflect.New returns a pointer, so we need Elem() to get the struct itself.
			newValue = reflect.New(valueType).Elem()
		case reflect.Map:
			// reflect.MakeMap creates a new, non-nil map.
			newValue = reflect.MakeMap(valueType)
		case reflect.Slice:
			// reflect.MakeSlice creates a new, non-nil slice with length 0.
			newValue = reflect.MakeSlice(valueType, 0, 0)
		case reflect.Pointer:
			// reflect.New creates a new zero-value pointer.
			newValue = reflect.New(valueType.Elem())
		default:
			// For other types, reflect.New() can create a zero value.
			newValue = reflect.New(valueType).Elem()
		}
	}

	return newValue, nil
}

func (n *Object) recursiveDescentSet(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) reflect.Value {
	const FunctionName = "recursiveDescentSet"

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("indexes empty").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return currentValue
	}

	recursiveDescentSearchSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveDescentSearchSegment == nil {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("recursive descent search segment is nil").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return currentValue
	}

	if core.IsNilOrInvalid(currentValue) {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("current value nil or invalid").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return currentValue
	}

	if recursiveDescentSearchSegment.IsKeyRoot {
		return n.recursiveSet(currentValue, currentPathSegmentIndexes, currentPath, currentValue.Type())
	}

	if !recursiveDescentSearchSegment.IsKey {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("recursive descent search segment %s is not key", recursiveDescentSearchSegment)).
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return currentValue
	}

	if currentValue.Kind() == reflect.Interface {
		return n.recursiveDescentSet(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if currentValue.Kind() == reflect.Ptr {
		recursiveDescentValue := n.recursiveDescentSet(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
		currentValue.Elem().Set(recursiveDescentValue)
		return currentValue
	}

	if _, mapValueType, ok := core.GetMapKeyValueType(currentValue); ok {
		for _, mapKey := range currentValue.MapKeys() {
			mapValue := currentValue.MapIndex(mapKey)
			if !mapValue.IsValid() {
				continue
			}

			keyPathSegment := &path.CollectionMemberSegment{
				IsKey: true,
				Key:   mapKeyString(mapKey),
			}

			if keyPathSegment.Key == recursiveDescentSearchSegment.Key {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						mapValueSchema, _ := schema.GetSchemaAtPath(append(currentPath, recursiveDescentSearchSegment), n.schema)
						if value, err := n.convertSourceToTargetType(n.valueToSet, mapValueSchema, mapValueType); err == nil {
							currentValue.SetMapIndex(mapKey, reflect.ValueOf(value))
							n.noOfResults++
						}
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentSet(mapValue, recursiveDescentIndexes, append(currentPath, keyPathSegment))
						currentValue.SetMapIndex(mapKey, recursiveDescentValue)
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveSet(mapValue, recursiveIndexes, append(currentPath, keyPathSegment), mapValue.Type())
					currentValue.SetMapIndex(mapKey, recursiveValue)
				}
			} else {
				recursiveDescentValue := n.recursiveDescentSet(mapValue, currentPathSegmentIndexes, append(currentPath, keyPathSegment))
				currentValue.SetMapIndex(mapKey, recursiveDescentValue)
			}
		}
	} else if _, ok := core.GetArraySliceValueType(currentValue); ok {
		for i := 0; i < currentValue.Len(); i++ {
			sliceArrayValue := currentValue.Index(i)
			if !sliceArrayValue.IsValid() {
				continue
			}

			if currentValue.Index(i).CanSet() {
				recursiveDescentValue := n.recursiveDescentSet(sliceArrayValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
				currentValue.Index(i).Set(recursiveDescentValue)
			}
		}
	} else if currentValue.Kind() == reflect.Struct {
		if core.StartsWithCapital(recursiveDescentSearchSegment.Key) {
			if structFieldValue := currentValue.FieldByName(recursiveDescentSearchSegment.Key); structFieldValue.IsValid() && structFieldValue.CanSet() {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						structFieldSchema, _ := schema.GetSchemaAtPath(append(currentPath, recursiveDescentSearchSegment), n.schema)
						if value, err := n.convertSourceToTargetType(n.valueToSet, structFieldSchema, structFieldValue.Type()); err == nil {
							structFieldValue.Set(reflect.ValueOf(value))
							n.noOfResults++
						}
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentSet(structFieldValue, recursiveDescentIndexes, append(currentPath, recursiveDescentSearchSegment))
						structFieldValue.Set(recursiveDescentValue)
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveSet(structFieldValue, recursiveIndexes, append(currentPath, recursiveDescentSearchSegment), structFieldValue.Type())
					structFieldValue.Set(recursiveValue)
				}
			}
		}

		for i := 0; i < currentValue.NumField(); i++ {
			if !core.IsStructFieldExported(currentValue.Type().Field(i)) {
				continue
			}

			structFieldValue := currentValue.Field(i)
			if !structFieldValue.IsValid() {
				continue
			}

			if structFieldValue.CanSet() {
				recursiveDescentValue := n.recursiveDescentSet(structFieldValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}))
				structFieldValue.Set(recursiveDescentValue)
			}
		}
	} else {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("unsupported value at recursive descent search segment %s", recursiveDescentSearchSegment)).
			WithNestedError(ErrValueAtPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
	}

	return currentValue
}

func (n *Object) convertSourceToTargetType(source any, sourceSchema *schema.DynamicSchemaNode, sourceType reflect.Type) (any, error) {
	const FunctionName = "convertSourceToTargetType"

	if (sourceSchema == nil || sourceSchema.Kind == reflect.Interface) && sourceType != nil {
		sourceSchema = &schema.DynamicSchemaNode{
			Kind: sourceType.Kind(),
			Type: sourceType,
		}
	}

	if sourceSchema != nil {
		var destination any
		if err := n.defaultConverter.Convert(source, sourceSchema, &destination); err != nil {
			return nil, err
		}
		return destination, nil
	}

	return nil, NewError().WithFunctionName(FunctionName).WithMessage("source schema not found").WithNestedError(ErrObjectError)
}
