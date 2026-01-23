package core

import (
	"testing"
)

func TestPtr(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		value := 42
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if *ptr != value {
			t.Errorf("Expected *ptr to be %d, got %d", value, *ptr)
		}
		// Verify it’s a different address
		if ptr == &value {
			t.Error("Expected different memory address from original variable")
		}
	})

	t.Run("string", func(t *testing.T) {
		value := "hello"
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if *ptr != value {
			t.Errorf("Expected *ptr to be %q, got %q", value, *ptr)
		}
	})

	t.Run("bool", func(t *testing.T) {
		value := true
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if *ptr != value {
			t.Errorf("Expected *ptr to be %v, got %v", value, *ptr)
		}
	})

	t.Run("float64", func(t *testing.T) {
		value := 3.14159
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if *ptr != value {
			t.Errorf("Expected *ptr to be %f, got %f", value, *ptr)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		value := TestStruct{Name: "Alice", Age: 30}
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if ptr.Name != value.Name || ptr.Age != value.Age {
			t.Errorf("Expected *ptr to be %+v, got %+v", value, *ptr)
		}
	})

	t.Run("slice", func(t *testing.T) {
		value := []int{1, 2, 3}
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if len(*ptr) != len(value) {
			t.Errorf("Expected slice length %d, got %d", len(value), len(*ptr))
		}
		for i := range value {
			if (*ptr)[i] != value[i] {
				t.Errorf("Expected (*ptr)[%d] to be %d, got %d", i, value[i], (*ptr)[i])
			}
		}
	})

	t.Run("map", func(t *testing.T) {
		value := map[string]int{"a": 1, "b": 2}
		ptr := Ptr(value)
		if ptr == nil {
			t.Fatal("Expected non-nil pointer")
		}
		if len(*ptr) != len(value) {
			t.Errorf("Expected map length %d, got %d", len(value), len(*ptr))
		}
		for k, v := range value {
			if (*ptr)[k] != v {
				t.Errorf("Expected (*ptr)[%s] to be %d, got %d", k, v, (*ptr)[k])
			}
		}
	})

	t.Run("zero values", func(t *testing.T) {
		intPtr := Ptr(0)
		if *intPtr != 0 {
			t.Errorf("Expected *intPtr to be 0, got %d", *intPtr)
		}

		strPtr := Ptr("")
		if *strPtr != "" {
			t.Errorf("Expected *strPtr to be empty string, got %q", *strPtr)
		}

		boolPtr := Ptr(false)
		if *boolPtr != false {
			t.Errorf("Expected *boolPtr to be false, got %v", *boolPtr)
		}
	})

	t.Run("pointer independence", func(t *testing.T) {
		// Verify that modifying the original doesn’t affect the pointer
		value := 10
		ptr := Ptr(value)
		originalValue := *ptr

		value = 20 // Modify original

		if *ptr != originalValue {
			t.Errorf("Expected *ptr to remain %d after modifying source, got %d", originalValue, *ptr)
		}
	})

	t.Run("multiple calls", func(t *testing.T) {
		// Verify each call creates a new pointer
		value := 100
		ptr1 := Ptr(value)
		ptr2 := Ptr(value)

		if ptr1 == ptr2 {
			t.Error("Expected different pointers from multiple Ptr calls")
		}

		if *ptr1 != *ptr2 {
			t.Errorf("Expected same value, got %d and %d", *ptr1, *ptr2)
		}
	})
}
