package object

import (
	"fmt"
	"reflect"

	"github.com/rogonion/go-json/internal"
	"github.com/rogonion/go-json/path"
	"github.com/rogonion/go-json/schema"
)

type jsonPathModifications struct {
	jsonPath
	noOfModifications uint64
	lastError         error
}

func (n *jsonPath) SetSchemaProcessor(schemaProcessor schema.DataProcessor) {
	n.schemaProcessor = schemaProcessor
}

func (n *jsonPath) MapKeyString(mapKey reflect.Value) string {
	x := fmt.Sprintf("%v", internal.JsonStringifyMust(mapKey.Interface()))
	return x[1 : len(x)-1]
}

type jsonPath struct {
	recursiveDescentSegments path.RecursiveDescentSegments
	schemaProcessor          schema.DataProcessor
}
