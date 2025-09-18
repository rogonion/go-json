package path

import (
	"fmt"
	"strings"
)

// Parse Extracts path member segment from path that adheres to JSON Path syntax
//
// Ensure that the path is a valid JSON Path.
//
// Splitting is done as follows:
//   - Breakdown path using recursive descent pattern.
//   - For each recursive descent pattern, breakdown using dot notation pattern followed by bracket notation pattern.
//
// Parameters:
//   - path - Path to data.
//
// Returns:
//
//   - Slice of Recursive Descent Path(s) to data.
//
//     Top level slices represents recursive descent path.
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

func (n RecursiveDescentSegments) String() string {
	segmentsStr := make([]string, 0)
	for _, segment := range n {
		segmentStr := segment.String()
		if segmentStr != "" {
			segmentsStr = append(segmentsStr, segmentStr)
		}
	}
	if len(segmentsStr) > 0 {
		return strings.Join(segmentsStr, "..")
	}
	return ""
}

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

func (n *LinearCollectionSelector) String() string {
	if n == nil {
		return ""
	}

	str := "["
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
	str += "]"
	return str
}

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
		return fmt.Sprintf("[%s]", JsonpathKeyIndexAll)
	}

	if n.IsKeyRoot {
		return fmt.Sprintf("%s", JsonpathKeyRoot)
	}

	if n.IsIndex {
		return fmt.Sprintf("[%d]", n.Index)
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
			return fmt.Sprintf("[%s]", strings.Join(segmentsStr, ","))
		}
	}

	return ""
}
