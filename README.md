# go-json

A reflection-based library for manipulating dynamic, JSON-like data structures in Go. It provides tools for traversing, modifying, validating, and converting deeply nested objects (maps, structs, slices, arrays) using JSONPath.

## Features

- **Dynamic Object Manipulation**: Get, Set, Delete, and Iterate over values in deeply nested structures using JSONPath.
- **Schema Validation**: Define schemas for your data and validate dynamic objects against them at runtime.
- **Type Conversion**: Convert loosely typed data (e.g., `map[string]any`) into strongly typed Go structs, maps, and slices based on schema definitions.
- **Deserialization**: Helpers for loading JSON and YAML data directly into schema-validated structures.
- **JSONPath Support**: Supports dot notation, recursive descent (`..`), wildcards (`*`), unions (`['a','b']`), and array slicing (`[start:end:step]`).

## Prerequisites

<table>
  <thead>
    <tr>
      <th>Tool</th>
      <th>Description</th>
      <th>Link</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Go</td>
      <td></td>
      <td><a href="https://go.dev/">Official Website</a></td>
    </tr>
    <tr>
      <td>Task</td>
      <td>
        <p>Task runner / build tool.</p> 
        <p>You can use the provided shell script <a href="taskw">taskw</a> that automatically downloads the binary and installs it in the <code>.task</code> folder.</p>
      </td>
      <td><a href="https://taskfile.dev/">Official Website</a></td>
    </tr>
    <tr>
      <td>Docker / Podman</td>
      <td>Optional container engine for isolated development environment.</td>
      <td><a href="https://www.docker.com/">Docker</a> / <a href="https://podman.io/">Podman</a></td>
    </tr>
  </tbody>
</table>

After building the dev container, below is a sample script that runs the container and mounts the project directory into the container:

```shell
#!/bin/bash

CONTAINER_ENGINE="podman"
CONTAINER="projects-go-json"
NETWORK="systemd-leap"
NETWORK_ALIAS="projects-go-json"
CONTAINER_UID=1000
IMAGE="localhost/projects/go-json:latest"
SSH_PORT="127.0.0.1:2200" # for local proxy vscode ssh access
PROJECT_DIRECTORY="$(pwd)"

# Check if container exists (Running or Stopped)
if $CONTAINER_ENGINE ps -a --format '{{.Names}}' | grep -q "^$CONTAINER$"; then
    echo "   Found existing container: $CONTAINER"
    # Check if it is currently running
    if $CONTAINER_ENGINE ps --format '{{.Names}}' | grep -q "^$CONTAINER$"; then
        echo "✅ Container is already running."
    else
        echo "🔄 Container stopped. Starting it..."
        $CONTAINER_ENGINE start $CONTAINER
        echo "✅ Started."
    fi
else
    # Container doesn't exist -> Create and Run it
    echo "🆕 Container not found. Creating new..."
    $CONTAINER_ENGINE run -d \
    # start container from scratch
    # `sudo` is used because systemd-leap network was created in `sudo`
    # Ensure container image exists in `sudo`
    # Not needed if target network is not in `sudo`
    sudo podman run -d \
        --name $CONTAINER \
        --network $NETWORK \
        --network-alias $NETWORK_ALIAS \
        --user $CONTAINER_UID:$CONTAINER_UID \
        -p $SSH_PORT:22 \
        -v $PROJECT_DIRECTORY:/home/dev/go-json:Z \
        $IMAGE
    echo "✅ Created and Started."
fi
```

## Installation

```shell
go get github.com/rogonion/go-json
```

## Environment Setup

This project uses `Taskfile` to manage the development environment and tasks.

<table>
  <thead>
    <tr>
      <th>Task</th>
      <th>Description</th>
      <th>Usage</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>env:build</code></td>
      <td>
        <p>Build the dev container image.</p>
        <p>Image runs an ssh server one can connect to with vscode.</p>
      </td>
      <td><code>task env:build</code></td>
    </tr>
    <tr>
      <td><code>env:info</code></td>
      <td>Show current environment configuration.</td>
      <td><code>task env:info</code></td>
    </tr>
    <tr>
      <td><code>deps</code></td>
      <td>Download and tidy dependencies.</td>
      <td><code>task deps</code></td>
    </tr>
    <tr>
      <td><code>test</code></td>
      <td>Run tests. Supports optional <code>TARGET</code> variable.</td>
      <td><code>task test</code><br><code>task test TARGET=./object</code></td>
    </tr>
  </tbody>
</table>

## Modules

### 1. Object

The `object` package is the core of the library, allowing you to manipulate data structures.

**Key Capabilities:**
- `Get`: Retrieve values.
- `Set`: Update or insert values (auto-creates nested structures if schema is provided).
- `Delete`: Remove values.
- `ForEach`: Iterate over matches.
- `AreEqual`: Deep comparison.

**Example:**

```go
package main

import (
	"fmt"
	"github.com/rogonion/go-json/object"
)

func main() {
	data := map[string]any{
		"users": []any{
			map[string]any{"name": "Alice", "id": 1},
			map[string]any{"name": "Bob", "id": 2},
		},
	}

	obj := object.NewObject().WithSourceInterface(data)

	// Get
	val, _ := obj.Get("$.users[0].name")
	fmt.Println(obj.GetValueFoundInterface()) // Output: Alice

	// Set
	obj.Set("$.users[1].active", true)

	// Delete
	obj.Delete("$.users[0]")
}
```

### 2. Schema

The `schema` package allows you to define the expected structure of your data. This is useful for validation and conversion of dynamic data.

**Example: Validation**

```go
package main

import (
	"reflect"
	"github.com/rogonion/go-json/schema"
)

func main() {
	// Define a schema
	userSchema := &schema.DynamicSchemaNode{
		Kind: reflect.Struct,
		Type: reflect.TypeOf(struct{ Name string; Age int }{}),
		ChildNodes: schema.ChildNodes{
			"Name": &schema.DynamicSchemaNode{Kind: reflect.String, Type: reflect.TypeOf("")},
			"Age":  &schema.DynamicSchemaNode{Kind: reflect.Int, Type: reflect.TypeOf(0)},
		},
	}

	validator := schema.NewValidation()
	data := map[string]any{"Name": "Alice", "Age": 30}
	
	// Validate
	ok, err := validator.ValidateData(data, userSchema)
}
```

**Example: Conversion**

```go
package main

import (
	"reflect"
	"github.com/rogonion/go-json/schema"
)

func main() {
	// Define schema (same as above)
	// ...

	source := map[string]any{"Name": "Alice", "Age": "30"} // Age is string in source
	var dest struct{ Name string; Age int }

	converter := schema.NewConversion()
	// Converts and coerces types (string "30" -> int 30)
	err := converter.Convert(source, userSchema, &dest)
}
```

### 3. Path

The `path` package handles parsing of JSONPath strings. It is primarily used internally by the `object` package but can be used directly to inspect paths.

```go
import "github.com/rogonion/go-json/path"

p := path.JSONPath("$.store.book[*].author")
segments := p.Parse()
```

## Supported Data Types

The library supports reflection-based manipulation of:
- **Primitives**: `int` (all sizes), `uint` (all sizes), `float32/64`, `bool`, `string`.
- **Collections**: `map`, `struct`, `slice`, `array`.
- **Pointers**: Pointers to any of the above.
- **Interfaces**: `any` / `interface{}`.
