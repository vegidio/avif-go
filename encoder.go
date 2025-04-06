package avif

import (
	"fmt"
	"image"
	"io"
)

// Encode encodes an image.Image to AVIF format and writes it to the provided io.Writer.
//
// If any error occurs during encoding or writing, it returns the error.
func Encode(writer io.Writer, img image.Image) error {
	// Convert the image to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	data, err := encodeAVIF(*rgba)
	if _, err = writer.Write(data); err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write AVIF image: %v", err)
	}

	return nil
}
