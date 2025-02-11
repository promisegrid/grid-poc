package cbordecode

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

var (
	encMode cbor.EncMode
	decMode cbor.DecMode
)

// The fxamacker/cbor/v2 package provides two related types: cbor.RawTag and cbor.Tag.
// cbor.RawTag is a simple structure with a Number field (the tag number) and a
// Content field which contains the raw, still-encoded CBOR bytes. This type is
// ideal for situations where you need to explicitly extract the tag number and then
// decode the inner content into a provided type without forcing the caller to use
// reflection.
//
// cbor.Tag, in contrast, is used to associate a tag with a value such that the content
// is automatically decoded as part of a larger structure. This can hide the tag and
// processing details. In our implementation, we want to give explicit control to the caller,
// so we make use of cbor.RawTag and then perform a secondary decode of its Content. This
// design avoids requiring the caller to import reflect or handle any reflection-related details.
//
// The codec package below provides three variants of a Decode() method:
//  1. DecodeInto: the caller supplies an existing instance (or pointer thereof) to decode into.
//  2. DecodeNew: the codec allocates a new instance of the desired type and returns a pointer to it.
//  3. DecodeValue: the codec decodes into a provided pointer; this variant is provided
//     for callers that do not want to use generics.
func init() {
	var err error
	encMode, err = cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		panic(fmt.Errorf("failed to create custom encoder mode: %w", err))
	}

	decMode, err = (cbor.DecOptions{}).DecMode()
	if err != nil {
		panic(fmt.Errorf("failed to create custom decoder mode: %w", err))
	}
}

// DecodeInto decodes CBOR-encoded data into an existing instance provided by out.
// It expects the data to contain a proper CBOR tag (major type 6) which wraps the actual
// encoded content. The decoded tag number is returned as an int along with any error encountered.
func DecodeInto(data []byte, out interface{}) (int, error) {
	var rawTag cbor.RawTag
	if err := decMode.Unmarshal(data, &rawTag); err != nil {
		return 0, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	tag := int(rawTag.Number)
	if err := decMode.Unmarshal(rawTag.Content, out); err != nil {
		return tag, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return tag, nil
}

// DecodeNew decodes CBOR-encoded data and returns a pointer to a newly allocated instance of type T.
// It expects the data to contain a proper CBOR tag (major type 6). The tag number is returned along with
// the pointer to the new instance.
func DecodeNew[T any](data []byte) (*T, int, error) {
	var rawTag cbor.RawTag
	if err := decMode.Unmarshal(data, &rawTag); err != nil {
		var zero *T
		return zero, 0, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	tag := int(rawTag.Number)
	p := new(T)
	if err := decMode.Unmarshal(rawTag.Content, p); err != nil {
		return p, tag, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return p, tag, nil
}

// DecodeRaw decodes a CBOR value from a byte slice into the given value.
// It performs a single-level decode and does not recursively unwrap CBOR tags.
// This function is useful for decoding the payload obtained from DecodeTag.
func DecodeRaw(data []byte, out interface{}) error {
	return decMode.Unmarshal(data, out)
}

// DecodeTag decodes a CBOR tag from a byte slice.
// It returns the tag number as a uint64, the tag data, and an error.
// If the byte slice does not contain a tag, it returns 0, nil, nil.
func DecodeTag(data []byte) (uint64, []byte, error) {
	if len(data) == 0 {
		return 0, nil, nil
	}
	// Check if the first byte indicates a CBOR tag (major type 6).
	if (data[0] >> 5) != 6 {
		return 0, nil, fmt.Errorf("data does not contain a CBOR tag")
	}

	var rawTag cbor.RawTag
	if err := decMode.Unmarshal(data, &rawTag); err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	return rawTag.Number, rawTag.Content, nil
}
