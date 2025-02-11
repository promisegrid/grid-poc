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

// TestDecodeRaw demonstrates the use of DecodeRaw which decodes a single-level CBOR value.
// It first extracts the raw content via DecodeTag, then decodes that content.
func TestDecodeRaw(t *testing.T) {
	// Example 1: decoding into a Message value.
	originalMsg := Message{Value: 175}
	// Use tag 262 for Message.
	dataMsg, err := encodeWithTag(262, originalMsg)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	tag, content, err := DecodeTag(dataMsg)
	if err != nil {
		t.Fatalf("DecodeTag failed: %v", err)
	}
	if tag != 262 {
		t.Errorf("expected tag 262, got %d", tag)
	}
	var msg Message
	err = DecodeRaw(content, &msg)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
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
	tag, content, err = DecodeTag(dataInt)
	if err != nil {
		t.Fatalf("DecodeTag failed: %v", err)
	}
	if tag != 263 {
		t.Errorf("expected tag 263, got %d", tag)
	}
	var number int
	err = DecodeRaw(content, &number)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
	}
	if number != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, number)
	}
}

// TestDecodeUnknown demonstrates decoding when the caller does not know the type at compile time.
// Here, the raw CBOR content is decoded into an interface{}.
func TestDecodeUnknown(t *testing.T) {
	// Example: decoding an int into a dynamic interface{}
	originalInt := 300
	// Use tag 264 for this test.
	dataInt, err := encodeWithTag(264, originalInt)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}
	_, content, err := DecodeTag(dataInt)
	if err != nil {
		t.Fatalf("DecodeTag failed: %v", err)
	}
	var decoded interface{}
	err = DecodeRaw(content, &decoded)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
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

// TestDecodeTag demonstrates the use of DecodeTag which decodes just the tag and its content.
func TestDecodeTag(t *testing.T) {
	originalInt := 1234
	// Use tag 270 for this test.
	dataInt, err := encodeWithTag(270, originalInt)
	if err != nil {
		t.Fatalf("encodeWithTag failed: %v", err)
	}

	tag, content, err := DecodeTag(dataInt)
	if err != nil {
		t.Fatalf("DecodeTag failed: %v", err)
	}
	if tag != 270 {
		t.Errorf("expected tag 270, got %d", tag)
	}

	// Now decode the content manually.
	var decodedInt int
	if err := decMode.Unmarshal(content, &decodedInt); err != nil {
		t.Fatalf("failed to unmarshal content from DecodeTag: %v", err)
	}
	if decodedInt != originalInt {
		t.Errorf("expected int value %d, got %d", originalInt, decodedInt)
	}

	// Test with data that does not start with a tag.
	nonTagData, err := encMode.Marshal(originalInt)
	if err != nil {
		t.Fatalf("failed to marshal non-tag data: %v", err)
	}
	tag, content, err = DecodeTag(nonTagData)
	if err == nil {
		t.Fatalf("DecodeTag with non-tag data did not return an error")
	}
	if tag != 0 || content != nil {
		t.Errorf("expected 0, nil for non-tag data, got %d, %v", tag, content)
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

// TestDecodeVarious demonstrates the use of DecodeTag and DecodeRaw with multiple tagged values.
func TestDecodeVarious(t *testing.T) {
	// Define another struct for testing.
	type Other struct {
		Name string
	}

	// Prepare a list of test cases with different tag numbers and corresponding values.
	type testCase struct {
		tag   uint64
		value interface{}
	}
	testCases := []testCase{
		{tag: 300, value: Message{Value: 111}},
		{tag: 301, value: 222},
		{tag: 302, value: Other{Name: "Test"}},
	}

	// Encode each test case.
	var encodedData [][]byte
	for _, tc := range testCases {
		data, err := encodeWithTag(tc.tag, tc.value)
		if err != nil {
			t.Fatalf("failed to encode value with tag %d: %v", tc.tag, err)
		}
		encodedData = append(encodedData, data)
	}

	// Iterate over the encoded data and decode based on tag.
	for i, data := range encodedData {
		tag, content, err := DecodeTag(data)
		if err != nil {
			t.Fatalf("DecodeTag failed for test case %d: %v", i, err)
		}
		switch tag {
		case 300:
			var msg Message
			if err := DecodeRaw(content, &msg); err != nil {
				t.Errorf("DecodeRaw failed for tag 300: %v", err)
				continue
			}
			expected := testCases[i].value.(Message)
			if msg != expected {
				t.Errorf("for tag 300, expected Message %+v, got %+v", expected, msg)
			}
		case 301:
			var number int
			if err := DecodeRaw(content, &number); err != nil {
				t.Errorf("DecodeRaw failed for tag 301: %v", err)
				continue
			}
			expected := testCases[i].value.(int)
			if number != expected {
				t.Errorf("for tag 301, expected int %d, got %d", expected, number)
			}
		case 302:
			var other struct {
				Name string
			}
			if err := DecodeRaw(content, &other); err != nil {
				t.Errorf("DecodeRaw failed for tag 302: %v", err)
				continue
			}
			expected := testCases[i].value.(Other)
			if other.Name != expected.Name {
				t.Errorf("for tag 302, expected Other %+v, got %+v", expected, other)
			}
		default:
			t.Errorf("unexpected tag %d in test case %d", tag, i)
		}
	}
}
