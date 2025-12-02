package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
	Bucket string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucket string) (*MinioClient, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Set to true if using HTTPS
	})
	if err != nil {
		return nil, err
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		// Set public policy for read access (simplified for dev)
		policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"]}]}`, bucket)
		err = minioClient.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			log.Printf("Failed to set bucket policy: %v", err)
		}
	}

	return &MinioClient{
		Client: minioClient,
		Bucket: bucket,
	}, nil
}

func (m *MinioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	info, err := m.Client.PutObject(ctx, m.Bucket, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	
	baseURL := os.Getenv("PUBLIC_STORAGE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9000"
	}
	
	return fmt.Sprintf("%s/%s/%s", baseURL, m.Bucket, info.Key), nil
}
