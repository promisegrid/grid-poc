package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// CWT claims structure
type CWTClaims struct {
	Issuer     string    `cbor:"1,keyasint"`
	Subject    string    `cbor:"2,keyasint"`
	Expiration time.Time `cbor:"4,keyasint"`
	IssuedAt   time.Time `cbor:"6,keyasint"`
}

// COSE_Sign1 structure
type COSESign1 struct {
	Protected   []byte
	Unprotected map[interface{}]interface{}
	Payload     []byte
	Signature   []byte
}

const (
	CWTTag       = 61 // CBOR tag for CWT
	COSESign1Tag = 18 // CBOR tag for COSE_Sign1
)

func main() {
	// Generate a key pair for signing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Create CWT claims
	claims := CWTClaims{
		Issuer:     "https://example.com",
		Subject:    "1234567890",
		Expiration: time.Now().Add(time.Hour),
		IssuedAt:   time.Now(),
	}

	// Encode CWT claims
	cwtBytes, err := cbor.Marshal(claims)
	if err != nil {
		log.Fatal(err)
	}

	// Create COSE_Sign1 structure
	coseSign1 := COSESign1{
		Protected:   []byte{}, // Empty for simplicity
		Unprotected: make(map[interface{}]interface{}),
		Payload:     cwtBytes,
	}

	// Sign the payload
	signatureInput, err := cbor.Marshal([]interface{}{
		"Signature1",
		coseSign1.Protected,
		[]byte{}, // External AAD, empty for this example
		coseSign1.Payload,
	})
	if err != nil {
		log.Fatal(err)
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, signatureInput)
	if err != nil {
		log.Fatal(err)
	}
	coseSign1.Signature = append(r.Bytes(), s.Bytes()...)

	// Encode COSE_Sign1
	coseBytes, err := cbor.Marshal(cbor.Tag{Number: COSESign1Tag, Content: coseSign1})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Encoded COSE-signed CWT: %x\n", coseBytes)

	// Decoding and verification
	var decodedCOSE COSESign1
	if err := cbor.Unmarshal(coseBytes, &decodedCOSE); err != nil {
		log.Fatal(err)
	}

	// Verify signature
	signatureInput, err = cbor.Marshal([]interface{}{
		"Signature1",
		decodedCOSE.Protected,
		[]byte{}, // External AAD
		decodedCOSE.Payload,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Instead of using nil pointers from new(ecdsa.PublicKey), create proper big.Int values.
	rBig := new(big.Int)
	sBig := new(big.Int)
	sigLen := len(decodedCOSE.Signature)
	if sigLen%2 != 0 {
		log.Fatal("invalid signature length")
	}
	rBig.SetBytes(decodedCOSE.Signature[:sigLen/2])
	sBig.SetBytes(decodedCOSE.Signature[sigLen/2:])

	if !ecdsa.Verify(&privateKey.PublicKey, signatureInput, rBig, sBig) {
		log.Fatal("Signature verification failed")
	}

	// Decode CWT claims
	var decodedClaims CWTClaims
	if err := cbor.Unmarshal(decodedCOSE.Payload, &decodedClaims); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decoded and verified CWT claims: %+v\n", decodedClaims)
}
