package object

import (
	"reflect"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

// ForEach loops through `Object.source` using jsonPath as the guide.
//
// Useful for the following JSONPath syntax:
//   - Recursive descent e.g, `$...One`
//   - Wildcard e.g., `$.One[*]`
//   - Union selector e.g., `$.['One','Two','Three']`
//   - Array selector e.g., `$.[1:6:2]`
//
// Parameters:
//   - jsonPath
//   - ifValueFoundInObject - Called when each value is found
func (n *Object) ForEach(jsonPath path.JSONPath, ifValueFoundInObject IfValueFoundInObject) {
	n.recursiveDescentSegments = jsonPath.Parse()
	n.ifValueFoundInObject = ifValueFoundInObject

	currentPathSegmentIndexes := internal.PathSegmentsIndexes{
		CurrentRecursive: 0,
		LastRecursive:    len(n.recursiveDescentSegments) - 1,
	}
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive {
		return
	}
	currentPathSegmentIndexes.CurrentCollection = 0
	currentPathSegmentIndexes.LastCollection = len(n.recursiveDescentSegments[0]) - 1
	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return
	}

	if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
		n.recursiveForEachValue(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{})
	} else {
		n.recursiveDescentForEachValue(n.source, currentPathSegmentIndexes, path.RecursiveDescentSegment{})
	}
}

