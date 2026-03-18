# avif-go

A Go library & CLI tool to encode/decode static and animated AVIF without external dependencies.

## 💡 Motivation

There are a couple of libraries to encode/decode AVIF images in Go, and even though they do the job well, they have some limitations that don't satisfy my needs:

- They need dependencies to be installed on the system to either build the app or later execute it.
- They rely on a WASM runtime - which is actually a really smart idea! - but it has a big impact on performance.
- There's currently no Go library capable of encoding animated AVIFs; this library aims to fill that gap.

**avif-go** uses CGO to create a static implementation of AVIF, so you don't need `libavif` (or any of its sub-dependencies) installed to build or run your Go application.

It also runs on native code (supports `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`), so it achieves the best performance possible.

## ⬇️ Installation

### Library

This library can be installed using Go modules. To do that, run the following command in your project's root directory:

```bash
$ go get github.com/vegidio/avif-go
```

### CLI

The binaries are available for Windows, macOS, and Linux. Download the [latest release](https://github.com/vegidio/avif-go/releases) that matches your computer architecture and operating system.

## 🤖 Usage

### Library

This is a CGO library, so to use it, you _must_ enable CGO while building your application. You can do that by setting the `CGO_ENABLED` environment variable to `1`:

```bash
$ CGO_ENABLED=1 go build /path/to/your/app.go
```

Here are some examples of how to encode and decode AVIF images using this library. These snippets don't have any error handling for the sake of simplicity, but you should always check for errors in production code.

#### Encoding

```go
var originalImage image.Image = ... // an image.Image to be encoded
avifFile, err := os.Create("/path/to/image.avif") // create the file to save the AVIF
err = avif.Encode(avifFile, originalImage, nil) // encode the image & save it to the file
```

#### Encoding Animation

```go
var frames []image.Image = ... // a slice of image.Image frames
avifFile, err := os.Create("/path/to/animation.avif") // create file to save the animation
a := &avif.AVIF{Image: frames, Delay: []int{10, 10, 10}, LoopCount: 0} // anim configuration
err = avif.EncodeAll(avifFile, a, nil) // encode the animation & save it to the file
```

#### Decoding

```go
import _ "github.com/vegidio/avif-go" // do a blank import to register the AVIF decoder
avifFile, err := os.Open("/path/to/image.avif") // open the AVIF file to be decoded
avifImage, _, err := image.Decode(avifFile) // decode the image
```

#### Decoding Animation

```go
avifFile, err := os.Open("/path/to/animation.avif") // open the animated AVIF file
a, err := avif.DecodeAll(avifFile) // decode all frames
// a.Image contains the frames, a.Delay the timing, a.LoopCount the loop behavior
```

For more examples on how to use this library, you can check the [cmd/](cmd) directory.

### CLI

If you want to decode an AVIF image, run the following command:

```bash
$ avif decode /path/to/image.avif /path/to/image.png
```

---

To encode an image to AVIF, run the following command:

```bash
$ avif encode /path/to/image.png /path/to/image.avif
```

For the full list of parameters, type `avif encode --help` in the terminal.

## 💣 Troubleshooting

### I cannot build my app after importing this library

If you cannot build your app after importing **avif-go**, it is probably because you didn't set the `CGO_ENABLED` environment variable to `1`.

You must either set a global environment variable with `export CGO_ENABLED=1` or set it in the command line when building your app with `CGO_ENABLED=1 go build /path/to/your/app.go`.

### "App Is Damaged/Blocked..." (Windows & macOS only)

For a couple of years now, Microsoft and Apple have required developers to join their "Developer Program" to gain the pretentious status of an _identified developer_ 😛.

Translating to non-BS language, this means that if you’re not registered with them (i.e., paying the fee), you can’t freely distribute Windows or macOS software. Apps from unidentified developers will display a message saying the app is damaged or blocked and can’t be opened.

To bypass this, open the Terminal and run one of the commands below (depending on your operating system), replacing `<path-to-app>` with the correct path to where you’ve installed the app:

- Windows: `Unblock-File -Path <path-to-app>`
- macOS: `xattr -d com.apple.quarantine <path-to-app>`

## 📝 License

**avif-go** is released under the Apache 2.0 License. See [LICENSE](LICENSE) for details.

## 👨🏾‍💻 Author

Vinicius Egidio ([vinicius.io](http://vinicius.io))
