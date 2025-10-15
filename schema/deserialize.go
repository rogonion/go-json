package schema

import (
	"encoding/json"
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"go.yaml.in/yaml/v4"
)

func (n *Deserialization) deserializeWithDynamicSchemaNode(source reflect.Value, schema *DynamicSchemaNode, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserializeWithDynamicSchemaNode"

	if internal.IsNilOrInvalid(source) {
		if !schema.Nilable {
			return reflect.Value{}, NewError(ErrDataValidationAgainstSchemaFailed, FunctionName, "source cannot be nil", schema, source.Interface(), pathSegments)
		}
		return reflect.ValueOf(schema.DefaultValue), nil
	}

	if schema.Kind == reflect.Interface {
		return source, nil
	}

	if schema.Converter != nil {
		return schema.Converter.Convert(source, schema, pathSegments)
	}

	if customDeserializer, ok := n.customConverters[schema.Type]; ok {
		return customDeserializer.Convert(source, schema, pathSegments)
	}

	return n.defaultConverter.RecursiveConvert(source, schema, pathSegments)
}

func (n *Deserialization) deserializeWithDynamicSchema(source reflect.Value, schema *DynamicSchema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserializeWithDynamicSchema"

	if len(schema.DefaultSchemaNodeKey) > 0 {
		if dynamicSchemaNode, found := schema.Nodes[schema.DefaultSchemaNodeKey]; found {
			if result, err := n.deserializeWithDynamicSchemaNode(source, dynamicSchemaNode, pathSegments); err == nil {
				schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schema.DefaultSchemaNodeKey)
				return result, nil
			}
		}
	}

	if len(schema.Nodes) == 0 {
		return reflect.Value{}, NewError(ErrDataDeserializationFailed, FunctionName, "no schema nodes found", schema, source, pathSegments)
	}

	var lastSchemaNodeErr error
	for schemaNodeKey, dynamicSchemaNode := range schema.Nodes {
		if schemaNodeKey == schema.DefaultSchemaNodeKey {
			continue
		}
		result, err := n.deserializeWithDynamicSchemaNode(source, dynamicSchemaNode, pathSegments)
		if err == nil {
			schema.ValidSchemaNodeKeys = append(schema.ValidSchemaNodeKeys, schemaNodeKey)
			return result, nil
		}
		lastSchemaNodeErr = err
	}

	return reflect.Value{}, lastSchemaNodeErr
}

func (n *Deserialization) deserialize(source reflect.Value, schema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "deserialize"

	switch s := schema.(type) {
	case *DynamicSchema:
		return n.deserializeWithDynamicSchema(source, s, pathSegments)
	case *DynamicSchemaNode:
		return n.deserializeWithDynamicSchemaNode(source, s, pathSegments)
	default:
		return reflect.Value{}, NewError(ErrDataDeserializationFailed, FunctionName, "unsupported schema type", schema, source, pathSegments)
	}
}

func (n *Deserialization) deserializeDeserializedData(deserializedData any, data string, schema Schema, destination any) error {
	const FunctionName = "deserializeDeserializedData"

	if result, err := n.deserialize(reflect.ValueOf(deserializedData), schema, path.RecursiveDescentSegment{
		{
			Key:       "$",
			IsKeyRoot: true,
		},
	}); err != nil {
		return err
	} else {
		dest := reflect.ValueOf(destination)
		if result.Kind() != reflect.Pointer {
			if result.Type() != reflect.TypeOf(destination) && reflect.TypeOf(destination).Elem().Kind() != reflect.Interface {
				return NewError(ErrDataDeserializationFailed, FunctionName, "destination and result type mismatch", schema, data, nil)
			}
			dest.Elem().Set(result)
		} else {
			if result.Elem().Type() != reflect.ValueOf(destination).Elem().Type() {
				return NewError(ErrDataDeserializationFailed, FunctionName, "destination and result type mismatch", schema, data, nil)
			}
			dest.Elem().Set(result.Elem())
		}
	}

	return nil
}

func (n *Deserialization) FromYAML(data []byte, schema Schema, destination any) error {
	const FunctionName = "FromYAML"

	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return NewError(ErrDataDeserializationFailed, FunctionName, "destination is not a pointer", schema, data, nil)
	}

	var deserializedData interface{}
	if err := yaml.Unmarshal(data, &deserializedData); err != nil {
		return NewError(err, FunctionName, "Unmarshal from Yaml failed", schema, data, nil)
	}

	return n.deserializeDeserializedData(deserializedData, string(data), schema, destination)
}

func (n *Deserialization) FromJSON(data []byte, schema Schema, destination any) error {
	const FunctionName = "FromJSON"

	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return NewError(ErrDataDeserializationFailed, FunctionName, "destination is not a pointer", schema, data, nil)
	}

	var deserializedData interface{}
	if err := json.Unmarshal(data, &deserializedData); err != nil {
		return NewError(err, FunctionName, "Unmarshal from JSON failed", schema, data, nil)
	}

	return n.deserializeDeserializedData(deserializedData, string(data), schema, destination)
}

func (n *Deserialization) WithCustomConverters(value Converters) *Deserialization {
	n.customConverters = value
	return n
}

func (n *Deserialization) SetCustomConverters(value Converters) {
	n.customConverters = value
}

func (n *Deserialization) WithDefaultConverter(value DefaultConverter) *Deserialization {
	n.defaultConverter = value
	return n
}

func (n *Deserialization) SetDefaultConverter(value DefaultConverter) {
	n.defaultConverter = value
}

func NewDeserialization() *Deserialization {
	n := new(Deserialization)
	n.defaultConverter = NewConversion()
	return n
}

type Deserialization struct {
	// Base converter to use.
	defaultConverter DefaultConverter

	// Specialized converter to use immediately after parsing from source if reflect.Type matches.
	customConverters Converters
}
