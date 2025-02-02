package codec

import (
	"fmt"
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

type Codec struct {
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
		em:        em,
		dm:        dm,
		tags:      tags,
		typeToTag: make(map[reflect.Type]uint64),
	}, nil
}

func (c *Codec) RegisterTag(tagNumber uint64, payloadType interface{}) error {
	payloadReflectType := reflect.TypeOf(payloadType)
	c.typeToTag[payloadReflectType] = tagNumber

	return c.tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		payloadReflectType,
		tagNumber,
	)
}

// Encode serializes the payload into a CBOR byte slice using wrapped
// self-describe tag per RFC 9277.
func (c *Codec) Encode(payload interface{}) ([]byte, error) {
	tagNumber := c.getTagForType(payload)
	if tagNumber == 0 {
		return nil, fmt.Errorf("no tag registered for type %T", payload)
	}

	wrapped := cbor.Tag{
		Number: 55799, // Outer self-describe tag
		Content: cbor.Tag{
			Number:  tagNumber,
			Content: payload,
		},
	}
	return c.em.Marshal(wrapped)
}

func (c *Codec) Decode(data []byte) (interface{}, error) {
	var outerTag cbor.Tag
	if err := c.dm.Unmarshal(data, &outerTag); err != nil {
		return nil, err
	}

	if outerTag.Number != 55799 {
		return nil, fmt.Errorf("expected outer tag number 55799, got %d", outerTag.Number)
	}

	innerTag, ok := outerTag.Content.(cbor.Tag)
	if !ok {
		return nil, fmt.Errorf("expected inner tag, got %T", outerTag.Content)
	}

	payloadType, ok := c.getTypeForTag(innerTag.Number)
	if !ok {
		return nil, fmt.Errorf("unknown tag")
	}

	payloadPtr := reflect.New(payloadType).Interface()
	if err := c.dm.Unmarshal(innerTag.Content, payloadPtr); err != nil {
		return nil, err
	}

	return reflect.ValueOf(payloadPtr).Elem().Interface(), nil
}

func (c *Codec) getTagForType(payload interface{}) uint64 {
	t := reflect.TypeOf(payload)
	if tag, ok := c.typeToTag[t]; ok {
		return tag
	}
	return 0
}

func (c *Codec) getTypeForTag(tagNumber uint64) (reflect.Type, bool) {
	for typ, tag := range c.typeToTag {
		if tag == tagNumber {
			return typ, true
		}
	}
	return nil, false
}
