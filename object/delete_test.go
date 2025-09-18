package object

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

func TestObject_Delete(t *testing.T) {

	for testData := range DeleteTestData {
		getValue := NewDeleteValue(schema.NewProcessor(true, nil, nil))
		res, ok, err := getValue.Delete(testData.Root, testData.Path)
		if ok != testData.ExpectedOk {
			t.Error(
				"expected ok=", testData.ExpectedOk, "got=", ok, "\n",
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

type DeleteData struct {
	internal.TestData
	Root          any
	Path          path.JSONPath
	ExpectedOk    uint64
	ExpectedValue any
}

func DeleteTestData(yield func(data *DeleteData) bool) {
	if !yield(
		&DeleteData{
			Root: []any{
				map[string]any{"User": &User{Name: "Alice"}},
				&struct {
					ID      int
					Details map[string]any
					Items   []struct {
						Name  string
						Value int
					}
					User *User
				}{
					User: &User{Name: "Bob"},
					Items: []struct {
						Name  string
						Value int
					}{
						{Name: "Item 1", Value: 10000},
						{Name: "Item 2", Value: 10000},
					},
				},
			},
			Path:       "$..Items[*].Value",
			ExpectedOk: 2,
			ExpectedValue: []any{
				map[string]any{"User": &User{Name: "Alice"}},
				&struct {
					ID      int
					Details map[string]any
					Items   []struct {
						Name  string
						Value int
					}
					User *User
				}{
					User: &User{Name: "Bob"},
					Items: []struct {
						Name  string
						Value int
					}{
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&DeleteData{
			Root: []any{
				map[string]any{"User": &User{Name: "OneName"}},
				&struct {
					ID      int
					Details map[string]any
					Items   []struct {
						Name  string
						Value int
					}
					User *User
				}{
					User: &User{Name: "OneName"},
					Items: []struct {
						Name  string
						Value int
					}{
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
			Path:       "$..User..Name",
			ExpectedOk: 2,
			ExpectedValue: []any{
				map[string]any{"User": &User{}},
				&struct {
					ID      int
					Details map[string]any
					Items   []struct {
						Name  string
						Value int
					}
					User *User
				}{
					User: &User{},
					Items: []struct {
						Name  string
						Value int
					}{
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
		},
	) {
		return
	}
	if !yield(
		&DeleteData{
			Root: []any{
				map[string]any{"User": &User{Name: "OneName"}},
				&struct {
					ID      int
					Details map[string]any
					Items   []struct {
						Name  string
						Value int
					}
					User *User
				}{
					User: &User{Name: "OneName"},
					Items: []struct {
						Name  string
						Value int
					}{
						{Name: "OneName"},
						{Name: "OneName"},
					},
				},
			},
			Path:       "$..Name",
			ExpectedOk: 4,
			ExpectedValue: []any{
				map[string]any{"User": &User{}},
				&struct {
					ID      int
					Details map[string]any
					Items   []struct {
						Name  string
						Value int
					}
					User *User
				}{
					User: &User{},
					Items: []struct {
						Name  string
						Value int
					}{
						{},
						{},
					},
				},
			},
		},
	) {
		return
	}
	if !yield(
		&DeleteData{
			Root: map[string]any{
				"Addresses": []any{
					nil,
					map[string]any{
						"City": []any{
							nil,
							nil,
							nil,
							nil,
							nil,
							map[string]any{
								"location":     "LocationSublocation",
								"sub-location": "LocationSublocation",
							},
						},
					},
					nil,
					nil,
					map[string]any{
						"City": []any{
							nil,
							nil,
							nil,
							nil,
							nil,
							map[string]any{
								"location":     "LocationSublocation",
								"sub-location": "LocationSublocation",
							},
						},
					},
				},
			},
			Path:       "$.Addresses[1,4].City[5]['location','sub-location']",
			ExpectedOk: 4,
			ExpectedValue: map[string]any{
				"Addresses": []any{
					nil,
					map[string]any{
						"City": []any{
							nil,
							nil,
							nil,
							nil,
							nil,
							map[string]any{},
						},
					},
					nil,
					nil,
					map[string]any{
						"City": []any{
							nil,
							nil,
							nil,
							nil,
							nil,
							map[string]any{},
						},
					},
				},
			},
		},
	) {
		return
	}
	if !yield(
		&DeleteData{
			Root: &ComplexData{
				User: User{
					Name:  "NameEmail",
					Email: "NameEmail",
				},
			},
			Path:       "$.User['Name','Email']",
			ExpectedOk: 2,
			ExpectedValue: &ComplexData{
				User: User{},
			},
		},
	) {
		return
	}

	if !yield(
		&DeleteData{
			Root: &[4]*Address{
				{
					City: "CityDouble",
				},
				{
					City: "City1",
				},
				{
					City: "CityDouble",
				},
				{
					City: "City3",
				},
			},
			Path:       "$[::2]City",
			ExpectedOk: 2,
			ExpectedValue: &[4]*Address{
				{},
				{
					City: "City1",
				},
				{},
				{
					City: "City3",
				},
			},
		},
	) {
		return
	}
	if !yield(
		&DeleteData{
			Root: &ComplexData{
				Items: []struct {
					Name  string
					Value int
				}{
					{},
					{
						Name: "I am User",
					},
				},
			},
			Path:       "$.Items[1].Name",
			ExpectedOk: 1,
			ExpectedValue: &ComplexData{
				Items: []struct {
					Name  string
					Value int
				}{
					{},
					{},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&DeleteData{
			Root:          map[string]any{"address": map[string]any{"address": "test"}},
			Path:          "$.address",
			ExpectedOk:    1,
			ExpectedValue: map[string]any{},
		},
	) {
		return
	}

	if !yield(
		&DeleteData{
			Root:          []any{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 10},
			Path:          "$.[10]",
			ExpectedOk:    1,
			ExpectedValue: []any{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
		},
	) {
		return
	}
}
