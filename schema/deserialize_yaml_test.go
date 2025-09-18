package schema

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/rogonion/go-json/internal"
)

func TestSchema_DeserializeFromYaml(t *testing.T) {
	for testData := range deserializeYamlDataTestData {
		schema := new(Processor)
		schema.SetValidateOnFirstMatch(testData.ValidateOnFirstMatch)
		schema.SetValidators(testData.Validators)
		schema.SetConverters(testData.Converters)

		var res any
		err := schema.DeserializeFromYaml([]byte(testData.Source), testData.Schema, &res)
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

func deserializeYamlDataTestData(yield func(data *deserializeData) bool) {
	if !yield(
		&deserializeData{
			Schema: &DynamicSchemaNode{
				Kind: reflect.Map,
				Type: reflect.TypeOf(map[string]interface{}{}),
				ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
				ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{Kind: reflect.Interface},
			},
			Source: strings.TrimSpace(`
name: Test User
id: 101
is_active: true
tags:
- alpha
- beta
- gamma
`),
			ExpectedOk: true,
			ExpectedData: map[string]interface{}{
				"name":      "Test User",
				"id":        101,
				"is_active": true,
				"tags":      []any{"alpha", "beta", "gamma"},
			},
		},
	) {
		return
	}

	if !yield(
		&deserializeData{
			Schema: ListOfShapesSchema(),
			Source: strings.TrimSpace(`
- Radius: 5
- Side: 10
- Radius: 7.5
`),
			ExpectedOk: true,
			ExpectedData: []Shape{
				&Circle{
					Radius: 5.0,
				},
				&Square{
					Side: 10.0,
				},
				&Circle{
					Radius: 7.5,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&deserializeData{
			Schema: ListOfShapesSchema(),
			Source: strings.TrimSpace(`
- Radius: 5
- Base: 10
`),
			ExpectedOk: false,
		},
	) {
		return
	}

	if !yield(
		&deserializeData{
			Schema: UserWithUuidIdSchema(),
			Source: strings.TrimSpace(`
ID: c1f20d6c-6a1e-4b9a-8a4b-91d5a7d7d5a7
Profile:
    Name: Jane
    Age: 28
    Occupation: Manager
`),
			ExpectedOk: true,
			ExpectedData: UserWithUuidId{
				ID: uuid.FromStringOrNil("c1f20d6c-6a1e-4b9a-8a4b-91d5a7d7d5a7"),
				Profile: UserProfile2{
					Name:       "Jane",
					Age:        28,
					Occupation: "Manager",
				},
			},
		},
	) {
		return
	}
}
