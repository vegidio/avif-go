# avif-go

A Go encoder/decoder for AVIF without system dependencies (CGO).

## 💡 Motivation

There are a couple of libraries to encode/decode AVIF images in Go, and even though they do the job well, they have some limitations that don't satisfy my needs:

- They either depend on libraries to be installed on the system in order to be built and/or later be executed.
- They rely on a WASM runtime - which is actually a really smart idea! - but it has a big impact in performance.

**avif-go** uses CGO to create a static implementation of AVIF, so you don't need to have `libavif` (or any of its sub-dependencies) installed to build or run your Go application.

It also runs on native code (supports `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`), so it achieves the best performance possible.

## ⬇️ Installation

This library can be installed using Go modules. To do that run the following command in your project's root directory:

```bash
$ go get github.com/vegidio/avif-go
```

## 🤖 Usage

This is a CGO library so in order to use it you _must_ enable CGO while building your application. You can do that by setting the `CGO_ENABLED` environment variable to `1`:

```bash
$ CGO_ENABLED=1 go build /path/to/your/app.go
```

Here are some examples of how to encode and decode AVIF images using this library. These snippets don't have any error handling for the sake of simplicity, but you should always check for errors in production code.

For more complete examples, check the [example](example) folder.

### Encoding

```go
var originalImage image.Image = ... // an image.Image to be encoded
avifFile, err := os.Create("/path/to/image.avif") // create the file to save the AVIF
err = avif.Encode(avifFile, originalImage) // encode the image and save it to the file
```

### Decoding

```go
import _ "github.com/vegidio/avif-go" // do a blank import to register the AVIF decoder
avifFile, err := os.Open("/path/to/image.avif") // open the AVIF file to be decoded
avifImage, _, err := image.Decode(avifFile) // decode the image
```

## 💣 Troubleshooting

### I cannot build my app after importing this library

If you cannot build your app after importing **avif-go**, it is probably because you didn't set the `CGO_ENABLED` environment variable to `1`.

You must either set a global environment variable with `export CGO_ENABLED=1` or set it in the command line when building your app with `CGO_ENABLED=1 go build /path/to/your/app.go`.

## 📝 License

**avif-go** is released under the MIT License. See [LICENSE](LICENSE) for details.

## 👨🏾‍💻 Author

Vinicius Egidio ([vinicius.io](http://vinicius.io))
