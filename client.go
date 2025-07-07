/*
Copyright 2023 Drew Hudson-Viles.

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
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	transport "github.com/aws/smithy-go/endpoints"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// For testing purposes, allowing this function to be replaced in tests
var newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
	return s3.NewFromConfig(cfg, optFns...)
}

// S3 contains an AWS S3 Client.
type S3 struct {
	Client *s3.Client
}

// fileInfo contains information on a file being uploaded to S3.
type fileInfo struct {
	buffer      []byte
	size        int64
	contentType string
}

type staticResolver struct {
	URL *url.URL
}

func (r *staticResolver) ResolveEndpoint(_ context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
	u := *r.URL
	u.Path += "/" + *params.Bucket
	return transport.Endpoint{URI: u}, nil
}

// New generates a new EndpointWithResolverOptions and returns an S3 containing the Bucket and S3Client.
func New(ctx context.Context, endpoint, accessKey, secretKey, region string) (*S3, error) {
	const defaultRegion = "us-east-1"
	r := defaultRegion
	if region != defaultRegion && region != "" {
		r = region
	}
	ep, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(r),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	return &S3{
		Client: newS3ClientFromConfig(cfg, func(o *s3.Options) {
			o.EndpointResolverV2 = &staticResolver{URL: ep}
		}),
	}, nil
}

// CreateBucket uses the client to create an S3 Bucket.
func (s *S3) CreateBucket(ctx context.Context, name string) error {
	_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	return err
}

// ListBuckets uses the client to list all of the S3 Buckets with a specific prefix.
func (s *S3) ListBuckets(ctx context.Context, prefix string) (*s3.ListBucketsOutput, error) {
	buckets, err := s.Client.ListBuckets(ctx, &s3.ListBucketsInput{
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	return buckets, nil
}

// DeleteBucket removes all objects from a bucket and then deletes the bucket itself
func (s *S3) DeleteBucket(ctx context.Context, name string) error {
	// Check the bucket exists first.
	buckets, err := s.ListBuckets(ctx, name)
	if err != nil {
		return err
	}
	if len(buckets.Buckets) == 0 {
		return nil
	}

	// List all objects in the bucket
	objects, err := s.ListObject(ctx, name, "")
	if err != nil {
		return err
	}

	// Delete all objects
	for _, key := range objects {
		if err := s.DeleteObject(ctx, name, key); err != nil {
			return err
		}
	}

	// Delete the bucket
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err = s.Client.DeleteBucket(ctx, input)
	return err
}

// FetchObject Downloads a file from an S3 bucket and returns its contents as a byte array.
func (s *S3) FetchObject(ctx context.Context, fileName, bucket string) ([]byte, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	}

	obj, err := s.Client.GetObject(ctx, params)
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()

	return io.ReadAll(obj.Body)
}

// PutObject Pushes a file to an S3 bucket. If the object is greater than 100MB it will be split into a multipart upload.
func (s *S3) PutObject(ctx context.Context, bucket, key string, body *os.File) error {
	fi, err := readFile(body)
	if err != nil {
		return err
	}

	params := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(fi.contentType),
		Body:        bytes.NewReader(fi.buffer),
	}

	var partMiBs int64 = 100
	maxPartSize := partMiBs * 1024 * 1024
	// If the file is greater than 100MB, then we'll do a multipart upload
	if fi.size > maxPartSize {
		uploader := manager.NewUploader(s.Client, func(u *manager.Uploader) {
			u.PartSize = maxPartSize
		})

		_, err = uploader.Upload(ctx, params)
		if err != nil {
			log.Printf("Couldn't upload large object to %v:%v. Here's why: %v\n",
				bucket, key, err)
		}
	} else {
		_, err = s.Client.PutObject(ctx, params)
		if err != nil {
			return err
		}
	}

	return nil
}

// readFile Reads the file into a buffer and returns the buffer along with the file content type.
func readFile(body *os.File) (*fileInfo, error) {
	info, _ := body.Stat()
	size := info.Size()
	buffer := make([]byte, size)
	fileType := http.DetectContentType(buffer)
	_, err := body.Read(buffer)
	if err != nil {
		return nil, err
	}

	fi := &fileInfo{
		buffer:      buffer,
		contentType: fileType,
		size:        size,
	}

	return fi, nil
}

// ListObject will list the contents of a bucket.
func (s *S3) ListObject(ctx context.Context, bucket, prefix string) ([]string, error) {
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	obj, err := s.Client.ListObjects(ctx, params)
	if err != nil {
		return nil, err
	}

	contents := []string{}
	for _, v := range obj.Contents {
		contents = append(contents, *v.Key)
	}

	return contents, nil
}

// DeleteObject removes a single object from an S3 bucket
func (s *S3) DeleteObject(ctx context.Context, bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.Client.DeleteObject(ctx, input)
	return err
}
