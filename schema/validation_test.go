package schema

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/rogonion/go-json/internal"
)

func TestSchema_ValidateData(t *testing.T) {
	for testData := range validationDataTestData {
		ok, err := NewValidation().WithValidateOnFirstMatch(testData.ValidateOnFirstMatch).WithCustomValidators(testData.Validators).ValidateData(testData.Data, testData.Schema)
		if ok != testData.ExpectedOk {
			t.Error(
				"expected ok=", testData.ExpectedOk, "got=", ok, "\n",
				"schema=", testData.Schema, "\n",
				"data=", internal.JsonStringifyMust(testData.Data), "\n",
			)
			var schemaError *Error
			if errors.As(err, &schemaError) {
				t.Error("Test Tile:", testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					schemaError.String(), "\n",
					"-----------------------",
				)
			}
		}

		if err != nil && testData.LogErrorsIfExpectedNotOk {
			var schemaError *Error
			if errors.As(err, &schemaError) {
				t.Log(
					"-----Error Details-----", "\n",
					"Test Title:", testData.TestTitle, "\n",
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
	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "Validate simple primitive",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "Validate invalid simple primitive",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test valid data",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test invalid data",
			},
			Schema:               DynamicUserSchema(),
			Data:                 123,
			ValidateOnFirstMatch: true,
			ExpectedOk:           false,
		},
	) {
		return
	}

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test valid list of products",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test invalid value in list of products",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test valid map of pointer to users",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test invalid value in map",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test map of any value type",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "test deeply nested data",
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

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "testing with global uuid custom validator for valid value",
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
				reflect.TypeOf(uuid.UUID{}): internal.Ptr(Pgxuuid{}),
			},
			ExpectedOk: true,
		},
	) {
		return
	}

	if !yield(
		&validationData{
			TestData: internal.TestData{
				TestTitle: "testing with node specific uuid custom validator for invalid nil value",
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
}
