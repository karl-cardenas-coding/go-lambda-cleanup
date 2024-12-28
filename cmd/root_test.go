// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestRoot(t *testing.T) {

	err := rootCmd.RunE(rootCmd, []string{})
	if err != nil {
		t.Errorf("expected no error to be returned but received %v", err)
	}

}

func TestRootCMD(t *testing.T) {
	ctx := context.TODO()
	newNetwork, err := network.New(ctx)
	if err != nil {
		t.Errorf("failed to create network: %s", err)
	}
	localstackContainer, err := localstack.Run(ctx,
		"localstack/localstack:3.6",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "lambda"}),
		network.WithNetwork([]string{"localstack-network-v2"}, newNetwork),
	)
	if err != nil {
		t.Errorf("failed to start localstack container: %s", err)
	}

	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			t.Error(err)
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
	newNetwork, err := network.New(ctx)
	if err != nil {
		t.Errorf("failed to create network: %s", err)
	}
	localstackContainer, err := localstack.Run(ctx,
		"localstack/localstack:3.6",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "lambda"}),
		network.WithNetwork([]string{"localstack-network-v2"}, newNetwork),
	)
	if err != nil {
		t.Errorf("failed to start localstack container: %s", err)
	}

	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			t.Error(err)
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
	newNetwork, err := network.New(ctx)
	if err != nil {
		t.Errorf("failed to create network: %s", err)
	}
	localstackContainer, err := localstack.Run(ctx,
		"localstack/localstack:3.6",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "lambda"}),
		network.WithNetwork([]string{"localstack-network-v2"}, newNetwork),
	)
	if err != nil {
		t.Errorf("failed to start localstack container: %s", err)
	}

	defer func() {
		if err := localstackContainer.Terminate(ctx); err != nil {
			t.Error(err)
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

func TestCreateHTTPClient(t *testing.T) {

	client := createHTTPClient()

	// Check if the client is not nil
	if client == nil {
		t.Fatalf("Expected non-nil HTTP client")
	}

	// Check if the transport is of type *http.Transport
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Expected transport to be of type *http.Transport, got %T", client.Transport)
	}

	// Check if the minimum TLS version is set to TLS 1.2
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion to be TLS 1.2, got %v", transport.TLSClientConfig.MinVersion)
	}

	// Check if ForceAttemptHTTP2 is set to true
	if !transport.ForceAttemptHTTP2 {
		t.Errorf("Expected ForceAttemptHTTP2 to be true, got %v", transport.ForceAttemptHTTP2)
	}
}
