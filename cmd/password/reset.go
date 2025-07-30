package cmd

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awssdk "github.com/yourusername/iamctl/internal/aws"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewResetCommand creates the password reset command
func NewResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset IAM user password",
		Long:  `Reset IAM user password with MFA verification.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile from flags or environment
			profile, _ := cmd.Flags().GetString("profile")
			username, _ := cmd.Flags().GetString("username")
			serialNumber, _ := cmd.Flags().GetString("mfa-serial")

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Create IAM client
			client, err := awssdk.NewIAMClient(profile)
			if err != nil {
				return handlePasswordResetErrors(err)
			}

			// Get current user if username not provided
			if username == "" {
				user, err := awssdk.GetCurrentUser(ctx, client)
				if err != nil {
					return handlePasswordResetErrors(err)
				}
				username = *user.UserName
			}

			// Get MFA token first (security critical order)
			mfaToken, err := getMFAToken()
			if err != nil {
				return fmt.Errorf("❌ Reset failed: failed to read MFA token")
			}

			// Validate MFA first (security critical)
			if err := validateMFA(ctx, client, profile, username, serialNumber, mfaToken); err != nil {
				return fmt.Errorf("❌ Reset failed: Invalid credentials")
			}

			// Get password securely after MFA validation
			password, err := getPassword()
			if err != nil {
				return fmt.Errorf("❌ Reset failed: failed to read password")
			}

			// Clear password from memory when done
			passwordBytes := []byte(password)
			defer func() {
				for i := range passwordBytes {
					passwordBytes[i] = 0
				}
			}()

			// Validate password complexity
			if !isValidPassword(password) {
				return fmt.Errorf("❌ Reset failed: password does not meet complexity requirements (14+ chars, 3/4 character types)")
			}

			// Reset password
			err = resetPassword(ctx, client, profile, username, password, serialNumber, mfaToken)
			if err != nil {
				return fmt.Errorf("❌ Reset failed: Invalid credentials")
			}

			fmt.Println("✅ Password reset complete")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("profile", "p", "", "Use a specific profile from your credential file")
	cmd.Flags().String("username", "", "Username to reset password for (defaults to current user)")
	cmd.Flags().String("mfa-serial", "", "MFA device serial number (required)")

	return cmd
}

// validateMFA validates the MFA token before proceeding with password reset
func validateMFA(ctx context.Context, client *iam.Client, profile, username, serialNumber, mfaToken string) error {
	// Create MFA-enabled client to validate MFA token
	_, err := awssdk.GetMFAEnabledClient(ctx, profile, "", serialNumber, mfaToken)
	if err != nil {
		return fmt.Errorf("MFA validation failed: %w", err)
	}
	return nil
}

// resetPassword resets the user's password with MFA verification
func resetPassword(ctx context.Context, client *iam.Client, profile, username, password, serialNumber, mfaToken string) error {
	// Create MFA-enabled client
	mfaClient, err := awssdk.GetMFAEnabledClient(ctx, profile, "", serialNumber, mfaToken)
	if err != nil {
		return fmt.Errorf("failed to create MFA-enabled client: %w", err)
	}

	// Reset password
	input := &iam.UpdateLoginProfileInput{
		UserName: aws.String(username),
		Password: aws.String(password),
	}

	_, err = mfaClient.UpdateLoginProfile(ctx, input)
	if err != nil {
		// If login profile doesn't exist, create it
		if strings.Contains(err.Error(), "NoSuchEntity") {
			createInput := &iam.CreateLoginProfileInput{
				UserName: aws.String(username),
				Password: aws.String(password),
			}
			_, err = mfaClient.CreateLoginProfile(ctx, createInput)
			if err != nil {
				return fmt.Errorf("failed to create login profile: %w", err)
			}
		} else {
			return fmt.Errorf("failed to update login profile: %w", err)
		}
	}

	return nil
}

// getPassword securely reads a password from the terminal
func getPassword() (string, error) {
	fmt.Print("Enter new password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after password input
	if err != nil {
		return "", err
	}

	fmt.Print("Confirm new password: ")
	confirm, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after password input
	if err != nil {
		return "", err
	}

	if string(password) != string(confirm) {
		return "", fmt.Errorf("passwords do not match")
	}

	return string(password), nil
}

// getMFAToken securely reads an MFA token from the terminal
func getMFAToken() (string, error) {
	fmt.Print("Enter MFA token: ")
	token, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add a newline after token input
	if err != nil {
		return "", err
	}
	return string(token), nil
}

// isValidPassword checks password complexity requirements
func isValidPassword(password string) bool {
	if len(password) < 14 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	charTypes := 0
	if hasLower {
		charTypes++
	}
	if hasUpper {
		charTypes++
	}
	if hasDigit {
		charTypes++
	}
	if hasSpecial {
		charTypes++
	}

	return charTypes >= 3
}

// handlePasswordResetErrors converts SDK errors to user-friendly messages with unified error messaging
func handlePasswordResetErrors(err error) error {
	// Always return generic error message for security
	return fmt.Errorf("❌ Reset failed: Invalid credentials")
}