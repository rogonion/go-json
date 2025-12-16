package path

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
)

func TestPath_SplitPathByRecursiveDescentPattern(t *testing.T) {
	for testData := range SplitPathByRecursiveDescentPatternTestData {
		result := testData.Path.SplitPathByRecursiveDescentPattern()

		if !reflect.DeepEqual(result, testData.ExpectedSegments) {
			t.Error(
				testData.TestTitle, "\n",
				"expected=", core.JsonStringifyMust(testData.ExpectedSegments), "\n",
				"got=", core.JsonStringifyMust(result),
			)
		}
	}
}

type SplitPathByRecursiveDescentPatternData struct {
	internal.TestData
	Path             JSONPath
	ExpectedSegments []JSONPath
}

func SplitPathByRecursiveDescentPatternTestData(yield func(data *SplitPathByRecursiveDescentPatternData) bool) {
	testCaseIndex := 1
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$.store.book[0].title",
			ExpectedSegments: []JSONPath{"$.store.book[0].title"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$..author",
			ExpectedSegments: []JSONPath{"$", "author"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$.store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$.store.bicycle['item-code']"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$.data[*].price",
			ExpectedSegments: []JSONPath{"$.data[*].price"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$['user info']['address.wind'][1].street",
			ExpectedSegments: []JSONPath{"$['user info']['address.wind'][1].street"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$.products['item-details..'].dimensions[0].width.dimensions[2][3].width",
			ExpectedSegments: []JSONPath{"$.products['item-details..'].dimensions[0].width.dimensions[2][3].width"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$..['1st_category'].name",
			ExpectedSegments: []JSONPath{"$", "['1st_category'].name"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             `$.data.user.preferences['theme-settings',"font-size",3]`,
			ExpectedSegments: []JSONPath{`$.data.user.preferences['theme-settings',"font-size",3]`},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$.transactions[1:5:2].amount",
			ExpectedSegments: []JSONPath{"$.transactions[1:5:2].amount"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$['report-data']..['total.sum']..store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$['report-data']", "['total.sum']", "store.bicycle['item-code']"},
		},
	) {
		return
	}

	testCaseIndex++
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			TestData: internal.TestData{
				TestTitle: fmt.Sprintf("Test Case %d", testCaseIndex),
			},
			Path:             "$['report-data']..['total.sum..()&^']..store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$['report-data']", "['total.sum..()&^']", "store.bicycle['item-code']"},
		},
	) {
		return
	}
}
