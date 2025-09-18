package schema

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
)

func TestSchema_Convert(t *testing.T) {
	for testData := range convertDataTestData {
		schema := new(Processor)
		schema.SetValidateOnFirstMatch(testData.ValidateOnFirstMatch)
		schema.SetValidators(testData.Validators)
		schema.SetConverters(testData.Converters)

		var res any
		err := schema.Convert(testData.Source, testData.Schema, &res)
		if testData.ExpectedOk && err != nil {
			t.Error(
				"expected ok=", testData.ExpectedOk, "got error=", err, "\n",
				"schema=", testData.Schema, "\n",
				"data=", internal.JsonStringifyMust(testData.Source), "\n",
			)
			var schemaError *Error
			if errors.As(err, &schemaError) {
				t.Error("Test Tile:", testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					schemaError.String(), "\n",
					"-----------------------",
				)
			}
		} else {
			if !reflect.DeepEqual(res, testData.ExpectedData) {
				t.Error(
					"expected res to be equal to testData.ExpectedData\n",
					"schema=", testData.Schema, "\n",
					"res", internal.JsonStringifyMust(res), "\n",
					"testData.ExpectedData", internal.JsonStringifyMust(testData.ExpectedData),
				)
			}
		}

		if err != nil && testData.LogErrorsIfExpectedNotOk {
			var schemaError *Error
			if errors.As(err, &schemaError) {
				t.Log(
					"-----Error Details-----", "\n",
					"Test Tile:", testData.TestTitle, "\n",
					schemaError.String(), "\n",
					"-----------------------",
				)
			}
		}
	}
}

type convertData struct {
	internal.TestData
	Schema               Schema
	Source               any
	ValidateOnFirstMatch bool
	Validators           map[reflect.Type]Validator
	Converters           map[reflect.Type]Converter
	ExpectedOk           bool
	ExpectedData         any
}

func convertDataTestData(yield func(data *convertData) bool) {
	if !yield(
		&convertData{
			Schema: DynamicUserSchema(),
			Source: map[string]interface{}{
				"ID":    "1",
				"Name":  "John Doe",
				"Email": "john.doe@email.com",
			},
			ExpectedOk: true,
			ExpectedData: User{
				ID:    1,
				Name:  "John Doe",
				Email: "john.doe@email.com",
			},
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema: &DynamicSchemaNode{
				Kind: reflect.Map,
				Type: reflect.TypeOf(map[int]int{}),
				ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
					Kind: reflect.Int,
					Type: reflect.TypeOf(0),
				},
				ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
					Kind: reflect.Int,
					Type: reflect.TypeOf(0),
				},
			},
			Source: map[string]string{
				"1": "1",
				"2": "2",
				"3": "3",
			},
			ExpectedOk: true,
			ExpectedData: map[int]int{
				1: 1,
				2: 2,
				3: 3,
			},
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema:       &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
			Source:       "123",
			ExpectedOk:   true,
			ExpectedData: 123,
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema:       &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
			Source:       456,
			ExpectedOk:   true,
			ExpectedData: "456",
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema:       &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
			Source:       123.45,
			ExpectedOk:   true,
			ExpectedData: 123,
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema:       &DynamicSchemaNode{Kind: reflect.Float64, Type: reflect.TypeOf(float64(0))},
			Source:       "25.7",
			ExpectedOk:   true,
			ExpectedData: 25.7,
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema:       &DynamicSchemaNode{Kind: reflect.Bool, Type: reflect.TypeOf(true)},
			Source:       25,
			ExpectedOk:   true,
			ExpectedData: true,
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema: UserWithAddressSchema(),
			Source: map[string]interface{}{
				"Name": "Bob",
				"Address": map[string]interface{}{
					"Street": "123 Main St",
					"City":   "Anytown",
				},
			},
			ExpectedOk: true,
			ExpectedData: UserWithAddress{
				Name: "Bob",
				Address: &Address{
					Street: "123 Main St",
					City:   "Anytown",
				},
			},
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema: &DynamicSchemaNode{
				Kind:                                     reflect.Slice,
				Type:                                     reflect.TypeOf(make([]any, 0)),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{Kind: reflect.Interface},
			},
			Source:       [3]interface{}{1, "two", true},
			ExpectedOk:   true,
			ExpectedData: []any{1, "two", true},
		},
	) {
		return
	}

	if !yield(
		&convertData{
			Schema: &DynamicSchemaNode{
				Kind:                                     reflect.Array,
				Type:                                     reflect.TypeOf([3]any{}),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{Kind: reflect.Interface},
			},
			Source:       []interface{}{1, "two", true},
			ExpectedOk:   true,
			ExpectedData: [3]any{1, "two", true},
		},
	) {
		return
	}

	{
		type custom struct {
			One   bool
			Two   []any
			Three string
			Four  int
		}

		if !yield(
			&convertData{
				Schema: &DynamicSchemaNode{
					Kind: reflect.Struct,
					Type: reflect.TypeOf(custom{}),
					ChildNodes: map[string]Schema{
						"One": &DynamicSchemaNode{
							Kind: reflect.Bool,
							Type: reflect.TypeOf(true),
						},
						"Two": &DynamicSchemaNode{
							Kind: reflect.Slice,
							Type: reflect.TypeOf(make([]any, 0)),
							ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{
								Kind: reflect.Interface,
							},
						},
						"Three": &DynamicSchemaNode{
							Kind: reflect.String,
							Type: reflect.TypeOf(""),
						},
						"Four": &DynamicSchemaNode{
							Kind: reflect.Int,
							Type: reflect.TypeOf(0),
						},
					},
				},
				Source: map[string]any{
					"One":   true,
					"Two":   []any{1, 2, 3},
					"Three": "three",
					"Four":  "4",
				},
				ExpectedOk: true,
				ExpectedData: custom{
					One:   true,
					Two:   []any{1, 2, 3},
					Three: "three",
					Four:  4,
				},
			},
		) {
			return
		}

		if !yield(
			&convertData{
				Schema: &DynamicSchemaNode{
					Kind: reflect.Map,
					Type: reflect.TypeOf(map[string]interface{}{}),
					ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
						Kind: reflect.String,
						Type: reflect.TypeOf(""),
					},
					ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
						Kind: reflect.Interface,
					},
				},
				ExpectedData: map[string]any{
					"One":   true,
					"Two":   []any{1, 2, 3},
					"Three": "three",
					"Four":  4,
				},
				ExpectedOk: true,
				Source: custom{
					One:   true,
					Two:   []any{1, 2, 3},
					Three: "three",
					Four:  4,
				},
			},
		) {
			return
		}
	}
}
