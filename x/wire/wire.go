package wire

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Message represents a protocol message with tag, protocol CID, and payload.
type Message struct {
	_        struct{} `cbor:",toarray"` // Force array encoding
	Tag      []byte
	Protocol []byte
	Payload  []byte
}

var (
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

// MarshalBinary implements encoding.BinaryMarshaler
func (m *Message) MarshalBinary() ([]byte, error) {
	return em.Marshal(m) // Use deterministic encoding mode
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (m *Message) UnmarshalBinary(data []byte) error {
	return dm.Unmarshal(data, m)
}
