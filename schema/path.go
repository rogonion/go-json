package schema

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/internal"
	jsonpath "github.com/rogonion/go-json/path"
)

/*
GetSchemaAtPath returns Schema found at SchemaPath.

Expects an absolute path to specific data.
*/
func GetSchemaAtPath[T SchemaPath](path T, schema Schema) (*DynamicSchemaNode, error) {
	const FunctionName = "GetSchemaAtPath"

	var pathToSchema jsonpath.RecursiveDescentSegment

	switch p := any(path).(type) {
	case jsonpath.RecursiveDescentSegment:
		pathToSchema = p
	case jsonpath.JSONPath:
		parsedJsonPath := p.Parse()
		if len(parsedJsonPath) == 1 {
			pathToSchema = parsedJsonPath[0]
			break
		}
		return nil, NewError(FunctionName, "Parsed JSON path contains multiple recursive descent segments").WithSchema(schema).WithNestedError(ErrSchemaPathError)
	case jsonpath.RecursiveDescentSegments:
		if len(p) == 1 {
			pathToSchema = p[0]
			break
		}
		return nil, NewError(FunctionName, "path contains multiple recursive descent segments").WithSchema(schema).WithNestedError(ErrSchemaPathError)
	default:
		return nil, NewError(FunctionName, "unsupported path type").WithSchema(schema).WithNestedError(ErrSchemaPathError)
	}

	n := new(schemaAtPath)
	n.RecursiveDescentSegment = pathToSchema
	return n.recursiveGetSchemaAtPath(internal.PathSegmentsIndexes{CurrentCollection: 0, LastCollection: len(pathToSchema) - 1}, schema)
}

func (n *schemaAtPath) recursiveGetSchemaAtPath(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema Schema) (*DynamicSchemaNode, error) {
	const FunctionName = "recursiveGetSchemaAtPath"

	switch s := currentSchema.(type) {
	case *DynamicSchema:
		return n.recursiveGetDynamicSchemaAtPath(currentPathSegmentIndexes, s)
	case *DynamicSchemaNode:
		return n.recursiveGetDynamicSchemaNodeAtPath(currentPathSegmentIndexes, s)
	default:
		return nil, NewError(FunctionName, "unsupported schema type").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
	}
}

