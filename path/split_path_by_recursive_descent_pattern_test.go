package path

import (
	"reflect"
	"testing"

	"github.com/rogonion/go-json/internal"
)

func TestPath_SplitPathByRecursiveDescentPattern(t *testing.T) {
	for testData := range SplitPathByRecursiveDescentPatternTestData {
		result := testData.Path.SplitPathByRecursiveDescentPattern()

		if !reflect.DeepEqual(result, testData.ExpectedSegments) {
			t.Error(
				"expected=", internal.JsonStringifyMust(testData.ExpectedSegments), "\n",
				"got=", internal.JsonStringifyMust(result),
			)
		}
	}
}

type SplitPathByRecursiveDescentPatternData struct {
	Path             JSONPath
	ExpectedSegments []JSONPath
}

func SplitPathByRecursiveDescentPatternTestData(yield func(data *SplitPathByRecursiveDescentPatternData) bool) {
	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$.store.book[0].title",
			ExpectedSegments: []JSONPath{"$.store.book[0].title"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$..author",
			ExpectedSegments: []JSONPath{"$", "author"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$.store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$.store.bicycle['item-code']"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$.data[*].price",
			ExpectedSegments: []JSONPath{"$.data[*].price"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$['user info']['address.wind'][1].street",
			ExpectedSegments: []JSONPath{"$['user info']['address.wind'][1].street"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$.products['item-details..'].dimensions[0].width.dimensions[2][3].width",
			ExpectedSegments: []JSONPath{"$.products['item-details..'].dimensions[0].width.dimensions[2][3].width"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$..['1st_category'].name",
			ExpectedSegments: []JSONPath{"$", "['1st_category'].name"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             `$.data.user.preferences['theme-settings',"font-size",3]`,
			ExpectedSegments: []JSONPath{`$.data.user.preferences['theme-settings',"font-size",3]`},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$.transactions[1:5:2].amount",
			ExpectedSegments: []JSONPath{"$.transactions[1:5:2].amount"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$['report-data']..['total.sum']..store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$['report-data']", "['total.sum']", "store.bicycle['item-code']"},
		},
	) {
		return
	}

	if !yield(
		&SplitPathByRecursiveDescentPatternData{
			Path:             "$['report-data']..['total.sum..()&^']..store.bicycle['item-code']",
			ExpectedSegments: []JSONPath{"$['report-data']", "['total.sum..()&^']", "store.bicycle['item-code']"},
		},
	) {
		return
	}
}
