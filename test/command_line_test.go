package test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"

	"sigs.k8s.io/yaml"
)

var tPath string = "config.yaml"

type CommandLineConfig struct {
	Playbook []string `json:"playbook"`
}

func TestCommandLines(t *testing.T) {
	raw, err := ioutil.ReadFile(tPath)
	if err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}
	conf := []CommandLineConfig{}
	if err := yaml.UnmarshalStrict(raw, &conf); err != nil {
		t.Error(fmt.Errorf("failed to read config %s %s", tPath, err))
	}
	// read in conf, ie config.yaml in this case
	for _, c := range conf {
		cmd := exec.Command("lathe", c.Playbook...)
		fmt.Printf("Running: %s\n", c.Playbook)
		var output []byte
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Failed running %s: %s", c.Playbook, err)
		}
		fmt.Printf("Command output:\n%s\n", output)
	}
}
