package expath

import (
	"os"
	"path/filepath"
	"runtime"
)

func Get() (string, error) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		return os.Getwd()
	}

	return unixExPath()
}

func unixExPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	exReal, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return "", err
	}
	dirAbsPath := filepath.Dir(exReal)

	return dirAbsPath, nil
}
