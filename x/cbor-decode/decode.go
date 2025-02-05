package cbordecode

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// Explanation:
//
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
// 1. DecodeInto: the caller supplies an existing instance (or pointer thereof) to decode into.
// 2. DecodeNew: the codec allocates a new instance of the desired type and returns a pointer to it.
// 3. DecodeValue: the codec returns a new instance as a conventional value (not a pointer).
//
// For callers that do not know the type at compile time, DecodeValue can be instantiated
// with the generic type interface{}. This allows for dynamic decoding without the caller having
// to import or use reflection.
 
// DecodeInto decodes CBOR-encoded data into an existing instance provided by out.
// It expects the data to contain a proper CBOR tag (major type 6) which wraps the actual
// encoded content. The decoded tag number is returned as an int along with any error encountered.
func DecodeInto(data []byte, out interface{}) (int, error) {
	var rawTag cbor.RawTag
	if err := cbor.Unmarshal(data, &rawTag); err != nil {
		return 0, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	tag := int(rawTag.Number)
	if err := cbor.Unmarshal(rawTag.Content, out); err != nil {
		return tag, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return tag, nil
}

// DecodeNew decodes CBOR-encoded data and returns a pointer to a newly allocated instance of type T.
// It expects the data to contain a proper CBOR tag (major type 6). The tag number is returned along with
// the pointer to the new instance.
func DecodeNew[T any](data []byte) (*T, int, error) {
	var rawTag cbor.RawTag
	if err := cbor.Unmarshal(data, &rawTag); err != nil {
		var zero *T
		return zero, 0, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	tag := int(rawTag.Number)
	p := new(T)
	if err := cbor.Unmarshal(rawTag.Content, p); err != nil {
		return p, tag, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return p, tag, nil
}

// DecodeValue decodes CBOR-encoded data and returns a new instance of type T as a conventional value.
// It expects the data to contain a proper CBOR tag (major type 6). The tag number is returned along with
// the decoded value. This function is especially useful when the caller wishes to decode dynamically by
// instantiating T as interface{}.
func DecodeValue[T any](data []byte) (T, int, error) {
	var zero T
	var rawTag cbor.RawTag
	if err := cbor.Unmarshal(data, &rawTag); err != nil {
		return zero, 0, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	tag := int(rawTag.Number)
	p := new(T)
	if err := cbor.Unmarshal(rawTag.Content, p); err != nil {
		return *p, tag, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return *p, tag, nil
}
