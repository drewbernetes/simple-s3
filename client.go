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
    "net/url"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    transport "github.com/aws/smithy-go/endpoints"
)

// For testing purposes, allowing this function to be replaced in tests
var newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
    return s3.NewFromConfig(cfg, optFns...)
}

// staticResolver provides custom endpoint resolution for non-AWS S3Client endpoints
type staticResolver struct {
    URL *url.URL
}

// ResolveEndpoint implements the EndpointResolverV2 interface
func (r *staticResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
    if r.URL != nil {
        u := *r.URL
        // Only append bucket to path if it's provided in the parameters
        if params.Bucket != nil {
            u.Path += "/" + *params.Bucket
        }
        return transport.Endpoint{URI: u}, nil
    }

    // Fall back to the default resolver if no custom endpoint is provided
    return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// New generates a new EndpointWithResolverOptions and returns an S3Client containing the Bucket and S3Client.
func New(ctx context.Context, endpoint, region, accessKey, secretKey string) (S3Interface, error) {
    if region == "" {
        region = "us-east-1"
    }

    // Configure AWS options
    options := []func(*config.LoadOptions) error{
        config.WithRegion(region),
    }

    // Add credentials if provided
    if accessKey != "" && secretKey != "" {
        options = append(options, config.WithCredentialsProvider(
            credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
        ))
    }

    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(ctx, options...)
    if err != nil {
        return nil, err
    }

    // Create a Client with standard config
    var client *s3.Client
    customEndpoint := false

    // If a custom endpoint is provided, configure the Client to use it
    if endpoint != "" {
        var ep *url.URL
        ep, err = url.Parse(endpoint)
        if err != nil {
            return nil, err
        }
        customEndpoint = true

        client = newS3ClientFromConfig(cfg, func(o *s3.Options) {
            o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
            o.EndpointResolverV2 = &staticResolver{URL: ep}
        })
    } else {
        // Use default AWS S3Client endpoint
        client = newS3ClientFromConfig(cfg)
    }

    return &S3Client{Client: client, customEndpoint: customEndpoint}, nil
}
