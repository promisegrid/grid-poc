package cbordecode

import (
	"testing"
)

// TestDecodeInto demonstrates the use of DecodeInto where the caller
// supplies an existing instance (or pointer) to be decoded into.
func TestDecodeInto(t *testing.T) {
	// Example 1: decoding into a Message
	dataMsg := []byte{1, 100} // tag = 1, value = 100
	var msg Message
	tag, err := DecodeInto(dataMsg, &msg)
	if err != nil {
		t.Fatalf("DecodeInto failed: %v", err)
	}
	if tag != 1 {
		t.Errorf("expected tag 1, got %d", tag)
	}
	if msg.Value != 100 {
		t.Errorf("expected Message.Value 100, got %d", msg.Value)
	}

	// Example 2: decoding into an int
	dataInt := []byte{2, 200} // tag = 2, value = 200
	var number int
	tag, err = DecodeInto(dataInt, &number)
	if err != nil {
		t.Fatalf("DecodeInto failed: %v", err)
	}
	if tag != 2 {
		t.Errorf("expected tag 2, got %d", tag)
	}
	if number != 200 {
		t.Errorf("expected int value 200, got %d", number)
	}
}

// TestDecodeNew demonstrates the use of DecodeNew which returns a pointer
// to a newly allocated instance of the decoded type.
func TestDecodeNew(t *testing.T) {
	// Example 1: decoding into a new Message instance
	dataMsg := []byte{3, 150} // tag = 3, value = 150
	msgPtr, tag, err := DecodeNew[Message](dataMsg)
	if err != nil {
		t.Fatalf("DecodeNew[Message] failed: %v", err)
	}
	if tag != 3 {
		t.Errorf("expected tag 3, got %d", tag)
	}
	if msgPtr.Value != 150 {
		t.Errorf("expected Message.Value 150, got %d", msgPtr.Value)
	}

	// Example 2: decoding into a new int instance
	dataInt := []byte{4, 250} // tag = 4, value = 250
	intPtr, tag, err := DecodeNew[int](dataInt)
	if err != nil {
		t.Fatalf("DecodeNew[int] failed: %v", err)
	}
	if tag != 4 {
		t.Errorf("expected tag 4, got %d", tag)
	}
	if *intPtr != 250 {
		t.Errorf("expected int value 250, got %d", *intPtr)
	}
}

// TestDecodeValue demonstrates the use of DecodeValue which returns a conventional
// Go value rather than a pointer.
func TestDecodeValue(t *testing.T) {
	// Example 1: decoding into a Message value
	dataMsg := []byte{5, 175} // tag = 5, value = 175
	msg, tag, err := DecodeValue[Message](dataMsg)
	if err != nil {
		t.Fatalf("DecodeValue[Message] failed: %v", err)
	}
	if tag != 5 {
		t.Errorf("expected tag 5, got %d", tag)
	}
	if msg.Value != 175 {
		t.Errorf("expected Message.Value 175, got %d", msg.Value)
	}

	// Example 2: decoding into an int value
	dataInt := []byte{6, 225} // tag = 6, value = 225
	number, tag, err := DecodeValue[int](dataInt)
	if err != nil {
		t.Fatalf("DecodeValue[int] failed: %v", err)
	}
	if tag != 6 {
		t.Errorf("expected tag 6, got %d", tag)
	}
	if number != 225 {
		t.Errorf("expected int value 225, got %d", number)
	}
}
