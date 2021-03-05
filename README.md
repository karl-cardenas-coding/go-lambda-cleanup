# go-lambda-cleanup
[![Actions Status](https://github.com/karl-cardenas-coding/go-lambda-cleanup/workflows/Go/badge.svg?branch=main)](https://github.com/karl-cardenas-coding/go-lambda-cleanup/actions?branch=main)
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)

<p align="center">
  <img src="/static/logo.jpg" alt="drawing" width="400"/>
</p>

A Golang based CLI for removing unused versions of AWS Lambdas. 

## Installation
go-lambda-cleanup is distributed as a single binary. [Download](https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases) the binary and install go-lambda-cleanup by unzipping the zip file and and moving the included binary a directory in your system's [PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them). `/usr/local/bin` is the recommended path for UNIX/LINUX environments. 

## Usage

```shell
Usage:
  gcl clean [flags]

Flags:
  -c, --count int32   The number of versions to retain from $LATEST - n-(x) (int) (default 1)
  -h, --help          help for clean

Global Flags:
  -s, --enableSharedCredentials   Leverages the default ~/.aws/crededentials file (bool)
  -r, --region string             Specify the desired AWS region to target.
  -v, --verbose                   Set to true to enable debugging (bool)
```


## Contributing to go-lambda-cleanup

For a complete guide to contributing to disaster-cli , see the [Contribution Guide](CONTRIBUTING.md).

Contributions to go-lambda-cleanup of any kind including documentation, organization, tutorials, blog posts, bug reports, issues, feature requests, feature implementations, pull requests, answering questions on the forum, helping to manage issues, etc.


## Helpful Links

Golang Cobra CLI Framework:https://github.com/spf13/cobra
