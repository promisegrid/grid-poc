package cbordecode

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// Message is an example type used to demonstrate CBOR decoding.
type Message struct {
	Value int
}

// encodeWithTag encodes the given value v into CBOR format and wraps it in a proper CBOR tag,
// where tagNum is the CBOR tag number. It first encodes v, then creates a cbor.RawTag whose Content
// is the encoding of v, and finally marshals the RawTag.
func encodeWithTag(tagNum uint64, v interface{}) ([]byte, error) {
	encodedContent, err := encMode.Marshal(v)
	if err != nil {
		return nil, err
	}
	tagged := cbor.RawTag{
		Number:  tagNum,
		Content: encodedContent,
	}
	return encMode.Marshal(tagged)
}

// TestDecodeInto demonstrates the use of DecodeInto where the caller supplies an existing instance.
func TestDecodeInto(t *testing.T) {
	// Example 1: decoding into a Message.
	originalMsg := Message{Value: 100}
	// Use tag 258 for Message.
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
	// Use tag 259 for int.
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

// TestDecodeNew demonstrates the use of DecodeNew which returns a pointer to a newly allocated instance.
func TestDecodeNew(t *testing.T) {
	// Example 1: decoding into a new Message instance.
	originalMsg := Message{Value: 150}
	// Use tag 260 for Message.
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
	// Use tag 261 for int.
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

// TestDecodeValue demonstrates the use of DecodeValue which returns a conventional Go value.
func TestDecodeValue(t *testing.T) {
	// Example 1: decoding into a Message value.
	originalMsg := Message{Value: 175}
	// Use tag 262 for Message.
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
	// Use tag 263 for int.
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

// TestDecodeUnknown demonstrates decoding when the caller does not know the type at compile time.
// In this example, we use the DecodeValue function with T instantiated as interface{}.
func TestDecodeUnknown(t *testing.T) {
	// Example: decoding an int into a dynamic interface{}
	originalInt := 300
	// Use tag 264 for this test.
	dataInt, err := encodeWithTag(264, originalInt)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	decoded, tag, err := DecodeValue[interface{}](dataInt)
	if err != nil {
		t.Fatalf("DecodeValue[interface{}] failed: %v", err)
	}
	if tag != 264 {
		t.Errorf("expected tag 264, got %d", tag)
	}
	// fxamacker/cbor decodes numeric values into uint64 when using interface{}.
	intResult, ok := decoded.(uint64)
	if !ok {
		t.Errorf("expected uint64 type, got %T", decoded)
	}
	if intResult != uint64(originalInt) {
		t.Errorf("expected uint64 value %d, got %d", originalInt, intResult)
	}
}

// TestDecodeFromReader demonstrates the use of decMode.NewDecoder to decode from an io.Reader.
// It encodes a Message with a specific tag and then decodes it using a bytes.Reader.
func TestDecodeFromReader(t *testing.T) {
	originalMsg := Message{Value: 999}
	// Use tag 265 for this test.
	data, err := encodeWithTag(265, originalMsg)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	reader := bytes.NewReader(data)
	decoder := decMode.NewDecoder(reader)
	
	var rawTag cbor.RawTag
	if err := decoder.Decode(&rawTag); err != nil {
		t.Fatalf("decoding raw tag from io.Reader failed: %v", err)
	}
	if int(rawTag.Number) != 265 {
		t.Errorf("expected tag 265, got %d", rawTag.Number)
	}

	var msg Message
	if err := decMode.Unmarshal(rawTag.Content, &msg); err != nil {
		t.Fatalf("decoding Message content failed: %v", err)
	}
	if msg != originalMsg {
		t.Errorf("expected Message %+v, got %+v", originalMsg, msg)
	}
}
