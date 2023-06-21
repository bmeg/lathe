package util

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func GetS3Client(u *url.URL) (*minio.Client, error) {

	useSSL := false
	if u.Scheme == "s3+https" {
		useSSL = true
	}

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID not set")
	}
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return nil, fmt.Errorf("AWS_SECRET_ACCESS_KEY not set")
	}

	mc, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	return mc, err
}

func GetS3URL(path string) *url.URL {
	if strings.HasPrefix(path, "s3+http://") || strings.HasPrefix(path, "s3+https://") {
		u, err := url.Parse(path)
		if err != nil {
			return nil
		}
		return u
	}
	return nil
}
