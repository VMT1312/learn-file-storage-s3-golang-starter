package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	objectInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	presignClient := s3.NewPresignClient(s3Client)
	presignRequest, err := presignClient.PresignGetObject(context.Background(), objectInput, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return presignRequest.URL, nil
}
