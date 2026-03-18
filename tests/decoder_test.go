//go:build cgo

package tests

import (
	"bytes"
	"errors"
	"image"
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

// errorReader is a helper type that always returns an error on Read
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
