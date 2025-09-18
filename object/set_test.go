package object

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

func TestObject_Set(t *testing.T) {

	for testData := range SetTestData {
		setValue := NewSetValue(schema.NewProcessor(true, nil, nil))

		res, ok, err := setValue.Set(testData.Root, testData.Path, testData.ValueToSet, testData.Schema)
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

type SetData struct {
	internal.TestData
	Root          any
	Path          path.JSONPath
	ValueToSet    any
	ExpectedOk    uint64
	ExpectedValue any
	Schema        schema.Schema
}

func SetTestData(yield func(data *SetData) bool) {
	if !yield(
		&SetData{
			Root:       nil,
			Path:       "$[5].Address.ZipCode",
			ValueToSet: "1234",
			Schema: &schema.DynamicSchemaNode{
				Kind: reflect.Slice,
				Type: reflect.TypeOf([]*UserProfile{}),
				ChildNodesLinearCollectionElementsSchema: &schema.DynamicSchemaNode{
					Kind:                    reflect.Pointer,
					Type:                    reflect.TypeOf(&UserProfile{}),
					ChildNodesPointerSchema: UserProfileSchema(),
				},
			},
			ExpectedOk:    1,
			ExpectedValue: []*UserProfile{nil, nil, nil, nil, nil, {Address: Address{ZipCode: internal.Ptr("1234")}}},
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:       nil,
			Path:       "$[10]",
			ValueToSet: 10,
			Schema: &schema.DynamicSchemaNode{
				Kind: reflect.Array,
				Type: reflect.TypeOf([11]int{}),
				ChildNodesLinearCollectionElementsSchema: &schema.DynamicSchemaNode{
					Kind: reflect.Int,
					Type: reflect.TypeOf(0),
				},
			},
			ExpectedOk:    1,
			ExpectedValue: [11]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10},
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:          nil,
			Path:          "$.Address.ZipCode",
			ValueToSet:    "1234",
			Schema:        UserProfileSchema(),
			ExpectedOk:    1,
			ExpectedValue: UserProfile{Address: Address{ZipCode: internal.Ptr("1234")}},
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:          nil,
			Path:          "$.ZipCode",
			ValueToSet:    "1234",
			Schema:        AddressSchema(),
			ExpectedOk:    1,
			ExpectedValue: Address{ZipCode: internal.Ptr("1234")},
		},
	) {
		return
	}

	if !yield(
		&SetData{
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
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
			Path:       "$..Items[*].Value",
			ExpectedOk: 2,
			ValueToSet: 10000,
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
						{Name: "Item 1", Value: 10000},
						{Name: "Item 2", Value: 10000},
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&SetData{
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
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
			Path:       "$..User..Name",
			ExpectedOk: 2,
			ValueToSet: "OneName",
			ExpectedValue: []any{
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
		},
	) {
		return
	}
	if !yield(
		&SetData{
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
						{Name: "Item 1"},
						{Name: "Item 2"},
					},
				},
			},
			Path:       "$..Name",
			ExpectedOk: 4,
			ValueToSet: "OneName",
			ExpectedValue: []any{
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
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:       nil,
			Path:       "$.Addresses[1,4].City[5]['location','sub-location']",
			ValueToSet: "LocationSublocation",
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
		},
	) {
		return
	}
	if !yield(
		&SetData{
			Root:       new(ComplexData),
			Path:       "$.User['Name','Email']",
			ValueToSet: "NameEmail",
			ExpectedOk: 2,
			ExpectedValue: &ComplexData{
				User: User{
					Name:  "NameEmail",
					Email: "NameEmail",
				},
			},
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root: &[4]*Address{
				{
					City: "City0",
				},
				{
					City: "City1",
				},
				{
					City: "City2",
				},
				{
					City: "City3",
				},
			},
			Path:       "$[::2]City",
			ValueToSet: "CityDouble",
			ExpectedOk: 2,
			ExpectedValue: &[4]*Address{
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
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:       new(ComplexData),
			Path:       "$.Items[1].Name",
			ValueToSet: "I am User",
			ExpectedOk: 1,
			ExpectedValue: &ComplexData{
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
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:          nil,
			Path:          "$.address",
			ValueToSet:    map[string]any{"address": "test"},
			ExpectedOk:    1,
			ExpectedValue: map[string]any{"address": map[string]any{"address": "test"}},
		},
	) {
		return
	}

	if !yield(
		&SetData{
			Root:          nil,
			Path:          "$.[10]",
			ValueToSet:    10,
			ExpectedOk:    1,
			ExpectedValue: []any{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 10},
		},
	) {
		return
	}
}
