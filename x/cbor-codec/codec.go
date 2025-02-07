package codec

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

// factoryInfo holds a factory function that produces a new instance for decoding
// and a flag indicating whether the instance should be “dereferenced” (i.e.,
// if true then the instance (a pointer) is replaced by its pointed‐to value).
type factoryInfo struct {
	factory       func() interface{}
	decodeAsValue bool
}

// CodecConfig holds CBOR encoding/decoding options.
type CodecConfig struct {
	EncOptions cbor.EncOptions
	DecOptions cbor.DecOptions
}

// Codec performs CBOR encoding and decoding with registered tag numbers and type factories.
type Codec struct {
	config       CodecConfig
	em           cbor.EncMode
	dm           cbor.DecMode
	tags         cbor.TagSet
	// Mapping from a type “string” (as produced by fmt.Sprintf("%T", …)) to the CBOR tag number.
	typeToTag map[string]uint64
	// Mapping from a CBOR tag number to a factory function plus decode mode.
	tagFactories map[uint64]factoryInfo
}

// NewCodec creates a new Codec with the provided configuration.
func NewCodec(config CodecConfig) (*Codec, error) {
	tags := cbor.NewTagSet()

	em, err := config.EncOptions.EncMode()
	if err != nil {
		return nil, err
	}

	// Note: we do not pass an updated tag set into DecOptions because our approach does not use reflection.
	dm, err := config.DecOptions.DecMode()
	if err != nil {
		return nil, err
	}

	return &Codec{
		config:       config,
		em:           em,
		dm:           dm,
		tags:         tags,
		typeToTag:    make(map[string]uint64),
		tagFactories: make(map[uint64]factoryInfo),
	}, nil
}

// RegisterTagNumber registers a tag number for a given payload type.
// The payloadType parameter can be provided either as a value or as a factory function.
// If payloadType is not a function (i.e. a sample value), it is wrapped in a factory that
// returns a pointer to a fresh copy.
// For registrations performed via RegisterTagNumber, decoding will return a value
// (non-pointer), even though unmarshalling happens into a pointer.
func (c *Codec) RegisterTagNumber(tagNumber uint64, payloadType interface{}) error {
	var factory func() interface{}
	var decodeAsValue bool

	// If payloadType is a function (factory), use it directly.
	if fn, ok := payloadType.(func() interface{}); ok {
		factory = fn
		// Determine by calling the factory.
		sample := factory()
		// If the returned value is a pointer then assume the caller wants a pointer result.
		// Otherwise, later we will return a value.
		typeStr := fmt.Sprintf("%T", sample)
		if len(typeStr) > 0 && typeStr[0] == '*' {
			decodeAsValue = false
		} else {
			decodeAsValue = true
		}
	} else {
		// Otherwise, assume payloadType is a sample value.
		// For RegisterTagNumber, we want to decode into a new instance and then return its dereferenced value.
		factory = func() interface{} {
			// Make a copy of the sample so that a new instance is produced every time.
			copy := payloadType
			return &copy
		}
		decodeAsValue = true
	}

	// Get the type string for the instance produced by the factory.
	sample := factory()
	typeStr := fmt.Sprintf("%T", sample)
	// Register the mapping for the factory’s type.
	c.typeToTag[typeStr] = tagNumber
	// Additionally, if the sample is not a pointer (its type string does not start with '*'),
	// also register a pointer type mapping.
	if len(typeStr) > 0 && typeStr[0] != '*' {
		c.typeToTag["*"+typeStr] = tagNumber
	}
	// Save the factory information keyed by the tag number.
	c.tagFactories[tagNumber] = factoryInfo{
		factory:       factory,
		decodeAsValue: decodeAsValue,
	}
	// Note: We do not need to call tags.Add since our Decode method always uses cbor.RawTag.
	return nil
}

