package codec_test

import (
	"encoding/hex"
	"testing"

	codec "github.com/stevegt/grid-poc/x/cbor-codec"

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

	assert.NoError(t, c.RegisterTagNumber(GridTag, GridPayload{}))
	assert.NoError(t, c.RegisterTagNumber(ImageTag, ImagePayload{}))
	assert.NoError(t, c.RegisterTagNumber(SensorTag, SensorData{}))

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
	// spew.Dump(obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tag")
	content := obj.(cbor.Tag).Content.(map[interface{}]interface{})
	ts := content["Timestamp"].(uint64)
	assert.Equal(t, ts, uint64(1614124800))
}

func TestInvalidStructure(t *testing.T) {
	c := setupCodec(t)

	// Malformed CBOR data
	data := []byte{0xd9, 0xd9, 0xf7, 0xda, 0x67, 0x72, 0x69, 0x64} // truncated
	_, err := c.Decode(data)
	assert.Error(t, err)
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
			input: "GoLang",
			// Expected: 0x476f4c616e67
			expectNum: 0x476f4c616e67,
		},
	}

	for _, tt := range tests {
		t.Run("StringToNum_"+tt.input, func(t *testing.T) {
			result := codec.StringToNum(tt.input)
			assert.Equal(t, tt.expectNum, result, "StringToNum(%q) should be %x", tt.input, tt.expectNum)
		})

		t.Run("NumToString_"+tt.input, func(t *testing.T) {
			// For the zero case, ensure that converting 0 returns an empty string.
			result := codec.NumToString(tt.expectNum)
			assert.Equal(t, tt.input, result, "NumToString(%x) should be %q", tt.expectNum, tt.input)
		})
	}

	// Additional round-trip test: converting a string to number and back should yield the original string.
	stringsToTest := []string{"testing", "123abc", "CBOR", "Go", ""}
	for _, s := range stringsToTest {
		t.Run("RoundTrip_"+s, func(t *testing.T) {
			num := codec.StringToNum(s)
			str := codec.NumToString(num)
			assert.Equal(t, s, str, "Round-trip conversion failed for %q", s)
		})
	}
}

func TestRegisterTagName(t *testing.T) {
	config := codec.CodecConfig{
		EncOptions: cbor.CoreDetEncOptions(),
		DecOptions: cbor.DecOptions{},
	}
	c, err := codec.NewCodec(config)
	assert.NoError(t, err)

	tagName := "example"
	expectedTag := codec.StringToNum(tagName)
	// TestPayload will be registered so that decoding returns a pointer.
	payload := TestPayload{Value: "hello"}

	// Register using tag name.
	err = c.RegisterTagName(tagName, payload)
	assert.NoError(t, err)

	// Verify that the tag number is correctly assigned for the payload type.
	gotTag := c.GetTagForType(payload)
	// It might be found via the pointer mapping.
	if gotTag == 0 {
		gotTag = c.GetTagForType(&payload)
	}
	assert.Equal(t, expectedTag, gotTag, "RegisterTagName did not assign the expected tag number")

	// Encode the payload and then decode it to ensure full round-trip functionality.
	encoded, err := c.Encode(payload)
	assert.NoError(t, err)
	decoded, err := c.Decode(encoded)
	assert.NoError(t, err)
	// For tag name registration, decoding returns a pointer.
	assert.Equal(t, &payload, decoded, "Decoded payload does not match the original")
}
