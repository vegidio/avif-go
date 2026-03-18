// Package avif is a Go library & CLI tool to encode/decode static and animated AVIF without external dependencies.
package avif

/*
#include <stdlib.h>
#include <avif/avif.h>

// Helper to get error string from avifResult
const char* get_error_string(avifResult result) {
    return avifResultToString(result);
}

// Config-only decode: reads the header and returns width and height.
// Returns error result via outResult.
void get_avif_config(const uint8_t * data, size_t size, uint32_t * width, uint32_t * height, avifResult *outResult) {
    avifDecoder* decoder = avifDecoderCreate();
    decoder->codecChoice = AVIF_CODEC_CHOICE_DAV1D;

    *outResult = avifDecoderSetIOMemory(decoder, data, size);
    if (*outResult != AVIF_RESULT_OK) {
         *width = 0;
         *height = 0;
         avifDecoderDestroy(decoder);
         return;
    }

    *outResult = avifDecoderParse(decoder);
    if (*outResult != AVIF_RESULT_OK) {
         *width = 0;
         *height = 0;
         avifDecoderDestroy(decoder);
         return;
    }

    *width = decoder->image->width;
    *height = decoder->image->height;
    avifDecoderDestroy(decoder);
}
*/
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"unsafe"
)

// Maximum tile dimensions supported by the SVT-AV1 encoder.
const (
	maxTileWidth  = 16384
	maxTileHeight = 8704
)

// encodeAVIF encodes an RGBA image to AVIF format.
//
// Speed ranges from 0 (slowest, best quality) to 10 (fastest, lower quality).
//
// ColorQuality and AlphaQuality range from 0 (worst) to 100 (lossless).
//
// Uses tiling to support images larger than SVT-AV1's dimension limits. For images within limits, creates a single tile
// (1x1 grid) with identical performance.
func encodeAVIF(rgba image.RGBA, options Options) ([]byte, error) {
	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()

	if width == 0 || height == 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	tileWidth := maxTileWidth
	tileHeight := maxTileHeight

	// Calculate the number of tiles needed (1x1 for images within limits)
	cols := (width + tileWidth - 1) / tileWidth
	rows := (height + tileHeight - 1) / tileHeight

	// Create tiles
	cellImages, err := createTiles(rgba, tileWidth, tileHeight)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, img := range cellImages {
			if img != nil {
				C.avifImageDestroy(img)
			}
		}
	}()

	// Create encoder
	encoder := C.avifEncoderCreate()
	if encoder == nil {
		return nil, fmt.Errorf("failed to create AVIF encoder")
	}
	defer C.avifEncoderDestroy(encoder)

	encoder.codecChoice = C.AVIF_CODEC_CHOICE_SVT
	encoder.speed = C.int(options.Speed)
	encoder.quality = C.int(options.ColorQuality)
	encoder.qualityAlpha = C.int(options.AlphaQuality)

	// Add the grid of images (1x1 for normal images, NxM for oversized)
	result := C.avifEncoderAddImageGrid(encoder, C.uint32_t(cols), C.uint32_t(rows),
		(**C.avifImage)(unsafe.Pointer(&cellImages[0])), C.AVIF_ADD_IMAGE_FLAG_SINGLE)

	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return nil, fmt.Errorf("failed to add image grid: %s", errStr)
	}

	// Finish encoding
	var encodedData C.avifRWData
	encodedData.data = nil
	encodedData.size = 0

	result = C.avifEncoderFinish(encoder, &encodedData)
	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return nil, fmt.Errorf("failed to finish encoding: %s", errStr)
	}
	defer C.avifRWDataFree(&encodedData)

	data := C.GoBytes(unsafe.Pointer(encodedData.data), C.int(encodedData.size))
	return data, nil
}

