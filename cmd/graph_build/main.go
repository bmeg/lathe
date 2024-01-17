package graph_build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmeg/lathe/util"
	"github.com/spf13/cobra"
)

var outdir = ""

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-build <path-to-playbook(s)>",
	Short: "process outputs of sifter file into graph elements",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		startDir, _ := filepath.Abs(args[0])
		graphOutDir := startDir
		if outdir != "" {
			graphOutDir = outdir
		}
		workingDir, _ := os.Getwd()
		steps := util.ScanObjectToGraph(startDir, workingDir, graphOutDir)

		for _, s := range steps {
			plan, _ := s.GenPlan()
			fmt.Printf("%s\n", plan)
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&outdir, "out", "o", outdir, "Change output Directory")
}
