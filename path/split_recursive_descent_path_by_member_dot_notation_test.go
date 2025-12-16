package path

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
)

func TestPath_SplitPathSegmentByDotNotationPattern(t *testing.T) {
	for testData := range SplitRecursiveDescentPathByMemberDotNotationPatternTestData {
		result := testData.Segment.SplitPathSegmentByDotNotationPattern()

		if !reflect.DeepEqual(result, testData.ExpectedSegments) {
			t.Error(
				testData.TestTitle, "\n",
				"expected=", core.JsonStringifyMust(testData.ExpectedSegments), "\n",
				"got=", core.JsonStringifyMust(result),
			)
		}
	}
}

type SplitRecursiveDescentPathByMemberDotNotationPatternData struct {
	internal.TestData
	Segment          JSONPath
	ExpectedSegments []JSONPath
}

func SplitRecursiveDescentPathByMemberDotNotationPatternTestData(yield func(data *SplitRecursiveDescentPathByMemberDotNotationPatternData) bool) {
	testCaseIndex := 1
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$.store.book[0].title",
			ExpectedSegments: []JSONPath{"$", "store", "book[0]", "title"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$",
			ExpectedSegments: []JSONPath{"$"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "author",
			ExpectedSegments: []JSONPath{"author"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$.store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$", "store", "bicycle['item-code']"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$.data[*].price",
			ExpectedSegments: []JSONPath{"$", "data[*]", "price"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$['user info']['address.wind'][1].street",
			ExpectedSegments: []JSONPath{"$['user info']['address.wind'][1]", "street"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$.products['item-details..'].dimensions[0].width.dimensions[2][3].width",
			ExpectedSegments: []JSONPath{"$", "products['item-details..']", "dimensions[0]", "width", "dimensions[2][3]", "width"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "['1st_category'].name",
			ExpectedSegments: []JSONPath{"['1st_category']", "name"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$.data.user.preferences['theme-settings','font-size',3]",
			ExpectedSegments: []JSONPath{"$", "data", "user", "preferences['theme-settings','font-size',3]"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$.transactions[1:5:2].amount",
			ExpectedSegments: []JSONPath{"$", "transactions[1:5:2]", "amount"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "$['report-data']",
			ExpectedSegments: []JSONPath{"$['report-data']"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "['total.sum']",
			ExpectedSegments: []JSONPath{"['total.sum']"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"store", "bicycle['item-code']"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitRecursiveDescentPathByMemberDotNotationPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Segment:          "['total.sum..()&^']",
			ExpectedSegments: []JSONPath{"['total.sum..()&^']"},
		},
	) {
		return
	}
}
