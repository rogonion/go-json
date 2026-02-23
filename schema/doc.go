/*
Package schema provides tools for defining, validating, converting, and deserializing data structures.

It allows you to define a "Schema" (using DynamicSchema or DynamicSchemaNode) that describes the expected structure and types of your data. This is particularly useful for working with dynamic or semi-structured data where compile-time types might be `any` or `map[string]any`, but a specific structure is enforced at runtime.

Key features:
  - **Definition**: Define schemas for primitives, structs, maps, slices, and arrays.
  - **Validation**: Check if a piece of data matches a defined schema.
  - **Conversion**: Convert raw data (e.g., `map[string]any` from JSON) into strongly-typed Go structures based on the schema.
  - **Deserialization**: Specific helpers for JSON and YAML that combine parsing and conversion.
  - **Path Traversal**: Retrieve the schema definition for a specific node within a larger schema using JSONPath.

# Core Concepts

- **DynamicSchemaNode**: Represents a single node in the schema tree (e.g., a field, a map value, an array element).
- **DynamicSchema**: Represents a collection of possible schemas (often used for root objects or polymorphic types).

# Usage

## Conversion

To convert data (e.g. a map) into a struct based on a schema:

1. Create a new instance of the `Conversion` struct using `NewConversion`.
2. Call the `Convert` method.

You can register custom converters for specific types (like UUIDs) using `WithCustomConverters`.

Example:

	schema := &DynamicSchemaNode{
		// Define a map[int]int schema
		Kind: reflect.Map,
		Type: reflect.TypeOf(map[int]int{}),
		ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
			Kind: reflect.Int,
			Type: reflect.TypeOf(0),
		},
		ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
			Kind: reflect.Int,
			Type: reflect.TypeOf(0),
		},
	}

	source := map[string]string{
		"1": "1",
		"2": "2",
		"3": "3",
	}
	var destination any
	converter := NewConversion()
	err := converter.Convert(source, schema, &destination)

## Deserialization

The module will first parse the raw data (JSON/YAML) into a generic Go structure (map/slice/any) and then convert it to the destination type using the provided schema.

1. Create a new instance of `Deserialization` using `NewDeserialization`.
2. Call `FromJSON` or `FromYAML`.

Example:

	deserializer := NewDeserialization()

	schema := &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(UserProfile2{}),
		ChildNodes: map[string]Schema{
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Age": &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
			"Country": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Occupation": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
		},
	}

	json := "{\"Name\":\"John Doe\"}"
	var jsonDestination UserProfile2
	err := deserializer.FromJSON([]byte(json), schema, &jsonDestination)

	yaml := strings.TrimSpace(`Name: John Doe`)
	var yamlDestination UserProfile2
	err := deserializer.FromYAML([]byte(yaml), schema, &yamlDestination)

## Validation

To check if data adheres to a schema without converting it:

Example:

	schema := &DynamicSchemaNode{
		Kind: reflect.String,
		Type: reflect.TypeOf(""),
	}

	validation := NewValidation()
	ok, err := validation.ValidateData("this is a string", schema)
*/
package schema
