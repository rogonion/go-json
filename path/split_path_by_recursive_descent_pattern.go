package path

func (jsonPath JSONPath) SplitPathByRecursiveDescentPattern() []JSONPath {
	matches := recursiveDescentPatternRegex().FindAllStringSubmatchIndex(string(jsonPath), -1)

	recursiveDescentPaths := make([]JSONPath, 0)
	recursiveDescentIndexes := make([][2]int, 0)
	for _, match := range matches {
		// The capturing group's indices are at positions 2 and 3
		if match[2] != -1 {
			recursiveDescentIndexes = append(recursiveDescentIndexes, [2]int{match[2], match[3]})
		}
	}

	if len(recursiveDescentIndexes) > 0 {
		start := 0
		for _, recursiveDescentIndex := range recursiveDescentIndexes {
			recursiveDescentPaths = append(recursiveDescentPaths, jsonPath[start:recursiveDescentIndex[0]])
			start = recursiveDescentIndex[1]
		}

		if start != len(jsonPath) {
			recursiveDescentPaths = append(recursiveDescentPaths, jsonPath[start:])
		}
	} else {
		recursiveDescentPaths = append(recursiveDescentPaths, jsonPath)
	}

	return recursiveDescentPaths
}
