package wire

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestRoundTrip(t *testing.T) {
	protocol := []byte{0x01, 0x02, 0x03}
	payload := []byte{0x04, 0x05, 0x06}

	orig := Message{
		Protocol: protocol,
		Payload:  payload,
	}

	data, err := cbor.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expectedPrefix := []byte{
		0x83,       // array(3)
		0x44, 0x67, 0x72, 0x69, 0x64,  // grid tag bytes
		0x43, 0x01, 0x02, 0x03,        // protocol bytes
		0x43, 0x04, 0x05, 0x06,        // payload bytes
	}
	if !bytes.HasPrefix(data, expectedPrefix) {
		t.Errorf("Invalid CBOR structure\nGot:  %x\nWant prefix: %x", data, expectedPrefix)
	}

	var decoded Message
	if err := cbor.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal(decoded.Protocol, protocol) {
		t.Errorf("Protocol mismatch\nGot:  %x\nWant: %x", decoded.Protocol, protocol)
	}
	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload mismatch\nGot:  %x\nWant: %x", decoded.Payload, payload)
	}
}

func TestEmptyMessage(t *testing.T) {
	testCases := []struct {
		name     string
		message  Message
		expected []byte
	}{
		{
			"empty fields",
			Message{Protocol: []byte{}, Payload: []byte{}},
			[]byte{
				0x83,                   // array(3)
				0x44, 0x67, 0x72, 0x69, 0x64, // grid tag
				0x40, // empty protocol bytes
				0x40, // empty payload bytes
			},
		},
		{
			"nil fields",
			Message{Protocol: nil, Payload: nil},
			[]byte{
				0x83,                   // array(3)
				0x44, 0x67, 0x72, 0x69, 0x64, // grid tag
				0xF6, // nil protocol
				0xF6, // nil payload
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := cbor.Marshal(tc.message)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			if !bytes.Equal(data, tc.expected) {
				t.Errorf("Encoding mismatch\nGot:  %x\nWant: %x", data, tc.expected)
			}

			var decoded Message
			if err := cbor.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if tc.message.Protocol == nil && decoded.Protocol != nil {
				t.Errorf("Expected nil protocol, got %x", decoded.Protocol)
			}
			if tc.message.Payload == nil && decoded.Payload != nil {
				t.Errorf("Expected nil payload, got %x", decoded.Payload)
			}
		})
	}
}
