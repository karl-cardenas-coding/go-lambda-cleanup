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
	ctx               context.Context
	CustomeDeleteList []string
	svc               *lambda.Client
	//go:embed aws-regions.txt
	f embed.FS
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes all former versions of AWS lambdas except for the $LATEST version",
	Long:  `Removes all former versions of AWS lambdas except for the $LATEST version. The user also has the ability specify n-? version to retain.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx = context.Background()

		var (
			awsEnvRegion  string
			awsEnvProfile string
			config        cliConfig
		)

		config = GlobalCliConfig
		awsEnvRegion = os.Getenv("AWS_DEFAULT_REGION")
		awsEnvProfile = os.Getenv("AWS_PROFILE")
		if *config.RegionFlag == "" {
			if awsEnvRegion != "" {
				*config.RegionFlag = validateRegion(f, awsEnvRegion)
			} else {
				log.Fatal("ERROR: Missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region.")
			}
		} else {
			*config.RegionFlag = validateRegion(f, *config.RegionFlag)
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

		if *config.LambdaListFile != "" {
			log.Info("******** CUSTOM LAMBDA LIST PROVIDED ********")
			customList, err := internal.GenerateLambdaDeleteList(LambdaListFile)
			if err != nil {
				log.Info(err.Error())
			}
			CustomeDeleteList = customList
		}

		cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfigOptions...)
		if err != nil {
			log.Fatal("ERROR ESTABLISHING AWS SESSION")
		}

		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			log.Fatal("ERROR RETRIEVING AWS CREDENTIALS")
		}
		if creds.Expired() {
			log.Fatal("AWS CREDENTIALS EXPIRED")
		}

		// svc = lambda.NewFromConfig(cfg)
		svc = lambda.NewFromConfig(cfg, func(o *lambda.Options) {
			// Set the User-Agent for all AWS with the Lambda client
			o.APIOptions = append(o.APIOptions, middleware.AddUserAgentKeyValue("go-lambda-cleanup", VersionString))
		})
		err = executeClean(&config)
		if err != nil {
			log.Fatal("ERROR: ", err)
		}
	},
}

// An action function that removes Lambda versions
func executeClean(config *cliConfig) error {
	startTime := time.Now()

	var (
		returnError                error
		globalLambdaStorage        []int64
		updatedGlobalLambdaStorage []int64
		globalLambdaVersionsList   [][]types.FunctionConfiguration
		counter                    int64 = 0
	)

	log.Info("Scanning AWS environment in " + *config.RegionFlag)
	lambdaList, err := getAllLambdas(ctx, svc, CustomeDeleteList)
	if err != nil {
		log.Error("ERROR: ", err)
		log.Fatal("ERROR: Failed to retrieve Lambda list.")
	}
	log.Info("............")

	if len(lambdaList) > 0 {
		tempCounter := 0
		for _, lambda := range lambdaList {
			lambdaItem := lambda
			lambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambdaItem)
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

		if DryRun {
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
		updatedLambdaList, err := getAllLambdas(ctx, svc, CustomeDeleteList)
		if err != nil {
			log.Error("ERROR: ", err)
			log.Fatal("ERROR: Failed to retrieve Lambda list.")
		}
		log.Info("............")

		for _, lambda := range updatedLambdaList {
			updatededlambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambda)
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

// Generates a list of Lambda version delete structs
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

// Returns a count of versions in a slice of lambda.DeleteFunctionInput
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

// Returns a count of versions in a slice of lambda.DeleteFunctionInput
func countDeleteVersions(deleteList [][]lambda.DeleteFunctionInput) int {
	var (
		versionsCount int
	)

	for _, lambda := range deleteList {
		versionsCount = versionsCount + len(lambda)
	}

	return versionsCount
}

// Deletes all Lambda versions specified in the input list
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

// Generate a list of Lambdas to remove based on the desired retain value
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

// Return a list of all Lambdas in the respective AWS account
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
			lambdasListOutput = append(lambdasListOutput, *result.Configuration)
		}
	}

	return lambdasListOutput, returnError
}

// A function that returns all the version of a Lambda
func getAllLambdaVersion(ctx context.Context, svc *lambda.Client, item types.FunctionConfiguration) ([]types.FunctionConfiguration, error) {
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

// A function that calculates the aggregate sum of all the functions' size
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

// Validates that the user passed in a valid AWS Region
func validateRegion(f embed.FS, input string) string {
	var output string

	rawData, _ := f.ReadFile(regionFile)
	regionsList := strings.Split(string(rawData), "	")

	for _, region := range regionsList {
		if strings.ToLower(input) == strings.TrimSpace(region) {
			output = strings.TrimSpace(region)
		}
	}

	if output == "" {
		log.Fatal(input, " is an invalid AWS region. If this is an error please")
	}

	return output
}

// calculateFileSize returns the size of a file in bytes. The function takes a cliConfig paramter to determine the number format type to return
func calculateFileSize(value uint64, config *cliConfig) string {
	if *config.SizeIEC {
		return humanize.IBytes(value)
	}
	return humanize.Bytes(value)
}
