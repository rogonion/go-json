/*
Package object provides a set of methods for manipulating dynamic objects using JSONPath.

It is designed to work with data structures where the type might not be known at compile time (e.g., `any`, `map[string]any`) or deeply nested structures (structs, slices, maps).

Key features:
  - **Get**: Retrieve values from an object using a JSONPath query.
  - **Set**: Create or update values at a specific JSONPath. Supports auto-creation of nested structures if a Schema is provided.
  - **Delete**: Remove values at a specific JSONPath.
  - **ForEach**: Iterate over all values matching a JSONPath query.
  - **AreEqual**: Deep equality check with support for custom equality handlers.

# Core Concepts

- **Object**: The main entry point. It wraps the source data and provides methods to manipulate it.

# Usage

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

	objManip := NewObject().WithSourceInterface(source)

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
