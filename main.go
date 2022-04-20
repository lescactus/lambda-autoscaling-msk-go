package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/slack-go/slack"
)

const (
	envSlackChannel              = "SLACK_CHANNEL"
	envAWSSecretsManagerName     = "AWS_SECRETS_MANAGER_NAME"
	jsonAWSSecretsManagerKeyName = "slackToken"

	failedScalingStatus  = "FAILED"
	successScalingStatus = "TRIGGERED"
)

var (
	slackToken            string // Slack token. Stored in AWS Secrets Manager
	slackChannel          string // Name of the Slack channel where to send the alert
	awsSecretsManagerName string // AWS Secrets Manager name where the Slack token is stored into
)

func handleRequest(ctx context.Context, autoScalingEvent events.AutoScalingEvent) (string, error) {
	// '.detail.eventName = UpdateBrokerStorage' is the event related to MSK storage scaling
	if autoScalingEvent.Detail["eventName"] != "UpdateBrokerStorage" {
		log.Printf("eventName is not 'UpdateBrokerStorage' but %s, abort...", autoScalingEvent.Detail["eventName"])
		return "", nil
	}

	// Dump event to stdout for logging purposes
	eventJson, _ := json.Marshal(autoScalingEvent)
	log.Printf("EVENT: %s", eventJson)

	// Parse JSON event
	event, err := gabs.ParseJSON(eventJson)
	if err != nil {
		return "", fmt.Errorf("error parsing json event: %w", err)
	}

	// The cluster arn is URL encoded in the event payload
	// It must be decoded first
	clusterArn := ""
	clusterArnPath, ok := event.Path("detail.requestParameters.clusterArn").Data().(string)
	if ok {
		log.Printf("Found detail.requestParameters.clusterArn: %s", clusterArnPath)
		clusterArn, err = url.QueryUnescape(clusterArnPath)
		if err != nil {
			fmt.Errorf("error url decoding %s: %w", clusterArnPath, err)
			clusterArn = clusterArnPath
		}
	}

	// Get the cluster resource from the arn
	clusterResource := ""
	a, err := arn.Parse(clusterArn)
	if err != nil || !arn.IsARN(clusterArn) {
		fmt.Errorf("error parsing ARN of %s: %w", clusterArn, err)
		clusterResource = clusterArn
	}
	clusterResource = a.Resource

	// The cluster resource in the event is formatted like so:
	//
	// cluster/<name>/<id>
	//
	clusterName := strings.Split(clusterResource, "/")[1]

	responseMessage := ""
	if responseMessagePath, ok := event.Path("detail.responseElements.message").Data().(string); ok {
		log.Printf("Found detail.responseElements.message: %s", responseMessagePath)
		responseMessage = responseMessagePath
	}

	targetBrokerEBSVolumeInfoStr := event.Path("detail.requestParameters.targetBrokerEBSVolumeInfo").String()
	log.Printf("Found detail.requestParameters.targetBrokerEBSVolumeInfo: %s", targetBrokerEBSVolumeInfoStr)

	// The event has a field ".detail.errorCode" when an autoscaling error occured
	//
	// {
	// 	"version": "0",
	// 	"source": "aws.kafka",
	//  ...
	// 	"detail": {
	// 		"awsRegion": "eu-central-1",
	// 		"errorCode": "BadRequestException",
	// 		"eventName": "UpdateBrokerStorage",
	// 		"eventSource": "kafka.amazonaws.com",
	//      ...
	//   }
	// }
	var status string
	if event.Exists("detail", "errorCode") {
		status = failedScalingStatus
		log.Printf("Found detail.errorCode")
	} else {
		status = successScalingStatus
		log.Printf("Didn't found detail.errorCode")
	}

	// Instanciate a new aws session
	awssession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Instanciate a new SecretsManager client with an aws session
	svc := secretsmanager.New(awssession)

	// Retrieve the Slack token stored in AWS Secrets Manager
	slackToken, err := GetSlackToken(svc, awsSecretsManagerName, jsonAWSSecretsManagerKeyName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve slack token from AWS Secrets Manager: %w", err)
	}

	// New slack client and message payload
	client := slack.New(slackToken)
	payload := slackMessage{
		Title:                     ":rotating_light: AWS MSK Storage Autoscaling triggered :rotating_light:\n",
		AWSAccountID:              autoScalingEvent.AccountID,
		AWSRegion:                 autoScalingEvent.Region,
		MSKClusterName:            clusterName,
		MSKClusterArn:             clusterArn,
		MSKConsoleLink:            fmt.Sprintf("https://%s.console.aws.amazon.com/msk/home?region=%s#/cluster/%s/view?tabId=details", os.Getenv("AWS_REGION"), os.Getenv("AWS_REGION"), clusterArnPath),
		ResponseMessage:           responseMessage,
		Source:                    autoScalingEvent.Source,
		DetailType:                autoScalingEvent.DetailType,
		TargetBrokerEBSVolumeInfo: targetBrokerEBSVolumeInfoStr,
		Time:                      autoScalingEvent.Time,
		Status:                    status,
	}

	_, _, err = client.PostMessage(
		slackChannel,
		slack.MsgOptionAttachments(payload.FormatAttachment()),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Fatalln(err)
		return "", nil
	}
	log.Printf("Message successfully sent to channel %s", slackChannel)

	return fmt.Sprintf("Message successfully sent to channel %s", slackChannel), nil
}

func main() {
	if slackChannel = os.Getenv(envSlackChannel); slackChannel == "" {
		log.Fatalln("Environment variable " + envSlackChannel + " is not set but is required! Exiting!")
	}
	if awsSecretsManagerName = os.Getenv(envAWSSecretsManagerName); awsSecretsManagerName == "" {
		log.Fatalln("Environment variable " + envAWSSecretsManagerName + " is not set but is required! Exiting!")
	}

	runtime.Start(handleRequest)
}
