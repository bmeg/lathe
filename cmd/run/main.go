package run

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/workflow"
	"github.com/spf13/cobra"
)

var dryRun bool = false

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run <plan file>",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		//baseDir := filepath.Dir(scriptPath)
		names := []string{}
		if len(args) > 1 {
			names = args[1:]
		}
		workflows, err := scriptfile.RunFile(scriptPath)
		if err != nil {
			log.Printf("Script Error: %s\n", err)
			return err
		}

		if len(names) == 0 {
			wNames := []string{}
			for k := range workflows {
				wNames = append(wNames, k)
			}
			fmt.Printf("workflows: %s\n", strings.Join(wNames, ", "))
		} else {
			for _, n := range names {
				if wfd, ok := workflows[n]; ok {
					wf, err := workflow.PrepWorkflow(wfd)
					if err == nil {
						//fmt.Printf("Running Workflow: %#v\n", wf)
						fwf, err := wf.BuildFlame()
						if err != nil {
							fmt.Printf("workflow build error: %s\n", err)
						}
						fmt.Printf("%#v\n", fwf)

						go func() {
							fwf.ProcessIn <- &workflow.WorkflowStatus{Name: "run", DryRun: dryRun}
							close(fwf.ProcessIn)
						}()

						fwf.Workflow.Start()
						fwf.Workflow.Wait()

						fmt.Printf("Workflow Done\n")
					}
				} else {
					fmt.Printf("Workflow %s not found\n", n)
				}
			}
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&dryRun, "dry-run", "x", dryRun, "Scan workflow without running commands")
}
