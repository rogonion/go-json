package object

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
)

func TestObject_AreEqual(t *testing.T) {
	for testData := range AreEqualTestData {
		if NewAreEqual().WithCustomEquals(testData.CustomAreEquals).AreEqual(reflect.ValueOf(testData.Left), reflect.ValueOf(testData.Right)) != testData.Expected {
			t.Error(
				testData.TestTitle, "\n",
				"Result of AreEqual not equal to testData.Expected\n",
				"testData.Expected", testData.Expected, "\n",
				"testData.Left", core.JsonStringifyMust(testData.Left), "\n",
				"testData.Right", core.JsonStringifyMust(testData.Right),
			)
		}
	}
}

type AreEqualData struct {
	internal.TestData
	Left, Right     any
	Expected        bool
	CustomAreEquals AreEquals
}

func AreEqualTestData(yield func(data *AreEqualData) bool) {
	testCaseIndex := 1
	if !yield(
		&AreEqualData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Check if two maps are equal", testCaseIndex),
			},
			Left: map[string]any{
				"a": "b",
				"c": "d",
				"e": false,
				"f": nil,
				"g": []any{
					23.56,
					&struct {
						One   bool
						Two   string
						Three int
						Four  map[int]string
						Five  *[2]int
					}{
						One:   true,
						Two:   "two",
						Three: 3,
						Four: map[int]string{
							1: "one",
							2: "two",
							3: "three",
						},
						Five: &[2]int{1, 2},
					},
				},
			},
			Right: map[string]any{
				"a": "b",
				"c": "d",
				"e": false,
				"f": nil,
				"g": []any{
					23.56,
					&struct {
						One   bool
						Two   string
						Three int
						Four  map[int]string
						Five  *[2]int
					}{
						One:   true,
						Two:   "two",
						Three: 3,
						Four: map[int]string{
							1: "one",
							2: "two",
							3: "three",
						},
						Five: &[2]int{1, 2},
					},
				},
			},
			Expected: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&AreEqualData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Check if two decimal numbers are equal", testCaseIndex),
			},
			Left:     2.3456567,
			Right:    2.2345678,
			Expected: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&AreEqualData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Check if two int numbers are equal", testCaseIndex),
			},
			Left:     2,
			Right:    2,
			Expected: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&AreEqualData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Check if two strings are equal", testCaseIndex),
			},
			Left:     "23",
			Right:    "23",
			Expected: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&AreEqualData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Check if two nil are equal", testCaseIndex),
			},
			Left:     nil,
			Right:    nil,
			Expected: true,
		},
	) {
		return
	}
}
