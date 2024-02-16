package viz

import (
	"fmt"
	"log"

	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/workflow"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "viz",
	Short: "Draw graph of workflows",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptPath := args[0]
		//dstBase := args[1]
		log.Printf("doing viz: %s\n", scriptPath)

		wfs, err := scriptfile.RunFile(scriptPath)
		if err != nil {
			return err
		}

		for wfn, wfd := range wfs.Workflows {
			fmt.Printf("digraph %s {\n", wfn)
			wf, err := workflow.PrepWorkflow(wfd, nil)
			if err == nil {
				nameMap := map[string]string{}
				for n := range wf.Steps {
					nameMap[n] = fmt.Sprintf("%d", len(nameMap))
				}

				for n, s := range wf.Steps {
					fmt.Printf("\t%s [label=\"%s\"]\n", nameMap[n], s.GetDesc())
				}

				for n, s := range wf.DepMap {
					for _, d := range s {
						fmt.Printf("\t%s -> %s\n", nameMap[d], nameMap[n])
					}
				}

			}
			fmt.Printf("}\n")
		}

		return nil
	},
}
