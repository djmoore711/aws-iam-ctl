package mfa

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awssdk "github.com/yourusername/iamctl/internal/aws"
	"github.com/spf13/cobra"
)

// NewDisableCommand creates the MFA disable command
func NewDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable MFA for the current user",
		Long:  `Disable MFA by deleting the VirtualMFADevice with double-confirmation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile from flags or environment
			profile, _ := cmd.Flags().GetString("profile")

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// Create IAM client
			client, err := awssdk.NewIAMClient(profile)
			if err != nil {
				return handleMFAErrors(err)
			}

			// Get current user
			user, err := awssdk.GetCurrentUser(ctx, client)
			if err != nil {
				return handleMFAErrors(err)
			}

			// Get current password
			password, err := getPassword()
			if err != nil {
				return fmt.Errorf("❌ Operation failed: Invalid credentials")
			}

			// Clear password from memory when done
			passwordBytes := []byte(password)
			defer func() {
				for i := range passwordBytes {
					passwordBytes[i] = 0
				}
			}()

			// Get MFA token
			mfaToken, err := getMFAToken()
			if err != nil {
				return fmt.Errorf("❌ Operation failed: Invalid credentials")
			}

			// Double confirmation
			fmt.Print("Are you sure you want to disable MFA? Type 'YES' to confirm: ")
			reader := bufio.NewReader(os.Stdin)
			confirmation, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("❌ Operation failed: Invalid credentials")
			}
			confirmation = strings.TrimSpace(confirmation)

			if confirmation != "YES" {
				return fmt.Errorf("❌ Operation cancelled")
			}

			// Disable MFA
			err = disableMFA(ctx, client, profile, *user.UserName, password, mfaToken)
			if err != nil {
				return fmt.Errorf("❌ Operation failed: Invalid credentials")
			}

			fmt.Println("✅ MFA disabled successfully")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("profile", "p", "", "Use a specific profile from your credential file")

	return cmd
}

// disableMFA disables MFA for the user
func disableMFA(ctx context.Context, client *iam.Client, profile string, username string, password, mfaToken string) error {
	// First validate current credentials
	if err := validateCredentials(ctx, client, profile, &username, password, mfaToken); err != nil {
		return err
	}

	// List MFA devices to get the serial number
	listInput := &iam.ListMFADevicesInput{
		UserName: &username,
	}

	listResult, err := client.ListMFADevices(ctx, listInput)
	if err != nil {
		return fmt.Errorf("failed to list MFA devices: %w", err)
	}

	// Check if any MFA devices exist
	if len(listResult.MFADevices) == 0 {
		return fmt.Errorf("no MFA devices found for user")
	}

	// Use the first MFA device (assuming only one)
	device := listResult.MFADevices[0]

	// Deactivate MFA device
	deactivateInput := &iam.DeactivateMFADeviceInput{
		UserName:     &username,
		SerialNumber: device.SerialNumber,
	}

	_, err = client.DeactivateMFADevice(ctx, deactivateInput)
	if err != nil {
		return fmt.Errorf("failed to deactivate MFA device: %w", err)
	}

	// Delete virtual MFA device
	deleteInput := &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: device.SerialNumber,
	}

	_, err = client.DeleteVirtualMFADevice(ctx, deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete virtual MFA device: %w", err)
	}

	return nil
}