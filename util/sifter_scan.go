package util

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/playbook"
)

func ScanSifter(baseDir string, userFunc func(*playbook.Playbook)) {
	filepath.Walk(baseDir,
		func(path string, info fs.FileInfo, err error) error {
			if strings.HasSuffix(path, ".yaml") {
				pb := playbook.Playbook{}
				if parseErr := playbook.ParseFile(path, &pb); parseErr == nil {
					userFunc(&pb)
				}
			}
			return nil
		})
}