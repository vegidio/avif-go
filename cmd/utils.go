package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/vegidio/avif-go"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var ValidImageTypes = []string{".bmp", ".gif", ".jpg", ".jpeg", ".png", ".tiff"}

func encodeAvif(input, output string, options *avif.Options) (image.Image, os.FileInfo, error) {
	inputFile, err := os.Open(input)
	if err != nil {
		return nil, nil, err
	}

	defer inputFile.Close()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		return nil, nil, err
	}

	outputFile, err := os.Create(output)
	if err != nil {
		return nil, nil, err
	}

	defer outputFile.Close()

	err = avif.Encode(outputFile, img, options)
	if err != nil {
		return nil, nil, err
	}

	info, err := outputFile.Stat()
	if err != nil {
		return nil, nil, err
	}

	return img, info, nil
}

func decodeAvif(input, output string) (image.Image, os.FileInfo, error) {
	ext := strings.ToLower(filepath.Ext(output))
	if !slices.Contains(ValidImageTypes, ext) {
		return nil, nil, fmt.Errorf("invalid output file type: %s", ext)
	}

	inputFile, err := os.Open(input)
	if err != nil {
		return nil, nil, err
	}

	defer inputFile.Close()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		return nil, nil, err
	}

	outputFile, err := os.Create(output)
	if err != nil {
		return nil, nil, err
	}

	defer outputFile.Close()

	switch ext {
	case ".bmp":
		err = bmp.Encode(outputFile, img)
	case ".gif":
		err = gif.Encode(outputFile, img, nil)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outputFile, img, nil)
	case ".png":
		err = png.Encode(outputFile, img)
	case ".tiff":
		err = tiff.Encode(outputFile, img, nil)
	}

	if err != nil {
		return nil, nil, err
	}

	info, err := outputFile.Stat()
	if err != nil {
		return nil, nil, err
	}

	return img, info, nil
}
