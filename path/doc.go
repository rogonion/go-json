/*
Package path provides the foundation for parsing and working with JSONPath in Go.

It implements a parser that converts a JSONPath string into a structured representation
(RecursiveDescentSegments) which can be used to traverse, query, and manipulate data structures.

The package supports a subset of the JSONPath specification, including:
  - Root identifier (`$`)
  - Dot notation (`.key`)
  - Bracket notation (`['key']`, `["key"]`)
  - Recursive descent (`..`)
  - Wildcards (`*`)
  - Array/Slice selectors (`[start:end:step]`)
  - Union selectors (`['key1','key2']`, `[1,3,5]`)
  - Index selectors (`[0]`, `[1]`)

# Usage

To parse a JSONPath string:

	var jsonPath JSONPath = "$[1,3,5]"
	var parsedPath RecursiveDescentSegments = jsonPath.Parse()
*/
package path
