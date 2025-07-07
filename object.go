package simple_s3

import (
    "bytes"
    "context"
    "io"
    "net/http"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

// fileInfo contains information on a file being uploaded to S3Client.
type fileInfo struct {
    buffer      []byte
    size        int64
    contentType string
}

// readFile Reads the file into a buffer and returns the buffer along with the file content type.
func readFile(body *os.File) (*fileInfo, error) {
    info, _ := body.Stat()
    size := info.Size()
    buffer := make([]byte, size)
    _, err := body.Read(buffer)
    if err != nil {
        return nil, err
    }

    // Detect content type after reading data into buffer
    fileType := http.DetectContentType(buffer)

    fi := &fileInfo{
        buffer:      buffer,
        contentType: fileType,
        size:        size,
    }

    return fi, nil
}

// GetObject Downloads a file from an S3Client bucket and returns its contents as a byte array.
func (s *S3Client) GetObject(ctx context.Context, fileName, bucket string) ([]byte, error) {
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

// ListObject will list the contents of a bucket.
func (s *S3Client) ListObject(ctx context.Context, bucket, prefix string) ([]string, error) {
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

// PutObject Pushes a file to an S3Client bucket. If the object is greater than 100MB it will be split into a multipart upload.
func (s *S3Client) PutObject(ctx context.Context, bucket, key string, body *os.File) *PutUploadResult {
    var uploadRes *manager.UploadOutput
    var putRes *s3.PutObjectOutput
    result := &PutUploadResult{}

    fi, err := readFile(body)
    if err != nil {
        result.Error = err
        return result
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
        uploadRes, err = uploader.Upload(ctx, params)
        if err != nil {
            result.Error = err
            return result
        }
    } else {
        putRes, err = s.Client.PutObject(ctx, params)
        if err != nil {
            result.Error = err
            return result
        }
    }

    result.UploadResults = uploadRes
    result.PutResults = putRes

    return result
}

// DeleteObject removes a single object from an S3Client bucket
func (s *S3Client) DeleteObject(ctx context.Context, bucket, key string) (*s3.DeleteObjectOutput, error) {
    input := &s3.DeleteObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    }

    return s.Client.DeleteObject(ctx, input)
}
