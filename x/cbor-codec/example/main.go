// User code example
package main

import (
	"cbor-codec/codec"

	"github.com/fxamacker/cbor/v2"
)

type CustomPayload struct {
	Field1 string `cbor:"1,keyasint"`
	Field2 int    `cbor:"2,keyasint"`
}

func main() {
	c, _ := codec.NewCodec(codec.CodecConfig{
		EncOptions: cbor.CoreDetEncOptions(),
	})

	c.RegisterTag(1234, CustomPayload{})

	payload := CustomPayload{Field1: "test", Field2: 42}
	encoded, _ := c.Encode(payload)
	decoded, _ := c.Decode(encoded)
}
