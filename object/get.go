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
Get retrieves value(s) in `Object.source` at jsonPath.

For JSONPath syntax patterns like the recursive descent pattern, wildcard, or union selector e.g., `$..One`, `$.One[*]`, `$.['One','Two','Three']`, expect a slice of type any which contains the values found.

Parameters:
  - jsonPath

Returns the number of results found and the last error encountered.
*/
func (n *Object) Get(jsonPath path.JSONPath) (uint64, error) {
	if string(jsonPath) == path.JsonpathKeyRoot || jsonPath == "" {
		n.valueFound = n.source
		return 1, nil
	}

	const FunctionName = "Get"

	n.recursiveDescentSegments = jsonPath.Parse()

	currentPathSegmentIndexes := internal.PathSegmentsIndexes{
		LastRecursive: len(n.recursiveDescentSegments) - 1,
	}
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive {
		return 0, NewError().WithFunctionName(FunctionName).WithMessage("recursiveDescentSegments empty").WithNestedError(ErrPathSegmentInvalidError).WithData(core.JsonObject{"Source": n.source.Interface()})
	}
	currentPathSegmentIndexes.LastCollection = len(n.recursiveDescentSegments[0]) - 1
	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return 0, NewError().WithFunctionName(FunctionName).WithMessage("recursiveDescentSegments empty").WithNestedError(ErrPathSegmentInvalidError).WithData(core.JsonObject{"Source": n.source.Interface()})
	}

	if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
		n.valueFound = n.recursiveGet(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]})
	} else {
		n.valueFound = n.recursiveDescentGet(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]})
	}

	if n.noOfResults > 0 {
		return n.noOfResults, nil
	}
	return n.noOfResults, n.lastError
}

