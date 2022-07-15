package cmd

import (
	"context"
	"crypto/tls"
	"embed"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/dustin/go-humanize"
	internal "github.com/karl-cardenas-coding/go-lambda-cleanup/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// Per AWS API Valid Range: Minimum value of 1. Maximum value of 10000.
	maxItems   int64  = 10000
	regionFile string = "aws-regions.txt"
)

var (
	ctx               context.Context
	CustomeDeleteList []string
	logLevel          aws.LogLevelType
	svc               *lambda.Lambda
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

		var (
			awsEnvRegion     string
			awsEnvProfile    string
			profile          string
			region           string
			sharedFileConfig session.SharedConfigState
			userAgent        string
		)

		awsEnvRegion = os.Getenv("AWS_DEFAULT_REGION")
		awsEnvProfile = os.Getenv("AWS_PROFILE")
		userAgent = fmt.Sprintf("go-lambda-cleanup-%s", VersionString)

		if RegionFlag == "" {
			if awsEnvRegion != "" {
				region = validateRegion(f, awsEnvRegion)
			} else {
				log.Fatal("ERROR: Missing region flag and AWS_DEFAULT_REGION env variable. Please use -r and provide a valid AWS region.")
			}
		} else {
			region = validateRegion(f, RegionFlag)
		}

		ctx = context.Background()
		// Initialize parameters

		// Setup client header to use TLS 1.2
		tr := &http.Transport{
			// Reads PROXY configuration from environment variables
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}

		// Needed due to custom client being leveraged, otherwise HTTP2 will not be used.
		tr.ForceAttemptHTTP2 = true

		// Create the client
		client := http.Client{Transport: tr}

		if ProfileFlag == "" {
			if awsEnvProfile != "" {
				log.Infof("AWS_PROFILE set to \"%s\"", awsEnvProfile)
				profile = awsEnvProfile
			} else {
				profile = ""
			}
		} else {
			log.Infof("The AWS Profile flag \"%s\" was passed in", ProfileFlag)
			profile = ProfileFlag
		}

		if Verbose {
			logLevel = aws.LogDebugWithRequestErrors
		} else {
			logLevel = aws.LogOff
		}

		if DryRun {
			log.Info("******** DRY RUN MODE ENABLED ********")
		}

		if LambdaListFile != "" {
			customList, err := internal.GenerateLambdaDeleteList(LambdaListFile)
			if err != nil {
				log.Info(err.Error())
			}

			CustomeDeleteList = customList
		}

		if CredentialsFile {
			sharedFileConfig = session.SharedConfigEnable
		} else {
			sharedFileConfig = session.SharedConfigDisable
		}

		sess, err := session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:                        aws.String(region),
				CredentialsChainVerboseErrors: aws.Bool(true),
				LogLevel:                      aws.LogLevel(logLevel),
				HTTPClient:                    &client,
				MaxRetries:                    aws.Int(1),
			},
			SharedConfigState: sharedFileConfig,
			Profile:           profile,
		})

		if err != nil {
			log.Fatal("ERROR ESTABLISHING AWS SESSION")
		}

		sess.Config.Credentials.Expire()
		_, err = sess.Config.Credentials.Get()
		if err != nil {
			log.Fatal("ERROR: Failed to acquire valid credentials.")
		}

		sessVerified := session.Must(sess, err)

		svc = lambda.New(sessVerified)
		// Set the User-Agent for all AWS connections
		svc.Handlers.Send.PushFront(func(r *request.Request) {
			r.HTTPRequest.Header.Set("User-Agent", userAgent)
		})

		err = executeClean(region)
		if err != nil {
			log.Fatal("ERROR: ", err)
		}
	},
}

