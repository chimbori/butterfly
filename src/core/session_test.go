package core

import (
	"testing"
	"time"
)

func TestInMemorySessionStore_Create(t *testing.T) {
	store := NewInMemorySessionStore(time.Hour)
	id, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if id == "" {
		t.Error("Create() returned empty ID")
	}

	if !store.IsValid(id) {
		t.Errorf("Newly created session %s should be valid", id)
	}
}

func TestInMemorySessionStore_IsValid_NonExistent(t *testing.T) {
	store := NewInMemorySessionStore(time.Hour)
	if store.IsValid("non-existent-id") {
		t.Error("IsValid() returned true for non-existent ID")
	}
}

func TestInMemorySessionStore_IsValid_Expired(t *testing.T) {
	// Use a very short TTL
	store := NewInMemorySessionStore(10 * time.Millisecond)
	id, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	if store.IsValid(id) {
		t.Error("IsValid() returned true for expired session")
	}
}

func TestInMemorySessionStore_Delete(t *testing.T) {
	store := NewInMemorySessionStore(time.Hour)
	id, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	store.Delete(id)

	if store.IsValid(id) {
		t.Error("IsValid() returned true for deleted session")
	}
}

func TestInMemorySessionStore_Concurrency(t *testing.T) {
	store := NewInMemorySessionStore(time.Minute)

	// Run many goroutines concurrently creating and checking sessions
	concurrency := 100
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			id, err := store.Create()
			if err != nil {
				t.Errorf("Create() error = %v", err)
				done <- true
				return
			}

			if !store.IsValid(id) {
				t.Errorf("Session %s should be valid", id)
			}

			store.Delete(id)

			if store.IsValid(id) {
				t.Errorf("Session %s should be invalid after delete", id)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}
}
