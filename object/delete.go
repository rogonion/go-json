package object

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

/*
Delete removes value(s) in `Object.source`.

Parameters:
  - jsonPath - path to data to remove.

Returns the number of modifications made through deletion and the last error encountered.
*/
func (n *Object) Delete(jsonPath path.JSONPath) (uint64, error) {
	const FunctionName = "Delete"

	n.noOfModifications = 0
	n.lastError = nil

	if jsonPath == "$" || jsonPath == "" {
		n.source = reflect.Zero(n.source.Type())
		return 1, nil
	}

	n.recursiveDescentSegments = jsonPath.Parse()

	currentPathSegmentIndexes := internal.PathSegmentsIndexes{
		CurrentRecursive: 0,
		LastRecursive:    len(n.recursiveDescentSegments) - 1,
	}
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive {
		return 0, NewError(FunctionName, "recursiveDescentSegments empty").WithData(n.source.Interface()).WithNestedError(ErrPathSegmentInvalidError)
	}
	currentPathSegmentIndexes.CurrentCollection = 0
	currentPathSegmentIndexes.LastCollection = len(n.recursiveDescentSegments[0]) - 1
	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return 0, NewError(FunctionName, "recursiveDescentSegments empty").WithData(n.source.Interface()).WithNestedError(ErrPathSegmentInvalidError)
	}

	if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
		n.source = n.recursiveDelete(n.source, currentPathSegmentIndexes, make(path.RecursiveDescentSegment, 0))
	} else {
		n.source = n.recursiveDescentDelete(n.source, currentPathSegmentIndexes, make(path.RecursiveDescentSegment, 0))
	}

	if n.noOfModifications > 0 {
		return n.noOfModifications, nil
	}
	return n.noOfModifications, n.lastError
}