// An action function that removes Lambda versions
func executeClean(region string) error {
	startTime := time.Now()

	var (
		returnError                error
		globalLambdaStorage        []int64
		updatedGlobalLambdaStorage []int64
		globalLambdaVersionsList   [][]*lambda.FunctionConfiguration
		counter                    int64 = 0
		elapsedTime                float64
		timeUnit                   string
	)

	log.Info("Scanning AWS environment in " + region)
	lambdaList, err := getAllLambdas(ctx, svc, CustomeDeleteList)
	checkError(err)
	log.Info("............")

	if len(lambdaList) > 0 {
		tempCounter := 0
		for _, lambda := range lambdaList {
			lambdaItem := lambda
			lambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambdaItem)
			checkError(err)

			globalLambdaVersionsList = append(globalLambdaVersionsList, lambdaVersionsList)

			totalLambdaStorage, err := getLambdaStorage(lambdaVersionsList)
			checkError(err)

			globalLambdaStorage = append(globalLambdaStorage, totalLambdaStorage)
			tempCounter++
		}

		log.Info(tempCounter, " Lambdas identified")
		for _, v := range globalLambdaStorage {
			counter = counter + v
		}

		log.Info("Current storage size: ", humanize.Bytes(uint64(counter)))
		log.Info("**************************")
		log.Info("Initiating clean-up process. This may take a few minutes....")
		// Begin delete process
		globalLambdaDeleteList := [][]*lambda.FunctionConfiguration{}

		for _, lambda := range globalLambdaVersionsList {
			lambdasDeleteList := getLambdasToDeleteList(lambda, Retain)
			globalLambdaDeleteList = append(globalLambdaDeleteList, lambdasDeleteList)
		}

		log.Info("............")
		globalLambdaDeleteInputStructs, err := generateDeleteInputStructs(globalLambdaDeleteList)
		checkError(err)

		log.Info("............")

		if DryRun {
			numVerDeleted := countDeleteVersions(globalLambdaDeleteInputStructs)
			log.Info(fmt.Sprintf("%d unique versions will be removed in an actual execution.", numVerDeleted))
			spaceRemovedPreview := calculateSpaceRemoval(globalLambdaDeleteList)
			log.Info(fmt.Sprintf("%s of storage space will be removed in an actual execution.", humanize.Bytes(uint64(spaceRemovedPreview))))
		} else {
			err = deleteLambdaVersion(ctx, svc, globalLambdaDeleteInputStructs...)
			checkError(err)

			// Recalculate storage size
			updatedLambdaList, err := getAllLambdas(ctx, svc, CustomeDeleteList)
			checkError(err)
			log.Info("............")

			for _, lambda := range updatedLambdaList {
				updatededlambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambda)
				checkError(err)

				updatedTotalLambdaStorage, err := getLambdaStorage(updatededlambdaVersionsList)
				checkError(err)

				updatedGlobalLambdaStorage = append(updatedGlobalLambdaStorage, updatedTotalLambdaStorage)
			}

			log.Info("............")
			var updatedCounter int64 = 0
			for _, v := range updatedGlobalLambdaStorage {
				updatedCounter = updatedCounter + v
			}

			log.Info("Total space freed up: ", (humanize.Bytes(uint64(counter - updatedCounter))))
			log.Info("Post clean-up storage size: ", humanize.Bytes(uint64(updatedCounter)))
			log.Info("*********************************************")
		}
	}

	if len(lambdaList) == 0 {
		log.Info("No lambdas found in ", region)
	}

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

	return returnError
}

// Generates a list of Lambda version delete structs
func generateDeleteInputStructs(versionsList [][]*lambda.FunctionConfiguration) ([][]lambda.DeleteFunctionInput, error) {
	var (
		returnError error
		output      [][]lambda.DeleteFunctionInput
	)

	for _, version := range versionsList {
		var tempList []lambda.DeleteFunctionInput
		var fName = ""

		for _, entry := range version {
			if *entry.Version != "$LATEST" {
				if fName == "" {
					fName = *entry.FunctionName
				}

				deleteItem := &lambda.DeleteFunctionInput{
					FunctionName: entry.FunctionName,
					Qualifier:    entry.Version,
				}

				tempList = append(tempList, *deleteItem)
			}
		}

		if MoreDetail && fName != "" {
			log.Info(fmt.Sprintf("%d versions of %s to be removed", len(tempList), fName))
		}

		output = append(output, tempList)
	}

	return output, returnError
}

