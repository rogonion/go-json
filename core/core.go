package core

import (
	"encoding/json"
	"fmt"
)

/*
JsonObject represents a generic JSON object structure.

This would be a map where keys are strings and values are of any type in Go.
*/
type JsonObject map[string]any

func (f JsonObject) String() string {
	if j, err := json.Marshal(f); err != nil {
		return fmt.Sprintf("<Marshalling error: %v>", err)
	} else {
		return string(j)
	}
}

/*
JsonArray represents a generic JSON array structure.

This would be a slice whose elements are of type any in Go.
*/
type JsonArray []any

func (f JsonArray) String() string {
	if j, err := json.Marshal(f); err != nil {
		return fmt.Sprintf("<Marshalling error: %v>", err)
	} else {
		return string(j)
	}
}
