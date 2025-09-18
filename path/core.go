package path

import (
	"fmt"
	"regexp"
)

// JSONPath alias for string intended to represent a JSON path.
type JSONPath string

// RecursiveDescentSegments alias that represents the final deconstructed JSONPath string using JSONPath.Parse.
type RecursiveDescentSegments []RecursiveDescentSegment

// RecursiveDescentSegment alias that represents a sequence of recursive descent segments.
type RecursiveDescentSegment []*CollectionMemberSegment

// LinearCollectionSelector For Path linear collections (slices and arrays) selector in JSON Path like this: [start:end:step]
type LinearCollectionSelector struct {
	Start   int
	IsStart bool
	End     int
	IsEnd   bool
	Step    int
	IsStep  bool
}

// CollectionMemberSegment For final individual path segment in JSONPath.
type CollectionMemberSegment struct {
	Key                      string
	Index                    int
	IsKey                    bool
	IsKeyIndexAll            bool
	IsKeyRoot                bool
	IsIndex                  bool
	ExpectLinear             bool
	ExpectAssociative        bool
	LinearCollectionSelector *LinearCollectionSelector
	UnionSelector            []*CollectionMemberSegment
}

const (
	JsonpathKeyIndexAll  string = "*"
	JsonpathKeyRoot      string = "$"
	JsonpathDotNotation  string = "."
	JsonpathLeftBracket  string = "["
	JsonpathRightBracket string = "]"
)

func getJsonKey(value string) string {
	if jsonKeyBeginDoesNotNeedBracketsRegex().MatchString(value) {
		if !jsonKeyRemainingNeedBracketsRegex().MatchString(value) {
			return value
		}
	}
	return fmt.Sprintf("['%s']", value)
}

func jsonKeyBeginDoesNotNeedBracketsRegex() *regexp.Regexp {
	return regexp.MustCompile("^[a-zA-Z]")
}

func jsonKeyRemainingNeedBracketsRegex() *regexp.Regexp {
	return regexp.MustCompile("[^a-zA-Z0-9_]")
}

func unionMemberPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`(\d+)|["']([^"']+)["']`)
}

func arraySelectorPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`(\d*):(\d*):(\d*)`)
}

func collectionMemberSegmentPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`\[(\d+|\*)]|\[(\d*:\d*:\d*)]|\[["']([^"']+)["']]|\[((?:[^,\n]+,?)+)]|([a-zA-Z0-9$*_]+)`)
}

func recursiveDescentPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`\[(?:["'][^"']+["']|[^]])+]|([.]{2})`)
}

func memberDotNotationPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`\[(?:["'][^"']+["']|[^]])+]|([.])`)
}
