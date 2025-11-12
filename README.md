# go-json

A library for working with objects i.e, data whose layout and structure resembles `JSON` tree structure.

An object can be a primitive value or a large and deeply nested collection.

Particularly useful for:

- Deeply nested objects.
- Objects whose type is dynamic.
- Objects whose type is discoverable at runtime only.

Relies on the language's reflection capabilities.

## Sections

- [Installation](#installation)
- [JSONPath](#jsonpath)
- [Modules](#modules)
    - [Object](#object)
    - [Path](#path)
    - [Schema](#schema)
- [Supported data types](#supported-data-types)

## Installation

```shell
go get github.com/rogonion/go-json
```

## JSONPath

As defined [here](https://en.wikipedia.org/wiki/JSONPath), it is a query language for working with values JSON style.

The module aims to extract path segments from a JSONPath string.

The module aims to support the entirety of the JSONPath [spec](https://www.rfc-editor.org/rfc/rfc9535.html) except for
the filter expression.

Noted supported spec as follows:

- Identifiers: `$`.
- Segments: Dot notation with recursive descent (search), bracket notation.
- Selectors: wildcard, array slice, union, name, index.

Example JSONPaths:

- `$..name..first`
- `$.address[1:4:2]['country-of-orign']`
- `$[*].countries[1,3,5]`

## Modules

### Object

This [module](object) allows one to manipulate data using [JSONPath](#jsonpath).

With this module, you can do the following:

- Set value(s) in an object.
- Get value(s) in an object.
- Delete value(s) in an object.
- Loop through each value(s) in an object (For Each).
- Check if two values are equal with options for setting up custom equal checkers.

Example usage for manipulating a single object:

```go
package main

import "github.com/rogonion/go-json/object"

type Address struct {
	Street  string
	City    string
	ZipCode *string
}

var source any = map[string]any{
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

var objManip *object.Object = object.NewObject(source)

var valueFound any
var ok bool
var err error
var noOfModifications uint64

valueFound, ok, err = objManip.Get("$.data.metadata.Address.City")

noOfModifications, err = objManip.Set("$.data.metadata.Status", "inactive")

noOfModifications, err = objManip.Delete("$.data.metadata.Status")

// retrieve modified source after Set/Delete
var modifiedSource any = objManip.GetSource()

```

### Path

[Module](path) for converting a [JSONPath](#jsonpath) string into a 2D array of detailed information about each path
segment.

Such information is used when manipulating data using the core modules like get, set, and delete.

Example parsing JSONPath string:

```go
package main

import "github.com/rogonion/go-json/path"

var jsonPath path.JSONPath = "$[1,3,5]"
var parsedPath path.RecursiveDescentSegments = jsonPath.Parse()
```

### Schema

[Module](schema) for defining and working with the schema of an object. This includes the data type as well as the tree
structure of
every simple primitive, linear collection element, or associative collection entry in an object.

Useful for the following purposes:

- Validating if an object adheres to a defined schema. Allows extension with custom validators.
- Converting an object to a schema defined type. Allows extension with custom converters.
- Deserializing data from json or yaml to a schema defined type. Allows extension with custom deserializers.
- Recursively creating new nested objects with the [Set](objectV1/set.go) module. For example, a source empty nil value
  of
  type any can end up being an array of pointers to structs if that is the schema definition.
- Fetch the schema of data at a `JSONPath`.

Example usage:

#### Conversion

```go
package main

import (
	"reflect"

	"github.com/rogonion/go-json/schema"
)

var sch schema.Schema = &schema.DynamicSchemaNode{
	Kind: reflect.Map,
	Type: reflect.TypeOf(map[int]int{}),
	ChildNodesAssociativeCollectionEntriesKeySchema: &schema.DynamicSchemaNode{
		Kind: reflect.Int,
		Type: reflect.TypeOf(0),
	},
	ChildNodesAssociativeCollectionEntriesValueSchema: &schema.DynamicSchemaNode{
		Kind: reflect.Int,
		Type: reflect.TypeOf(0),
	},
}

var source any = map[string]string{
	"1": "1",
	"2": "2",
	"3": "3",
}
var destination any
var converter schema.DefaultConverter = schema.NewConversion()
var err error = converter.Convert(source, sch, &destination)

```

#### Deserialization

```go
package main

import (
	"reflect"
	"strings"

	"github.com/rogonion/go-json/schema"
)

type UserProfile2 struct {
	Name       string
	Age        int
	Country    string
	Occupation string
}

var sch schema.Schema = &schema.DynamicSchemaNode{
	Kind: reflect.Struct,
	Type: reflect.TypeOf(UserProfile2{}),
	ChildNodes: map[string]schema.Schema{
		"Name": &schema.DynamicSchemaNode{
			Kind: reflect.String,
			Type: reflect.TypeOf(""),
		},
		"Age": &schema.DynamicSchemaNode{
			Kind: reflect.Int,
			Type: reflect.TypeOf(0),
		},
		"Country": &schema.DynamicSchemaNode{
			Kind: reflect.String,
			Type: reflect.TypeOf(""),
		},
		"Occupation": &schema.DynamicSchemaNode{
			Kind: reflect.String,
			Type: reflect.TypeOf(""),
		},
	},
}

var deserializer schema.Deserializer = schema.NewDeserialization()

var json string = "{\"Name\":\"John Doe\"}"
var jsonDestination UserProfile2
var err error = deserializer.FromJSON([]byte(json), sch, &jsonDestination)

var yaml string = strings.TrimSpace(`Name: John Doe`)
var yamlDestination UserProfile2
err = deserializer.FromYAML([]byte(yaml), schema, &yamlDestination)

```

#### Validation

```go
package main

import (
	"reflect"

	"github.com/rogonion/go-json/schema"
)

var sch schema.Schema = &schema.DynamicSchemaNode{
	Kind: reflect.String,
	Type: reflect.TypeOf(""),
}

var validation schema.DefaultValidator = schema.NewValidation()
var ok bool
var err error
ok, err = validation.ValidateData("this is a string", schema)

```

## Supported data types

- Primitive types:
    - Signed integer: `int`, `int8`, `int16`, `int32`, `int64`.
    - Unsigned integer: `uint`, `uint8`, `uint16`, `uint32`, `uint64`.
    - Float: `float32`, `float64`.
    - `boolean`
    - `string`
- Collection types:
    - Linear:
        - `arrays`
        - `slices`
    - Associative:
        - `structs`
        - `maps`
- `pointers` to values of primitive and collection types.

