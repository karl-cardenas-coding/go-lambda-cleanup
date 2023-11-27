// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"testing"
	"time"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
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

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})

}

func TestDeleteLambdaVersion(t *testing.T) {

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

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(true),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	deleteList := []lambda.DeleteFunctionInput{
		{
			FunctionName: aws.String("func1"),
			Qualifier:    aws.String("2"),
		},
	}

	err = deleteLambdaVersion(ctx, svc, deleteList)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	result, err := listFunctionVersions(ctx, svc, "func1")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}
	if result != 2 {
		t.Errorf("expected 2 functions to be returned but received %v", result)
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})

}
func TestGetAllLambdas(t *testing.T) {

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

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(true),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	lambdaListResult, err := getAllLambdas(ctx, svc, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	result, err := listFunctions(ctx, svc)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	if len(lambdaListResult) != result {
		t.Errorf("expected 3 functions to be returned but received %v", result)
	}

	lambdaListResult2, err := getAllLambdas(ctx, svc, []string{"func1"})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	if len(lambdaListResult2) != 1 {
		t.Errorf("Scenario 2: expected 1 functions to be returned but received %v", len(lambdaListResult2))
	}

	_, err = getAllLambdas(ctx, svc, []string{"func22"})
	if err != nil {
		t.Errorf("expected an error to be returned but received %v", err)
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})

}

func TestGetAllLambdaVersion(t *testing.T) {

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

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(true),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	versions, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("expected 3 versions to be returned but received %v", len(versions))
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(true),
			Retain:            aws.Int8(0),
		}

	})

}
func TestGetAllLambdaVersionWithAlias(t *testing.T) {

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

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(true),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(true),
		Retain:            aws.Int8(0),
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = publishAlias(ctx, svc, "func1", "DEMO", "2")
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	versions, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("expected 2 versions to be returned but received %v", len(versions))
	}

	_, err = getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func22"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func22"),
	}, GlobalCliConfig)
	if err == nil {
		t.Errorf("expected an error to be returned but received %v", err)
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})

}

func TestExecuteClean(t *testing.T) {
	ctx := context.TODO()
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

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String(""),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(true),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = publishAlias(ctx, svc, "func1", "DEMO", "2")
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	err = executeClean(ctx, &GlobalCliConfig, svc, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}
}

func TestCleanCMDDryRun(t *testing.T) {

	ctx := context.TODO()
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

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		panic(err)
	}

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		panic(err)
	}

	mappedPort, err := localstackContainer.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		panic(err)
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	bf3, err := getZipPackage("../tests/handler3.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf3)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(false),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	versions, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	if len(versions) != 4 {
		t.Errorf("expected 4 versions to be returned but received %v", len(versions))
	}

	t.Logf("Pre-Clean # of versions: %v", len(versions))

	os.Setenv("AWS_ENDPOINT_URL", fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	err = cleanCmd.RunE(cleanCmd, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	actual, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	t.Logf("Post-Clean # of versions: %v", len(actual))

	if len(actual) != 4 {
		t.Errorf("expected 4 versions to be returned but received %v", len(actual))
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(2),
		}

		os.Unsetenv("AWS_ENDPOINT_URL")
		os.Unsetenv("AWS_EC2_METADATA_DISABLED")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	})

}

func TestCleanCMD(t *testing.T) {

	ctx := context.TODO()
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

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		panic(err)
	}

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		panic(err)
	}

	mappedPort, err := localstackContainer.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		panic(err)
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	bf3, err := getZipPackage("../tests/handler3.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf3)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(false),
		Verbose:           aws.Bool(false),
		LambdaListFile:    aws.String(""),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	versions, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	if len(versions) != 4 {
		t.Errorf("expected 4 versions to be returned but received %v", len(versions))
	}

	t.Logf("Pre-Clean # of versions: %v", len(versions))

	os.Setenv("AWS_ENDPOINT_URL", fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	err = cleanCmd.RunE(cleanCmd, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	actual, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	t.Logf("Post-Clean # of versions: %v", len(actual))

	if len(actual) != 2 {
		t.Errorf("expected 2 versions to be returned but received %v", len(actual))
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(2),
		}

		os.Unsetenv("AWS_ENDPOINT_URL")
		os.Unsetenv("AWS_EC2_METADATA_DISABLED")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	})

}

