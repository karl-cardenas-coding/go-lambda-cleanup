package cmd

import (
	"embed"
	"testing"

	_ "embed"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var (
	//go:embed aws-regions.txt
	rr embed.FS
)

func TestGetLambdaStorage(t *testing.T) {

	var (
		lambdaList []*lambda.FunctionConfiguration
		want       int64
	)

	lambdaList = []*lambda.FunctionConfiguration{
		&lambda.FunctionConfiguration{
			CodeSha256:       new(string),
			CodeSize:         aws.Int64(1200),
			DeadLetterConfig: &lambda.DeadLetterConfig{},
			Description:      aws.String("Test A"),
		},
		&lambda.FunctionConfiguration{
			CodeSha256: new(string),
			CodeSize:   aws.Int64(1500),
		},
	}

	want = 2700
	got, err := getLambdaStorage(lambdaList)
	if got != want || err != nil {
		t.Fatalf("Lambda storage calculation invalid. Expected %d but received %d", want, got)
	}
}

func TestValidateRegion(t *testing.T) {
	input := "us-east-1"
	want := "us-east-1"
	got := validateRegion(rr, input)

	if got != want {
		t.Fatalf("The provided input is valid, %s is a valid region", input)
	}

}

func TestGetLambdasToDelteList(t *testing.T) {
	var (
		retainNumber int8 = 2
		lambdaList   []*lambda.FunctionConfiguration
		want         int = 2
	)

	lambdaList = []*lambda.FunctionConfiguration{
		&lambda.FunctionConfiguration{
			CodeSha256:       new(string),
			Version:          aws.String("1"),
			CodeSize:         aws.Int64(1200),
			DeadLetterConfig: &lambda.DeadLetterConfig{},
			Description:      aws.String("Test A"),
		},
		&lambda.FunctionConfiguration{
			CodeSha256: new(string),
			Version:    aws.String("2"),
			CodeSize:   aws.Int64(1500),
		},
		&lambda.FunctionConfiguration{
			CodeSha256: new(string),
			Version:    aws.String("3"),
			CodeSize:   aws.Int64(1500),
		},
		&lambda.FunctionConfiguration{
			CodeSha256: new(string),
			Version:    aws.String("4"),
			CodeSize:   aws.Int64(1500),
		},
		&lambda.FunctionConfiguration{
			CodeSha256: new(string),
			Version:    aws.String("5"),
			CodeSize:   aws.Int64(1500),
		},
	}

	got := getLambdasToDelteList(lambdaList, retainNumber)

	if len(got) != want {
		t.Fatalf("Expected 2 lambda configuration items to be returned but instead received %d", len(got))
	}

}

func TestGenerateDeleteInputStructs(t *testing.T) {

	lambdaList := [][]*lambda.FunctionConfiguration{
		[]*lambda.FunctionConfiguration{
			&lambda.FunctionConfiguration{
				CodeSha256:       new(string),
				FunctionName:     aws.String("A"),
				Version:          aws.String("1"),
				CodeSize:         aws.Int64(1200),
				DeadLetterConfig: &lambda.DeadLetterConfig{},
				Description:      aws.String("Test A"),
			},
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("B"),
				Version:      aws.String("2"),
				CodeSize:     aws.Int64(1500),
			},
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("C"),
				Version:      aws.String("3"),
				CodeSize:     aws.Int64(1500),
			},
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("D"),
				Version:      aws.String("4"),
				CodeSize:     aws.Int64(1500),
			},
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("E"),
				Version:      aws.String("5"),
				CodeSize:     aws.Int64(1500),
			},
		},
		[]*lambda.FunctionConfiguration{
			&lambda.FunctionConfiguration{
				CodeSha256:       new(string),
				Version:          aws.String("1"),
				FunctionName:     aws.String("A1"),
				CodeSize:         aws.Int64(1200),
				DeadLetterConfig: &lambda.DeadLetterConfig{},
				Description:      aws.String("Test A"),
			},
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A2"),
				Version:      aws.String("2"),
				CodeSize:     aws.Int64(1500),
			},
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A3"),
				Version:      aws.String("3"),
				CodeSize:     aws.Int64(1500),
			},
		},
	}

	got, err := generateDeleteInputStructs(lambdaList)
	if len(got) != 2 || err != nil {
		t.Fatalf("Expected a lambda delete struct list to be 2 but go a length of %d", len(got))
	}

	if (*got[1][1].FunctionName != "A2") || err != nil {
		t.Fatalf("Expected a lambda delete struct to have item A2 but instead got %v", *got[1][1].FunctionName)
	}

}

// func TestGetAllLambdaVersion(t *testing.T) {
// 	inputList := []lambda.FunctionConfiguration{
// 		CodeSha256: "3rCuTG9h7yez7ZU+U9Zc3wQE3DTcmSUPY+Vr73RvmKg=",
// 		CodeSize: 3078323,
// 		Description: "A lambda function that populates RDS with mock data",
// 		Environment: {
// 		  Variables: {
// 			ENV: "test",
// 			REGION: "us-east-1",
// 		  },
// 		},
// 		FunctionArn: "arn:aws:lambda:us-east-1:731696908010:function:populate_rds:33",
// 		FunctionName: "populate_rds",
// 		Handler: "populate_rds.lambda_handler",
// 		LastModified: "2021-03-01T05:54:45.642+0000",
// 		MemorySize: 128,
// 		PackageType: "Zip",
// 		RevisionId: "21f4e275-7b24-46e9-bf5b-acf9c70bd476",
// 		Role: "arn:aws:iam::731696908010:role/system/LambdaExecution-RDS-write-role",
// 		Runtime: "python3.7",
// 		SigningJobArn: "arn:aws:signer:us-east-1:731696908010:/signing-jobs/5b5544c6-0851-443b-91c6-4e26a73cf51f",
// 		SigningProfileVersionArn: "arn:aws:signer:us-east-1:731696908010:/signing-profiles/SawyerBrink_TF/B1oks0DM6r",
// 		Timeout: 30,
// 		TracingConfig: {
// 		  Mode: "PassThrough",
// 		},
// 		Version: "33",
// 		VpcConfig: {
// 		  SecurityGroupIds: ["sg-9b77c7bf"],
// 		  SubnetIds: [
// 			"subnet-ff5e8ba0",
// 			"subnet-6ed21a08",
// 			"subnet-938254b2",
// 			"subnet-8131a9cc",
// 			"subnet-8f7966b1",
// 			"subnet-d870f2d6"
// 		  ],
// 		  VpcId: "vpc-4bc62236"
// 		}
// 	  }
// }
