package codec

import (
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

type Codec struct {
	em   cbor.EncMode
	dm   cbor.DecMode
	tags cbor.TagSet
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
		em:   em,
		dm:   dm,
		tags: tags,
	}, nil
}

func (c *Codec) RegisterTag(tagNumber uint64, payloadType interface{}) error {
	return c.tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(payloadType),
		tagNumber,
	)
}

func (c *Codec) Encode(payload interface{}) ([]byte, error) {
	wrapped := cbor.Tag{
		Number: 55799, // Outer self-describe tag
		Content: cbor.Tag{
			Number:  c.getTagForType(payload),
			Content: payload,
		},
	}
	return c.em.Marshal(wrapped)
}

func (c *Codec) Decode(data []byte) (interface{}, error) {
	var decoded interface{}
	err := c.dm.Unmarshal(data, &decoded)
	return decoded, err
}

func (c *Codec) getTagForType(payload interface{}) uint64 {
	t := reflect.TypeOf(payload)
	for _, tag := range c.tags.Tags() {
		if tag.ContentType == t {
			return tag.Number
		}
	}
	return 0
}
