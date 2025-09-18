package object

import (
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/schema"
)

// User is a simple struct with exported fields.
type User struct {
	ID    int
	Name  string
	Email string
}

// Address is a nested struct with exported fields.
type Address struct {
	Street  string
	City    string
	ZipCode *string
}

// ComplexData is a highly nested struct with different data types.
type ComplexData struct {
	ID      int
	Details map[string]any
	Items   []struct {
		Name  string
		Value int
	}
	User User
}

func AddressSchema() *schema.DynamicSchemaNode {
	return &schema.DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Address{}),
		ChildNodes: map[string]schema.Schema{
			"Street": &schema.DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"City": &schema.DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"ZipCode": &schema.DynamicSchemaNode{
				Kind: reflect.Pointer,
				Type: reflect.TypeOf(internal.Ptr("")),
				ChildNodesPointerSchema: &schema.DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
			},
		},
	}
}

type UserProfile struct {
	Name    string
	Age     int
	Address Address
}

func UserProfileSchema() *schema.DynamicSchemaNode {
	return &schema.DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(UserProfile{}),
		ChildNodes: map[string]schema.Schema{
			"Name": &schema.DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Age": &schema.DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
			"Address": AddressSchema(),
		},
	}
}
