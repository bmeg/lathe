package workflow

import (
	"errors"
	"os"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func IsFile(path string) bool {
	s, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return !s.IsDir()
}
