package wire

import (
	"bytes"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Message represents a protocol message with tag, protocol CID, and payload.
type Message struct {
	_        struct{} `cbor:",toarray"` // Enable CBOR array encoding
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
	encOpts = cbor.EncOptions{}
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

// MarshalCBOR encodes Message to CBOR using array format.
func (m *Message) MarshalCBOR() ([]byte, error) {
	var buf bytes.Buffer
	enc := em.NewEncoder(&buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}

// UnmarshalCBOR decodes CBOR to Message using array format.
func (m *Message) UnmarshalCBOR(data []byte) error {
	return dm.Unmarshal(data, m)
}
