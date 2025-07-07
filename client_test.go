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
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/drewbernetes/simple-s3/pkg/util"
)

var _ util.S3Interface = (*S3)(nil)

var (
	origNewS3ClientFromConfig = newS3ClientFromConfig
	origLoadDefaultConfig     = loadDefaultConfig
	origS3CreateBucket        = s3CreateBucket
	origS3ListBuckets         = s3ListBuckets
	origS3HeadBucket          = s3HeadBucket
	origS3DeleteBucket        = s3DeleteBucket
	origS3GetObject           = s3GetObject
	origS3DeleteObject        = s3DeleteObject
	origS3DeleteObjects       = s3DeleteObjects
	origNewTransferManager    = newTransferManager
	origListObjectsV2All      = listObjectsV2All
)

func restoreHooks() {
	newS3ClientFromConfig = origNewS3ClientFromConfig
	loadDefaultConfig = origLoadDefaultConfig
	s3CreateBucket = origS3CreateBucket
	s3ListBuckets = origS3ListBuckets
	s3HeadBucket = origS3HeadBucket
	s3DeleteBucket = origS3DeleteBucket
	s3GetObject = origS3GetObject
	s3DeleteObject = origS3DeleteObject
	s3DeleteObjects = origS3DeleteObjects
	newTransferManager = origNewTransferManager
	listObjectsV2All = origListObjectsV2All
}

