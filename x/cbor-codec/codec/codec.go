package codec

import (
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	. "github.com/stevegt/goadapt"
)

type Codec struct {
	config    CodecConfig
	em        cbor.EncMode
	dm        cbor.DecMode
	tags      cbor.TagSet
	typeToTag map[reflect.Type]uint64
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
	}, nil
}

func (c *Codec) RegisterTagNumber(tagNumber uint64, payloadType interface{}) error {
	payloadReflectType := reflect.TypeOf(payloadType)
	c.typeToTag[payloadReflectType] = tagNumber

	err := c.tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		payloadReflectType,
		tagNumber,
	)
	Ck(err)

	// re-initialize the decoder mode with the updated tag set
	c.dm, err = c.config.DecOptions.DecModeWithTags(c.tags)
	Ck(err)

	return nil
}

// RegisterTagName converts the given tag name to a tag number and
// registers it with the codec.
func (c *Codec) RegisterTagName(tagName string, payloadType interface{}) error {
	num := StringToNum(tagName)
	return c.RegisterTagNumber(num, payloadType)
}

// Encode serializes the payload into a CBOR byte slice
func (c *Codec) Encode(payload interface{}) ([]byte, error) {
	tagNumber := c.GetTagForType(payload)
	if tagNumber == 0 {
		return nil, fmt.Errorf("no tag registered for type %T", payload)
	}

	obj := cbor.Tag{
		Number:  tagNumber,
		Content: payload,
	}
	return c.em.Marshal(obj)
}

func (c *Codec) Decode(data []byte) (interface{}, error) {
	var payload interface{}
	if err := c.dm.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	// typ, ok := c.GetTypeForTag(c.GetTagForType(payload))
	typ := reflect.TypeOf(payload)

	// return payload as well as an error if the tag is not registered
	_, ok := c.typeToTag[typ]
	if !ok {
		return payload, fmt.Errorf("unknown tag")
	}

	// cast payload to typ
	obj := reflect.New(typ).Interface()
	reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(payload))

	return obj, nil
}

func (c *Codec) GetTagForType(payload interface{}) uint64 {
	t := reflect.TypeOf(payload)
	if tag, ok := c.typeToTag[t]; ok {
		return tag
	}
	return 0
}

func (c *Codec) GetTypeForTag(tagNumber uint64) (reflect.Type, bool) {
	for typ, tag := range c.typeToTag {
		if tag == tagNumber {
			return typ, true
		}
	}
	return nil, false
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
