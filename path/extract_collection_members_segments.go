package path

import (
	"strconv"
)

/*
ExtractCollectionMemberSegments parses a single JSONPath segment string into a structured RecursiveDescentSegment.

It identifies and extracts various selector types such as indices (`[1]`), wildcards (`[*]`), slices (`[1:5]`), unions (`['a','b']`), and standard keys.
*/
func (jsonPath JSONPath) ExtractCollectionMemberSegments() RecursiveDescentSegment {
	collectionMemberSegments := make(RecursiveDescentSegment, 0)

	matches := collectionMemberSegmentPatternRegex().FindAllStringSubmatch(string(jsonPath), -1)

	for _, match := range matches {
		for j, segment := range match {
			if segment == "" {
				continue
			}

			var collectionMemberSegment *CollectionMemberSegment

			switch j {
			case 1: // E.g. [1] , [*]
				collectionMemberSegment = new(CollectionMemberSegment)
				if index, err := strconv.Atoi(segment); err == nil {
					collectionMemberSegment.Index = index
					collectionMemberSegment.IsIndex = true
					collectionMemberSegment.ExpectLinear = true
				} else {
					collectionMemberSegment.Key = segment
					collectionMemberSegment.IsKeyIndexAll = true
					collectionMemberSegment.ExpectAssociative = true
					collectionMemberSegment.ExpectLinear = true
				}
			case 2: // E.g. [1:5:2] , [1::]
				startEndStepMatch := arraySelectorPatternRegex().FindStringSubmatch(segment)
				if len(startEndStepMatch) == 0 {
					break
				}

				sesAvailable := false
				collectionMemberSegment = new(CollectionMemberSegment)
				for _, ses := range startEndStepMatch {
					if len(ses) > 0 {
						sesAvailable = true
						break
					}
				}
				if !sesAvailable {
					collectionMemberSegment.Key = JsonpathKeyIndexAll
					collectionMemberSegment.IsKeyIndexAll = true
					collectionMemberSegment.ExpectAssociative = true
					collectionMemberSegment.ExpectLinear = true
				} else {
					collectionMemberSegment.LinearCollectionSelector = new(LinearCollectionSelector)
					if index, err := strconv.Atoi(startEndStepMatch[1]); err == nil {
						collectionMemberSegment.LinearCollectionSelector.Start = index
						collectionMemberSegment.LinearCollectionSelector.IsStart = true
					}
					if index, err := strconv.Atoi(startEndStepMatch[2]); err == nil {
						collectionMemberSegment.LinearCollectionSelector.End = index
						collectionMemberSegment.LinearCollectionSelector.IsEnd = true
					}
					if index, err := strconv.Atoi(startEndStepMatch[3]); err == nil {
						collectionMemberSegment.LinearCollectionSelector.Step = index
						collectionMemberSegment.LinearCollectionSelector.IsStep = true
					}
					collectionMemberSegment.ExpectLinear = true
				}
			case 3: // E.g. ['report-data'] , ["reach..records"]
				collectionMemberSegment = new(CollectionMemberSegment)
				collectionMemberSegment.Key = segment
				collectionMemberSegment.IsKey = true
				collectionMemberSegment.ExpectAssociative = true
			case 4: // E.g. ['theme-settings',"font-size",3]
				unionMemberMatch := unionMemberPatternRegex().FindAllStringSubmatch(segment, -1)

				if len(unionMemberMatch) == 0 {
					break
				}

				collectionMemberSegment = new(CollectionMemberSegment)
				collectionMemberSegment.UnionSelector = make(RecursiveDescentSegment, 0)

				for _, umm := range unionMemberMatch {
					for ummSegmentIndex, ummSegment := range umm {
						if ummSegment == "" {
							continue
						}

						switch ummSegmentIndex {
						case 1:
							if index, err := strconv.Atoi(ummSegment); err == nil {
								ummMemberSegment := new(CollectionMemberSegment)
								ummMemberSegment.Index = index
								ummMemberSegment.IsIndex = true
								collectionMemberSegment.UnionSelector = append(collectionMemberSegment.UnionSelector, ummMemberSegment)
							}
						case 2:
							ummMemberSegment := new(CollectionMemberSegment)
							ummMemberSegment.Key = ummSegment
							ummMemberSegment.IsKey = true
							collectionMemberSegment.UnionSelector = append(collectionMemberSegment.UnionSelector, ummMemberSegment)
						}
					}
				}

				collectionMemberSegment.ExpectAssociative = true
				collectionMemberSegment.ExpectLinear = true
			case 5: // E.g. _threeFour5 , *
				collectionMemberSegment = new(CollectionMemberSegment)
				if segment == JsonpathKeyIndexAll {
					collectionMemberSegment.Key = segment
					collectionMemberSegment.IsKeyIndexAll = true
					collectionMemberSegment.ExpectAssociative = true
				} else if segment == JsonpathKeyRoot {
					collectionMemberSegment.Key = segment
					collectionMemberSegment.IsKeyRoot = true
					collectionMemberSegment.ExpectAssociative = true
					collectionMemberSegment.ExpectLinear = true
				} else {
					collectionMemberSegment.Key = segment
					collectionMemberSegment.IsKey = true
					collectionMemberSegment.ExpectAssociative = true
				}
			}

			if collectionMemberSegment != nil {
				collectionMemberSegments = append(collectionMemberSegments, collectionMemberSegment)
			}
		}
	}

	return collectionMemberSegments
}