// encodeAnimatedAVIF encodes multiple RGBA frames into an animated AVIF sequence.
//
// Each frame is encoded with its corresponding delay (in milliseconds).
// repetitionCount controls looping: -1 for infinite, 0 for play once, N for N+1 plays.
func encodeAnimatedAVIF(frames []image.RGBA, delaysMs []int, repetitionCount int, options Options) ([]byte, error) {
	if len(frames) == 0 {
		return nil, fmt.Errorf("at least one frame is required")
	}

	// Create encoder
	encoder := C.avifEncoderCreate()
	if encoder == nil {
		return nil, fmt.Errorf("failed to create AVIF encoder")
	}
	defer C.avifEncoderDestroy(encoder)

	encoder.codecChoice = C.AVIF_CODEC_CHOICE_SVT
	encoder.speed = C.int(options.Speed)
	encoder.quality = C.int(options.ColorQuality)
	encoder.qualityAlpha = C.int(options.AlphaQuality)
	encoder.timescale = C.uint64_t(1000) // millisecond units
	encoder.repetitionCount = C.int(repetitionCount)

	// Add each frame
	for i, frame := range frames {
		width := frame.Bounds().Dx()
		height := frame.Bounds().Dy()
		stride := width * 4
		pixels := make([]byte, height*stride)

		for y := 0; y < height; y++ {
			srcOffset := y * frame.Stride
			dstOffset := y * stride
			copy(pixels[dstOffset:dstOffset+stride], frame.Pix[srcOffset:srcOffset+stride])
		}

		avifImage, err := createAVIFTile(pixels, width, height, stride, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to create AVIF image for frame %d: %w", i, err)
		}

		result := C.avifEncoderAddImage(encoder, avifImage, C.uint64_t(delaysMs[i]), C.AVIF_ADD_IMAGE_FLAG_NONE)
		C.avifImageDestroy(avifImage)

		if result != C.AVIF_RESULT_OK {
			errStr := C.GoString(C.get_error_string(result))
			return nil, fmt.Errorf("failed to add frame %d: %s", i, errStr)
		}
	}

	// Finish encoding
	var encodedData C.avifRWData
	encodedData.data = nil
	encodedData.size = 0

	result := C.avifEncoderFinish(encoder, &encodedData)
	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return nil, fmt.Errorf("failed to finish encoding: %s", errStr)
	}
	defer C.avifRWDataFree(&encodedData)

	data := C.GoBytes(unsafe.Pointer(encodedData.data), C.int(encodedData.size))
	return data, nil
}

// createTiles splits the input RGBA image into tiles and converts them to AVIF format.
// Returns a slice of avifImage pointers that must be freed by the caller.
func createTiles(rgba image.RGBA, tileWidth, tileHeight int) ([]*C.avifImage, error) {
	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()

	cols := (width + tileWidth - 1) / tileWidth
	rows := (height + tileHeight - 1) / tileHeight

	// Pre-allocate slice with exact capacity
	cellImages := make([]*C.avifImage, 0, cols*rows)

	// Pre-allocate tile buffer once and reuse, sized to actual needs
	actualTileW := min(tileWidth, width)
	actualTileH := min(tileHeight, height)
	tileBuffer := make([]byte, actualTileW*actualTileH*4)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			// Calculate tile boundaries
			x0 := col * tileWidth
			y0 := row * tileHeight
			x1 := x0 + tileWidth
			y1 := y0 + tileHeight
			if x1 > width {
				x1 = width
			}
			if y1 > height {
				y1 = height
			}

			tileW := x1 - x0
			tileH := y1 - y0

			// Use pre-allocated buffer
			stride := tileW * 4
			tileSize := tileH * stride
			tilePix := tileBuffer[:tileSize]

			// Copy rows
			for y := 0; y < tileH; y++ {
				srcY := y0 + y
				dstOffset := y * stride
				srcOffset := srcY*rgba.Stride + x0*4
				copy(tilePix[dstOffset:dstOffset+stride], rgba.Pix[srcOffset:srcOffset+stride])
			}

			// Create and convert tile
			avifImage, err := createAVIFTile(tilePix, tileW, tileH, stride, col, row)
			if err != nil {
				// Clean up already created tiles
				for _, img := range cellImages {
					if img != nil {
						C.avifImageDestroy(img)
					}
				}
				return nil, err
			}

			cellImages = append(cellImages, avifImage)
		}
	}

	return cellImages, nil
}

// createAVIFTile creates an avifImage from raw RGBA pixel data.
func createAVIFTile(pixels []byte, width, height, stride, col, row int) (*C.avifImage, error) {
	avifImage := C.avifImageCreate(C.uint32_t(width), C.uint32_t(height), 8, C.AVIF_PIXEL_FORMAT_YUV420)
	if avifImage == nil {
		return nil, fmt.Errorf("failed to create AVIF image for tile (%d,%d)", col, row)
	}

	// Convert to YUV
	rgb := (*C.avifRGBImage)(C.malloc(C.size_t(unsafe.Sizeof(C.avifRGBImage{}))))
	if rgb == nil {
		C.avifImageDestroy(avifImage)
		return nil, fmt.Errorf("failed to allocate avifRGBImage for tile (%d,%d)", col, row)
	}

	C.avifRGBImageSetDefaults(rgb, avifImage)
	rgb.format = C.AVIF_RGB_FORMAT_RGBA
	rgb.depth = 8
	rgb.pixels = (*C.uint8_t)(unsafe.Pointer(&pixels[0]))
	rgb.rowBytes = C.uint32_t(stride)

	result := C.avifImageRGBToYUV(avifImage, rgb)
	C.free(unsafe.Pointer(rgb))

	if result != C.AVIF_RESULT_OK {
		C.avifImageDestroy(avifImage)
		return nil, fmt.Errorf("failed to convert tile (%d,%d) from RGB to YUV", col, row)
	}

	return avifImage, nil
}

