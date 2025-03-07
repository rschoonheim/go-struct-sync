package comparing_structs_for_changes

import "testing"

func TestCompareStructsDetectsModifiedValues(t *testing.T) {
	old := Person{
		Name:    "John",
		Age:     30,
		Active:  true,
		Address: "123 Main St",
	}

	new := Person{
		Name:    "John",
		Age:     35,
		Active:  false,
		Address: "123 Main St",
	}

	changes, err := CompareStructs(old, new)
	if err != nil {
		t.Fatalf("CompareStructs failed: %v", err)
	}

	if len(changes) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(changes))
	}

	// Verify Age was changed
	ageChange := findChangeByField(changes, "Age")
	if ageChange == nil || ageChange.ChangeType != Modified || ageChange.OldValue.(int) != 30 || ageChange.NewValue.(int) != 35 {
		t.Errorf("Age change not detected correctly")
	}

	// Verify Active was changed
	activeChange := findChangeByField(changes, "Active")
	if activeChange == nil || activeChange.ChangeType != Modified || activeChange.OldValue.(bool) != true || activeChange.NewValue.(bool) != false {
		t.Errorf("Active change not detected correctly")
	}
}

func TestCompareStructsDetectsDeletedValues(t *testing.T) {
	old := Person{
		Name:     "John",
		Children: []string{"Alice", "Bob"},
		Manager:  &Person{Name: "Boss"},
	}

	new := Person{
		Name:     "John",
		Children: []string{},
		Manager:  nil,
	}

	changes, err := CompareStructs(old, new)
	if err != nil {
		t.Fatalf("CompareStructs failed: %v", err)
	}

	childrenChange := findChangeByField(changes, "Children")
	if childrenChange == nil || childrenChange.ChangeType != Deleted {
		t.Errorf("Deleted Children not detected correctly")
	}

	managerChange := findChangeByField(changes, "Manager")
	if managerChange == nil || managerChange.ChangeType != Deleted {
		t.Errorf("Deleted Manager not detected correctly")
	}
}

func TestCompareStructsDetectsAddedValues(t *testing.T) {
	old := Person{
		Name:     "John",
		Children: []string{},
		Manager:  nil,
		Address:  "",
	}

	new := Person{
		Name:     "John",
		Children: []string{"Charlie"},
		Manager:  &Person{Name: "NewBoss"},
		Address:  "456 Oak St",
	}

	changes, err := CompareStructs(old, new)
	if err != nil {
		t.Fatalf("CompareStructs failed: %v", err)
	}

	childrenChange := findChangeByField(changes, "Children")
	if childrenChange == nil || childrenChange.ChangeType != Added {
		t.Errorf("Added Children not detected correctly")
	}

	managerChange := findChangeByField(changes, "Manager")
	if managerChange == nil || managerChange.ChangeType != Added {
		t.Errorf("Added Manager not detected correctly")
	}

	addressChange := findChangeByField(changes, "Address")
	if addressChange == nil || addressChange.ChangeType != Added {
		t.Errorf("Added Address not detected correctly")
	}
}

func TestCompareStructsWithIdenticalStructs(t *testing.T) {
	person1 := Person{Name: "John", Age: 30}
	person2 := Person{Name: "John", Age: 30}

	changes, err := CompareStructs(person1, person2)
	if err != nil {
		t.Fatalf("CompareStructs failed: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("Expected 0 changes for identical structs, got %d", len(changes))
	}
}

func TestCompareStructsWithPointers(t *testing.T) {
	old := &Person{Name: "John", Age: 30}
	new := &Person{Name: "Jane", Age: 30}

	changes, err := CompareStructs(old, new)
	if err != nil {
		t.Fatalf("CompareStructs failed: %v", err)
	}

	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	nameChange := findChangeByField(changes, "Name")
	if nameChange == nil || nameChange.ChangeType != Modified || nameChange.OldValue.(string) != "John" || nameChange.NewValue.(string) != "Jane" {
		t.Errorf("Name change not detected correctly with pointers")
	}
}

func TestCompareStructsFailsOnNonStruct(t *testing.T) {
	old := "not a struct"
	new := "also not a struct"

	_, err := CompareStructs(old, new)
	if err == nil {
		t.Error("Expected error when comparing non-structs")
	}
}

func TestCompareStructsFailsOnDifferentTypes(t *testing.T) {
	type OtherStruct struct {
		Field string
	}

	old := Person{Name: "John"}
	new := OtherStruct{Field: "value"}

	_, err := CompareStructs(old, new)
	if err == nil {
		t.Error("Expected error when comparing different struct types")
	}
}

// Helper function to find a change by field name
func findChangeByField(changes []Change, fieldName string) *Change {
	for _, c := range changes {
		if c.Field == fieldName {
			return &c
		}
	}
	return nil
}
