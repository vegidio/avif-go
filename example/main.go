package main

import (
	"fmt"
	"github.com/vegidio/avif-go"
	_ "github.com/vegidio/avif-go"
	"image"
	_ "image/jpeg"
	"os"
	"time"
)

func main() {
	// Encode a JPEG image to AVIF
	encodeAvif("assets/image.jpg", "assets/image.avif")

	// Decode an AVIF image
	decodeAvif("assets/image.avif")
}

func encodeAvif(inputFile, outputFile string) {
	jpgFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("failed to open JPEG -", err)
		return
	}

	defer jpgFile.Close()

	jpgImg, _, err := image.Decode(jpgFile)
	if err != nil {
		fmt.Println("failed to decode JPEG -", err)
		return
	}

	avifFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("failed to create AVIF -", err)
		return
	}

	defer avifFile.Close()

	start := time.Now()

	err = avif.Encode(avifFile, jpgImg, nil)
	if err != nil {
		fmt.Println("failed to encode AVIF -", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("Encoding completed in %s\n", duration)
}

func decodeAvif(inputFile string) {
	avifFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("failed to open AVIF -", err)
		return
	}

	defer avifFile.Close()

	start := time.Now()

	avifImage, _, err := image.Decode(avifFile)
	if err != nil {
		fmt.Println("failed to decode AVIF -", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("Decoding completed in %s - decoded AVIF image: %dx%d\n", duration, avifImage.Bounds().Dx(),
		avifImage.Bounds().Dy())
}
