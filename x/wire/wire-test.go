package wire

import (
	"encoding/hex"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestMessageRoundTrip(t *testing.T) {
	// Create sample message with protocol requirement values
	msg := wire.Message{
		Tag:      []byte{0x67, 0x72, 0x69, 0x64}, // "grid" in ASCII
		Protocol: []byte{0x01, 0x02, 0x03},
		Payload:  []byte{0x04, 0x05, 0x06},
	}

	// Encode to CBOR
	data, err := cbor.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Validate CBOR structure is array (major type 4)
	if (data[0] >> 5) != 4 {
		t.Fatalf("CBOR data is not an array, got major type %d", data[0]>>5)
	}

	// Expected hex: 8344677269644301020343040506
	// Breakdown:
	// 83 - array(3)
	//   44 - bytes(4) "grid"
	//   43 - bytes(3) protocol
	//   43 - bytes(3) payload
	expected := "8344677269644301020343040506"
	if hexStr := hex.EncodeToString(data); hexStr != expected {
		t.Fatalf("Unexpected CBOR encoding:\nGot:  %s\nWant: %s", hexStr, expected)
	}

	// Decode and verify
	var decoded wire.Message
	if err := cbor.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if string(decoded.Tag) != "grid" {
		t.Errorf("Tag mismatch:\nGot:  % x\nWant: 67 72 69 64", decoded.Tag)
	}

	if string(decoded.Protocol) != string(msg.Protocol) {
		t.Errorf("Protocol mismatch:\nGot:  % x\nWant: % x", decoded.Protocol, msg.Protocol)
	}

	if string(decoded.Payload) != string(msg.Payload) {
		t.Errorf("Payload mismatch:\nGot:  % x\nWant: % x", decoded.Payload, msg.Payload)
	}
}

func TestEmptyFields(t *testing.T) {
	msg := wire.Message{
		Tag:      []byte{},
		Protocol: []byte{},
		Payload:  []byte{},
	}

	data, err := cbor.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Expected empty array: 83404040
	// 83 - array(3)
	//   40 - bytes(0)
	//   40 - bytes(0)
	//   40 - bytes(0)
	expected := "83404040"
	if hexStr := hex.EncodeToString(data); hexStr != expected {
		t.Fatalf("Unexpected empty encoding:\nGot:  %s\nWant: %s", hexStr, expected)
	}
}

func TestInvalidArray(t *testing.T) {
	// Invalid array (wrong element count)
	invalidData, _ := hex.DecodeString("820440") // array(2), bytes(0)
	var msg wire.Message
	if err := cbor.Unmarshal(invalidData, &msg); err == nil {
		t.Fatal("Expected error for wrong array length, got nil")
	}
}
