package object

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

func (n *Object) Get(jsonPath path.JSONPath) (any, bool, error) {
	if string(jsonPath) == path.JsonpathKeyRoot || jsonPath == "" {
		return n.source.Interface(), true, nil
	}

	const FunctionName = "Get"

	n.recursiveDescentSegments = jsonPath.Parse()

	currentPathSegmentIndexes := internal.PathSegmentsIndexes{
		LastRecursive: len(n.recursiveDescentSegments) - 1,
	}
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive {
		return nil, false, NewError(ErrPathSegmentInvalidError, FunctionName, "recursiveDescentSegments empty", n.source.Interface(), nil)
	}
	currentPathSegmentIndexes.LastCollection = len(n.recursiveDescentSegments[0]) - 1
	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return nil, false, NewError(ErrPathSegmentInvalidError, FunctionName, "recursiveDescentSegments empty", n.source.Interface(), nil)
	}

	if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
		if value, ok, err := n.recursiveGet(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]}); ok {
			return value.Interface(), ok, err
		} else {
			return nil, false, err
		}
	}

	if value, ok, err := n.recursiveDescentGet(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]}); ok {
		return value.Interface(), ok, err
	} else {
		return nil, false, err
	}
}

func (n *Object) recursiveGet(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) (reflect.Value, bool, error) {
	const FunctionName = "recursiveGet"

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, "currentPathSegmentIndexes empty", currentValue.Interface(), currentPath)
	}

	recursiveSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveSegment == nil {
		return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, "recursiveSegment is nil", currentValue.Interface(), currentPath)
	}

	if internal.IsNilOrInvalid(currentValue) {
		return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, "value nil or invalid", currentValue.Interface(), currentPath)
	}

	if currentValue.Kind() == reflect.Pointer || currentValue.Kind() == reflect.Interface {
		return n.recursiveGet(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if recursiveSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				return currentValue, true, nil
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
				return reflect.Value{}, false, NewError(err, FunctionName, fmt.Sprintf("convert mapKey %s to type %v failed", recursiveSegment, mapKeyType), currentValue.Interface(), currentPath)
			}

			mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if mapValue.IsValid() {
						return mapValue, true, nil
					}
					return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("value of map entry %s not valid", recursiveSegment), currentValue.Interface(), currentPath)
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
						return reflect.Value{}, false, NewError(err, FunctionName, fmt.Sprintf("convert mapKey %s to type %v failed", recursiveSegment, mapKeyType), currentValue.Interface(), currentPath)
					}

					mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
					if mapValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, mapValue)
					}
				}
			}

			return n.selectorGetLoop(dataKind, selectorSlice, recursiveSegment, currentValue, currentPathSegmentIndexes, currentPath)
		}

		return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, unsupported recursive segment %s", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
	}

	if _, ok := core.GetArraySliceValueType(currentValue); ok {
		const dataKind = "array/slice"

		if recursiveSegment.IsIndex {
			if recursiveSegment.Index >= currentValue.Len() {
				return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, index %s out of range", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
			}

			arraySliceValue := currentValue.Index(recursiveSegment.Index)
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if arraySliceValue.IsValid() {
						return arraySliceValue, true, nil
					}
					return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("value in %s at index %s not valid", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
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
						return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, linear collection selector %s Start is out of range", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
					}
					start = recursiveSegment.LinearCollectionSelector.Start
				}
				step := 1
				if recursiveSegment.LinearCollectionSelector.IsStep {
					if recursiveSegment.LinearCollectionSelector.Step >= currentValue.Len() {
						return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, linear collection selector %s Step is out of range", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
					}
					if recursiveSegment.LinearCollectionSelector.Step > 0 {
						step = recursiveSegment.LinearCollectionSelector.Step
					}
				}
				end := currentValue.Len()
				if recursiveSegment.LinearCollectionSelector.IsEnd {
					if recursiveSegment.LinearCollectionSelector.End >= end {
						return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, linear collection selector %s End is out of range", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
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

		return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, unsupported recursive segment %s", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
	}

	if currentValue.Kind() == reflect.Struct {
		const dataKind = "struct"

		if recursiveSegment.IsKey {
			if !internal.StartsWithCapital(recursiveSegment.Key) {
				return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("key %s is not valid for struct", recursiveSegment), currentValue.Interface(), currentPath)
			}

			structFieldValue := currentValue.FieldByName(recursiveSegment.Key)
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if structFieldValue.IsValid() {
						return structFieldValue, true, nil
					}
					return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("value of field %s in struc is not valid", recursiveSegment), currentValue.Interface(), currentPath)
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
					if !internal.IsStructFieldExported(currentValue.Type().Field(i)) {
						continue
					}

					structFieldValue := currentValue.Field(i)
					if structFieldValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, structFieldValue)
					}
				}
			} else {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsKey || !internal.StartsWithCapital(unionKey.Key) {
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

		return reflect.Value{}, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, unsupported recursive segment %s", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
	}

	return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, "unsupported value at recursive segment", currentValue.Interface(), currentPath)
}

