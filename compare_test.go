package comparing_structs_for_changes

import (
	"strings"
	"testing"
)

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

func TestFilterChangesWithChangeTypes(t *testing.T) {
	changes := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Age", ChangeType: Modified, OldValue: 30, NewValue: 31},
		{Field: "Address", ChangeType: Deleted, OldValue: "123 Main St"},
		{Field: "Active", ChangeType: Added, NewValue: true},
	}

	// Filter by Modified changes
	modified := FilterChanges(changes, []ChangeType{Modified}, nil)
	if len(modified) != 2 {
		t.Errorf("Expected 2 modified changes, got %d", len(modified))
	}
	if modified[0].Field != "Name" || modified[1].Field != "Age" {
		t.Errorf("Modified changes filtered incorrectly")
	}

	// Filter by Deleted changes
	deleted := FilterChanges(changes, []ChangeType{Deleted}, nil)
	if len(deleted) != 1 || deleted[0].Field != "Address" {
		t.Errorf("Deleted change filtered incorrectly")
	}

	// Filter by multiple change types
	combined := FilterChanges(changes, []ChangeType{Modified, Added}, nil)
	if len(combined) != 3 {
		t.Errorf("Expected 3 changes when filtering by Modified and Added, got %d", len(combined))
	}
}

func TestFilterChangesByFields(t *testing.T) {
	changes := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Age", ChangeType: Modified, OldValue: 30, NewValue: 31},
		{Field: "Address", ChangeType: Deleted, OldValue: "123 Main St"},
	}

	filtered := FilterChanges(changes, nil, []string{"Name", "Address"})
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered changes, got %d", len(filtered))
	}

	nameChange := findChangeByField(filtered, "Name")
	addressChange := findChangeByField(filtered, "Address")
	if nameChange == nil || addressChange == nil {
		t.Errorf("Field filtering didn't return correct changes")
	}
}

func TestChangesToMapProvidesEfficientLookup(t *testing.T) {
	changes := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Age", ChangeType: Modified, OldValue: 30, NewValue: 31},
		{Field: "Address", ChangeType: Deleted, OldValue: "123 Main St"},
	}

	changeMap := ChangesToMap(changes)
	if len(changeMap) != 3 {
		t.Errorf("Expected 3 items in change map, got %d", len(changeMap))
	}

	nameChange := changeMap["Name"]
	if nameChange.ChangeType != Modified || nameChange.NewValue.(string) != "Jane" {
		t.Errorf("ChangesToMap did not preserve change data correctly")
	}

	// Check a non-existent field
	_, exists := changeMap["NonExistent"]
	if exists {
		t.Errorf("Map should not contain entry for non-existent field")
	}
}

func TestMergeChangesCombinesChangeLists(t *testing.T) {
	changes1 := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Age", ChangeType: Modified, OldValue: 30, NewValue: 31},
	}

	changes2 := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Bob"}, // This should win
		{Field: "Address", ChangeType: Added, NewValue: "456 Oak St"},
	}

	merged := MergeChanges(changes1, changes2)
	if len(merged) != 3 {
		t.Errorf("Expected 3 merged changes, got %d", len(merged))
	}

	nameChange := findChangeByField(merged, "Name")
	if nameChange == nil || nameChange.NewValue.(string) != "Bob" {
		t.Errorf("Later changes did not take precedence in merged result")
	}

	addressChange := findChangeByField(merged, "Address")
	if addressChange == nil || addressChange.ChangeType != Added {
		t.Errorf("Added fields not preserved in merged result")
	}
}

func TestFormatChangesCreatesReadableOutput(t *testing.T) {
	changes := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Address", ChangeType: Deleted, OldValue: "123 Main St"},
		{Field: "Active", ChangeType: Added, NewValue: true},
	}

	formatted := FormatChanges(changes)

	if !strings.Contains(formatted, "Modified Name: John → Jane") {
		t.Errorf("Modified change not formatted correctly")
	}
	if !strings.Contains(formatted, "Deleted Address (was 123 Main St)") {
		t.Errorf("Deleted change not formatted correctly")
	}
	if !strings.Contains(formatted, "Added Active: true") {
		t.Errorf("Added change not formatted correctly")
	}
}

func TestChangesToJSONAndFromJSONRoundTrip(t *testing.T) {
	original := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Age", ChangeType: Modified, OldValue: 30, NewValue: 31},
		{Field: "Active", ChangeType: Added, NewValue: true},
	}

	// To JSON
	jsonData, err := ChangesToJSON(original)
	if err != nil {
		t.Fatalf("ChangesToJSON failed: %v", err)
	}

	// From JSON
	restored, err := ChangesFromJSON(jsonData)
	if err != nil {
		t.Fatalf("ChangesFromJSON failed: %v", err)
	}

	// Check if the round trip preserved essential data
	if len(restored) != len(original) {
		t.Errorf("JSON round trip changed the number of changes")
	}

	// Find a specific field and verify its properties
	nameChange := findChangeByField(restored, "Name")
	if nameChange == nil || nameChange.ChangeType != Modified {
		t.Errorf("JSON round trip did not preserve change data correctly")
	}
}

func TestChangesFromJSONHandlesInvalidJSON(t *testing.T) {
	invalidJSON := []byte("{not valid json")

	_, err := ChangesFromJSON(invalidJSON)
	if err == nil {
		t.Errorf("ChangesFromJSON should return error for invalid JSON")
	}
}

func TestRevertChangesCreatesInverseChanges(t *testing.T) {
	original := []Change{
		{Field: "Name", ChangeType: Modified, OldValue: "John", NewValue: "Jane"},
		{Field: "Age", ChangeType: Added, NewValue: 31},
		{Field: "Active", ChangeType: Deleted, OldValue: true},
	}

	reverted := RevertChanges(original)

	if len(reverted) != len(original) {
		t.Errorf("Expected same number of changes in reversion")
	}

	nameChange := findChangeByField(reverted, "Name")
	if nameChange == nil || nameChange.ChangeType != Modified ||
		nameChange.OldValue.(string) != "Jane" || nameChange.NewValue.(string) != "John" {
		t.Errorf("Modified change not reverted correctly")
	}

	ageChange := findChangeByField(reverted, "Age")
	if ageChange == nil || ageChange.ChangeType != Deleted || ageChange.OldValue != 31 {
		t.Errorf("Added change not reverted to Deleted correctly")
	}

	activeChange := findChangeByField(reverted, "Active")
	if activeChange == nil || activeChange.ChangeType != Added || activeChange.NewValue != true {
		t.Errorf("Deleted change not reverted to Added correctly")
	}
}
