# Simple S3

This repo is an abstraction of the AWS SDK to enable easy interaction with S3.

The point was to remove the fluff that is required to configure a client just so you can push, pull or list anything in
S3.

Now you just provide the access and secret keys, a region, a bucket, and an optional endpoint, if you are using your 
own S3 storage, and you're off to the races.

It's pretty basic at the moment as it does not support versioning etc, but it's a start!

I may go further with it, I may not - it purely depends on my requirements and time available.


## Update the Changelog

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

Once tested and validated using your branch, get the next available tag by running the following command, and incrementing by one. e.g if this output `v0.1.31`, you should use v0.1.32.

```shell
git tag | sort -V | tail -n1
```
