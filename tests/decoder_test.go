//go:build cgo

package tests

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/jpeg"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vegidio/avif-go"
)

func TestDecode(t *testing.T) {
	t.Run("valid AVIF file", func(t *testing.T) {
		if _, err := os.Stat("../assets/image.avif"); err != nil {
			t.Skip("assets/image.avif not found, skipping test")
			return
		}

		file, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file.Close()

		img, err := avif.Decode(file)

		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Equal(t, img.Bounds().Dx(), 1024)
		assert.Equal(t, img.Bounds().Dy(), 1536)
	})

	t.Run("with image package", func(t *testing.T) {
		if _, err := os.Stat("../assets/image.avif"); err != nil {
			t.Skip("assets/image.avif not found, skipping test")
			return
		}

		file, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file.Close()

		img, format, err := image.Decode(file)

		assert.NoError(t, err)
		assert.Equal(t, "avif", format)
		assert.NotNil(t, img)
	})

	t.Run("reader error", func(t *testing.T) {
		errReader := &errorReader{err: errors.New("read error")}

		img, err := avif.Decode(errReader)

		assert.Error(t, err)
		assert.Nil(t, img)
		assert.Contains(t, err.Error(), "failed to decode AVIF data")
	})

	t.Run("invalid data", func(t *testing.T) {
		invalidData := []byte("not a valid AVIF file")
		reader := bytes.NewReader(invalidData)

		img, err := avif.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
	})

	t.Run("empty data", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		img, err := avif.Decode(reader)

		assert.Error(t, err)
		assert.Nil(t, img)
	})

	t.Run("consistency with DecodeConfig", func(t *testing.T) {
		if _, err := os.Stat("../assets/image.avif"); err != nil {
			t.Skip("assets/image.avif not found, skipping test")
			return
		}

		// Get config
		file1, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file1.Close()

		config, err := avif.DecodeConfig(file1)
		require.NoError(t, err)

		// Decode image
		file2, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file2.Close()

		img, err := avif.Decode(file2)
		require.NoError(t, err)

		// Compare dimensions
		assert.Equal(t, config.Width, img.Bounds().Dx())
		assert.Equal(t, config.Height, img.Bounds().Dy())
	})
}

func TestDecodeConfig(t *testing.T) {
	t.Run("valid AVIF file", func(t *testing.T) {
		if _, err := os.Stat("../assets/image.avif"); err != nil {
			t.Skip("assets/image.avif not found, skipping test")
			return
		}

		file, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file.Close()

		config, err := avif.DecodeConfig(file)

		assert.NoError(t, err)
		assert.Equal(t, config.Width, 1024)
		assert.Equal(t, config.Height, 1536)
		assert.NotNil(t, config.ColorModel)
	})

	t.Run("with image package", func(t *testing.T) {
		if _, err := os.Stat("../assets/image.avif"); err != nil {
			t.Skip("assets/image.avif not found, skipping test")
			return
		}

		file, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file.Close()

		config, format, err := image.DecodeConfig(file)

		assert.NoError(t, err)
		assert.Equal(t, "avif", format)
		assert.Equal(t, config.Width, 1024)
		assert.Equal(t, config.Height, 1536)
	})

	t.Run("reader error", func(t *testing.T) {
		errReader := &errorReader{err: errors.New("read error")}

		config, err := avif.DecodeConfig(errReader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get config of AVIF data")
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})

	t.Run("invalid data", func(t *testing.T) {
		invalidData := []byte("not a valid AVIF file")
		reader := bytes.NewReader(invalidData)

		config, err := avif.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})

	t.Run("empty data", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		config, err := avif.DecodeConfig(reader)

		assert.Error(t, err)
		assert.Equal(t, 0, config.Width)
		assert.Equal(t, 0, config.Height)
	})
}

func TestMultipleFormats(t *testing.T) {
	if _, err := os.Stat("../assets/image.avif"); err != nil {
		t.Skip("assets/image.avif not found, skipping test")
		return
	}

	if _, err := os.Stat("../assets/image.jpg"); err != nil {
		t.Skip("assets/image.jpg not found, skipping test")
		return
	}

	t.Run("decode AVIF", func(t *testing.T) {
		avifFile, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer avifFile.Close()

		avifImg, avifFormat, err := image.Decode(avifFile)
		require.NoError(t, err)
		assert.Equal(t, "avif", avifFormat)
		assert.NotNil(t, avifImg)
	})

	t.Run("decode JPEG", func(t *testing.T) {
		jpegFile, err := os.Open("../assets/image.jpg")
		require.NoError(t, err)
		defer jpegFile.Close()

		jpegImg, jpegFormat, err := image.Decode(jpegFile)
		require.NoError(t, err)
		assert.Equal(t, "jpeg", jpegFormat)
		assert.NotNil(t, jpegImg)
	})
}

func TestDecodeAll(t *testing.T) {
	t.Run("animated AVIF", func(t *testing.T) {
		if _, err := os.Stat("../assets/spider.avif"); err != nil {
			t.Skip("assets/spider.avif not found, skipping test")
		}

		file, err := os.Open("../assets/spider.avif")
		require.NoError(t, err)
		defer file.Close()

		a, err := avif.DecodeAll(file)

		require.NoError(t, err)
		assert.Greater(t, len(a.Image), 1, "should have multiple frames")
		assert.Equal(t, len(a.Image), len(a.Delay), "frames and delays should match")

		for i, delay := range a.Delay {
			assert.Greater(t, delay, 0, "frame %d delay should be > 0", i)
		}

		t.Logf("Decoded %d frames, LoopCount=%d", len(a.Image), a.LoopCount)
	})

	t.Run("still AVIF", func(t *testing.T) {
		if _, err := os.Stat("../assets/image.avif"); err != nil {
			t.Skip("assets/image.avif not found, skipping test")
		}

		file, err := os.Open("../assets/image.avif")
		require.NoError(t, err)
		defer file.Close()

		a, err := avif.DecodeAll(file)

		require.NoError(t, err)
		assert.Equal(t, 1, len(a.Image), "still image should have 1 frame")
		assert.Equal(t, 1, len(a.Delay))
	})

	t.Run("round-trip animated", func(t *testing.T) {
		// Create a synthetic 3-frame animation
		frames := make([]image.Image, 3)
		for i := range frames {
			img := image.NewRGBA(image.Rect(0, 0, 16, 16))
			for y := 0; y < 16; y++ {
				for x := 0; x < 16; x++ {
					img.Set(x, y, color.RGBA{R: uint8(i * 80), G: 100, B: 200, A: 255})
				}
			}
			frames[i] = img
		}

		original := &avif.AVIF{
			Image:     frames,
			Delay:     []int{10, 20, 30},
			LoopCount: 0,
		}

		// Encode
		buf := &bytes.Buffer{}
		err := avif.EncodeAll(buf, original, &avif.Options{Speed: 10})
		require.NoError(t, err)

		// Decode back
		decoded, err := avif.DecodeAll(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)

		assert.Equal(t, len(original.Image), len(decoded.Image), "frame count should match")
		assert.Equal(t, len(original.Delay), len(decoded.Delay), "delay count should match")
		assert.Equal(t, original.LoopCount, decoded.LoopCount, "loop count should match")
	})

	t.Run("empty data", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})
		a, err := avif.DecodeAll(reader)

		assert.Error(t, err)
		assert.Nil(t, a)
	})
}

// errorReader is a helper type that always returns an error on Read
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
