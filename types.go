package simple_s3

import (
    "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client contains an AWS S3Client Client.
type S3Client struct {
    Client         *s3.Client
    customEndpoint bool
    ChunkSize      int64
}

type PutUploadResult struct {
    UploadResults *manager.UploadOutput
    PutResults    *s3.PutObjectOutput
    Error         error
}
