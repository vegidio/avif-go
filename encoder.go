package avif

import (
	"fmt"
	"image"
	"image/draw"
	"io"
)

// Options represent the configuration options for encoding an AVIF image.
//
//   - Speed: Controls the encoding speed, from 0-10. Higher values result in faster encoding but lower quality
//     (default 6).
//   - AlphaQuality: Specifies the quality of the alpha channel (transparency), from 0-100 (default 60).
//   - ColorQuality: Specifies the quality of the color channels, from 0-100 (default 60).
type Options struct {
	Speed        int
	AlphaQuality int
	ColorQuality int
}

// AVIF represents a multi-frame AVIF image (animation/sequence).
//
//   - Image: The successive images (frames).
//   - Delay: The successive delay times, one per frame, in 100ths of a second (same unit as gif.GIF).
//   - LoopCount: Controls animation looping. 0 means loop forever, -1 means show each frame only once, and
//     N > 0 means the sequence is played N+1 times.
type AVIF struct {
	Image     []image.Image
	Delay     []int
	LoopCount int
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
	opts, err := normalizeOptions(options)
	if err != nil {
		return err
	}

	rgba := toRGBA(img)

	data, err := encodeAVIF(*rgba, opts)
	if err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write AVIF image: %w", err)
	}

	return nil
}

// EncodeAll encodes a multi-frame AVIF animation and writes it to the provided writer.
//
// Parameters:
//   - writer: The destination where the encoded AVIF animation will be written.
//   - a: An AVIF struct containing the frames, delays, and loop count.
//   - options: A pointer to an Options struct that specifies encoding parameters. If nil, default values are used.
//
// Returns:
//   - An error if encoding or writing fails, otherwise nil.
func EncodeAll(writer io.Writer, a *AVIF, options *Options) error {
	if len(a.Image) == 0 {
		return fmt.Errorf("at least one frame is required")
	}

	// Single frame: delegate to still image encoding
	if len(a.Image) == 1 {
		return Encode(writer, a.Image[0], options)
	}

	opts, err := normalizeOptions(options)
	if err != nil {
		return err
	}

	// Convert frames to RGBA and delays to milliseconds
	rgbaFrames := make([]image.RGBA, len(a.Image))
	delaysMs := make([]int, len(a.Image))

	for i, img := range a.Image {
		rgbaFrames[i] = *toRGBA(img)

		// Convert delay from centiseconds to milliseconds; treat 0 as 100ms (GIF default)
		delay := 10
		if i < len(a.Delay) && a.Delay[i] > 0 {
			delay = a.Delay[i]
		}
		delaysMs[i] = delay * 10
	}

	// Map LoopCount to AVIF repetitionCount
	repetitionCount := -1 // AVIF_REPETITION_COUNT_INFINITE
	if a.LoopCount == -1 {
		repetitionCount = 0
	} else if a.LoopCount > 0 {
		repetitionCount = a.LoopCount
	}

	data, err := encodeAnimatedAVIF(rgbaFrames, delaysMs, repetitionCount, opts)
	if err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return fmt.Errorf("failed to write AVIF animation: %w", err)
	}

	return nil
}

// region - Private functions

// normalizeOptions applies defaults and validates encoding options.
func normalizeOptions(options *Options) (Options, error) {
	if options == nil {
		return Options{Speed: 6, AlphaQuality: 60, ColorQuality: 60}, nil
	}

	if options.Speed < 0 || options.Speed > 10 {
		return Options{}, fmt.Errorf("speed must be between 0 and 10")
	}
	if options.AlphaQuality < 0 || options.AlphaQuality > 100 {
		return Options{}, fmt.Errorf("alpha quality must be between 0 and 100")
	}
	if options.ColorQuality < 0 || options.ColorQuality > 100 {
		return Options{}, fmt.Errorf("color quality must be between 0 and 100")
	}

	return *options, nil
}

// toRGBA converts an image.Image to *image.RGBA, returning it directly if it already is one.
func toRGBA(img image.Image) *image.RGBA {
	if r, ok := img.(*image.RGBA); ok {
		return r
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	return rgba
}

// endregion
