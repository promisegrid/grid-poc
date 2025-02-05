package cbordecode

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// DecodeInto decodes CBOR-encoded data into an existing instance provided by out.
// It uses the fxamacker/cbor/v2 package as the decoding engine.
// The expected format for the data is a single tag byte followed by
// a proper CBOR encoding of the value. No fallback simulation is performed.
func DecodeInto(data []byte, out interface{}) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("data too short")
	}
	tag := int(data[0])
	// Decode the CBOR data (data[1:]) into the provided output.
	if err := cbor.Unmarshal(data[1:], out); err != nil {
		return tag, err
	}
	return tag, nil
}

// DecodeNew decodes CBOR-encoded data and returns a pointer to a newly allocated
// instance of type T. It is a generic function, and the caller ultimately receives
// a conventional Go pointer to the decoded type without any usage of reflect.
// The data is expected to have a leading tag byte followed by a proper CBOR encoding.
// No fallback simulation is performed.
func DecodeNew[T any](data []byte) (*T, int, error) {
	if len(data) < 2 {
		var zero *T
		return zero, 0, fmt.Errorf("data too short")
	}
	tag := int(data[0])
	p := new(T)

	// Use the cbor package for decoding the remainder of data.
	if err := cbor.Unmarshal(data[1:], p); err != nil {
		return p, tag, err
	}
	return p, tag, nil
}

// DecodeValue decodes CBOR-encoded data and returns a value of type T.
// It is generic and returns a conventional Go value (not a pointer).
// The data is expected to have a leading tag byte followed by a valid CBOR encoding.
// No fallback simulation is performed.
func DecodeValue[T any](data []byte) (T, int, error) {
	var zero T
	if len(data) < 2 {
		return zero, 0, fmt.Errorf("data too short")
	}
	tag := int(data[0])
	p := new(T)
	if err := cbor.Unmarshal(data[1:], p); err != nil {
		return *p, tag, err
	}
	return *p, tag, nil
}
