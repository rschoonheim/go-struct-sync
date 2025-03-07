# GO Struct Compare
A Go package for comparing struct instances and applying detected changes.


### Installation
```bash
go get github.com/rschoonheim/go-struct-compare
````

## Overview
This package provides utilities to:
- Compare two structs and detect differences (added, modified, or deleted fields)
- Apply a set of changes to a struct
- Filter, merge, and manipulate change sets
- Convert changes to human-readable format or JSON

## Usage

### Basic Comparison

```go
package main

import (
    "fmt"
    "github.com/rschoonheim/go-struct-compare"
)

func main() {
    type Person struct {
        Name    string
        Age     int
        Address string
        Active  bool
    }

    // Original struct
    old := Person{
        Name:    "John Doe",
        Age:     30,
        Address: "123 Main St", 
        Active:  true,
    }

    // Modified struct
    new := Person{
        Name:    "John Doe",
        Age:     31,
        Address: "456 Oak Ave",
        Active:  false,
    }

    // Compare the structs
    changes, err := comparing_structs_for_changes.CompareStructs(old, new)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // Print the changes
    fmt.Println(comparing_structs_for_changes.FormatChanges(changes))
}
```

### Applying Changes

```go
// Apply changes to the original struct
result, err := comparing_structs_for_changes.ApplyChanges(old, changes)
if err != nil {
    fmt.Printf("Error applying changes: %v\n", err)
    return
}

// Cast to the correct type
modifiedPerson := result.(Person)
fmt.Printf("Modified person: %+v\n", modifiedPerson)
```

### Filtering Changes

```go
// Get only modified fields
modifiedChanges := comparing_structs_for_changes.FilterChanges(
    changes, 
    []comparing_structs_for_changes.ChangeType{comparing_structs_for_changes.Modified}, 
    nil,
)

// Get only changes to specific fields
addressChanges := comparing_structs_for_changes.FilterChanges(
    changes, 
    nil, 
    []string{"Address"},
)
```

### Working with Change Maps

```go
// Convert changes to map for efficient lookup
changeMap := comparing_structs_for_changes.ChangesToMap(changes)

// Get change for specific field
addressChange, exists := changeMap["Address"]
if exists {
    fmt.Printf("Address changed from %v to %v\n", addressChange.OldValue, addressChange.NewValue)
}
```

## API Reference

### Types

```go
type ChangeType int

const (
    Modified ChangeType = iota
    Added
    Deleted
)

type Change struct {
    Field      string
    ChangeType ChangeType
    OldValue   interface{}
    NewValue   interface{}
}
```

### Functions

```go
// Compares two structs and returns a list of changes
func CompareStructs(old, new interface{}) ([]Change, error)

// Applies a list of changes to a struct
func ApplyChanges(original interface{}, changes []Change) (interface{}, error)

// Filters changes by type and/or field name
func FilterChanges(changes []Change, changeTypes []ChangeType, fields []string) []Change

// Converts a list of changes to a map keyed by field name
func ChangesToMap(changes []Change) map[string]Change

// Merges multiple change lists, with later changes taking precedence
func MergeChanges(changeLists ...[]Change) []Change

// Returns a human-readable representation of changes
func FormatChanges(changes []Change) string

// Serializes changes to JSON
func ChangesToJSON(changes []Change) ([]byte, error)

// Deserializes changes from JSON
func ChangesFromJSON(data []byte) ([]Change, error)

// Creates a new change list that would undo the given changes
func RevertChanges(changes []Change) []Change
```

## License
MIT License

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.