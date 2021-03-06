# Development Environment

## Requirements
* node v14.15.0+
* [Go](https://golang.org/doc/install) 1.16+

## Setup
If you wish to work on the go-lambda-cleanup CLI, you'll first need Go installed on your machine (please check the requirements before proceeding).

Note: This project uses Go Modules making it safe to work with it outside of your existing GOPATH. The instructions that follow assume a directory in your home directory outside of the standard GOPATH.

Clone repository to: `$HOME/projects/go-lambda-cleanup/`
```
$ mkdir -p $HOME/projects/go-lambda-cleanup/; cd $HOME/projects/go-lambda-cleanup/
$ git clone git@github.com:karl-cardenas-coding/go-lambda-cleanup.git
```

Issue the command `npm install`. This will setup the commit hooks for the project through [Husky](https://github.com/typicode/husky) v4
```
$ npm install
```

The `clean` command expects the file `cmd/aws-regions.txt` to be present. Otherwise binary builds or `go run main.go clean` will fail. Issue the command below from the root of the project namespace. This assumes you have your local aws credentials configured.
```shell
 aws ec2 describe-regions --region us-east-1 --query "Regions[].RegionName" --output text >> cmd/aws-regions.txt
```

## Testing