func (n *Object) recursiveDelete(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) reflect.Value {
	const FunctionName = "recursiveDelete"

	recursiveSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveSegment == nil {
		n.lastError = NewError(FunctionName, "recursiveSegment is nil").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		return currentValue
	}

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		n.lastError = NewError(FunctionName, "indexes empty").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		return currentValue
	}

	if core.IsNilOrInvalid(currentValue) {
		n.lastError = NewError(FunctionName, "value nil or invalid").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		return currentValue
	}

	if recursiveSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				return reflect.Zero(currentValue.Type())
			}

			recursiveDescentIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: 0,
				LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
			}

			return n.recursiveDescentDelete(currentValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
		}

		recursiveIndexes := internal.PathSegmentsIndexes{
			CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
			LastRecursive:     currentPathSegmentIndexes.LastRecursive,
			CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
			LastCollection:    currentPathSegmentIndexes.LastCollection,
		}
		return n.recursiveDelete(currentValue, recursiveIndexes, append(currentPath, recursiveSegment))
	}

	if currentValue.Kind() == reflect.Interface {
		return n.recursiveDelete(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if currentValue.Kind() == reflect.Ptr {
		recursiveDescentValue := n.recursiveDelete(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
		currentValue.Elem().Set(recursiveDescentValue)
		return currentValue
	}

	if mapKeyType, _, ok := core.GetMapKeyValueType(currentValue); ok {
		if recursiveSegment.IsKey {
			var mapKey any
			if err := n.defaultConverter.Convert(recursiveSegment.Key, &schema.DynamicSchemaNode{Kind: mapKeyType.Kind(), Type: mapKeyType}, &mapKey); err != nil {
				n.lastError = NewError(FunctionName, fmt.Sprintf("convert mapKey %s to type %v failed", recursiveSegment, mapKeyType)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(err)
			} else {
				mapKeyR := reflect.ValueOf(mapKey)
				mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						currentValue.SetMapIndex(mapKeyR, reflect.Value{})
						n.noOfModifications++
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentDelete(mapValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
						currentValue.SetMapIndex(mapKeyR, recursiveDescentValue)
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveDelete(mapValue, recursiveIndexes, append(currentPath, recursiveSegment))
					currentValue.SetMapIndex(mapKeyR, recursiveValue)
				}
			}
		} else if recursiveSegment.IsKeyIndexAll {
			for _, mapKey := range currentValue.MapKeys() {
				mapValue := currentValue.MapIndex(mapKey)
				if !mapValue.IsValid() {
					continue
				}

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						currentValue.SetMapIndex(mapKey, reflect.Value{})
						n.noOfModifications++
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentValue := n.recursiveDescentDelete(mapValue, recursiveDescentIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: fmt.Sprintf("%v", core.JsonStringifyMust(mapKey.Interface()))}))
					currentValue.SetMapIndex(mapKey, recursiveDescentValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveValue := n.recursiveDelete(mapValue, recursiveIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: fmt.Sprintf("%v", core.JsonStringifyMust(mapKey.Interface()))}))
				currentValue.SetMapIndex(mapKey, recursiveValue)
			}
		} else if len(recursiveSegment.UnionSelector) > 0 {
			for _, unionKey := range recursiveSegment.UnionSelector {
				if !unionKey.IsKey {
					continue
				}

				var mapKey any
				if err := n.defaultConverter.Convert(unionKey.Key, &schema.DynamicSchemaNode{Kind: mapKeyType.Kind(), Type: mapKeyType}, &mapKey); err != nil {
					n.lastError = NewError(FunctionName, fmt.Sprintf("convert key %s to type %v failed", unionKey.Key, mapKeyType)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(err)
					continue
				}
				mapKeyR := reflect.ValueOf(mapKey)

				mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
				if !mapValue.IsValid() {
					continue
				}

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						currentValue.SetMapIndex(mapKeyR, reflect.Value{})
						n.noOfModifications++
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentValue := n.recursiveDescentDelete(mapValue, recursiveDescentIndexes, append(currentPath, unionKey))
					currentValue.SetMapIndex(mapKeyR, recursiveDescentValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveValue := n.recursiveDelete(mapValue, recursiveIndexes, append(currentPath, unionKey))
				currentValue.SetMapIndex(mapKeyR, recursiveValue)
			}
		} else {
			n.lastError = NewError(FunctionName, fmt.Sprintf("in map, unsupported recursive segment %s", recursiveSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		}
	} else if arraySliceType, ok := core.GetArraySliceValueType(currentValue); ok {
		if recursiveSegment.IsIndex {
			if recursiveSegment.Index >= currentValue.Len() {
				n.lastError = NewError(FunctionName, fmt.Sprintf("in linear collection, index %s out of range", recursiveSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrValueAtPathSegmentInvalidError)
			} else {
				arraySliceValue := currentValue.Index(recursiveSegment.Index)
				if arraySliceValue.IsValid() && arraySliceValue.CanSet() {
					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
							if currentValue.Kind() == reflect.Array {
								arraySliceValue.Set(reflect.Zero(arraySliceType))
							} else {
								newSlice := reflect.MakeSlice(currentValue.Type(), currentValue.Len()-1, currentValue.Len()-1)
								skip := 0
								for i := 0; i < currentValue.Len(); i++ {
									if i == recursiveSegment.Index {
										skip++
										continue
									}
									newSlice.Index(i - skip).Set(currentValue.Index(i))
								}
								currentValue = newSlice
							}
							n.noOfModifications++
						} else {
							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentValue := n.recursiveDescentDelete(arraySliceValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
							arraySliceValue.Set(recursiveDescentValue)
						}
					} else {
						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveValue := n.recursiveDelete(arraySliceValue, recursiveIndexes, append(currentPath, recursiveSegment))
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
				n.noOfModifications++
			} else {
				for i := 0; i < currentValue.Len(); i++ {
					arraySliceValue := currentValue.Index(i)
					if !arraySliceValue.CanSet() || !arraySliceValue.IsValid() {
						continue
					}

					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentDelete(arraySliceValue, recursiveDescentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
						arraySliceValue.Set(recursiveDescentValue)
						continue
					}

					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveDelete(arraySliceValue, recursiveIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
					arraySliceValue.Set(recursiveValue)
				}
			}
		} else if len(recursiveSegment.UnionSelector) > 0 {
			if currentValue.Kind() == reflect.Array || (currentPathSegmentIndexes.CurrentCollection != currentPathSegmentIndexes.LastCollection || currentPathSegmentIndexes.CurrentRecursive != currentPathSegmentIndexes.LastRecursive) {
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
							arraySliceValue.Set(reflect.Zero(arraySliceType))
							n.noOfModifications++
							continue
						}

						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentDelete(arraySliceValue, recursiveDescentIndexes, append(currentPath, unionKey))
						arraySliceValue.Set(recursiveDescentValue)
						continue
					}

					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveDelete(arraySliceValue, recursiveIndexes, append(currentPath, unionKey))
					arraySliceValue.Set(recursiveValue)
				}
			} else if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection && currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				indexesToExclude := make([]int, 0)
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsIndex || unionKey.Index >= currentValue.Len() {
						continue
					}
					indexesToExclude = append(indexesToExclude, unionKey.Index)
				}
				newSlice := reflect.MakeSlice(currentValue.Type(), currentValue.Len()-len(indexesToExclude), currentValue.Len()-len(indexesToExclude))
				skip := 0
				for i := 0; i < currentValue.Len(); i++ {
					if slices.Contains(indexesToExclude, i) {
						skip++
						continue
					}
					newSlice.Index(i - skip).Set(currentValue.Index(i))
				}
				currentValue = newSlice
			} else {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsIndex || unionKey.Index >= currentValue.Len() {
						continue
					}

					arraySliceValue := currentValue.Index(unionKey.Index)
					if !arraySliceValue.CanSet() || !arraySliceValue.IsValid() {
						continue
					}

					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentDelete(arraySliceValue, recursiveDescentIndexes, append(currentPath, unionKey))
						arraySliceValue.Set(recursiveDescentValue)
						continue
					}

					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveDelete(arraySliceValue, recursiveIndexes, append(currentPath, unionKey))
					arraySliceValue.Set(recursiveValue)
				}
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

			if end > currentValue.Len() {
				n.lastError = NewError(FunctionName, fmt.Sprintf("in linear collection, linear collection selector %s End is out of range", recursiveSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
			} else {
				if currentValue.Kind() == reflect.Array || (currentPathSegmentIndexes.CurrentCollection != currentPathSegmentIndexes.LastCollection || currentPathSegmentIndexes.CurrentRecursive != currentPathSegmentIndexes.LastRecursive) {
					for i := start; i < end; i += step {
						if i >= currentValue.Len() {
							continue
						}

						valueFromSliceArray := currentValue.Index(i)
						if !valueFromSliceArray.CanSet() || !valueFromSliceArray.IsValid() {
							continue
						}

						if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
							if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
								valueFromSliceArray.Set(reflect.Zero(arraySliceType))
								n.noOfModifications++
								continue
							}

							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentValue := n.recursiveDescentDelete(valueFromSliceArray, recursiveDescentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
							valueFromSliceArray.Set(recursiveDescentValue)
							continue
						}

						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveValue := n.recursiveDelete(valueFromSliceArray, recursiveIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
						valueFromSliceArray.Set(recursiveValue)
					}
				} else if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection && currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					indexesToExclude := make([]int, 0)
					for i := start; i < end; i += step {
						if i >= currentValue.Len() {
							continue
						}

						indexesToExclude = append(indexesToExclude, i)
					}
					newSlice := reflect.MakeSlice(currentValue.Type(), currentValue.Len()-len(indexesToExclude), currentValue.Len()-len(indexesToExclude))
					skip := 0
					for i := 0; i < currentValue.Len(); i++ {
						if slices.Contains(indexesToExclude, i) {
							skip++
							continue
						}
						newSlice.Index(i - skip).Set(currentValue.Index(i))
					}
					currentValue = newSlice
				} else {
					for i := start; i < end; i += step {
						if i >= currentValue.Len() {
							continue
						}

						arraySliceValue := currentValue.Index(i)
						if !arraySliceValue.CanSet() || !arraySliceValue.IsValid() {
							continue
						}

						if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentValue := n.recursiveDescentDelete(arraySliceValue, recursiveDescentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
							arraySliceValue.Set(recursiveDescentValue)
							continue
						}

						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveValue := n.recursiveDelete(arraySliceValue, recursiveIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
						arraySliceValue.Set(recursiveValue)
					}
				}
			}
		} else {
			n.lastError = NewError(FunctionName, fmt.Sprintf("in linear collection, unsupported recursive segment %s", recursiveSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		}
	} else if currentValue.Kind() == reflect.Struct {
		if recursiveSegment.IsKey {
			if !core.StartsWithCapital(recursiveSegment.Key) {
				n.lastError = NewError(FunctionName, fmt.Sprintf("key %v is not valid for struct", recursiveSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
			} else {
				structFieldValue := currentValue.FieldByName(recursiveSegment.Key)
				if structFieldValue.IsValid() && structFieldValue.CanSet() {
					if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
						if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
							structFieldValue.Set(reflect.Zero(structFieldValue.Type()))
							n.noOfModifications++
						} else {
							recursiveDescentIndexes := internal.PathSegmentsIndexes{
								CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
								LastRecursive:     currentPathSegmentIndexes.LastRecursive,
								CurrentCollection: 0,
								LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
							}

							recursiveDescentDeleteValue := n.recursiveDescentDelete(structFieldValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
							structFieldValue.Set(recursiveDescentDeleteValue)
						}
					} else {
						recursiveIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
							LastCollection:    currentPathSegmentIndexes.LastCollection,
						}

						recursiveSetValue := n.recursiveDelete(structFieldValue, recursiveIndexes, append(currentPath, recursiveSegment))
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

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						structField.Set(reflect.Zero(structField.Type()))
						n.noOfModifications++
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentValue := n.recursiveDescentDelete(structField, recursiveDescentIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}))
					structField.Set(recursiveDescentValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveValue := n.recursiveDelete(structField, recursiveIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}))
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
						structFieldValue.Set(reflect.Zero(structFieldValue.Type()))
						n.noOfModifications++
						continue
					}

					recursiveDescentIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: 0,
						LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
					}

					recursiveDescentDeleteValue := n.recursiveDescentDelete(structFieldValue, recursiveDescentIndexes, append(currentPath, unionKey))
					structFieldValue.Set(recursiveDescentDeleteValue)
					continue
				}

				recursiveIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
					LastCollection:    currentPathSegmentIndexes.LastCollection,
				}

				recursiveDeleteValue := n.recursiveDelete(structFieldValue, recursiveIndexes, append(currentPath, unionKey))
				structFieldValue.Set(recursiveDeleteValue)
			}
		} else {
			n.lastError = NewError(FunctionName, fmt.Sprintf("in struct, unsupported recursive segment %s", recursiveSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		}
	} else {
		n.lastError = NewError(FunctionName, "unsupported value at recursive segment").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrValueAtPathSegmentInvalidError)
	}

	return currentValue
}

func (n *Object) recursiveDescentDelete(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) reflect.Value {
	const FunctionName = "recursiveDescentDelete"

	recursiveDescentSearchSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveDescentSearchSegment == nil {
		n.lastError = NewError(FunctionName, "recursive descent search segment is nil").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		return currentValue
	}

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		n.lastError = NewError(FunctionName, "indexes empty").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		return currentValue
	}

	if core.IsNilOrInvalid(currentValue) {
		n.lastError = NewError(FunctionName, "value nil or invalid").WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrValueAtPathSegmentInvalidError)
		return currentValue
	}

	if recursiveDescentSearchSegment.IsKeyRoot {
		return n.recursiveDelete(currentValue, currentPathSegmentIndexes, currentPath)
	}

	if !recursiveDescentSearchSegment.IsKey {
		n.lastError = NewError(FunctionName, fmt.Sprintf("recursive descent search segment %s is not key", recursiveDescentSearchSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrPathSegmentInvalidError)
		return currentValue
	}

	if currentValue.Kind() == reflect.Interface {
		return n.recursiveDescentDelete(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if currentValue.Kind() == reflect.Ptr {
		recursiveDescentValue := n.recursiveDescentDelete(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
		currentValue.Elem().Set(recursiveDescentValue)
		return currentValue
	}

	if _, _, ok := core.GetMapKeyValueType(currentValue); ok {
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
						currentValue.SetMapIndex(mapKey, reflect.Value{})
						n.noOfModifications++
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentDelete(mapValue, recursiveDescentIndexes, append(currentPath, keyPathSegment))
						currentValue.SetMapIndex(mapKey, recursiveDescentValue)
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveDelete(mapValue, recursiveIndexes, append(currentPath, keyPathSegment))
					currentValue.SetMapIndex(mapKey, recursiveValue)
				}
			} else {
				recursiveDescentValue := n.recursiveDescentDelete(mapValue, currentPathSegmentIndexes, append(currentPath, keyPathSegment))
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
				recursiveDescentValue := n.recursiveDescentDelete(sliceArrayValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
				currentValue.Index(i).Set(recursiveDescentValue)
			}
		}
	} else if currentValue.Kind() == reflect.Struct {
		if core.StartsWithCapital(recursiveDescentSearchSegment.Key) {
			if structFieldValue := currentValue.FieldByName(recursiveDescentSearchSegment.Key); structFieldValue.IsValid() && structFieldValue.CanSet() {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						structFieldValue.Set(reflect.Zero(structFieldValue.Type()))
						n.noOfModifications++
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentDelete(structFieldValue, recursiveDescentIndexes, append(currentPath, recursiveDescentSearchSegment))
						structFieldValue.Set(recursiveDescentValue)
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveDelete(structFieldValue, recursiveIndexes, append(currentPath, recursiveDescentSearchSegment))
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
				recursiveDescentValue := n.recursiveDescentDelete(structFieldValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}))
				structFieldValue.Set(recursiveDescentValue)
			}
		}
	} else {
		n.lastError = NewError(FunctionName, fmt.Sprintf("unsupported value at recursive descent search segment %s", recursiveDescentSearchSegment)).WithData(currentValue.Interface()).WithPathSegments(currentPath).WithNestedError(ErrValueAtPathSegmentInvalidError)
	}
	return currentValue
}
