package wire

import (
	"github.com/fxamacker/cbor/v2"
)

// Message represents a protocol message with tag, protocol CID, and payload.
type Message struct {
	_        struct{} `cbor:",toarray"` // Force array encoding
	Tag      []byte
	Protocol []byte
	Payload  []byte
}

// EncMode configures CBOR encoding to use array format for structs
var encOpts = cbor.EncOptions{
	Sort:          cbor.SortCoreDeterministic,
	StructToArray: true,
}

// encMode is the reusable CBOR encoding mode with struct array configuration
var encMode, _ = encOpts.EncMode()

// MarshalCBOR implements CBOR marshaling with array format
func (m *Message) MarshalCBOR() ([]byte, error) {
	return encMode.Marshal(m)
}

// UnmarshalCBOR implements CBOR unmarshaling expecting array format
func (m *Message) UnmarshalCBOR(data []byte) error {
	return cbor.Unmarshal(data, m)
}
