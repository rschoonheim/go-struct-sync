package comparing_structs_for_changes

import (
	"reflect"
	"testing"
)

// Test struct for applying changes
type Person struct {
	Name     string
	Age      int
	Active   bool
	Address  string
	Children []string
	Manager  *Person
	private  string
}

func TestApplyChangesModifiesValues(t *testing.T) {
	original := Person{
		Name:    "John",
		Age:     30,
		Active:  true,
		Address: "123 Main St",
	}

	changes := []Change{
		{Field: "Name", ChangeType: Modified, NewValue: "Jane"},
		{Field: "Age", ChangeType: Modified, NewValue: 31},
		{Field: "Active", ChangeType: Modified, NewValue: false},
	}

	result, err := ApplyChanges(original, changes)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	modified := result.(Person)
	if modified.Name != "Jane" || modified.Age != 31 || modified.Active != false {
		t.Errorf("Expected modified values, got: %+v", modified)
	}
	if modified.Address != "123 Main St" {
		t.Errorf("Unchanged fields should remain the same")
	}
}

func TestApplyChangesToPointer(t *testing.T) {
	original := &Person{
		Name: "John",
		Age:  30,
	}

	changes := []Change{
		{Field: "Name", ChangeType: Modified, NewValue: "Jane"},
	}

	result, err := ApplyChanges(original, changes)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	modified := result.(*Person)
	if modified.Name != "Jane" || modified.Age != 30 {
		t.Errorf("Expected only Name to be modified, got: %+v", modified)
	}
}

func TestApplyChangesDeletesFields(t *testing.T) {
	original := Person{
		Name:     "John",
		Age:      30,
		Children: []string{"Alice", "Bob"},
		Manager:  &Person{Name: "Boss"},
	}

	changes := []Change{
		{Field: "Name", ChangeType: Deleted, OldValue: "John"},
		{Field: "Children", ChangeType: Deleted, OldValue: []string{"Alice", "Bob"}},
		{Field: "Manager", ChangeType: Deleted, OldValue: &Person{Name: "Boss"}},
	}

	result, err := ApplyChanges(original, changes)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	modified := result.(Person)
	if modified.Name != "" {
		t.Errorf("Name should be empty string after deletion")
	}
	if modified.Children != nil {
		t.Errorf("Children should be nil after deletion")
	}
	if modified.Manager != nil {
		t.Errorf("Manager should be nil after deletion")
	}
}

func TestApplyChangesAddsValues(t *testing.T) {
	original := Person{
		Name: "John",
	}

	changes := []Change{
		{Field: "Age", ChangeType: Added, NewValue: 25},
		{Field: "Active", ChangeType: Added, NewValue: true},
		{Field: "Children", ChangeType: Added, NewValue: []string{"Child"}},
	}

	result, err := ApplyChanges(original, changes)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	modified := result.(Person)
	if modified.Name != "John" || modified.Age != 25 || modified.Active != true {
		t.Errorf("Expected values to be added, got: %+v", modified)
	}
	if len(modified.Children) != 1 || modified.Children[0] != "Child" {
		t.Errorf("Expected Children slice to be added")
	}
}

func TestApplyChangesWithEmptyChangesList(t *testing.T) {
	original := Person{Name: "John", Age: 30}
	changes := []Change{}

	result, err := ApplyChanges(original, changes)
	if err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}

	modified := result.(Person)
	if !reflect.DeepEqual(modified, original) {
		t.Errorf("With empty changes list, result should equal original")
	}
}

func TestApplyChangesFailsOnNonStruct(t *testing.T) {
	original := "not a struct"
	changes := []Change{
		{Field: "something", ChangeType: Modified, NewValue: "new"},
	}

	_, err := ApplyChanges(original, changes)
	if err == nil {
		t.Error("Expected error when applying changes to non-struct")
	}
}

func TestApplyChangesFailsOnNonExistentField(t *testing.T) {
	original := Person{Name: "John"}
	changes := []Change{
		{Field: "NonExistentField", ChangeType: Modified, NewValue: "value"},
	}

	_, err := ApplyChanges(original, changes)
	if err == nil {
		t.Error("Expected error when field doesn't exist")
	}
}

func TestApplyChangesFailsOnUnexportedField(t *testing.T) {
	original := Person{Name: "John"}
	changes := []Change{
		{Field: "private", ChangeType: Modified, NewValue: "value"},
	}

	_, err := ApplyChanges(original, changes)
	if err == nil {
		t.Error("Expected error when field is unexported")
	}
}

func TestApplyChangesFailsOnTypeConversionError(t *testing.T) {
	original := Person{Name: "John"}
	changes := []Change{
		{Field: "Name", ChangeType: Modified, NewValue: struct{}{}},
	}

	_, err := ApplyChanges(original, changes)
	if err == nil {
		t.Error("Expected error when type conversion isn't possible")
	}
}