func (n *Object) recursiveForEachValue(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) bool {
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return false
	}

	if internal.IsNilOrInvalid(currentValue) {
		return false
	}

	recursiveSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]

	if recursiveSegment == nil {
		return false
	}

	if currentValue.Kind() == reflect.Pointer || currentValue.Kind() == reflect.Interface {
		// Unpack pointers/interfaces at the start.
		return n.recursiveForEachValue(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if recursiveSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				return n.ifValueFoundInObject(currentPath, currentValue.Interface())
			}

			recursiveDescentIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: 0,
				LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
			}

			return n.recursiveDescentForEachValue(currentValue, recursiveDescentIndexes, currentPath)
		}

		recursiveIndexes := internal.PathSegmentsIndexes{
			CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
			LastRecursive:     currentPathSegmentIndexes.LastRecursive,
			CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
			LastCollection:    currentPathSegmentIndexes.LastCollection,
		}
		return n.recursiveForEachValue(currentValue, recursiveIndexes, currentPath)
	}

	if mapKeyType, _, ok := core.GetMapKeyValueType(currentValue); ok {
		if recursiveSegment.IsKey {
			var mapKey any
			if err := n.defaultConverter.Convert(recursiveSegment.Key, &schema.DynamicSchemaNode{Kind: mapKeyType.Kind(), Type: mapKeyType}, &mapKey); err != nil {
				return false
			}

			mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if mapValue.IsValid() {
						return n.ifValueFoundInObject(append(currentPath, recursiveSegment), mapValue.Interface())
					}
					return false
				}

				recursiveDescentIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: 0,
					LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
				}

				return n.recursiveDescentForEachValue(mapValue, recursiveDescentIndexes, append(currentPath, recursiveSegment))
			}

			recursiveIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
				LastCollection:    currentPathSegmentIndexes.LastCollection,
			}
			return n.recursiveForEachValue(mapValue, recursiveIndexes, append(currentPath, recursiveSegment))
		}

		if recursiveSegment.IsKeyIndexAll || len(recursiveSegment.UnionSelector) > 0 {
			_sliceAny := make([]any, 0)
			selectorSlice := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)
			selectorSliceElementPaths := make(path.RecursiveDescentSegment, 0)

			if recursiveSegment.IsKeyIndexAll {
				for _, valueKey := range currentValue.MapKeys() {
					mapValue := currentValue.MapIndex(valueKey)
					if mapValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, mapValue)
						selectorSliceElementPaths = append(selectorSliceElementPaths, &path.CollectionMemberSegment{IsKey: true, Key: valueKey.String()})
					}
				}
			} else {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsKey {
						continue
					}

					var mapKey any
					if err := n.defaultConverter.Convert(unionKey.Key, &schema.DynamicSchemaNode{Kind: reflect.ValueOf(unionKey.Key).Kind(), Type: reflect.ValueOf(unionKey.Key).Type()}, &mapKey); err != nil {
						return false
					}

					mapValue := currentValue.MapIndex(reflect.ValueOf(mapKey))
					if mapValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, mapValue)
						selectorSliceElementPaths = append(selectorSliceElementPaths, unionKey)
					}
				}
			}

			return n.selectorForEachLoop(selectorSlice, selectorSliceElementPaths, currentPathSegmentIndexes, currentPath)
		}

		return false
	}

	if _, ok := core.GetArraySliceValueType(currentValue); ok {
		if recursiveSegment.IsIndex {
			if recursiveSegment.Index >= currentValue.Len() {
				return false
			}

			sliceArrayElementValue := currentValue.Index(recursiveSegment.Index)
			nextPathSegments := append(currentPath, recursiveSegment)
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if sliceArrayElementValue.IsValid() {
						if n.ifValueFoundInObject(nextPathSegments, sliceArrayElementValue.Interface()) {
							return true
						}
					}
					return false
				}

				recursiveDescentIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: 0,
					LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
				}

				return n.recursiveDescentForEachValue(sliceArrayElementValue, recursiveDescentIndexes, nextPathSegments)
			}

			recursiveIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
				LastCollection:    currentPathSegmentIndexes.LastCollection,
			}
			return n.recursiveForEachValue(sliceArrayElementValue, recursiveIndexes, nextPathSegments)
		}

		if recursiveSegment.IsKeyIndexAll || len(recursiveSegment.UnionSelector) > 0 || recursiveSegment.LinearCollectionSelector != nil {
			_sliceAny := make([]any, 0)
			selectorSlice := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)
			selectorSliceElementPaths := make(path.RecursiveDescentSegment, 0)

			if recursiveSegment.IsKeyIndexAll {
				for i := 0; i < currentValue.Len(); i++ {
					arraySliceValue := currentValue.Index(i)
					if arraySliceValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, arraySliceValue)
						selectorSliceElementPaths = append(selectorSliceElementPaths, &path.CollectionMemberSegment{IsIndex: true, Index: i})
					}
				}
			} else if len(recursiveSegment.UnionSelector) > 0 {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsIndex || unionKey.Index >= currentValue.Len() {
						continue
					}

					sliceArrayValue := currentValue.Index(unionKey.Index)
					if sliceArrayValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, sliceArrayValue)
						selectorSliceElementPaths = append(selectorSliceElementPaths, unionKey)
					}
				}
			} else {
				start := 0
				if recursiveSegment.LinearCollectionSelector.IsStart {
					if recursiveSegment.LinearCollectionSelector.Start >= currentValue.Len() {
						return false
					}
					start = recursiveSegment.LinearCollectionSelector.Start
				}
				step := 1
				if recursiveSegment.LinearCollectionSelector.IsStep {
					if recursiveSegment.LinearCollectionSelector.Step >= currentValue.Len() {
						return false
					}
					if recursiveSegment.LinearCollectionSelector.Step > 0 {
						step = recursiveSegment.LinearCollectionSelector.Step
					}
				}
				end := currentValue.Len()
				if recursiveSegment.LinearCollectionSelector.IsEnd {
					if recursiveSegment.LinearCollectionSelector.End >= end {
						return false
					}
					end = recursiveSegment.LinearCollectionSelector.End
				}

				for i := start; i < end; i += step {
					if i >= currentValue.Len() {
						continue
					}
					sliceArrayValue := currentValue.Index(i)
					if sliceArrayValue.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, sliceArrayValue)
						selectorSliceElementPaths = append(selectorSliceElementPaths, &path.CollectionMemberSegment{IsIndex: true, Index: i})
					}
				}
			}

			return n.selectorForEachLoop(selectorSlice, selectorSliceElementPaths, currentPathSegmentIndexes, currentPath)
		}

		return false
	}

	if currentValue.Kind() == reflect.Struct {
		if recursiveSegment.IsKey {
			if !internal.StartsWithCapital(recursiveSegment.Key) {
				return false
			}

			structFieldValue := currentValue.FieldByName(recursiveSegment.Key)
			nextPathSegments := append(currentPath, recursiveSegment)
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
					if structFieldValue.IsValid() {
						if n.ifValueFoundInObject(nextPathSegments, structFieldValue.Interface()) {
							return true
						}
					}
					return false
				}

				recursiveDescentIndexes := internal.PathSegmentsIndexes{
					CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
					LastRecursive:     currentPathSegmentIndexes.LastRecursive,
					CurrentCollection: 0,
					LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
				}

				return n.recursiveDescentForEachValue(structFieldValue, recursiveDescentIndexes, nextPathSegments)
			}

			recursiveIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
				LastCollection:    currentPathSegmentIndexes.LastCollection,
			}
			return n.recursiveForEachValue(structFieldValue, recursiveIndexes, nextPathSegments)
		}

		if recursiveSegment.IsKeyIndexAll || len(recursiveSegment.UnionSelector) > 0 {
			_sliceAny := make([]any, 0)
			selectorSlice := reflect.MakeSlice(reflect.TypeOf(_sliceAny), 0, 0)
			selectorSliceElementPaths := make(path.RecursiveDescentSegment, 0)

			if recursiveSegment.IsKeyIndexAll {
				for i := 0; i < currentValue.NumField(); i++ {
					if !internal.IsStructFieldExported(currentValue.Type().Field(i)) {
						continue
					}

					structField := currentValue.Field(i)
					if structField.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, structField)
						selectorSliceElementPaths = append(selectorSliceElementPaths, &path.CollectionMemberSegment{IsKey: true, Key: currentValue.Type().Field(i).Name})
					}
				}
			} else {
				for _, unionKey := range recursiveSegment.UnionSelector {
					if !unionKey.IsKey || !internal.StartsWithCapital(unionKey.Key) {
						continue
					}

					valueFromStruct := currentValue.FieldByName(unionKey.Key)
					if valueFromStruct.IsValid() {
						selectorSlice = reflect.Append(selectorSlice, valueFromStruct)
						selectorSliceElementPaths = append(selectorSliceElementPaths, unionKey)
					}
				}
			}

			return n.selectorForEachLoop(selectorSlice, selectorSliceElementPaths, currentPathSegmentIndexes, currentPath)
		}

		return false
	}

	return false
}

