// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"regexp"
	"sort"
	"testing"
	"time"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

var (
	//go:embed aws-regions.txt
	rr embed.FS
)

func TestGetLambdaStorage(t *testing.T) {

	var (
		lambdaList []types.FunctionConfiguration
		want       int64
	)

	lambdaList = []types.FunctionConfiguration{
		{
			CodeSha256:       new(string),
			CodeSize:         1200,
			DeadLetterConfig: &types.DeadLetterConfig{},
			Description:      aws.String("Test A"),
		},
		{
			CodeSha256: new(string),
			CodeSize:   1500,
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
	got, err := validateRegion(rr, input)
	if err != nil || got != want {
		t.Fatalf("The provided input is valid, %s is a valid region", input)
	}

}

func TestValidateRegionWithEnv(t *testing.T) {

	os.Setenv("AWS_DEFAULT_REGION", "not-valid")
	expectedErr := "not-valid is an invalid AWS region. If this is an error please report it"
	err := CleanCmd.RunE(CleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	os.Setenv("AWS_DEFAULT_REGION", "")
	expectedErr = "missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region"
	err = CleanCmd.RunE(CleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	// Set rootCmd to use the region flag
	RegionFlag = "not-valid"
	expectedErr = "not-valid is an invalid AWS region. If this is an error please report it"
	err = CleanCmd.RunE(CleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

}

func TestValidateRegionWithFlag(t *testing.T) {

	expectedErr := "missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region"
	err := CleanCmd.RunE(CleanCmd, []string{"--profile", "default", "--region", "not-valid", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}
}

func TestInvalidRegion(t *testing.T) {
	input := "not-valid"
	want := "not-valid is an invalid AWS region. If this is an error please report it"
	got, err := validateRegion(rr, input)

	if err == nil || err.Error() != want {
		t.Fatalf("The provided input is invalid, %s is not a valid region", got)
	}
}

func TestGetLambdasToDeleteList(t *testing.T) {
	var (
		retainNumber int8 = 2
		lambdaList   []types.FunctionConfiguration
		want         int = 3
	)

	lambdaList = []types.FunctionConfiguration{
		{

			CodeSha256:  new(string),
			Version:     aws.String("1"),
			CodeSize:    1200,
			Description: aws.String("Test A"),
		},
		{

			CodeSha256: new(string),
			Version:    aws.String("2"),
			CodeSize:   1500,
		},
		{

			CodeSha256: new(string),
			Version:    aws.String("3"),
			CodeSize:   1500,
		},
		{

			CodeSha256: new(string),
			Version:    aws.String("4"),
			CodeSize:   1500,
		},
		{

			CodeSha256: new(string),
			Version:    aws.String("5"),
			CodeSize:   1500,
		},
	}

	sort.Sort(byVersion(lambdaList))

	got := getLambdasToDeleteList(lambdaList, retainNumber)

	if len(got) != want {
		t.Fatalf("Expected %d lambda configuration items to be returned but instead received %d", want, len(got))
	}

}

func TestGenerateDeleteInputStructs(t *testing.T) {

	lambdaList := [][]types.FunctionConfiguration{
		{
			types.FunctionConfiguration{
				CodeSha256:       new(string),
				FunctionName:     aws.String("A"),
				Version:          aws.String("1"),
				CodeSize:         1200,
				DeadLetterConfig: &types.DeadLetterConfig{},
				Description:      aws.String("Test A"),
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("B"),
				Version:      aws.String("2"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("C"),
				Version:      aws.String("3"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("D"),
				Version:      aws.String("4"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("E"),
				Version:      aws.String("5"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("F"),
				Version:      aws.String("$LATEST"),
				CodeSize:     1500,
			},
		},
		{
			types.FunctionConfiguration{
				CodeSha256:       new(string),
				Version:          aws.String("1"),
				FunctionName:     aws.String("A1"),
				CodeSize:         1200,
				DeadLetterConfig: &types.DeadLetterConfig{},
				Description:      aws.String("Test A"),
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A2"),
				Version:      aws.String("2"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A3"),
				Version:      aws.String("3"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A4"),
				Version:      aws.String("$LATEST"),
				CodeSize:     1500,
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

	lambdaList := [][]types.FunctionConfiguration{
		{
			types.FunctionConfiguration{
				CodeSha256:       new(string),
				FunctionName:     aws.String("A"),
				Version:          aws.String("1"),
				CodeSize:         1200,
				DeadLetterConfig: &types.DeadLetterConfig{},
				Description:      aws.String("Test A"),
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("B"),
				Version:      aws.String("2"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("C"),
				Version:      aws.String("3"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("D"),
				Version:      aws.String("4"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("E"),
				Version:      aws.String("5"),
				CodeSize:     1500,
			},
		},
		{
			types.FunctionConfiguration{
				CodeSha256:       new(string),
				Version:          aws.String("1"),
				FunctionName:     aws.String("A1"),
				CodeSize:         1200,
				DeadLetterConfig: &types.DeadLetterConfig{},
				Description:      aws.String("Test A"),
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A2"),
				Version:      aws.String("2"),
				CodeSize:     1500,
			},
			types.FunctionConfiguration{
				CodeSha256:   new(string),
				FunctionName: aws.String("A3"),
				Version:      aws.String("3"),
				CodeSize:     1500,
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
	got := calculateFileSize(308000000, &cliConfig)

	if got != want {
		t.Fatalf("Expected the size output to be %s but received %s instead", want, got)
	}

	cliConfig.SizeIEC = aws.Bool(false)

	want2 := "308 MB"
	got2 := calculateFileSize(308000000, &cliConfig)
	if got2 != want2 {
		t.Fatalf("Expected the size output to be %s but received %s instead", want2, got2)
	}
}

func TestDisplayDuration(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	startTime := time.Now().Add(-time.Second * 30)

	displayDuration(startTime)

	got := buf.String()
	want := "time=.* level=.* msg=\"Job Duration Time: 30.00"
	if match, _ := regexp.MatchString(want, got); !match {
		t.Errorf("displayDuration() = %q, want %q", got, want)
	}
	buf.Reset()
}

func TestEnteryMissingEnvRegion(t *testing.T) {

	expectedErr := "Missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region"

	err := CleanCmd.RunE(CleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil && err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

}

func TestDeleteLambdaVersionError(t *testing.T) {

	ctx := context.Background()
	networkName := "localstack-network-v2"

	localstackContainer, err := localstack.RunContainer(ctx,
		localstack.WithNetwork(networkName, "localstack"),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:latest",
				Env:   map[string]string{"SERVICES": "lambda"},
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	lambdaClient, err := lambdaClient(ctx, localstackContainer)
	if err != nil {
		t.Fatal(err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(true),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(2),
	}

	deleteList := []lambda.DeleteFunctionInput{
		{
			FunctionName: aws.String("test"),
			Qualifier:    aws.String("1"),
		},
	}

	err = deleteLambdaVersion(ctx, lambdaClient, deleteList)
	if err == nil {
		t.Errorf("expected an error to be returned but received %v", err)
	}

}

// func TestDeleteLambdaVersion(t *testing.T) {

// 	ctx := context.Background()
// 	networkName := "localstack-network-v2"

// 	localstackContainer, err := localstack.RunContainer(ctx,
// 		localstack.WithNetwork(networkName, "localstack"),
// 		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
// 			ContainerRequest: testcontainers.ContainerRequest{
// 				Image: "localstack/localstack:latest",
// 				Env:   map[string]string{"SERVICES": "lambda"},
// 			},
// 		}),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Clean up the container
// 	defer func() {
// 		if err := localstackContainer.Terminate(ctx); err != nil {
// 			panic(err)
// 		}
// 	}()

// 	lambdaClient, err := lambdaClient(ctx, localstackContainer)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	GlobalCliConfig = cliConfig{
// 		RegionFlag:        aws.String("us-east-1"),
// 		ProfileFlag:       aws.String(""),
// 		DryRun:            aws.Bool(true),
// 		Verbose:           aws.Bool(true),
// 		LambdaListFile:    aws.String(""),
// 		MoreLambdaDetails: aws.Bool(true),
// 		SizeIEC:           aws.Bool(false),
// 		SkipAliases:       aws.Bool(false),
// 		Retain:            aws.Int8(2),
// 	}

// 	// verify file existis
// 	zipFile := "../tests/archive.zip"
// 	if _, err := os.Stat(zipFile); os.IsNotExist(err) {
// 		t.Fatalf("zip file does not exist, %v", err)
// 	}

// 	zipContent, err := decodeZipFile(zipFile)
// 	if err != nil {
// 		t.Fatalf("failed to decode zip file, %v", err)
// 	}

// 	fnCode := []byte(zipContent)

// 	deleteList := []lambda.DeleteFunctionInput{
// 		{
// 			FunctionName: aws.String("test"),
// 			Qualifier:    aws.String("1"),
// 		},
// 	}

// 	err = deleteLambdaVersion(ctx, lambdaClient, deleteList)
// 	if err != nil {
// 		t.Errorf("expected no error to be returned but received %v", err)
// 	}

// }

// lambdaClient returns a lambda client configured to use the localstack containers
func lambdaClient(ctx context.Context, l *localstack.LocalStackContainer) (*lambda.Client, error) {
	mappedPort, err := l.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		return nil, err
	}

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return nil, err
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           fmt.Sprintf("http://%s:%d", host, mappedPort.Int()),
				SigningRegion: region,
			}, nil
		})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	client := lambda.NewFromConfig(awsCfg, func(o *lambda.Options) {})

	return client, nil
}

// // decodeZipFile returns a slice of base64 encoded strings from the provided zip file
// func decodeZipFile(zipFilePath string) (string, error) {
// 	// Open the zip file
// 	r, err := zip.OpenReader(zipFilePath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer r.Close()

// 	var base64Content string

// 	// Iterate over the files in the zip file
// 	for _, file := range r.File {
// 		// Get the file's content
// 		content, err := file.Open()
// 		if err != nil {
// 			return "", err
// 		}
// 		defer content.Close()

// 		// Read the file's content
// 		fileContent, err := io.ReadAll(content)
// 		if err != nil {
// 			return "", err
// 		}

// 		base64Content = base64.StdEncoding.EncodeToString(fileContent)
// 	}

// 	return base64Content, nil
// }

// func createLambdaFunction(scope *awscdk.Construct, id string, handler string, runtime awscdk.LambdaRuntime) awslambda.Function {
// 	// Create a new lambda function
// 	function := awslambda.NewFunction(scope, id, handler, runtime)

// 	// Add a trigger to the lambda function
// 	function.AddEventSource(awslambda.NewEventSource(
// 		awslambda.EventSourceType.S3,
// 		awslambda.S3Config{
// 			bucket: aws.String("my-bucket"),
// 			events: []string{
// 				"s3:ObjectCreated:*",
// 			},
// 		},
// 	))

// 	// Return the lambda function
// 	return function
// }
