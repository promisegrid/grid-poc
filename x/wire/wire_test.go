package wire

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
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
		0x83,                         // array(3)
		0x44, 0x67, 0x72, 0x69, 0x64, // 'grid' tag bytes
		0x43, 0x01, 0x02, 0x03, // subprotocol bytes
		0x43, 0x04, 0x05, 0x06, // payload bytes
	}
	if !bytes.HasPrefix(data, expectedPrefix) {
		t.Errorf("Invalid CBOR structure\nGot:  %x\nWant prefix: %x",
			data, expectedPrefix)
	}

	var decoded Message
	if err := cbor.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal(decoded.Protocol, protocol) {
		t.Errorf("Protocol mismatch\nGot:  %x\nWant: %x",
			decoded.Protocol, protocol)
	}
	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload mismatch\nGot:  %x\nWant: %x",
			decoded.Payload, payload)
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
				0x83,                         // array(3)
				0x44, 0x67, 0x72, 0x69, 0x64, // grid tag
				0x40, // empty protocol bytes
				0x40, // empty payload bytes
			},
		},
		{
			"nil fields",
			Message{Protocol: nil, Payload: nil},
			[]byte{
				0x83,                         // array(3)
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
				t.Errorf("Encoding mismatch\nGot:  %x\nWant: %x",
					data, tc.expected)
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

func TestMessageWithCID(t *testing.T) {
	// Create a CIDv1 with a SHA2-256 multihash.
	// We use multihash.Sum to generate the hash.
	mh, err := multihash.Sum([]byte("example data"), multihash.SHA2_256, -1)
	if err != nil {
		t.Fatalf("failed to create multihash: %v", err)
	}
	// Create a CIDv1 using the Raw codec.
	c := cid.NewCidV1(cid.Raw, mh)

	// Marshal the CID to CBOR using fxamacker/cbor.
	cborCID, err := cbor.Marshal(c)
	if err != nil {
		t.Fatalf("failed to CBOR encode CID: %v", err)
	}

	// Create the message with the full CBOR encoded CID as the
	// protocol identifier and "hello" as the payload.
	msg := Message{
		Protocol: cborCID,
		Payload:  []byte("hello"),
	}

	data, err := msg.MarshalCBOR()
	if err != nil {
		t.Fatalf("MarshalCBOR failed: %v", err)
	}

	var out Message
	if err := out.UnmarshalCBOR(data); err != nil {
		t.Fatalf("UnmarshalCBOR failed: %v", err)
	}

	// Verify that the payload matches.
	if !bytes.Equal(out.Payload, []byte("hello")) {
		t.Errorf("Payload mismatch: got %s, expected hello",
			string(out.Payload))
	}
	// Verify that the protocol (CID) matches the original.
	if !bytes.Equal(out.Protocol, cborCID) {
		t.Errorf("Protocol mismatch: got %x, expected %x", out.Protocol,
			cborCID)
	}

	// Further verify that the CBOR encoded protocol can be decoded back
	// to a valid CID.
	var rawCID []byte
	if err := cbor.Unmarshal(out.Protocol, &rawCID); err != nil {
		t.Errorf("failed to unmarshal protocol as raw bytes: %v", err)
	}
	decodedCID, err := cid.Cast(rawCID)
	if err != nil {
		t.Errorf("failed to cast raw bytes to CID: %v", err)
	}
	if decodedCID.Version() != 1 {
		t.Errorf("expected CIDv1, got CIDv%d", decodedCID.Version())
	}
	mhDecoded, err := multihash.Decode(decodedCID.Hash())
	if err != nil {
		t.Errorf("failed to decode multihash: %v", err)
	}
	if mhDecoded.Code != multihash.SHA2_256 {
		t.Errorf("expected SHA2_256 multihash, got code %d", mhDecoded.Code)
	}
}
