package avif

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
)

// Encode encodes an image.Image to AVIF format and writes it to the provided io.Writer.
//
// If any error occurs during encoding or writing, it returns the error.
func Encode(writer io.Writer, img image.Image) error {
	// Disable SVT‑AV1 logs by setting the environment variable
	os.Setenv("SVT_LOG", "-1")

	// Convert the image to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)

	data, err := encodeAVIF(*rgba)
	if _, err = writer.Write(data); err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write AVIF image: %v", err)
	}

	return nil
}
