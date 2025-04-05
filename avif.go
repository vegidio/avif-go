package avif

/*
#include <stdlib.h>
#include <avif/avif.h>

// Full decode: creates a decoder, sets up the memory I/O, and decodes the image.
// Returns the avifImage pointer (which contains width, height, etc.) and leaves the
// decoder pointer for cleanup.
avifImage* decode_avif_image(const uint8_t * data, size_t size, avifDecoder ** outDecoder) {
    avifDecoder* decoder = avifDecoderCreate();
    if (avifDecoderSetIOMemory(decoder, data, size) != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return NULL;
    }
    if (avifDecoderParse(decoder) != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return NULL;
    }
    if (avifDecoderNextImage(decoder) != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return NULL;
    }
    if (outDecoder) {
        *outDecoder = decoder;
    }
    return decoder->image;
}

// Config-only decode: reads the header and returns width and height.
void get_avif_config(const uint8_t * data, size_t size, uint32_t * width, uint32_t * height) {
    avifDecoder* decoder = avifDecoderCreate();
    if (avifDecoderSetIOMemory(decoder, data, size) != AVIF_RESULT_OK) {
         *width = 0;
         *height = 0;
         avifDecoderDestroy(decoder);
         return;
    }
    if (avifDecoderParse(decoder) != AVIF_RESULT_OK) {
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

func decodeAVIFToRGBA(data []byte) (*image.RGBA, error) {
	// Allocate C memory and copy data.
	cData := C.CBytes(data)
	defer C.free(cData)

	var decoder *C.avifDecoder
	avifImg := C.decode_avif_image((*C.uint8_t)(cData), C.size_t(len(data)), &decoder)
	if avifImg == nil {
		return nil, fmt.Errorf("failed to decode AVIF image")
	}

	// Set up an avifRGBImage struct to hold the converted image.
	var rgb C.avifRGBImage
	C.avifRGBImageSetDefaults(&rgb, avifImg)
	rgb.format = C.AVIF_RGB_FORMAT_RGBA
	rgb.depth = 8 // 8-bit per channel

	// Allocate pixel buffer for the RGB data.
	if C.avifRGBImageAllocatePixels(&rgb) != C.AVIF_RESULT_OK {
		C.avifDecoderDestroy(decoder)
		return nil, fmt.Errorf("failed to allocate RGB pixels")
	}

	// Convert the image from YUV to RGB.
	if C.avifImageYUVToRGB(avifImg, &rgb) != C.AVIF_RESULT_OK {
		C.avifRGBImageFreePixels(&rgb)
		C.avifDecoderDestroy(decoder)
		return nil, fmt.Errorf("failed to convert image to RGB")
	}

	width := int(avifImg.width)
	height := int(avifImg.height)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rowBytes := int(rgb.rowBytes)

	// Copy the pixel data row by row into the Go image.
	srcPtr := unsafe.Pointer(rgb.pixels)
	srcSlice := C.GoBytes(srcPtr, C.int(height*rowBytes))
	for y := 0; y < height; y++ {
		srcOffset := y * rowBytes
		dstOffset := y * img.Stride
		copy(img.Pix[dstOffset:dstOffset+4*width], srcSlice[srcOffset:srcOffset+4*width])
	}

	// Free C resources.
	C.avifRGBImageFreePixels(&rgb)
	C.avifDecoderDestroy(decoder)

	return img, nil
}

// DecodeConfig reads enough of r to determine the image's configuration (dimensions, etc.).
// Here we read the entire data and call a lightweight C function that only parses the header.
func decodeConfig(data []byte) (image.Config, error) {
	var width, height C.uint32_t
	C.get_avif_config((*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), &width, &height)

	if width == 0 || height == 0 {
		return image.Config{}, fmt.Errorf("failed to get AVIF image config")
	}

	// We assume an RGBA color model for simplicity.
	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      int(width),
		Height:     int(height),
	}, nil
}
