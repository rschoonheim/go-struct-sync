package comparing_structs_for_changes

import (
	"fmt"
	"reflect"
)

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
