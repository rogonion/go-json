package object

import (
	"reflect"

	"github.com/rogonion/go-json/core"
)

/*
Equal Define custom equal check logic.

Meant to be implemented by custom data types that need to perform specific value-based equality checks beyond defaults.
*/
type Equal interface {
	// AreEqual Checks if left and right are equal
	//
	// Parameters:
	//   - left - Value to check.
	//   - right - Value to check.
	//
	// Returns true if left and right are equal.
	AreEqual(left any, right any) bool

	AreEqualReflect(left reflect.Value, right reflect.Value) bool
}

/*
AreEquals Map of custom equal checkers.

Intended to be used for custom equality check logic of user-defined types like structs.
*/
type AreEquals map[reflect.Type]Equal

/*
AreEqual Recursively checks if left and right are equal

Actively checks if elements of  slices, arrays, maps, and/or structs are equal and defaults to reflect.DeepEqual for the remaining checks.

May only panic if reflect functions panic though measures have been set to ensure they are called appropriately.

Parameters:
  - left - Value to check.
  - right - Value to check.

Returns `true` if left and right are equal.
*/
func (n *AreEqual) AreEqual(left any, right any) bool {
	return n.AreEqualReflect(reflect.ValueOf(left), reflect.ValueOf(right))
}

func (n *AreEqual) AreEqualReflect(left reflect.Value, right reflect.Value) bool {
	leftNilOrInvalid := core.IsNilOrInvalid(left)
	rightNilOrInvalid := core.IsNilOrInvalid(right)

	if (!leftNilOrInvalid && rightNilOrInvalid) || (!rightNilOrInvalid && leftNilOrInvalid) {
		return false
	}

	if leftNilOrInvalid {
		return true
	}

	if left.Kind() != right.Kind() {
		return false
	}

	if customEqualityCheck, ok := n.customEquals[left.Type()]; ok {
		return customEqualityCheck.AreEqualReflect(left, right)
	}

	switch left.Kind() {
	case reflect.Ptr, reflect.Interface:
		return n.AreEqualReflect(left.Elem(), right.Elem())
	case reflect.Slice, reflect.Array:
		if left.Len() != right.Len() {
			return false
		}

		for i := 0; i < left.Len(); i++ {
			if !n.AreEqualReflect(left.Index(i), right.Index(i)) {
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
				if n.AreEqualReflect(leftKey, rightKey) {
					leftKeyMatchRightKey = true
					if !n.AreEqualReflect(left.MapIndex(leftKey), right.MapIndex(rightKey)) {
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
			if !n.AreEqualReflect(left.Field(i), right.Field(i)) {
				return false
			}
		}
	default:
		return reflect.DeepEqual(left.Interface(), right.Interface())
	}

	return true
}

func (n *AreEqual) WithCustomEquals(value AreEquals) *AreEqual {
	n.customEquals = value
	return n
}

func (n *AreEqual) SetCustomEquals(value AreEquals) {
	n.customEquals = value
}

func NewAreEqual() *AreEqual {
	n := new(AreEqual)
	return n
}

/*
AreEqual checks if two values are equal.
*/
type AreEqual struct {
	// Pass custom equality check logic.
	//
	// Useful for user defined types like structs.
	customEquals AreEquals
}
