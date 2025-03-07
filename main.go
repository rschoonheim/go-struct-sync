package main

import (
	"fmt"
	"reflect"
	"sync"
)

// ChangeType represents the type of change
type ChangeType string

const (
	Modified ChangeType = "modified"
	Deleted  ChangeType = "deleted"
	Added    ChangeType = "added"
)

// Change represents a difference between two struct fields
type Change struct {
	Field      string
	ChangeType ChangeType
	OldValue   interface{}
	NewValue   interface{}
}

// CompareStructs compares two struct instances and returns a list of changes
func CompareStructs(old, new interface{}) ([]Change, error) {
	oldVal := reflect.ValueOf(old)
	newVal := reflect.ValueOf(new)

	// Dereference if pointers
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	// Validate input types
	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("both arguments must be structs")
	}
	if oldVal.Type() != newVal.Type() {
		return nil, fmt.Errorf("both structs must be of the same type")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	changes := make([]Change, 0, oldVal.NumField())

	// Cache field information
	type fieldInfo struct {
		oldField reflect.Value
		newField reflect.Value
		name     string
	}
	fields := make([]fieldInfo, oldVal.NumField())
	for i := 0; i < oldVal.NumField(); i++ {
		fields[i] = fieldInfo{
			oldField: oldVal.Field(i),
			newField: newVal.Field(i),
			name:     oldVal.Type().Field(i).Name,
		}
	}

	// Iterate through struct fields
	for _, field := range fields {
		wg.Add(1)
		go func(field fieldInfo) {
			defer wg.Done()

			// Skip unexported fields
			if !field.oldField.CanInterface() {
				return
			}

			// Compare values
			if !reflect.DeepEqual(field.oldField.Interface(), field.newField.Interface()) {
				changeType := Modified

				// Detect deletion based on type
				switch field.oldField.Kind() {
				case reflect.Ptr, reflect.Interface:
					if !field.oldField.IsNil() && field.newField.IsNil() {
						changeType = Deleted
					} else if field.oldField.IsNil() && !field.newField.IsNil() {
						changeType = Added
					}
				case reflect.Slice, reflect.Map:
					if field.oldField.Len() > 0 && field.newField.Len() == 0 {
						changeType = Deleted
					} else if field.oldField.Len() == 0 && field.newField.Len() > 0 {
						changeType = Added
					}
				case reflect.String:
					if field.oldField.String() != "" && field.newField.String() == "" {
						changeType = Deleted
					} else if field.oldField.String() == "" && field.newField.String() != "" {
						changeType = Added
					}
				}

				mu.Lock()
				changes = append(changes, Change{
					Field:      field.name,
					ChangeType: changeType,
					OldValue:   field.oldField.Interface(),
					NewValue:   field.newField.Interface(),
				})
				mu.Unlock()
			}
		}(field)
	}

	wg.Wait()
	return changes, nil
}

// PrintChanges prints the changes in a readable format
func PrintChanges(changes []Change) {
	if len(changes) == 0 {
		fmt.Println("No changes detected")
		return
	}

	fmt.Println("Changes detected:")
	for _, change := range changes {
		fmt.Printf("Field: %s (%s)\n", change.Field, change.ChangeType)
		fmt.Printf("  - Old: %v\n", change.OldValue)
		fmt.Printf("  - New: %v\n", change.NewValue)
	}
}

// ApplyChanges applies a list of changes to the original struct and returns a modified copy
func ApplyChanges(original interface{}, changes []Change) (interface{}, error) {
	// Extract and validate original value
	originalVal := reflect.ValueOf(original)
	isPointer := originalVal.Kind() == reflect.Ptr

	if isPointer {
		originalVal = originalVal.Elem()
	}

	if originalVal.Kind() != reflect.Struct {
		return nil, fmt.Errorf("original must be a struct")
	}

	originalType := originalVal.Type()

	// Create a new instance
	resultVal := reflect.New(originalType).Elem()

	// Copy all fields from original to result
	for i := 0; i < originalVal.NumField(); i++ {
		if originalVal.Field(i).CanInterface() && resultVal.Field(i).CanSet() {
			resultVal.Field(i).Set(originalVal.Field(i))
		}
	}

	// Create a field cache to avoid repeated lookups
	fieldCache := make(map[string]reflect.Value, len(changes))

	// Apply each change
	for _, change := range changes {
		// Check cache first before using reflection to find the field
		field, ok := fieldCache[change.Field]
		if !ok {
			field = resultVal.FieldByName(change.Field)
			if !field.IsValid() {
				return nil, fmt.Errorf("field %s not found", change.Field)
			}
			fieldCache[change.Field] = field
		}

		if !field.CanSet() {
			return nil, fmt.Errorf("field %s is not settable", change.Field)
		}

		switch change.ChangeType {
		case Deleted:
			// Set zero value for deleted fields
			field.Set(reflect.Zero(field.Type()))
		case Modified, Added:
			// Fast path for nil values
			if change.NewValue == nil {
				if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface ||
					field.Kind() == reflect.Map || field.Kind() == reflect.Slice {
					field.Set(reflect.Zero(field.Type()))
					continue
				}
			}

			// Handle non-nil values
			newValue := reflect.ValueOf(change.NewValue)

			// Direct set if types match
			if field.Type() == newValue.Type() {
				field.Set(newValue)
			} else if newValue.Type().ConvertibleTo(field.Type()) {
				field.Set(newValue.Convert(field.Type()))
			} else {
				return nil, fmt.Errorf("cannot convert value for field %s", change.Field)
			}
		}
	}

	// Return with the correct type
	if isPointer {
		ptr := reflect.New(originalType)
		ptr.Elem().Set(resultVal)
		return ptr.Interface(), nil
	}
	return resultVal.Interface(), nil
}

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
