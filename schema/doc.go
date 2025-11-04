/*
Package schema is a module for bringing structure to an object.

Offers the following set of functions:
  - Conversion - Convert data from one type to another using Schema. Works with simple and nested objects as well as offers the option to supply custom Converter.
  - Deserialization - Deserialize data from sources such as `json` or `yaml` using schema.
  - Validation - Validate data using Schema. Offers option to supply custom Validator.

# Usage

## Conversion

To begin using the module:

1. Create a new instance of the Conversion struct. You can use the convenience method NewConversion.

	The following parameters can be set using the builder method (prefixed `With`) or Set (prefixed `Set):

		- customConverters - A map of custom converters. Useful especially for user-defined structs e.g. `uuid`. Can be set using `Conversion.WithCustomConverters` or `Conversion.SetCustomConverters`. The conversion logic will prioritize custom converters for any types encountered.

2. Call the Conversion.Convert method to convert data from one type to another.

Example:

	schema := &DynamicSchemaNode{
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

To begin using the module:

1. Create a new instance of the Deserialization struct. You can use the convenience method NewDeserialization which sets the `Deserialization.defaultConverter` using NewConversion.

	The following parameters can be set using the builder method (prefixed `With`) or Set (prefixed `Set):

		- defaultConverter - Can be set using `Deserialization.WithDefaultConverter` or `Deserialization.SetDefaultConverter`.

		- customConverters - Converter used immediately after deserialization if deserialized type matches. Can be set using `Deserialization.WithCustomConverters` or `Deserialization.SetCustomConverters`.

2. Deserialize the data using the following methods:
  - Deserialization.FromJSON
  - Deserialization.FromYAML

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

	json := "{"Name":"John Doe"}"
	var jsonDestination UserProfile2
	err := deserializer.FromJSON([]byte(json), schema, &jsonDestination)

	yaml := strings.TrimSpace(`Name: John Doe`)
	var yamlDestination UserProfile2
	err := deserializer.FromYAML([]byte(yaml), schema, &yamlDestination)

## Validation

To begin using the module:

1. Create a new instance of the Validation struct. You can use the convenience method NewValidation which sets Validation.validateOnFirstMatch to `true`.

	The following parameters can be set using the builder method (prefixed `With`) or Set (prefixed `Set):

		- validateOnFirstMatch - Using `Validation.WithValidateOnFirstMatch` or `Validation.SetValidateOnFirstMatch`.
		- customValidators - Custom validation logic based on data type. Can be set using `Validation.WithCustomValidators` or `Validation.SetCustomValidators`.

2. Validate data using Validation.ValidateData.

Example:

	schema := &DynamicSchemaNode{
		Kind: reflect.String,
		Type: reflect.TypeOf(""),
	}

	validation := NewValidation()
	ok, err := validation.ValidateData("this is a string", schema)
*/
package schema
