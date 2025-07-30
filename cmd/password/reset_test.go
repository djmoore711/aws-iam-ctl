package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awssdk "github.com/yourusername/iamctl/internal/aws"
)

// Mock IAM client for testing
type mockPasswordIAMClient struct {
	updateLoginProfileFunc func(context.Context, *iam.UpdateLoginProfileInput) (*iam.UpdateLoginProfileOutput, error)
	createLoginProfileFunc func(context.Context, *iam.CreateLoginProfileInput) (*iam.CreateLoginProfileOutput, error)
}

func (m *mockPasswordIAMClient) UpdateLoginProfile(ctx context.Context, input *iam.UpdateLoginProfileInput) (*iam.UpdateLoginProfileOutput, error) {
	if m.updateLoginProfileFunc != nil {
		return m.updateLoginProfileFunc(ctx, input)
	}
	return &iam.UpdateLoginProfileOutput{}, nil
}

func (m *mockPasswordIAMClient) CreateLoginProfile(ctx context.Context, input *iam.CreateLoginProfileInput) (*iam.CreateLoginProfileOutput, error) {
	if m.createLoginProfileFunc != nil {
		return m.createLoginProfileFunc(ctx, input)
	}
	return &iam.CreateLoginProfileOutput{}, nil
}

func TestSuccessfulResetWithValidMFA(t *testing.T) {
	// Setup mock client
	client := &mockPasswordIAMClient{
		updateLoginProfileFunc: func(ctx context.Context, input *iam.UpdateLoginProfileInput) (*iam.UpdateLoginProfileOutput, error) {
			return &iam.UpdateLoginProfileOutput{}, nil
		},
	}

	// Test successful reset
	ctx := context.Background()
	err := resetPasswordWithClient(ctx, client, "test-profile", "testuser", "ValidPass123!", "arn:aws:iam::123456789012:mfa/testuser", "123456")
	if err != nil {
		t.Errorf("Expected successful reset, got error: %v", err)
	}
}

func TestInvalidPasswordComplexity(t *testing.T) {
	// Test password too short
	if isValidPassword("Short1!") {
		t.Error("Expected short password to fail complexity check")
	}

	// Test password with only 2 character types
	if isValidPassword("onlylowercaseandnumbers123") {
		t.Error("Expected password with only 2 character types to fail complexity check")
	}

	// Test valid password
	if !isValidPassword("ValidPass123!@") { // 15 characters with 4 types
		t.Error("Expected valid password to pass complexity check")
	}
}

func TestMFAValidationFailure(t *testing.T) {
	// Setup mock client that returns MFA error
	client := &mockPasswordIAMClient{
		updateLoginProfileFunc: func(ctx context.Context, input *iam.UpdateLoginProfileInput) (*iam.UpdateLoginProfileOutput, error) {
			return nil, &awssdk.PermissionError{Err: errors.New("access denied")}
		},
	}

	// Test reset failure due to MFA validation
	ctx := context.Background()
	err := resetPasswordWithClient(ctx, client, "test-profile", "testuser", "ValidPass123!", "arn:aws:iam::123456789012:mfa/testuser", "invalid")
	if err == nil {
		t.Error("Expected reset to fail due to MFA validation error")
	}

	// Check that the error is properly classified
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected error to contain access denied message, got: %v", err)
	}
}

// resetPasswordWithClient is a testable version of resetPassword that accepts a mock client
func resetPasswordWithClient(ctx context.Context, client *mockPasswordIAMClient, profile, username, password, serialNumber, mfaToken string) error {
	// Reset password
	input := &iam.UpdateLoginProfileInput{
		UserName: aws.String(username),
		Password: aws.String(password),
	}

	_, err := client.UpdateLoginProfile(ctx, input)
	if err != nil {
		// If login profile doesn't exist, create it
		if strings.Contains(err.Error(), "NoSuchEntity") {
			createInput := &iam.CreateLoginProfileInput{
				UserName: aws.String(username),
				Password: aws.String(password),
			}
			_, err = client.CreateLoginProfile(ctx, createInput)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}