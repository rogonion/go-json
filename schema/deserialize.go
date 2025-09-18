package schema

import (
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
)

func (n *Processor) deserialize(data reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserialize"

	switch s := schema.(type) {
	case *DynamicSchema:
		return n.deserializeWithDynamicSchema(data, s, pathSegments)
	case *DynamicSchemaNode:
		return n.deserializeWithDynamicSchemaNode(data, s, pathSegments)
	default:
		return reflect.Value{}, NewError(ErrDataDeserializationFailed, FunctionName, "unsupported schema type", schema, data, pathSegments)
	}
}

func (n *Processor) deserializeWithDynamicSchema(data reflect.Value, schema *DynamicSchema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserializeWithDynamicSchema"

	if len(schema.DefaultSchemaNodeKey) > 0 {
		if dynamicSchemaNode, found := schema.Nodes[schema.DefaultSchemaNodeKey]; found {
			if result, err := n.deserializeWithDynamicSchemaNode(data, dynamicSchemaNode, pathSegments); err == nil {
				schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schema.DefaultSchemaNodeKey)
				return result, nil
			}
		}
	}

	if len(schema.Nodes) == 0 {
		return reflect.Value{}, NewError(ErrDataDeserializationFailed, FunctionName, "no schema nodes found", schema, data, pathSegments)
	}

	var lastSchemaNodeErr error
	for schemaNodeKey, dynamicSchemaNode := range schema.Nodes {
		if schemaNodeKey == schema.DefaultSchemaNodeKey {
			continue
		}
		result, err := n.deserializeWithDynamicSchemaNode(data, dynamicSchemaNode, pathSegments)
		if err == nil {
			schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schemaNodeKey)
			return result, nil
		}
		lastSchemaNodeErr = err
	}

	return reflect.Value{}, lastSchemaNodeErr
}

func (n *Processor) deserializeWithDynamicSchemaNode(data reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserializeWithDynamicSchemaNode"

	if internal.IsNilOrInvalid(data) {
		if !schema.Nilable {
			return reflect.Value{}, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "data cannot be nil", schema, data.Interface(), pathSegments)
		}
		return reflect.ValueOf(schema.DefaultValue), nil
	}

	if schema.Kind == reflect.Interface {
		return data, nil
	}

	if schema.Converter != nil {
		return schema.Converter.Convert(data, schema, pathSegments)
	}

	if customDeserializer, ok := n.converters[schema.Type]; ok {
		return customDeserializer.Convert(data, schema, pathSegments)
	}

	return n.convert(data, schema, pathSegments)
}
