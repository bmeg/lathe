package prep_manifest

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/manifest"
	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/util"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var changeDir = "."
var exclude = []string{}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "prep-manifest",
	Short: "Build manifest of all files that are needed before run",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(exclude) > 0 {
			log.Printf("Excluding %#v", exclude)
		}

		sources := []manifest.ScriptRecord{}

		changeDir, _ = filepath.Abs(changeDir)

		totalFiles := 0
		totalSize := uint64(0)

		for _, dir := range args {
			startDir, _ := filepath.Abs(dir)
			filepath.Walk(startDir,
				func(path string, info fs.FileInfo, err error) error {
					scriptRelpath, _ := filepath.Rel(changeDir, path)
					if strings.HasSuffix(path, ".yaml") {
						sourceFiles := []manifest.FileRecord{}
						//log.Printf("Checking %s\n", path)
						doExclude := false

						for _, e := range exclude {
							ePath, _ := filepath.Abs(e)
							if match, err := filepath.Match(ePath, path); match && err == nil {
								doExclude = true
							}
						}
						if !doExclude {
							pl := scriptfile.ScriptFile{}
							if latheErr := scriptfile.ParseFile(path, &pl); latheErr == nil {
								dir := filepath.Dir(path)

								for _, pr := range pl.Prep {
									for _, v := range pr.Outputs {
										dst := filepath.Join(dir, v)
										if util.Exists(dst) {
											if util.IsDir(dst) {
												m, _ := filepath.Glob(filepath.Join(dst, "*"))
												for _, f := range m {
													//fmt.Printf("Found dir: %s %d\n", f, util.FileSize(f))
													sha1val, _ := util.QuickSHA1(f)
													relPath, _ := filepath.Rel(changeDir, f)
													sourceFiles = append(sourceFiles, manifest.FileRecord{Path: relPath, Size: util.FileSize(f), QuickSHA1: sha1val})
													totalFiles++
													totalSize += util.FileSize(f)
												}
											} else {
												//fmt.Printf("Found: %s: %s %d\n", o, dst, util.FileSize(dst))
												sha1val, _ := util.QuickSHA1(dst)
												relPath, _ := filepath.Rel(changeDir, dst)
												sourceFiles = append(sourceFiles, manifest.FileRecord{Path: relPath, Size: util.FileSize(dst), QuickSHA1: sha1val})
												totalFiles++
												totalSize += util.FileSize(dst)
											}
										} else {
											//fmt.Printf("Missing: %s: %s\n", o, dst)
										}
									}
								}
							}
						}
						if len(sourceFiles) > 0 {
							sources = append(sources, manifest.ScriptRecord{
								Path:  scriptRelpath,
								Files: sourceFiles,
							})
						}
					}

					return nil
				})
		}

		m := manifest.Manifest{Sources: sources, Summary: manifest.SummaryRecord{TotalSize: totalSize, FileCount: totalFiles}}

		out, _ := yaml.Marshal(m)

		fmt.Printf("%s", out)

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&changeDir, "dir", "C", changeDir, "Change Directory for script base")
	flags.StringArrayVarP(&exclude, "exclude", "e", exclude, "Paths to exclude")
}
