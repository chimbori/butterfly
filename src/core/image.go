package core

import (
	"bytes"
	"image"
	"image/png"
)

func CompressPNG(input []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
