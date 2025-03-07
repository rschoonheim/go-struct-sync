package comparing_structs_for_changes

import (
	"encoding/json"
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

// FilterChanges - returns a subset of changes that match the provided criteria
func FilterChanges(changes []Change, changeTypes []ChangeType, fields []string) []Change {
	if len(changeTypes) == 0 && len(fields) == 0 {
		return changes
	}

	typeMap := make(map[ChangeType]bool)
	for _, ct := range changeTypes {
		typeMap[ct] = true
	}

	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}

	result := make([]Change, 0, len(changes))
	for _, change := range changes {
		matchesType := len(typeMap) == 0 || typeMap[change.ChangeType]
		matchesField := len(fieldMap) == 0 || fieldMap[change.Field]
		if matchesType && matchesField {
			result = append(result, change)
		}
	}
	return result
}

// RevertChanges - creates a new change list that would undo the given changes
func RevertChanges(changes []Change) []Change {
	reverted := make([]Change, len(changes))

	for i, change := range changes {
		reverted[i] = Change{
			Field:    change.Field,
			OldValue: change.NewValue,
			NewValue: change.OldValue,
		}

		switch change.ChangeType {
		case Added:
			reverted[i].ChangeType = Deleted
		case Deleted:
			reverted[i].ChangeType = Added
		case Modified:
			reverted[i].ChangeType = Modified
		}
	}

	return reverted
}

// ChangesToMap - converts a list of changes to a map for efficient lookup by field name
func ChangesToMap(changes []Change) map[string]Change {
	result := make(map[string]Change, len(changes))
	for _, change := range changes {
		result[change.Field] = change
	}
	return result
}

// MergeChanges - combines multiple change lists, with later changes taking precedence
func MergeChanges(changeLists ...[]Change) []Change {
	merged := make(map[string]Change)

	for _, list := range changeLists {
		for _, change := range list {
			merged[change.Field] = change
		}
	}

	result := make([]Change, 0, len(merged))
	for _, change := range merged {
		result = append(result, change)
	}
	return result
}

// FormatChanges - returns a human-readable representation of the changes
func FormatChanges(changes []Change) string {
	var result string

	for _, change := range changes {
		switch change.ChangeType {
		case Modified:
			result += fmt.Sprintf("Modified %s: %v → %v\n", change.Field, change.OldValue, change.NewValue)
		case Added:
			result += fmt.Sprintf("Added %s: %v\n", change.Field, change.NewValue)
		case Deleted:
			result += fmt.Sprintf("Deleted %s (was %v)\n", change.Field, change.OldValue)
		}
	}

	return result
}

// ChangesToJSON - serializes a list of changes to JSON
func ChangesToJSON(changes []Change) ([]byte, error) {
	return json.Marshal(changes)
}

// ChangesFromJSON - deserializes a list of changes from JSON
func ChangesFromJSON(data []byte) ([]Change, error) {
	var changes []Change
	err := json.Unmarshal(data, &changes)
	return changes, err
}
