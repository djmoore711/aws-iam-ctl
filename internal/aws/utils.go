package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
)

// GetCurrentUser retrieves the current IAM user with proper error classification
func GetCurrentUser(ctx context.Context, client *iam.Client) (*types.User, error) {
	// Create the request
	input := &iam.GetUserInput{}

	// Execute request with timeout
	result, err := client.GetUser(ctx, input)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "UnrecognizedClientException", "InvalidClientTokenId":
				return nil, &CredentialError{Err: err}
			case "AccessDeniedException":
				return nil, &PermissionError{Err: err}
			}
		}
		return nil, &ServiceError{Err: err}
	}

	return result.User, nil
}

// Custom error types for proper classification
type CredentialError struct{ Err error }
func (e *CredentialError) Error() string { return "credential error: " + e.Err.Error() }

type PermissionError struct{ Err error }
func (e *PermissionError) Error() string { return "permission error: " + e.Err.Error() }

type ServiceError struct{ Err error }
func (e *ServiceError) Error() string { return "service error: " + e.Err.Error() }