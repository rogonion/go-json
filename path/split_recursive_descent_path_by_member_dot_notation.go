package path

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
