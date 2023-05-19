package prep_upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/manifest"
	"github.com/bmeg/lathe/util"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "prep-upload",
	Short: "Upload files in manfiest",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestPath := args[0]
		dstBase := args[1]

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			return err
		}
		mfile := manifest.Manifest{}
		yaml.Unmarshal(data, &mfile)

		dstURL := util.GetS3URL(dstBase)
		mc, err := util.GetS3Client(dstURL)

		if err != nil {
			fmt.Printf("err: %s\n", err)
			return nil
		}

		tmp := strings.SplitN(dstURL.Path, "/", 3)
		bucketName := tmp[1]
		path := ""
		if len(tmp) > 2 {
			path = tmp[2]
		}
		fmt.Printf("(%s)(%s)\n", bucketName, path)

		for _, source := range mfile.Sources {
			for _, file := range source.Files {
				key := filepath.Join(path, file.Path)
				stats, err := mc.StatObject(cmd.Context(), bucketName, key, minio.GetObjectOptions{})
				if err != nil {
					errResponse := minio.ToErrorResponse(err)
					if errResponse.Code == "NoSuchKey" {
						fmt.Printf("Find not found: %s\n", key)
					}
				} else {
					fmt.Printf("%s %d\n", key, stats.Size)
				}

			}
		}

		/*
			files := mc.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Prefix: path, Recursive: true})
			for f := range files {
				obj, _ := mc.GetObject(context.Background(), bucketName, f.Key, minio.GetObjectOptions{})
				stats, _ := obj.Stat()
				fmt.Printf("%s %d\n", f.Key, stats.Size)
			}
		*/

		//cred := credentials.NewFileMinioClient("", "rgw-lab")
		//values, err := cred.Get()

		//values.
		//minioClient, err := minio.New(endpoint, &minio.Options{
		//	Creds:  cred,
		//		Secure: true,
		//	})

		return nil
	},
}

func init() {
	//flags := Cmd.Flags()
}
