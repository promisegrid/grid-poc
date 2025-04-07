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

// Removed custom MarshalCBOR/UnmarshalCBOR methods to prevent recursion
// Encoding/decoding is handled by cbor.EncMode and struct tags
