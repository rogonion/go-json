package object

import (
	"reflect"

	"github.com/rogonion/go-json/internal"
)

// AreEqual Recursively checks if left and right are equal
//
// Actively checks if elements of  slices, arrays, maps, and/or structs are equal and defaults to reflect.DeepEqual for the remaining checks.
//
//	May only panic if reflect functions panic though measures have been set to ensure they are called appropriately.
//
// Parameters:
//   - left - Value to check.
//   - right - Value to check.
//
// Returns true if left and right are equal.
func AreEqual(left reflect.Value, right reflect.Value) bool {
	leftNilOrInvalid := internal.IsNilOrInvalid(left)
	rightNilOrInvalid := internal.IsNilOrInvalid(right)

	if (!leftNilOrInvalid && rightNilOrInvalid) || (!rightNilOrInvalid && leftNilOrInvalid) {
		return false
	}

	if leftNilOrInvalid {
		return true
	}

	if left.Kind() != right.Kind() {
		return false
	}

	switch left.Kind() {
	case reflect.Ptr, reflect.Interface:
		return AreEqual(left.Elem(), right.Elem())
	case reflect.Slice, reflect.Array:
		if left.Len() != right.Len() {
			return false
		}

		for i := 0; i < left.Len(); i++ {
			if !AreEqual(left.Index(i), right.Index(i)) {
				return false
			}
		}
	case reflect.Map:
		leftMapKeys := left.MapKeys()
		rightMapKeys := right.MapKeys()

		if len(leftMapKeys) != len(rightMapKeys) {
			return false
		}

		for _, leftKey := range leftMapKeys {
			leftKeyMatchRightKey := false
			for _, rightKey := range rightMapKeys {
				if AreEqual(leftKey, rightKey) {
					leftKeyMatchRightKey = true
					if !AreEqual(left.MapIndex(leftKey), right.MapIndex(rightKey)) {
						return false
					}
					break
				}
			}
			if !leftKeyMatchRightKey {
				return false
			}
		}
	case reflect.Struct:
		leftNumFields := left.NumField()
		rightNumFields := right.NumField()

		if leftNumFields != rightNumFields {
			return false
		}
		for i := 0; i < leftNumFields; i++ {
			if !AreEqual(left.Field(i), right.Field(i)) {
				return false
			}
		}
	default:
		return reflect.DeepEqual(left.Interface(), right.Interface())
	}

	return true
}
