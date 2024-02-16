package inputs

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/lathe/scriptfile"
	"github.com/spf13/cobra"
)

var outJson = false

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "inputs",
	Short: "Input Operations",
}

var Push = &cobra.Command{
	Use:   "push",
	Short: "Push inputs to storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		//manifestPath := args[0]
		//dstBase := args[1]
		log.Printf("doing push\n")
		return nil
	},
}

var Pull = &cobra.Command{
	Use:   "pull",
	Short: "Pull inputs from storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		//manifestPath := args[0]
		//dstBase := args[1]
		log.Printf("doing pull\n")
		return nil
	},
}

var List = &cobra.Command{
	Use:   "list",
	Short: "List Inputs from playbook",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptPath := args[0]
		wfName := args[1]
		//dstBase := args[1]
		log.Printf("doing list\n")
		workflows, err := scriptfile.RunFile(scriptPath)
		if err != nil {
			log.Printf("Script Error: %s\n", err)
			return err
		}

		if wf, ok := workflows.Workflows[wfName]; ok {
			if outJson {
				for _, p := range wf.Steps {
					for k, v := range p.GetInputs() {
						path := filepath.Join(p.GetBasePath(), v)
						data := map[string]any{
							"step": p.GetName(),
							"name": k,
							"path": path,
						}
						b, err := json.Marshal(data)
						if err == nil {
							fmt.Printf("%s\n", b)
						}
					}
				}
			} else {
				paths := map[string]bool{}
				for _, p := range wf.Steps {
					for _, v := range p.GetInputs() {
						path := filepath.Join(p.GetBasePath(), v)
						paths[path] = true
					}
				}
				for k := range paths {
					fmt.Printf("%s\n", k)
				}
			}
		}
		return nil
	},
}

func init() {
	Cmd.AddCommand(Push)
	Cmd.AddCommand(Pull)
	Cmd.AddCommand(List)

	listFlags := List.Flags()
	listFlags.BoolVarP(&outJson, "json", "j", outJson, "Output JSON")
}
