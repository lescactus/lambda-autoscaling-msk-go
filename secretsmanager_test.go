package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/stretchr/testify/assert"
)

// Define a mock struct to be used in unit tests
type mockSecretsManagerClient struct {
	secretsmanageriface.SecretsManagerAPI
}

var (
	mockSlackChannel                 = "MockSlackChannel"
	mockSlackToken                   = "ThisIsAMockSlackToken"
	mockSecretManagerName            = "moodagent/alerting/slack"
	mockJSONAWSSecretsManagerKeyName = "slackToken"
)

func init() {
	os.Setenv("SLACK_CHANNEL", mockSlackChannel)
	os.Setenv("AWS_SECRETS_MANAGER_NAME", mockSecretManagerName)
}

func (m *mockSecretsManagerClient) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	if *input.SecretId == mockSecretManagerName {
		// Secrets Manager store secret string in json
		str := fmt.Sprintf("{\"%s\":\"%s\"}", jsonAWSSecretsManagerKeyName, mockSlackToken)

		return &secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(str),
		}, nil
	}
	return nil, awserr.New(secretsmanager.ErrCodeResourceNotFoundException, "Secrets Manager can't find the specified secret.", errors.New(""))
}

func TestGetSlackToken(t *testing.T) {
	// Setup Test
	mockSvc := &mockSecretsManagerClient{}

	t.Run("AWS Secret exists - key exists", func(t *testing.T) {
		token, err := GetSlackToken(mockSvc, mockSecretManagerName, mockJSONAWSSecretsManagerKeyName)
		assert.NoError(t, err)
		assert.EqualValues(t, mockSlackToken, token)
	})

	t.Run("AWS Secrets exists - key doesn't exists", func(t *testing.T) {
		token, err := GetSlackToken(mockSvc, mockSecretManagerName, "InexistentKey")
		assert.Error(t, err)
		assert.EqualValues(t, "", token)
	})

	t.Run("AWS Secret doesn't exists", func(t *testing.T) {
		token, err := GetSlackToken(mockSvc, "InexistentSecret", "InexistentKey")
		assert.Error(t, err)
		assert.EqualValues(t, "", token)
	})
}
