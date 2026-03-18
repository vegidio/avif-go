package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vegidio/avif-go"
)

func main() {
	var speed uint
	var alphaQuality uint
	var colorQuality uint

	cmd := &cli.Command{
		Name:            "avif",
		Usage:           "a tool to encode & decode AVIF images",
		UsageText:       "avif <enc|dec> <input> <output>",
		Version:         "<version>",
		HideHelpCommand: true,
		Commands: []*cli.Command{
			{
				Name:      "encode",
				Aliases:   []string{"enc"},
				Usage:     "encode an image to AVIF",
				UsageText: "avif enc <input> <output>",
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:        "speed",
						Aliases:     []string{"s"},
						Usage:       "encoding speed between 0-10; higher values result in faster encoding but lower quality.",
						Value:       6,
						DefaultText: "6",
						Destination: &speed,
						Required:    false,
					},
					&cli.UintFlag{
						Name:        "alpha-quality",
						Aliases:     []string{"a"},
						Usage:       "alpha quality between 0-100; higher values result in better quality.",
						Value:       60,
						DefaultText: "60",
						Destination: &alphaQuality,
						Required:    false,
					},
					&cli.UintFlag{
						Name:        "color-quality",
						Aliases:     []string{"c"},
						Usage:       "color quality between 0-100; higher values result in better quality.",
						Value:       60,
						DefaultText: "60",
						Destination: &colorQuality,
						Required:    false,
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					if command.Args().Len() < 1 {
						return fmt.Errorf("missing input file")
					}
					if command.Args().Len() < 2 {
						return fmt.Errorf("missing output file")
					}

					input := command.Args().First()
					output := command.Args().Tail()[0]

					options := &avif.Options{
						Speed:        int(speed),
						AlphaQuality: int(alphaQuality),
						ColorQuality: int(colorQuality),
					}

					now := time.Now()
					img, info, err := encodeAvif(input, output, options)
					duration := time.Since(now)

					if err == nil {
						printResult(img, info, duration, true)
					}

					return err
				},
			},
			{
				Name:      "decode",
				Aliases:   []string{"dec"},
				Usage:     "decode an AVIF image to a different format",
				UsageText: "avif dec <input> <output>",
				Action: func(ctx context.Context, command *cli.Command) error {
					if command.Args().Len() < 1 {
						return fmt.Errorf("missing input file")
					}
					if command.Args().Len() < 2 {
						return fmt.Errorf("missing output file")
					}

					input := command.Args().First()
					output := command.Args().Tail()[0]

					now := time.Now()
					img, info, err := decodeAvif(input, output)
					duration := time.Since(now)

					if err == nil {
						printResult(img, info, duration, false)
					}

					return err
				},
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			return fmt.Errorf("either the command <encode> or <decode> must be used")
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		msg := fmt.Sprintf("🧨 %v", err)
		fmt.Println(red.Render(msg))
	}
}

func printResult(img image.Image, info os.FileInfo, duration time.Duration, isEncode bool) {
	cmd := "decoded"
	if isEncode {
		cmd = "encoded"
	}

	msg := fmt.Sprintf("✅ Successfully %s image to %s in %s",
		cmd, info.Name(), duration.Truncate(time.Millisecond))
	fmt.Println(green.Render(msg))

	msg = fmt.Sprintf("🖼 Image dimensions: %dx%d; size: %d bytes",
		img.Bounds().Dx(), img.Bounds().Dy(), info.Size())
	fmt.Println(yellow.Render(msg))
}
