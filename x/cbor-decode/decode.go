package cbordecode

import (
	"fmt"
)

// Message is an example type to demonstrate CBOR decoding.
type Message struct {
	Value int
}

// DecodeInto decodes CBOR-encoded data into an existing instance provided by out.
// It simulates CBOR decoding by expecting data to be at least 2 bytes,
// where the first byte is interpreted as a tag and the second byte as the value.
// The caller passes a pointer to an instance (or a nil pointer if type unknown),
// much like how encoding/json.Unmarshal works.
func DecodeInto(data []byte, out interface{}) (int, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("data too short")
	}
	tag := int(data[0])
	if out == nil {
		return tag, fmt.Errorf("cannot decode into nil")
	}

	// This switch shows examples for decoding into known types.
	switch v := out.(type) {
	case *Message:
		// For Message, we simulate that data[1] holds the decoded value.
		v.Value = int(data[1])
		return tag, nil
	case *int:
		*v = int(data[1])
		return tag, nil
	}

	return tag, fmt.Errorf("unsupported type in DecodeInto")
}

// DecodeNew decodes CBOR-encoded data and returns a pointer to a newly allocated
// instance of type T. The function is generic, and the caller can get a concrete
// pointer of the decoded type without using reflect.
func DecodeNew[T any](data []byte) (*T, int, error) {
	if len(data) < 2 {
		var zero *T
		return zero, 0, fmt.Errorf("data too short")
	}
	tag := int(data[0])
	p := new(T)
	switch v := any(p).(type) {
	case *Message:
		v.Value = int(data[1])
	case *int:
		*v = int(data[1])
		// Add further cases as needed for other types.
	}
	return p, tag, nil
}

// DecodeValue decodes CBOR-encoded data and returns a value of type T.
// It is generic, and returns a conventional Go value rather than a pointer.
func DecodeValue[T any](data []byte) (T, int, error) {
	var zero T
	if len(data) < 2 {
		return zero, 0, fmt.Errorf("data too short")
	}
	tag := int(data[0])
	// Create a temporary pointer to T so we can update it.
	p := new(T)
	switch v := any(p).(type) {
	case *Message:
		v.Value = int(data[1])
	case *int:
		*v = int(data[1])
		// Other types can be handled here.
	}
	return *p, tag, nil
}
