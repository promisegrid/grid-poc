package main

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

func main() {
	data, _ := cbor.Marshal("hello world")

	// show the message data in hex
	for _, b := range data {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
	// Output: 6b 68 65 6c 6c 6f 20 77 6f 72 6c 64

	// show the first byte in binary
	fmt.Printf("%08b\n", data[0])
	// Output: 01101011

	// show the CBOR major type
	fmt.Printf("%d\n", data[0]>>5)
	// Output: 3
}
