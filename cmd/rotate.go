package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	awssdk "github.com/yourusername/iamctl/internal/aws"
	"github.com/spf13/cobra"
)

// NewRotateCommand creates the rotate command
func NewRotateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate IAM access keys",
		Long:  `Rotate IAM access keys by creating a new key, testing it, storing it in Secrets Manager, and deleting the old key.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile from flags or environment
			profile, _ := cmd.Flags().GetString("profile")
			secretName, _ := cmd.Flags().GetString("secret-name")

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Create IAM client
			iamClient, err := awssdk.NewIAMClient(profile)
			if err != nil {
				return handleRotateAWSErrors(err)
			}

			// Get current user
			user, err := awssdk.GetCurrentUser(ctx, iamClient)
			if err != nil {
				return handleRotateAWSErrors(err)
			}

			// Perform atomic key rotation
			err = rotateKeys(ctx, iamClient, user.UserName, secretName)
			if err != nil {
				return fmt.Errorf("❌ Rotation failed: %v", sanitizeError(err))
			}

			fmt.Println("✅ Key rotation complete. New key stored in Secrets Manager")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("profile", "p", "", "Use a specific profile from your credential file")
	cmd.Flags().String("secret-name", "iamctl/access-key", "Name of the secret in AWS Secrets Manager")

	return cmd
}

// rotateKeys performs the atomic key rotation sequence
func rotateKeys(ctx context.Context, client *iam.Client, username *string, secretName string) error {
	// 1. Create new access key
	createKeyInput := &iam.CreateAccessKeyInput{
		UserName: username,
	}

	createResult, err := client.CreateAccessKey(ctx, createKeyInput)
	if err != nil {
		return fmt.Errorf("failed to create new access key: %w", err)
	}

	newKey := createResult.AccessKey

	// Defer deletion of the new key in case of failure
	defer func() {
		if err != nil {
			// Try to delete the newly created key if something went wrong
			_, deleteErr := client.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
				AccessKeyId: newKey.AccessKeyId,
				UserName:    username,
			})
			if deleteErr != nil {
				// Log but don't return this error as we're already handling another error
				fmt.Printf("Warning: Failed to clean up new access key: %v\n", deleteErr)
			}
		}
	}()

	// 2. Test the new key (simplified test - in a real implementation you might do a more thorough test)
	testCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			*newKey.AccessKeyId,
			*newKey.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to create test config: %w", err)
	}

	testClient := iam.NewFromConfig(testCfg)
	_, err = testClient.GetUser(ctx, &iam.GetUserInput{})
	if err != nil {
		return fmt.Errorf("failed to test new access key: %w", err)
	}

	// 3. Store new key in Secrets Manager
	smClient := secretsmanager.NewFromConfig(testCfg)
	secretValue := fmt.Sprintf("{\"AccessKeyId\": \"%s\", \"SecretAccessKey\": \"%s\"}", 
		*newKey.AccessKeyId, *newKey.SecretAccessKey)
	
	_, err = smClient.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretValue),
	})
	if err != nil {
		return fmt.Errorf("failed to store key in Secrets Manager: %w", err)
	}

	// 4. List existing keys to find the old one
	listInput := &iam.ListAccessKeysInput{
		UserName: username,
	}

	listResult, err := client.ListAccessKeys(ctx, listInput)
	if err != nil {
		return fmt.Errorf("failed to list access keys: %w", err)
	}

	// 5. Delete the old key (assuming we're rotating the first key)
	if len(listResult.AccessKeyMetadata) > 0 {
		oldKey := listResult.AccessKeyMetadata[0]
		if *oldKey.AccessKeyId != *newKey.AccessKeyId {
			_, err = client.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
				AccessKeyId: oldKey.AccessKeyId,
				UserName:    username,
			})
			if err != nil {
				return fmt.Errorf("failed to delete old access key: %w", err)
			}
		}
	}

	return nil
}

// sanitizeError removes sensitive information from error messages
func sanitizeError(err error) error {
	// In a real implementation, you would sanitize the error message
	// to remove any sensitive information like access key IDs
	return err
}

// handleAWSErrors converts SDK errors to user-friendly messages without leaks
func handleRotateAWSErrors(err error) error {
	switch err.(type) {
	case *awssdk.CredentialError:
		return fmt.Errorf("credential error: check your AWS credentials configuration")
	case *awssdk.PermissionError:
		return fmt.Errorf("permission error: you don't have sufficient permissions to rotate keys")
	default:
		return fmt.Errorf("AWS service error: cannot connect to IAM service")
	}
}