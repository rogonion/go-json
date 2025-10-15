# go-json

A library for working with objects i.e, data whose layout and structure resembles `JSON`'s tree structure.

An object can be a primitive value or a large and deeply nested collection.

Particularly useful for:

- Deeply nested objects.
- Objects whose type is dynamic.
- Objects whose type is discoverable at runtime only.

Relies on the language's reflection capabilities.

The [core set of modules](object) for manipulating an object using `JSONPath` are as follows:

- Set value(s) in an object.
- Get value(s) in an object.
- Delete value(s) in an object.
- Loop through each value(s) in an object (For Each).
- Check if two values are equal.

## Sections

- [JSONPath](#jsonpath)
- [Supported data type](#supported-data-types)
- [Additional modules](#additional-modules)

## JSONPath

As defined [here](https://en.wikipedia.org/wiki/JSONPath), it is a query language for querying values JSON-style.

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

## Additional modules

These modules support the core objective of the library.

### Path

[Module](path) for converting a [JSONPath](#jsonpath) string into a 2D array of detailed information about each path
segment.

Such information is used when manipulating data using the core modules like get, set, and delete.

### Schema

[Module](schema) for defining and working with the schema of an object. This includes the data type as well as the tree structure of
every simple primitive, linear collection element, or associative collection entry in an object.

Useful for the following purposes:

- Validating if an object adheres to a defined schema. Allows extension with custom validators.
- Converting an object to a schema defined type. Allows extension with custom converters.
- Deserializing data from json or yaml to a schema defined type. Allows extension with custom deserializers.
- Recursively creating new nested objects with the [Set](object/set.go) module. For example, a source empty nil value of
  type any can end up being an array of pointers to structs if that is the schema definition.
- Fetch the schema of data at a `JSONPath`.

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
    - Associative:
        - `structs`
        - `slices`
- `pointers` to values of primitive and collection types.

