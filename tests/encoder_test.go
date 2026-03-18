package tests

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg" // Register JPEG format
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vegidio/avif-go"
)

func TestEncode_Options(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	t.Run("default options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := avif.Encode(buf, img, nil)

		assert.NoError(t, err)
		assert.NotEmpty(t, buf.Bytes())
	})

	t.Run("custom options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		options := &avif.Options{
			Speed:        8,
			AlphaQuality: 80,
			ColorQuality: 90,
		}

		err := avif.Encode(buf, img, options)

		assert.NoError(t, err)
		assert.NotEmpty(t, buf.Bytes())
	})
}

func TestEncode_Validation(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buf := &bytes.Buffer{}

	t.Run("speed validation", func(t *testing.T) {
		tests := []struct {
			name    string
			speed   int
			wantErr bool
		}{
			{"speed -1", -1, true},
			{"speed 0", 0, false},
			{"speed 5", 5, false},
			{"speed 10", 10, false},
			{"speed 11", 11, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				options := &avif.Options{Speed: tt.speed, AlphaQuality: 60, ColorQuality: 60}
				err := avif.Encode(buf, img, options)

				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "speed must be between 0 and 10")
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("alpha quality validation", func(t *testing.T) {
		tests := []struct {
			name         string
			alphaQuality int
			wantErr      bool
		}{
			{"alpha -1", -1, true},
			{"alpha 0", 0, false},
			{"alpha 50", 50, false},
			{"alpha 100", 100, false},
			{"alpha 101", 101, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				options := &avif.Options{Speed: 6, AlphaQuality: tt.alphaQuality, ColorQuality: 60}
				err := avif.Encode(buf, img, options)

				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "alpha quality must be between 0 and 100")
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("color quality validation", func(t *testing.T) {
		tests := []struct {
			name         string
			colorQuality int
			wantErr      bool
		}{
			{"color -1", -1, true},
			{"color 0", 0, false},
			{"color 50", 50, false},
			{"color 100", 100, false},
			{"color 101", 101, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				options := &avif.Options{Speed: 6, AlphaQuality: 60, ColorQuality: tt.colorQuality}
				err := avif.Encode(buf, img, options)

				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "color quality must be between 0 and 100")
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestEncode_ImageConversion(t *testing.T) {
	// Create a non-RGBA image
	img := image.NewGray(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.Gray{Y: uint8(x * y)})
		}
	}

	buf := &bytes.Buffer{}
	err := avif.Encode(buf, img, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
}

func TestEncode_Errors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	t.Run("writer error", func(t *testing.T) {
		errWriter := &errorWriter{}
		err := avif.Encode(errWriter, img, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write AVIF image")
	})
}

func TestEncode_WithRealImage(t *testing.T) {
	// Try to load the test image
	if _, err := os.Stat("../assets/image.jpg"); err == nil {
		file, err := os.Open("../assets/image.jpg")
		require.NoError(t, err)
		defer file.Close()

		img, _, err := image.Decode(file)
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		options := &avif.Options{Speed: 6, AlphaQuality: 60, ColorQuality: 60}
		err = avif.Encode(buf, img, options)

		assert.NoError(t, err)
		assert.NotEmpty(t, buf.Bytes())
	} else {
		t.Skip("assets/image.jpg not found, skipping real image test")
	}
}

func TestEncodeAll_MultiFrame(t *testing.T) {
	frames := make([]image.Image, 3)
	for i := range frames {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				img.Set(x, y, color.RGBA{R: uint8(i * 80), G: 100, B: 200, A: 255})
			}
		}
		frames[i] = img
	}

	a := &avif.AVIF{
		Image:     frames,
		Delay:     []int{10, 10, 10},
		LoopCount: 0,
	}

	buf := &bytes.Buffer{}
	err := avif.EncodeAll(buf, a, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
}

func TestEncodeAll_SingleFrame(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	a := &avif.AVIF{
		Image:     []image.Image{img},
		Delay:     []int{10},
		LoopCount: 0,
	}

	buf := &bytes.Buffer{}
	err := avif.EncodeAll(buf, a, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
}

func TestEncodeAll_EmptyFrames(t *testing.T) {
	a := &avif.AVIF{
		Image: []image.Image{},
	}

	buf := &bytes.Buffer{}
	err := avif.EncodeAll(buf, a, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one frame is required")
}

func TestEncodeAll_Validation(t *testing.T) {
	frames := []image.Image{
		image.NewRGBA(image.Rect(0, 0, 10, 10)),
		image.NewRGBA(image.Rect(0, 0, 10, 10)),
	}

	a := &avif.AVIF{
		Image: frames,
		Delay: []int{10, 10},
	}

	buf := &bytes.Buffer{}
	err := avif.EncodeAll(buf, a, &avif.Options{Speed: 11})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "speed must be between 0 and 10")
}

func TestEncodeAll_WithRealGIF(t *testing.T) {
	if _, err := os.Stat("../assets/spider.gif"); err != nil {
		t.Skip("assets/spider.gif not found, skipping real GIF test")
	}

	file, err := os.Open("../assets/spider.gif")
	require.NoError(t, err)
	defer file.Close()

	gifData, err := gif.DecodeAll(file)
	require.NoError(t, err)
	require.Greater(t, len(gifData.Image), 1, "spider.gif should have multiple frames")

	// Build AVIF struct from GIF frames with proper full-canvas compositing
	canvasWidth := gifData.Config.Width
	canvasHeight := gifData.Config.Height
	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	frames := make([]image.Image, len(gifData.Image))
	for i, frame := range gifData.Image {
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		cloned := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
		copy(cloned.Pix, canvas.Pix)
		frames[i] = cloned
	}

	a := &avif.AVIF{
		Image:     frames,
		Delay:     gifData.Delay,
		LoopCount: gifData.LoopCount,
	}

	buf := &bytes.Buffer{}
	err = avif.EncodeAll(buf, a, &avif.Options{Speed: 10})

	assert.NoError(t, err)
	assert.NotEmpty(t, buf.Bytes())
	t.Logf("Encoded %d frames to %d bytes", len(frames), buf.Len())
}

// errorWriter is a helper type that always returns an error on Write
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}
