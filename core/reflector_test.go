package core

import (
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
)

func TestIsMap(t *testing.T) {
	for isMapData := range IsMapTestData {
		mapKeyType, mapValueType, isMap := IsMap(isMapData.Data)

		if isMap != isMapData.ExpectedToBeMap {
			t.Error(
				"expected ExpectedToBeMap=", isMapData.ExpectedToBeMap, "\n",
				"got=", isMap, "\n",
				"data", internal.JsonStringifyMust(isMapData.Data),
			)
		}

		if isMap {
			if isMapData.MapKeyType.String() != mapKeyType.Kind().String() {
				t.Error(
					"expected MapKeyType=", isMapData.MapKeyType.String(), "\n",
					"got=", mapKeyType.String(), "\n",
					"data", internal.JsonStringifyMust(isMapData.Data),
				)
			}

			if isMapData.MapValueType.String() != mapValueType.Kind().String() {
				t.Error(
					"expected MapValueType=", isMapData.MapValueType.String(), "\n",
					"got=", mapValueType.String(), "\n",
					"data", internal.JsonStringifyMust(isMapData.Data),
				)
			}
		}
	}
}

func TestIsArray(t *testing.T) {
	for isArrayData := range IsArrayTestData {
		listItemType, isArray := IsArray(isArrayData.Data)

		if isArray != isArrayData.ExpectedToBeArray {
			t.Error(
				"expected ExpectedToBeArray=", isArrayData.ExpectedToBeArray, "\n",
				"got=", isArray, "\n",
				"data", internal.JsonStringifyMust(isArrayData.Data),
			)
		}

		if isArray {
			if isArrayData.ArrayItemType.String() != listItemType.Kind().String() {
				t.Error(
					"expected ArrayItemType=", isArrayData.ArrayItemType.String(), "\n",
					"got=", listItemType.String(), "\n",
					"data", internal.JsonStringifyMust(isArrayData.Data),
				)
			}
		}
	}
}

func TestIsSlice(t *testing.T) {
	for isSliceData := range IsSliceTestData {
		listItemType, isSlice := IsSlice(isSliceData.Data)

		if isSlice != isSliceData.ExpectedToBeSlice {
			t.Error(
				"expected ExpectedToBeSlice=", isSliceData.ExpectedToBeSlice, "\n",
				"got=", isSliceData, "\n",
				"data", internal.JsonStringifyMust(isSliceData.Data),
			)
		}

		if isSlice {
			if isSliceData.SliceItemType.String() != listItemType.Kind().String() {
				t.Error(
					"expected ArrayItemType=", isSliceData.SliceItemType.String(), "\n",
					"got=", listItemType.String(), "\n",
					"data", internal.JsonStringifyMust(isSliceData.Data),
				)
			}
		}
	}
}

type IsMapData struct {
	Data            any
	ExpectedToBeMap bool
	MapKeyType      reflect.Kind
	MapValueType    reflect.Kind
}

func IsMapTestData(yield func(data *IsMapData) bool) {
	if !yield(
		&IsMapData{
			&IsMapData{},
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsMapData{
			IsMapData{},
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsMapData{
			0,
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsMapData{
			map[string]int{},
			true,
			reflect.String,
			reflect.Int,
		},
	) {
		return
	}

	if !yield(
		&IsMapData{
			[]int{},
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsMapData{
			map[int]int{},
			true,
			reflect.Int,
			reflect.Int,
		},
	) {
		return
	}
}

type IsArrayData struct {
	Data              any
	ExpectedToBeArray bool
	ArrayItemType     reflect.Kind
}

func IsArrayTestData(yield func(*IsArrayData) bool) {
	if !yield(
		&IsArrayData{
			&IsArrayData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsArrayData{
			IsArrayData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsArrayData{
			0,
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsArrayData{
			[]any{1, "23", IsArrayData{}},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsArrayData{
			[3]any{1, "23", IsArrayData{}},
			true,
			reflect.Interface,
		},
	) {
		return
	}

	if !yield(
		&IsArrayData{
			[2]int{1, 2},
			true,
			reflect.Int,
		},
	) {
		return
	}

	if !yield(
		&IsArrayData{
			map[int]int{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}
}

type IsSliceData struct {
	Data              any
	ExpectedToBeSlice bool
	SliceItemType     reflect.Kind
}

func IsSliceTestData(yield func(*IsSliceData) bool) {
	if !yield(
		&IsSliceData{
			&IsSliceData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsSliceData{
			IsSliceData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsSliceData{
			0,
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	if !yield(
		&IsSliceData{
			[]any{},
			true,
			reflect.Interface,
		},
	) {
		return
	}

	if !yield(
		&IsSliceData{
			[]int{},
			true,
			reflect.Int,
		},
	) {
		return
	}

	if !yield(
		&IsSliceData{
			map[int]int{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}
}
