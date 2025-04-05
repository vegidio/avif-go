//go:build cgo

package avif

import (
	"fmt"
	"image"
	"io"
)

// The init function registers the AVIF decoder with Go's image package.
// The second argument ("????ftypavif" and "????ftypavis") are substrings expected in the file header.
// "ftypavif" is for still images, while "ftypavis" is for image sequences.
func init() {
	image.RegisterFormat("avif", "????ftypavif", Decode, DecodeConfig)
	image.RegisterFormat("avif", "????ftypavis", Decode, DecodeConfig)
}

// Decode reads all data from r, decodes it using libavif, and returns an image.Image.
// This function fully decodes the image into an *image.RGBA.
func Decode(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode AVIF data: %w", err)
	}
	return decodeAVIFToRGBA(data)
}

// DecodeConfig reads enough of r to determine the image's configuration (dimensions, etc.).
// Here we read the entire data and call a lightweight C function that only parses the header.
func DecodeConfig(r io.Reader) (image.Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return image.Config{}, fmt.Errorf("failed get config of AVIF data: %w", err)
	}

	return decodeConfig(data)
}
