package wire

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
)

// Message represents a protocol message with protocol CID and payload.
type Message struct {
	Protocol []byte
	Payload  []byte
}

var (
	gridTagNum uint64 = 0x67726964 // 'grid' as 4-byte big-endian integer
	encOpts    cbor.EncOptions
	decOpts    cbor.DecOptions
	Em         cbor.EncMode
	Dm         cbor.DecMode
)

func init() {
	var err error
	// Use Core Deterministic Encoding for consistent serialization
	encOpts = cbor.CoreDetEncOptions()
	decOpts = cbor.DecOptions{}

	Em, err = encOpts.EncMode()
	if err != nil {
		panic(fmt.Sprintf("failed to create CBOR enc mode: %v", err))
	}

	Dm, err = decOpts.DecMode()
	if err != nil {
		panic(fmt.Sprintf("failed to create CBOR dec mode: %v", err))
	}
}

// NewMessage creates a new CBOR-encoded message with protocol CID and payload
func NewMessage(cidV1 cid.Cid, payload []byte) ([]byte, error) {
	msg := Message{
		Protocol: cidV1.Bytes(),
		Payload:  payload,
	}
	return Em.Marshal(msg)
}

func (m Message) MarshalCBOR() ([]byte, error) {
	tag := cbor.Tag{
		Number:  gridTagNum,
		Content: []interface{}{m.Protocol, m.Payload},
	}
	return Em.Marshal(tag)
}

func (m *Message) UnmarshalCBOR(data []byte) error {
	var tag cbor.Tag
	if err := Dm.Unmarshal(data, &tag); err != nil {
		return err
	}

	if tag.Number != gridTagNum {
		return fmt.Errorf("invalid grid tag number: %d", tag.Number)
	}

	parts, ok := tag.Content.([]interface{})
	if !ok || len(parts) != 2 {
		return fmt.Errorf("invalid content format, expected 2-element array")
	}

	// Decode protocol CID
	switch v := parts[0].(type) {
	case []byte:
		m.Protocol = v
	case nil:
		m.Protocol = nil
	default:
		return fmt.Errorf("protocol field has unexpected type: %T", parts[0])
	}

	// Decode payload
	switch v := parts[1].(type) {
	case []byte:
		m.Payload = v
	case nil:
		m.Payload = nil
	default:
		return fmt.Errorf("payload field has unexpected type: %T", parts[1])
	}

	return nil
}
