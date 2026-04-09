# Simple S3

A simple Go package for interacting with S3-compatible object storage.

## Installation

```bash
go get github.com/drewbernetes/simple-s3
```

## Usage

This package provides a simplified interface for interacting with AWS S3 or S3-compatible storage services.

### Creating a Client

```go
package main

import (
	"context"

	simple_s3 "github.com/drewbernetes/simple-s3"
)

func main() {
	ctx := context.Background()

	// Connect to AWS S3
	client, err := simple_s3.New(ctx, "", "YOUR_ACCESS_KEY", "YOUR_SECRET_KEY", "us-west-2")
	if err != nil {
		panic(err)
	}

	// Or connect to an S3-compatible service (MinIO, LocalStack, etc.)
	client, err = simple_s3.New(ctx, "http://localhost:9000", "minioadmin", "minioadmin", "")
	if err != nil {
		panic(err)
	}

	_ = client
}
```

### Bucket Operations

```go
// Create a bucket
err := client.CreateBucket(ctx, "my-bucket")

// List buckets (with optional prefix filter)
buckets, err := client.ListBuckets(ctx, "prod-")

// Delete a bucket (also removes all objects inside it)
err = client.DeleteBucket(ctx, "my-bucket")
```

### Object Operations

```go
import (
	"bytes"
	"os"
)

// Upload from a byte slice
data := []byte("hello world")
err := client.PutObject(ctx, "my-bucket", "path/to/object.txt", bytes.NewReader(data))

// Upload a file
f, err := os.Open("photo.jpg")
if err != nil {
	panic(err)
}
defer f.Close()
err = client.PutObject(ctx, "my-bucket", "images/photo.jpg", f)

// Download an object
content, err := client.FetchObject(ctx, "path/to/object.txt", "my-bucket")

// List objects (with optional prefix filter)
keys, err := client.ListObject(ctx, "my-bucket", "images/")

// Delete an object
err = client.DeleteObject(ctx, "my-bucket", "path/to/object.txt")
```

### Mocking for Tests

An `S3Interface` is provided for dependency injection and testing:

```go
import "github.com/drewbernetes/simple-s3/pkg/util"

type MyService struct {
	storage util.S3Interface
}
```

Mock generation (requires [mockgen](https://github.com/uber-go/mock)):

```bash
mockgen -source=pkg/util/interfaces.go -destination=pkg/mock/interfaces.go -package=mock
```

## Development
### Update the Changelog

Get yourself a GitHub access token with permissions to read the repository, if you don't already have one.

```shell
gh auth login
gh auth token
```

Run [git cliff](https://github.com/orhun/git-cliff/)

```
export GITHUB_TOKEN=<token> # You can also add this to your ~/.bashrc or ~/.zshrc etc
git cliff -o
```
It's worth noting that `--bump` will update the changelog with what it thinks will be the next release. Make sure to check this and ensure your next tag matches this value.
See [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)) for information on commit message style.

```shell
git tag | sort -V | tail -n1
```

### Running Tests

```bash
go test -v -cover ./...
```

## License

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
Once tested and validated using your branch, get the next available tag by running the following command, and incrementing by one. e.g if this output `v0.1.31`, you should use v0.1.32.


