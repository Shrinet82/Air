package main

import (
	"testing"
)

// Simple test for Buffer creation
func TestNewBuffer(t *testing.T) {
	// Test creating a new empty buffer
	buffer, err := NewBuffer("")
	if err != nil {
		t.Fatalf("Expected no error for empty buffer, got: %v", err)
	}
	if len(buffer.Lines) != 1 || buffer.Lines[0] != "" {
		t.Errorf("Expected empty buffer to have one empty line, got: %v", buffer.Lines)
	}
	if buffer.Dirty {
		t.Error("New buffer should not be marked as dirty")
	}
	
	// Test BaseName function
	if buffer.BaseName() != "[No Name]" {
		t.Errorf("Expected unnamed buffer to have BaseName '[No Name]', got: %s", buffer.BaseName())
	}
}
