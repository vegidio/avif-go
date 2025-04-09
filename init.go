package avif

import "os"

func init() {
	// Disable SVT‑AV1 logs by setting the environment variable
	os.Setenv("SVT_LOG", "-1")
}
