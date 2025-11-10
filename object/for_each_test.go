package object

import (
	"reflect"
	"testing"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

func TestObject_ForEachValue(t *testing.T) {
	for testData := range ForEachValueTestData {
		res := make([]any, 0)
		NewObject(testData.Object).ForEach(testData.Path, func(jsonPath path.RecursiveDescentSegment, value any) bool {
			res = append(res, value)
			return false
		})

		if !reflect.DeepEqual(res, testData.Expected) {
			t.Error(
				"expected res to be equal to testData.Expected\n",
				"path=", testData.Path, "\n",
				"res=", core.JsonStringifyMust(res), "\n",
				"JSON testData.Expected=", core.JsonStringifyMust(testData.Expected),
			)
		}
	}
}

type ForEachData struct {
	internal.TestData
	Object   any
	Path     path.JSONPath
	Expected any
}

func ForEachValueTestData(yield func(data *ForEachData) bool) {
	if !yield(
		&ForEachData{
			Object: []any{
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
			Path:     "$..Name",
			Expected: []any{"Alice", "Item 1", "Item 2", "Bob"},
		},
	) {
		return
	}

	if !yield(
		&ForEachData{
			Object: []any{
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
			Path:     "$..Three[::2]['TwentyFour','04']",
			Expected: []any{[]int{0, 4}, 24},
		},
	) {
		return
	}

	if !yield(
		&ForEachData{
			Object: []any{
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
			Path:     "$..Three[::2]['TwentyFour','04']",
			Expected: []any{"04", 24},
		},
	) {
		return
	}

	if !yield(
		&ForEachData{
			Object: []any{
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
			Path:     "$..Three[::2]..TwentyFour",
			Expected: []any{24},
		},
	) {
		return
	}

	if !yield(
		&ForEachData{
			Object: map[string]any{
				"child": []any{
					nil,
					map[string]any{
						"nectar": map[string]any{
							"willy": []any{
								nil, nil, nil, nil, nil, []any{nil, nil, nil, "smitty"}, nil, nil, nil, nil, nil, nil, nil, nil, nil,
								map[string]any{
									"oxford": "willow",
									"bee":    []any{nil, nil, nil, 5},
								},
							},
							"two": []any{1, 2, 3, 4, 5},
						},
						"mocha": struct {
							Nacho  string
							Amount float64
						}{
							Nacho:  "cheese",
							Amount: 45.56,
						},
					},
					nil,
					nil,
					"another child",
				},
			},
			Path:     "$.child[20].wind",
			Expected: []any{},
		},
	) {
		return
	}

	if !yield(
		&ForEachData{
			Object: map[string]any{
				"child": []any{
					nil,
					map[string]any{
						"nectar": map[string]any{
							"willy": []any{
								nil, nil, nil, nil, nil, []any{nil, nil, nil, "smitty"}, nil, nil, nil, nil, nil, nil, nil, nil, nil,
								map[string]any{
									"oxford": "willow",
									"bee":    []any{nil, nil, nil, 5},
								},
							},
							"two": []any{1, 2, 3, 4, 5},
						},
						"mocha": struct {
							Nacho  string
							Amount float64
						}{
							Nacho:  "cheese",
							Amount: 45.56,
						},
					},
					nil,
					nil,
					"another child",
				},
			},
			Path:     "$.child[1].nectar.two[*]",
			Expected: []any{1, 2, 3, 4, 5},
		},
	) {
		return
	}
}
