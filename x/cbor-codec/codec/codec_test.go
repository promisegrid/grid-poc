package codec_test

import (
	"cbor-codec/codec"
	"encoding/hex"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fxamacker/cbor/v2"
	. "github.com/stevegt/goadapt"
	"github.com/stretchr/testify/assert"
)

// PrintDiag prints a human-readable diagnostic string for the given CBOR
// buffer.
func PrintDiag(buf []byte) {
	// Diagnose the CBOR data
	dm, err := cbor.DiagOptions{
		ByteStringText:         true,
		ByteStringEmbeddedCBOR: true,
	}.DiagMode()
	Ck(err)
	diagnosis, err := dm.Diagnose(buf)
	Ck(err)
	Pl(diagnosis)
}

// Test payload types
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

const (
	GridTag   = 1735551332
	ImageTag  = 1735551333
	SensorTag = 1735551334
)

func setupCodec(t *testing.T) *codec.Codec {
	config := codec.CodecConfig{
		EncOptions: cbor.CoreDetEncOptions(),
		DecOptions: cbor.DecOptions{},
	}

	c, err := codec.NewCodec(config)
	assert.NoError(t, err)

	assert.NoError(t, c.RegisterTag(GridTag, GridPayload{}))
	assert.NoError(t, c.RegisterTag(ImageTag, ImagePayload{}))
	assert.NoError(t, c.RegisterTag(SensorTag, SensorData{}))

	return c
}

func TestEncode(t *testing.T) {
	c := setupCodec(t)

	tests := []struct {
		name    string
		payload interface{}
		wantHex string
	}{
		{
			name: "GridPayload",
			payload: GridPayload{
				ID:        []byte{0x01, 0x02},
				Timestamp: 1614124800,
			},
			wantHex: "da67726964a26249444201026954696d657374616d701a60359700",
		},
		{
			name: "ImagePayload",
			payload: ImagePayload{
				Width:  800,
				Height: 600,
				Format: "JPEG",
			},
			wantHex: "da67726965a365576964746819032066466f726d6174644a50454766486569676874190258",
		},
		{
			name: "SensorData",
			payload: SensorData{
				SensorID:  "temp-1",
				Temp:      25.0,
				Precision: 2,
			},
			wantHex: "da67726966a36454656d70f94e406853656e736f7249446674656d702d3169507265636973696f6e02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// encode
			encoded, err := c.Encode(tt.payload)
			assert.NoError(t, err)
			PrintDiag(encoded)
			assert.Equal(t, tt.wantHex, hex.EncodeToString(encoded))

			// decode
			decoded, err := c.Decode(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.payload, decoded)
		})
	}
}

func TestUnknownTag(t *testing.T) {
	// set up a codec without registering any tags
	config := codec.CodecConfig{
		EncOptions: cbor.CoreDetEncOptions(),
		DecOptions: cbor.DecOptions{},
	}
	c, err := codec.NewCodec(config)
	assert.NoError(t, err)

	// Data with unregistered tag (GridTag)
	data, _ := hex.DecodeString("da67726964a26249444201026954696d657374616d701a60359700")
	obj, err := c.Decode(data)
	spew.Dump(obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tag")
}

func TestInvalidStructure(t *testing.T) {
	c := setupCodec(t)

	// Malformed CBOR data
	data := []byte{0xd9, 0xd9, 0xf7, 0xda, 0x67, 0x72, 0x69, 0x64} // truncated
	_, err := c.Decode(data)
	assert.Error(t, err)
}
