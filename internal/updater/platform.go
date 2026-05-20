package updater

import (
	"fmt"
	"runtime"
)

var osMap = map[string]string{
	"linux":   "Linux",
	"darwin":  "Darwin",
	"windows": "Windows",
	"freebsd": "FreeBSD",
}

var archMap = map[string]string{
	"amd64": "x86_64",
	"arm64": "aarch64",
	"386":   "i386",
	"arm":   "armv7",
}

func PlatformArchiveName() (string, error) {
	osLabel, ok := osMap[runtime.GOOS]
	if !ok {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	archLabel, ok := archMap[runtime.GOARCH]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	return fmt.Sprintf("xxaitk_%s_%s", osLabel, archLabel), nil
}

func ArchiveExt() string {
	if runtime.GOOS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "aitk.exe"
	}
	return "aitk"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}
