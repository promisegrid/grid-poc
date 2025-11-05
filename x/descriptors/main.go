package main

import (
	"fmt"
	"log"

	"github.com/fxamacker/cbor/v2"
)

// PromiseGridMessage represents the 5-element CBOR message structure from PromiseGrid
// RFC 8949 (CBOR), RFC 8392 (CWT), RFC 9052 (COSE)
type PromiseGridMessage struct {
	_ struct{} `cbor:",toarray"` // Force encoding as CBOR array instead of map

	// Element 1: Protocol Tag
	ProtocolTag string `cbor:"0,keyasint"`

	// Element 2: Protocol Handler CID (Content Identifier)
	ProtocolCID string `cbor:"1,keyasint"`

	// Element 3: Grid Instance CID (isolation namespace)
	GridCID string `cbor:"2,keyasint"`

	// Element 4: CBOR Web Token Payload (claims and proof-of-possession)
	CWTPayload map[string]interface{} `cbor:"3,keyasint"`

	// Element 5: COSE Signature (cryptographic proof)
	Signature []byte `cbor:"4,keyasint"`
}

// CWTClaims represents standard CBOR Web Token claims
type CWTClaims struct {
	Issuer    string `cbor:"1,keyasint"` // iss
	Subject   string `cbor:"2,keyasint"` // sub
	Audience  string `cbor:"3,keyasint"` // aud
	ExpiresAt int64  `cbor:"4,keyasint"` // exp
	NotBefore int64  `cbor:"5,keyasint"` // nbf
	IssuedAt  int64  `cbor:"6,keyasint"` // iat
	CWTID     []byte `cbor:"7,keyasint"` // cti
}

func main() {
	// Example: Create a PromiseGrid message
	msg := PromiseGridMessage{
		ProtocolTag: "grid",
		ProtocolCID: "bafyreigmitjgwhpx2vgrzp7knbqdu2ju5ytyibfybll7tfb7eqjqujtd3y",
		GridCID:     "bafyreigmitjgwhpx2vgrzp7knbqdu2ju5ytyibfybll7tfb7eqjqujtd3y",
		CWTPayload: map[string]interface{}{
			"iss": "issuer-system",
			"sub": "subject-node",
			"aud": "audience-grid",
			"iat": int64(1704067200),
		},
		Signature: []byte("cose_signature_bytes"),
	}

	// Encode to CBOR binary format
	encoded, err := cbor.Marshal(msg)
	if err != nil {
		log.Fatalf("Encoding failed: %v", err)
	}

	fmt.Printf("Encoded CBOR (hex): %x\n", encoded)
	fmt.Printf("Message size: %d bytes\n", len(encoded))

	// Decode from CBOR binary format
	var decoded PromiseGridMessage
	err = cbor.Unmarshal(encoded, &decoded)
	if err != nil {
		log.Fatalf("Decoding failed: %v", err)
	}

	fmt.Printf("\nDecoded message:\n")
	fmt.Printf("  Protocol Tag: %s\n", decoded.ProtocolTag)
	fmt.Printf("  Protocol CID: %s\n", decoded.ProtocolCID)
	fmt.Printf("  Grid CID: %s\n", decoded.GridCID)
	fmt.Printf("  CWT Payload: %+v\n", decoded.CWTPayload)
	fmt.Printf("  Signature: %x\n", decoded.Signature)
}
