//go:build darwin && arm64

package avif

/*
#cgo CFLAGS: -I./include
#cgo LDFLAGS: -L./libs/darwin_arm64 -lavif -ldav1d -ljpeg -lturbojpeg -lyuv
*/
import "C"
