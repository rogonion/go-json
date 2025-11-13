/*
Package object provides a set of methods for manipulating objects (data that can be deeply nested whose structure can be determined at runtime) using path.JSONPath.

An object can be a simple primitive or a complex nested mixed structure of structs, maps, slices, and arrays.

Provides the following methods for manipulating a single object as a source based on path.JSONPath:
 1. Object.Get - Retrieve value(s) from an object.
 2. Object.Set - Create or update value(s) in an object.
 3. Object.Delete - Delete value(s) in an object.
 4. Object.ForEach - Loop through each value in an object.

Optionally, there is the AreEqual package that performs the same function as reflect.DeepEqual with the addition of setting up custom AreEquals handlers.

# Usage

Except AreEqual, the remaining methods work on a single object as a source. This means you can chain manipulation method calls against one a single source object after instantiation.

To begin:

1. Create an instance of the Object struct. You can use the NewObject function which provides convenience for setting the `Object.source` and the default value for `Object.defaultConverter`.

The following parameters can be set using the builder method (prefixed `With`) or Set (prefixed `Set) before calling the manipulation methods:
  - Object.source - Mandatory. This is the root object to work with.
  - Object.defaultConverter - Mandatory for Object.Set. Module to use when converting `Object.valueToSet` to the destination type at path.JSONPath. When Object is instantiated using NewObject, it will be set with schema.NewConversion.
  - Object.schema - Optional but useful for Object.Set. Used to determine the new collections (especially structs) to create in the nested object. Defaults to json collections (core.JsonObject and core.JsonArray).

2. Once the Object has been successfully initialized, you can begin calling the manipulation methods: Object.Get, Object.Set, Object.Delete, and `Object.ForEach, which will work on the same `Object.source`.

3. Once you are satisfied, you can call the `Object.GetSourceInterface` method to retrieve the modified source especially if changed using `Object.Set` or `Object.Delete`.

Example:

	type Address struct {
		Street  string
		City    string
		ZipCode *string
	}

	source := map[string]any{
		"data": map[string]any{
			"metadata": struct {
				Address Address
				Status  string
			}{
				Address: Address{
					Street: "123 Main St",
					City:   "Anytown",
				},
				Status: "active",
			},
		},
	}

	objManip := NewObject().WithSource(source)

	noOfResults, err := objManip.Get("$.data.metadata.Address.City")
	var valueFound any
	if noOfResults > 0 {
		// retrieve value found if no of results is greater than 0
		valueFound = objManip.GetValueFoundInterface()
	}

	noOfModifications, err := objManip.Set("$.data.metadata.Status", "inactive")

	noOfModifications, err = objManip.Delete("$.data.metadata.Status")

	// retrieve modified source after Set/Delete
	var modifiedSource any = objManip.GetSourceInterface()
*/
package object
