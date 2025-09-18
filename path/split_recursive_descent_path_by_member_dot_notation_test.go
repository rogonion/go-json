package path

import (
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
)

func TestPath_SplitPathSegmentByDotNotationPattern(t *testing.T) {
	for testData := range SplitRecursiveDescentPathByMemberDotNotationPatternTestData {
		result := testData.Segment.SplitPathSegmentByDotNotationPattern()

		if !reflect.DeepEqual(result, testData.ExpectedSegments) {
			t.Error(
				"expected=", internal.JsonStringifyMust(testData.ExpectedSegments), "\n",
				"got=", internal.JsonStringifyMust(result),
			)
		}
	}
}

type SplitRecursiveDescentPathByMemberDotNotationPatternData struct {
	Segment          JSONPath
	ExpectedSegments []JSONPath
}

func SplitRecursiveDescentPathByMemberDotNotationPatternTestData(yield func(data *SplitRecursiveDescentPathByMemberDotNotationPatternData) bool) {
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$.store.book[0].title",
			ExpectedSegments: []JSONPath{"$", "store", "book[0]", "title"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$",
			ExpectedSegments: []JSONPath{"$"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "author",
			ExpectedSegments: []JSONPath{"author"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$.store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$", "store", "bicycle['item-code']"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$.data[*].price",
			ExpectedSegments: []JSONPath{"$", "data[*]", "price"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$['user info']['address.wind'][1].street",
			ExpectedSegments: []JSONPath{"$['user info']['address.wind'][1]", "street"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$.products['item-details..'].dimensions[0].width.dimensions[2][3].width",
			ExpectedSegments: []JSONPath{"$", "products['item-details..']", "dimensions[0]", "width", "dimensions[2][3]", "width"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "['1st_category'].name",
			ExpectedSegments: []JSONPath{"['1st_category']", "name"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$.data.user.preferences['theme-settings','font-size',3]",
			ExpectedSegments: []JSONPath{"$", "data", "user", "preferences['theme-settings','font-size',3]"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$.transactions[1:5:2].amount",
			ExpectedSegments: []JSONPath{"$", "transactions[1:5:2]", "amount"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "$['report-data']",
			ExpectedSegments: []JSONPath{"$['report-data']"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "['total.sum']",
			ExpectedSegments: []JSONPath{"['total.sum']"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"store", "bicycle['item-code']"},
		},
	) {
		return
	}

	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			Segment:          "['total.sum..()&^']",
			ExpectedSegments: []JSONPath{"['total.sum..()&^']"},
		},
	) {
		return
	}
}
