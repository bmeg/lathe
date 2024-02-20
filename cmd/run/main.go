package run

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/runner"
	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/workflow"
	"github.com/spf13/cobra"
)

var dryRun bool = false
var tesServer = ""

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

		var run runner.CommandRunner
		if tesServer == "" {
			run = runner.NewSingleMachineRunner(16, 32000)
		} else {
			run = runner.NewTesRunner(tesServer, "ubuntu")
		}
		if len(names) == 0 {
			wNames := []string{}
			for k := range workflows.Workflows {
				wNames = append(wNames, k)
			}
			if len(wNames) == 1 {
				names = wNames
			} else {
				log.Printf("Choose a workflow: %s\n", strings.Join(wNames, ", "))
				return nil
			}
		}

		for _, n := range names {
			if wfd, ok := workflows.Workflows[n]; ok {
				wf, err := workflow.PrepWorkflow(wfd, run)
				if err == nil {
					//fmt.Printf("Running Workflow: %#v\n", wf)
					fwf, err := wf.BuildFlame()
					if err != nil {
						log.Printf("workflow build error: %s\n", err)
					}
					//fmt.Printf("%#v\n", fwf)

					go func() {
						fwf.ProcessIn <- &workflow.WorkflowStatus{Name: "run", DryRun: dryRun}
						close(fwf.ProcessIn)
					}()

					/*
						go func() {
							for i := range fwf.ProcessOut {
								fmt.Printf("%#v\n", i)
							}
						}()
					*/
					fwf.Workflow.Start()
					fwf.Workflow.Wait()

					log.Printf("Workflow Done\n")
				}
			} else {
				log.Printf("Workflow %s not found\n", n)
			}
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&dryRun, "dry-run", "x", dryRun, "Scan workflow without running commands")
	flags.StringVarP(&tesServer, "tes", "t", tesServer, "TES Server")
}
