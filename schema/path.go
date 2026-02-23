package schema

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/core"
	"github.com/rogonion/go-json/internal"
	jsonpath "github.com/rogonion/go-json/path"
)

/*
GetSchemaAtPath traverses the provided schema using the given path and returns the schema node corresponding to that path.

It expects an absolute path (starting with $) and does not support recursive descent ('..') in the path query for schema retrieval.
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
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("Parsed JSON path contains multiple recursive descent segments").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": schema})
	case jsonpath.RecursiveDescentSegments:
		if len(p) == 1 {
			pathToSchema = p[0]
			break
		}
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("path contains multiple recursive descent segments").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": schema})
	default:
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported path type").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": schema})
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
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported schema type").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("no schema nodes found").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("no schema nodes found").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("current path segment indexes exhausted").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment})
	}

	currentPathSegment := n.RecursiveDescentSegment[currentPathSegmentIndexes.CurrentCollection]

	if currentPathSegment == nil {
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("current path segment empty").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
		return nil, NewError().WithFunctionName(FunctionName).WithMessage("current path segment is not key or index").
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
	}

	switch currentSchema.Kind {
	case reflect.Pointer:
		if currentSchema.ChildNodesPointerSchema == nil {
			return nil, NewError().WithFunctionName(FunctionName).WithMessage("schema for value that pointer points to not found").
				WithNestedError(ErrSchemaPathError).
				WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
						return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported schema type").
							WithNestedError(ErrSchemaPathError).
							WithData(core.JsonObject{"Schema": aces, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
					}
				}

				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				if result, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, associativeCollectionEntrySchema); err == nil {
					return result, nil
				}
			}
		}

		if currentSchema.ChildNodesAssociativeCollectionEntriesKeySchema == nil || currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema == nil {
			return nil, NewError().WithFunctionName(FunctionName).WithMessage("schema for associative collection keys and/or values not found").
				WithNestedError(ErrSchemaPathError).
				WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
		}

		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			newAssociativeCollectionEntryKeySchema := new(DynamicSchemaNode)
			if value, err := n.recursiveGetSchemaAtPath(currentPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesKeySchema); err != nil {
				return nil, NewError().WithFunctionName(FunctionName).WithMessage("default schema for all keys in associative entries not found").
					WithNestedError(err).
					WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
			} else {
				*newAssociativeCollectionEntryKeySchema = *value
			}

			newAssociativeCollectionEntrySchema := new(DynamicSchemaNode)
			if value, err := n.recursiveGetSchemaAtPath(currentPathSegmentIndexes, currentSchema.ChildNodesAssociativeCollectionEntriesValueSchema); err != nil {
				return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported path type").
					WithNestedError(err).
					WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
						return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported path type").
							WithNestedError(ErrSchemaPathError).
							WithData(core.JsonObject{"Schema": lces, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
					}
				}

				nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
				if result, err := n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, linearCollectionElementSchema); err == nil {
					return result, nil
				}
			}
		}

		if currentSchema.ChildNodesLinearCollectionElementsSchema == nil {
			return nil, NewError().WithFunctionName(FunctionName).WithMessage("schema for linear collection elements not found").
				WithNestedError(ErrSchemaPathError).
				WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
		}

		if currentPathSegmentIndexes.CurrentCollection == currentPathSegmentIndexes.LastCollection {
			switch cnlces := currentSchema.ChildNodesLinearCollectionElementsSchema.(type) {
			case *DynamicSchemaNode:
				return cnlces, nil
			case *DynamicSchema:
				return n.getDefaultDynamicSchemaNode(currentPathSegmentIndexes, cnlces)
			default:
				return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported path type").
					WithNestedError(ErrSchemaPathError).
					WithData(core.JsonObject{"Schema": cnlces, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
			}
		}

		nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
		return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, currentSchema.ChildNodesLinearCollectionElementsSchema)
	case reflect.Struct:
		if currentSchema.ChildNodes == nil {
			return nil, NewError().WithFunctionName(FunctionName).WithMessage("schema child nodes empty").
				WithNestedError(ErrSchemaPathError).
				WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
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
					return nil, NewError().WithFunctionName(FunctionName).WithMessage("unsupported path type").
						WithNestedError(ErrSchemaPathError).
						WithData(core.JsonObject{"Schema": sfs, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
				}
			}

			nextPathSegmentIndexes := internal.PathSegmentsIndexes{CurrentCollection: currentPathSegmentIndexes.CurrentCollection + 1, LastCollection: currentPathSegmentIndexes.LastCollection}
			return n.recursiveGetSchemaAtPath(nextPathSegmentIndexes, structFieldSchema)
		}

		return nil, NewError().WithFunctionName(FunctionName).WithMessage(fmt.Sprintf("schema for struct field %s not found", collectionKey)).
			WithNestedError(ErrSchemaPathError).
			WithData(core.JsonObject{"Schema": currentSchema, "PathSegments": n.RecursiveDescentSegment[:currentPathSegmentIndexes.CurrentCollection+1]})
	default:
		return currentSchema, nil
	}
}

type schemaAtPath struct {
	RecursiveDescentSegment jsonpath.RecursiveDescentSegment
}
