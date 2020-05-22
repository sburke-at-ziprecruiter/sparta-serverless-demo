package store

// Re http://gosparta.io/reference/apigateway/echo_event/
//    https://github.com/mweagle/SpartaHTML/blob/master/main.go

import (
	"MyGo/dynamo_db/pkg/customer"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	sparta "github.com/mweagle/Sparta"
	spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
	spartaAWSEvents "github.com/mweagle/Sparta/aws/events"
)

type storeResponse struct {
	Store   Store
	Message string
	Request spartaAWSEvents.APIGatewayRequest
}

type customersResponse struct {
	Customers []customer.Customer
	Message   string
	Request   spartaAWSEvents.APIGatewayRequest
}

type moviesResponse struct {
	Movies  []StoreMovie
	Message string
	Request spartaAWSEvents.APIGatewayRequest
}

func getStore(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// Get the store from DynamoDB
	sto := Store{}
	msg := "no path param phone"
	if phone, ok := gatewayEvent.PathParams["phone"]; ok {
		var err error
		msg = "getStore phone:" + phone
		if loggerOk {
			logger.Info(msg)
		}
		sto, err = Get(phone)
		if err != nil {
			msg = "store.Get() -> " + err.Error()
			if loggerOk {
				logger.Errorf(msg)
			}
		}
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, &storeResponse{
		Store:   sto,
		Message: msg,
		Request: gatewayEvent,
	}), nil
}

func getCustomers(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// Get the store's customers from DynamoDB
	cus := []customer.Customer{}
	msg := "no path param phone"
	if phone, ok := gatewayEvent.PathParams["phone"]; ok {
		msg = "getStore phone:" + phone
		if loggerOk {
			logger.Info(msg)
		}
		if sto, err := Get(phone); err == nil {
			cus, err = sto.Customers()
			if err != nil {
				msg = "store.Customers() -> " + err.Error()
				if loggerOk {
					logger.Errorf(msg)
				}
			}
		} else {
			msg = "store.Get() -> " + err.Error()
			if loggerOk {
				logger.Errorf(msg)
			}
		}
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, &customersResponse{
		Customers: cus,
		Message:   msg,
		Request:   gatewayEvent,
	}), nil
}

func getMovies(ctx context.Context,
	gatewayEvent spartaAWSEvents.APIGatewayRequest) (*spartaAPIGateway.Response, error) {

	logger, loggerOk := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// Get the store's movies from DynamoDB
	mov := []StoreMovie{}
	msg := "no path param phone"
	if phone, ok := gatewayEvent.PathParams["phone"]; ok {
		var err error
		year := 0
		title := ""
		msg = "getMovies phone:" + phone
		if ystr, ok := gatewayEvent.PathParams["year"]; ok {
			if year, err = strconv.Atoi(ystr); err == nil {
				msg = msg + " year:" + ystr
				if title, ok = gatewayEvent.PathParams["title"]; ok {
					msg = msg + " title:" + title
				}
			} else {
				msg = msg + fmt.Sprintf(" year %q - %s", ystr, err.Error())
			}
		}
		if loggerOk {
			logger.Info(msg)
		}
		if sto, err := Get(phone); err == nil {
			mov, err = sto.Movies(year, title)
			if err != nil {
				msg = "store.getMovies() -> " + err.Error()
				if loggerOk {
					logger.Errorf(msg)
				}
			}
		} else {
			msg = "store.Get() -> " + err.Error()
			if loggerOk {
				logger.Errorf(msg)
			}
		}
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, &moviesResponse{
		Movies:  mov,
		Message: msg,
		Request: gatewayEvent,
	}), nil
}

func LambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo

	// Lambda getStore
	lambdaFn, lambdaErr := sparta.NewAWSLambda(
		sparta.LambdaName(getStore),
		getStore,
		"Sparta-Lambda-DynamoDB",
	)
	if nil != lambdaErr {
		panic("Failed to create getStore lambda: " + lambdaErr.Error())
	}
	lambdaFunctions = append(lambdaFunctions, lambdaFn)
	if nil != api {
		apiPath := "/store/{phone}"
		apiGatewayResource, _ := api.NewResource(apiPath, lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create resource " + apiPath + " - " + apiMethodErr.Error())
		}
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path.phone"] = true
	}

	// Lambda getCustomers
	lambdaFn, lambdaErr = sparta.NewAWSLambda(
		sparta.LambdaName(getCustomers),
		getCustomers,
		"Sparta-Lambda-DynamoDB",
	)
	if nil != lambdaErr {
		panic("Failed to create getCustomers lambda: " + lambdaErr.Error())
	}
	lambdaFunctions = append(lambdaFunctions, lambdaFn)
	if nil != api {
		apiPath := "/store/{phone}/customer"
		apiGatewayResource, _ := api.NewResource(apiPath, lambdaFn)

		// We only return http.StatusOK
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create resource " + apiPath + " - " + apiMethodErr.Error())
		}
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path.phone"] = true
	}

	// Lambda getMovies
	lambdaFn, lambdaErr = sparta.NewAWSLambda(
		sparta.LambdaName(getMovies),
		getMovies,
		"Sparta-Lambda-DynamoDB",
	)
	if nil != lambdaErr {
		panic("Failed to create getMovies lambda: " + lambdaErr.Error())
	}
	lambdaFunctions = append(lambdaFunctions, lambdaFn)
	if nil != api {
		apiPath := "/store/{phone}/movie"
		apiGatewayResource, _ := api.NewResource(apiPath, lambdaFn)
		apiMethod, apiMethodErr := apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create resource " + apiPath + " - " + apiMethodErr.Error())
		}
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path.phone"] = true

		apiPath = "/store/{phone}/movie/{year}"
		apiGatewayResource, _ = api.NewResource(apiPath, lambdaFn)
		apiMethod, apiMethodErr = apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create resource " + apiPath + " - " + apiMethodErr.Error())
		}
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path.phone"] = true
		apiMethod.Parameters["method.request.path.year"] = true

		apiPath = "/store/{phone}/movie/{year}/{title}"
		apiGatewayResource, _ = api.NewResource(apiPath, lambdaFn)
		apiMethod, apiMethodErr = apiGatewayResource.NewMethod("GET",
			http.StatusOK,
			http.StatusInternalServerError)
		if nil != apiMethodErr {
			panic("Failed to create resource " + apiPath + " - " + apiMethodErr.Error())
		}
		apiMethod.SupportedRequestContentTypes = []string{"application/json"}
		apiMethod.Parameters["method.request.path.phone"] = true
		apiMethod.Parameters["method.request.path.year"] = true
		apiMethod.Parameters["method.request.path.title"] = true
	}

	return lambdaFunctions
}
