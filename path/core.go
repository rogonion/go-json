package path

import (
	"fmt"
	"regexp"
)

/*
JSONPath is an alias for a string intended to represent a JSON path query.
Example: "$.store.book[0].title"
*/
type JSONPath string

/*
RecursiveDescentSegments represents the final deconstructed JSONPath string after parsing.
It is a 2D slice where the first dimension separates segments by the recursive descent operator ('..').
*/
type RecursiveDescentSegments []RecursiveDescentSegment

/*
RecursiveDescentSegment represents a sequence of path segments that are directly connected
(e.g. by dot notation or brackets) without any recursive descent operators.
*/
type RecursiveDescentSegment []*CollectionMemberSegment

/*
LinearCollectionSelector represents a slice selector for linear collections (arrays/slices).
Syntax: [start:end:step]
*/
type LinearCollectionSelector struct {
	Start   int  // The starting index (inclusive)
	IsStart bool // True if Start was specified
	End     int  // The ending index (exclusive)
	IsEnd   bool // True if End was specified
	Step    int  // The step value
	IsStep  bool // True if Step was specified
}

/*
CollectionMemberSegment represents a single atomic segment in a JSONPath.
It holds information about the key, index, or selector used at that specific point in the path.
*/
type CollectionMemberSegment struct {
	// Key represents the map key or property name.
	Key   string
	IsKey bool
	// IsKeyIndexAll is true if the segment is a wildcard '*'.
	IsKeyIndexAll bool
	// IsKeyRoot is true if the segment is the root '$'.
	IsKeyRoot                bool
	Index                    int
	IsIndex                  bool
	ExpectLinear             bool
	ExpectAssociative        bool
	LinearCollectionSelector *LinearCollectionSelector
	UnionSelector            []*CollectionMemberSegment
}

const (
	JsonpathKeyIndexAll              string = "*"
	JsonpathKeyRoot                  string = "$"
	JsonpathDotNotation              string = "."
	JsonpathRecursiveDescentNotation string = ".."
	JsonpathLeftBracket              string = "["
	JsonpathRightBracket             string = "]"
)

// getJsonKey formats a string key for JSONPath output.
// It returns the key as-is if it's a valid identifier, otherwise wraps it in brackets and quotes.
// E.g., "name" -> "name", "first-name" -> "['first-name']"
func getJsonKey(value string) string {
	if jsonKeyBeginDoesNotNeedBracketsRegex().MatchString(value) {
		if !jsonKeyRemainingNeedBracketsRegex().MatchString(value) {
			return value
		}
	}
	return fmt.Sprintf("['%s']", value)
}

// jsonKeyBeginDoesNotNeedBracketsRegex matches keys starting with a letter.
func jsonKeyBeginDoesNotNeedBracketsRegex() *regexp.Regexp {
	return regexp.MustCompile("^[a-zA-Z]")
}

// jsonKeyRemainingNeedBracketsRegex matches keys containing characters that require brackets.
func jsonKeyRemainingNeedBracketsRegex() *regexp.Regexp {
	return regexp.MustCompile("[^a-zA-Z0-9_]")
}

// unionMemberPatternRegex matches individual members inside a union selector (integers or quoted strings).
func unionMemberPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`(\d+)|["']([^"']+)["']`)
}

// arraySelectorPatternRegex matches the array slice syntax start:end:step.
func arraySelectorPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`(\d*):(\d*):(\d*)`)
}

// collectionMemberSegmentPatternRegex matches various forms of path segments (index, slice, quoted key, union, simple key).
func collectionMemberSegmentPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`\[(\d+|\*)]|\[(\d*:\d*:\d*)]|\[["']([^"']+)["']]|\[((?:[^,\n]+,?)+)]|([a-zA-Z0-9$*_]+)`)
}

// recursiveDescentPatternRegex matches segments separated by '..'.
func recursiveDescentPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`\[(?:["'][^"']+["']|[^]])+]|([.]{2})`)
}

func memberDotNotationPatternRegex() *regexp.Regexp {
	return regexp.MustCompile(`\[(?:["'][^"']+["']|[^]])+]|([.])`)
}
