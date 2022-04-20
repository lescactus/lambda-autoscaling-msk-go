package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestFormatAttachment(t *testing.T) {
	s := slackMessage{
		Title:                     ":rotating_light: AWS MSK Storage Autoscaling triggered :rotating_light:\n",
		AWSAccountID:              "123456789012",
		AWSRegion:                 "eu-central-1",
		MSKClusterName:            "msk-cluster",
		MSKClusterArn:             "arn:aws:kafka:eu-central-1:123456789012:cluster/msk-cluster/a1adaf3f-a963-45c2-879f-c4bc09a97285-5",
		MSKConsoleLink:            fmt.Sprintf("https://%s.console.aws.amazon.com/msk/home?region=%s#/cluster/%s/view?tabId=details", os.Getenv("AWS_REGION"), os.Getenv("AWS_REGION"), "arn%3Aaws%3Akafka%3Aeu-central-1%123456789%3Acluster%2Fmsk%2F6e494b06-9803-410a-a237-ec5e7068c0ad-5"),
		ResponseMessage:           "--",
		Source:                    "aws.kafka",
		DetailType:                "AWS API Call via CloudTrail",
		TargetBrokerEBSVolumeInfo: "[{\"kafkaBrokerNodeId\":\"All\",\"volumeSizeGB\":39}]",
		Time:                      time.Now(),
		Status:                    successScalingStatus,
	}

	want := slack.Attachment{
		Color:     "danger",
		Title:     s.Title,
		TitleLink: s.MSKConsoleLink,
		Pretext:   ":fire: *MSK Storage Autoscaling - Cluster: " + s.MSKClusterName + "* :fire:",
		Text:      "Automatic alert\n",
		Fields: []slack.AttachmentField{
			{
				Title: "MSK Cluster",
				Value: "_" + s.MSKClusterName + "_",
				Short: true,
			},
			{
				Title: "Time",
				Value: "_" + s.Time.Format(time.RFC1123) + "_",
				Short: true,
			},
			{
				Title: "MSK Cluster ARN",
				Value: "_" + s.MSKClusterArn + "_",
				Short: false,
			},
			{
				Title: "Target Broker EBS Volume Info",
				Value: "_" + s.TargetBrokerEBSVolumeInfo + "_",
				Short: false,
			},
			{
				Title: "Scaling Status",
				Value: fmt.Sprintf("_%v_", s.Status),
			},
			{
				Title: "Message",
				Value: "_" + s.ResponseMessage + "_",
			},
			{
				Title: "Source",
				Value: "_" + s.Source + "_",
				Short: true,
			},
			{
				Title: "Detail Type",
				Value: "_" + s.DetailType + "_",
				Short: true,
			},
			{
				Title: "AWS Account ID",
				Value: "_" + s.AWSAccountID + "_",
				Short: true,
			},
			{
				Title: "AWS Region",
				Value: "_" + s.AWSRegion + "_",
				Short: true,
			},
		},
	}

	t.Run("Format Slack message attachment", func(t *testing.T) {
		attachment := s.FormatAttachment()
		assert.EqualValues(t, want, attachment)
	})
}
