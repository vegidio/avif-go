package avif

/*
#include <stdlib.h>
#include <avif/avif.h>

// A simple wrapper to decode an AVIF image from memory.
// Returns a pointer to the decoded avifImage (owned by the decoder)
// and assigns the decoder to outDecoder (so you can clean it up later).
avifImage* decode_avif_image(const uint8_t * data, size_t size, avifDecoder ** outDecoder) {
    avifDecoder* decoder = avifDecoderCreate();
    if(avifDecoderSetIOMemory(decoder, data, size) != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return NULL;
    }
    if(avifDecoderParse(decoder) != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return NULL;
    }
    if(avifDecoderNextImage(decoder) != AVIF_RESULT_OK) {
        avifDecoderDestroy(decoder);
        return NULL;
    }
    if(outDecoder) {
        *outDecoder = decoder;
    }
    return decoder->image;
}
*/
import "C"
import (
	"fmt"
	"image"
	"os"
	"unsafe"
)

func DecodeAvifToImage(path string) (*image.RGBA, error) {
	// Read the entire file into a byte slice using os.ReadFile.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Allocate C memory for the file data.
	cData := C.CBytes(data)
	defer C.free(cData)

	// Decode the AVIF image.
	var decoder *C.avifDecoder
	avifImg := C.decode_avif_image((*C.uint8_t)(cData), C.size_t(len(data)), &decoder)
	if avifImg == nil {
		return nil, fmt.Errorf("failed to decode AVIF image")
	}

	// Set up an avifRGBImage to hold the converted RGBA data.
	var rgb C.avifRGBImage
	C.avifRGBImageSetDefaults(&rgb, avifImg)
	rgb.format = C.AVIF_RGB_FORMAT_RGBA
	rgb.depth = 8 // use 8-bit per channel

	// Allocate memory for the RGB pixels.
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

	// Create a Go image.RGBA to hold the final image.
	width := int(avifImg.width)
	height := int(avifImg.height)
	// Go’s RGBA uses 4 bytes per pixel.
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// The rowBytes in rgb is the number of bytes per row in the converted image.
	rowBytes := int(rgb.rowBytes)
	// Copy pixel data row by row.
	srcPtr := unsafe.Pointer(rgb.pixels)
	srcSlice := C.GoBytes(srcPtr, C.int(height*rowBytes))
	for y := 0; y < height; y++ {
		srcOffset := y * rowBytes
		dstOffset := y * img.Stride
		// Ensure we only copy up to the width in bytes (4 * width)
		copy(img.Pix[dstOffset:dstOffset+4*width], srcSlice[srcOffset:srcOffset+4*width])
	}

	// Free resources allocated by libavif.
	C.avifRGBImageFreePixels(&rgb)
	C.avifDecoderDestroy(decoder)

	return img, nil
}
