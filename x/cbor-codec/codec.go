package codec

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// CodecConfig holds CBOR encoding/decoding options.
type CodecConfig struct {
	EncOptions cbor.EncOptions
	DecOptions cbor.DecOptions
}

// Codec performs CBOR encoding and decoding with persistent configurations.
type Codec struct {
	config CodecConfig
	Em     cbor.EncMode
	Dm     cbor.DecMode
}

// NewCodec creates a new Codec with the provided configuration.
func NewCodec(config CodecConfig) (*Codec, error) {
	em, err := config.EncOptions.EncMode()
	if err != nil {
		return nil, err
	}
	dm, err := config.DecOptions.DecMode()
	if err != nil {
		return nil, err
	}
	return &Codec{
		config: config,
		Em:     em,
		Dm:     dm,
	}, nil
}

// Encode serializes the payload into a CBOR byte slice, wrapping it
// in the given tag number.
func (c *Codec) Encode(tagNum uint64, payload interface{}) ([]byte, error) {
	innerEncoded, err := c.Em.Marshal(payload)
	if err != nil {
		return nil, err
	}
	rawTag := cbor.RawTag{
		Number:  tagNum,
		Content: innerEncoded,
	}
	return c.Em.Marshal(rawTag)
}

// DecodeRaw decodes a CBOR value from a byte slice into the given value.
// It performs a single-level decode and does not recursively unwrap CBOR tags.
// This function is useful for decoding the payload obtained from DecodeTag.
func (c *Codec) DecodeRaw(data []byte, out interface{}) error {
	return c.Dm.Unmarshal(data, out)
}

// DecodeTag decodes a CBOR tag from a byte slice.
// It returns the tag number as a uint64, the tag data, and an error.
// If the byte slice does not contain a tag, it returns 0, nil, nil.
func (c *Codec) DecodeTag(data []byte) (uint64, []byte, error) {
	if len(data) == 0 {
		return 0, nil, nil
	}
	// Check if the first byte indicates a CBOR tag (major type 6).
	if (data[0] >> 5) != 6 {
		return 0, nil, fmt.Errorf("data does not contain a CBOR tag")
	}

	var rawTag cbor.RawTag
	if err := c.Dm.Unmarshal(data, &rawTag); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	return rawTag.Number, rawTag.Content, nil
}

// StringToNum converts a string to a number, e.g. "grid" -> 0x67726964.
func StringToNum(s string) (n uint64) {
	for _, r := range s {
		n = (n << 8) + uint64(r)
	}
	return n
}

// NumToString converts a number to a string, e.g. 0x67726964 -> "grid".
func NumToString(n uint64) (s string) {
	for n > 0 {
		s = fmt.Sprintf("%c%s", n&0xff, s)
		n >>= 8
	}
	return s
}
