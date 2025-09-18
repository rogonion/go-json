package schemapath

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/internal"
	jsonpath "github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

// GetSchemaAtPath returns schema.Schema found at SchemaPath.
//
// Expects an absolute path to specific data.
func GetSchemaAtPath[T SchemaPath](path T, schema schema.Schema) (*schema.DynamicSchemaNode, error) {
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
		return nil, NewError(ErrSchemaPathError, FunctionName, "Parsed JSON path contains multiple recursive descent segments", schema, nil)
	case jsonpath.RecursiveDescentSegments:
		if len(p) == 1 {
			pathToSchema = p[0]
			break
		}
		return nil, NewError(ErrSchemaPathError, FunctionName, "path contains multiple recursive descent segments", schema, nil)
	default:
		return nil, NewError(ErrSchemaPathError, FunctionName, "unsupported path type", schema, nil)
	}

	n := new(getSchemaAtPath)
	n.RecursiveDescentSegment = pathToSchema
	return n.recursiveGetSchemaAtPath(internal.PathSegmentsIndexes{CurrentCollection: 0, LastCollection: len(pathToSchema) - 1}, schema)
}

func (n *getSchemaAtPath) recursiveGetSchemaAtPath(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema schema.Schema) (*schema.DynamicSchemaNode, error) {
	const FunctionName = "recursiveGetSchemaAtPath"

	switch s := currentSchema.(type) {
	case *schema.DynamicSchema:
		return n.recursiveGetDynamicSchemaAtPath(currentPathSegmentIndexes, s)
	case *schema.DynamicSchemaNode:
		return n.recursiveGetDynamicSchemaNodeAtPath(currentPathSegmentIndexes, s)
	default:
		return nil, NewError(ErrSchemaPathError, FunctionName, "unsupported schema type", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
	}
}

func (n *getSchemaAtPath) recursiveGetDynamicSchemaAtPath(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema *schema.DynamicSchema) (*schema.DynamicSchemaNode, error) {
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
		return nil, NewError(ErrSchemaPathError, FunctionName, "no schema nodes found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
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

func (n *getSchemaAtPath) getDefaultDynamicSchemaNode(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema *schema.DynamicSchema) (*schema.DynamicSchemaNode, error) {
	const FunctionName = "getDefaultDynamicSchemaNode"

	if len(currentSchema.Nodes) == 0 {
		return nil, NewError(ErrSchemaPathError, FunctionName, "no schema nodes found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
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

func (n *getSchemaAtPath) recursiveGetDynamicSchemaNodeAtPath(currentPathSegmentIndexes internal.PathSegmentsIndexes, currentSchema *schema.DynamicSchemaNode) (*schema.DynamicSchemaNode, error) {
	const FunctionName = "recursiveGetDynamicSchemaNodeAtPath"

	if currentPathSegmentIndexes.CurrentCollection > currentPathSegmentIndexes.LastCollection {
		return nil, NewError(ErrSchemaPathError, FunctionName, "current path segment indexes exhausted", currentSchema, n.RecursiveDescentSegment)
	}

	currentPathSegment := n.RecursiveDescentSegment[currentPathSegmentIndexes.CurrentCollection]

	if currentPathSegment == nil {
		return nil, NewError(ErrSchemaPathError, FunctionName, "current path segment empty", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
	}

	if currentPathSegment.IsKeyRoot {
		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			return currentSchema, nil
		}

		newIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
		return n.recursiveGetDynamicSchemaNodeAtPath(newIndexes, currentSchema)
	}

	var collectionKey string
	if currentPathSegment.IsKey {
		collectionKey = currentPathSegment.Key
	} else if currentPathSegment.IsIndex {
		collectionKey = fmt.Sprintf("%d", currentPathSegment.Index)
	} else {
		return nil, NewError(ErrSchemaPathError, FunctionName, "current path segment is not key or index", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
	}

	switch currentSchema.Kind {
	case reflect.Pointer:
		if currentSchema.ChildNodesPointerSchema == nil {
			return nil, NewError(ErrSchemaPathError, FunctionName, "schema for value that pointer points to not found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
		}

		return n.recursiveGetSchemaAtPath(currentPathSegmentIndexes, currentSchema.ChildNodesPointerSchema)
	case reflect.Map:
		if len(currentSchema.ChildNodes) > 0 {
			if associativeCollectionEntrySchema, ok := currentSchema.ChildNodes[collectionKey]; ok {
				if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
					switch aces := associativeCollectionEntrySchema.(type) {
					case *schema.DynamicSchemaNode:
						return aces, nil
					case *schema.DynamicSchema:
						nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
						return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, aces)
					default:
						return nil, NewError(ErrSchemaPathError, FunctionName, "unsupported path type", aces, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
					}
				}

				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, associativeCollectionEntrySchema)
			}
		}

		if currentSchema.ChildNodesAssociativeCollectionEntriesKeySchema == nil || currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema == nil {
			return nil, NewError(ErrSchemaPathError, FunctionName, "schema for associative collection keys and/or values not found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
		}

		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection, LastCollection: currentPathSegmentIndexes.LastCollection}

			newAssociativeCollectionEntryKeySchema := new(schema.DynamicSchemaNode)
			if value, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesKeySchema); err != nil {
				return nil, NewError(err, FunctionName, "default schema for all keys in associative entries not found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
			} else {
				*newAssociativeCollectionEntryKeySchema = *value
			}

			newAssociativeCollectionEntrySchema := new(schema.DynamicSchemaNode)
			if value, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema); err != nil {
				return nil, NewError(err, FunctionName, "default schema for all values in associative entries not found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
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
					case *schema.DynamicSchemaNode:
						return lces, nil
					case *schema.DynamicSchema:
						nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
						return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, lces)
					default:
						return nil, NewError(ErrSchemaPathError, FunctionName, "unsupported path type", lces, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
					}
				}

				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				if result, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, linearCollectionElementSchema); err == nil {
					return result, nil
				}
			}
		}

		if currentSchema.ChildNodesLinearCollectionElementsSchema == nil {
			return nil, NewError(ErrSchemaPathError, FunctionName, "schema for linear collection elements not found", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
		}

		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			switch cnlces := currentSchema.ChildNodesLinearCollectionElementsSchema.(type) {
			case *schema.DynamicSchemaNode:
				return cnlces, nil
			case *schema.DynamicSchema:
				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, cnlces)
			default:
				return nil, NewError(ErrSchemaPathError, FunctionName, "unsupported path type", cnlces, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
			}
		}

		nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
		return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, currentSchema.ChildNodesLinearCollectionElementsSchema)
	case reflect.Struct:
		if currentSchema.ChildNodes == nil {
			return nil, NewError(ErrSchemaPathError, FunctionName, "schema for struct fields is empty", currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
		}

		if structFieldSchema, ok := currentSchema.ChildNodes[collectionKey]; ok {
			if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
				switch sfs := structFieldSchema.(type) {
				case *schema.DynamicSchemaNode:
					return sfs, nil
				case *schema.DynamicSchema:
					nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
					return n.getDefaultDynamicSchemaNode(nextPathSegmentIndexes, sfs)
				default:
					return nil, NewError(ErrSchemaPathError, FunctionName, "unsupported path type", sfs, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
				}
			}

			nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
			return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, structFieldSchema)
		}

		return nil, NewError(ErrSchemaPathError, FunctionName, fmt.Sprintf("schema for struct field %s not found", collectionKey), currentSchema, n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1])
	default:
		return currentSchema, nil
	}
}

type getSchemaAtPath struct {
	RecursiveDescentSegment jsonpath.RecursiveDescentSegment
}
