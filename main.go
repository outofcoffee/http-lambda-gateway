package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"lambdahttpgw/config"
	"lambdahttpgw/stats"
	"net/http"
	"strings"
	"time"
)

var (
	region          = config.GetRegion()
	requestIdHeader = config.GetRequestIdHeader()
	version         = "dev"
)

func main() {
	logrus.SetLevel(config.GetConfigLevel())
	stats.Init()

	http.HandleFunc("/system/stats", statsHandler)
	http.HandleFunc("/system/status", statusHandler)
	http.HandleFunc("/", handler)

	port := config.GetPort()
	logrus.Infof("starting http lambda gateway %v for region %v on port %v", version, region, port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func statusHandler(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "ok\n")
}

func statsHandler(w http.ResponseWriter, req *http.Request) {
	logrus.Debugf("fetching stats")
	statsJson, err := json.Marshal(stats.GetAllStats())
	if err != nil {
		logrus.Errorf("error marshalling stats to JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, "%s\n", statsJson)
}

func handler(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	log := logrus.WithField("requestId", getRequestId(requestIdHeader, req))

	client := req.RemoteAddr
	log.Debugf("received request %v %v from client %v", req.Method, req.URL, client)

	functionName, path, requestHeaders, requestBody, err := parseRequest(req)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code, body, responseHeaders, err := invoke(log, functionName, req.Method, path, requestHeaders, requestBody)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = sendResponse(log, w, responseHeaders, code, body, client)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(startTime)
	log.Infof("proxied request to %v [code: %v, body %v bytes] for client %v in %v", functionName, code, len(body), client, elapsed)
	stats.RecordHit(functionName)
}

func getRequestId(headerName string, req *http.Request) string {
	var requestId string
	if headerName != "" {
		requestId = req.Header.Get(headerName)
	}
	if requestId == "" {
		requestId = uuid.NewString()
	}
	return requestId
}

func parseRequest(req *http.Request) (string, string, map[string]string, []byte, error) {
	splitPath := strings.SplitN(strings.TrimPrefix(req.URL.Path, "/"), "/", 2)

	var functionName string
	path := "/"
	if len(splitPath) >= 1 && splitPath[0] != "" {
		functionName = splitPath[0]
		if len(splitPath) >= 2 {
			path = "/" + splitPath[1]
		}
	} else {
		return "", "", nil, nil, fmt.Errorf("path must include function name and request path")
	}

	requestHeaders := make(map[string]string)
	for requestHeaderKey, requestHeaderValue := range req.Header {
		requestHeaders[requestHeaderKey] = requestHeaderValue[0]
	}

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", "", map[string]string{}, nil, fmt.Errorf("error parsing request body: %v", err)
	}
	return functionName, path, requestHeaders, requestBody, err
}

func invoke(
	log *logrus.Entry,
	functionName string,
	httpMethod string,
	path string,
	requestHeaders map[string]string,
	requestBody []byte,
) (statusCode int, body []byte, responseHeaders map[string]string, err error) {
	log.Debugf("invoking function %v with %v %v [body: %v bytes]", functionName, httpMethod, path, len(requestBody))

	// Create Lambda service client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := lambda.New(sess, &aws.Config{Region: aws.String(region)})

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

	log.Debugf("received response from function %v [code: %v, body: %v bytes]", functionName, resp.StatusCode, len(responseBody))
	return resp.StatusCode, responseBody, resp.Headers, nil
}

func sendResponse(log *logrus.Entry, w http.ResponseWriter, responseHeaders map[string]string, statusCode int, body []byte, client string) (err error) {
	for responseHeaderKey, responseHeaderValue := range responseHeaders {
		w.Header().Add(responseHeaderKey, responseHeaderValue)
	}
	w.WriteHeader(statusCode)
	_, err = w.Write(body)
	if err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}

	log.Debugf("wrote response [code: %v, body %v bytes] to client %v", statusCode, len(body), client)
	return nil
}
