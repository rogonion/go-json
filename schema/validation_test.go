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

func TestSchema_ValidateData(t *testing.T) {
	for testData := range validationDataTestData {
		ok, err := NewValidation().WithValidateOnFirstMatch(testData.ValidateOnFirstMatch).WithCustomValidators(testData.Validators).ValidateData(testData.Data, testData.Schema)
		if ok != testData.ExpectedOk {
			t.Error(
				testData.TestTitle, "\n",
				"expected ok=", testData.ExpectedOk, "got=", ok, "\n",
				"schema=", testData.Schema, "\n",
				"data=", core.JsonStringifyMust(testData.Data), "\n",
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

type validationData struct {
	internal.TestData
	Schema               Schema
	Data                 any
	ValidateOnFirstMatch bool
	Validators           Validators
	ExpectedOk           bool
}

func validationDataTestData(yield func(data *validationData) bool) {
	testCaseIndex := 1
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate simple primitive", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			Data:       "this is a string",
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate invalid simple primitive", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
			Data:       0.0,
			ExpectedOk: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test valid data", testCaseIndex),
			},
			Schema: DynamicUserSchema(),
			Data: User{
				ID:    1,
				Name:  "John Doe",
				Email: "john@example.com",
			},
			ValidateOnFirstMatch: true,
			ExpectedOk:           true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test invalid data", testCaseIndex),
			},
			Schema:               DynamicUserSchema(),
			Data:                 123,
			ValidateOnFirstMatch: true,
			ExpectedOk:           false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test valid list of products", testCaseIndex),
			},
			Schema: ListOfProductsSchema(),
			Data: []Product{
				{
					ID:    1,
					Name:  "Laptop",
					Price: 1200.0,
				},
				{
					ID:    2,
					Name:  "R2-D2",
					Price: 25.5,
				},
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test invalid value in list of products", testCaseIndex),
			},
			Schema: ListOfProductsSchema(),
			Data: []any{
				Product{
					ID:    1,
					Name:  "Laptop",
					Price: 1200.0,
				},
				Product{
					ID:    2,
					Name:  "R2-D2",
					Price: 25.5,
				},
				12,
			},
			ExpectedOk: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test valid map of pointer to users", testCaseIndex),
			},
			Schema: MapUserSchema(),
			Data: map[int]*User{
				1: {
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				},
				2: {
					ID:    2,
					Name:  "R2-D2",
					Email: "r2d2@email.com",
				},
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test invalid value in map", testCaseIndex),
			},
			Schema: MapUserSchema(),
			Data: map[int]any{
				1: &User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				},
				2: &User{
					ID:    2,
					Name:  "R2-D2",
					Email: "r2d2@email.com",
				},
				3: "invalid value",
			},
			ExpectedOk: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test map of any value type", testCaseIndex),
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
			Data: map[string]interface{}{
				"name":      "John Doe",
				"age":       12,
				"isStudent": true,
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: test deeply nested data", testCaseIndex),
			},
			Schema: ListOfNestedItemSchema(),
			Data: []NestedItem{
				{
					ID: 1,
					MapData: map[string]interface{}{
						"name":       "Item A",
						"properties": map[string]interface{}{"size": 10, "color": "red"},
					},
					ListData: []interface{}{
						"value1",
						123,
						true,
						map[string]interface{}{
							"nestedKey": "nestedValue",
						},
					},
				},
				{
					ID: 2,
					MapData: map[string]interface{}{
						"name":       "Item B",
						"properties": map[string]interface{}{"size": 20, "color": "blue"},
					},
					ListData: []interface{}{
						map[string]interface{}{
							"anotherKey": "anotherValue",
						},
						"value2",
						456,
					},
				},
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: testing with global uuid custom validator for valid value", testCaseIndex),
			},
			Schema: UserWithUuidIdSchema(),
			Data: UserWithUuidId{
				ID: func() uuid.UUID {
					randomUuid, _ := uuid.NewV7()
					return randomUuid
				}(),
				Profile: UserProfile2{
					Name:       "John Doe",
					Age:        12,
					Country:    "USA",
					Occupation: "busy",
				},
			},
			Validators: Validators{
				reflect.TypeOf(uuid.UUID{}): core.Ptr(Pgxuuid{}),
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: testing with node specific uuid custom validator for invalid nil value", testCaseIndex),
			},
			Schema: UserWithUuidIdSchema(),
			Data: UserWithUuidId{
				ID: uuid.Nil,
				Profile: UserProfile2{
					Name:       "John Doe",
					Age:        12,
					Country:    "USA",
					Occupation: "busy",
				},
			},
			ExpectedOk: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate tuple (slice with specific schemas at indices)", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind: reflect.Slice,
				Type: reflect.TypeOf([]any{}),
				ChildNodes: ChildNodes{
					"0": &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
					"1": &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
				},
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{Kind: reflect.Interface},
			},
			Data:       []any{"Title", 42, true}, // 0: String (OK), 1: Int (OK), 2: Interface (OK)
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate mixed map (specific keys + generic rest)", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind: reflect.Map,
				Type: reflect.TypeOf(map[string]any{}),
				ChildNodes: ChildNodes{
					"id": &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
				},
				ChildNodesAssociativeCollectionEntriesKeySchema:   &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
				ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
			},
			Data: map[string]any{
				"id":          101,       // Matches specific ChildNode
				"description": "generic", // Matches generic ValueSchema
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate Nilable=true", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind:    reflect.String,
				Type:    reflect.TypeOf(""),
				Nilable: true,
			},
			Data:       nil,
			ExpectedOk: true,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate Nilable=false (default)", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			Data:       nil,
			ExpectedOk: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate strict map (ChildNodesMustBeValid=true) with missing key", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind: reflect.Map,
				Type: reflect.TypeOf(map[string]any{}),
				ChildNodes: ChildNodes{
					"required_key": &DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
				},
				ChildNodesMustBeValid: true, // Enforces that "required_key" must exist
			},
			Data: map[string]any{
				"other_key": "value",
			},
			ExpectedOk: false,
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d: Validate fixed-size Array", testCaseIndex),
			},
			Schema: &DynamicSchemaNode{
				Kind:                                     reflect.Array,
				Type:                                     reflect.TypeOf([3]int{}),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
			},
			Data:       [3]int{1, 2, 3},
			ExpectedOk: true,
		},
	) {
		return
	}
}
