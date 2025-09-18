package internal

import (
	"encoding/json"
	"reflect"
	"unicode"
)

// StartsWithCapital Checks if a string starts with a capital letter.
func StartsWithCapital(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstChar := rune(s[0])
	return unicode.IsUpper(firstChar)
}

// IsStructFieldExported Checks if field is exported.
func IsStructFieldExported(field reflect.StructField) bool {
	// A simpler, more common way to check is to look at the first character.
	// We check if the name is not empty to avoid a panic.
	if len(field.Name) > 0 {
		return unicode.IsUpper(rune(field.Name[0]))
	}
	return false
}

// IsNilOrInvalid Checks if v is valid or nil depending on the type.
func IsNilOrInvalid(v reflect.Value) bool {
	// 1. Check if the value is valid. This is always safe.
	if !v.IsValid() {
		return true
	}

	// 2. Check if the kind of the value is one of the "nil-able" types.
	// This is required to call IsNil() safely.
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		// 3. For all other types (like struct, int, string), they can never be nil.
		return false
	}
}

// JsonStringifyMust Attempt to stringify value.
//
// Arguments:
//   - value - Must be a pointer to data.
//
// Returns json representation of data if successful else value as is.
func JsonStringifyMust(value any) any {
	if reflect.ValueOf(value).Kind() == reflect.Ptr {
		if jsonData, err := json.Marshal(value); err == nil {
			return string(jsonData)
		}
	} else {
		if jsonData, err := json.Marshal(Ptr(value)); err == nil {
			return string(jsonData)
		}
	}

	return value
}

// Ptr Returns pointer to data.
//
// Useful for functions that require pointers to data.
func Ptr[T any](v T) *T {
	return &v
}
