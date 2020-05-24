package movie

// Re http://gosparta.io/reference/apigateway/echo_event/
//    https://github.com/mweagle/SpartaHTML/blob/master/main.go

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sburke-at-ziprecruiter/sparta-serverless-demo/pkg/config"
	"github.com/sirupsen/logrus"

	sparta "github.com/mweagle/Sparta"
	spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
)

type movieResponse struct {
	Movie   Movie
	Message string
	Request spartaAWSEvents.APIGatewayRequest
}

func getMovie(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// Get the movie from DynamoDB
	mov := Movie{}
	msg := "no path param year,title"
	if ystr, ok := gatewayEvent.PathParams["year"]; ok {
		if year, err := strconv.Atoi(ystr); err != nil {
			msg = fmt.Sprintf("year %q - %s", ystr, err.Error())
		} else {
			msg = "no path param title"
			if title, ok := gatewayEvent.PathParams["title"]; ok {
				var err error
				msg = "getMovie year:" + ystr + " title:" + title
				if loggerOk {
					logger.Info(msg)
				}
				mov, err = Get(year, title)
				if err != nil {
					msg = "movie.Get() -> " + err.Error()
					if loggerOk {
						logger.Errorf(msg)
					}
				}
			}
		}
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, &movieResponse{
		Movie:   mov,
		Message: msg,
		Request: gatewayEvent,
	}), nil
}

func LambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn, _ := sparta.NewAWSLambda(
		config.ShortLambdaName(getMovie),
		getMovie,
		"Sparta-Lambda-DynamoDB",
	)
	if nil != api {
		apiGatewayResource, _ := api.NewResource("/movie/{year}/{title}", lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create /movie resource: " + apiMethodErr.Error())
		}
		// The lambda resource only supports application/json Unmarshallable requests.
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path.year"] = true
		apiMethod.Parameters["method.request.path.title"] = true
	}
	return append(lambdaFunctions, lambdaFn)
}
