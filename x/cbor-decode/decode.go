package cbordecode

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

var (
	encMode cbor.EncMode
	decMode cbor.DecMode
)

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

// DecodeValue decodes CBOR-encoded data and returns a new instance of type T as a conventional value.
// It expects the data to contain a proper CBOR tag (major type 6). The tag number is returned along with
// the decoded value. This function is especially useful when the caller wishes to decode dynamically by
// instantiating T as interface{}.
func DecodeValue[T any](data []byte) (T, int, error) {
	var zero T
	var rawTag cbor.RawTag
	if err := decMode.Unmarshal(data, &rawTag); err != nil {
		return zero, 0, fmt.Errorf("failed to unmarshal tag: %w", err)
	}
	tag := int(rawTag.Number)
	p := new(T)
	if err := decMode.Unmarshal(rawTag.Content, p); err != nil {
		return *p, tag, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return *p, tag, nil
}
