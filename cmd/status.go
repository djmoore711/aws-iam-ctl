package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/yourusername/iamctl/internal/aws"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current IAM identity information",
		Long: `Display information about the current IAM identity including:
- User ARN
- User ID
- Account ID
- MFA status
- Credentials expiration (if temporary)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Get profile from flags or environment
			profile, _ := cmd.Flags().GetString("profile")
			
			// 2. Get output format
			output, _ := cmd.Flags().GetString("output")

			// 3. Create context with timeout (matches Phase 2 pattern)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// 4. Create IAM client (uses Pattern from Phase 2)
			client, err := aws.NewIAMClient(profile)
			if err != nil {
				// 5. Proper error classification (matches Phase 2)
				return handleAWSErrors(err)
			}

			// 6. Get current user information
			user, err := aws.GetCurrentUser(ctx, client)
			if err != nil {
				return handleAWSErrors(err)
			}

			// 7. Prepare user information
			userName := *user.UserName
			arn := *user.Arn
			accountID := extractAccountID(*user.Arn)
			
			// 8. Check MFA status (security feature)
			mfaStatus := "disabled"
			if isMFAEnabled(ctx, client, user) {
				mfaStatus = "enabled"
			}

			// 9. Output based on format
			if output == "csv" {
				// Write CSV to stdout
				writer := csv.NewWriter(os.Stdout)
				defer writer.Flush()
				
				// Write header
				header := []string{"User", "ARN", "AccountID", "MFA"}
				if err := writer.Write(header); err != nil {
					return fmt.Errorf("failed to write CSV header: %w", err)
				}
				
				// Write data
				record := []string{userName, arn, accountID, mfaStatus}
				if err := writer.Write(record); err != nil {
					return fmt.Errorf("failed to write CSV record: %w", err)
				}
			} else {
				// Default text output (security-conscious formatting)
				fmt.Printf("User: %s\n", userName)
				fmt.Printf("ARN: %s\n", arn)
				fmt.Printf("Account ID: %s\n", accountID)
				fmt.Printf("MFA: %s\n", mfaStatus)
			}

			return nil
		},
	}

	// Add profile flag matching AWS CLI behavior
	cmd.Flags().StringP("profile", "p", "", "Use a specific profile from your credential file")
	
	// Add output flag
	cmd.Flags().StringP("output", "o", "", "Output format (csv)")

	return cmd
}

// handleAWSErrors converts SDK errors to user-friendly messages without leaks
func handleAWSErrors(err error) error {
	switch err.(type) {
	case *aws.CredentialError:
		return fmt.Errorf("credential error: check your AWS credentials configuration")
	case *aws.PermissionError:
		return fmt.Errorf("permission error: you don't have sufficient permissions to get user information")
	default:
		return fmt.Errorf("AWS service error: cannot connect to IAM service")
	}
}

// Helper function to extract account ID from ARN
func extractAccountID(arn string) string {
	// arn:aws:iam::123456789012:user/username
	parts := strings.Split(arn, ":")
	if len(parts) >= 5 {
		return parts[4]
	}
	return "unknown"
}

// Check if MFA is enabled for the current user
func isMFAEnabled(ctx context.Context, client *iam.Client, user *types.User) bool {
	// Implementation would check MFA devices
	// This is a placeholder for actual implementation
	return false
}