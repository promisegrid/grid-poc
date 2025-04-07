package wire

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestRoundTrip(t *testing.T) {
	// Test vectors
	tag := []byte{0x67, 0x72, 0x69, 0x64} // 'grid' in ASCII
	protocol := []byte{0x01, 0x02, 0x03}
	payload := []byte{0x04, 0x05, 0x06}

	// Create original message
	orig := Message{
		Tag:      tag,
		Protocol: protocol,
		Payload:  payload,
	}

	// Marshal to CBOR
	data, err := orig.MarshalCBOR()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal from CBOR
	var decoded Message
	if err := decoded.UnmarshalCBOR(data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify decoded fields
	if !bytes.Equal(decoded.Tag, tag) {
		t.Errorf("Tag mismatch: got %x, want %x", decoded.Tag, tag)
	}
	if !bytes.Equal(decoded.Protocol, protocol) {
		t.Errorf("Protocol mismatch: got %x, want %x", decoded.Protocol, protocol)
	}
	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload mismatch: got %x, want %x", decoded.Payload, payload)
	}
}

func TestEmptyFields(t *testing.T) {
	// Test empty message
	orig := Message{
		Tag:      []byte{},
		Protocol: []byte{},
		Payload:  []byte{},
	}

	data, err := orig.MarshalCBOR()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify CBOR structure by decoding to raw array
	var arr []cbor.RawMessage
	if err := cbor.Unmarshal(data, &arr); err != nil {
		t.Fatalf("Raw unmarshal failed: %v", err)
	}

	if len(arr) != 3 {
		t.Fatalf("Unexpected array length: got %d, want 3", len(arr))
	}

	// Check all elements are empty byte strings
	for i, el := range arr {
		var b []byte
		if err := cbor.Unmarshal(el, &b); err != nil {
			t.Errorf("Element %d not byte string: %v", i, err)
		}
		if len(b) != 0 {
			t.Errorf("Element %d not empty: len=%d", i, len(b))
		}
	}
}
