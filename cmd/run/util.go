package run

import "os"

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err != os.ErrNotExist
}
