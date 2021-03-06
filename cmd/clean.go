package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

const (
	maxItems int64 = 1000
)

var (
	ctx      context.Context
	logLevel aws.LogLevelType
	svc      *lambda.Lambda
)

func init() {
	rootCmd.AddCommand(cleanCmd)

}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes all versions of lambda except for the $LATEST version",
	Long:  `Removes all versions of lambda except for the $LATEST bersion. The user also has the ability specify n-? version to retain.`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			awsEnvRegion     string
			profile          string
			region           string
			sharedFileConfig session.SharedConfigState
		)

		awsEnvRegion = os.Getenv("AWS_DEFAULT_REGION")
		awsEnvProfile := os.Getenv("AWS_PROFILE")

		if awsEnvRegion == "" {
			if RegionFlag != "" {
				region = RegionFlag
			}
		} else {
			region = awsEnvRegion
		}

		if awsEnvProfile == "" {
			if ProfileFlag != "" {
				profile = ProfileFlag
			}
		} else {
			profile = awsEnvProfile
		}

		if region != "" {
			ctx = context.Background()
			// Initialize parameters

			// Setup client header to use TLS 1.2
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			}
			// Needed due to custom client being leveraged, otherwise HTTP2 will not be used.
			tr.ForceAttemptHTTP2 = true

			// Create the client
			client := http.Client{Transport: tr}

			if Debug {
				logLevel = aws.LogDebugWithRequestErrors
			} else {
				logLevel = aws.LogOff
			}

			if CredentialsFile {
				sharedFileConfig = session.SharedConfigEnable

			} else {
				sharedFileConfig = session.SharedConfigDisable
			}

			// The must() helps us ensure that the connection/session is leveraging all our specified client configurations.
			// sessVerified := session.Must(sess, err)
			sessVerified := session.Must(session.NewSessionWithOptions(session.Options{
				Config: aws.Config{
					Region:                        aws.String(region),
					CredentialsChainVerboseErrors: aws.Bool(true),
					LogLevel:                      aws.LogLevel(logLevel),
					HTTPClient:                    &client,
					MaxRetries:                    aws.Int(1),
				},
				Profile:           profile,
				SharedConfigState: sharedFileConfig,
			}))

			sessVerified.Config.Credentials.Expire()
			_, err := sessVerified.Config.Credentials.Get()
			if err != nil {
				log.Fatal("ERROR: Failed to valid aquire credentials.")
			}

			svc = lambda.New(sessVerified)
			err = executeClean(region)
			if err != nil {
				log.Fatal("ERROR: ", err)
			}

		} else {
			log.Println("ERROR: Missing region flag. Please use -r and provide a valid AWS region.")
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
	)

	log.Println("Scanning AWS environment in " + region + ".....")
	lambdaList, err := getAlllambdas(ctx, svc)
	checkError(err)
	log.Println("............")

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
		log.Println(tempCounter, " Lambdas identified")
		for _, v := range globalLambdaStorage {
			counter = counter + v
		}
		log.Println("Current storage size: ", humanize.Bytes(uint64(counter)))
		log.Println("**************************")
		log.Println("Initiating clean-up process. This may take a few minutes....")
		// Begin delete process
		globalLambdaDeleteList := [][]*lambda.FunctionConfiguration{}

		for _, lambda := range globalLambdaVersionsList {
			lambdasDeleteList := getLambdasToDelteList(lambda, Retain)
			globalLambdaDeleteList = append(globalLambdaDeleteList, lambdasDeleteList)

		}
		log.Println("............")
		globalLambdaDeleteInputStructs, err := generateDeleteInputStructs(globalLambdaDeleteList)
		checkError(err)

		log.Println("............")
		err = deleteLambdaVersion(ctx, svc, globalLambdaDeleteInputStructs...)
		checkError(err)

		// Recalculate storage size
		updatedLambdaList, err := getAlllambdas(ctx, svc)
		checkError(err)
		log.Println("............")

		for _, lambda := range updatedLambdaList {
			updatededlambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambda)
			checkError(err)

			updatedTotalLambdaStorage, err := getLambdaStorage(updatededlambdaVersionsList)
			checkError(err)

			updatedGlobalLambdaStorage = append(updatedGlobalLambdaStorage, updatedTotalLambdaStorage)
		}
		log.Println("............")
		var updatedCounter int64 = 0
		for _, v := range updatedGlobalLambdaStorage {
			updatedCounter = updatedCounter + v
		}
		log.Println("Total space freed up: ", (humanize.Bytes(uint64(counter - updatedCounter))))
		log.Println("Post clean-up storage size: ", humanize.Bytes(uint64(updatedCounter)))
		log.Println("*********************************************")

		t := time.Now()
		elapsedTime := time.Duration(t.Sub(startTime).Minutes())
		log.Println("Job Duration Time: ", elapsedTime)
	}

	if len(lambdaList) == 0 {
		log.Println("No lambdas found in", region)
	}

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

		for _, version := range version {
			deleteItem := &lambda.DeleteFunctionInput{
				FunctionName: version.FunctionName,
				Qualifier:    version.Version,
			}

			tempList = append(tempList, *deleteItem)

		}

		output = append(output, tempList)

	}

	return output, returnError

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
							fmt.Println(lambda.ErrCodeServiceException, aerr.Error())
							returnError = aerr
						case lambda.ErrCodeResourceNotFoundException:
							fmt.Println(lambda.ErrCodeResourceNotFoundException, aerr.Error())
							returnError = aerr
						case lambda.ErrCodeTooManyRequestsException:
							fmt.Println(lambda.ErrCodeTooManyRequestsException, aerr.Error())
							returnError = aerr
						case lambda.ErrCodeInvalidParameterValueException:
							fmt.Println(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
							returnError = aerr
						case lambda.ErrCodeResourceConflictException:
							fmt.Println(lambda.ErrCodeResourceConflictException, aerr.Error())
							returnError = aerr
						default:
							fmt.Println(aerr.Error())
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
func getLambdasToDelteList(list []*lambda.FunctionConfiguration, retainCount int8) []*lambda.FunctionConfiguration {

	var retainNumber int
	// Ensure the passed in parameter is greater than zero
	if retainCount >= 1 {
		retainNumber = int(retainCount)
	}

	// If passed in parameter is less than zero than set the default value to 0
	if retainCount < 1 {
		retainNumber = 1
	}

	// This checks to ensure that we are not deleting a list tha only contains $LATEST
	if (len(list)-1) > 1 && (int(retainNumber) < len(list)-1) {
		return list[retainNumber:(len(list) - 1)]
	} else {
		return nil
	}
}

// Return a list of all Lambdas in the respective AWS account
func getAlllambdas(ctx context.Context, svc *lambda.Lambda) ([]*lambda.FunctionConfiguration, error) {

	var (
		lambdasLisOutput []*lambda.FunctionConfiguration
		returnError      error
		input            *lambda.ListFunctionsInput
	)

	input = &lambda.ListFunctionsInput{
		FunctionVersion: aws.String("ALL"),
		MaxItems:        aws.Int64(maxItems),
	}

	pageNum := 0
	err := svc.ListFunctionsPagesWithContext(ctx, input,
		func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
			pageNum++
			lambdasLisOutput = append(lambdasLisOutput, page.Functions...)
			return lastPage
		})
	if err != nil {
		returnError = err
	}

	return lambdasLisOutput, returnError

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

	pageNum := 0
	err := svc.ListVersionsByFunctionPagesWithContext(ctx, input,
		func(page *lambda.ListVersionsByFunctionOutput, lastPage bool) bool {
			pageNum++
			lambdasLisOutput = append(lambdasLisOutput, page.Versions...)
			return lastPage
		})
	if err != nil {
		returnError = err
	}

	// Sort list so that the fomer versions are listed firstm and $LATEST is listed last
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
			default:
				log.Fatal("ERROR: ", err.Error())
			}
		}
	}
}
