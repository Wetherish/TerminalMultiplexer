package session

import (
	"bytes"
	"testing"
)

func TestRingBuffer_Write(t *testing.T) {
	rb := NewRingBuffer(10)

	rb.Write([]byte("12345"))
	if string(rb.Bytes()) != "12345" {
		t.Errorf("Expected '12345', got '%s'", string(rb.Bytes()))
	}

	rb.Write([]byte("67890"))
	if string(rb.Bytes()) != "1234567890" {
		t.Errorf("Expected '1234567890', got '%s'", string(rb.Bytes()))
	}

	// Test overwriting
	rb.Write([]byte("A"))
	if string(rb.Bytes()) != "234567890A" {
		t.Errorf("Expected '234567890A', got '%s'", string(rb.Bytes()))
	}

	// Test overwriting multiple
	rb.Write([]byte("BC"))
	if string(rb.Bytes()) != "4567890ABC" {
		t.Errorf("Expected '4567890ABC', got '%s'", string(rb.Bytes()))
	}

	// Test writing more than size
	rb.Write([]byte("123456789012345"))
	if string(rb.Bytes()) != "6789012345" {
		t.Errorf("Expected '6789012345', got '%s'", string(rb.Bytes()))
	}
}

func TestRingBuffer_Bytes(t *testing.T) {
	rb := NewRingBuffer(5)
	rb.Write([]byte("123"))

	b := rb.Bytes()
	if !bytes.Equal(b, []byte("123")) {
		t.Errorf("Bytes() returned wrong data")
	}

	// Ensure modifying returned slice doesn't affect buffer
	b[0] = 'X'
	if string(rb.Bytes()) != "123" {
		t.Errorf("Bytes() return value should be a copy")
	}
}
