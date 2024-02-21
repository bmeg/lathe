package run

import (
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/logger"
	"github.com/bmeg/lathe/runner"
	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/workflow"
	"github.com/spf13/cobra"
)

var verbose = false
var jsonLog = false
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

		logger.Init(verbose, jsonLog)

		//baseDir := filepath.Dir(scriptPath)
		names := []string{}
		if len(args) > 1 {
			names = args[1:]
		}
		workflows, err := scriptfile.RunFile(scriptPath)
		if err != nil {
			logger.Error("Script Error %s : %s\n", scriptPath, err)
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
				logger.Error("Need to choose a workflow", "options", strings.Join(wNames, ", "))
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
						logger.Error("workflow build error: %s\n", err)
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

					logger.Info("Workflow Done")
				}
			} else {
				logger.Error("Workflow not found", "name", n)
			}
		}

		logger.Close()

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&dryRun, "dry-run", "x", dryRun, "Scan workflow without running commands")
	flags.StringVarP(&tesServer, "tes", "t", tesServer, "TES Server")
	flags.BoolVarP(&jsonLog, "jsonlog", "j", jsonLog, "JSON logging output")
	flags.BoolVarP(&verbose, "verbose", "v", verbose, "Vebose logging")
}
