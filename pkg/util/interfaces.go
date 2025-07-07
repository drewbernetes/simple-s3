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

package util

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:generate mockgen -source=interfaces.go -destination=../mock/interfaces.go -package=mock

// S3Interface defines the supported S3 operations exposed by this library.
type S3Interface interface {
	// CreateBucket creates a bucket.
	CreateBucket(context.Context, string) error
	// ListBuckets lists buckets filtered by prefix.
	ListBuckets(context.Context, string) (*s3.ListBucketsOutput, error)
	// DeleteBucket deletes a bucket and any objects it contains.
	DeleteBucket(context.Context, string) error
	// FetchObject reads and returns the full object content.
	FetchObject(context.Context, string, string) ([]byte, error)
	// PutObject uploads data to the provided bucket and key.
	PutObject(context.Context, string, string, io.ReadSeeker) error
	// ListObject lists object keys in a bucket filtered by prefix.
	ListObject(context.Context, string, string) ([]string, error)
	// DeleteObject deletes a single object key from a bucket.
	DeleteObject(context.Context, string, string) error
}
