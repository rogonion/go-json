package schema

import (
	"encoding/json"
	"reflect"

	"github.com/rogonion/go-json/path"
	"go.yaml.in/yaml/v4"
)

type Processor struct {
	validateOnFirstMatch bool
	validators           map[reflect.Type]Validator
	converters           map[reflect.Type]Converter
}

func NewProcessor(validateOnFirstMatch bool, validators map[reflect.Type]Validator, converters map[reflect.Type]Converter) *Processor {
	n := new(Processor)
	n.validateOnFirstMatch = validateOnFirstMatch
	n.validators = validators
	n.converters = converters
	return n
}

func (n *Processor) Convert(data any, schema Schema, destination any) error {
	const FunctionName = "Convert"

	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return NewError(ErrDataConversionFailed, FunctionName, "destination is not a pointer", schema, data, nil)
	}

	if result, err := n.convert(reflect.ValueOf(data), schema, path.RecursiveDescentSegment{
		{
			Key:       "$",
			IsKeyRoot: true,
		},
	}); err != nil {
		return err
	} else {
		dest := reflect.ValueOf(destination)
		if result.Type() != reflect.TypeOf(destination) && reflect.TypeOf(destination).Elem().Kind() != reflect.Interface {
			return NewError(ErrDataConversionFailed, FunctionName, "destination and result type mismatch", schema, data, nil)
		}
		dest.Elem().Set(result)
	}
	return nil
}

func (n *Processor) DeserializeFromYaml(data []byte, schema Schema, destination any) error {
	const FunctionName = "DeserializeFromYaml"

	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return NewError(ErrDataDeserializationFailed, FunctionName, "destination is not a pointer", schema, data, nil)
	}

	var deserializedData interface{}
	if err := yaml.Unmarshal(data, &deserializedData); err != nil {
		return NewError(err, FunctionName, "Unmarshal from Yaml failed", schema, data, nil)
	}

	return n.deserializeDeserializedData(deserializedData, string(data), schema, destination)
}

func (n *Processor) DeserializeFromJson(data []byte, schema Schema, destination any) error {
	const FunctionName = "DeserializeFromJson"

	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return NewError(ErrDataDeserializationFailed, FunctionName, "destination is not a pointer", schema, data, nil)
	}

	var deserializedData interface{}
	if err := json.Unmarshal(data, &deserializedData); err != nil {
		return NewError(err, FunctionName, "Unmarshal from JSON failed", schema, data, nil)
	}

	return n.deserializeDeserializedData(deserializedData, string(data), schema, destination)
}

func (n *Processor) deserializeDeserializedData(deserializedData any, data string, schema Schema, destination any) error {
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

func (n *Processor) ValidateData(data any, schema Schema) (bool, error) {
	return n.validateData(reflect.ValueOf(data), schema, path.RecursiveDescentSegment{
		{
			Key:       "$",
			IsKeyRoot: true,
		},
	})
}

func (n *Processor) SetConverter(key reflect.Type, value Converter) {
	if n.converters == nil {
		n.converters = make(map[reflect.Type]Converter)
	}
	n.converters[key] = value
}

func (n *Processor) SetConverters(converters map[reflect.Type]Converter) {
	n.converters = converters
}

func (n *Processor) SetValidator(key reflect.Type, value Validator) {
	if n.validators == nil {
		n.validators = make(map[reflect.Type]Validator)
	}
	n.validators[key] = value
}

func (n *Processor) SetValidators(validators map[reflect.Type]Validator) {
	n.validators = validators
}

func (n *Processor) SetValidateOnFirstMatch(value bool) {
	n.validateOnFirstMatch = value
}
