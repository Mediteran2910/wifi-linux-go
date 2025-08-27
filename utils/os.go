package utils

import "runtime"

// IsLinux checks if the current operating system is Linux.
// The function name starts with a capital letter to make it public (exported).
func IsLinux() bool {
	return runtime.GOOS == "linux"
}