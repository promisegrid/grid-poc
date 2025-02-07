package codec

import (
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	// . "github.com/stevegt/goadapt"
)

type Codec struct {
	config CodecConfig
	em     cbor.EncMode
	dm     cbor.DecMode
	tags   cbor.TagSet
	// map from a Go type to a CBOR tag number.
	typeToTag map[reflect.Type]uint64
	// map from CBOR tag number to a canonical Go type.
	tagToType map[uint64]reflect.Type
}

type CodecConfig struct {
	EncOptions cbor.EncOptions
	DecOptions cbor.DecOptions
}

func NewCodec(config CodecConfig) (*Codec, error) {
	tags := cbor.NewTagSet()

	em, err := config.EncOptions.EncMode()
	if err != nil {
		return nil, err
	}

	dm, err := config.DecOptions.DecModeWithTags(tags)
	if err != nil {
		return nil, err
	}

	return &Codec{
		config:    config,
		em:        em,
		dm:        dm,
		tags:      tags,
		typeToTag: make(map[reflect.Type]uint64),
		tagToType: make(map[uint64]reflect.Type),
	}, nil
}

// RegisterTagNumber registers a tag number for the given payload type.
// It stores the type mapping in both directions. For types passed as non-pointers,
// an additional mapping for the pointer type is added. (This allows c.Encode to find
// the tag number whether a value or pointer is passed.)
func (c *Codec) RegisterTagNumber(tagNumber uint64, payloadType interface{}) error {
	payloadReflectType := reflect.TypeOf(payloadType)
	c.typeToTag[payloadReflectType] = tagNumber
	c.tagToType[tagNumber] = payloadReflectType
	// also register the pointer type if not already a pointer.
	if payloadReflectType.Kind() != reflect.Ptr {
		ptrType := reflect.PtrTo(payloadReflectType)
		c.typeToTag[ptrType] = tagNumber
	}

	err := c.tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		payloadReflectType,
		tagNumber,
	)
	if err != nil {
		return err
	}

	// re-initialize the decoder with the updated tag set
	c.dm, err = c.config.DecOptions.DecModeWithTags(c.tags)
	if err != nil {
		return err
	}

	return nil
}

// RegisterTagName converts the given tag name to a tag number and
// registers it with the codec. When using tag name registration the
// canonical type is forced to be a pointer type so that decoding
// returns a pointer.
func (c *Codec) RegisterTagName(tagName string, payloadType interface{}) error {
	// If payloadType is not already a pointer, convert it.
	t := reflect.TypeOf(payloadType)
	if t.Kind() != reflect.Ptr {
		// Create a nil pointer of type *T.
		payloadType = reflect.Zero(reflect.PtrTo(t)).Interface()
	}
	num := StringToNum(tagName)
	return c.RegisterTagNumber(num, payloadType)
}

// GetTagForType returns the tag number for the given payload value.
// It first checks the concrete type and then, if not found and the
// type is not a pointer, checks the pointer type.
func (c *Codec) GetTagForType(payload interface{}) uint64 {
	t := reflect.TypeOf(payload)
	if tag, ok := c.typeToTag[t]; ok {
		return tag
	}
	if t.Kind() != reflect.Ptr {
		if tag, ok := c.typeToTag[reflect.PtrTo(t)]; ok {
			return tag
		}
	}
	return 0
}

// GetTypeForTag returns the registered Go type for the given tag number.
func (c *Codec) GetTypeForTag(tagNumber uint64) (reflect.Type, bool) {
	typ, ok := c.tagToType[tagNumber]
	return typ, ok
}

// Encode serializes the payload into a CBOR byte slice.
// It first marshals the payload, then wraps the encoded content
// in a CBOR raw tag (major type 6) using the tag number registered
// for the payload’s type.
func (c *Codec) Encode(payload interface{}) ([]byte, error) {
	tagNumber := c.GetTagForType(payload)
	if tagNumber == 0 {
		return nil, fmt.Errorf("no tag registered for type %T", payload)
	}

	// First encode the payload.
	innerEncoded, err := c.em.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Wrap the encoded payload in a CBOR raw tag.
	rawTag := cbor.RawTag{
		Number:  tagNumber,
		Content: innerEncoded,
	}

	return c.em.Marshal(rawTag)
}

// Decode deserializes the CBOR data.
// It first decodes the data into a CBOR raw tag to extract the tag number
// and the raw encoded content. If the tag is registered, it then decodes
// the inner content into an instance of the registered type. If the tag is
// unknown, it decodes the content into an interface{} and returns a cbor.Tag.
func (c *Codec) Decode(data []byte) (interface{}, error) {
	var rawTag cbor.RawTag
	if err := c.dm.Unmarshal(data, &rawTag); err != nil {
		return nil, err
	}
	tagNumber := rawTag.Number
	typ, found := c.tagToType[tagNumber]
	if !found {
		var unknown interface{}
		if err := c.dm.Unmarshal(rawTag.Content, &unknown); err != nil {
			return nil, err
		}
		return cbor.Tag{Number: tagNumber, Content: unknown}, fmt.Errorf("unknown tag %d", tagNumber)
	}
	// Allocate a new instance of the registered type.
	instance := reflect.New(typ)
	if err := c.dm.Unmarshal(rawTag.Content, instance.Interface()); err != nil {
		return nil, err
	}

	// If the canonical type is a pointer, return it directly.
	// Otherwise (e.g., if a value type was registered) return the dereferenced value.
	if typ.Kind() == reflect.Ptr {
		return instance.Interface(), nil
	} else if typ.Kind() == reflect.Struct {
		return instance.Elem().Interface(), nil
	}
	return instance.Interface(), nil
}

// StringToNum converts a string to a number, e.g. "grid" -> 0x67726964
func StringToNum(s string) (n uint64) {
	for _, r := range s {
		n = (n << 8) + uint64(r)
	}
	return n
}

// NumToString converts a number to a string, e.g. 0x67726964 -> "grid"
func NumToString(n uint64) (s string) {
	for n > 0 {
		s = fmt.Sprintf("%c%s", n&0xff, s)
		n >>= 8
	}
	return s
}
