package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
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

	ext := strings.ToLower(filepath.Ext(input))

	// Animated GIF path
	if ext == ".gif" {
		gifData, err := gif.DecodeAll(inputFile)
		if err != nil {
			return nil, nil, err
		}

		// It has more than 1 frame, so it's an image sequence (animated)
		if len(gifData.Image) > 1 {
			a := composeGIFFrames(gifData)

			outputFile, err := os.Create(output)
			if err != nil {
				return nil, nil, err
			}
			defer outputFile.Close()

			err = avif.EncodeAll(outputFile, a, options)
			if err != nil {
				return nil, nil, err
			}

			info, err := outputFile.Stat()
			if err != nil {
				return nil, nil, err
			}

			return a.Image[0], info, nil
		}

		// Single-frame GIF: encode as still image
		outputFile, err := os.Create(output)
		if err != nil {
			return nil, nil, err
		}
		defer outputFile.Close()

		err = avif.Encode(outputFile, gifData.Image[0], options)
		if err != nil {
			return nil, nil, err
		}

		info, err := outputFile.Stat()
		if err != nil {
			return nil, nil, err
		}

		return gifData.Image[0], info, nil
	}

	// Non-GIF path: existing behavior
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

// composeGIFFrames composites a decoded animated GIF into an AVIF struct with
// full-size RGBA frames and timing information, handling disposal methods correctly.
func composeGIFFrames(g *gif.GIF) *avif.AVIF {
	width := g.Config.Width
	height := g.Config.Height

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	var backup *image.RGBA

	frames := make([]image.Image, len(g.Image))
	delays := make([]int, len(g.Image))

	for i, frame := range g.Image {
		// Apply previous frame's disposal method
		if i > 0 {
			prevDisposal := byte(0)
			if i-1 < len(g.Disposal) {
				prevDisposal = g.Disposal[i-1]
			}

			switch prevDisposal {
			case gif.DisposalBackground:
				prevBounds := g.Image[i-1].Bounds()
				draw.Draw(canvas, prevBounds, image.NewUniform(color.Transparent), image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				if backup != nil {
					copy(canvas.Pix, backup.Pix)
				}
			}
		}

		// Save backup before drawing if current frame's disposal is DisposalPrevious
		disposal := byte(0)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		if disposal == gif.DisposalPrevious {
			backup = image.NewRGBA(canvas.Bounds())
			copy(backup.Pix, canvas.Pix)
		}

		// Draw current frame onto canvas
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		// Clone canvas as this frame
		cloned := image.NewRGBA(canvas.Bounds())
		copy(cloned.Pix, canvas.Pix)
		frames[i] = cloned

		// Use delay directly (both GIF and AVIF.Delay use centiseconds)
		if i < len(g.Delay) {
			delays[i] = g.Delay[i]
		}
	}

	return &avif.AVIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: g.LoopCount,
	}
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

	// Decode all frames from the AVIF
	a, err := avif.DecodeAll(inputFile)
	if err != nil {
		return nil, nil, err
	}

	outputFile, err := os.Create(output)
	if err != nil {
		return nil, nil, err
	}
	defer outputFile.Close()

	// Animated AVIF → animated GIF
	if ext == ".gif" && len(a.Image) > 1 {
		g := &gif.GIF{
			LoopCount: a.LoopCount,
			Delay:     a.Delay,
			Image:     make([]*image.Paletted, len(a.Image)),
		}

		for i, frame := range a.Image {
			bounds := frame.Bounds()
			paletted := image.NewPaletted(bounds, palette.Plan9)
			draw.FloydSteinberg.Draw(paletted, bounds, frame, bounds.Min)
			g.Image[i] = paletted
		}

		err = gif.EncodeAll(outputFile, g)
	} else {
		// Single frame or non-GIF output: use first frame
		img := a.Image[0]

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
	}

	if err != nil {
		return nil, nil, err
	}

	info, err := outputFile.Stat()
	if err != nil {
		return nil, nil, err
	}

	return a.Image[0], info, nil
}
