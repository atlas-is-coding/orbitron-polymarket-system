package updater

import "runtime"

// buildOS and buildArch reflect the OS/arch the binary was compiled for.
var (
	buildOS   = runtime.GOOS
	buildArch = runtime.GOARCH
)

func binaryKey() string {
	return buildOS + "_" + buildArch
}
