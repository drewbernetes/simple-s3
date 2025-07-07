# Simple S3

A simple Go package for interacting with S3-compatible object storage.

## Usage

This package provides a simplified interface for interacting with AWS S3 or S3-compatible storage services.

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


