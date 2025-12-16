package object

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

func TestObject_Get(t *testing.T) {
	for testData := range GetTestData {
		obj := NewObject().WithSourceInterface(testData.Root)
		noOfResults, err := obj.Get(testData.Path)
		if noOfResults != testData.ExpectedOk {
			t.Error(
				testData.TestTitle, "\n",
				"expected ok=", testData.ExpectedOk, "got ok=", noOfResults, "\n",
				"path=", testData.Path,
			)
		}

		if err != nil && testData.LogErrorsIfExpectedNotOk {
			var objectProcessorError *core.Error
			if errors.As(err, &objectProcessorError) {
				t.Error(
					testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					objectProcessorError.String(), "\n",
					"-----------------------",
				)
			}
		}

		valueFound := obj.GetValueFoundInterface()
		if !reflect.DeepEqual(valueFound, testData.ExpectedValue) {
			t.Error(
				testData.TestTitle, "\n",
				"res not equal to testData.ExpectedValue\n",
				"Path", testData.Path, "\n",
				"res=", core.JsonStringifyMust(valueFound), "\n",
				"JSON testData.Expected=", core.JsonStringifyMust(testData.ExpectedValue),
			)
		}
	}
}

type GetData struct {
	internal.TestData
	Root          any
	Path          path.JSONPath
	ExpectedOk    uint64
	ExpectedValue any
}

func GetTestData(yield func(data *GetData) bool) {
	testCaseIndex := 1
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Complex nested path with maps and structs\nThis test navigates through nested maps and structs to a specific field.", testCaseIndex),
			},
			Root: map[string]any{
				"data": map[string]any{
					"metadata": struct {
						Address Address
						Status  string
					}{
						Address: Address{
							Street: "123 Main St",
							City:   "Anytown",
						},
						Status: "active",
					},
				},
			},
			Path:          "$.data.metadata.Address.City",
			ExpectedOk:    1,
			ExpectedValue: "Anytown",
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Recursive descent with a wildcard to get all names\nThis tests if the function can find all \"Name\" fields regardless of nesting.", testCaseIndex),
			},
			Root: []any{
				map[string]any{"User": User{Name: "Alice"}},
				ComplexData{
					User: User{Name: "Bob"},
					Items: []struct {
						Name  string
						Value int
					}{
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
			Path:          "$..Name",
			ExpectedOk:    4,
			ExpectedValue: []any{"Alice", "Item 1", "Item 2", "Bob"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Union operator to get multiple fields from a single struct\nThis tests the `['fieldA', 'fieldB']` operator.", testCaseIndex),
			},
			Root: ComplexData{
				ID: 99,
				User: User{
					ID:    10,
					Name:  "Charlie",
					Email: "charlie@example.com",
				},
			},
			Path:          "$.User['ID', 'Name']",
			ExpectedOk:    2,
			ExpectedValue: []any{10, "Charlie"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Deep path with a slice index and field extraction\nThis tests a combination of operators to get a specific value from a deeply nested slice.", testCaseIndex),
			},
			Root: ComplexData{
				Items: []struct {
					Name  string
					Value int
				}{
					{Name: "First", Value: 10},
					{Name: "Second", Value: 20},
					{Name: "Third", Value: 30},
				},
			},
			Path:          "$.Items[1].Value",
			ExpectedOk:    1,
			ExpectedValue: 20,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Path that should not find a match\nThis test ensures the function correctly returns false and a nil value for an invalid path.", testCaseIndex),
			},
			Root: map[string]any{
				"store": map[string]any{"book": "fiction"},
			},
			Path:          "$.store.magazine",
			ExpectedOk:    0,
			ExpectedValue: nil,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Root: []any{
				map[string]any{
					"one": struct {
						Three []any
					}{
						Three: []any{
							0, 1, 2, 3, map[string]any{
								"04": []int{0, 4},
							}, 5, 6, 7, 8, 9,
						},
					},
				},
				&struct {
					Two map[string]any
				}{
					Two: map[string]any{
						"Three": []any{
							0, 1, 2, 3, []any{4}, 5, 6, 7, 8, 9,
						},
					},
				},
				[]any{
					map[string]any{
						"Three": []any{
							0, 1, 2, 3, struct {
								TwentyFour int
							}{
								TwentyFour: 24,
							}, 5, 6, 7, 8, 9,
						},
					},
				},
			},
			Path:          "$..Three[::2]['TwentyFour','04']",
			ExpectedOk:    2,
			ExpectedValue: []any{[]int{0, 4}, 24},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Root: []any{
				map[string]any{
					"one": struct {
						Three []any
					}{
						Three: []any{
							0, 1, 2, 3, []any{4}, 5, 6, 7, 8, 9,
						},
					},
				},
				&struct {
					Two map[string]any
				}{
					Two: map[string]any{
						"Three": []any{
							0, 1, 2, 3, map[string]any{
								"04": "04",
							}, 5, 6, 7, 8, 9,
						},
					},
				},
				[]any{
					map[string]any{
						"Three": []any{
							0, 1, 2, 3, struct {
								TwentyFour int
							}{
								TwentyFour: 24,
							}, 5, 6, 7, 8, 9,
						},
					},
				},
			},
			Path:          "$..Three[::2]['TwentyFour','04']",
			ExpectedOk:    2,
			ExpectedValue: []any{"04", 24},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Root: []any{
				map[string]any{
					"one": struct {
						Three []any
					}{
						Three: []any{
							0, 1, 2, 3, []any{
								map[string]any{
									"04": "04",
								},
							}, 5, 6, 7, 8, 9,
						},
					},
				},
				&struct {
					Two map[string]any
				}{
					Two: map[string]any{
						"Three": []any{
							0, 1, 2, 3, []any{4}, 5, 6, 7, 8, 9,
						},
					},
				},
				[]any{
					map[string]any{
						"Three": []any{
							0, 1, 2, 3, []any{
								struct {
									TwentyFour int
								}{
									TwentyFour: 24,
								},
							}, 5, 6, 7, 8, 9,
						},
					},
				},
			},
			Path:          "$..Three[::2]..TwentyFour",
			ExpectedOk:    1,
			ExpectedValue: []any{24},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Root: map[string]any{
				"one": []any{
					1,
					map[string]any{
						"Two": []any{
							12,
						},
					},
					map[string]any{
						"three": []any{
							"four",
						},
					},
					map[string]any{
						"five": []any{
							struct {
								Two []any
							}{
								Two: []any{
									13,
								},
							},
						},
					},
				},
			},
			Path:          "$.one..Two[*]",
			ExpectedOk:    2,
			ExpectedValue: []any{12, 13},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Root: []struct {
				Five  string
				Six   string
				Seven string
			}{
				{
					Five: "five",
					Six:  "0_six",
				},
				{
					Five:  "five",
					Seven: "seven",
				},
				{
					Five: "five",
					Six:  "2_six",
				},
			},
			Path:          "$..Six",
			ExpectedOk:    3,
			ExpectedValue: []any{"0_six", "", "2_six"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&GetData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Root: map[string]any{
				"one": 1,
				"two": 2,
				"three": []any{
					"four",
					map[int]int{
						1: 1,
						2: 2,
					},
				},
			},
			Path:          "$.three[1].['1']",
			ExpectedOk:    1,
			ExpectedValue: 1,
		},
	) {
		return
	}
}
