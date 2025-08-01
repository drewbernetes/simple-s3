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
    "os"

    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Interface interface {
    CreateBucket(context.Context, string) (*s3.CreateBucketOutput, error)
    BucketExists(context.Context, string) (bool, error)
    ListBuckets(context.Context) (*s3.ListBucketsOutput, error)
    DeleteBucket(context.Context, string) (*s3.DeleteBucketOutput, error)
    GetObject(context.Context, string, string) ([]byte, error)
    PutObject(context.Context, string, string, *os.File) *PutUploadResult
    ListObject(context.Context, string, string) ([]string, error)
    DeleteObject(context.Context, string, string) (*s3.DeleteObjectOutput, error)
}
