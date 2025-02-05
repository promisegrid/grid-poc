package cbordecode

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// Message is an example type used to demonstrate CBOR decoding.
type Message struct {
	Value int
}

// encodeWithTag encodes the given value v into CBOR format and wraps it in a proper CBOR tag,
// where tagNum is the CBOR tag number. It first encodes v, then creates a cbor.RawTag
// whose Content is the encoding of v, and finally marshals the RawTag.
// Note: To avoid conflicting with reserved tag numbers in the CBOR specification,
// the test cases use non-reserved tag numbers (e.g., 258, 259, etc.).
func encodeWithTag(tagNum uint64, v interface{}) ([]byte, error) {
	encodedContent, err := cbor.Marshal(v)
	if err != nil {
		return nil, err
	}
	tagged := cbor.RawTag{
		Number:  tagNum,
		Content: encodedContent,
	}
	return cbor.Marshal(tagged)
}

// TestDecodeInto demonstrates the use of DecodeInto where the caller supplies an existing instance
// for decoding into. The encoded data uses a proper CBOR tag.
func TestDecodeInto(t *testing.T) {
	// Example 1: decoding into a Message.
	originalMsg := Message{Value: 100}
	// Use tag 258 (non-reserved) for Message.
	dataMsg, err := encodeWithTag(258, originalMsg)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	var msg Message
	tag, err := DecodeInto(dataMsg, &msg)
	if err != nil {
		t.Fatalf("DecodeInto failed: %v", err)
	}
	if tag != 258 {
		t.Errorf("expected tag 258, got %d", tag)
	}
	if msg != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, msg)
	}

	// Example 2: decoding into an int.
	originalInt := 200
	// Use tag 259 (non-reserved) for int.
	dataInt, err := encodeWithTag(259, originalInt)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	var number int
	tag, err = DecodeInto(dataInt, &number)
	if err != nil {
		t.Fatalf("DecodeInto failed: %v", err)
	}
	if tag != 259 {
		t.Errorf("expected tag 259, got %d", tag)
	}
	if number != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, number)
	}
}

// TestDecodeNew demonstrates the use of DecodeNew which returns a pointer to a newly allocated instance
// of the decoded type. The encoded data uses a proper CBOR tag.
func TestDecodeNew(t *testing.T) {
	// Example 1: decoding into a new Message instance.
	originalMsg := Message{Value: 150}
	// Use tag 260 (non-reserved) for Message.
	dataMsg, err := encodeWithTag(260, originalMsg)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	msgPtr, tag, err := DecodeNew[Message](dataMsg)
	if err != nil {
		t.Fatalf("DecodeNew[Message] failed: %v", err)
	}
	if tag != 260 {
		t.Errorf("expected tag 260, got %d", tag)
	}
	if *msgPtr != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, *msgPtr)
	}

	// Example 2: decoding into a new int instance.
	originalInt := 250
	// Use tag 261 (non-reserved) for int.
	dataInt, err := encodeWithTag(261, originalInt)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	intPtr, tag, err := DecodeNew[int](dataInt)
	if err != nil {
		t.Fatalf("DecodeNew[int] failed: %v", err)
	}
	if tag != 261 {
		t.Errorf("expected tag 261, got %d", tag)
	}
	if *intPtr != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, *intPtr)
	}
}

// TestDecodeValue demonstrates the use of DecodeValue which returns a conventional Go value rather than a pointer.
// The encoded data uses a proper CBOR tag.
func TestDecodeValue(t *testing.T) {
	// Example 1: decoding into a Message value.
	originalMsg := Message{Value: 175}
	// Use tag 262 (non-reserved) for Message.
	dataMsg, err := encodeWithTag(262, originalMsg)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	msg, tag, err := DecodeValue[Message](dataMsg)
	if err != nil {
		t.Fatalf("DecodeValue[Message] failed: %v", err)
	}
	if tag != 262 {
		t.Errorf("expected tag 262, got %d", tag)
	}
	if msg != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, msg)
	}

	// Example 2: decoding into an int value.
	originalInt := 225
	// Use tag 263 (non-reserved) for int.
	dataInt, err := encodeWithTag(263, originalInt)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	number, tag, err := DecodeValue[int](dataInt)
	if err != nil {
		t.Fatalf("DecodeValue[int] failed: %v", err)
	}
	if tag != 263 {
		t.Errorf("expected tag 263, got %d", tag)
	}
	if number != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, number)
	}
}
