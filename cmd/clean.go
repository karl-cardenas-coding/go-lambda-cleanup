// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/dustin/go-humanize"
	internal "github.com/karl-cardenas-coding/go-lambda-cleanup/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// Per AWS API Valid Range: Minimum value of 1. Maximum value of 10000.
	maxItems   int32  = 10000
	regionFile string = "aws-regions.txt"
)

var (
	//go:embed aws-regions.txt
	f embed.FS

	a string
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes all former versions of AWS lambdas except for the $LATEST version",
	Long:  `Removes all former versions of AWS lambdas except for the $LATEST version. The user also has the ability specify n-? version to retain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		var (
			awsEnvRegion      string
			awsEnvProfile     string
			config            cliConfig
			err               error
			customeDeleteList []string
		)

		config = GlobalCliConfig
		awsEnvRegion = os.Getenv("AWS_DEFAULT_REGION")
		awsEnvProfile = os.Getenv("AWS_PROFILE")
		if *config.RegionFlag == "" {
			if awsEnvRegion != "" {
				*config.RegionFlag, err = validateRegion(f, awsEnvRegion)
				if err != nil {
					return err
				}

			} else {
				return errors.New("missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region")
			}
		} else {
			*config.RegionFlag, err = validateRegion(f, *config.RegionFlag)
			if err != nil {
				return err
			}
		}

		// Create a list of AWS Configurations Options
		awsConfigOptions := []func(*awsConfig.LoadOptions) error{
			awsConfig.WithRegion(*config.RegionFlag),
			awsConfig.WithHTTPClient(GlobalHTTPClient),
			awsConfig.WithAssumeRoleCredentialOptions(func(aro *stscreds.AssumeRoleOptions) {
				aro.TokenProvider = stscreds.StdinTokenProvider
			}),
		}
		if *config.ProfileFlag == "" {
			if awsEnvProfile != "" {
				log.Infof("AWS_PROFILE set to \"%s\"", awsEnvProfile)
				config.ProfileFlag = &awsEnvProfile
			}
		} else {
			log.Infof("The AWS Profile flag \"%s\" was passed in", *config.ProfileFlag)
		}
		awsConfigOptions = append(awsConfigOptions, awsConfig.WithSharedConfigProfile(*config.ProfileFlag))

		if *config.Verbose {
			awsConfigOptions = append(awsConfigOptions, awsConfig.WithClientLogMode(aws.LogRetries|aws.LogRequest))
		}

		if *config.DryRun {
			log.Info("******** DRY RUN MODE ENABLED ********")
		}

		config.SkipAliases = &SkipAliases

		if *config.SkipAliases {
			log.Info("Skip Aliases enabled")
		}

		if *config.LambdaListFile != "" {
			log.Info("******** CUSTOM LAMBDA LIST PROVIDED ********")
			list, err := internal.GenerateLambdaDeleteList(*config.LambdaListFile)
			if err != nil {
				log.Infof("an issue occurred while processing %s", *config.LambdaListFile)
				log.Info(err.Error())
			}
			customeDeleteList = list
		}

		cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfigOptions...)
		if err != nil {
			return errors.New("ERROR ESTABLISHING AWS SESSION")
		}

		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return errors.New("ERROR RETRIEVING AWS CREDENTIALS")
		}
		if creds.Expired() {
			return errors.New("AWS CREDENTIALS EXPIRED")
		}

		// svc = lambda.NewFromConfig(cfg)
		initSvc := lambda.NewFromConfig(cfg, func(o *lambda.Options) {
			// Set the User-Agent for all AWS with the Lambda client
			o.APIOptions = append(o.APIOptions, middleware.AddUserAgentKeyValue("go-lambda-cleanup", VersionString))
		})
		err = executeClean(ctx, &config, initSvc, customeDeleteList)
		if err != nil {
			return err
		}

		return err
	},
}

// executeClean is the main function that executes the clean-up process
// It takes a context, a pointer to a cliConfig struct, a pointer to a lambda client, and a list of custom lambdas to delete
// An error is returned if the function fails to execute
func executeClean(ctx context.Context, config *cliConfig, svc *lambda.Client, customList []string) error {
	startTime := time.Now()

	var (
		returnError                error
		globalLambdaStorage        []int64
		updatedGlobalLambdaStorage []int64
		globalLambdaVersionsList   [][]types.FunctionConfiguration
		counter                    int64 = 0
	)

	log.Info("Scanning AWS environment in " + *config.RegionFlag)
	lambdaList, err := getAllLambdas(ctx, svc, customList)
	if err != nil {
		log.Error("ERROR: ", err)
		log.Fatal("ERROR: Failed to retrieve Lambda list.")
	}
	log.Info("............")

	if len(lambdaList) > 0 {
		tempCounter := 0
		for _, lambda := range lambdaList {
			lambdaItem := lambda
			lambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambdaItem, *config)
			if err != nil {
				log.Error("ERROR: ", err)
				log.Fatal("ERROR: Failed to retrieve Lambda version list.")
			}

			globalLambdaVersionsList = append(globalLambdaVersionsList, lambdaVersionsList)

			totalLambdaStorage, err := getLambdaStorage(lambdaVersionsList)
			if err != nil {
				log.Error("ERROR: ", err)
				log.Fatal("ERROR: Failed to retrieve Lambda storage.")
			}

			globalLambdaStorage = append(globalLambdaStorage, totalLambdaStorage)
			tempCounter++
		}

		log.Info(tempCounter, " Lambdas identified")
		for _, v := range globalLambdaStorage {
			counter = counter + v
		}

		log.Info("Current storage size: ", calculateFileSize(uint64(counter), config))
		log.Info("**************************")
		log.Info("Initiating clean-up process. This may take a few minutes....")
		// Begin delete process
		globalLambdaDeleteList := [][]types.FunctionConfiguration{}

		for _, lambda := range globalLambdaVersionsList {
			lambdasDeleteList := getLambdasToDeleteList(lambda, *config.Retain)
			globalLambdaDeleteList = append(globalLambdaDeleteList, lambdasDeleteList)
		}

		log.Info("............")
		globalLambdaDeleteInputStructs, err := generateDeleteInputStructs(globalLambdaDeleteList, *config.MoreLambdaDetails)
		if err != nil {
			log.Error("ERROR: ", err)
			log.Fatal("ERROR: Failed to generate delete input structs")
		}

		log.Info("............")

		if *config.DryRun {
			numVerDeleted := countDeleteVersions(globalLambdaDeleteInputStructs)
			log.Info(fmt.Sprintf("%d unique versions will be removed in an actual execution.", numVerDeleted))
			spaceRemovedPreview := calculateSpaceRemoval(globalLambdaDeleteList)
			log.Info(fmt.Sprintf("%s of storage space will be removed in an actual execution.", calculateFileSize(uint64(spaceRemovedPreview), config)))

			displayDuration(startTime)

			return returnError
		}

		err = deleteLambdaVersion(ctx, svc, globalLambdaDeleteInputStructs...)
		if err != nil {
			log.Error("ERROR: ", err)
			log.Fatal("ERROR: Failed to delete Lambda versions.")
		}

		// Recalculate storage size
		updatedLambdaList, err := getAllLambdas(ctx, svc, customList)
		if err != nil {
			log.Error("ERROR: ", err)
			log.Fatal("ERROR: Failed to retrieve Lambda list.")
		}
		log.Info("............")

		for _, lambda := range updatedLambdaList {
			updatededlambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambda, *config)
			if err != nil {
				log.Error("ERROR: ", err)
				log.Fatal("ERROR: Failed to retrieve Lambda version list.")
			}

			updatedTotalLambdaStorage, err := getLambdaStorage(updatededlambdaVersionsList)
			if err != nil {
				log.Error("ERROR: ", err)
				log.Fatal("ERROR: Failed to retrieve Lambda storage size.")
			}

			updatedGlobalLambdaStorage = append(updatedGlobalLambdaStorage, updatedTotalLambdaStorage)
		}

		log.Info("............")
		var updatedCounter int64 = 0
		for _, v := range updatedGlobalLambdaStorage {
			updatedCounter = updatedCounter + v
		}

		log.Info("Total space freed up: ", (calculateFileSize(uint64(counter-updatedCounter), config)))
		log.Info("Post clean-up storage size: ", calculateFileSize(uint64(updatedCounter), config))
		log.Info("*********************************************")
	}

	if len(lambdaList) == 0 {
		log.Info("No lambdas found in ", *config.RegionFlag)
	}

	displayDuration(startTime)

	return returnError
}

func displayDuration(startTime time.Time) {
	var (
		elapsedTime float64
		timeUnit    string
	)

	t1 := time.Now()
	tempTime := t1.Sub(startTime)
	if tempTime.Minutes() > 1 {
		elapsedTime = tempTime.Minutes()
		timeUnit = "m"
	} else {
		elapsedTime = tempTime.Seconds()
		timeUnit = "s"
	}

	log.Infof("Job Duration Time: %f%s", elapsedTime, timeUnit)
}

// generateDeleteInputStructs takes a list of lambda.DeleteFunctionInput and a boolean value to determine if the user wants more details. The function returns a list of lambda.DeleteFunctionInput
// An error is returned if the function fails to execute
func generateDeleteInputStructs(versionsList [][]types.FunctionConfiguration, details bool) ([][]lambda.DeleteFunctionInput, error) {
	var (
		returnError error
		output      [][]lambda.DeleteFunctionInput
	)

	for _, version := range versionsList {
		var tempList []lambda.DeleteFunctionInput
		var functionName string

		for _, entry := range version {
			if *entry.Version != "$LATEST" {
				if functionName == "" {
					functionName = *entry.FunctionName
				}

				deleteItem := &lambda.DeleteFunctionInput{
					FunctionName: entry.FunctionName,
					Qualifier:    entry.Version,
				}

				tempList = append(tempList, *deleteItem)
			}
		}

		if details && functionName != "" {
			log.Info(fmt.Sprintf("%5d versions of %s to be removed", len(tempList), functionName))
		}

		output = append(output, tempList)
	}

	return output, returnError
}

// calculateSpaceRemoval returns the total size of all the versions to be deleted.
// The function takes a list of lambda.DeleteFunctionInput and returns an int
func calculateSpaceRemoval(deleteList [][]types.FunctionConfiguration) int {
	var (
		size int
	)

	for _, lambda := range deleteList {
		for _, version := range lambda {
			if *version.Version != "$LATEST" {
				size = size + int(version.CodeSize)
			}
		}
	}

	return size
}

// countDeleteVersions returns the total number of versions to be deleted.
// The function takes a list of lambda.DeleteFunctionInput and returns an int
func countDeleteVersions(deleteList [][]lambda.DeleteFunctionInput) int {
	var (
		versionsCount int
	)

	for _, lambda := range deleteList {
		versionsCount = versionsCount + len(lambda)
	}

	return versionsCount
}

// deleteLambdaVersion takes a list of lambda.DeleteFunctionInput and deletes all the versions in the list
// The function takes a context, a pointer to a lambda client, and a list of lambda.DeleteFunctionInput. A variadic operator is used to allow the user to pass in multiple lists of lambda.DeleteFunctionInput
// Use this function with caution as it will delete all the versions in the list.
func deleteLambdaVersion(ctx context.Context, svc *lambda.Client, deleteList ...[]lambda.DeleteFunctionInput) error {
	var (
		returnError error
		wg          sync.WaitGroup
	)

	for _, versions := range deleteList {
		for _, version := range versions {
			wg.Add(1)
			func() {
				defer wg.Done()
				_, err := svc.DeleteFunction(ctx, &version)
				if err != nil {
					log.Error(err)
					returnError = err
				}
			}()
		}
	}

	wg.Wait()
	return returnError
}

// getLambdasToDeleteList takes a list of lambda.FunctionConfiguration and a int8 value to determine how many versions to retain. The function returns a list of lambda.FunctionConfiguration
func getLambdasToDeleteList(list []types.FunctionConfiguration, retainCount int8) []types.FunctionConfiguration {
	var retainNumber int
	// Ensure the passed in parameter is greater than zero
	if retainCount >= 1 {
		retainNumber = int(retainCount)
	}

	// If passed in parameter is less than zero than set the default value to 0
	if retainCount < 1 {
		retainNumber = 1
	}

	// This checks to ensure that we are not deleting a list that only contains $LATEST
	if (len(list)) > 1 && (int(retainNumber) < len(list)) {
		return list[retainNumber:]
	} else {
		return nil
	}
}

// getAllLambdas returns a list of all available lambdas in the AWS environment. The function takes a context, a pointer to a lambda client, and a list of custom lambdas function names to delete
func getAllLambdas(ctx context.Context, svc *lambda.Client, customList []string) ([]types.FunctionConfiguration, error) {
	var (
		lambdasListOutput []types.FunctionConfiguration
		returnError       error
		input             *lambda.ListFunctionsInput
	)

	if len(customList) == 0 {
		input = &lambda.ListFunctionsInput{
			MaxItems: aws.Int32(maxItems),
		}

		p := lambda.NewListFunctionsPaginator(svc, input)
		for p.HasMorePages() {
			page, err := p.NextPage(ctx)
			if err != nil {
				log.Error(err)
				return lambdasListOutput, err
			}
			lambdasListOutput = append(lambdasListOutput, page.Functions...)
		}
	}

	if len(customList) > 0 {
		for _, item := range customList {

			input := &lambda.GetFunctionInput{
				FunctionName: aws.String(item),
			}

			result, err := svc.GetFunction(ctx, input)
			if err != nil {
				var rnf *types.ResourceNotFoundException
				if errors.As(err, &rnf) {
					log.Warn(fmt.Sprintf("The lambda function %s does not exist. Ensure you specified the correct name and that function exists and try again. ", item))
					log.Warn(fmt.Sprintf("Skipping %s", item))
					continue
				}
				returnError = err
			}
			if result != nil && result.Configuration != nil {
				lambdasListOutput = append(lambdasListOutput, *result.Configuration)
			}
		}
	}

	return lambdasListOutput, returnError
}

// getAllLambdaVersion returns a list of all available versions for a given lambda. The function takes a context, a pointer to a lambda client, and a lambda.FunctionConfiguration
func getAllLambdaVersion(
	ctx context.Context,
	svc *lambda.Client,
	item types.FunctionConfiguration,
	flags cliConfig,
) ([]types.FunctionConfiguration, error) {
	var (
		lambdasLisOutput []types.FunctionConfiguration
		returnError      error
		input            *lambda.ListVersionsByFunctionInput
	)

	input = &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(*item.FunctionName),
		MaxItems:     aws.Int32(maxItems),
	}

	p := lambda.NewListVersionsByFunctionPaginator(svc, input)
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			log.Error(err)
			return lambdasLisOutput, err
		}
		lambdasLisOutput = append(lambdasLisOutput, page.Versions...)
	}

	if *flags.SkipAliases {
		// fetch the list of aliases for this function
		aliasesOut, err := svc.ListAliases(ctx, &lambda.ListAliasesInput{
			FunctionName: aws.String(*item.FunctionArn),
		})
		if err != nil {
			log.Error(err)
			return lambdasLisOutput, err
		}

		// produce a new slice that includes only versions for which there is no alias
		var result []types.FunctionConfiguration

		// iterate over the list of versions and check if there is an alias for each
		// if there is no alias, add the version to the result
		// if there is an alias, skip the version
		// this is done to avoid deleting versions that are in use by an alias
		for _, funConf := range lambdasLisOutput {
			isAlias := false

			for _, alias := range aliasesOut.Aliases {
				if alias.FunctionVersion != nil && *alias.FunctionVersion == *funConf.Version {
					isAlias = true
					break
				}
			}

			if !isAlias {
				result = append(result, funConf)
			}
		}

		// return the pared down list of versions
		lambdasLisOutput = result
	}

	// Sort list so that the former versions are listed first and $LATEST is listed last
	sort.Sort(byVersion(lambdasLisOutput))

	return lambdasLisOutput, returnError
}

type byVersion []types.FunctionConfiguration

func (a byVersion) Len() int { return len(a) }

func (a byVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byVersion) Less(i, j int) bool {
	one, _ := strconv.ParseInt(*a[i].Version, 10, 32)
	two, _ := strconv.ParseInt(*a[j].Version, 10, 32)
	return one > two
}

// getLambdaStorage calculates the aggregate sum of all the functions' size
func getLambdaStorage(list []types.FunctionConfiguration) (int64, error) {
	var (
		sizeCounter int64
		returnError error
	)

	for _, item := range list {
		sizeCounter = sizeCounter + item.CodeSize
	}

	return sizeCounter, returnError
}

// validateRegion validates the user input to ensure it is a valid AWS region. The function takes a embed.FS and a string. The function returns a string and an error
// An embedded file is used to validate the user input. The embedded file contains a list of all the AWS regions
// Example of the embedded file: ap-south-2	ap-south-1	eu-south-1	eu-south-2	me-central-1	ca-central-1	eu-central-1	eu-central-2
func validateRegion(f embed.FS, input string) (string, error) {
	var output string
	var err error

	rawData, _ := f.ReadFile(regionFile)
	regionsList := strings.Split(string(rawData), "	")

	for _, region := range regionsList {
		if strings.ToLower(input) == strings.TrimSpace(region) {
			output = strings.TrimSpace(region)
		}
	}

	if output == "" {
		err = errors.New(input + " is an invalid AWS region. If this is an error please report it")
		return "", err
	}

	return output, err
}

// calculateFileSize returns the size of a file in bytes. The function takes a cliConfig parameter to determine the number format type to return
func calculateFileSize(value uint64, config *cliConfig) string {
	if *config.SizeIEC {
		return humanize.IBytes(value)
	}
	return humanize.Bytes(value)
}