func (n *Object) recursiveGet(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) reflect.Value {
	const FunctionName = "recursiveGet"

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("currentPathSegmentIndexes empty").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	recursiveSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveSegment == nil {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("recursiveSegment is nil").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if core.IsNilOrInvalid(currentValue) {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("value nil or invalid").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if currentValue.Kind() == reflect.Pointer || currentValue.Kind() == reflect.Interface {
		return n.recursiveGet(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if recursiveSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				n.noOfResults++
				return currentValue
			}

			recursiveDescentIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: 0,
				LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
			}
			return n.recursiveDescentGet(currentValue, recursiveDescentIndexes, currentPath)
		}

		recursiveIndexes := internal.PathSegmentsIndexes{
			CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
			LastRecursive:     currentPathSegmentIndexes.LastRecursive,
			CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
			LastCollection:    currentPathSegmentIndexes.LastCollection,
		}
		return n.recursiveGet(currentValue, recursiveIndexes, currentPath)
	}

	if mapKeyType, _, ok := core.GetMapKeyValueType(currentValue); ok {
		const dataKind = "map"

		if recursiveSegment.IsKey {
			var mapKey any
			if err := n.defaultConverter.Convert(recursiveSegment.Key, &schema.DynamicSchemaNode{Kind: mapKeyType.Kind(), Type: mapKeyType}, &mapKey); err != nil {
				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("convert mapKey %s to type %v failed", recursiveSegment, mapKeyType)).
					WithNestedError(err).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
				return reflect.Value{}
			}

			mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if mapValue.IsValid() {
						n.noOfResults++
						return mapValue
					}
					n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("value of map entry %s not valid", recursiveSegment)).
						WithNestedError(ErrValueAtPathSegmentInvalidError).
						WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
					return reflect.Value{}
				}

				recursiveDescentIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: 0,
					LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
				}
				return n.recursiveDescentGet(mapValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
			}

			recursiveIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
				LastCollection:    currentPathSegmentIndexes.LastCollection,
			}
			return n.recursiveGet(mapValue, recursiveIndexes, append(currentPath, recursiveSegment))
		}

		if recursiveSegment.IsKeyIndexAll || len(recursiveSegment.UnionSelector) > 0 {
			_sliceAny := make([]any, 0)
			selectorSlice := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)

			if recursiveSegment.IsKeyIndexAll {
				for _, valueKey := range currentValue.MapKeys() {
					mapValue := currentValue.MapIndex(valueKey)
					if mapValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, mapValue)
					}
				}
			} else {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsKey {
						continue
					}

					var mapKey any
					if err := n.defaultConverter.Convert(unionKey.Key, &schema.DynamicSchemaNode{Kind: mapKeyType.Kind(), Type: mapKeyType}, &mapKey); err != nil {
						n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("convert mapKey %s to type %v failed", recursiveSegment, mapKeyType)).
							WithNestedError(err).
							WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
						return reflect.Value{}
					}

					mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
					if mapValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, mapValue)
					}
				}
			}
			return n.selectorGetLoop(dataKind, selectorSlice, recursiveSegment, currentValue, currentPathSegmentIndexes, currentPath)
		}

		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, unsupported recursive segment %s", dataKind, recursiveSegment)).
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if _, ok := core.GetArraySliceValueType(currentValue); ok {
		const dataKind = "array/slice"

		if recursiveSegment.IsIndex {
			if recursiveSegment.Index >= currentValue.Len() {
				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, index %s out of range", dataKind, recursiveSegment)).
					WithNestedError(ErrValueAtPathSegmentInvalidError).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
				return reflect.Value{}
			}

			arraySliceValue := currentValue.Index(recursiveSegment.Index)
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if arraySliceValue.IsValid() {
						n.noOfResults++
						return arraySliceValue
					}
					n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("value in %s at index %s not valid", dataKind, recursiveSegment)).
						WithNestedError(ErrValueAtPathSegmentInvalidError).
						WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
					return reflect.Value{}
				}

				recursiveDescentIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: 0,
					LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
				}
				return n.recursiveDescentGet(arraySliceValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
			}

			recursiveIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
				LastCollection:    currentPathSegmentIndexes.LastCollection,
			}
			return n.recursiveGet(arraySliceValue, recursiveIndexes, append(currentPath, recursiveSegment))
		}

		if recursiveSegment.IsKeyIndexAll || len(recursiveSegment.UnionSelector) > 0 || recursiveSegment.LinearCollectionSelector != nil {
			_sliceAny := make([]any, 0)
			selectorSlice := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)

			if recursiveSegment.IsKeyIndexAll {
				for i := 0; i < currentValue.Len(); i++ {
					arraySliceValue := currentValue.Index(i)
					if arraySliceValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, arraySliceValue)
					}
				}
			} else if len(recursiveSegment.UnionSelector) > 0 {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsIndex || unionKey.Index >= currentValue.Len() {
						continue
					}

					valueFromSliceArray := currentValue.Index(unionKey.Index)
					if valueFromSliceArray.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, valueFromSliceArray)
					}
				}
			} else {
				start := 0
				if recursiveSegment.LinearCollectionSelector.IsStart {
					if recursiveSegment.LinearCollectionSelector.Start >= currentValue.Len() {
						n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, linear collection selector %s Start is out of range", dataKind, recursiveSegment)).
							WithNestedError(ErrPathSegmentInvalidError).
							WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
						return reflect.Value{}
					}
					start = recursiveSegment.LinearCollectionSelector.Start
				}
				step := 1
				if recursiveSegment.LinearCollectionSelector.IsStep {
					if recursiveSegment.LinearCollectionSelector.Step >= currentValue.Len() {
						n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, linear collection selector %s Step is out of range", dataKind, recursiveSegment)).
							WithNestedError(ErrPathSegmentInvalidError).
							WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
						return reflect.Value{}
					}
					if recursiveSegment.LinearCollectionSelector.Step > 0 {
						step = recursiveSegment.LinearCollectionSelector.Step
					}
				}
				end := currentValue.Len()
				if recursiveSegment.LinearCollectionSelector.IsEnd {
					if recursiveSegment.LinearCollectionSelector.End >= end {
						n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, linear collection selector %s End is out of range", dataKind, recursiveSegment)).
							WithNestedError(ErrPathSegmentInvalidError).
							WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
						return reflect.Value{}
					}
					end = recursiveSegment.LinearCollectionSelector.End
				}

				for i := start; i < end; i += step {
					if i >= currentValue.Len() {
						continue
					}
					valueFromSliceArray := currentValue.Index(i)
					if valueFromSliceArray.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, valueFromSliceArray)
					}
				}
			}

			return n.selectorGetLoop(dataKind, selectorSlice, recursiveSegment, currentValue, currentPathSegmentIndexes, currentPath)
		}

		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, unsupported recursive segment %s", dataKind, recursiveSegment)).
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if currentValue.Kind() == reflect.Struct {
		const dataKind = "struct"

		if recursiveSegment.IsKey {
			if !core.StartsWithCapital(recursiveSegment.Key) {
				n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("key %s is not valid for struct", recursiveSegment)).
					WithNestedError(ErrPathSegmentInvalidError).
					WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
				return reflect.Value{}
			}

			structFieldValue := currentValue.FieldByName(recursiveSegment.Key)
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if structFieldValue.IsValid() {
						n.noOfResults++
						return structFieldValue
					}
					n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("value of field %s in struct is not valid", recursiveSegment)).
						WithNestedError(ErrValueAtPathSegmentInvalidError).
						WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
					return reflect.Value{}
				}

				recursiveDescentIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: 0,
					LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
				}

				return n.recursiveDescentGet(structFieldValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
			}

			recursiveIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
				LastCollection:    currentPathSegmentIndexes.LastCollection,
			}

			return n.recursiveGet(structFieldValue, recursiveIndexes, append(currentPath, recursiveSegment))
		}

		if recursiveSegment.IsKeyIndexAll || len(recursiveSegment.UnionSelector) > 0 {
			_sliceAny := make([]any, 0)
			selectorSlice := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)

			if recursiveSegment.IsKeyIndexAll {
				for i := 0; i < currentValue.NumField(); i++ {
					if !core.IsStructFieldExported(currentValue.Type().Field(i)) {
						continue
					}

					structFieldValue := currentValue.Field(i)
					if structFieldValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, structFieldValue)
					}
				}
			} else {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsKey || !core.StartsWithCapital(unionKey.Key) {
						continue
					}

					structFieldValue := currentValue.FieldByName(unionKey.Key)
					if structFieldValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, structFieldValue)
					}
				}
			}

			return n.selectorGetLoop(dataKind, selectorSlice, recursiveSegment, currentValue, currentPathSegmentIndexes, currentPath)
		}

		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, unsupported recursive segment %s", dataKind, recursiveSegment)).
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("unsupported value at recursive segment").
		WithNestedError(ErrValueAtPathSegmentInvalidError).
		WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
	return reflect.Value{}
}

