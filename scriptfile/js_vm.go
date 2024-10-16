package scriptfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

func RunFile(path string) (*Plan, error) {

	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(path)
	if abserr != nil {
		return nil, abserr
	}

	// Read file
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config at path %s: \n%v", path, err)
	}

	vm := goja.New()

	pl := &Plan{Workflows: map[string]*WorkflowDesc{}, Path: path, VM: vm, Images: []*DockerImage{}}

	latheObj := map[string]any{
		"Params": map[string]string{
			"mode": "prep",
		},
		"Workflow":    pl.Workflow,
		"LoadPlan":    pl.LoadPlan,
		"Process":     pl.Process,
		"File":        pl.File,
		"Plugin":      pl.Plugin,
		"DockerImage": pl.DockerImage,
	}

	vm.Set("print", pl.Print)
	vm.Set("println", pl.Println)
	vm.Set("glob", pl.Glob)
	vm.Set("lathe", latheObj)

	_, err = vm.RunScript("main", string(source))
	if err != nil {
		return nil, fmt.Errorf("error parsing: %s = %s", path, err)
	}
	//fmt.Printf("%#v\n", pl.Workflows)
	return pl, nil
}
