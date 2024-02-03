// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

func TestRoot(t *testing.T) {

	err := rootCmd.RunE(rootCmd, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

}

func TestRootCMD(t *testing.T) {
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

	if len(versions) != 3 {
		t.Errorf("expected 3 versions to be returned but received %v", len(versions))
	}

	t.Logf("Pre-Clean # of versions: %v", len(versions))

	os.Setenv("AWS_ENDPOINT_URL", fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	err = rootCmd.RunE(rootCmd, []string{})
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

	if len(actual) != 3 {
		t.Errorf("expected 3 versions to be returned but received %v", len(actual))
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

func TestRootExecute(t *testing.T) {
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

	if len(versions) != 3 {
		t.Errorf("expected 3 versions to be returned but received %v", len(versions))
	}

	t.Logf("Pre-Clean # of versions: %v", len(versions))

	os.Setenv("AWS_ENDPOINT_URL", fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	Execute()

	actual, err := getAllLambdaVersion(ctx, svc, types.FunctionConfiguration{
		FunctionName: aws.String("func1"),
		FunctionArn:  aws.String("arn:aws:lambda:us-east-1:000000000000:function:func1"),
	}, GlobalCliConfig)
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

	t.Logf("Post-Clean # of versions: %v", len(actual))

	if len(actual) != 3 {
		t.Errorf("expected 3 versions to be returned but received %v", len(actual))
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

func TestNoLambdas(t *testing.T) {
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

	// svc, err := getAWSCredentials(ctx, localstackContainer)
	// if err != nil {
	// 	panic(err)
	// }

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

	os.Setenv("AWS_ENDPOINT_URL", fmt.Sprintf("http://%s:%d", host, mappedPort.Int()))
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	Execute()

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
