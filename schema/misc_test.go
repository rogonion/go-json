package schema

import (
	"fmt"
	"reflect"

	"github.com/gofrs/uuid/v5"
	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/path"
)

func JsonMapSchema() Schema {
	return &DynamicSchemaNode{
		Kind: reflect.Map,
		Type: reflect.TypeOf(map[string]interface{}{}),
		ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
			Kind: reflect.String,
			Type: reflect.TypeOf(""),
		},
		ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
			Kind: reflect.Interface,
		},
	}
}

type NestedItem struct {
	ID       int
	MapData  map[string]interface{}
	ListData []interface{}
}

func NestedItemSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(NestedItem{}),
		ChildNodes: ChildNodes{
			"ID": &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
			"MapData": &DynamicSchemaNode{
				Kind: reflect.Map,
				Type: reflect.TypeOf(map[string]interface{}{}),
				ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
				ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
					Kind: reflect.Interface,
				},
			},
			"ListData": &DynamicSchemaNode{
				Kind: reflect.Slice,
				Type: reflect.TypeOf([]interface{}{}),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{
					Kind: reflect.Interface,
				},
			},
		},
	}
}

func ListOfNestedItemSchema() Schema {
	return &DynamicSchemaNode{
		Kind:                                     reflect.Slice,
		Type:                                     reflect.TypeOf([]NestedItem{}),
		ChildNodesLinearCollectionElementsSchema: NestedItemSchema(),
	}
}

type Shape interface {
	isShape() bool
}

func ShapeSchema() Schema {
	return &DynamicSchema{
		DefaultSchemaNodeKey: DynamicSchemaDefaultNodeKey,
		Nodes: DynamicSchemaNodes{
			"Circle": circleSchema(),
			"Square": squareSchema(),
		},
	}
}

func ListOfShapesSchema() Schema {
	var x any
	return &DynamicSchemaNode{
		Kind: reflect.Slice,
		Type: reflect.TypeOf([]Shape{}),
		ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{
			Kind:                    reflect.Pointer,
			Type:                    reflect.TypeOf(core.Ptr(x)),
			ChildNodesPointerSchema: ShapeSchema(),
		},
	}
}

type Circle struct {
	Radius float64
}

func (c *Circle) isShape() bool {
	return true
}

func circleSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Circle{}),
		ChildNodes: ChildNodes{
			"Radius": &DynamicSchemaNode{
				Kind: reflect.Float64,
				Type: reflect.TypeOf(0.0),
			},
		},
	}
}

type Square struct {
	Side float64
}

func (s *Square) isShape() bool {
	return true
}

func squareSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Square{}),
		ChildNodes: ChildNodes{
			"Side": &DynamicSchemaNode{
				Kind: reflect.Float64,
				Type: reflect.TypeOf(0.0),
			},
		},
	}
}

type user2 struct {
	ID   uuid.UUID
	Name string
}

func User2Schema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(user2{}),
		ChildNodes: ChildNodes{
			"ID": &DynamicSchemaNode{
				Kind: reflect.TypeOf(uuid.UUID{}).Kind(),
				Type: reflect.TypeOf(uuid.UUID{}),
			},
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
		},
	}
}

type UserWithAddress struct {
	Name    string
	Address *Address
}

func UserWithAddressSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(UserWithAddress{}),
		ChildNodes: ChildNodes{
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Address": &DynamicSchemaNode{
				Kind:                    reflect.Pointer,
				Type:                    reflect.TypeOf(core.Ptr(Address{})),
				ChildNodesPointerSchema: AddressSchema(),
			},
		},
	}
}

type UserWithUuidId struct {
	ID      uuid.UUID
	Profile UserProfile2
}

func UserWithUuidIdSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(UserWithUuidId{}),
		ChildNodes: ChildNodes{
			"ID": &DynamicSchemaNode{
				Kind:      reflect.TypeOf(uuid.UUID{}).Kind(),
				Type:      reflect.TypeOf(uuid.UUID{}),
				Validator: core.Ptr(Pgxuuid{}),
				Converter: core.Ptr(Pgxuuid{}),
			},
			"Profile": UserProfile2Schema(),
		},
	}
}

type User struct {
	ID    int
	Name  string
	Email string
}

func UserSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(User{}),
		ChildNodes: ChildNodes{
			"ID": &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(int(0)),
			},
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Email": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
		},
	}
}

func MapUserSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Map,
		Type: reflect.TypeOf(map[int]*User{}),
		ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
			Kind: reflect.Int,
			Type: reflect.TypeOf(0),
		},
		ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
			Kind:                    reflect.Pointer,
			Type:                    reflect.TypeOf(core.Ptr(User{})),
			ChildNodesPointerSchema: UserSchema(),
		},
	}
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

func ProductSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Product{}),
		ChildNodes: ChildNodes{
			"ID": &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(int(0)),
			},
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Price": &DynamicSchemaNode{
				Kind: reflect.Float64,
				Type: reflect.TypeOf(float64(0)),
			},
		},
	}
}

func ListOfProductsSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind:                                     reflect.Slice,
		Type:                                     reflect.TypeOf([]Product{}),
		ChildNodesLinearCollectionElementsSchema: ProductSchema(),
	}
}

type Company struct {
	Name      string
	Employees []*User
}

func CompanySchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Company{}),
		ChildNodes: ChildNodes{
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Employees": &DynamicSchemaNode{
				Kind: reflect.Slice,
				Type: reflect.TypeOf([]*User{}),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{
					Kind:                    reflect.Pointer,
					Type:                    reflect.TypeOf(core.Ptr(User{})),
					ChildNodesPointerSchema: UserSchema(),
				},
			},
		},
	}
}

type Address struct {
	Street  string
	City    string
	ZipCode *string
}

func AddressSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Address{}),
		ChildNodes: ChildNodes{
			"Street": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"City": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"ZipCode": &DynamicSchemaNode{
				Kind: reflect.Pointer,
				Type: reflect.TypeOf(core.Ptr("")),
				ChildNodesPointerSchema: &DynamicSchemaNode{
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

func UserProfileSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(UserProfile{}),
		ChildNodes: ChildNodes{
			"Name": &DynamicSchemaNode{
				Kind: reflect.String,
				Type: reflect.TypeOf(""),
			},
			"Age": &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
			"Address": AddressSchema(),
		},
	}
}

type Employee struct {
	ID           int
	Profile      UserProfile
	Skills       []string
	ProjectHours map[string]int
}

func EmployeeSchema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(Employee{}),
		ChildNodes: ChildNodes{
			"ID": &DynamicSchemaNode{
				Kind: reflect.Int,
				Type: reflect.TypeOf(0),
			},
			"Profile": UserProfileSchema(),
			"Skills": &DynamicSchemaNode{
				Kind: reflect.Slice,
				Type: reflect.TypeOf([]string{}),
				ChildNodesLinearCollectionElementsSchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
			},
			"ProjectHours": &DynamicSchemaNode{
				Kind: reflect.Map,
				Type: reflect.TypeOf(map[string]int{}),
				ChildNodesAssociativeCollectionEntriesKeySchema: &DynamicSchemaNode{
					Kind: reflect.String,
					Type: reflect.TypeOf(""),
				},
				ChildNodesAssociativeCollectionEntriesValueSchema: &DynamicSchemaNode{
					Kind: reflect.Int,
					Type: reflect.TypeOf(0),
				},
			},
		},
	}
}

type UserProfile2 struct {
	Name       string
	Age        int
	Country    string
	Occupation string
}

func UserProfile2Schema() *DynamicSchemaNode {
	return &DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(UserProfile2{}),
		ChildNodes: ChildNodes{
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
}

func DynamicUserSchema() *DynamicSchema {
	return &DynamicSchema{
		DefaultSchemaNodeKey: "UserWithAddress",
		Nodes: DynamicSchemaNodes{
			"User":            UserSchema(),
			"User2":           User2Schema(),
			"UserProfile":     UserProfileSchema(),
			"UserProfile2":    UserProfile2Schema(),
			"UserWithUuidID":  UserWithUuidIdSchema(),
			"UserWithAddress": UserWithAddressSchema(),
		},
	}
}

type Pgxuuid struct{}

func (n *Pgxuuid) Convert(data reflect.Value, currentSchema Schema, pathSegments path.RecursiveDescentSegment) (reflect.Value, error) {
	const FunctionName = "RecursiveConvert"

	rawValue := data.Interface()
	switch d := rawValue.(type) {
	case string:
		if uuidString, err := uuid.FromString(d); err == nil {
			return reflect.ValueOf(uuidString), nil
		} else {
			return reflect.Value{}, NewError().WithFunctionName(FunctionName).WithMessage("RecursiveConvert to uuid from string failed").WithNestedError(err).WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
		}
	case []byte:
		if uuidBytes, err := uuid.FromBytes(d); err == nil {
			return reflect.ValueOf(uuidBytes), nil
		} else {
			return reflect.Value{}, NewError().WithFunctionName(FunctionName).WithMessage("RecursiveConvert to uuid from bytes failed").WithNestedError(err).WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
		}
	default:
		return reflect.Value{}, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("unsupported type %T", data)).WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
	}
}

func (n *Pgxuuid) ValidateData(data any, currentSchema Schema, pathSegments path.RecursiveDescentSegment) (bool, error) {
	const FunctionName = "ValidateData"

	switch d := data.(type) {
	case uuid.UUID:
		if d.IsNil() {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("uuid is nil").WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
		}
		return true, nil
	case string:
		if _, err := uuid.FromString(d); err != nil {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("string not valid uuid").WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
		}
		return true, nil
	case []byte:
		if _, err := uuid.FromBytes(d); err != nil {
			return false, NewError().WithFunctionName(FunctionName).WithMessage("[]bytes not valid uuid").WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
		}
		return true, nil
	default:
		return false, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("unsupported type %T", data)).WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": pathSegments, "Data": data})
	}
}
