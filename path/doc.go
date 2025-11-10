/*
Package path provides the foundation for setting up JSON Path in go.

It provides a custom type called JSONPath (alias for string) which offers the following core set of methods:
  - JSONPath.Parse - Converts JSONPath `string` into a 2D slice of CollectionMemberSegment which provides detailed information about each path segment.

# Usage

Example parsing JSONPath string:

	var jsonPath JSONPath = "$[1,3,5]"
	var parsedPath RecursiveDescentSegments = jsonPath.Parse()
*/
package path
