package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/spf13/cobra"
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

// ExecutableDescriptor describes an embedded executable
type ExecutableDescriptor struct {
	Name        string `cbor:"0,keyasint"`
	ContentType string `cbor:"1,keyasint"`
	Size        int64  `cbor:"2,keyasint"`
	Executable  []byte `cbor:"3,keyasint"`
	Checksum    []byte `cbor:"4,keyasint"`
}

// example demonstrates CBOR encoding and decoding of PromiseGrid messages
func example() {
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

// embed reads an executable file and embeds it in a CBOR descriptor
func embed(executableName string) {
	// Read the executable file
	data, err := ioutil.ReadFile(executableName)
	if err != nil {
		log.Fatalf("Failed to read executable: %v", err)
	}

	// Get file info
	info, err := os.Stat(executableName)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}

	// Create executable descriptor
	descriptor := ExecutableDescriptor{
		Name:        executableName,
		ContentType: "application/octet-stream",
		Size:        info.Size(),
		Executable:  data,
		Checksum:    []byte("sha256-placeholder"),
	}

	// Encode descriptor to CBOR
	encoded, err := cbor.Marshal(descriptor)
	if err != nil {
		log.Fatalf("Failed to encode descriptor: %v", err)
	}

	fmt.Printf("Executable: %s\n", executableName)
	fmt.Printf("Size: %d bytes\n", info.Size())
	fmt.Printf("Descriptor size: %d bytes (CBOR encoded)\n", len(encoded))
	fmt.Printf("Descriptor (hex): %x\n", encoded)

	// Create a PromiseGridMessage wrapping the descriptor
	descriptorBytes, _ := cbor.Marshal(descriptor)
	msg := PromiseGridMessage{
		ProtocolTag: "grid",
		ProtocolCID: "bafyreigmitjgwhpx2vgrzp7knbqdu2ju5ytyibfybll7tfb7eqjqujtd3y",
		GridCID:     "bafyreigmitjgwhpx2vgrzp7knbqdu2ju5ytyibfybll7tfb7eqjqujtd3y",
		CWTPayload: map[string]interface{}{
			"descriptor_type": "executable",
			"executable_name": executableName,
		},
		Signature: descriptorBytes,
	}

	// Encode full message
	fullMsg, err := cbor.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed to encode message: %v", err)
	}

	fmt.Printf("\nPromiseGrid Message size: %d bytes (CBOR encoded)\n", len(fullMsg))
	fmt.Printf("PromiseGrid Message (hex): %x\n", fullMsg)
}

// exampleCmd is the Cobra subcommand for running the example
var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Run PromiseGrid CBOR encoding/decoding example",
	Long:  "Demonstrates how to create, encode, and decode PromiseGrid 5-element CBOR messages",
	Run: func(cmd *cobra.Command, args []string) {
		example()
	},
}

// embedCmd is the Cobra subcommand for embedding executables
var embedCmd = &cobra.Command{
	Use:   "embed <executable>",
	Short: "Embed an executable in a CBOR descriptor",
	Long:  "Reads an executable file and creates a CBOR-encoded descriptor containing the binary data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		embed(args[0])
	},
}

// rootCmd is the root Cobra command
var rootCmd = &cobra.Command{
	Use:   "promisegrid",
	Short: "PromiseGrid CLI - Decentralized Computing System",
	Long:  "PromiseGrid command-line interface for managing messages and grid operations",
}

func init() {
	rootCmd.AddCommand(exampleCmd)
	rootCmd.AddCommand(embedCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
