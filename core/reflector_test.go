package core

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
)

func TestCore_Reflector_IsMap(t *testing.T) {
	for testData := range IsMapTestData {
		mapKeyType, mapValueType, isMap := IsMap(testData.Data)

		if isMap != testData.ExpectedToBeMap {
			t.Error(
				testData.TestTitle, "\n",
				"expected ExpectedToBeMap=", testData.ExpectedToBeMap, "\n",
				"got=", isMap, "\n",
			)
		}

		if isMap {
			if testData.MapKeyType.String() != mapKeyType.Kind().String() {
				t.Error(
					testData.TestTitle, "\n",
					"expected MapKeyType=", testData.MapKeyType.String(), "\n",
					"got=", mapKeyType.String(),
				)
			}

			if testData.MapValueType.String() != mapValueType.Kind().String() {
				t.Error(
					testData.TestTitle, "\n",
					"expected MapValueType=", testData.MapValueType.String(), "\n",
					"got=", mapValueType.String(),
				)
			}
		}
	}
}

func TestCore_Reflector_IsArray(t *testing.T) {
	for testData := range IsArrayTestData {
		listItemType, isArray := IsArray(testData.Data)

		if isArray != testData.ExpectedToBeArray {
			t.Error(
				testData.TestTitle, "\n",
				"expected ExpectedToBeArray=", testData.ExpectedToBeArray, "\n",
				"got=", isArray,
			)
		}

		if isArray {
			if testData.ArrayItemType.String() != listItemType.Kind().String() {
				t.Error(
					testData.TestTitle, "\n",
					"expected ArrayItemType=", testData.ArrayItemType.String(), "\n",
					"got=", listItemType.String(),
				)
			}
		}
	}
}

func TestCore_Reflector_IsSlice(t *testing.T) {
	for testData := range IsSliceTestData {
		listItemType, isSlice := IsSlice(testData.Data)

		if isSlice != testData.ExpectedToBeSlice {
			t.Error(
				testData.TestTitle, "\n",
				"expected ExpectedToBeSlice=", testData.ExpectedToBeSlice, "\n",
				"got=", testData,
			)
		}

		if isSlice {
			if testData.SliceItemType.String() != listItemType.Kind().String() {
				t.Error(
					testData.TestTitle, "\n",
					"expected ArrayItemType=", testData.SliceItemType.String(), "\n",
					"got=", listItemType.String(),
				)
			}
		}
	}
}

type IsMapData struct {
	internal.TestData
	Data            any
	ExpectedToBeMap bool
	MapKeyType      reflect.Kind
	MapValueType    reflect.Kind
}

func IsMapTestData(yield func(data *IsMapData) bool) {
	testCaseIndex := 1
	if !yield(
		&IsMapData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing IsMapTestData{}", testCaseIndex),
			},
			IsMapData{},
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsMapData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing &IsMapTestData{}", testCaseIndex),
			},
			&IsMapData{},
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsMapData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing int", testCaseIndex),
			},
			0,
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsMapData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing map[string]int{}", testCaseIndex),
			},
			map[string]int{},
			true,
			reflect.String,
			reflect.Int,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsMapData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing []int{}", testCaseIndex),
			},
			[]int{},
			false,
			reflect.Invalid,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsMapData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing map[int]int{}", testCaseIndex),
			},
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
	internal.TestData
	Data              any
	ExpectedToBeArray bool
	ArrayItemType     reflect.Kind
}

func IsArrayTestData(yield func(*IsArrayData) bool) {
	testCaseIndex := 1
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing &IsArrayData{}", testCaseIndex),
			},
			&IsArrayData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing IsArrayData{}", testCaseIndex),
			},
			IsArrayData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing int", testCaseIndex),
			},
			0,
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing []any", testCaseIndex),
			},
			[]any{1, "23", IsArrayData{}},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing [3]any", testCaseIndex),
			},
			[3]any{1, "23", IsArrayData{}},
			true,
			reflect.Interface,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing [2]int", testCaseIndex),
			},
			[2]int{1, 2},
			true,
			reflect.Int,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsArrayData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing map[int]int", testCaseIndex),
			},
			map[int]int{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}
}

type IsSliceData struct {
	internal.TestData
	Data              any
	ExpectedToBeSlice bool
	SliceItemType     reflect.Kind
}

func IsSliceTestData(yield func(*IsSliceData) bool) {
	testCaseIndex := 1
	if !yield(
		&IsSliceData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing &IsSliceData{}", testCaseIndex),
			},
			&IsSliceData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsSliceData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing IsSliceData{}", testCaseIndex),
			},
			IsSliceData{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsSliceData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing int", testCaseIndex),
			},
			0,
			false,
			reflect.Invalid,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsSliceData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing []any", testCaseIndex),
			},
			[]any{},
			true,
			reflect.Interface,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsSliceData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing []int", testCaseIndex),
			},
			[]int{},
			true,
			reflect.Int,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&IsSliceData{
			internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Testing map[int]int", testCaseIndex),
			},
			map[int]int{},
			false,
			reflect.Invalid,
		},
	) {
		return
	}
}
