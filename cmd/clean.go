package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
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

	// Setup the AWS Session with our desired configurations
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(RegionFlag),
		CredentialsChainVerboseErrors: aws.Bool(true),
		LogLevel:                      aws.LogLevel(logLevel),
		HTTPClient:                    &client,
	})
	if err != nil {
		panic("ERROR SETTING UP SESSION")
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Fatal("ERROR: Failed to valid aquire credentials.")
	}

	// The must() helps us ensure that the connection/session is leveraging all our specified client configurations.
	sessVerified := session.Must(sess, err)

	svc = lambda.New(sessVerified)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes all versions of lambda expect for the $LATEST version",
	Long:  `Removes all versions of lambda expect for the $LATEST bersion. The user also has the ability specify n-? version to retain.`,
	Run: func(cmd *cobra.Command, args []string) {
		if RegionFlag != "" {
			err := executeClean()
			if err != nil {
				log.Fatal("ERROR: ", err)
			}

		} else {
			log.Println("ERROR: Missing region flag. Please use -r and provide a valid AWS region.")
		}
	},
}

func executeClean() error {

	var (
		returnError error
		// errorList                []error
		globalLambdaStorage        []int64
		updatedGlobalLambdaStorage []int64
		globalLambdaVersionsList   [][]*lambda.FunctionConfiguration
	)

	// g, ctx := errgroup.WithContext(ctx)
	log.Println("Scanning AWS environment in " + RegionFlag + ".....")
	lambdaList, err := getAlllambdas(ctx, svc)
	checkError(err)
	log.Println("............")
	for _, lambda := range lambdaList {
		// g.Go(func() error {
		lambdaVersionsList, err := getAllLambdaVersion(ctx, svc, lambda)
		checkError(err)

		// if err != nil {
		// 	errorList = append(errorList, returnError)
		// }
		globalLambdaVersionsList = append(globalLambdaVersionsList, lambdaVersionsList)
		totalLambdaStorage, err := getLambdaStorage(lambdaVersionsList)
		checkError(err)
		// if err != nil {
		// 	errorList = append(errorList, returnError)
		// }

		// mutex.Lock()
		globalLambdaStorage = append(globalLambdaStorage, totalLambdaStorage)
		// mutex.Unlock()
	}

	var counter int64 = 0
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
	returnError = checkError(err)
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

	// 		return err
	// 	})
	// }
	// if err := g.Wait(); err == nil || err == context.Canceled {
	// 	log.Println(".....Processing")
	// } else {
	// 	log.Println(err)
	// 	for _, err := range errorList {
	// 		log.Println(err.Error())
	// 	}
	// }

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
func getLambdasToDelteList(list []*lambda.FunctionConfiguration, retainCount int32) []*lambda.FunctionConfiguration {

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
	returnError = checkError(err)

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
	returnError = checkError(err)

	// Sort list so that the fomer versions are listed firstm and $LATEST is listed last
	sort.Sort(ByVersion(lambdasLisOutput))

	return lambdasLisOutput, returnError

}

type ByVersion []*lambda.FunctionConfiguration

func (a ByVersion) Len() int      { return len(a) }
func (a ByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool {
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

func checkError(err error) error {
	if err != nil {
		return err
	} else {
		return nil
	}
}
