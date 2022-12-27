package class_concat

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bmeg/golib"
	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

var outFile = ""

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "class-concat",
	Short: "Concatinate output files of class type",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir := args[0]
		className := args[1]
		//userInputs := map[string]string{}

		var outStream io.Writer = os.Stdout

		if outFile != "" {
			var err error
			outFile, err := os.Create(outFile)
			if err != nil {
				return err
			}
			defer outFile.Close()
			outStream = gzip.NewWriter(outFile)
		}

		util.ScanSifter(baseDir, func(pb *playbook.Playbook) {
			//localInputs, err := pb.PrepConfig(userInputs, baseDir)
			//task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), localInputs)

			for pname, p := range pb.Pipelines {
				emitName := ""
				for _, s := range p {
					if s.Emit != nil {
						emitName = s.Emit.Name
					}
				}
				if emitName != "" {
					for _, s := range p {
						if s.ObjectValidate != nil {
							if s.ObjectValidate.Class == className {
								outdir := pb.GetDefaultOutDir()
								outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
								outpath := filepath.Join(outdir, outname)

								data, err := golib.ReadGzipLines(outpath)
								if err == nil {
									for line := range data {
										outStream.Write(line)
										outStream.Write([]byte("\n"))
									}
								}
							}
						}
					}
				}
			}
		})
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&outFile, "out", "o", outFile, "Output file")
}