func (n *schemaAtPath) recursiveGetDynamicSchemaAtPath(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema *DynamicSchema) (*DynamicSchemaNode, error) {
	const FunctionName = "recursiveGetDynamicSchemaAtPath"

	if len(currentSchema.DefaultSchemaNodeKey) > 0 {
		if dynamicSchemaNode, found := currentSchema.Nodes[currentSchema.DefaultSchemaNodeKey]; found {
			if result, err := n.recursiveGetDynamicSchemaNodeAtPath(currentPathSegmentIndexes, dynamicSchemaNode); err == nil {
				currentSchema.ValidSchemaNodeKeys = append(currentSchema.ValidSchemaNodeKeys, currentSchema.DefaultSchemaNodeKey)
				return result, nil
			}
		}
	}

	if len(currentSchema.Nodes) == 0 {
		return nil, NewError(FunctionName, "no schema nodes found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
	}

	var lastSchemaNodeErr error
	for schemaNodeKey, dynamicSchemaNode := range currentSchema.Nodes {
		if schemaNodeKey == currentSchema.DefaultSchemaNodeKey {
			continue
		}
		result, err := n.recursiveGetDynamicSchemaNodeAtPath(currentPathSegmentIndexes, dynamicSchemaNode)
		if err == nil {
			currentSchema.ValidSchemaNodeKeys = append(currentSchema.ValidSchemaNodeKeys, schemaNodeKey)
			return result, nil
		}
		lastSchemaNodeErr = err
	}

	return nil, lastSchemaNodeErr
}

func (n *schemaAtPath) getDefaultDynamicSchemaNode(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema *DynamicSchema) (*DynamicSchemaNode, error) {
	const FunctionName = "getDefaultDynamicSchemaNode"

	if len(currentSchema.Nodes) == 0 {
		return nil, NewError(FunctionName, "no schema nodes found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
	}

	if currentSchema.DefaultSchemaNodeKey != "" {
		if defaultDynamicSchemaNode, ok := currentSchema.Nodes[currentSchema.DefaultSchemaNodeKey]; ok {
			return defaultDynamicSchemaNode, nil
		}
	}
	for _, node := range currentSchema.Nodes {
		return node, nil
	}

	return nil, nil
}

func (n *schemaAtPath) recursiveGetDynamicSchemaNodeAtPath(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema *DynamicSchemaNode) (*DynamicSchemaNode, error) {
	const FunctionName = "recursiveGetDynamicSchemaNodeAtPath"

	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return nil, NewError(FunctionName, "current path segment indexes exhausted").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment).WithNestedError(ErrSchemaPathError)
	}

	currentPathSegment := n.RecursiveDescentSegment[currentPathSegmentIndexes.CurrentCollection]

	if currentPathSegment == nil {
		return nil, NewError(FunctionName, "current path segment empty").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
	}

	if currentPathSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			return currentSchema, nil
		}

		nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
		return n.recursiveGetDynamicSchemaNodeAtPath(nextPathSegmentIndexes, currentSchema)
	}

	var collectionKey string
	if currentPathSegment.IsKey {
		collectionKey = currentPathSegment.Key
	} else if currentPathSegment.IsIndex {
		collectionKey = fmt.Sprintf("%d", currentPathSegment.Index)
	} else {
		return nil, NewError(FunctionName, "current path segment is not key or index").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
	}

	switch currentSchema.Kind {
	case reflect.Pointer:
		if currentSchema.ChildNodesPointerSchema == nil {
			return nil, NewError(FunctionName, "schema for value that pointer points to not found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
		}

		return n.recursiveGetSchemaAtPath(currentPathSegmentIndexes, currentSchema.ChildNodesPointerSchema)
	case reflect.Map:
		if len(currentSchema.ChildNodes) > 0 {
			if associativeCollectionEntrySchema, ok := currentSchema.ChildNodes[collectionKey]; ok {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					switch aces := associativeCollectionEntrySchema.(type) {
					case *DynamicSchemaNode:
						return aces, nil
					case *DynamicSchema:
						nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
						return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, aces)
					default:
						return nil, NewError(FunctionName, "unsupported schema type").WithSchema(aces).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
					}
				}

				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				if result, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, associativeCollectionEntrySchema); err == nil {
					return result, nil
				}
			}
		}

		if currentSchema.ChildNodesAssociativeCollectionEntriesKeySchema == nil || currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema == nil {
			return nil, NewError(FunctionName, "schema for associative collection keys and/or values not found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
		}

		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			newAssociativeCollectionEntryKeySchema := new(DynamicSchemaNode)
			if value, err := n.recursiveGetSchemaAtPath(currentPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesKeySchema); err != nil {
				return nil, NewError(FunctionName, "default schema for all keys in associative entries not found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(err)
			} else {
				*newAssociativeCollectionEntryKeySchema = *value
			}

			newAssociativeCollectionEntrySchema := new(DynamicSchemaNode)
			if value, err := n.recursiveGetSchemaAtPath(currentPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema); err != nil {
				return nil, NewError(FunctionName, "default schema for all values in associative entries not found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(err)
			} else {
				*newAssociativeCollectionEntrySchema = *value
			}

			newAssociativeCollectionEntrySchema.AssociativeCollectionEntryKeySchema = newAssociativeCollectionEntryKeySchema
			return newAssociativeCollectionEntrySchema, nil
		}

		nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
		return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema)
	case reflect.Slice, reflect.Array:
		if len(currentSchema.ChildNodes) > 0 {
			if linearCollectionElementSchema, ok := currentSchema.ChildNodes[collectionKey]; ok {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					switch lces := linearCollectionElementSchema.(type) {
					case *DynamicSchemaNode:
						return lces, nil
					case *DynamicSchema:
						nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
						return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, lces)
					default:
						return nil, NewError(FunctionName, "unsupported path type").WithSchema(lces).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
					}
				}

				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				if result, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, linearCollectionElementSchema); err == nil {
					return result, nil
				}
			}
		}

		if currentSchema.ChildNodesLinearCollectionElementsSchema == nil {
			return nil, NewError(FunctionName, "schema for linear collection elements not found").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
		}

		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			switch cnlces := currentSchema.ChildNodesLinearCollectionElementsSchema.(type) {
			case *DynamicSchemaNode:
				return cnlces, nil
			case *DynamicSchema:
				return n.getDefaultDynamicSchemaNode(currentPathSegmentIndexes, cnlces)
			default:
				return nil, NewError(FunctionName, "unsupported path type").WithSchema(cnlces).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
			}
		}

		nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
		return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, currentSchema.ChildNodesLinearCollectionElementsSchema)
	case reflect.Struct:
		if currentSchema.ChildNodes == nil {
			return nil, NewError(FunctionName, "schema for struct fields is empty").WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
		}

		if structFieldSchema, ok := currentSchema.ChildNodes[collectionKey]; ok {
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				switch sfs := structFieldSchema.(type) {
				case *DynamicSchemaNode:
					return sfs, nil
				case *DynamicSchema:
					nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
					return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, sfs)
				default:
					return nil, NewError(FunctionName, "unsupported path type").WithSchema(sfs).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
				}
			}

			nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
			return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, structFieldSchema)
		}

		return nil, NewError(FunctionName, fmt.Sprintf("schema for struct field %s not found", collectionKey)).WithSchema(currentSchema).WithPathSegments(n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]).WithNestedError(ErrSchemaPathError)
	default:
		return currentSchema, nil
	}
}

type schemaAtPath struct {
	RecursiveDescentSegment jsonpath.RecursiveDescentSegment
}
