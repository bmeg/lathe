package schema_lint

import (
	"fmt"

	"github.com/bmeg/grip/log"
	"github.com/bmeg/sifter/schema"
	"github.com/spf13/cobra"
)

var httpDir string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-lint",
	Short: "Schema lint",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		loadOpts := []schema.LoadOpt{
			{
				LogError: func(uri string, err error) {
					log.Errorf("Error compiling %s : %s\n", uri, err)
				},
			},
		}
		/*
			if httpDir != "" {
				httpMap := map[string]string{}

				files, _ := filepath.Glob(filepath.Join(httpDir, "*.yaml"))
				if t, err := filepath.Glob(filepath.Join(httpDir, "*.json")); err == nil {
					files = append(files, t...)
				}
				for _, f := range files {
					d := map[string]any{}
					source, err := os.ReadFile(f)
					if err == nil {
						yaml.Unmarshal(source, &d)
						if id, ok := d["$id"]; ok {
							if idStr, ok := id.(string); ok {
								idStr = strings.TrimSuffix(idStr, "#")
								httpMap[idStr] = f
							}
						}
					}
				}
				fmt.Printf("IDMap: %#v\n", httpMap)
				loadOpts = append(loadOpts, schema.LoadOpt{
					HttpFileMap: func(url string) string {
						for k, v := range httpMap {
							if strings.HasPrefix(url, k) {
								return strings.Replace(url, k, v, 1)
							}
						}
						return url
					},
				})
			}
		*/
		sch, err := schema.Load(args[0], loadOpts...)
		if err == nil {
			for _, cls := range sch.Classes {
				fmt.Printf("OK: %s (%s)\n", cls.Title, cls.Location)
			}
		} else {
			log.Errorf("Loading error: %s", err)
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&httpDir, "httpdir", "d", httpDir, "Map directory to http")
}
