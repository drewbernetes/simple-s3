# Simple S3

This repo is an abstraction of the AWS SDK to enable easy interaction with S3.

The point was to remove the fluff that is required to configure a client just so you can push, pull or list anything in
S3.

Now you just provide the access and secret keys, a region and a bucket, and you're off to the races.

It's pretty basic at the moment as it doesn't support versioning etc, but it's a start!