func (n *Object) selectorGetLoop(dataKind string, selectorSlice reflect.Value, recursiveSegment *path.CollectionMemberSegment, currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) (reflect.Value, bool, error) {
	const FunctionName = "selectorGetLoop"
	_sliceAny := make([]any, 0)

	if selectorSlice.Len() == 0 {
		return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, selector %s yielded no results", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
	}

	if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
		if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
			return selectorSlice, true, nil
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
		v, ok, _ := n.recursiveGet(selectorSlice.Index(i), recursiveIndexes, append(currentPath, recursiveSegment))
		if ok {
			newSliceResult = n.flattenNewSliceResult(newSliceResult, currentPathSegmentIndexes, v)
		}
	}

	if newSliceResult.Len() == 0 {
		return reflect.Value{}, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("in %s, recursively working with selector %s results yielded no ok results", dataKind, recursiveSegment), currentValue.Interface(), currentPath)
	}

	return newSliceResult, true, nil
}

func (n *Object) recursiveDescentGet(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) (reflect.Value, bool, error) {
	const FunctionName = "recursiveDescentGet"

	var valueFound reflect.Value
	{
		_sliceAny := make([]any, 0)
		valueFound = reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)
	}

	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return valueFound, false, NewError(ErrPathSegmentInvalidError, FunctionName, "currentPathSegmentIndexes exhausted", currentValue.Interface(), currentPath)
	}

	recursiveDescentSearchSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]
	if recursiveDescentSearchSegment == nil {
		return valueFound, false, NewError(ErrPathSegmentInvalidError, FunctionName, "recursive descent search segment is not empty", currentValue.Interface(), currentPath)
	}

	if internal.IsNilOrInvalid(currentValue) {
		return valueFound, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, "current value nil or invalid", currentValue.Interface(), currentPath)
	}

	if recursiveDescentSearchSegment.IsKeyRoot {
		return n.recursiveGet(currentValue, currentPathSegmentIndexes, currentPath)
	}

	if !recursiveDescentSearchSegment.IsKey {
		return valueFound, false, NewError(ErrPathSegmentInvalidError, FunctionName, fmt.Sprintf("recursive descent search segment %s is not key", recursiveDescentSearchSegment), currentValue.Interface(), currentPath)
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

						recursiveDescentValue, ok, err := n.recursiveDescentGet(mapValue, recursiveDescentIndexes, nextPathSegments)
						if ok {
							if recursiveDescentValue.Kind() == reflect.Slice {
								for i := 0; i < recursiveDescentValue.Len(); i++ {
									valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
								}
							} else {
								valueFound = reflect.Append(valueFound, recursiveDescentValue)
							}
						} else {
							return valueFound, false, err
						}
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue, ok, err := n.recursiveGet(mapValue, recursiveIndexes, nextPathSegments)
					if ok {
						if recursiveValue.Kind() == reflect.Slice || recursiveValue.Kind() == reflect.Array {
							for i := 0; i < recursiveValue.Len(); i++ {
								valueFound = reflect.Append(valueFound, recursiveValue.Index(i))
							}
						} else {
							valueFound = reflect.Append(valueFound, recursiveValue)
						}
					} else {
						return valueFound, false, err
					}
				}
			}

			recursiveDescentValue, ok, _ := n.recursiveDescentGet(mapValue, currentPathSegmentIndexes, nextPathSegments)
			if ok {
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

			recursiveDescentValue, ok, _ := n.recursiveDescentGet(sliceArrayValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i}))
			if ok {
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
		if internal.StartsWithCapital(recursiveDescentSearchSegment.Key) {
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

						recursiveDescentValue, ok, err := n.recursiveDescentGet(structFieldValue, recursiveDescentIndexes, nextPathSegments)
						if ok {
							if recursiveDescentValue.Kind() == reflect.Slice {
								for i := 0; i < recursiveDescentValue.Len(); i++ {
									valueFound = reflect.Append(valueFound, recursiveDescentValue.Index(i))
								}
							} else {
								valueFound = reflect.Append(valueFound, recursiveDescentValue)
							}
						} else {
							return valueFound, false, err
						}
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					recursiveValue, ok, err := n.recursiveGet(structFieldValue, recursiveIndexes, nextPathSegments)
					if ok {
						if recursiveValue.Kind() == reflect.Slice || recursiveValue.Kind() == reflect.Array {
							for i := 0; i < recursiveValue.Len(); i++ {
								valueFound = reflect.Append(valueFound, recursiveValue.Index(i))
							}
						} else {
							valueFound = reflect.Append(valueFound, recursiveValue)
						}
					} else {
						return valueFound, false, err
					}
				}
			}
		}

		for i := 0; i < currentValue.NumField(); i++ {
			if !internal.IsStructFieldExported(currentValue.Type().Field(i)) {
				continue
			}

			structFieldValue := currentValue.Field(i)
			if !structFieldValue.IsValid() {
				continue
			}

			recursiveDescentValue, ok, err := n.recursiveDescentGet(structFieldValue, currentPathSegmentIndexes, append(currentPath, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name}))
			if ok && err == nil {
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
		return valueFound, false, NewError(ErrValueAtPathSegmentInvalidError, FunctionName, fmt.Sprintf("unsupported value at recursive descent search segment %s", recursiveDescentSearchSegment), currentValue.Interface(), currentPath)
	}

	if valueFound.Len() == 0 {
		return reflect.Value{}, false, NewError(ErrObjectProcessorError, FunctionName, fmt.Sprintf("no search value found at recursive descent search segment %s", recursiveDescentSearchSegment), currentValue.Interface(), currentPath)
	}

	return valueFound, true, nil
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
