package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
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
func example(outputFile string) {
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

	fmt.Printf("Message size: %d bytes\n", len(encoded))

	// Write to file if specified
	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, encoded, 0644)
		if err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		fmt.Printf("Written to: %s\n", outputFile)
		return
	}

	// Decode from CBOR binary format for display
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
func embed(executableName string, outputFile string) {
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

	// Write to file if specified
	if outputFile != "" {
		err := ioutil.WriteFile(outputFile, encoded, 0644)
		if err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		fmt.Printf("Written to: %s\n", outputFile)
		return
	}
}

// execFromMemory executes a binary from memory using memfd_create
func execFromMemory(descriptorFile string, args []string) {
	// Read the CBOR-encoded descriptor
	data, err := ioutil.ReadFile(descriptorFile)
	if err != nil {
		log.Fatalf("Failed to read descriptor file: %v", err)
	}

	// Decode the descriptor
	var descriptor ExecutableDescriptor
	err = cbor.Unmarshal(data, &descriptor)
	if err != nil {
		log.Fatalf("Failed to decode descriptor: %v", err)
	}

	fmt.Printf("Executing: %s (%d bytes from memory)\n", descriptor.Name, descriptor.Size)

	// Create anonymous file in RAM using memfd_create without MFD_CLOEXEC
	// fd must remain open for kernel to map the executable during exec
	fd, err := unix.MemfdCreate(descriptor.Name, 0)
	if err != nil {
		log.Fatalf("memfd_create failed: %v", err)
	}

	// Write executable data to the memory file
	n, err := unix.Write(fd, descriptor.Executable)
	if err != nil {
		unix.Close(fd)
		log.Fatalf("Failed to write to memfd: %v", err)
	}

	if n != len(descriptor.Executable) {
		unix.Close(fd)
		log.Fatalf("Incomplete write to memfd: wrote %d of %d bytes", n, len(descriptor.Executable))
	}

	// Get the /proc/self/fd/<fd> path for execution
	procPath := fmt.Sprintf("/proc/self/fd/%d", fd)

	// Prepare arguments for exec (first arg is program name)
	execArgs := append([]string{descriptor.Name}, args...)

	// Use unix.Exec for true process replacement
	// This preserves the fd across the exec boundary
	// XXX we probably don't really want exec here because we might want to continue in our current process
	err = unix.Exec(procPath, execArgs, os.Environ())
	if err != nil {
		unix.Close(fd)
		log.Fatalf("Execution failed: %v", err)
	}

	// unix.Exec does not return on success (replaces current process)
}

// exampleCmd is the Cobra subcommand for running the example
var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Run PromiseGrid CBOR encoding/decoding example",
	Long:  "Demonstrates how to create, encode, and decode PromiseGrid 5-element CBOR messages",
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, _ := cmd.Flags().GetString("output")
		example(outputFile)
	},
}

// embedCmd is the Cobra subcommand for embedding executables
var embedCmd = &cobra.Command{
	Use:   "embed <executable>",
	Short: "Embed an executable in a CBOR descriptor",
	Long:  "Reads an executable file and creates a CBOR-encoded descriptor containing the binary data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, _ := cmd.Flags().GetString("output")
		embed(args[0], outputFile)
	},
}

// execCmd is the Cobra subcommand for executing in-memory binaries
var execCmd = &cobra.Command{
	Use:   "exec <descriptor> [arguments...]",
	Short: "Execute a binary from a CBOR descriptor in memory",
	Long:  "Reads a CBOR descriptor, extracts the executable, and runs it from RAM using memfd_create",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		descriptorFile := args
		execArgs := args[1:]
		execFromMemory(descriptorFile[0], execArgs)
	},
}

// rootCmd is the root Cobra command
var rootCmd = &cobra.Command{
	Use:   "promisegrid",
	Short: "PromiseGrid CLI - Decentralized Computing System",
	Long:  "PromiseGrid command-line interface for managing messages and grid operations",
}

func init() {
	exampleCmd.Flags().StringP("output", "o", "", "Output file for CBOR-encoded message")
	embedCmd.Flags().StringP("output", "o", "", "Output file for CBOR-encoded descriptor")
	rootCmd.AddCommand(exampleCmd)
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(execCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