func (n *Object) selectorGetLoop(dataKind string, selectorSlice reflect.Value, recursiveSegment *path.CollectionMemberSegment, currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) reflect.Value {
	const FunctionName = "selectorGetLoop"
	_sliceAny := make([]any, 0)

	if selectorSlice.Len() == 0 {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, selector %s yielded no results", dataKind, recursiveSegment)).
			WithNestedError(ErrValueAtPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
		if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
			n.noOfResults = uint64(selectorSlice.Len())
			return selectorSlice
		}

		recursiveDescentIndexes := internal.PathSegmentsIndexes{
			CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
			LastRecursive:     currentPathSegmentIndexes.LastRecursive,
			CurrentCollection: 0,
			LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
		}

		return n.recursiveDescentGet(selectorSlice, recursiveDescentIndexes, append(currentPath, recursiveSegment))
	}

	recursiveIndexes := internal.PathSegmentsIndexes{
		CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
		LastRecursive:     currentPathSegmentIndexes.LastRecursive,
		CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
		LastCollection:    currentPathSegmentIndexes.LastCollection,
	}

	newSliceResult := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)
	for i := 0; i < selectorSlice.Len(); i++ {
		if v := n.recursiveGet(selectorSlice.Index(i), recursiveIndexes, append(currentPath, recursiveSegment)); v.IsValid() {
			newSliceResult = n.flattenNewSliceResult(newSliceResult, currentPathSegmentIndexes, v)
		}
	}

	if newSliceResult.Len() == 0 {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("in %s, recursively working with selector %s results yielded no ok results", dataKind, recursiveSegment)).
			WithNestedError(ErrValueAtPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	n.noOfResults = uint64(newSliceResult.Len())
	return newSliceResult
}