func TestCleanCMDWithCustomList(t *testing.T) {

	ctx := context.TODO()
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

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		panic(err)
	}

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		panic(err)
	}

	mappedPort, err := localstackContainer.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		panic(err)
	}

	bf, err := getZipPackage("../tests/handler.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	svc, err := getAWSCredentials(ctx, localstackContainer)
	if err != nil {
		panic(err)
	}

	_, err = addFunctions(ctx, svc, bf)
	if err != nil {
		panic(err)
	}

	bf2, err := getZipPackage("../tests/handler2.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf2)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	bf3, err := getZipPackage("../tests/handler3.zip")
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	_, err = updateFunctions(ctx, svc, *bf3)
	if err != nil {
		t.Logf("expected no error to be returned but received %v", err)
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(true),
		Verbose:           aws.Bool(false),
		LambdaListFile:    aws.String("../tests/custom.yml"),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	versions, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	if len(versions) != 4 {
		t.Errorf("expected 4 versions to be returned but received %v", len(versions))
	}

	t.Logf("Pre-Clean # of versions: %v", len(versions))

	os.Setenv("AWS_ENDPOINT_URL", fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	err = cleanCmd.RunE(cleanCmd, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	actual, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func3"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	t.Logf("Post-Clean # of versions: %v", len(actual))

	if len(actual) != 4 {
		t.Errorf("expected 4 versions to be returned but received %v", len(actual))
	}

	GlobalCliConfig = cliConfig{
		RegionFlag:        aws.String("us-east-1"),
		CredentialsFile:   aws.Bool(false),
		ProfileFlag:       aws.String(""),
		DryRun:            aws.Bool(false),
		Verbose:           aws.Bool(false),
		LambdaListFile:    aws.String("../tests/custom.yml"),
		MoreLambdaDetails: aws.Bool(true),
		SizeIEC:           aws.Bool(false),
		SkipAliases:       aws.Bool(false),
		Retain:            aws.Int8(0),
	}

	err = cleanCmd.RunE(cleanCmd, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	actual2, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func3"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	t.Logf("Post-Clean Non-Dry # of versions: %v", len(actual))

	if len(actual2) != 2 {
		t.Errorf("expected 2 versions to be returned but received %v", len(actual2))
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(2),
		}

		os.Unsetenv("AWS_ENDPOINT_URL")
		os.Unsetenv("AWS_EC2_METADATA_DISABLED")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	})

}

func TestAWSEnteryMissingEnvRegion(t *testing.T) {

	expectedErr := "Missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region"

	err := cleanCmd.RunE(cleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil && err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(2),
		}

	})

}

func TestAWSValidateRegion(t *testing.T) {

	input := "us-east-1"
	want := "us-east-1"
	got, err := validateRegion(rr, input)
	if err != nil || got != want {
		t.Fatalf("The provided input is valid, %s is a valid region", input)
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})

}

func TestValidateRegionWithEnv(t *testing.T) {

	os.Setenv("AWS_DEFAULT_REGION", "not-valid")
	expectedErr := "not-valid is an invalid AWS region. If this is an error please report it"
	err := cleanCmd.RunE(cleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	os.Setenv("AWS_DEFAULT_REGION", "")
	expectedErr = "missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region"
	err = cleanCmd.RunE(cleanCmd, []string{"--profile", "default", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	// Set rootCmd to use the region flag
	GlobalCliConfig.RegionFlag = aws.String("not-valid")
	expectedErr = "not-valid is an invalid AWS region. If this is an error please report it"
	err = cleanCmd.RunE(cleanCmd, []string{"--profile", "default", "--region", "not-valid", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(false),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})

}

func TestAWSValidateRegionWithFlag(t *testing.T) {

	expectedErr := "missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region"
	err := cleanCmd.RunE(cleanCmd, []string{"--profile", "default", "--region", "not-valid", "--retain", "2", "--dry-run"})
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected an error to be returned but received %v", err.Error())
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})
}

func TestAWSInvalidRegion(t *testing.T) {

	input := "not-valid"
	want := "not-valid is an invalid AWS region. If this is an error please report it"
	got, err := validateRegion(rr, input)

	if err == nil || err.Error() != want {
		t.Fatalf("The provided input is invalid, %s is not a valid region", got)
	}

	t.Cleanup(func() {
		GlobalCliConfig = cliConfig{
			RegionFlag:        aws.String(""),
			CredentialsFile:   aws.Bool(false),
			ProfileFlag:       aws.String(""),
			DryRun:            aws.Bool(true),
			Verbose:           aws.Bool(true),
			LambdaListFile:    aws.String(""),
			MoreLambdaDetails: aws.Bool(true),
			SizeIEC:           aws.Bool(false),
			SkipAliases:       aws.Bool(false),
			Retain:            aws.Int8(0),
		}

	})
}

/*

THE CODE BELOW IS FOR TESTING PURPOSES ONLY


*/

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
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("aaaa", "bbbb", "cccc")),
	)
	if err != nil {
		return nil, err
	}

	client := lambda.NewFromConfig(awsCfg, func(o *lambda.Options) {})

	return client, nil
}

