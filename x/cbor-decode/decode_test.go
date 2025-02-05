package cbordecode

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// Message is an example type to demonstrate CBOR decoding.
type Message struct {
	Value int
}

// Helper function to prepend a tag byte to CBOR-encoded data.
func prependTag(tag byte, encoded []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(tag)
	buf.Write(encoded)
	return buf.Bytes()
}

// TestDecodeInto demonstrates the use of DecodeInto where the caller
// supplies an existing instance (or pointer) to be decoded into.
func TestDecodeInto(t *testing.T) {
	// Example 1: decoding into a Message.
	originalMsg := Message{Value: 100}
	encodedMsg, err := cbor.Marshal(originalMsg)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}
	dataMsg := prependTag(1, encodedMsg)
	var msg Message
	tag, err := DecodeInto(dataMsg, &msg)
	if err != nil {
		t.Fatalf("DecodeInto failed: %v", err)
	}
	if tag != 1 {
		t.Errorf("expected tag 1, got %d", tag)
	}
	if msg != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, msg)
	}

	// Example 2: decoding into an int.
	originalInt := 200
	encodedInt, err := cbor.Marshal(originalInt)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}
	dataInt := prependTag(2, encodedInt)
	var number int
	tag, err = DecodeInto(dataInt, &number)
	if err != nil {
		t.Fatalf("DecodeInto failed: %v", err)
	}
	if tag != 2 {
		t.Errorf("expected tag 2, got %d", tag)
	}
	if number != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, number)
	}
}

// TestDecodeNew demonstrates the use of DecodeNew which returns a pointer
// to a newly allocated instance of the decoded type.
func TestDecodeNew(t *testing.T) {
	// Example 1: decoding into a new Message instance.
	originalMsg := Message{Value: 150}
	encodedMsg, err := cbor.Marshal(originalMsg)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}
	dataMsg := prependTag(3, encodedMsg)
	msgPtr, tag, err := DecodeNew[Message](dataMsg)
	if err != nil {
		t.Fatalf("DecodeNew[Message] failed: %v", err)
	}
	if tag != 3 {
		t.Errorf("expected tag 3, got %d", tag)
	}
	if *msgPtr != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, *msgPtr)
	}

	// Example 2: decoding into a new int instance.
	originalInt := 250
	encodedInt, err := cbor.Marshal(originalInt)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}
	dataInt := prependTag(4, encodedInt)
	intPtr, tag, err := DecodeNew[int](dataInt)
	if err != nil {
		t.Fatalf("DecodeNew[int] failed: %v", err)
	}
	if tag != 4 {
		t.Errorf("expected tag 4, got %d", tag)
	}
	if *intPtr != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, *intPtr)
	}
}

// TestDecodeValue demonstrates the use of DecodeValue which returns a conventional
// Go value rather than a pointer.
func TestDecodeValue(t *testing.T) {
	// Example 1: decoding into a Message value.
	originalMsg := Message{Value: 175}
	encodedMsg, err := cbor.Marshal(originalMsg)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}
	dataMsg := prependTag(5, encodedMsg)
	msg, tag, err := DecodeValue[Message](dataMsg)
	if err != nil {
		t.Fatalf("DecodeValue[Message] failed: %v", err)
	}
	if tag != 5 {
		t.Errorf("expected tag 5, got %d", tag)
	}
	if msg != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, msg)
	}

	// Example 2: decoding into an int value.
	originalInt := 225
	encodedInt, err := cbor.Marshal(originalInt)
	if err != nil {
		t.Fatalf("cbor.Marshal failed: %v", err)
	}
	dataInt := prependTag(6, encodedInt)
	number, tag, err := DecodeValue[int](dataInt)
	if err != nil {
		t.Fatalf("DecodeValue[int] failed: %v", err)
	}
	if tag != 6 {
		t.Errorf("expected tag 6, got %d", tag)
	}
	if number != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, number)
	}
}
