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

// S3 contains an S3 Client and a Bucket.
type S3 struct {
	Bucket string
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
func New(endpoint, accessKey, secretKey, bucket, region string) (*S3, error) {
	const defaultRegion = "us-east-1"
	r := defaultRegion
	if region != defaultRegion && region != "" {
		r = region
	}
	ep, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(r),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	return &S3{
		Bucket: bucket,
		Client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.EndpointResolverV2 = &staticResolver{URL: ep}
		}),
	}, nil
}

// CreateBucket uses the client to create an S3 Bucket
func (s *S3) CreateBucket() error {
	ctx := context.Background()
	_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.Bucket),
	})
	return err
}

// DeleteBucket removes all objects from a bucket and then deletes the bucket itself
func (s *S3) DeleteBucket() error {
	// List all objects in the bucket
	objects, err := s.List("")
	if err != nil {
		return err
	}

	// Delete all objects
	for _, key := range objects {
		if err := s.Delete(key); err != nil {
			return err
		}
	}

	// Delete the bucket
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(s.Bucket),
	}

	_, err = s.Client.DeleteBucket(context.Background(), input)
	return err
}

// Fetch Downloads a file from an S3 bucket and returns its contents as a byte array.
func (s *S3) Fetch(fileName string) ([]byte, error) {
	params := &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &fileName,
	}

	obj, err := s.Client.GetObject(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(obj.Body)
}

// Put Pushes a file to an S3 bucket. If the object is greater than 100MB it will be split into a multipart upload.
func (s *S3) Put(key string, body *os.File) error {
	ctx := context.Background()

	fi, err := readFile(body)
	if err != nil {
		return err
	}

	params := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
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
				s.Bucket, key, err)
		}
	} else {
		_, err = s.Client.PutObject(context.Background(), params)
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

// List will list the contents of a bucket.
func (s *S3) List(prefix string) ([]string, error) {

	params := &s3.ListObjectsInput{
		Bucket: &s.Bucket,
		Prefix: &prefix,
	}
	obj, err := s.Client.ListObjects(context.Background(), params)
	if err != nil {
		return nil, err
	}

	contents := []string{}
	for _, v := range obj.Contents {
		contents = append(contents, *v.Key)
	}

	return contents, nil
}

// Delete removes a single object from an S3 bucket
func (s *S3) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.Client.DeleteObject(context.Background(), input)
	return err
}
