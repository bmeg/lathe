package scriptfile

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// Parse parses a YAML doc into the given Config instance.
func parse(raw []byte, conf *ScriptFile) error {
	return yaml.UnmarshalStrict(raw, conf)
}

// ParseFile parses a Sifter playbook file, which is formatted in YAML,
// and returns a Playbook struct.
func ParseFile(relpath string, conf *ScriptFile) error {
	if relpath == "" {
		return nil
	}

	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(relpath)
	if abserr != nil {
		path = relpath
	}

	// Read file
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config at path %s: \n%v", path, err)
	}

	// Parse file
	err = parse(source, conf)
	if err != nil {
		return fmt.Errorf("failed to parse config at path %s: \n%v", path, err)
	}

	conf.path = path

	for k, v := range conf.Scripts {
		v.name = k
		v.path = path
	}

	for _, v := range conf.Prep {
		v.path = path
	}

	for i := range conf.Collections {
		conf.Collections[i].path = path
	}

	return nil
}
