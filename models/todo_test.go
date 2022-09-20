package models

import "testing"

func TestTodo_Serialize(t *testing.T) {
	// Arrange
	//
	todoTest := Todo{Id: "99", Title: "Test1", Description: "Beschrieb", Terminated: false}
	var want []string = []string{"99", "Test1", "Beschrieb", "false"}

	// Act
	//
	got := todoTest.Serialize()

	// Assert
	//
	if areStringSlicesEqual(got, want) == false {
		t.Error("Fehler")
	}
}

func TestTodo_AddTodo(t *testing.T) {
	// Arrange
	//
	todoTest := Todo{Id: "0", Title: "Test1", Description: "Beschrieb", Terminated: false}
	var want Todo = todoTest

	// Act
	//
	got := AddTodo(todoTest)

	// Assert
	//
	if got != want {
		t.Error("Fehler")
	}
}

// areStringSlicesEqual tells whether a and b contain the same elements.
func areStringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
