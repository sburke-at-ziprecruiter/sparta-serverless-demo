// File: main.go
package main

import (
	"MyGo/dynamo_db/pkg/config"
	"MyGo/dynamo_db/pkg/customer"
	"MyGo/dynamo_db/pkg/movie"
	"MyGo/dynamo_db/pkg/store"
	"archive/zip"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	"github.com/sirupsen/logrus"
)

// archiveHook adds config.json to the Lambda zip archive.
func archiveHook(context map[string]interface{},
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {

	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic("Reading greeting " + err.Error())
	}
	iow, err := zipWriter.Create("greeting")
	if err != nil {
		panic("Creating zip/greeting " + err.Error())
	}
	_, err = iow.Write(data)
	if err != nil {
		panic("Writing zip/greeting " + err.Error())
	}
	return err
}

func main() {

	// Register the function with the API Gateway
	apiStage := sparta.NewStage(config.Config.APIStage)
	apiGateway := sparta.NewAPIGateway(config.Config.APIName, apiStage)

	lambdaFunctions := []*sparta.LambdaAWSInfo{}
	lambdaFunctions = append(lambdaFunctions, customer.LambdaFunctions(apiGateway)...)
	lambdaFunctions = append(lambdaFunctions, movie.LambdaFunctions(apiGateway)...)
	lambdaFunctions = append(lambdaFunctions, store.LambdaFunctions(apiGateway)...)

	// Deploy it
	stackName := spartaCF.UserScopedStackName(config.Config.APIName)
	sparta.MainEx(stackName,
		fmt.Sprintf("Provision API Gateway %s backed by Lambda functions and DynamoDB table %s", config.Config.APIName, config.Config.Table),
		lambdaFunctions,
		apiGateway,
		nil,
		&sparta.WorkflowHooks{
			Archives: []sparta.ArchiveHookHandler{sparta.ArchiveHookFunc(config.ArchiveHook)},
		},
		false)
}
