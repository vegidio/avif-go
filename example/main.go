package main

import (
	"fmt"
	_ "github.com/vegidio/avif-go"
	"image"
	"os"
)

func main() {
	file, err := os.Open("assets/image.avif")
	if err != nil {
		fmt.Println("failed to open file -", err)
		return
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("failed to decode AVIF -", err)
		return
	}

	fmt.Printf("Decoded image: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
}
