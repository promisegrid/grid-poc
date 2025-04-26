package wire

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func TestNewMessageRoundTrip(t *testing.T) {
	// Create CIDv1 with raw codec
	mh, _ := multihash.Sum([]byte("test"), multihash.SHA2_256, -1)
	c := cid.NewCidV1(cid.Raw, mh)
	payload := []byte{0x01, 0x02, 0x03}

	// Create and encode message
	encoded, err := NewMessage(c, payload)
	if err != nil {
		t.Fatalf("NewMessage failed: %v", err)
	}

	// Verify CBOR structure
	expectedPrefix := []byte{
		0xDA,             // Tag(4 bytes)
		0x67, 0x72, 0x69, 0x64, // Tag number 0x67726964 ('grid')
		0x82, // Array(2)
	}
	if !bytes.HasPrefix(encoded, expectedPrefix) {
		t.Errorf("Invalid CBOR structure\nGot:  %x\nWant prefix: %x", encoded, expectedPrefix)
	}

	// Decode message
	var decoded Message
	if err := cbor.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify CID roundtrip
	decodedCID, err := cid.Cast(decoded.Protocol)
	if err != nil {
		t.Fatalf("CID cast failed: %v", err)
	}
	if !decodedCID.Equals(c) {
		t.Errorf("CID mismatch\nGot:  %s\nWant: %s", decodedCID, c)
	}

	// Verify payload
	if !bytes.Equal(decoded.Payload, payload) {
		t.Errorf("Payload mismatch\nGot:  %x\nWant: %x", decoded.Payload, payload)
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty payload", func(t *testing.T) {
		c, _ := cid.Decode("bafkreihdwdcefgh4dqkjv67uzcmw7ojee6xedzdetojuzjevtenxquvyku")
		encoded, err := NewMessage(c, []byte{})
		if err != nil {
			t.Fatal(err)
		}

		var decoded Message
		if err := cbor.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}

		if len(decoded.Payload) != 0 {
			t.Errorf("Expected empty payload, got %x", decoded.Payload)
		}
	})

	t.Run("nil payload", func(t *testing.T) {
		c, _ := cid.Decode("bafkreihdwdcefgh4dqkjv67uzcmw7ojee6xedzdetojuzjevtenxquvyku")
		encoded, err := NewMessage(c, nil)
		if err != nil {
			t.Fatal(err)
		}

		var decoded Message
		if err := cbor.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}

		if decoded.Payload != nil {
			t.Errorf("Expected nil payload, got %x", decoded.Payload)
		}
	})
}

func TestInvalidMessages(t *testing.T) {
	t.Run("invalid tag number", func(t *testing.T) {
		data := []byte{
			0xDA, // Tag(4 bytes)
			0x00, 0x00, 0x00, 0x00, // Invalid tag number
			0x82, // Array(2)
			0x40, // Empty bytes
			0x40, // Empty bytes
		}
		var m Message
		err := cbor.Unmarshal(data, &m)
		if err == nil {
			t.Error("Expected error for invalid tag number")
		}
	})

	t.Run("insufficient array elements", func(t *testing.T) {
		data := []byte{
			0xDA, 0x67, 0x72, 0x69, 0x64, // Valid grid tag
			0x81, // Array(1) instead of 2
			0x40, // Empty bytes
		}
		var m Message
		err := cbor.Unmarshal(data, &m)
		if err == nil {
			t.Error("Expected error for insufficient array elements")
		}
	})

	t.Run("non-array content", func(t *testing.T) {
		data := []byte{
			0xDA, 0x67, 0x72, 0x69, 0x64, // Valid grid tag
			0xA0, // Empty map instead of array
		}
		var m Message
		err := cbor.Unmarshal(data, &m)
		if err == nil {
			t.Error("Expected error for non-array content")
		}
	})
}
