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

// Decode reads AVIF image data from the provided io.Reader and decodes it into an image.Image.
//
// It returns the decoded image or an error if the decoding process fails.
func Decode(reader io.Reader) (image.Image, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode AVIF data: %w", err)
	}
	return decodeAVIFToRGBA(data)
}

// DecodeConfig reads the configuration of an AVIF image from the provided io.Reader.
//
// It returns an image.Config containing the width, height, and color model of the image, or an error if the
// configuration cannot be determined.
func DecodeConfig(reader io.Reader) (image.Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return image.Config{}, fmt.Errorf("failed get config of AVIF data: %w", err)
	}

	return decodeConfig(data)
}
