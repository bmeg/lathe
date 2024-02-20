package scriptfile

import "path/filepath"

type FileCheck struct {
	File *File
}

func (fc *FileCheck) GetName() string {
	b := filepath.Dir(fc.File.BasePath)
	p := filepath.Join(b, fc.File.Path)
	return p
}

func (fc *FileCheck) GetBasePath() string {
	return fc.File.BasePath
}

func (fc *FileCheck) GetInputs() map[string]string {
	return map[string]string{
		"file": fc.File.Path,
	}
}

func (fc *FileCheck) GetProcess() *ProcessDesc {
	return nil
}