func addFunctions(ctx context.Context, svc *lambda.Client, zipPackage *bytes.Buffer) (string, error) {

	list := []lambda.CreateFunctionInput{
		{
			Code: &types.FunctionCode{
				ZipFile: zipPackage.Bytes(),
			},
			Description:  aws.String("func1"),
			FunctionName: aws.String("func1"),
			Handler:      aws.String("index.handler"),
			Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
			Runtime:      types.RuntimeNodejs18x,
			Publish:      true,
		},
		{
			Code: &types.FunctionCode{
				ZipFile: zipPackage.Bytes(),
			},
			Description:  aws.String("func2"),
			FunctionName: aws.String("func2"),
			Handler:      aws.String("index.handler"),
			Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
			Runtime:      types.RuntimeNodejs18x,
			Publish:      true,
		},
		{
			Code: &types.FunctionCode{
				ZipFile: zipPackage.Bytes(),
			},
			Description:  aws.String("func3"),
			FunctionName: aws.String("func3"),
			Handler:      aws.String("index.handler"),
			Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
			Runtime:      types.RuntimeNodejs18x,
			Publish:      true,
		},
	}

	var result string

	for _, input := range list {
		var state types.State
		item := input
		output, err := svc.CreateFunction(ctx, &item)
		if err != nil {
			var resConflict *types.ResourceConflictException
			if errors.As(err, &resConflict) {
				log.Printf("Function %v already exists.\n", "test")
				state = types.StateActive
			} else {
				log.Panicf("Couldn't create function %v. Here's why: %v\n", "test", err)
			}
		} else {
			waiter := lambda.NewFunctionActiveV2Waiter(svc)
			funcOutput, err := waiter.WaitForOutput(context.TODO(), &lambda.GetFunctionInput{
				FunctionName: aws.String(*input.FunctionName)}, 2*time.Minute)
			if err != nil {
				log.Panicf("Couldn't wait for function %v to be active. Here's why: %v\n", "test", err)
			} else {
				state = funcOutput.Configuration.State
			}
		}

		fmt.Println("Function ARN: ", *output.FunctionArn)

		result += fmt.Sprintf("Function %v is %v\n", *input.FunctionName, state)
	}

	return fmt.Sprintf("Result: %v", result), nil
}

func listFunctions(ctx context.Context, svc *lambda.Client) (int, error) {

	output, err := svc.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		return 0, err
	}

	return len(output.Functions), err

}

func getAWSCredentials(ctx context.Context, l *localstack.LocalStackContainer) (*lambda.Client, error) {

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := l.MappedPort(ctx, nat.Port("4566/tcp"))
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
		config.WithSharedConfigFiles([]string{"/tests/config"}),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("aaaa", "bbbb", "cccc")),
	)
	if err != nil {
		return nil, err
	}

	return lambda.NewFromConfig(awsCfg), nil

}

func updateFunctions(ctx context.Context, svc *lambda.Client, zipPackage bytes.Buffer) (string, error) {

	list := []lambda.UpdateFunctionCodeInput{
		{
			FunctionName: aws.String("func1"),
			ZipFile:      zipPackage.Bytes(),
			Publish:      true,
		},
		{
			FunctionName: aws.String("func2"),
			ZipFile:      zipPackage.Bytes(),
			Publish:      true,
		},
		{
			FunctionName: aws.String("func3"),
			ZipFile:      zipPackage.Bytes(),
			Publish:      true,
		},
	}

	var result string

	for _, input := range list {
		var state types.State
		item := input
		_, err := svc.UpdateFunctionCode(ctx, &item)
		if err != nil {
			var resConflict *types.ResourceConflictException
			if errors.As(err, &resConflict) {
				log.Printf("Function %v already exists.\n", "test")
				state = types.StateActive
			} else {
				log.Panicf("Couldn't create function %v. Here's why: %v\n", "test", err)
			}
		} else {
			waiter := lambda.NewFunctionActiveV2Waiter(svc)
			funcOutput, err := waiter.WaitForOutput(context.TODO(), &lambda.GetFunctionInput{
				FunctionName: aws.String(*input.FunctionName)}, 2*time.Minute)
			if err != nil {
				log.Panicf("Couldn't wait for function %v to be active. Here's why: %v\n", "test", err)
			} else {
				state = funcOutput.Configuration.State
			}
		}

		result += fmt.Sprintf("Function %v was updated and the state is  %v\n", *input.FunctionName, state)
	}

	return fmt.Sprintf("Result: %v", result), nil
}

func publishAlias(ctx context.Context, svc *lambda.Client, functionName string, aliasName string, version string) (string, error) {

	input := lambda.CreateAliasInput{
		FunctionName:    aws.String(functionName),
		FunctionVersion: aws.String(version),
		Name:            aws.String(aliasName),
		Description:     aws.String("Alias Test"),
	}

	_, err := svc.CreateAlias(ctx, &input)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return fmt.Sprintf("Alias %v was created for function %v", aliasName, functionName), nil
}

func getZipPackage(zipFile string) (*bytes.Buffer, error) {
	_, err := os.Stat(zipFile)
	if err != nil {
		return nil, err
	}

	zipPackage, err := os.Open(zipFile)
	if err != nil {
		return nil, err
	}

	defer zipPackage.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(zipPackage)
	if err != nil {
		return nil, err
	}

	return buf, err

}

func listFunctionVersions(ctx context.Context, svc *lambda.Client, funcName string) (int, error) {

	output, err := svc.ListVersionsByFunction(ctx, &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(funcName),
	})

	if err != nil {
		return 0, err
	}

	return len(output.Versions), nil
}
