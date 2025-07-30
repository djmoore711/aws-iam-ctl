package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	awssdk "github.com/yourusername/iamctl/internal/aws"
	"github.com/spf13/cobra"
)

// NewDisableCommand creates the disable command
func NewDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable IAM access key",
		Long:  `Instantly disable an IAM access key by its ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile from flags or environment
			profile, _ := cmd.Flags().GetString("profile")
			keyID, _ := cmd.Flags().GetString("key-id")
			username, _ := cmd.Flags().GetString("username")

			// Validate required parameters
			if keyID == "" {
				return fmt.Errorf("key-id is required")
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Create IAM client
			client, err := awssdk.NewIAMClient(profile)
			if err != nil {
				return handleDisableAWSErrors(err)
			}

			// Disable the access key
			err = disableKey(ctx, client, keyID, username)
			if err != nil {
				return fmt.Errorf("❌ Disable failed: %v", sanitizeError(err))
			}

			fmt.Println("✅ Key disabled successfully")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("profile", "p", "", "Use a specific profile from your credential file")
	cmd.Flags().String("key-id", "", "ID of the access key to disable (required)")
	cmd.Flags().String("username", "", "Username of the key owner (defaults to current user)")

	return cmd
}

// disableKey disables an access key
func disableKey(ctx context.Context, client *iam.Client, keyID, username string) error {
	input := &iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(keyID),
		Status:      types.StatusTypeInactive,
	}

	// Add username if provided
	if username != "" {
		input.UserName = aws.String(username)
	}

	_, err := client.UpdateAccessKey(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to disable access key: %w", err)
	}

	return nil
}

// handleDisableErrors converts SDK errors to user-friendly messages
func handleDisableAWSErrors(err error) error {
	switch err.(type) {
	case *awssdk.CredentialError:
		return fmt.Errorf("credential error: check your AWS credentials configuration")
	case *awssdk.PermissionError:
		return fmt.Errorf("permission error: you don't have sufficient permissions to disable keys")
	default:
		return fmt.Errorf("AWS service error: cannot connect to IAM service")
	}
}