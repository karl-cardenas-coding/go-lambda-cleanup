# go-lambda-cleanup
[![Actions Status](https://github.com/karl-cardenas-coding/go-lambda-cleanup/workflows/test.yml/badge.svg?branch=main)](https://github.com/karl-cardenas-coding/go-lambda-cleanup/actions?branch=main)
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)

<p align="center">
  <img src="/static/logo.png" alt="drawing" width="400"/>
</p>

<p align="center">A  Golang based CLI for removing unused versions of AWS Lambdas. </p>

<p align="center">
<img src="/static/demo.gif" alt="drawing" width="400"/>
</p>


## Installation
go-lambda-cleanup is distributed as a single binary. [Download](https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases) the binary and install go-lambda-cleanup by unzipping the zip file and and moving the included binary to a directory in your system's [PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them). `/usr/local/bin` is the recommended path for UNIX/LINUX environments. 

```shell
wget https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v1.0.0/go-lambda-cleanup_linux_amd64.zip
unzip go-lambda-cleanup_linux_amd64.zip 
sudo mv go-clean-lambda /usr/local/bin/
```

## Usage

```shell
Usage:
  glc clean [flags]

Flags:
  -c, --count int8   The number of versions to retain from $LATEST-(n) (default 1)
  -h, --help         help for clean

Global Flags:
  -s, --enableSharedCredentials   Leverages the default ~/.aws/credentials file (bool)
  -p, --profile string            Specify the AWS profile to leverage for authentication.
  -r, --region string             Specify the desired AWS region to target.
  -v, --verbose                   Set to true to enable debugging (bool)

Use "gcl [command] --help" for more information about a command.
```

To retain `2` version excluding `$LATEST`
```shell
glc clean -r us-east-2 -c 2 -s true -p myProfile
```

### Authentication
go-lambda-clean utilizes the default AWS Go SDK credentials provider to find AWS credentials. The default provider chain looks for credentials in the following order:

1. Environment variables.

2. Shared credentials file.

3. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.

4. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.

#### Shared File Example
If `~/.aws/config` and `~/.aws/config` is setup for the AWS CLI then you may leverage the existing profile confugrations for authentication.
```shell
$ export AWS_PROFILE=sb-test
$ glc clean -r us-west-2 -s true
INFO[03/05/21] Scanning AWS environment in us-west-2
INFO[03/05/21] ............
```
Alternatively, the `--profile` flag may be used.
```shell
$ glc clean -r us-west-2 -s true -p myProfile
INFO[03/05/21] Scanning AWS environment in us-west-2
INFO[03/05/21] ............
```

#### Environment Variables
Static credentials may be utlized to authenticate into AWS.
* AWS_ACCESS_KEY_ID

* AWS_SECRET_ACCESS_KEY

* AWS_SESSION_TOKEN (optional)
```shell
$ export AWS_ACCESS_KEY_ID=YOUR_AKID
$ export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
$ export AWS_SESSION_TOKEN=TOKEN
$ gcl clean -r us-west-2
2021/03/04 20:42:46 Scanning AWS environment in us-west-2.....
2021/03/04 20:42:46 ............
```

## Contributing to go-lambda-cleanup

For a complete guide to contributing to go-lambda-clean, see the [Contribution Guide](documentation/CONTRIBUTING.md).

Contributions to go-lambda-cleanup of any kind including documentation, organization, tutorials, blog posts, bug reports, issues, feature requests, feature implementations, pull requests, answering questions on the forum, helping to manage issues, etc.


## Helpful Links
[AWS Credentials Configuration](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
[Golang Cobra CLI Framework](https://github.com/spf13/cobra)
[AWS Go SDK Credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)
