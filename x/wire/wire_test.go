package wire

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestRoundTrip(t *testing.T) {
	tag := []byte{0x67, 0x72, 0x69, 0x64} // 'grid' tag
	protocol := []byte{0x01, 0x02, 0x03}
	payload := []byte{0x04, 0x05, 0x06}

	orig := Message{
		Tag:      tag,
		Protocol: protocol,
		Payload:  payload,
	}

	data, err := cbor.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify CBOR array structure (0x83 = array(3), 0x44 = bytes(4), etc.)
	expectedPrefix := []byte{
		0x83,       // array(3)
		0x44, 0x67, 0x72, 0x69, 0x64,  // bytes(4) "grid"
		0x43, 0x01, 0x02, 0x03,        // bytes(3) protocol
		0x43, 0x04, 0x05, 0x06,        // bytes(3) payload
	}
	if !bytes.HasPrefix(data, expectedPrefix) {
		t.Errorf("Invalid CBOR structure\nGot:  %x\nWant prefix: %x", data, expectedPrefix)
	}

	var decoded Message
	if err := cbor.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal(decoded.Tag, tag) {
		t.Errorf("Tag mismatch\nGot:  %x\nWant: %x", decoded.Tag, tag)
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
			Message{Tag: []byte{}, Protocol: []byte{}, Payload: []byte{}},
			[]byte{0x83, 0x40, 0x40, 0x40}, // [bstr(), bstr(), bstr()]
		},
		{
			"nil fields",
			Message{Tag: nil, Protocol: nil, Payload: nil},
			[]byte{0x83, 0xF6, 0xF6, 0xF6}, // [null, null, null]
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
		})
	}
}
