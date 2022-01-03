package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	splitPath := strings.SplitN(strings.TrimPrefix(req.URL.Path, "/"), "/", 2)
	functionName := splitPath[0]
	path := "/" + splitPath[1]
	requestHeaders := map[string]string{}

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	// invoke function
	code, body, responseHeaders, err := invoke(functionName, req.Method, path, requestHeaders, requestBody)
	if err != nil {
		panic(err)
	}

	for responseHeaderKey, responseHeaderValue := range responseHeaders {
		w.Header().Add(responseHeaderKey, responseHeaderValue)
	}
	w.WriteHeader(code)
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

func invoke(
	functionName string,
	httpMethod string,
	path string,
	requestHeaders map[string]string,
	requestBody []byte,
) (statusCode int, body []byte, responseHeaders map[string]string, err error) {

	// Create Lambda service client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := lambda.New(sess, &aws.Config{Region: aws.String("eu-west-1")})

	encodedBody := b64.StdEncoding.EncodeToString(requestBody)
	request := events.APIGatewayProxyRequest{
		HTTPMethod:      httpMethod,
		Path:            path,
		Headers:         requestHeaders,
		Body:            encodedBody,
		IsBase64Encoded: true,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("error marshalling request: %v", err)
	}

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: payload})
	if err != nil {
		return 0, nil, nil, fmt.Errorf("error calling %v: %v", functionName, err)
	}

	var resp events.APIGatewayProxyResponse

	err = json.Unmarshal(result.Payload, &resp)
	if err != nil || resp.StatusCode == 0 {
		return 0, nil, nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	var responseBody []byte
	if resp.IsBase64Encoded {
		responseBody, err = b64.StdEncoding.DecodeString(resp.Body)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("error decoding body %v: %v", resp.Body, err)
		}
	} else {
		responseBody = []byte(resp.Body)
	}

	return resp.StatusCode, responseBody, resp.Headers, nil
}