func (n *Object) recursiveDescentGet(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) reflect.Value {
	const FunctionName = "recursiveDescentGet"

	var valueFound reflect.Value
	{
		_sliceAny := make([]any, 0)
		valueFound = reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)
	}

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("currentPathSegmentIndexes exhausted").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	recursiveDescentSearchSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveDescentSearchSegment == nil {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("recursive descent search segment is not empty").
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if core.IsNilOrInvalid(currentValue) {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage("current value nil or invalid").
			WithNestedError(ErrValueAtPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if recursiveDescentSearchSegment.IsKeyRoot {
		return n.recursiveGet(currentValue, currentPathSegmentIndexes, currentPath)
	}

	if !recursiveDescentSearchSegment.IsKey {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("recursive descent search segment %s is not key", recursiveDescentSearchSegment)).
			WithNestedError(ErrPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if currentValue.Kind() == reflect.Pointer || currentValue.Kind() == reflect.Interface {
		return n.recursiveDescentGet(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if _, _, ok := core.GetMapKeyValueType(currentValue); ok {
		for _, mapKey := range currentValue.MapKeys() {
			mapValue := currentValue.MapIndex(mapKey)
			if !mapValue.IsValid() {
				continue
			}

			pathSegment := &path.CollectionMemberSegment{
				IsKey: true,
				Key:   mapKeyString(mapKey),
			}

			nextPathSegments := append(currentPath, pathSegment)

			if pathSegment.Key == recursiveDescentSearchSegment.Key {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						valueFound = reflect.Append(valueFound, mapValue)
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentGet(mapValue, recursiveDescentIndexes, nextPathSegments)
						if recursiveDescentValue.IsValid() {
							if recursiveDescentValue.Kind() == reflect.Slice {
								for i := 0; i < recursiveDescentValue.Len(); i++ {
									valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
								}
							} else {
								valueFound = reflect.Append(valueFound, recursiveDescentValue)
							}
						} else {
							n.noOfResults = uint64(valueFound.Len())
							return valueFound
						}
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveGet(mapValue, recursiveIndexes, nextPathSegments)
					if recursiveValue.IsValid() {
						if recursiveValue.Kind() == reflect.Slice || recursiveValue.Kind() == reflect.Array {
							for i := 0; i < recursiveValue.Len(); i++ {
								valueFound = reflect.Append(valueFound, recursiveValue.Index(i))
							}
						} else {
							valueFound = reflect.Append(valueFound, recursiveValue)
						}
					} else {
						n.noOfResults = uint64(valueFound.Len())
						return valueFound
					}
				}
			}

			recursiveDescentValue := n.recursiveDescentGet(mapValue, currentPathSegmentIndexes, nextPathSegments)
			if recursiveDescentValue.IsValid() {
				if recursiveDescentValue.Kind() == reflect.Slice {
					for i := 0; i < recursiveDescentValue.Len(); i++ {
						valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
					}
				} else {
					valueFound = reflect.Append(valueFound, recursiveDescentValue)
				}
			}
		}
	} else if _, ok := core.GetArraySliceValueType(currentValue); ok {
		for i := 0; i < currentValue.Len(); i++ {
			sliceArrayValue := currentValue.Index(i)
			if !sliceArrayValue.IsValid() {
				continue
			}

			recursiveDescentValue := n.recursiveDescentGet(sliceArrayValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
			if recursiveDescentValue.IsValid() {
				if recursiveDescentValue.Kind() == reflect.Slice {
					for i := 0; i < recursiveDescentValue.Len(); i++ {
						valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
					}
				} else {
					valueFound = reflect.Append(valueFound, recursiveDescentValue)
				}
			}
		}
	} else if currentValue.Kind() == reflect.Struct {
		if core.StartsWithCapital(recursiveDescentSearchSegment.Key) {
			if structFieldValue := currentValue.FieldByName(recursiveDescentSearchSegment.Key); structFieldValue.IsValid() {
				nextPathSegments := append(currentPath, recursiveDescentSearchSegment)
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						valueFound = reflect.Append(valueFound, structFieldValue)
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						recursiveDescentValue := n.recursiveDescentGet(structFieldValue, recursiveDescentIndexes, nextPathSegments)
						if recursiveDescentValue.IsValid() {
							if recursiveDescentValue.Kind() == reflect.Slice {
								for i := 0; i < recursiveDescentValue.Len(); i++ {
									valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
								}
							} else {
								valueFound = reflect.Append(valueFound, recursiveDescentValue)
							}
						} else {
							n.noOfResults = uint64(valueFound.Len())
							return valueFound
						}
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue := n.recursiveGet(structFieldValue, recursiveIndexes, nextPathSegments)
					if recursiveValue.IsValid() {
						if recursiveValue.Kind() == reflect.Slice || recursiveValue.Kind() == reflect.Array {
							for i := 0; i < recursiveValue.Len(); i++ {
								valueFound = reflect.Append(valueFound, recursiveValue.Index(i))
							}
						} else {
							valueFound = reflect.Append(valueFound, recursiveValue)
						}
					} else {
						n.noOfResults = uint64(valueFound.Len())
						return valueFound
					}
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

			recursiveDescentValue := n.recursiveDescentGet(structFieldValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}))
			if recursiveDescentValue.IsValid() {
				if recursiveDescentValue.Kind() == reflect.Slice {
					for i := 0; i < recursiveDescentValue.Len(); i++ {
						valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
					}
				} else {
					valueFound = reflect.Append(valueFound, recursiveDescentValue)
				}
			}
		}
	} else {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("unsupported value at recursive descent search segment %s", recursiveDescentSearchSegment)).
			WithNestedError(ErrValueAtPathSegmentInvalidError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	if valueFound.Len() == 0 {
		n.lastError = NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("no search value found at recursive descent search segment %s", recursiveDescentSearchSegment)).
			WithNestedError(ErrObjectError).
			WithData(core.JsonObject{"CurrentValue": currentValue.Interface(), "CurrentPathSegment": currentPath})
		return reflect.Value{}
	}

	n.noOfResults = uint64(valueFound.Len())
	return valueFound
}

// convert nested slice result v from recursiveGet into a single 1D slice if the next pathSegment contains CollectionMemberSegment.IsKeyIndexAll, CollectionMemberSegment.UnionSelector, or CollectionMemberSegment.LinearCollectionSelector.
func (n *Object) flattenNewSliceResult(newSliceResult reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, v reflect.Value) reflect.Value {
	if currentPathSegmentIndexes.CurrentCollection < currentPathSegmentIndexes.LastCollection {
		if v.Kind() == reflect.Slice {
			if n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection+1].IsKeyIndexAll || len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection+1].UnionSelector) > 0 || n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection+1].LinearCollectionSelector != nil {
				for i := 0; i < v.Len(); i++ {
					newSliceResult = reflect.Append(newSliceResult, v.Index(i))
				}
				return newSliceResult
			}
		}
	}

	newSliceResult = reflect.Append(newSliceResult, v)
	return newSliceResult
}
