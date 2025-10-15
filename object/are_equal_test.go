package object

import (
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
)

func TestObject_AreEqual(t *testing.T) {
	for testData := range AreEqualTestData {
		if NewAreEqual().WithCustomEquals(testData.CustomAreEquals).AreEqual(reflect.ValueOf(testData.Left), reflect.ValueOf(testData.Right)) != testData.Expected {
			t.Error(
				"AreEqual(testData.Left, testData.Right) not equal to testData.Expected\n",
				"testData.Expected", testData.Expected, "\n",
				"testData.Left", internal.JsonStringifyMust(testData.Left), "\n",
				"testData.Right", internal.JsonStringifyMust(testData.Right),
			)
		}
	}
}

type AreEqualData struct {
	Left, Right     any
	Expected        bool
	CustomAreEquals AreEquals
}

func AreEqualTestData(yield func(data *AreEqualData) bool) {
	if !yield(
		&AreEqualData{
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

	if !yield(
		&AreEqualData{
			Left:     2.3456567,
			Right:    2.2345678,
			Expected: false,
		},
	) {
		return
	}

	if !yield(
		&AreEqualData{
			Left:     2,
			Right:    2,
			Expected: true,
		},
	) {
		return
	}

	if !yield(
		&AreEqualData{
			Left:     "23",
			Right:    "23",
			Expected: true,
		},
	) {
		return
	}

	if !yield(
		&AreEqualData{
			Left:     nil,
			Right:    nil,
			Expected: true,
		},
	) {
		return
	}
}
