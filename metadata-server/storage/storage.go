package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bfbarry/coop-storage/metadata-server/config"
)

// Client wraps the S3-compatible RustFS connection and presigner.
type Client struct {
	s3        *s3.Client
	presigner *s3.PresignClient
	cfg       config.RustFSConfig
}

func NewClient(cfg config.RustFSConfig) *Client {
	s3Client := s3.NewFromConfig(
		aws.Config{
			// Region: cfg.Region,
			Credentials: credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"", // session token — not used
			),
			EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               cfg.Endpoint,
						HostnameImmutable: true, // keeps the endpoint URL as-is
					}, nil
				},
			),
		},
		func(o *s3.Options) {
			o.UsePathStyle = cfg.UsePathStyle // required for RustFS / MinIO
		},
	)

	return &Client{
		s3:        s3Client,
		presigner: s3.NewPresignClient(s3Client),
		cfg:       cfg,
	}
}

// PresignUpload returns a pre-signed PUT URL the client can use to upload
// directly to RustFS without routing bytes through this service.
//
// objectKey should be a unique, caller-controlled path (e.g. "{userID}/{fileID}").
// contentType is forwarded as a signed header so the client must send it unchanged.
func (c *Client) PresignUpload(
	ctx context.Context,
	objectKey string,
	contentType string,
	contentLength int64,
) (url string, expiresAt time.Time, err error) {
	req, err := c.presigner.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket:        aws.String(c.cfg.Bucket),
			Key:           aws.String(objectKey),
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(contentLength),
		},
		s3.WithPresignExpires(c.cfg.PresignDuration),
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presign put object: %w", err)
	}

	expiresAt = time.Now().UTC().Add(c.cfg.PresignDuration)
	return req.URL, expiresAt, nil
}

//TODO: PresignDownload
