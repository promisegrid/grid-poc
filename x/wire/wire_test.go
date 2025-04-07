package wire

import (
	"bytes"
	"testing"
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
	data, err := orig.MarshalBinary()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal from CBOR
	var decoded Message
	if err := decoded.UnmarshalBinary(data); err != nil {
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

	data, err := orig.MarshalBinary()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify CBOR structure by decoding to array
	var decoded Message
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Tag) != 0 || len(decoded.Protocol) != 0 || len(decoded.Payload) != 0 {
		t.Errorf("Empty fields not preserved: %+v", decoded)
	}
}
