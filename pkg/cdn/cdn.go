package cdn

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Cdn struct {
	client *s3.Client
}

func New() (Cdn, error) {
	cdn := Cdn{}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("B3_S3_ACCESS_KEY_ID"),
				os.Getenv("B3_S3_ACCESS_KEY_SECRET"),
				"",
			),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		return Cdn{}, fmt.Errorf("cdn.New: failed to load config")
	}

	cdn.client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("B3_S3_ENDPOINT"))
	})

	return cdn, nil
}

func (cdn Cdn) UploadAsset(path string) (string, error) {
	asset, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("UploadAsset: failed to open asset file at %v: %v", path, err)
	}

	h := sha256.New()
	h.Write(asset)

	assetSha256 := base64.StdEncoding.EncodeToString(h.Sum(nil))
	assetExt := path[strings.LastIndex(path, ".")+1:]
	assetKey := fmt.Sprintf("%v/%v.%v", os.Getenv("B3_S3_FILE_PREFIX"), assetSha256, assetExt)

	_, err = cdn.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("B3_S3_BUCKET_NAME")),
		Key:         aws.String(assetKey),
		Body:        bytes.NewReader(asset),
		ContentType: aws.String(fmt.Sprintf("image/%v", assetExt)),
	})

	if err != nil {
		return "", fmt.Errorf("UploadAsset: failed to upload asset %v: %v", path, err)
	}

	return fmt.Sprintf("%v/%v", os.Getenv("B3_S3_BUCKET_PUBLIC_HOST"), assetKey), nil
}
