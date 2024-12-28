# go-lambda-cleanup
[![Go Reference](https://pkg.go.dev/badge/github.com/karl-cardenas-coding/go-lambda-cleanup.svg)](https://pkg.go.dev/github.com/karl-cardenas-coding/go-lambda-cleanup/v2)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  
[![Go version](https://img.shields.io/github/go-mod/go-version/karl-cardenas-coding/go-lambda-cleanup)](https://golang.org/dl/)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![codecov](https://codecov.io/gh/karl-cardenas-coding/go-lambda-cleanup/graph/badge.svg?token=S8SY4ZA2ZA)](https://codecov.io/gh/karl-cardenas-coding/go-lambda-cleanup)
[![Go Report Card](https://goreportcard.com/badge/github.com/karl-cardenas-coding/go-lambda-cleanup/v2)](https://goreportcard.com/report/github.com/karl-cardenas-coding/go-lambda-cleanup/v2)


<p align="center">
  <img src="/static/logo.png" alt="drawing" width="400"/>
</p>

<p align="center">A Go based CLI for removing unused versions of AWS Lambdas. One binary, no additional dependencies required. </p>

<p align="center">
<img src="/static/glc.gif" alt="drawing" width="800"/>
</p>


## Installation
go-lambda-cleanup is distributed as a single binary. [Download](https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases) the binary and install go-lambda-cleanup in a directory in your system's [PATH](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them). `/usr/local/bin` is the recommended path for UNIX/LINUX environments. 

```shell
VERSION=2.0.15
wget https://github.com/karl-cardenas-coding/go-lambda-cleanup/releases/download/v$VERSION/go-lambda-cleanup-v$VERSION-linux-amd64.zip
unzip go-lambda-cleanup-v$VERSION-linux-amd64.zip 
sudo mv glc /usr/local/bin/
```


## Docker
go-lambda-cleanup is also available as a Docker image. Check out the [GitHub Packages](https://github.com/karl-cardenas-coding/go-lambda-cleanup/pkgs/container/go-lambda-cleanup) page for this repository to learn more about the available images.

```
VERSION=v2.0.15
docker pull ghcr.io/karl-cardenas-coding/go-lambda-cleanup:$VERSION
```

You can pass AWS credentials to the container through ENVIRONMENT variables.
```
export AWS_ACCESS_KEY_ID=47as12fdsdg....
export AWS_SECRET_ACCESS_KEY=21a5sf5dg8e...

docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY ghcr.io/karl-cardenas-coding/go-lambda-cleanup:$VERSION clean -r us-east-1 -d
time=05/23/22 level=info msg="******** DRY RUN MODE ENABLED ********"
time=05/23/22 level=info msg="Scanning AWS environment in us-east-1"
time=05/23/22 level=info msg=............
time=05/23/22 level=info msg="8 Lambdas identified"
time=05/23/22 level=info msg="Current storage size: 193 MB"
time=05/23/22 level=info msg="**************************"
time=05/23/22 level=info msg="Initiating clean-up process. This may take a few minutes...."
time=05/23/22 level=info msg=............
time=05/23/22 level=info msg=............
time=05/23/22 level=info msg="24 unique versions will be removed in an actual execution."
time=05/23/22 level=info msg="124 MB of storage space will be removed in an actual execution."
time=05/23/22 level=info msg="Job Duration Time: 1.454406s"
```

### Nightly Release

A nightly release is available as a Docker image. The nightly release is a snapshot of the main branch and automatically updated with the latest minor depedencies updates.  

```shell
docker pull ghcr.io/karl-cardenas-coding/go-lambda-cleanup:nightly
```



## Usage

```shell
Usage:
  glc [flags]
  glc [command]

Available Commands:
  clean       Removes all former versions of AWS lambdas except for the $LATEST version
  help        Help about any command
  version     Print the current version number of glc

Flags:
  -d, --dryrun                    Executes a dry run (bool)
  -h, --help                      help for glc
  -l, --listFile string           Specify a file containing Lambdas to delete.
  -m, --moreLambdaDetails         Set to true to show Lambda names and count of versions to be removed (bool)
  -p, --profile string            Specify the AWS profile to leverage for authentication.
  -r, --region string             Specify the desired AWS region to target.
  -i, --size-iec                  Displays file sizes in IEC units (bool)
  -v, --verbose                   Set to true to enable debugging (bool)

Use "glc [command] --help" for more information about a command.
```

### Versions Retention 

To retain previous version excluding `$LATEST`, use the `-c` flag. Use this flag to control the number of versions to retain.
```shell
$ glc clean -r us-east-2 -c 2 -p myProfile
```

### Additonal Lambda Details
To view additional details, such as the Lambda names and version counts, set the `-m` flag to true.

```shell
glc clean -r us-east-1 -dmp mySuperAwesomeProfile
```

```shell
INFO[07/31/22] Scanning AWS environment in us-east-1
INFO[07/31/22] ............
INFO[07/31/22] 72 Lambdas identified
INFO[07/31/22] Current storage size: 817 MiB
INFO[07/31/22] **************************
INFO[07/31/22] Initiating clean-up process. This may take a few minutes....
INFO[07/31/22] ............
INFO[07/31/22]     1 versions of YourLambda to be removed
INFO[07/31/22]     1 versions of AnotherLambda to be removed
INFO[07/31/22] ............
INFO[07/31/22] 2 unique versions will be removed in an actual execution.
INFO[07/31/22] 12 MiB of storage space will be removed in an actual execution.
INFO[07/31/22] Job Duration Time: 9.409405s
```


### Lambda Aliases

AWS disallows the deletion of Lambda versions that are attached to an alias. The default behavior of `glc` is to attempt to delete a Lambda version, regardless of whether it has an alias attachment. If a Lambda version is attached to an alias and `glc` attempts to delete the version, an error will occur, and the program will exit with a non-zero exit code.

You can use the CLI flags `--skip-aliases` or `-s` to check
the Lambda version for the existence of aliases and skip the removal step if an alias is attached to the version. This check entails one additional API query per lambda, so consider not enabling this functionality if you do not use aliases.

## Compile
If you want to complile the binary, clone the project to your local system. Ensure you have `Go 1.18` installed. This tool leverages the Golang [embed](https://golang.org/pkg/embed/) functionality. A file named `aws-regions.txt` is expected in the `cmd/` directory.  You need valid AWS credentials in order to generate the file.
```shell
git clone git@github.com:karl-cardenas-coding/go-lambda-cleanup.git
aws ec2 describe-regions --region us-east-1 --all-regions --query "Regions[].RegionName" --output text >> cmd/aws-regions.txt
go build -o glc
```

### File Size Format
To display file sizes in [IEC format](https://en.wikipedia.org/wiki/Binary_prefix), enable the `-i` flag.
```shell
$ glc clean -i -r us-east-1 -p myProfile
```

### Dry Run

You also have the ability to preview an execution by leveraging the dry run flag `-d`

```shell
 $ glc clean -p myProfile -r us-east-1 -d
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
You can provide an input file containing a list of Lambda functions to be cleaned-up. The input file can be of the following types; `json`, `yaml`, or `yml.`  An input file allows you to control the execution more granularly. 

#### YAML
```yaml
# custom_list.yaml
lambdas:
  - stopEC2-instances
  - putControls
```

```shell
$ glc clean -r us-east-1 -p myProfile -l custom_list.yaml
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
glc clean -r us-east-1 -p myProfile -l custom_list.json
```


### IAM Permissions

go-lambda-cleanup requires the following IAM permissions to operate. 

- `lambda:ListFunctions`
- `lambda:ListVersionsByFunction`
- `lambda:ListAliases`
- `lambda:DeleteFunction`

The following code snippet is an IAM policy you may assign to the IAM User or IAM Role used by go-lambda-cleanup.


```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "goLambdaCleanup",
            "Effect": "Allow",
            "Action": [
                "lambda:ListFunctions",
                "lambda:ListVersionsByFunction",
                "lambda:ListAliases",
                "lambda:DeleteFunction"
            ],
            "Resource": "*"
        }
    ]
}
```

### Authentication
go-lambda-clean utilizes the default AWS Go SDK credentials provider to find AWS credentials. The default provider chain looks for credentials in the following order:

1. Environment variables.

2. Shared credentials file.

3. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.

4. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.

_If there is an MFA serial attached to the credentials, you will be prompted for an MFA token._

#### Shared File Example
If `~/.aws/config` and `~/.aws/config` is setup for the AWS CLI then you may leverage the existing profile configuration for authentication.
```shell
$ export AWS_PROFILE=sb-test
$ glc clean -r us-west-2
INFO[03/05/21] Scanning AWS environment in us-west-2
INFO[03/05/21] ............
```
Alternatively, the `--profile` flag may be used.
```shell
$ glc clean -r us-west-2 -p myProfile
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

## Proxy
The tool supports network proxy configurations and will honor the following proxy environment variables.

* `HTTP_PROXY`
* `HTTPS_PROXY`
* `NO_PROXY`

The environment values may be either a complete URL or a "host[:port]", in which case the "http" scheme is assumed. An error is returned if the value is a different form.

```shell
$ export HTTP_PROXY=http://proxy.example.org:9000

$ glc clean -r us-west-2
2021/03/04 20:42:46 Scanning AWS environment in us-west-2.....
2021/03/04 20:42:46 ............
```

## GitHub Actions Cron

go-lambda-cleanup is a good fit for cron jobs. Below is an example snippet for how you can setup a cron job through GitHub Actions.

```yml
name: Nightly Lambda Version Cleanup

on:
  schedule:
    # At 04:00 on every day
    - cron: '0 04 * * *'
env:
  VERSION: 'v2.0.15'

jobs:
  build:
    name: Run go-lambda-cleanup
    runs-on: ubuntu-latest
    steps:

      - name: Pull Docker Image
        run: docker pull ghcr.io/karl-cardenas-coding/go-lambda-cleanup:$VERSION


      - name: Run go-lambda-cleanup in Test
        env:
          AWS_ACCESS_KEY_ID: ${{secrets.AWS_TEST_ACCESS_KEY}}
          AWS_SECRET_ACCESS_KEY: ${{secrets.AWS_TEST_SECRET_ACCESS_KEY}}
          REGION: us-east-1
        run: docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY ghcr.io/karl-cardenas-coding/go-lambda-cleanup:$VERSION clean -r $REGION

      - name: Run go-lambda-cleanup in Prod
        env:
          AWS_ACCESS_KEY_ID: ${{secrets.AWS_PROD_ACCESS_KEY}}
          AWS_SECRET_ACCESS_KEY: ${{secrets.AWS_PROD_SECRET_ACCESS_KEY}}
          REGION: us-east-1
        run: docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY ghcr.io/karl-cardenas-coding/go-lambda-cleanup:$VERSION clean -r $REGION
```

## Contributing to go-lambda-cleanup

For a complete guide to contributing to go-lambda-clean, please review the [Contribution Guide](documentation/CONTRIBUTING.md).

Contributions to go-lambda-cleanup of any kind are welcome. Contributions include, but not limited to; documentation, organization, tutorials, blog posts, bug reports, issues, feature requests, feature implementations, pull requests, answering questions on the forum, helping to manage issues, etc.

## FAQ

---
<table><tr>

Q: On MacOS I am unable to open the binary due to Apple not trusting the binary. What are my options?

A: You have four options. 

- Option A (Recommended) is to use the Docker container. Please review the [Docker steps](#docker).

- Option B is to to grant permission for the application to run. Use [this guide](https://support.apple.com/en-us/HT202491) to help you grant permission to the application.

- Option C is not recommended but it's an avaiable option. You can remove the binary from quarantine mode. 
  ```shell
  xattr -d com.apple.quarantine /path/to/file
  ```
- Option D is to clone this project and compile the binary. Issue `go build -o glc`, and the end result is a binary compatible for your system. If you still encounter issues after this, invoke the code signing command on the binary `codesign -s -`
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

# Open Source Licenses

For a list of all open source packages and software used, check out the open source [acknowledgment](./documentation/open-source.md) resource page.
