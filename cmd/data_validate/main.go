package data_validate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bmeg/golib"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func yamlLoader(s string) (io.ReadCloser, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	f := u.Path
	if runtime.GOOS == "windows" {
		f = strings.TrimPrefix(f, "/")
		f = filepath.FromSlash(f)
	}
	fmt.Printf("Loading: %s\n", f)
	if strings.HasSuffix(f, ".yaml") {
		source, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		d := map[string]any{}
		yaml.Unmarshal(source, &d)
		schemaText, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(string(schemaText))), nil
	}
	return os.Open(f)
}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "data-validate",
	Short: "Data Validate",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		schemaFile := args[0]
		inputPath := args[1]

		jsonschema.Loaders["file"] = yamlLoader

		compiler := jsonschema.NewCompiler()
		compiler.ExtractAnnotations = true

		sch, err := compiler.Compile(schemaFile)
		if err != nil {
			fmt.Printf("Error compiling %s : %s\n", schemaFile, err)
		} else {
			if len(sch.Types) == 1 && sch.Types[0] == "object" {
				fmt.Printf("OK: %s %s (%s)\n", schemaFile, sch.Id, sch.Title)
			}
		}

		var reader chan []byte
		if strings.HasSuffix(inputPath, ".gz") {
			reader, err = golib.ReadGzipLines(inputPath)
		} else {
			reader, err = golib.ReadFileLines(inputPath)
		}
		if err != nil {
			return err
		}

		procChan := make(chan map[string]interface{}, 100)
		go func() {
			for line := range reader {
				o := map[string]interface{}{}
				if len(line) > 0 {
					json.Unmarshal(line, &o)
					procChan <- o
				}
			}
			close(procChan)
		}()

		validCount := 0
		errorCount := 0
		for row := range procChan {
			err = sch.Validate(row)
			if err != nil {
				errorCount++
				fmt.Printf("Error: %s\n", err)
			} else {
				validCount++
			}
		}
		fmt.Printf("%s results: %d valid records %d invalid records\n", inputPath, validCount, errorCount)
		return nil
	},
}
