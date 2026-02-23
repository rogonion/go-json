package path

/*
SplitPathSegmentByDotNotationPattern splits a path segment into smaller segments using the dot ('.') delimiter.

This function is typically called after splitting by recursive descent. It respects brackets and quotes,
ensuring that dots inside string literals or bracket notation are not treated as delimiters.
*/
func (jsonPath JSONPath) SplitPathSegmentByDotNotationPattern() []JSONPath {
	dotNotationPaths := make([]JSONPath, 0)

	matches := memberDotNotationPatternRegex().FindAllStringSubmatchIndex(string(jsonPath), -1)

	memberDotNotationIndexes := make([][2]int, 0)
	for _, match := range matches {
		// The capturing group's indices are at positions 2 and 3
		if match[2] != -1 {
			memberDotNotationIndexes = append(memberDotNotationIndexes, [2]int{match[2], match[3]})
		}
	}

	if len(memberDotNotationIndexes) > 0 {
		start := 0
		for _, memberDotNotationIndex := range memberDotNotationIndexes {
			dotNotationPaths = append(dotNotationPaths, jsonPath[start:memberDotNotationIndex[0]])
			start = memberDotNotationIndex[1]
		}

		if start != len(jsonPath) {
			dotNotationPaths = append(dotNotationPaths, jsonPath[start:])
		}
	} else {
		dotNotationPaths = append(dotNotationPaths, jsonPath)
	}

	return dotNotationPaths
}
