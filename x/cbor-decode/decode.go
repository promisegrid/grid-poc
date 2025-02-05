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
// ideal for situations where you first need to extract the tag number, then perform
// a secondary decode on the contained data without imposing reflection or additional
// type constraints on the caller.
//
// cbor.Tag, in contrast, is used to automatically decode the content as part of a
// higher-level structure and may involve extra processing. In our implementation,
// we use cbor.RawTag so that we can explicitly handle extracting the tag and then
// decode the content into a provided type. This approach avoids requiring the caller
// to import or use reflection, thus keeping the API clean and straightforward.
//
// Note: The CBOR specification reserves several tag numbers for standard purposes.
// For example, tag 1 is reserved for epoch-based date/time and tag 2 and 3 for bignums.
// Using these reserved tags with arbitrary data may result in encoding errors.
// Therefore, our test cases use non-reserved tag numbers (e.g. 258, 259, etc.) to avoid such conflicts.

// DecodeInto decodes CBOR-encoded data into an existing instance provided by out.
// It uses the fxamacker/cbor/v2 package as the decoding engine.
// The data is expected to contain a proper CBOR tag (major type 6) containing the CBOR-encoded value.
// The decoded tag number is returned as an int along with any decoding error.
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
// It uses the fxamacker/cbor/v2 package as the decoding engine.
// The data is expected to contain a proper CBOR tag; the tag number is returned along with the pointer.
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

// DecodeValue decodes CBOR-encoded data and returns a value of type T.
// It uses the fxamacker/cbor/v2 package as the decoding engine.
// The data is expected to contain a proper CBOR tag; the tag number is returned along with the value.
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
