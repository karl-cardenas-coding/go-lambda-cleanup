# Copyright (c) karl-cardenas-coding
# SPDX-License-Identifier: MIT

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region  = "us-east-1"
  profile = var.profile
}

resource "aws_lambda_function" "test_lambda" {
  count         = 55
  filename      = "lambda.zip"
  function_name = "blank_go-${count.index}"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "main"
  timeout       = 30
  memory_size   = 128
  runtime       = "go1.x"
  tags = {
      Name = "blank_go-${count.index}"
  }

  depends_on = [
    data.archive_file.zip
  ]
}

variable "profile" {
  type = string
  description = "AWS Profile variable"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "archive_file" "zip" {
  type        = "zip"
  source_file = "${path.module}/main"
  output_path = "${path.module}/lambda.zip"
}