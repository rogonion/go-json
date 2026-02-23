package path

import (
	"fmt"
	"strings"
)

/*
Parse breaks down a JSONPath string into a structured 2D slice of segments.

The parsing process involves:
 1. Splitting the path by the recursive descent operator (`..`).
 2. For each resulting section:
    a. Splitting by the dot notation pattern (`.`).
    b. Extracting individual collection members (brackets, indices, keys).

It returns a RecursiveDescentSegments object, which is a 2D slice. The top-level slice
represents parts of the path separated by recursive descent, and the inner slice contains
the linear sequence of segments.

Example:

	path := path.JSONPath("$..book[0]")
	segments := path.Parse()
*/
func (jsonPath JSONPath) Parse() RecursiveDescentSegments {
	recursiveDescentSegments := jsonPath.SplitPathByRecursiveDescentPattern()

	segments := make(RecursiveDescentSegments, 0)
	for _, recursiveDescentSegment := range recursiveDescentSegments {
		splitDotNotationSegments := recursiveDescentSegment.SplitPathSegmentByDotNotationPattern()
		collectionSegments := make([]*CollectionMemberSegment, 0)
		for _, splitDotNotationSegment := range splitDotNotationSegments {
			res := splitDotNotationSegment.ExtractCollectionMemberSegments()
			collectionSegments = append(collectionSegments, res...)
		}
		segments = append(segments, collectionSegments)
	}

	return segments
}

// String reconstructs the JSONPath string from the parsed RecursiveDescentSegments.
func (n RecursiveDescentSegments) String() string {
	segmentsStr := make([]string, 0)
	for _, segment := range n {
		segmentStr := segment.String()
		if segmentStr != "" {
			segmentsStr = append(segmentsStr, segmentStr)
		}
	}
	if len(segmentsStr) > 0 {
		return strings.Join(segmentsStr, JsonpathRecursiveDescentNotation)
	}
	return ""
}

// String reconstructs the string representation of a single recursive descent segment.
func (n RecursiveDescentSegment) String() string {
	segmentsStr := make([]string, 0)
	for _, s := range n {
		sString := s.String()
		if len(sString) > 0 {
			segmentsStr = append(segmentsStr, sString)
		}
	}
	if len(segmentsStr) > 0 {
		newPath := segmentsStr[0]
		if len(segmentsStr) > 1 {
			for i := 1; i < len(segmentsStr); i++ {
				segmentStr := segmentsStr[i]
				if strings.HasPrefix(segmentStr, JsonpathLeftBracket) && strings.HasSuffix(segmentStr, JsonpathRightBracket) {
					newPath += segmentStr
					continue
				}
				if !strings.HasSuffix(newPath, JsonpathDotNotation) {
					newPath += JsonpathDotNotation
				}
				newPath += segmentStr
				if i != len(segmentsStr)-1 {
					nextSegmentStr := segmentsStr[i+1]
					if !strings.HasPrefix(nextSegmentStr, JsonpathLeftBracket) && !strings.HasSuffix(nextSegmentStr, JsonpathRightBracket) {
						newPath += JsonpathDotNotation
					}
				}
			}
		}
		return newPath
	}
	return ""
}

// String returns the string representation of a linear collection selector (e.g., "[1:5:2]").
func (n *LinearCollectionSelector) String() string {
	if n == nil {
		return ""
	}

	str := JsonpathLeftBracket
	if n.IsStart {
		str += fmt.Sprintf("%d", n.Start)
	}
	str += ":"
	if n.IsEnd {
		str += fmt.Sprintf("%d", n.End)
	}
	str += ":"
	if n.IsStep {
		str += fmt.Sprintf("%d", n.Step)
	}
	str += JsonpathRightBracket
	return str
}

// String returns the string representation of a single collection member segment.
func (n *CollectionMemberSegment) String() string {
	if n == nil {
		return ""
	}

	if n.IsKey && n.Key != "" {
		return getJsonKey(n.Key)
	}

	if n.IsKeyIndexAll {
		if n.ExpectAssociative {
			return JsonpathKeyIndexAll
		}
		return fmt.Sprintf("%s%s%s", JsonpathLeftBracket, JsonpathKeyIndexAll, JsonpathRightBracket)
	}

	if n.IsKeyRoot {
		return JsonpathKeyRoot
	}

	if n.IsIndex {
		return fmt.Sprintf("%s%d%s", JsonpathLeftBracket, n.Index, JsonpathRightBracket)
	}

	if n.LinearCollectionSelector != nil {
		return n.LinearCollectionSelector.String()
	}

	if len(n.UnionSelector) > 0 {
		segmentsStr := make([]string, 0)
		for _, u := range n.UnionSelector {
			uStr := u.String()
			if uStr != "" {
				if strings.HasPrefix(uStr, JsonpathLeftBracket) {
					uStr = uStr[1:]
				}
				if strings.HasSuffix(uStr, JsonpathRightBracket) {
					uStr = uStr[:len(uStr)-1]
				}
				segmentsStr = append(segmentsStr, uStr)
			}
		}
		if len(segmentsStr) > 0 {
			return fmt.Sprintf("%s%s%s", JsonpathLeftBracket, strings.Join(segmentsStr, ","), JsonpathRightBracket)
		}
	}

	return ""
}
