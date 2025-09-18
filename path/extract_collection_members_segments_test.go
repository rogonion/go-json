package path

import (
	"reflect"

	"github.com/rogonion/go-json/internal"

	"testing"
)

func TestPath_ExtractCollectionMembersSegments(t *testing.T) {
	for testData := range ExtractCollectionMemberSegmentsTestData {
		result := testData.Segment.ExtractCollectionMemberSegments()

		if !reflect.DeepEqual(internal.JsonStringifyMust(&result), internal.JsonStringifyMust(&testData.ExpectedSegments)) {
			t.Error(
				"expected=", internal.JsonStringifyMust(testData.ExpectedSegments), "\n",
				"got=", result, "\n",
				"Test Data Segment=", testData.Segment,
			)
		}
	}
}

type ExtractCollectionMemberSegmentsData struct {
	Segment          JSONPath
	ExpectedSegments RecursiveDescentSegment
}

func ExtractCollectionMemberSegmentsTestData(yield func(data *ExtractCollectionMemberSegmentsData) bool) {
	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "storebicycle['item-code']",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "storebicycle",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Key:               "item-code",
					IsKey:             true,
					ExpectAssociative: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "['total.sum']",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "total.sum",
					IsKey:             true,
					ExpectAssociative: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "$['report-data']",
			ExpectedSegments: RecursiveDescentSegment{
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
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "widow[::2][4:5:]",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "widow",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					LinearCollectionSelector: &LinearCollectionSelector{
						Step:   2,
						IsStep: true,
					},
					ExpectLinear: true,
				},
				{
					LinearCollectionSelector: &LinearCollectionSelector{
						Start:   4,
						IsStart: true,
						End:     5,
						IsEnd:   true,
					},
					ExpectLinear: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "transactions[1:5:2]",
			ExpectedSegments: RecursiveDescentSegment{
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
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: `preferences['theme-settings',"font-size",3]`,
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "preferences",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					UnionSelector: RecursiveDescentSegment{
						{
							Key:   "theme-settings",
							IsKey: true,
						},
						{
							Key:   "font-size",
							IsKey: true,
						},
						{
							Index:   3,
							IsIndex: true,
						},
					},
					ExpectAssociative: true,
					ExpectLinear:      true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "['1st_category']",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "1st_category",
					IsKey:             true,
					ExpectAssociative: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "dimensions[2][3]",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "dimensions",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Index:        2,
					IsIndex:      true,
					ExpectLinear: true,
				},
				{
					Index:        3,
					IsIndex:      true,
					ExpectLinear: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "products['item-details..']",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "products",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Key:               "item-details..",
					IsKey:             true,
					ExpectAssociative: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "$['user info']['address.wind'][1]",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "$",
					IsKeyRoot:         true,
					ExpectLinear:      true,
					ExpectAssociative: true,
				},
				{
					Key:               "user info",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Key:               "address.wind",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Index:        1,
					IsIndex:      true,
					ExpectLinear: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "data[*]",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "data",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Key:               "*",
					IsKeyIndexAll:     true,
					ExpectAssociative: true,
					ExpectLinear:      true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "bicycle['item-code']",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "bicycle",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Key:               "item-code",
					IsKey:             true,
					ExpectAssociative: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "book[0]",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "book",
					IsKey:             true,
					ExpectAssociative: true,
				},
				{
					Index:        0,
					IsIndex:      true,
					ExpectLinear: true,
				},
			},
		},
	) {
		return
	}

	if !yield(
		&ExtractCollectionMemberSegmentsData{
			Segment: "$",
			ExpectedSegments: RecursiveDescentSegment{
				{
					Key:               "$",
					IsKeyRoot:         true,
					ExpectLinear:      true,
					ExpectAssociative: true,
				},
			},
		},
	) {
		return
	}
}
