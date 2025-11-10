package core

import "reflect"

/*
IsArray Checks if an obj is an array.

Returns:
  - array element type.
  - bool to indicate if it is an array.
*/
func IsArray(obj any) (reflect.Type, bool) {
	objType := reflect.TypeOf(obj)

	if objType.Kind() != reflect.Array {
		return nil, false
	}

	return objType.Elem(), true
}

/*
IsSlice Checks if an objectV1 is a slice.

Returns:
  - slice element type.
  - bool to indicate if it is a slice.
*/
func IsSlice(obj any) (reflect.Type, bool) {
	objType := reflect.TypeOf(obj)

	if objType.Kind() != reflect.Slice {
		return nil, false
	}

	return objType.Elem(), true
}

/*
IsMap Checks if an objectV1 is a map.

Returns:
  - map key type.
  - map value type.
  - bool to indicate if it is a map.
*/
func IsMap(obj any) (reflect.Type, reflect.Type, bool) {
	objType := reflect.TypeOf(obj)

	if objType.Kind() != reflect.Map {
		return nil, nil, false
	}

	return objType.Key(), objType.Elem(), true
}

/*
GetArraySliceValueType Extracts Array/Slice Element Value Type.

Returns:
  - Array/Slice Element Value Type or nil if value is not a slice/array.
  - True if value is a slice/array.
*/
func GetArraySliceValueType(value reflect.Value) (reflect.Type, bool) {
	if value.Kind() != reflect.Slice && value.Type().Kind() != reflect.Array {
		return nil, false
	}

	return value.Type().Elem(), true
}

/*
GetMapKeyValueType Extracts Map Key Type, and Map Value Type.

Returns:
  - Map Key Type or nil if value is not a map.
  - Map Value Type or nil if value is not a map.
  - True if value is a map.
*/
func GetMapKeyValueType(value reflect.Value) (reflect.Type, reflect.Type, bool) {
	if value.Kind() != reflect.Map {
		return nil, nil, false
	}

	return value.Type().Key(), value.Type().Elem(), true
}
