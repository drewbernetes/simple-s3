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
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type S3 struct {
	Bucket string
	Client *s3.Client
}

func New(endpoint, accessKey, secretKey, bucket, region string) (*S3, error) {
	const defaultRegion = "us-east-1"
	r := defaultRegion
	if region != "" {
		r = region
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               endpoint,
			SigningRegion:     r,
			HostnameImmutable: true,
		}, nil
	})

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(r),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	return &S3{
		Bucket: bucket,
		Client: s3.NewFromConfig(cfg),
	}, nil
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

// Put Pushes a file to an S3 bucket.
func (s *S3) Put(contentType, key string, body io.ReadSeeker) error {
	params := &s3.PutObjectInput{
		Bucket:      &s.Bucket,
		Key:         &key,
		ContentType: &contentType,
		Body:        body,
	}

	_, err := s.Client.PutObject(context.Background(), params)
	if err != nil {
		return err
	}

	return nil
}

// List will list the contents of a bucket
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
