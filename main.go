// File: main.go
package main

import (
	"fmt"

	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/config"
	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/customer"
	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/movie"
	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/store"

	sparta "github.com/mweagle/Sparta"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
)

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
