//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	simple_s3 "github.com/drewbernetes/simple-s3"
)

var _ = Describe("S3 Integration", Ordered, func() {
	var (
		client   *simple_s3.S3
		ctx      context.Context
		bucket   string
		endpoint string
	)

	BeforeAll(func() {
		endpoint = os.Getenv("LOCALSTACK_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://localhost:4566"
		}

		if os.Getenv("S3_INTEGRATION_TEST") != "true" {
			Skip("S3_INTEGRATION_TEST not set — skipping integration tests")
		}

		ctx = context.Background()
		bucket = fmt.Sprintf("integration-test-%d", time.Now().UnixNano())

		var err error
		client, err = simple_s3.New(ctx, endpoint, "test", "test", "us-east-1")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		if client != nil {
			_ = client.DeleteBucket(ctx, bucket)
		}
	})

	It("should create a bucket", func() {
		err := client.CreateBucket(ctx, bucket)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should list the created bucket", func() {
		buckets, err := client.ListBuckets(ctx, "integration-test-")
		Expect(err).NotTo(HaveOccurred())

		found := false
		for _, b := range buckets.Buckets {
			if b.Name != nil && *b.Name == bucket {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "expected bucket %s in list", bucket)
	})

	It("should upload an object", func() {
		body := bytes.NewReader([]byte("integration-test-payload"))
		err := client.PutObject(ctx, bucket, "test-key.txt", body)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should list objects", func() {
		keys, err := client.ListObject(ctx, bucket, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(keys).To(ContainElement("test-key.txt"))
	})

	It("should fetch the uploaded object", func() {
		data, err := client.FetchObject(ctx, "test-key.txt", bucket)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("integration-test-payload"))
	})

	It("should delete a single object", func() {
		err := client.DeleteObject(ctx, bucket, "test-key.txt")
		Expect(err).NotTo(HaveOccurred())

		keys, err := client.ListObject(ctx, bucket, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(keys).NotTo(ContainElement("test-key.txt"))
	})

	It("should cascade-delete a bucket with objects", func() {
		// Put a few objects back in
		for i := 0; i < 3; i++ {
			body := bytes.NewReader([]byte(fmt.Sprintf("payload-%d", i)))
			err := client.PutObject(ctx, bucket, fmt.Sprintf("obj-%d", i), body)
			Expect(err).NotTo(HaveOccurred())
		}

		err := client.DeleteBucket(ctx, bucket)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should no-op when deleting a non-existent bucket", func() {
		err := client.DeleteBucket(ctx, "nonexistent-bucket-that-should-not-exist")
		Expect(err).NotTo(HaveOccurred())
	})
})
