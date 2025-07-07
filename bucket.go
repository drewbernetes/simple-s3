package simple_s3

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

// BucketExists checks if a bucket exists
func (s *S3Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
    // List buckets
    result, err := s.ListBuckets(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to list buckets: %w", err)
    }

    // Might as well check if there are any buckets at all, if there isn't it's pretty darn safe to presume the bucket doesn't exist!
    if len(result.Buckets) > 0 {
        // Check if our bucket exists in the list
        for _, bucket := range result.Buckets {
            if *bucket.Name == bucketName {
                return true, nil
            }
        }
    }

    return false, nil
}

// CreateBucket creates a new S3 bucket
func (s *S3Client) CreateBucket(ctx context.Context, bucketName string) (*s3.CreateBucketOutput, error) {
    // Check if the bucket already exists
    exists, err := s.BucketExists(ctx, bucketName)
    if err != nil {
        return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
    }

    if exists {
        return nil, fmt.Errorf("bucket %s already exists", bucketName)
    }

    fmt.Println("Bucket doesn't exist, creating it", bucketName)

    // Create bucket
    return s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
        Bucket: aws.String(bucketName),
    })
}

// ListBuckets lists all S3 buckets
func (s *S3Client) ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error) {
    // Call S3 to list buckets
    return s.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
}

// DeleteBucket removes all objects from a bucket and then deletes the bucket itself
func (s *S3Client) DeleteBucket(ctx context.Context, bucketName string) (*s3.DeleteBucketOutput, error) {
    // Check if the bucket already exists
    exists, err := s.BucketExists(ctx, bucketName)
    if err != nil {
        return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
    }

    if !exists {
        return nil, fmt.Errorf("bucket %s does not exist", bucketName)
    }

    // List all objects in the bucket
    objects, err := s.ListObject(ctx, bucketName, "")
    if err != nil {
        return nil, err
    }

    // Delete all objects
    for _, key := range objects {
        if _, err := s.DeleteObject(ctx, bucketName, key); err != nil {
            return nil, err
        }
    }

    // Delete the bucket
    input := &s3.DeleteBucketInput{
        Bucket: aws.String(bucketName),
    }

    return s.Client.DeleteBucket(ctx, input)
}