// decodeAVIFToRGBA decodes the first frame of AVIF image data to an RGBA image.
func decodeAVIFToRGBA(data []byte) (*image.RGBA, error) {
	frames, _, _, err := decodeAllAVIFToRGBA(data)
	if err != nil {
		return nil, err
	}
	return frames[0], nil
}

// decodeAllAVIFToRGBA decodes all frames from an AVIF image sequence.
// Returns the frames as RGBA images, delays in centiseconds, and the repetition count.
func decodeAllAVIFToRGBA(data []byte) ([]*image.RGBA, []int, int, error) {
	if len(data) == 0 {
		return nil, nil, 0, fmt.Errorf("cannot decode empty data")
	}

	cData := C.CBytes(data)
	defer C.free(cData)

	// Create and configure decoder
	decoder := C.avifDecoderCreate()
	if decoder == nil {
		return nil, nil, 0, fmt.Errorf("failed to create AVIF decoder")
	}
	defer C.avifDecoderDestroy(decoder)

	decoder.codecChoice = C.AVIF_CODEC_CHOICE_DAV1D

	result := C.avifDecoderSetIOMemory(decoder, (*C.uint8_t)(cData), C.size_t(len(data)))
	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return nil, nil, 0, fmt.Errorf("failed to set decoder I/O: %s", errStr)
	}

	result = C.avifDecoderParse(decoder)
	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return nil, nil, 0, fmt.Errorf("failed to parse AVIF: %s", errStr)
	}

	frameCount := int(decoder.imageCount)
	repetitionCount := int(decoder.repetitionCount)

	frames := make([]*image.RGBA, 0, frameCount)
	delays := make([]int, 0, frameCount)

	for i := 0; i < frameCount; i++ {
		result = C.avifDecoderNextImage(decoder)
		if result != C.AVIF_RESULT_OK {
			errStr := C.GoString(C.get_error_string(result))
			return nil, nil, 0, fmt.Errorf("failed to decode frame %d: %s", i, errStr)
		}

		img, err := avifImageToRGBA(decoder.image)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("frame %d: %w", i, err)
		}
		frames = append(frames, img)

		// Get frame timing and convert to centiseconds
		var timing C.avifImageTiming
		C.avifDecoderNthImageTiming(decoder, C.uint32_t(i), &timing)
		delay := 0
		if timing.timescale > 0 {
			delay = int(timing.durationInTimescales * 100 / timing.timescale)
		}
		delays = append(delays, delay)
	}

	return frames, delays, repetitionCount, nil
}

// avifImageToRGBA converts a C avifImage (YUV) to a Go image.RGBA.
func avifImageToRGBA(avifImg *C.avifImage) (*image.RGBA, error) {
	var rgb C.avifRGBImage
	C.avifRGBImageSetDefaults(&rgb, avifImg)
	rgb.format = C.AVIF_RGB_FORMAT_RGBA
	rgb.depth = 8

	if C.avifRGBImageAllocatePixels(&rgb) != C.AVIF_RESULT_OK {
		return nil, fmt.Errorf("failed to allocate RGB pixels")
	}
	defer C.avifRGBImageFreePixels(&rgb)

	result := C.avifImageYUVToRGB(avifImg, &rgb)
	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return nil, fmt.Errorf("failed to convert YUV to RGB: %s", errStr)
	}

	width := int(avifImg.width)
	height := int(avifImg.height)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rowBytes := int(rgb.rowBytes)

	for y := 0; y < height; y++ {
		srcPtr := unsafe.Add(unsafe.Pointer(rgb.pixels), y*rowBytes)
		dstOffset := y * img.Stride
		copy(img.Pix[dstOffset:dstOffset+4*width],
			unsafe.Slice((*byte)(srcPtr), 4*width))
	}

	return img, nil
}

// decodeConfig reads enough of the data to determine the image's configuration (dimensions, etc.).
//
// This is a lightweight operation that only parses the header.
func decodeConfig(data []byte) (image.Config, error) {
	if len(data) == 0 {
		return image.Config{}, fmt.Errorf("failed to get AVIF image config: empty data")
	}

	// Use C.CBytes for safer memory handling
	cData := C.CBytes(data)
	defer C.free(cData)

	var width, height C.uint32_t
	var result C.avifResult
	C.get_avif_config((*C.uint8_t)(cData), C.size_t(len(data)), &width, &height, &result)

	if result != C.AVIF_RESULT_OK {
		errStr := C.GoString(C.get_error_string(result))
		return image.Config{}, fmt.Errorf("failed to get AVIF image config: %s", errStr)
	}

	if width == 0 || height == 0 {
		return image.Config{}, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	// We assume an RGBA color model for simplicity.
	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      int(width),
		Height:     int(height),
	}, nil
}
