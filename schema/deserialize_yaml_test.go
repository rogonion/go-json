package schema

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
)

func TestSchema_DeserializeFromYaml(t *testing.T) {
	for testData := range deserializeYamlDataTestData {
		var res any
		err := NewDeserialization().WithCustomConverters(testData.Converters).FromYAML([]byte(testData.Source), testData.Schema, &res)
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

func deserializeYamlDataTestData(yield func(data *deserializeData) bool) {
	testCaseIndex := 1
	if !yield(
		&deserializeData{
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

	testCaseIndex++
	if !yield(
		&deserializeData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

	testCaseIndex++
	if !yield(
		&deserializeData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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

	testCaseIndex++
	if !yield(
		&deserializeData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
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
