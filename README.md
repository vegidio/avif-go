# avif-go

A Go encoder/decoder for AVIF without external dependencies (CGO).

## 💡 Motivation

There are a couple of libraries to encode/decode AVIF images in Go, and even though they do the job well, they have some limitations that don't satisfy my needs:

- They either depend on external libraries to be installed in the system in order to be built and/or later be executed.
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

### The AVIF encoding is taking too long

This build of `libavif` uses the reference implementation `libaom` to encode images. Even though it offers the best image quality, it is also the slowest encoder available.

There are other AVIF implementations that encode much quicker, like `svt-av1`, but their builds are not easily available. My goal is to replace `libam` with `svt-av1` in the future, but for now we will have to deal with the slow encoding.

Decoding on the other hand should not be a problem, since this library uses `dav1d` which is the fastest AV1 decoder available.

## Building

To build the `libavif` static binaries used by this library, you need:

- Install [vcpkg](https://github.com/microsoft/vcpkg) in your computer.
- Copy the triplets from the folder `vcpkg/` to `$VCPKG_ROOT/triplets/community`.
- Run the necessary command depending which platform you are building for:
  - darwin/amd64: `vcpkg install "libavif[aom,dav1d]:darwin-amd64-static"`
  - darwin/arm64: `vcpkg install "libavif[aom,dav1d]:darwin-arm64-static"`
  - linux/amd64: `vcpkg install "libavif[aom,dav1d]:linux-amd64-static"`
  - linux/arm64: `vcpkg install "libavif[aom,dav1d]:linux-arm64-static"`
  - windows/amd64: `vcpkg install "libavif[aom,dav1d]:windows-amd64-static"`

## 📝 License

**avif-go** is released under the MIT License. See [LICENSE](LICENSE) for details.

## 👨🏾‍💻 Author

Vinicius Egidio ([vinicius.io](http://vinicius.io))
