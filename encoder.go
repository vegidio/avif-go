package avif

import (
	"fmt"
	"image"
	"image/draw"
	"io"
)

// Options represents the configuration options for encoding an AVIF image.
//   - Speed: Controls the encoding speed, from 0-10. Higher values result in faster encoding but lower quality
//     (default 6).
//   - AlphaQuality: Specifies the quality of the alpha channel (transparency), from 0-100 (default 60).
//   - ColorQuality: Specifies the quality of the color channels, from 0-100 (default 60).
type Options struct {
	Speed        int
	AlphaQuality int
	ColorQuality int
}

// Encode encodes an image into the AVIF format and writes it to the provided writer.
//
// Parameters:
//   - writer: The destination where the encoded AVIF image will be written.
//   - img: The input image to be encoded.
//   - options: A pointer to an Options struct that specifies encoding parameters. If nil, default values are used.
//
// Returns:
//   - An error if encoding or writing fails, otherwise nil.
func Encode(writer io.Writer, img image.Image, options *Options) error {
	// Convert the image to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)

	// Set default values for options if they are not set
	if options == nil {
		options = &Options{Speed: 6, AlphaQuality: 60, ColorQuality: 60}
	}

	if options.Speed < 0 || options.Speed > 10 {
		return fmt.Errorf("speed must be between 0 and 10")
	}
	if options.AlphaQuality < 0 || options.AlphaQuality > 100 {
		return fmt.Errorf("alpha quality must be between 0 and 100")
	}
	if options.ColorQuality < 0 || options.ColorQuality > 100 {
		return fmt.Errorf("color quality must be between 0 and 100")
	}

	data, err := encodeAVIF(*rgba, *options)
	if err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write AVIF image: %v", err)
	}

	return nil
}
