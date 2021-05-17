package customer

// Re http://gosparta.io/reference/apigateway/echo_event/
//    https://github.com/mweagle/SpartaHTML/blob/master/main.go

import (
	"context"
	"fmt"
	"net/http"

	"github.com/seanpburke/sparta-serverless-demo/pkg/config"
	"github.com/sirupsen/logrus"

	sparta "github.com/mweagle/Sparta"
	spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
)

type customerResponse struct {
	Customer Customer
	Message  string
	Request  spartaAWSEvents.APIGatewayRequest
}

func getCustomer(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// Get my customer from DynamoDB
	cus := Customer{}
	msg := "no path param phone"
	if phone, ok := gatewayEvent.PathParams["phone"]; ok {
		var err error
		msg = "GetCustomer phone:" + phone
		if loggerOk {
			logger.Info(msg)
		}
		cus, err = Get(phone)
		if err != nil {
			msg = "customer.Get() -> " + err.Error()
			if loggerOk {
				logger.Errorf(msg)
			}
		}
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, &customerResponse{
		Customer: cus,
		Message:  msg,
		Request:  gatewayEvent,
	}), nil
}

func LambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	paramName := "phone"
	lambdaFn, _ := sparta.NewAWSLambda(
		config.ShortLambdaName(getCustomer),
		getCustomer,
		"Sparta-Lambda-DynamoDB",
	)
	if nil != api {
		apiGatewayResource, _ := api.NewResource(fmt.Sprintf("/customer/{%s}", paramName), lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create /hello resource: " + apiMethodErr.Error())
		}
		// The lambda resource only supports application/json Unmarshallable requests.
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path."+paramName] = true
	}
	return append(lambdaFunctions, lambdaFn)
}
