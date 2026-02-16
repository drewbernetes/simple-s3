/*
Copyright 2026 Drew Hudson-Viles.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package simple_s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	transport "github.com/aws/smithy-go/endpoints"
)

// newS3ClientFromConfig is a test hook for constructing an S3 client.
var newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
	return s3.NewFromConfig(cfg, optFns...)
}

// Test hooks for AWS SDK calls to keep behavior unit-testable.
var loadDefaultConfig = config.LoadDefaultConfig

var s3CreateBucket = func(c *s3.Client, ctx context.Context, params *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	return c.CreateBucket(ctx, params)
}

var s3ListBuckets = func(c *s3.Client, ctx context.Context, params *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return c.ListBuckets(ctx, params)
}

var s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
	return c.HeadBucket(ctx, params)
}

var s3DeleteBucket = func(c *s3.Client, ctx context.Context, params *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	return c.DeleteBucket(ctx, params)
}

var s3GetObject = func(c *s3.Client, ctx context.Context, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return c.GetObject(ctx, params)
}

var s3DeleteObject = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return c.DeleteObject(ctx, params)
}

var s3DeleteObjects = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	return c.DeleteObjects(ctx, params)
}

var newTransferManager = func(c *s3.Client, optFns ...func(*transfermanager.Options)) transferManagerAPI {
	return transfermanager.New(c, optFns...)
}

var listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	paginator := s3.NewListObjectsV2Paginator(c, params)

	contents := make([]s3types.Object, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		contents = append(contents, page.Contents...)
	}

	return contents, nil
}

// S3 wraps an AWS S3 client with simplified helper methods.
type S3 struct {
	// Client is the underlying AWS SDK S3 client used to execute requests.
	Client *s3.Client
}

// transferManagerAPI captures the transfermanager client behavior used by PutObject.
type transferManagerAPI interface {
	UploadObject(ctx context.Context, params *transfermanager.UploadObjectInput, optFns ...func(*transfermanager.Options)) (*transfermanager.UploadObjectOutput, error)
}

// staticResolver resolves all S3 requests to a fixed endpoint URL.
type staticResolver struct {
	URL *url.URL
}

func (r *staticResolver) ResolveEndpoint(_ context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
	u := *r.URL
	if params.Bucket != nil && *params.Bucket != "" {
		u.Path = strings.TrimSuffix(u.Path, "/") + "/" + *params.Bucket
	}
	return transport.Endpoint{URI: u}, nil
}

// New creates a configured S3 wrapper.
//
// If an endpoint is provided, requests are routed to that endpoint using path-style addressing.
// If a region is empty, us-east-1 is used.
func New(ctx context.Context, endpoint, accessKey, secretKey, region string) (*S3, error) {
	const defaultRegion = "us-east-1"
	r := defaultRegion
	if region != defaultRegion && region != "" {
		r = region
	}

	cfg, err := loadDefaultConfig(ctx,
		config.WithRegion(r),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	options := make([]func(*s3.Options), 0, 2)
	if endpoint != "" {
		ep, parseErr := url.Parse(endpoint)
		if parseErr != nil {
			return nil, parseErr
		}
		options = append(options, func(o *s3.Options) {
			o.EndpointResolverV2 = &staticResolver{URL: ep}
			o.UsePathStyle = true
		})
	}

	return &S3{Client: newS3ClientFromConfig(cfg, options...)}, nil
}

// CreateBucket creates a bucket with the provided name.
func (s *S3) CreateBucket(ctx context.Context, name string) error {
	_, err := s3CreateBucket(s.Client, ctx, &s3.CreateBucketInput{Bucket: aws.String(name)})
	return err
}

// ListBuckets lists buckets filtered by the provided prefix.
func (s *S3) ListBuckets(ctx context.Context, prefix string) (*s3.ListBucketsOutput, error) {
	buckets, err := s3ListBuckets(s.Client, ctx, &s3.ListBucketsInput{Prefix: aws.String(prefix)})
	if err != nil {
		return nil, err
	}
	return buckets, nil
}

// DeleteBucket removes all objects from a bucket and then deletes the bucket.
//
// If the bucket does not exist, DeleteBucket returns nil.
func (s *S3) DeleteBucket(ctx context.Context, name string) error {
	_, err := s3HeadBucket(s.Client, ctx, &s3.HeadBucketInput{Bucket: aws.String(name)})
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}

	objects, err := listObjectsV2All(ctx, s.Client, name, "")
	if err != nil {
		return err
	}

	identifiers := make([]s3types.ObjectIdentifier, 0, len(objects))
	for _, object := range objects {
		if object.Key == nil || *object.Key == "" {
			continue
		}
		identifiers = append(identifiers, s3types.ObjectIdentifier{Key: object.Key})
	}

	for i := 0; i < len(identifiers); i += 1000 {
		end := min(i+1000, len(identifiers))
		_, err = s3DeleteObjects(s.Client, ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(name),
			Delete: &s3types.Delete{
				Objects: identifiers[i:end],
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}
	}

	_, err = s3DeleteBucket(s.Client, ctx, &s3.DeleteBucketInput{Bucket: aws.String(name)})
	return err
}

// FetchObject downloads an object and returns its full contents.
func (s *S3) FetchObject(ctx context.Context, fileName, bucket string) ([]byte, error) {
	obj, err := s3GetObject(s.Client, ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, err
	}

	defer obj.Body.Close()

	return io.ReadAll(obj.Body)
}

// PutObject uploads content to a bucket key.
//
// Objects larger than 100 MiB are uploaded using multipart transfer settings.
func (s *S3) PutObject(ctx context.Context, bucket, key string, body io.ReadSeeker) error {
	contentType, err := readContentType(body)
	if err != nil {
		return err
	}

	params := &transfermanager.UploadObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        body,
	}

	var partMiBs int64 = 100
	maxPartSize := partMiBs * 1024 * 1024
	client := newTransferManager(s.Client, func(o *transfermanager.Options) {
		o.PartSizeBytes = maxPartSize
		o.MultipartUploadThreshold = maxPartSize
	})

	_, err = client.UploadObject(ctx, params)
	return err
}

// readContentType reads up to 512 bytes to detect content type and rewinds the reader.
func readContentType(body io.ReadSeeker) (string, error) {
	header := make([]byte, 512)
	n, err := body.Read(header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	if _, err := body.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	if n == 0 {
		return http.DetectContentType(nil), nil
	}
	return http.DetectContentType(header[:n]), nil
}

// ListObject lists object keys in a bucket filtered by prefix.
func (s *S3) ListObject(ctx context.Context, bucket, prefix string) ([]string, error) {
	objects, err := listObjectsV2All(ctx, s.Client, bucket, prefix)
	if err != nil {
		return nil, err
	}

	contents := make([]string, 0, len(objects))
	for _, object := range objects {
		if object.Key == nil {
			continue
		}
		contents = append(contents, *object.Key)
	}

	return contents, nil
}

// DeleteObject removes a single object from a bucket.
func (s *S3) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := s3DeleteObject(s.Client, ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func isNotFoundError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "NotFound" || code == "NoSuchBucket"
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
