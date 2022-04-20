# lambda-autoscaling-msk

This repository contains the code for a lambda function to send MSK storage autoscaling events to Slack

## Installation

### Build the lambda

#### From source with go

You need a working [go](https://golang.org/doc/install) toolchain (It has been developped and tested with go 1.16 only, but should work with go >= 1.11 ). Refer to the official documentation for more information (or from your Linux/Mac/Windows distribution documentation to install it from your favorite package manager).

```sh
# Clone this repository
git clone https://github.com/lescactus/lambda-autoscaling-msk-go.git && cd lambda-autoscaling-msk-go

# Build from sources. Use the '-o' flag to change the compiled binary name
GOOS=linux go build -o main

# Zip the binary to upload it to Lambda
zip main.zip main
```

### Update the lambda

Once the lambda has been compiled and zipped, it need to be uploaded:

```sh
aws lambda update-function-code \
  --function-name lambda-autoscaling-msk  \
  --zip-file fileb://./function.zip
```

### AWS Secrets Manager

`lambda-autoscaling-msk-go` will lookup the Slack token from AWS Secrets Manager. Before using, you need to create a secret with the following value:

```json
{
  "slackToken": "<slack token>"
}
```

### IAM Permissions

`lambda-autoscaling-msk-go` require the following IAM permissions:

* `AWSLambdaBasicExecutionRole` role

* `secretsmanager:GetSecretValue`

* `kms:Decrypt`

## Configuration

`lambda-autoscaling-msk-go` is looking for the following environment variables:

| Name     | Type | Default value    | Description |
| --------|---------|---------|-------|
| `SLACK_CHANNEL`  | `string` | `""`   | Name of the slack channel to send the notification into   |
| `AWS_SECRETS_MANAGER_NAME` | `string` |`""` | Name of the AWS Secrets Manager Secret where `lambda-autoscaling-msk-go` will look for the slack token  |

## Testing

### Unit tests
To run the test suite, run the following commands:

```sh
# Run the unit tests. Remove the '-v' flag to reduce verbosity
go test -v ./... 

# Get coverage to html format
go test -coverprofile -v /tmp/cover.out ./...
go tool cover -html=/tmp/cover.out -o /tmp/cover.out.html
```

### Testing the lambda

There is two json files in the `tests/` folder:

* `event-failure.json`

* `event-success.json`

They can be used as a test event to make an e2e test of the lambda.