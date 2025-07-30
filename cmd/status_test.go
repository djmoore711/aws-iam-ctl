package cmd

import (
	"testing"

	"github.com/yourusername/iamctl/internal/aws"
	"github.com/stretchr/testify/assert"
)

func TestExtractAccountID(t *testing.T) {
	tests := []struct {
		name     string
		arn      string
		expected string
	}{
		{
			name:     "valid ARN",
			arn:      "arn:aws:iam::123456789012:user/test-user",
			expected: "123456789012",
		},
		{
			name:     "invalid ARN",
			arn:      "invalid-arn",
			expected: "unknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractAccountID(tc.arn)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHandleAWSErrors(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		errorContains string
	}{
		{
			name:          "credential error",
			err:           &aws.CredentialError{Err: nil},
			errorContains: "credential error",
		},
		{
			name:          "permission error",
			err:           &aws.PermissionError{Err: nil},
			errorContains: "permission error",
		},
		{
			name:          "service error",
			err:           &aws.ServiceError{Err: nil},
			errorContains: "AWS service error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := handleAWSErrors(tc.err)
			assert.Contains(t, result.Error(), tc.errorContains)
		})
	}
}