package attachment

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	konfaconfig "github.com/konfa-chat/hub/src/config"
)

// NewStorageFromConfig creates a new storage provider based on the application configuration
func NewStorageFromConfig(cfg *konfaconfig.AttachmentStorage) (Storage, error) {
	switch cfg.Type {
	case "local":
		return NewLocalStorage(cfg.Local.Path)
	case "s3":
		return createS3Storage(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

// createS3Storage initializes an S3 storage provider with the given configuration
func createS3Storage(cfg *konfaconfig.AttachmentStorage) (Storage, error) {
	s3Cfg := cfg.S3

	// Create AWS SDK configuration
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if s3Cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               s3Cfg.Endpoint,
				SigningRegion:     s3Cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		// Fallback to default resolver
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Create AWS configuration
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(s3Cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3Cfg.AccessKeyID,
			s3Cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s3Cfg.UsePathStyle
	})

	return NewS3Storage(s3Client, s3Cfg.Bucket), nil
}
