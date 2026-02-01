package config

import (
	"context"
	"encoding/json"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinIOClient(cfg *Config) (*minio.Client, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		log.Printf("Created MinIO bucket: %s", cfg.MinIOBucket)
	}

	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{"arn:aws:s3:::" + cfg.MinIOBucket + "/*"},
			},
		},
	}
	policyJSON, _ := json.Marshal(policy)
	err = client.SetBucketPolicy(ctx, cfg.MinIOBucket, string(policyJSON))
	if err != nil {
		log.Printf("Warning: Failed to set bucket policy: %v", err)
	}

	return client, nil
}
