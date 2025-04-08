package wire

import (
	"bytes"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Message represents a protocol message with protocol CID and payload.
type Message struct {
	Protocol []byte
	Payload  []byte
}

var (
	gridTag = []byte{0x67, 0x72, 0x69, 0x64} // "grid"
	encOpts cbor.EncOptions
	decOpts cbor.DecOptions
	em      cbor.EncMode
	dm      cbor.DecMode
)

func init() {
	var err error
	// Use Core Deterministic Encoding for consistent serialization
	encOpts = cbor.CoreDetEncOptions()
	decOpts = cbor.DecOptions{}

	em, err = encOpts.EncMode()
	if err != nil {
		panic(fmt.Sprintf("failed to create CBOR enc mode: %v", err))
	}

	dm, err = decOpts.DecMode()
	if err != nil {
		panic(fmt.Sprintf("failed to create CBOR dec mode: %v", err))
	}
}

func (m Message) MarshalCBOR() ([]byte, error) {
	return em.Marshal([]interface{}{gridTag, m.Protocol, m.Payload})
}

func (m *Message) UnmarshalCBOR(data []byte) error {
	var parts []interface{}
	if err := dm.Unmarshal(data, &parts); err != nil {
		return err
	}

	if len(parts) != 3 {
		return fmt.Errorf("invalid array length: expected 3, got %d", len(parts))
	}

	tag, ok := parts[0].([]byte)
	if !ok || !bytes.Equal(tag, gridTag) {
		return fmt.Errorf("invalid grid tag: %v", parts[0])
	}

	switch v := parts[1].(type) {
	case []byte:
		m.Protocol = v
	case nil:
		m.Protocol = nil
	default:
		return fmt.Errorf("protocol field has unexpected type: %T", parts[1])
	}

	switch v := parts[2].(type) {
	case []byte:
		m.Payload = v
	case nil:
		m.Payload = nil
	default:
		return fmt.Errorf("payload field has unexpected type: %T", parts[2])
	}

	return nil
}
