# Lambda HTTP Gateway

Simple HTTP gateway for AWS Lambda. Acts as a lightweight API gateway for AWS Lambda functions.

## Run

You can download the binary for your platform from the [releases](https://github.com/outofcoffee/lambda-http-gateway/releases) page, or use the [Docker image](#docker-image).

> **Important:** Ensure the relevant AWS credentials are configured before run. The gateway uses the standard AWS mechanisms to authenticate/authorise with the AWS Lambda API, so the usual approaches of profiles/credentials apply.

### Binary

    ./lambdahttpgw

### Docker

    docker run -it -p 8090:8090 outofcoffee/lambdahttpgw

> Note: the home directory for the `gateway` user that runs the binary is `/opt/gateway`

## Call Lambda function

Call the Lambda function via the gateway:

    curl http://localhost:8090/MyLambdaName/some/path
    ...
    <Lambda HTTP response>

> Note the prefix of the Lambda function name (`MyLambdaName` above), before the path. The function receives the portion of the path without the function name, i.e. `/some/path` in this example.

The Lambda function receives events in the standard AWS API Gateway JSON format, and is expected to respond in kind.

## Configuration

Environment variables:

| Variable              | Meaning                                                                                        | Default     | Example               |
|-----------------------|------------------------------------------------------------------------------------------------|-------------|-----------------------|
| AWS_REGION            | AWS region in which to connect to Lambda functions.                                            | `eu-west-1` | `us-east-1`           |
| LOG_LEVEL             | Log level (trace, debug, info, warn, error).                                                   | `debug`     | `warn`                |
| PORT                  | Port on which to listen.                                                                       | `8090`      | `8080`                |
| REQUEST_ID_HEADER     | Name of request header to use as request ID for logging. If absent, a UUID will be used.       | Empty       | `x-correlation-id`    |
| STATS_REPORT_INTERVAL | The frequency with which stats should be reported, if enabled.                                 | `5s`        | `2m`                  |
| STATS_REPORT_URL      | URL to which stats should be reported. If not empty, hits are recorded for each function name. | Empty       | `https://example.com` |

## Build

Prerequisites:

- Go 1.17+

Steps:

    go build

## Docker image

Image on Docker Hub:

    outofcoffee/lambdahttpgw
