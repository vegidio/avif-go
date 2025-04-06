# avif-go

## Building

Install `vcpkg` and run the command below to create the static library:

```bash
MACOSX_DEPLOYMENT_TARGET=14 vcpkg install "libavif[aom,dav1d]:darwin-arm64-static"
```

Now you can build the Go binary with the command below:

```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build ./example/main.go
```