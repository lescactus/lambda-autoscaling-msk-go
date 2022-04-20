package main

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// Represents the info we are interested in
type slackMessage struct {
	Title                     string
	AWSAccountID              string
	AWSRegion                 string
	MSKClusterArn             string
	MSKClusterName            string
	MSKConsoleLink            string
	Source                    string // events.Source
	DetailType                string // events.DetailType
	TargetBrokerEBSVolumeInfo string
	ResponseMessage           string
	Time                      time.Time // events.Time
	Status                    string
}

// FormatAttachment will create a Slack attachment to wrap all the alarm info we want in Slack
// Slack attachments supports formatting as per https://api.slack.com/docs/formatting
// It returns a slack.Attachment with the correct formatting ready to be sent to Slack
func (s *slackMessage) FormatAttachment() slack.Attachment {

	return slack.Attachment{
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
}
