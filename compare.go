package comparing_structs_for_changes

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