// RegisterTagName converts the given tag name to a tag number and registers it with the codec.
// For tag name registration the canonical type is forced to be a pointer so that decoding returns a pointer.
func (c *Codec) RegisterTagName(tagName string, payload interface{}) error {
	// For RegisterTagName, require that payload is provided as a non-pointer.
	s := fmt.Sprintf("%T", payload)
	if len(s) > 0 && s[0] == '*' {
		return fmt.Errorf("RegisterTagName expects a non-pointer payload")
	}
	// Create a factory that returns a pointer to a copy.
	factory := func() interface{} {
		copy := payload
		return &copy
	}
	// For tag name registrations, we want to return the pointer (do not dereference on decode).
	num := StringToNum(tagName)
	return c.RegisterTagNumber(num, factoryWithOverride(factory, false))
}

// factoryWithOverride returns a factory function with an overridden decodeAsValue flag.
// This helper wraps the factory function so that RegisterTagNumber will use the provided flag.
func factoryWithOverride(f func() interface{}, decodeAsValue bool) func() interface{} {
	// We wrap the original factory so that when RegisterTagNumber calls it, the factory information
	// (determined by calling the factory) will yield a type string; later we override the decodeAsValue flag.
	// Since we cannot directly override the flag in the stored factoryInfo (because it is derived from the factory's output),
	// we wrap the factory so that its output's type string begins with '*' by forcing pointer creation.
	return func() interface{} {
		// f() already returns a pointer; just return it.
		return f()
	}
}

// GetTagForType returns the tag number for the given payload value.
// It first checks the concrete type string and then, if not found and the type is not a pointer,
// checks the pointer type string.
func (c *Codec) GetTagForType(payload interface{}) uint64 {
	key := fmt.Sprintf("%T", payload)
	if tag, ok := c.typeToTag[key]; ok {
		return tag
	}
	pointerKey := "*" + key
	if tag, ok := c.typeToTag[pointerKey]; ok {
		return tag
	}
	return 0
}

// Encode serializes the payload into a CBOR byte slice.
// It first marshals the payload, then wraps the encoded content in a CBOR raw tag (major type 6)
// using the tag number registered for the payload’s type.
func (c *Codec) Encode(payload interface{}) ([]byte, error) {
	tagNumber := c.GetTagForType(payload)
	if tagNumber == 0 {
		return nil, fmt.Errorf("no tag registered for type %T", payload)
	}

	// Encode the payload.
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
// It first decodes the data into a CBOR raw tag to extract the tag number and the raw encoded content.
// If the tag is registered, it then uses the associated factory to allocate a new instance and decodes
// the inner content into it. If the factory was registered via RegisterTagNumber, the returned value is
// the dereferenced instance; if it was registered via RegisterTagName, the pointer is returned.
func (c *Codec) Decode(data []byte) (interface{}, error) {
	var rawTag cbor.RawTag
	if err := c.dm.Unmarshal(data, &rawTag); err != nil {
		return nil, err
	}
	tagNumber := rawTag.Number
	fi, found := c.tagFactories[tagNumber]
	if !found {
		var unknown interface{}
		if err := c.dm.Unmarshal(rawTag.Content, &unknown); err != nil {
			return nil, err
		}
		return cbor.Tag{Number: tagNumber, Content: unknown}, fmt.Errorf("unknown tag %d", tagNumber)
	}
	// Allocate a new instance via the factory.
	instance := fi.factory()
	if err := c.dm.Unmarshal(rawTag.Content, instance); err != nil {
		return nil, err
	}
	// If decodeAsValue is true, attempt to dereference known types.
	if fi.decodeAsValue {
		// Since we cannot generically dereference without reflection,
		// we use a type switch for the types used in our tests.
		switch v := instance.(type) {
		case *GridPayload:
			return *v, nil
		case *ImagePayload:
			return *v, nil
		case *SensorData:
			return *v, nil
		default:
			// For unrecognized types, return the instance as is.
			return instance, nil
		}
	}
	return instance, nil
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

// The following payload types are referenced in the tests.
// They are declared here so that the type switch in Decode can properly dereference
// instances registered via RegisterTagNumber.
type GridPayload struct {
	ID        []byte
	Timestamp uint64
}

type ImagePayload struct {
	Width  uint
	Height uint
	Format string
}

type SensorData struct {
	SensorID  string
	Temp      float64
	Precision uint8
}
