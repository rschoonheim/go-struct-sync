# GO Struct Compare
This repository contains utilities to compare two Go structs and apply the differences to a target struct.


## Example
The following example demonstrates how to compare two structs and apply the changes to the original struct.

```go
package main

import "fmt"

func main() {
	// Example usage
	type Person struct {
		Name    string
		Age     int
		Address string
		Active  bool
	}

	old := Person{
		Name:    "John Doe",
		Age:     30,
		Address: "123 Main St",
		Active:  true,
	}

	new := Person{
		Name:    "John Doe",
		Age:     31,
		Address: "456 Oak Ave",
		Active:  false,
	}

	changes, err := CompareStructs(old, new)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Apply changes to the original struct
	result, err := ApplyChanges(old, changes)
	if err != nil {
		fmt.Printf("Error applying changes: %v\n", err)
		return
	}

	// Cast to the correct type
	modifiedPerson := result.(Person)
	fmt.Printf("Modified person: %+v\n", modifiedPerson)
}

```