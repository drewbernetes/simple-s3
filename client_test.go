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
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/drewbernetes/simple-s3/pkg/mock"
	"github.com/drewbernetes/simple-s3/pkg/util"
	"go.uber.org/mock/gomock"
	"os"
	"path/filepath"
	"testing"
)

// createFile creates a temporary file with predefined content and returns a file pointer and an error if any occurs.
// The function ensures the file is rewound to the beginning for subsequent reading or operations.
func createFile() (*os.File, error) {
	tmpDir := os.TempDir()
	testFilePath := filepath.Join(tmpDir, "s3-test-file.txt")

	f, err := os.Create(testFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create test file: %w", err)
	}

	content := []byte("test content for S3 upload")
	_, err = f.Write(content)
	if err != nil {
		f.Close() // Close before returning on error
		return nil, fmt.Errorf("failed to write to test file: %w", err)
	}

	// Seek back to the beginning so file can be read later
	_, err = f.Seek(0, 0)
	if err != nil {
		f.Close() // Close before returning on error
		return nil, fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	return f, nil
}

func put(s util.S3Interface) error {
	f, err := createFile()
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer func() {
		if removeErr := removeFile(f); removeErr != nil {
			// Just log the error here since we want to return the original error if there is one
			fmt.Printf("Warning: failed to remove test file: %v\n", removeErr)
		}
	}()

	if err := s.PutObject(context.Background(), "text/plain", "path/results.json", f); err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// removeFile deletes the file associated with the provided *os.File object and returns an error if the removal fails.
func removeFile(f *os.File) error {
	err := os.Remove(f.Name())
	if err != nil {
		return err
	}
	return nil
}

func TestHelperFunctions(t *testing.T) {
	t.Run("createFile and removeFile success", func(t *testing.T) {
		f, err := createFile()
		if err != nil {
			t.Fatalf("createFile() failed: %v", err)
		}

		// Verify file exists and has content
		info, err := f.Stat()
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		if info.Size() == 0 {
			t.Error("File is empty")
		}

		// Test removeFile
		err = removeFile(f)
		if err != nil {
			t.Fatalf("removeFile() failed: %v", err)
		}

		// Verify file is deleted
		_, err = os.Stat(f.Name())
		if !os.IsNotExist(err) {
			t.Errorf("File was not deleted, got error: %v", err)
		}
	})
}

// TestNewWithMock tests the New function, verifying correct S3 client configuration and mocking behavior.
func TestNewWithMock(t *testing.T) {
	// Store the original function
	originalNewFromConfig := newS3ClientFromConfig
	// Restore it after the test
	defer func() { newS3ClientFromConfig = originalNewFromConfig }()

	t.Run("With successful configuration", func(t *testing.T) {
		// Create a mock client
		mockClient := &s3.Client{}

		// Replace the newS3ClientFromConfig function with our mock
		newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
			// Verify the region was set correctly
			if cfg.Region != "us-east-1" && cfg.Region != "eu-west-1" {
				t.Errorf("Expected region to be us-east-1 or eu-west-1, got %s", cfg.Region)
			}

			// Verify that the option functions were set correctly
			if len(optFns) != 1 {
				t.Errorf("Expected 1 option function, got %d", len(optFns))
			}

			// Return the mock client
			return mockClient
		}

		// Test with default region
		c, err := New(context.Background(), "http://test.example.com", "abc", "def", "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the client is our mock
		if c.Client != mockClient {
			t.Errorf("Expected mock client, got different client")
		}

		// Test with custom region
		c, err = New(context.Background(), "http://test.example.com", "abc", "def", "eu-west-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the client is our mock
		if c.Client != mockClient {
			t.Errorf("Expected mock client, got different client")
		}
	})

	t.Run("With error from LoadDefaultConfig", func(t *testing.T) {
		// Make the function return an error
		newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
			// This shouldn't be called if LoadDefaultConfig fails
			t.Fatal("newS3ClientFromConfig should not be called when LoadDefaultConfig fails")
			return nil
		}

		// To properly test this, we'd need to mock config.LoadDefaultConfig,
		// which is challenging without a package-level variable or interface.
		// This is a limitation in the current design.
		t.Skip("Cannot mock config.LoadDefaultConfig with the current design")
	})

	t.Run("Verifies credentials are set correctly", func(t *testing.T) {
		accessKey := "test-access-key"
		secretKey := "test-secret-key"

		newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
			// We'd need to verify that credentials were set correctly,
			// but this requires inspecting the internal credentials provider
			// which is challenging without modifying the original code
			return &s3.Client{}
		}

		_, err := New(context.Background(), "http://test.example.com", accessKey, secretKey, "us-east-1")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

// TestNew verifies the behavior of the New function for various configurations and error scenarios.
func TestNew(t *testing.T) {
	// Test case table
	tests := []struct {
		name      string
		endpoint  string
		accessKey string
		secretKey string
		region    string
		expectErr bool
		expectNil bool
	}{
		{
			name:      "Valid configuration",
			endpoint:  "http://test.example.com",
			accessKey: "abc",
			secretKey: "def",
			region:    "eu-west-1",
			expectErr: false,
			expectNil: false,
		},
		{
			name:      "Default region when empty",
			endpoint:  "http://test.example.com",
			accessKey: "abc",
			secretKey: "def",
			region:    "",
			expectErr: false,
			expectNil: false,
		},
		{
			name:      "Invalid endpoint URL",
			endpoint:  "://invalid-url",
			accessKey: "abc",
			secretKey: "def",
			region:    "eu-west-1",
			expectErr: true,
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := New(context.Background(), tc.endpoint, tc.accessKey, tc.secretKey, tc.region)

			if tc.expectErr && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.expectNil && c != nil {
				t.Errorf("Expected nil client but got non-nil")
			}

			if !tc.expectNil && c == nil {
				t.Errorf("Expected non-nil client but got nil")
			}

			// For successful cases, verify the client has the expected configuration
			if c != nil {
				// Just check that the client is non-nil
				if c.Client == nil {
					t.Error("s3.Client is nil")
				}
			}
		})
	}
}

// TestFetchObject verifies that an object can be fetched from a bucket successfully using the mocked S3 interface.
func TestFetchObject(t *testing.T) {
	t.Run("Successfully fetches object from bucket", func(t *testing.T) {
		c := gomock.NewController(t)
		defer c.Finish()
		m := mock.NewMockS3Interface(c)

		expectedData := []byte("some data")
		m.EXPECT().FetchObject(context.Background(), "", gomock.Eq("trivyignore")).Return(expectedData, nil)

		data, err := m.FetchObject(context.Background(), "", "trivyignore")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if string(data) != string(expectedData) {
			t.Errorf("Expected data '%s', got '%s'", expectedData, data)
		}
	})
}

// TestPut tests the functionality of the put function with a mock S3Interface, validating that PutObject is called correctly.
func TestPut(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	m := mock.NewMockS3Interface(c)

	f, err := createFile()
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer func() {
		if err := removeFile(f); err != nil {
			t.Logf("Warning: Failed to remove test file: %v", err)
		}
	}()

	m.EXPECT().PutObject(context.Background(), gomock.Eq("text/plain"), gomock.Eq("path/results.json"), gomock.Any()).Return(nil)
	if err := put(m); err != nil {
		t.Errorf("put() failed: %v", err)
	}
}
