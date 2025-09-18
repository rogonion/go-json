package path

import (
	"reflect"
	"testing"
)

func TestPath_ParseReverse(t *testing.T) {
	for testData := range ParseDataTestData {
		result := testData.Path.Parse()

		if !reflect.DeepEqual(result.String(), string(testData.Path)) {
			t.Error(
				"expected=", testData.Path, "\n",
				"got=", result.String(),
			)
		}
	}
}

func TestPath_Parse(t *testing.T) {
	for testData := range ParseDataTestData {
		result := testData.Path.Parse()

		if !reflect.DeepEqual(result, testData.ExpectedPathSegment) {
			t.Error(
				"expected=", testData.ExpectedPathSegment, "\n",
				"got=", result,
			)
		}
	}
}

type ParseData struct {
	Path                JSONPath
	ExpectedPathSegment RecursiveDescentSegments
}

func ParseDataTestData(yield func(data *ParseData) bool) {
	if !yield(
		&ParseData{
			Path: "$[1,3,5]",
			ExpectedPathSegment: RecursiveDescentSegments{
				{
					{
						Key:               "$",
						IsKeyRoot:         true,
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
					{
						UnionSelector: []*CollectionMemberSegment{
							{
								Index:   1,
								IsIndex: true,
							},
							{
								Index:   3,
								IsIndex: true,
							},
							{
								Index:   5,
								IsIndex: true,
							},
						},
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ParseData{
			Path: "$['report-data']..['total.sum']",
			ExpectedPathSegment: RecursiveDescentSegments{
				{
					{
						Key:               "$",
						IsKeyRoot:         true,
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
					{
						Key:               "report-data",
						IsKey:             true,
						ExpectAssociative: true,
					},
				},
				{
					{
						Key:               "total.sum",
						IsKey:             true,
						ExpectAssociative: true,
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ParseData{
			Path: "$.transactions[1:5:2].*",
			ExpectedPathSegment: RecursiveDescentSegments{
				{
					{
						Key:               "$",
						IsKeyRoot:         true,
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
					{
						Key:               "transactions",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						LinearCollectionSelector: &LinearCollectionSelector{
							Start:   1,
							IsStart: true,
							End:     5,
							IsEnd:   true,
							Step:    2,
							IsStep:  true,
						},
						ExpectLinear: true,
					},
					{
						Key:               "*",
						IsKeyIndexAll:     true,
						ExpectAssociative: true,
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ParseData{
			Path: "$.data.user.preferences['theme-settings','font-size']",
			ExpectedPathSegment: RecursiveDescentSegments{
				{
					{
						Key:               "$",
						IsKeyRoot:         true,
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
					{
						Key:               "data",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						Key:               "user",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						Key:               "preferences",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						UnionSelector: []*CollectionMemberSegment{
							{
								Key:   "theme-settings",
								IsKey: true,
							},
							{
								Key:   "font-size",
								IsKey: true,
							},
						},
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ParseData{
			Path: "$..['1st_category'].name",
			ExpectedPathSegment: RecursiveDescentSegments{
				{
					{
						Key:               "$",
						IsKeyRoot:         true,
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
				},
				{
					{
						Key:               "1st_category",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						Key:               "name",
						IsKey:             true,
						ExpectAssociative: true,
					},
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ParseData{
			Path: "$.products['item-details'].dimensions[0].width",
			ExpectedPathSegment: RecursiveDescentSegments{
				{
					{
						Key:               "$",
						IsKeyRoot:         true,
						ExpectLinear:      true,
						ExpectAssociative: true,
					},
					{
						Key:               "products",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						Key:               "item-details",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						Key:               "dimensions",
						IsKey:             true,
						ExpectAssociative: true,
					},
					{
						Index:        0,
						IsIndex:      true,
						ExpectLinear: true,
					},
					{
						Key:               "width",
						IsKey:             true,
						ExpectAssociative: true,
					},
				},
			},
		},
	) {
		return
	}
}
