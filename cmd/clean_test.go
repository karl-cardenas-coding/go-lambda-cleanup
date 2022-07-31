package cmd

import (
	"embed"
	"sort"
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
		{
			CodeSha256:       new(string),
			CodeSize:         aws.Int64(1200),
			DeadLetterConfig: &lambda.DeadLetterConfig{},
			Description:      aws.String("Test A"),
		},
		{
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

func TestGetLambdasToDeleteList(t *testing.T) {
	var (
		retainNumber int8 = 2
		lambdaList   []*lambda.FunctionConfiguration
		want         int = 3
	)

	lambdaList = []*lambda.FunctionConfiguration{
		{
			CodeSha256:  new(string),
			Version:     aws.String("1"),
			CodeSize:    aws.Int64(1200),
			Description: aws.String("Test A"),
		},
		{
			CodeSha256: new(string),
			Version:    aws.String("2"),
			CodeSize:   aws.Int64(1500),
		},
		{
			CodeSha256: new(string),
			Version:    aws.String("3"),
			CodeSize:   aws.Int64(1500),
		},
		{
			CodeSha256: new(string),
			Version:    aws.String("4"),
			CodeSize:   aws.Int64(1500),
		},
		{
			CodeSha256: new(string),
			Version:    aws.String("5"),
			CodeSize:   aws.Int64(1500),
		},
	}

	sort.Sort(byVersion(lambdaList))

	got := getLambdasToDeleteList(lambdaList, retainNumber)

	if len(got) != want {
		t.Fatalf("Expected %d lambda configuration items to be returned but instead received %d", want, len(got))
	}

}

func TestGenerateDeleteInputStructs(t *testing.T) {

	lambdaList := [][]*lambda.FunctionConfiguration{
		{
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
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("F"),
				Version:      aws.String("$LATEST"),
				CodeSize:     aws.Int64(1500),
			},
		},
		{
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
			&lambda.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A4"),
				Version:      aws.String("$LATEST"),
				CodeSize:     aws.Int64(1500),
			},
		},
	}

	got, err := generateDeleteInputStructs(lambdaList, false)
	if len(got) != 2 || err != nil {
		t.Fatalf("Expected a lambda delete struct list to be 2 but go a length of %d", len(got))
	}

	if (*got[1][1].FunctionName != "A2") || err != nil {
		t.Fatalf("Expected a lambda delete struct to have item A2 but instead got %v", *got[1][1].FunctionName)
	}

}

func TestCountDeleteVersions(t *testing.T) {

	lambdaList := [][]lambda.DeleteFunctionInput{
		{
			{

				FunctionName: aws.String("A"),
				Qualifier:    aws.String("1"),
			},
			{
				FunctionName: aws.String("B"),
				Qualifier:    aws.String("2"),
			},
			{
				FunctionName: aws.String("C"),
				Qualifier:    aws.String("3"),
			},
			{
				FunctionName: aws.String("D"),
				Qualifier:    aws.String("4"),
			},
			{
				FunctionName: aws.String("E"),
				Qualifier:    aws.String("5"),
			},
		},
		{
			lambda.DeleteFunctionInput{
				FunctionName: aws.String("A1"),
				Qualifier:    aws.String("1"),
			},
			lambda.DeleteFunctionInput{

				FunctionName: aws.String("A2"),
				Qualifier:    aws.String("2"),
			},
			lambda.DeleteFunctionInput{
				FunctionName: aws.String("A3"),
				Qualifier:    aws.String("3"),
			},
		},
	}

	got := countDeleteVersions(lambdaList)

	want := 8

	if got != want {
		t.Fatalf("Expected count of versions to be %d but received %d instead", want, got)
	}

}

func TestCalculateSpaceRemoval(t *testing.T) {

	lambdaList := [][]*lambda.FunctionConfiguration{
		{
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
		{
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

	got := calculateSpaceRemoval(lambdaList)

	want := 11400

	if got != want {
		t.Fatalf("Expected the size of all versions to be %d but received %d instead", want, got)
	}

}

func TestCalculateFileSize(t *testing.T) {

	cliConfig := cliConfig{
		SizeIEC: aws.Bool(true),
	}

	want := "294 MiB"
	got := calculateFileSize(308000000, cliConfig)

	if got != want {
		t.Fatalf("Expected the size output to be %s but received %s instead", want, got)
	}

	cliConfig.SizeIEC = aws.Bool(false)

	want2 := "308 MB"
	got2 := calculateFileSize(308000000, cliConfig)
	if got2 != want2 {
		t.Fatalf("Expected the size output to be %s but received %s instead", want2, got2)
	}
}
