package codec_test

import (
	"encoding/hex"
	"testing"

	codec "github.com/stevegt/grid-poc/x/cbor-codec"

	"github.com/fxamacker/cbor/v2"
	. "github.com/stevegt/goadapt"
	"github.com/stretchr/testify/assert"
)

// PrintDiag prints a human-readable diagnostic string for the given CBOR buffer.
func PrintDiag(buf []byte) {
	dm, err := cbor.DiagOptions{
		ByteStringText:         true,
		ByteStringEmbeddedCBOR: true,
	}.DiagMode()
	Ck(err)
	diagnosis, err := dm.Diagnose(buf)
	Ck(err)
	Pl(diagnosis)
}

// Test payload types.
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

type TestPayload struct {
	Value string
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
	return c
}

func TestEncode(t *testing.T) {
	c := setupCodec(t)

	tests := []struct {
		name    string
		tag     uint64
		payload interface{}
		wantHex string
	}{
		{
			name: "GridPayload",
			tag:  GridTag,
			payload: GridPayload{
				ID:        []byte{0x01, 0x02},
				Timestamp: 1614124800,
			},
			wantHex: "da67726964a26249444201026954696d657374616d701a60359700",
		},
		{
			name: "ImagePayload",
			tag:  ImageTag,
			payload: ImagePayload{
				Width:  800,
				Height: 600,
				Format: "JPEG",
			},
			wantHex: "da67726965a365576964746819032066466f726d6174644a50454766486569676874190258",
		},
		{
			name: "SensorData",
			tag:  SensorTag,
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
			// Encode the payload.
			encoded, err := c.Encode(tt.tag, tt.payload)
			assert.NoError(t, err)
			PrintDiag(encoded)
			assert.Equal(t, tt.wantHex, hex.EncodeToString(encoded))

			// decode the tag
			tag, payload, err := c.DecodeTag(encoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.tag, tag)

			switch tag {
			case GridTag:
				decoded := &GridPayload{}
				err := c.DecodeRaw(payload, decoded)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload, *decoded)
			case ImageTag:
				decoded := &ImagePayload{}
				err := c.DecodeRaw(payload, decoded)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload, *decoded)
			case SensorTag:
				decoded := &SensorData{}
				err := c.DecodeRaw(payload, decoded)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload, *decoded)
			}
		})
	}
}

func TestStringToNumAndNumToString(t *testing.T) {
	tests := []struct {
		input     string
		expectNum uint64
	}{
		{
			input:     "grid",
			expectNum: 0x67726964,
		},
		{
			input:     "hello",
			expectNum: 0x68656c6c6f,
		},
		{
			input:     "",
			expectNum: 0,
		},
		{
			input:     "GoLang",
			expectNum: 0x476f4c616e67,
		},
	}

	for _, tt := range tests {
		t.Run("StringToNum_"+tt.input, func(t *testing.T) {
			result := codec.StringToNum(tt.input)
			assert.Equal(t, tt.expectNum, result, "StringToNum(%q) should be %x", tt.input, tt.expectNum)
		})

		t.Run("NumToString_"+tt.input, func(t *testing.T) {
			result := codec.NumToString(tt.expectNum)
			assert.Equal(t, tt.input, result, "NumToString(%x) should be %q", tt.expectNum, tt.input)
		})
	}

	// Round-trip test.
	stringsToTest := []string{"testing", "123abc", "CBOR", "Go", ""}
	for _, s := range stringsToTest {
		t.Run("RoundTrip_"+s, func(t *testing.T) {
			num := codec.StringToNum(s)
			str := codec.NumToString(num)
			assert.Equal(t, s, str, "Round-trip conversion failed for %q", s)
		})
	}
}
