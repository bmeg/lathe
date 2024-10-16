package outputs

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/bmeg/lathe/logger"
	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/util"
	"github.com/spf13/cobra"
)

var outJson = false
var verbose = false
var jsonLog = false
var relPath = "./"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "outputs",
	Short: "Input Operations",
}

var Push = &cobra.Command{
	Use:   "push",
	Short: "Push inputs to storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		//manifestPath := args[0]
		//dstBase := args[1]
		logger.Info("doing push")
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
		logger.Info("doing pull")
		return nil
	},
}

var List = &cobra.Command{
	Use:   "list",
	Short: "List Outputs from playbook",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptPath := args[0]
		wfName := args[1]
		//dstBase := args[1]

		logger.Init(verbose, jsonLog)

		logger.Info("doing list")
		workflows, err := scriptfile.RunFile(scriptPath)
		if err != nil {
			logger.Info("Script Error", "error", err)
			return err
		}

		if relPath != "" {
			relPath, _ = filepath.Abs(relPath)
		}

		if wf, ok := workflows.Workflows[wfName]; ok {
			if outJson {
				for _, p := range wf.Steps {
					//fmt.Printf("step: %s\n", p.GetName())
					proc := p.GetProcess()
					if proc != nil {
						for k, v := range proc.Outputs {
							path := filepath.Join(p.GetBasePath(), v)
							if relPath != "" {
								path, _ = filepath.Rel(relPath, path)
							}
							data := map[string]any{
								"step":   p.GetName(),
								"name":   k,
								"path":   path,
								"exists": util.Exists(path),
							}
							b, err := json.Marshal(data)
							if err == nil {
								fmt.Printf("%s\n", b)
							}
						}
					}
				}
			} else {
				paths := map[string]bool{}
				for _, p := range wf.Steps {
					proc := p.GetProcess()
					if proc != nil {
						for _, v := range proc.Outputs {
							path := filepath.Join(p.GetBasePath(), v)
							paths[path] = true
						}
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
	listFlags.BoolVarP(&verbose, "verbose", "v", verbose, "Vebose logging")

}