func (n *Object) selectorForEachLoop(selectorSlice reflect.Value, selectorSliceElementPaths path.RecursiveDescentSegment, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) bool {
	if selectorSlice.Len() == 0 {
		return false
	}

	for i := 0; i < len(selectorSliceElementPaths); i++ {
		nextPathSegments := append(currentPath, selectorSliceElementPaths[i])
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
				if n.ifValueFoundInObject(nextPathSegments, selectorSlice.Index(i).Interface()) {
					return true
				}
				continue
			}

			recursiveDescentIndexes := internal.PathSegmentsIndexes{
				CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
				LastRecursive:     currentPathSegmentIndexes.LastRecursive,
				CurrentCollection: 0,
				LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
			}

			if n.recursiveDescentForEachValue(selectorSlice.Index(i), recursiveDescentIndexes, nextPathSegments) {
				return true
			}
			continue
		}

		recursiveIndexes := internal.PathSegmentsIndexes{
			CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
			LastRecursive:     currentPathSegmentIndexes.LastRecursive,
			CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
			LastCollection:    currentPathSegmentIndexes.LastCollection,
		}

		if n.recursiveForEachValue(selectorSlice.Index(i), recursiveIndexes, nextPathSegments) {
			return true
		}
	}

	return false
}

func (n *Object) recursiveDescentForEachValue(currentValue reflect.Value, currentPathSegmentIndexes internal.PathSegmentsIndexes, currentPath path.RecursiveDescentSegment) bool {
	if currentPathSegmentIndexes.CurrentRecursive > currentPathSegmentIndexes.LastRecursive || currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return false
	}

	if internal.IsNilOrInvalid(currentValue) {
		return false
	}

	recursiveDescentSearchSegment := n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive][currentPathSegmentIndexes.CurrentCollection]

	if recursiveDescentSearchSegment == nil {
		return false
	}

	if recursiveDescentSearchSegment.IsKeyRoot {
		return n.recursiveForEachValue(currentValue, currentPathSegmentIndexes, append(currentPath, recursiveDescentSearchSegment))
	}

	if !recursiveDescentSearchSegment.IsKey {
		return false
	}

	if currentValue.Kind() == reflect.Pointer || currentValue.Kind() == reflect.Interface {
		return n.recursiveDescentForEachValue(currentValue.Elem(), currentPathSegmentIndexes, currentPath)
	}

	if _, _, ok := core.GetMapKeyValueType(currentValue); ok {
		for _, valueKey := range currentValue.MapKeys() {
			mapEntryValue := currentValue.MapIndex(valueKey)
			if !mapEntryValue.IsValid() {
				continue
			}

			nextPathSegments := append(currentPath, recursiveDescentSearchSegment)
			if valueKey.Interface() == recursiveDescentSearchSegment.Key {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						if n.ifValueFoundInObject(nextPathSegments, mapEntryValue.Interface()) {
							return true
						}
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						if n.recursiveDescentForEachValue(mapEntryValue, recursiveDescentIndexes, nextPathSegments) {
							return true
						}
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					if n.recursiveForEachValue(mapEntryValue, recursiveIndexes, nextPathSegments) {
						return true
					}
				}
			}

			if n.recursiveDescentForEachValue(mapEntryValue, currentPathSegmentIndexes, nextPathSegments) {
				return true
			}
		}
	} else if _, ok := core.GetArraySliceValueType(currentValue); ok {
		for i := 0; i < currentValue.Len(); i++ {
			sliceArrayElementValue := currentValue.Index(i)
			if !sliceArrayElementValue.IsValid() {
				continue
			}

			nextPathSegments := append(currentPath, &path.CollectionMemberSegment{IsIndex: true, Index: i})

			if n.recursiveDescentForEachValue(sliceArrayElementValue, currentPathSegmentIndexes, nextPathSegments) {
				return true
			}
		}
	} else if currentValue.Kind() == reflect.Struct {
		if internal.StartsWithCapital(recursiveDescentSearchSegment.Key) {
			if structFieldValue := currentValue.FieldByName(recursiveDescentSearchSegment.Key); structFieldValue.IsValid() {
				nextPathSegments := append(currentPath, recursiveDescentSearchSegment)

				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					if currentPathSegmentIndexes.CurrentRecursive == currentPathSegmentIndexes.LastRecursive {
						if n.ifValueFoundInObject(nextPathSegments, structFieldValue.Interface()) {
							return true
						}
					} else {
						recursiveDescentIndexes := internal.PathSegmentsIndexes{
							CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive + 1,
							LastRecursive:     currentPathSegmentIndexes.LastRecursive,
							CurrentCollection: 0,
							LastCollection:    len(n.recursiveDescentSegments[currentPathSegmentIndexes.CurrentRecursive+1]) - 1,
						}

						if n.recursiveDescentForEachValue(structFieldValue, recursiveDescentIndexes, nextPathSegments) {
							return true
						}
					}
				} else {
					recursiveIndexes := internal.PathSegmentsIndexes{
						CurrentRecursive:  currentPathSegmentIndexes.CurrentRecursive,
						LastRecursive:     currentPathSegmentIndexes.LastRecursive,
						CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1,
						LastCollection:    currentPathSegmentIndexes.LastCollection,
					}

					if n.recursiveForEachValue(structFieldValue, recursiveIndexes, nextPathSegments) {
						return true
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

			if n.recursiveDescentForEachValue(structFieldValue, currentPathSegmentIndexes, currentPath) {
				return true
			}
		}
	}

	return false
}
