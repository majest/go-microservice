package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	service "github.com/majest/go-service-test/service"
	"golang.org/x/net/context"
)

func DecodeRequest(r *http.Request) (interface{}, error) {
	var request service.CountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func EncodeResponse(w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func MakeEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		svc.Handle(request)
		response := svc.GetResponse()
		fmt.Printf("Request %v: \n", response)
		return response, nil
	}
}
