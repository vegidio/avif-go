package main

import (
	avif "avif-go"
	"fmt"
)

func main() {
	fmt.Println("Combined libavif library is linked!")

	img, err := avif.DecodeAvifToImage("/Users/vegidio/Development/Source/avif-go/cmd/image.avif")
	if err != nil {
		fmt.Println("Error decoding AVIF:", err)
		return
	}

	fmt.Printf("Decoded image: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
}