// Returns a count of versions in a slice of lambda.DeleteFunctionInput
func calculateSpaceRemoval(deleteList [][]*lambda.FunctionConfiguration) int {
	var (
		size int
	)

	for _, lambda := range deleteList {
		for _, version := range lambda {
			if *version.Version != "$LATEST" {
				size = size + int(*version.CodeSize)
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
func deleteLambdaVersion(ctx context.Context, svc *lambda.Lambda, deleteList ...[]lambda.DeleteFunctionInput) error {
	var (
		returnError error
		wg          sync.WaitGroup
	)

	for _, versions := range deleteList {
		for _, version := range versions {
			wg.Add(1)
			func() {
				defer wg.Done()
				_, err := svc.DeleteFunction(&version)
				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						case lambda.ErrCodeServiceException:
							log.Error("Function Name: ", *version.FunctionName)
							returnError = aerr
						case lambda.ErrCodeResourceNotFoundException:
							log.Error("Function Name: ", *version.FunctionName)
							returnError = aerr
						case lambda.ErrCodeTooManyRequestsException:
							log.Error("Function Name: ", *version.FunctionName)
							returnError = aerr
						case lambda.ErrCodeInvalidParameterValueException:
							log.Error("Function Name: ", *version.FunctionName)
							returnError = aerr
						case lambda.ErrCodeResourceConflictException:
							log.Error("Function Name: ", *version.FunctionName)
							returnError = aerr
						default:
							log.Error("Function Name: ", *version.FunctionName)
							returnError = aerr
						}
					}
				}
			}()
		}
	}

	wg.Wait()
	return returnError
}

// Generate a list of Lambdas to remove based on the desired retain value
func getLambdasToDeleteList(list []*lambda.FunctionConfiguration, retainCount int8) []*lambda.FunctionConfiguration {
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
func getAllLambdas(ctx context.Context, svc *lambda.Lambda, customList []string) ([]*lambda.FunctionConfiguration, error) {
	var (
		lambdasListOutput []*lambda.FunctionConfiguration
		returnError       error
		input             *lambda.ListFunctionsInput
	)

	if len(customList) == 0 {
		input = &lambda.ListFunctionsInput{
			MaxItems: aws.Int64(maxItems),
		}

		// Loop condition variable
		loopBreaker := false

		for {
			err := svc.ListFunctionsPagesWithContext(ctx, input,
				func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
					lambdasListOutput = append(lambdasListOutput, page.Functions...)

					// Set the next marker indicator for the next iteration
					input.Marker = page.NextMarker
					//  Set condition variable to a new condition value
					loopBreaker = lastPage
					return lastPage
				})
			if err != nil {
				return lambdasListOutput, err
			}

			if loopBreaker {
				break
			}
		}

	}

	if len(customList) > 0 {
		for _, item := range customList {

			input := &lambda.GetFunctionInput{
				FunctionName: aws.String(item),
			}

			result, err := svc.GetFunctionWithContext(ctx, input)
			if err != nil {
				returnError = err
			}

			lambdasListOutput = append(lambdasListOutput, result.Configuration)
		}
	}

	return lambdasListOutput, returnError
}

// A function that returns all the version of a Lambda
func getAllLambdaVersion(ctx context.Context, svc *lambda.Lambda, item *lambda.FunctionConfiguration) ([]*lambda.FunctionConfiguration, error) {
	var (
		lambdasLisOutput []*lambda.FunctionConfiguration
		returnError      error
		input            *lambda.ListVersionsByFunctionInput
	)

	input = &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(*item.FunctionName),
		MaxItems:     aws.Int64(maxItems),
	}

	// Loop condition variable
	loopBreaker := false

	for {
		err := svc.ListVersionsByFunctionPagesWithContext(ctx, input,
			func(page *lambda.ListVersionsByFunctionOutput, lastPage bool) bool {

				lambdasLisOutput = append(lambdasLisOutput, page.Versions...)
				// Set the next marker indicator for the next iteration
				input.Marker = page.NextMarker
				//  Set condition variable to a new condition value
				loopBreaker = lastPage
				return lastPage
			})

		if err != nil {
			return lambdasLisOutput, returnError
		}

		if loopBreaker {
			break
		}
	}

	// Sort list so that the former versions are listed first and $LATEST is listed last
	sort.Sort(byVersion(lambdasLisOutput))

	return lambdasLisOutput, returnError
}

type byVersion []*lambda.FunctionConfiguration

func (a byVersion) Len() int      { return len(a) }
func (a byVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byVersion) Less(i, j int) bool {
	one, _ := strconv.ParseInt(*a[i].Version, 10, 32)
	two, _ := strconv.ParseInt(*a[j].Version, 10, 32)
	return one > two
}

// A function that calculates the aggregate sum of all the functions' size
func getLambdaStorage(list []*lambda.FunctionConfiguration) (int64, error) {
	var (
		sizeCounter int64
		returnError error
	)

	for _, item := range list {
		sizeCounter = sizeCounter + *item.CodeSize
	}

	return sizeCounter, returnError
}

func checkError(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case organizations.ErrCodeAccessDeniedException:
				log.Fatal("ERROR: Access Denied - Please verify AWS credentials and permissions\n", aerr.Code())
			case lambda.ErrCodeResourceConflictException:
				log.Fatal("ERROR: ", err.Error())
			case lambda.ErrCodeResourceNotFoundException:
				log.Fatal("ERROR: ", "Invalid Lambda(s) provided. Please check the function name provided. \n")
			default:
				log.Fatal("ERROR: ", err.Error())
			}
		}
	}
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
