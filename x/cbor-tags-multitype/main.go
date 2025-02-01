package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/fxamacker/cbor/v2"
)

// Define multiple payload types
type GridPayload struct {
	ID        []byte `cbor:"1,keyasint"`
	Timestamp uint64 `cbor:"2,keyasint"`
}

type ImagePayload struct {
	Width  uint   `cbor:"1,keyasint"`
	Height uint   `cbor:"2,keyasint"`
	Format string `cbor:"3,keyasint"`
}

type SensorData struct {
	SensorID  string  `cbor:"1,keyasint"`
	Temp      float64 `cbor:"2,keyasint"`
	Precision uint8   `cbor:"3,keyasint"`
}

// Define tag numbers
const (
	GridTag   = 1735551332
	ImageTag  = 1735551333
	SensorTag = 1735551334
)

func main() {
	// Create encoding mode
	em, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		log.Fatal(err)
	}

	// Create tag set for decoding
	tags := cbor.NewTagSet()
	registerTag(tags, GridPayload{}, GridTag)
	registerTag(tags, ImagePayload{}, ImageTag)
	registerTag(tags, SensorData{}, SensorTag)

	// Create decoding mode
	dm, err := cbor.DecOptions{}.DecModeWithTags(tags)
	if err != nil {
		log.Fatal(err)
	}

	// Example payloads
	payloads := []interface{}{
		GridPayload{ID: []byte{0x01, 0x02}, Timestamp: 1614124800},
		ImagePayload{Width: 800, Height: 600, Format: "JPEG"},
		SensorData{SensorID: "temp-1", Temp: 23.5, Precision: 2},
	}

	for _, payload := range payloads {
		// Encoding
		encoded := encodePayload(em, payload)
		fmt.Printf("Encoded (%T): %x\n", payload, encoded)

		// Decoding
		decoded := decodePayload(dm, encoded)
		fmt.Printf("Decoded: %#v\n\n", decoded)
	}
}

func registerTag(tags cbor.TagSet, typ interface{}, tagNum uint64) {
	err := tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(typ),
		tagNum,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func encodePayload(em cbor.EncMode, payload interface{}) []byte {
	wrapped := cbor.Tag{
		Number: 55799, // Outer self-describe tag
		Content: cbor.Tag{
			Number:  getTagForType(payload),
			Content: payload,
		},
	}

	encoded, err := em.Marshal(wrapped)
	if err != nil {
		log.Fatal(err)
	}
	return encoded
}

func decodePayload(dm cbor.DecMode, data []byte) interface{} {
	var decoded interface{}
	if err := dm.Unmarshal(data, &decoded); err != nil {
		log.Fatal(err)
	}
	return decoded
}

func getTagForType(payload interface{}) uint64 {
	switch payload.(type) {
	case GridPayload:
		return GridTag
	case ImagePayload:
		return ImageTag
	case SensorData:
		return SensorTag
	default:
		log.Fatal("Unknown payload type")
		return 0
	}
}
