package schema

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
)

func TestSchema_Convert(t *testing.T) {
	for testData := range conversionDataTestData {
		var res any
		err := NewConversion().WithCustomConverters(testData.Converters).Convert(testData.Source, testData.Schema, &res)
		if testData.ExpectedOk && err != nil {
			t.Error(
				testData.TestTitle, "\n",
				"expected ok=", testData.ExpectedOk, "got error=", err, "\n",
				"schema=", testData.Schema, "\n",
				"data=", core.JsonStringifyMust(testData.Source), "\n",
			)
			var schemaError *core.Error
			if errors.As(err, &schemaError) {
				t.Error(
					testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					schemaError.String(), "\n",
					"-----------------------",
				)
			}
		} else {
			if !reflect.DeepEqual(res, testData.ExpectedData) {
				t.Error(
					testData.TestTitle, "\n",
					"expected res to be equal to testData.ExpectedData\n",
					"schema=", testData.Schema, "\n",
					"res", core.JsonStringifyMust(res), "\n",
					"testData.ExpectedData", core.JsonStringifyMust(testData.ExpectedData),
				)
			}
		}

		if err != nil && testData.LogErrorsIfExpectedNotOk {
			var schemaError *core.Error
			if errors.As(err, &schemaError) {
				t.Log(
					testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					schemaError.String(), "\n",
					"-----------------------",
				)
			}
		}
	}
}

type conversionData struct {
	internal.TestData
	Schema       Schema
	Source       any
	Converters   Converters
	ExpectedOk   bool
	ExpectedData any
}

func conversionDataTestData(yield func(data *conversionData) bool) {
	testCaseIndex := 1
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Schema:       &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
			Source:       "123",
			ExpectedOk:   true,
			ExpectedData: 123,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Schema:       &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
			Source:       456,
			ExpectedOk:   true,
			ExpectedData: "456",
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Schema:       &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
			Source:       123.45,
			ExpectedOk:   true,
			ExpectedData: 123,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Schema:       &DynamicSchemaNode{Kind: reflect.Float64, Type: reflect.TypeOf(float64(0))},
			Source:       "25.7",
			ExpectedOk:   true,
			ExpectedData: 25.7,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Schema:       &DynamicSchemaNode{Kind: reflect.Bool, Type: reflect.TypeOf(true)},
			Source:       25,
			ExpectedOk:   true,
			ExpectedData: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

		testCaseIndex++
		if !yield(
			&conversionData{
				TestData: internal.TestData{
					TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
				},
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

		testCaseIndex++
		if !yield(
			&conversionData{
				TestData: internal.TestData{
					TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
				},
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

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Convert JSON string directly to Struct", testCaseIndex),
			},
			Schema:       UserProfile2Schema(),
			Source:       `{"Name": "James Bond", "Age": 40, "Country": "UK", "Occupation": "Agent"}`,
			ExpectedOk:   true,
			ExpectedData: UserProfile2{Name: "James Bond", Age: 40, Country: "UK", Occupation: "Agent"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Convert JSON string directly to Slice", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind:                                     reflect.Slice,
				Type:                                     reflect.TypeOf([]int{}),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
			},
			Source:       `[1, 2, 3, 4]`,
			ExpectedOk:   true,
			ExpectedData: []int{1, 2, 3, 4},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&conversionData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Convert using Custom Converter (UUID string to UUID struct)", testCaseIndex),
			},
			Schema: UserWithUuidIdSchema(),
			Source: map[string]any{
				"ID": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
				"Profile": map[string]any{
					"Name":       "Alice",
					"Age":        30,
					"Country":    "Wonderland",
					"Occupation": "Explorer",
				},
			},
			ExpectedOk: true,
			ExpectedData: UserWithUuidId{
				ID:      uuid.FromStringOrNil("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
				Profile: UserProfile2{Name: "Alice", Age: 30, Country: "Wonderland", Occupation: "Explorer"},
			},
		},
	) {
		return
	}
}

func TestSchema_ConvertStoreResultInTypedDestination(t *testing.T) {
	cvt := NewConversion()

	var schema Schema = &DynamicSchemaNode{
		Kind: reflect.Int64,
		Type: reflect.TypeOf(int64(0)),
	}
	numberCast := int64(0)
	err := cvt.Convert(0.0, schema, &numberCast)
	if err != nil {
		t.Fatal("Convert float64 to int64 failed", err)
	}

	schema = AddressSchema()
	pointerToAddress := new(Address)
	err = cvt.Convert(map[string]any{"Street": "Turnkit Boulevard", "City": "NewYork", "ZipCode": "1234"}, schema, pointerToAddress)
	if err != nil {
		t.Fatal("Convert Address map to struct failed", err)
	}
}
