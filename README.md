# avif-go

## Building

Install `vcpkg` and run the command below to create the static library:

```bash
MACOSX_DEPLOYMENT_TARGET=15.0 vcpkg install "libavif[dav1d]:arm64-osx-static"
```

Now you can build the Go binary with the command below:

```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build ./example/main.go
```