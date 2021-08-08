# go-lambda-cleanup
[![Test](https://github.com/karl-cardenas-coding/go-lambda-cleanup/actions/workflows/test.yml/badge.svg)](https://github.com/karl-cardenas-coding/go-lambda-cleanup/actions/workflows/test.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)

<p align="center">
  <img src="/static/logo.png" alt="drawing" width="400"/>
</p>

<p align="center">A Golang based CLI for removing unused versions of AWS Lambdas. One binary,  no additional dependencies required. </p>

<p align="center">
<img src="/static/demo.gif" alt="drawing" width="400"/>
</p>


## Installation
go-lambda-cleanup is distributed as a single binary. [Download](https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases) the binary and install go-lambda-cleanup in a directory in your system's [PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them). `/usr/local/bin` is the recommended path for UNIX/LINUX environments. 

```shell
VERSION=1.0.8
wget https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v$VERSION/go-lambda-cleanup-v$VERSION-linux-amd64.zip
unzip go-lambda-cleanup-v$VERSION-linux-amd64.zip 
sudo mv glc /usr/local/bin/
```

## Usage

```shell
Usage:
  glc [flags]
  glc [command]

Available Commands:
  clean       Removes all former versions of AWS lambdas except for the $LATEST version
  help        Help about any command
  version     Print the current version number of disaster-cli

Flags:
  -d, --dryrun                    Executes a dry run (bool)
  -s, --enableSharedCredentials   Leverages the default ~/.aws/credentials file (bool)
  -h, --help                      help for glc
  -l, --listFile string           Specify a file containing Lambdas to delete.
  -p, --profile string            Specify the AWS profile to leverage for authentication.
  -r, --region string             Specify the desired AWS region to target.
  -v, --verbose                   Set to true to enable debugging (bool)

Use "glc [command] --help" for more information about a command.
```

To retain `2` version excluding `$LATEST`
```shell
glc clean -r us-east-2 -c 2 -s -p myProfile
```

You also have the ability to preview an execution by leveraging the dry run flag `-d`

```shell
 $ glc clean -s -p myProfile -r us-east-1 -d
INFO[03/19/21] The AWS Profile flag "myProfile" was passed in
INFO[03/19/21] ******** DRY RUN MODE ENABLED ********
INFO[03/19/21] Scanning AWS environment in us-east-1
INFO[03/19/21] ............
INFO[03/19/21] 50 Lambdas identified
INFO[03/19/21] Current storage size: 1.2 GB
INFO[03/19/21] **************************
INFO[03/19/21] Initiating clean-up process. This may take a few minutes....
INFO[03/19/21] ............
INFO[03/19/21] ............
INFO[03/19/21] 82 unique versions will be removed in an actual execution.
INFO[03/19/21] 554 MB of storage space will be removed in an actual execution.
INFO[03/19/21] Job Duration Time: 7.834585s
```

### Custom List
You can provide an input file contaitning a list of Lambda functions to be cleaned-up. The input file can be of the following types; `json`, `yaml`, or `yml.`  An input file allows you to control the execution more granularly. 

#### YAML
```yaml
# custom_list.yaml
lambdas:
  - stopEC2-instances
  - putControls
```

```shell
glc clean -r us-east-1 -sp myProfile -l custom_list.yaml
```

#### JSON
```json
{
    "lambdas": [
        "stopEC2-instances",
        "putControls"
    ]
}
```

```shell
glc clean -r us-east-1 -sp myProfile -l custom_list.json
```

### Authentication
go-lambda-clean utilizes the default AWS Go SDK credentials provider to find AWS credentials. The default provider chain looks for credentials in the following order:

1. Environment variables.

2. Shared credentials file.

3. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.

4. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.

#### Shared File Example
If `~/.aws/config` and `~/.aws/config` is setup for the AWS CLI then you may leverage the existing profile configuration for authentication.
```shell
$ export AWS_PROFILE=sb-test
$ glc clean -r us-west-2 -s
INFO[03/05/21] Scanning AWS environment in us-west-2
INFO[03/05/21] ............
```
Alternatively, the `--profile` flag may be used.
```shell
$ glc clean -r us-west-2 -s -p myProfile
INFO[03/05/21] Scanning AWS environment in us-west-2
INFO[03/05/21] ............
```

#### Environment Variables
Static credentials may be also be used to authenticate into AWS.
* AWS_ACCESS_KEY_ID

* AWS_SECRET_ACCESS_KEY

* AWS_SESSION_TOKEN (optional)
```shell
$ export AWS_ACCESS_KEY_ID=YOUR_AKID
$ export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
$ export AWS_SESSION_TOKEN=TOKEN
$ glc clean -r us-west-2
2021/03/04 20:42:46 Scanning AWS environment in us-west-2.....
2021/03/04 20:42:46 ............
```
## Compile
If you want to complile the binary, clone the project to your local system. Ensure you have `Go 1.16` installed. This tool leverages the Golang [embed](https://golang.org/pkg/embed/) functionality. A file named `aws-regions.txt` is expected in the `cmd/` directory.  You need valid AWS credentials in order to generate the file.
```shell
git clone git@github.com:karl-cardenas-coding/go-lambda-cleanup.git
aws ec2 describe-regions --region us-east-1 --query "Regions[].RegionName" --output text >> cmd/aws-regions.txt
go build -o glc
```

## Proxy
The tool supports network proxy configurations and will honor the following proxy environment variables.

* `HTTP_PROXY`,
* `HTTPS_PROXY`
* `NO_PROXY`

The environment values may be either a complete URL or a "host[:port]", in which case the "http" scheme is assumed. An error is returned if the value is a different form.

```shell
$ export HTTP_PROXY=http://proxy.example.org:9000

$ glc clean -r us-west-2
2021/03/04 20:42:46 Scanning AWS environment in us-west-2.....
2021/03/04 20:42:46 ............
```

## Contributing to go-lambda-cleanup

For a complete guide to contributing to go-lambda-clean, see the [Contribution Guide](documentation/CONTRIBUTING.md).

Contributions to go-lambda-cleanup of any kind including documentation, organization, tutorials, blog posts, bug reports, issues, feature requests, feature implementations, pull requests, answering questions on the forum, helping to manage issues, etc.

## FAQ
---
<table><tr>

Q: On MacOS I am unable to open the binary due to Apple not trusting the binary. What are my options?

A: You have two options. 

Option A is to clone this project and compile the binary. Issue `go build -o glc`, and the end result is a binary compatible for your system. If you still encounter issues after this, invoke the code signing command on the binary `codesign -s -`

Option B is not recommended but I'll offer it up. You can remove the binary from quarantine mode. 
```shell
xattr -d com.apple.quarantine /path/to/file
```
---

Q: This keeps timing out when attempting to connect to AWS and I have verified my AWS credentials are valid?

A: This could be related to a corporate firewall. If your organization has a proxy endpoint configure the proxy environment variable with the correct proxy endpoint. Consult your organization's networking team to learn more about the proper proxy settings.

---

Q: I don't want to execute this command without understanding exactly what it will do. Is there a way to preview the actions?

A: Yes, leverage the dry run mode. Dry run can be invoked through the `-d`, `--dryrun` flag.

---

</tr></table>

## Helpful Links
[AWS Credentials Configuration](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

[Golang Cobra CLI Framework](https://github.com/spf13/cobra)

[AWS Go SDK Credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)
