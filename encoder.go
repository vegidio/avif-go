package avif

import (
	"fmt"
	"image"
	"io"
)

func Encode(w io.Writer, img image.Image) error {
	// Convert the image to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	data, err := encodeAVIF(*rgba)
	if _, err = w.Write(data); err != nil {
		return err
	}

	if _, err = w.Write(data); err != nil {
		return fmt.Errorf("failed to write AVIF image: %v", err)
	}

	return nil
}
