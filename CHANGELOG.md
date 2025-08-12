# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-08-12

### :bug: Fixed
- Fix correcting the reference to the bot ID in the changelog pipeline by @drew-viles in [#53](https://github.com/drewbernetes/simple-s3/pull/53)

### :rocket: Added
- Adding ability for user to override the default chunk sizing for multipart uploads & updating go modules by @drew-viles in [#52](https://github.com/drewbernetes/simple-s3/pull/52)


## [1.0.0] - 2025-08-01

### :gear: Changed
- Bump github.com/aws/aws-sdk-go-v2/feature/s3/manager by @dependabot[bot] in [#41](https://github.com/drewbernetes/simple-s3/pull/41)
- Updating modules by @drew-viles in [#36](https://github.com/drewbernetes/simple-s3/pull/36)
- Bump github.com/aws/aws-sdk-go-v2/feature/s3/manager by @dependabot[bot] in [#31](https://github.com/drewbernetes/simple-s3/pull/31)
- Bump github.com/aws/aws-sdk-go-v2/service/s3 by @dependabot[bot] in [#32](https://github.com/drewbernetes/simple-s3/pull/32)
- Bump github.com/aws/aws-sdk-go-v2/config by @dependabot[bot] in [#33](https://github.com/drewbernetes/simple-s3/pull/33)
- Bump github.com/aws/aws-sdk-go-v2/feature/s3/manager by @dependabot[bot] in [#27](https://github.com/drewbernetes/simple-s3/pull/27)
- Bump github.com/aws/aws-sdk-go-v2/credentials by @dependabot[bot] in [#28](https://github.com/drewbernetes/simple-s3/pull/28)

### :rocket: Added
- Updating to support bucket interactions and huge refactor to support bypassing aws checksum validation when not using aws s3 by @drew-viles in [#42](https://github.com/drewbernetes/simple-s3/pull/42)


## [0.1.7] - 2025-05-30

### :gear: Changed
- Updating modules by @drew-viles


## [0.1.6] - 2025-05-30

### :gear: Changed
- Bump github.com/aws/aws-sdk-go-v2/feature/s3/manager by @dependabot[bot] in [#21](https://github.com/drewbernetes/simple-s3/pull/21)
- Bump github.com/aws/aws-sdk-go-v2/service/s3 from 1.79.2 to 1.79.3 by @dependabot[bot] in [#20](https://github.com/drewbernetes/simple-s3/pull/20)
- Bump go.uber.org/mock from 0.5.1 to 0.5.2 by @dependabot[bot] in [#19](https://github.com/drewbernetes/simple-s3/pull/19)
- Create dependabot.yml by @drew-viles in [#17](https://github.com/drewbernetes/simple-s3/pull/17)
- Go mod tidy run by @drew-viles in [#16](https://github.com/drewbernetes/simple-s3/pull/16)
- Module update by @drew-viles in [#15](https://github.com/drewbernetes/simple-s3/pull/15)
- Bump github.com/aws/aws-sdk-go-v2/feature/s3/manager by @dependabot[bot] in [#10](https://github.com/drewbernetes/simple-s3/pull/10)
- Create dependabot.yml by @drew-viles in [#9](https://github.com/drewbernetes/simple-s3/pull/9)
- Correcting the way the region is parsed and update go modules by @drew-viles in [#8](https://github.com/drewbernetes/simple-s3/pull/8)

### :rocket: Added
- Added ability create buckets, to delete files and buckets by @drew-viles in [#18](https://github.com/drewbernetes/simple-s3/pull/18)


## New Contributors
* @dependabot[bot] made their first contribution in [#21](https://github.com/drewbernetes/simple-s3/pull/21)
## [0.1.5] - 2024-10-21

### :gear: Changed
- Changelog update by @drew-viles
- Module updates by @drew-viles in [#7](https://github.com/drewbernetes/simple-s3/pull/7)


## [0.1.4] - 2024-09-04

### :rocket: Added
- Adding codeowners by @drew-viles in [#6](https://github.com/drewbernetes/simple-s3/pull/6)


## [0.1.3] - 2024-08-21

### :gear: Changed
- Endpoint fix by @drew-viles in [#5](https://github.com/drewbernetes/simple-s3/pull/5)


## [0.1.2] - 2024-08-21

### :bug: Fixed
- Fixing release process by @drew-viles in [#3](https://github.com/drewbernetes/simple-s3/pull/3)

### :gear: Changed
- Changelog update by @drew-viles
- Updating go builder action in pipeline by @drew-viles in [#4](https://github.com/drewbernetes/simple-s3/pull/4)


## [0.1.1] - 2024-08-20

### :gear: Changed
- Updated modules and go version by @drew-viles in [#2](https://github.com/drewbernetes/simple-s3/pull/2)


## [0.1.0] - 2024-02-14

### :rocket: Added
- Added support for multipart upload by @drew-viles in [#1](https://github.com/drewbernetes/simple-s3/pull/1)
- Added support for multipart upload by @drew-viles
- Added prefix option to list command by @drew-viles
- Adding initial code by @drew-viles


## New Contributors
* @drew-viles made their first contribution in [#1](https://github.com/drewbernetes/simple-s3/pull/1)
[1.1.0]: https://github.com/drewbernetes/simple-s3/compare/v1.0.0..v1.1.0
[1.0.0]: https://github.com/drewbernetes/simple-s3/compare/v0.1.7..v1.0.0
[0.1.7]: https://github.com/drewbernetes/simple-s3/compare/v0.1.6..v0.1.7
[0.1.6]: https://github.com/drewbernetes/simple-s3/compare/v0.1.5..v0.1.6
[0.1.5]: https://github.com/drewbernetes/simple-s3/compare/v0.1.4..v0.1.5
[0.1.4]: https://github.com/drewbernetes/simple-s3/compare/v0.1.3..v0.1.4
[0.1.3]: https://github.com/drewbernetes/simple-s3/compare/v0.1.2..v0.1.3
[0.1.2]: https://github.com/drewbernetes/simple-s3/compare/v0.1.1..v0.1.2
[0.1.1]: https://github.com/drewbernetes/simple-s3/compare/v0.1.0..v0.1.1

<!-- generated by git-cliff -->
