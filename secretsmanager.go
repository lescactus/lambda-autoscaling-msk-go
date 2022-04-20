package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/tidwall/gjson"
)

// GetSlackToken will lookup the AWS Secrets Manager json value to find the Slack token
//
// Minimum IAM permission:
//
// * secretsmanager:GetSecretValue
//
// * kms:Decrypt
//
//
// It returns the Slack token as a string or any error encountered
func GetSlackToken(svc secretsmanageriface.SecretsManagerAPI, secretName, keyName string) (string, error) {
	// Prepare input to retrieve secret's value string
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	// Retrieve secret's value
	output, err := svc.GetSecretValue(input)
	if err != nil {
		return "", err
	}

	// Lookup if the key already exists.
	// gjson.Get() return the value if the key exists
	jsonValue := gjson.Get(*output.SecretString, keyName)
	if jsonValue.String() == "" {
		return "", fmt.Errorf("error finding the Slack token value in AWS Secrets Manager. Key '%s' doesn't exists", keyName)
	}
	return jsonValue.String(), nil
}