var _ = Describe("S3 Client", func() {
	BeforeEach(func() {
		restoreHooks()
	})

	AfterEach(func() {
		restoreHooks()
	})

	Describe("New", func() {
		BeforeEach(func() {
			loadDefaultConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
				opts := config.LoadOptions{}
				for _, fn := range optFns {
					err := fn(&opts)
					Expect(err).NotTo(HaveOccurred())
				}
				return aws.Config{Region: opts.Region}, nil
			}
		})

		It("uses default region and no endpoint options when endpoint is empty", func() {
			cfgRegion := ""
			optCount := 0
			newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
				cfgRegion = cfg.Region
				optCount = len(optFns)
				return &s3.Client{}
			}

			_, err := New(context.Background(), "", "ak", "sk", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(cfgRegion).To(Equal("us-east-1"))
			Expect(optCount).To(Equal(0))
		})

		It("uses supplied region and endpoint options", func() {
			seenPathStyle := false
			newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
				Expect(cfg.Region).To(Equal("eu-west-1"))
				Expect(optFns).To(HaveLen(1))

				o := &s3.Options{}
				optFns[0](o)
				seenPathStyle = o.UsePathStyle
				Expect(o.EndpointResolverV2).NotTo(BeNil())
				return &s3.Client{}
			}

			_, err := New(context.Background(), "http://example.local:9000", "ak", "sk", "eu-west-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(seenPathStyle).To(BeTrue())
		})

		It("returns an error for invalid endpoint", func() {
			_, err := New(context.Background(), "://bad-url", "ak", "sk", "eu-west-1")
			Expect(err).To(HaveOccurred())
		})

		It("returns load config error", func() {
			loadDefaultConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
				return aws.Config{}, errors.New("load config failed")
			}

			_, err := New(context.Background(), "", "ak", "sk", "")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("staticResolver", func() {
		It("adds bucket path when present", func() {
			base, err := url.Parse("http://localhost:9000/base")
			Expect(err).NotTo(HaveOccurred())

			r := &staticResolver{URL: base}
			bucket := "my-bucket"
			ep, err := r.ResolveEndpoint(context.Background(), s3.EndpointParameters{Bucket: &bucket})
			Expect(err).NotTo(HaveOccurred())
			Expect(ep.URI.String()).To(Equal("http://localhost:9000/base/my-bucket"))
		})

		It("handles nil bucket safely", func() {
			base, err := url.Parse("http://localhost:9000/base")
			Expect(err).NotTo(HaveOccurred())

			r := &staticResolver{URL: base}
			ep, err := r.ResolveEndpoint(context.Background(), s3.EndpointParameters{})
			Expect(err).NotTo(HaveOccurred())
			Expect(ep.URI.String()).To(Equal("http://localhost:9000/base"))
		})
	})

	Describe("CreateBucket", func() {
		It("creates a bucket", func() {
			sut := &S3{Client: &s3.Client{}}
			s3CreateBucket = func(c *s3.Client, ctx context.Context, params *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
				Expect(aws.ToString(params.Bucket)).To(Equal("bucket-a"))
				return &s3.CreateBucketOutput{}, nil
			}

			err := sut.CreateBucket(context.Background(), "bucket-a")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an underlying error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3CreateBucket = func(c *s3.Client, ctx context.Context, params *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
				return nil, errors.New("boom")
			}

			err := sut.CreateBucket(context.Background(), "bucket-a")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListBuckets", func() {
		It("lists buckets", func() {
			sut := &S3{Client: &s3.Client{}}
			s3ListBuckets = func(c *s3.Client, ctx context.Context, params *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
				Expect(aws.ToString(params.Prefix)).To(Equal("prefix-"))
				return &s3.ListBucketsOutput{Buckets: []s3types.Bucket{{Name: aws.String("prefix-a")}}}, nil
			}

			out, err := sut.ListBuckets(context.Background(), "prefix-")
			Expect(err).NotTo(HaveOccurred())
			Expect(out.Buckets).To(HaveLen(1))
			Expect(aws.ToString(out.Buckets[0].Name)).To(Equal("prefix-a"))
		})

		It("returns an underlying error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3ListBuckets = func(c *s3.Client, ctx context.Context, params *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
				return nil, errors.New("boom")
			}

			_, err := sut.ListBuckets(context.Background(), "prefix-")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteBucket", func() {
		It("returns nil when bucket does not exist", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return nil, apiErr{code: "NoSuchBucket"}
			}

			calledDelete := false
			s3DeleteBucket = func(c *s3.Client, ctx context.Context, params *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
				calledDelete = true
				return &s3.DeleteBucketOutput{}, nil
			}

			err := sut.DeleteBucket(context.Background(), "missing")
			Expect(err).NotTo(HaveOccurred())
			Expect(calledDelete).To(BeFalse())
		})

		It("returns head bucket error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return nil, errors.New("head failed")
			}

			err := sut.DeleteBucket(context.Background(), "bucket-a")
			Expect(err).To(HaveOccurred())
		})

		It("returns list objects error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return &s3.HeadBucketOutput{}, nil
			}
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				return nil, errors.New("list failed")
			}

			err := sut.DeleteBucket(context.Background(), "bucket-a")
			Expect(err).To(HaveOccurred())
		})

		It("deletes objects in 1000-item chunks and then deletes bucket", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return &s3.HeadBucketOutput{}, nil
			}

			objects := make([]s3types.Object, 1001)
			for i := range objects {
				key := "key"
				objects[i] = s3types.Object{Key: &key}
			}
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				return objects, nil
			}

			chunkSizes := make([]int, 0)
			s3DeleteObjects = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
				chunkSizes = append(chunkSizes, len(params.Delete.Objects))
				return &s3.DeleteObjectsOutput{}, nil
			}

			deletedBucket := false
			s3DeleteBucket = func(c *s3.Client, ctx context.Context, params *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
				deletedBucket = true
				return &s3.DeleteBucketOutput{}, nil
			}

			err := sut.DeleteBucket(context.Background(), "bucket-a")
			Expect(err).NotTo(HaveOccurred())
			Expect(chunkSizes).To(Equal([]int{1000, 1}))
			Expect(deletedBucket).To(BeTrue())
		})

		It("returns delete objects error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return &s3.HeadBucketOutput{}, nil
			}
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				key := "key"
				return []s3types.Object{{Key: &key}}, nil
			}
			s3DeleteObjects = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
				return nil, errors.New("delete objects failed")
			}

			err := sut.DeleteBucket(context.Background(), "bucket-a")
			Expect(err).To(HaveOccurred())
		})

		It("skips objects with nil or empty keys", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return &s3.HeadBucketOutput{}, nil
			}

			empty := ""
			valid := "valid-key"
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				return []s3types.Object{
					{},
					{Key: &empty},
					{Key: &valid},
				}, nil
			}
			s3DeleteObjects = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
				Expect(params.Delete.Objects).To(HaveLen(1))
				Expect(aws.ToString(params.Delete.Objects[0].Key)).To(Equal(valid))
				return &s3.DeleteObjectsOutput{}, nil
			}
			s3DeleteBucket = func(c *s3.Client, ctx context.Context, params *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
				return &s3.DeleteBucketOutput{}, nil
			}

			err := sut.DeleteBucket(context.Background(), "bucket-a")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns delete bucket error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3HeadBucket = func(c *s3.Client, ctx context.Context, params *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
				return &s3.HeadBucketOutput{}, nil
			}
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				return nil, nil
			}
			s3DeleteBucket = func(c *s3.Client, ctx context.Context, params *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
				return nil, errors.New("delete bucket failed")
			}

			err := sut.DeleteBucket(context.Background(), "bucket-a")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("FetchObject", func() {
		It("fetches an object", func() {
			sut := &S3{Client: &s3.Client{}}
			s3GetObject = func(c *s3.Client, ctx context.Context, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("payload")))}, nil
			}

			data, err := sut.FetchObject(context.Background(), "key-a", "bucket-a")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("payload"))
		})

		It("returns get object error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3GetObject = func(c *s3.Client, ctx context.Context, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				return nil, errors.New("get object failed")
			}

			_, err := sut.FetchObject(context.Background(), "key-a", "bucket-a")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("PutObject", func() {
		It("uploads with transfer manager and rewinds body after content sniff", func() {
			sut := &S3{Client: &s3.Client{}}
			fakeClient := &fakeTransferManager{}
			newTransferManager = func(c *s3.Client, optFns ...func(*transfermanager.Options)) transferManagerAPI {
				o := &transfermanager.Options{}
				for _, fn := range optFns {
					fn(o)
				}
				Expect(o.PartSizeBytes).To(Equal(int64(100 * 1024 * 1024)))
				Expect(o.MultipartUploadThreshold).To(Equal(int64(100 * 1024 * 1024)))
				return fakeClient
			}

			body := bytes.NewReader([]byte("hello world"))
			err := sut.PutObject(context.Background(), "bucket-a", "key-a", body)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeClient.uploadInput).NotTo(BeNil())
			Expect(aws.ToString(fakeClient.uploadInput.ContentType)).To(Equal("text/plain; charset=utf-8"))

			uploaded, readErr := io.ReadAll(fakeClient.uploadInput.Body)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(string(uploaded)).To(Equal("hello world"))
		})

		It("returns read content type error", func() {
			sut := &S3{Client: &s3.Client{}}
			err := sut.PutObject(context.Background(), "bucket-a", "key-a", &failingReadSeeker{})
			Expect(err).To(HaveOccurred())
		})

		It("returns transfer upload error", func() {
			sut := &S3{Client: &s3.Client{}}
			newTransferManager = func(c *s3.Client, optFns ...func(*transfermanager.Options)) transferManagerAPI {
				return &fakeTransferManager{err: errors.New("upload failed")}
			}

			err := sut.PutObject(context.Background(), "bucket-a", "key-a", bytes.NewReader([]byte("hello world")))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("readContentType", func() {
		It("returns a default type for empty reader", func() {
			ct, err := readContentType(bytes.NewReader(nil))
			Expect(err).NotTo(HaveOccurred())
			Expect(ct).NotTo(BeEmpty())
		})

		It("returns seek error", func() {
			_, err := readContentType(&seekErrorReadSeeker{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListObject", func() {
		It("lists object keys", func() {
			sut := &S3{Client: &s3.Client{}}
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				k1 := "a"
				k2 := "b"
				return []s3types.Object{{Key: &k1}, {}, {Key: &k2}}, nil
			}

			keys, err := sut.ListObject(context.Background(), "bucket-a", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(keys).To(Equal([]string{"a", "b"}))
		})

		It("returns list error", func() {
			sut := &S3{Client: &s3.Client{}}
			listObjectsV2All = func(ctx context.Context, c *s3.Client, bucket, prefix string) ([]s3types.Object, error) {
				return nil, errors.New("list failed")
			}

			_, err := sut.ListObject(context.Background(), "bucket-a", "")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteObject", func() {
		It("deletes a single object", func() {
			sut := &S3{Client: &s3.Client{}}
			s3DeleteObject = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
				Expect(aws.ToString(params.Bucket)).To(Equal("bucket-a"))
				Expect(aws.ToString(params.Key)).To(Equal("key-a"))
				return &s3.DeleteObjectOutput{}, nil
			}

			err := sut.DeleteObject(context.Background(), "bucket-a", "key-a")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns delete error", func() {
			sut := &S3{Client: &s3.Client{}}
			s3DeleteObject = func(c *s3.Client, ctx context.Context, params *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
				return nil, errors.New("delete failed")
			}

			err := sut.DeleteObject(context.Background(), "bucket-a", "key-a")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("isNotFoundError", func() {
		It("returns true for not found codes", func() {
			Expect(isNotFoundError(apiErr{code: "NotFound"})).To(BeTrue())
			Expect(isNotFoundError(apiErr{code: "NoSuchBucket"})).To(BeTrue())
		})

		It("returns false for other errors", func() {
			Expect(isNotFoundError(apiErr{code: "AccessDenied"})).To(BeFalse())
			Expect(isNotFoundError(errors.New("plain error"))).To(BeFalse())
		})
	})
})

type fakeTransferManager struct {
	uploadInput *transfermanager.UploadObjectInput
	err         error
}

func (f *fakeTransferManager) UploadObject(ctx context.Context, params *transfermanager.UploadObjectInput, optFns ...func(*transfermanager.Options)) (*transfermanager.UploadObjectOutput, error) {
	f.uploadInput = params
	if f.err != nil {
		return nil, f.err
	}
	return &transfermanager.UploadObjectOutput{}, nil
}

type failingReadSeeker struct{}

func (f *failingReadSeeker) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func (f *failingReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

type seekErrorReadSeeker struct{}

func (s *seekErrorReadSeeker) Read(p []byte) (int, error) {
	copy(p, []byte("abc"))
	return 3, io.EOF
}

func (s *seekErrorReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek failed")
}

type apiErr struct {
	code string
}

func (e apiErr) Error() string                 { return e.code }
func (e apiErr) ErrorCode() string             { return e.code }
func (e apiErr) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }
func (e apiErr) ErrorMessage() string          { return e.code }
