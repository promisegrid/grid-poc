package codec_test

import (
	"cbor-codec/codec"
	"encoding/hex"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

// Test payload types
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
			wantHex: "d9d9f7da67726964a201420102021a60359700",
		},
		{
			name: "ImagePayload",
			payload: ImagePayload{
				Width:  800,
				Height: 600,
				Format: "JPEG",
			},
			wantHex: "d9d9f7da67726965a30119032002190358031804444a504547",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := c.Encode(tt.payload)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHex, hex.EncodeToString(encoded))
		})
	}
}

func TestDecode(t *testing.T) {
	c := setupCodec(t)

	gridData, _ := hex.DecodeString("d9d9f7da67726964a201420102021a60359700")
	imageData, _ := hex.DecodeString("d9d9f7da67726965a30119032002190358031804444a504547")

	tests := []struct {
		name     string
		data     []byte
		expected interface{}
	}{
		{
			name: "GridPayload",
			data: gridData,
			expected: GridPayload{
				ID:        []byte{0x01, 0x02},
				Timestamp: 1614124800,
			},
		},
		{
			name: "ImagePayload",
			data: imageData,
			expected: ImagePayload{
				Width:  800,
				Height: 600,
				Format: "JPEG",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := c.Decode(tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, decoded)
		})
	}
}

func TestUnknownTag(t *testing.T) {
	c := setupCodec(t)

	// Data with unregistered tag
	data, _ := hex.DecodeString("d9d9f7da00000001a10102")
	_, err := c.Decode(data)
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
