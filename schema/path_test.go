package schema

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

func TestSchemaPath_GetSchemaAtPath(t *testing.T) {
	for testData := range getSchemaAtPathDataTestData {
		res, err := GetSchemaAtPath(testData.Path, testData.Schema)

		if testData.ExpectedOk && err != nil {
			t.Error(
				"expected ok=", testData.ExpectedOk, "got error=", err, "\n",
				"schema=", testData.Schema, "\n",
				"path=", testData.Path, "\n",
			)

			var schemaPathError *core.Error
			if errors.As(err, &schemaPathError) {
				t.Error("Test Tile:", testData.TestTitle, "\n",
					"-----Error Details-----", "\n",
					schemaPathError.String(), "\n",
					"-----------------------",
				)
			}
		} else {
			if !reflect.DeepEqual(res, testData.ExpectedData) {
				t.Error(
					"expected res to be equal to testData.ExpectedData\n",
					"schema=", testData.Schema, "\n",
					"res", res, "\n",
					"testData.ExpectedData", testData.ExpectedData,
				)
			}
		}

		if err != nil && testData.LogErrorsIfExpectedNotOk {
			var schemaError *core.Error
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

type getSchemaAtPathData struct {
	internal.TestData
	Schema       Schema
	Path         path.JSONPath
	ExpectedOk   bool
	ExpectedData *DynamicSchemaNode
}

func getSchemaAtPathDataTestData(yield func(data *getSchemaAtPathData) bool) {
	if !yield(
		&getSchemaAtPathData{
			Schema:     JsonMapSchema(),
			Path:       "$.Name",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Interface,
				AssociativeCollectionEntryKeySchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     JsonMapSchema(),
			Path:       "$.Addresses[1].Zipcode",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Interface,
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     ShapeSchema(),
			Path:       "$.Side",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Float64,
				Type: reflect.TypeOf(0.0),
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     UserWithAddressSchema(),
			Path:       "$.Address.ZipCode",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Pointer,
				Type: reflect.TypeOf(core.Ptr("")),
				ChildNodesPointerSchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     CompanySchema(),
			Path:       "$.Employees[2].ID",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     EmployeeSchema(),
			Path:       "$.ProjectHours['1']",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
				AssociativeCollectionEntryKeySchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     EmployeeSchema(),
			Path:       "$.ProjectHours['1'].two",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
		},
	) {
		return
	}

	if !yield(
		&getSchemaAtPathData{
			Schema:     ListOfShapesSchema(),
			Path:       "$[1].Radius",
			ExpectedOk: true,
			ExpectedData: &DynamicSchemaNode{
				Kind: reflect.Float64,
				Type: reflect.TypeOf(0.0),
			},
		},
	) {
		return
	}
}
