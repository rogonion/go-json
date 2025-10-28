package object

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

func TestObject_Get(t *testing.T) {
	for testData := range GetTestData {
		res, ok, err := NewObject(testData.Root).Get(testData.Path)
		if ok != testData.ExpectedOk {
			t.Error(
				"expected ok=", testData.ExpectedOk, "got ok=", ok, "\n",
				"path=", testData.Path,
			)
		}

		if err != nil && testData.LogErrorsIfExpectedNotOk {
			var objectProcessorError *Error
			if errors.As(err, &objectProcessorError) {
				t.Error("Test Tile:", testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					objectProcessorError.String(), "\n",
					"-----------------------",
				)
			}
		}

		if !reflect.DeepEqual(res, testData.ExpectedValue) {
			t.Error(
				"res not equal to testData.ExpectedValue\n",
				"Path", testData.Path, "\n",
				"res=", internal.JsonStringifyMust(res), "\n",
				"JSON testData.Expected=", internal.JsonStringifyMust(testData.ExpectedValue),
			)
		}
	}
}

type GetData struct {
	internal.TestData
	Root          any
	Path          path.JSONPath
	ExpectedOk    bool
	ExpectedValue any
}

func GetTestData(yield func(data *GetData) bool) {
	// --- Test Case 1: Complex nested path with maps and structs ---
	// This test navigates through nested maps and structs to a specific field.
	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: "Anytown",
		},
	) {
		return
	}

	// --- Test Case 2: Recursive descent with a wildcard to get all names ---
	// This tests if the function can find all "Name" fields regardless of nesting.
	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: []any{"Alice", "Item 1", "Item 2", "Bob"},
		},
	) {
		return
	}

	// --- Test Case 3: Union operator to get multiple fields from a single struct ---
	// This tests the `['fieldA', 'fieldB']` operator.
	if !yield(
		&GetData{
			Root: ComplexData{
				ID: 99,
				User: User{
					ID:    10,
					Name:  "Charlie",
					Email: "charlie@example.com",
				},
			},
			Path:          "$.User['ID', 'Name']",
			ExpectedOk:    true,
			ExpectedValue: []any{10, "Charlie"},
		},
	) {
		return
	}

	// --- Test Case 4: Deep path with a slice index and field extraction ---
	// This tests a combination of operators to get a specific value from a deeply nested slice.
	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: 20,
		},
	) {
		return
	}

	// --- Test Case 5: Path that should not find a match ---
	// This test ensures the function correctly returns false and a nil value for an invalid path.
	if !yield(
		&GetData{
			Root: map[string]any{
				"store": map[string]any{"book": "fiction"},
			},
			Path:          "$.store.magazine",
			ExpectedOk:    false,
			ExpectedValue: nil,
		},
	) {
		return
	}

	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: []any{[]int{0, 4}, 24},
		},
	) {
		return
	}

	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: []any{"04", 24},
		},
	) {
		return
	}

	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: []any{24},
		},
	) {
		return
	}

	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: []any{12, 13},
		},
	) {
		return
	}

	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: []any{"0_six", "", "2_six"},
		},
	) {
		return
	}

	if !yield(
		&GetData{
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
			ExpectedOk:    true,
			ExpectedValue: 1,
		},
	) {
		return
	}
}
