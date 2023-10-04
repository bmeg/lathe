package scriptfile

import (
	"fmt"

	"github.com/bmeg/flame"
)

func (pl *ScriptFile) RunScripts() error {

	wf := flame.NewWorkflow()

	for _, s := range pl.Scripts {

		fmt.Printf("Input: %s\n", s.Inputs)
	}

	fmt.Printf("%#v\n", wf)

	return nil
}
