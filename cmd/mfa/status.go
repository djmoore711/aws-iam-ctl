package mfa

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awssdk "github.com/yourusername/iamctl/internal/aws"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the MFA status command
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show MFA enrollment status",
		Long:  `Display information about the current user's MFA enrollment status.`,
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

			// Get MFA status
			mfaStatus, err := getMFAStatus(ctx, client, user.UserName)
			if err != nil {
				return fmt.Errorf("❌ Operation failed: Invalid credentials")
			}

			// Display MFA status
			fmt.Printf("User: %s\n", *user.UserName)
			fmt.Printf("MFA Status: %s\n", mfaStatus.Status)
			
			if mfaStatus.Enabled {
				fmt.Printf("MFA Device: %s\n", mfaStatus.Device)
				fmt.Printf("Enrolled: %s\n", mfaStatus.Enrolled.Format("2006-01-02 15:04:05 MST"))
				
				// Check if rotation is needed (90 days)
				if time.Since(mfaStatus.Enrolled) > 90*24*time.Hour {
					// For enterprise security, we enforce rotation after 90 days
					return fmt.Errorf("❌ MFA device requires rotation (older than 90 days)")
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("profile", "p", "", "Use a specific profile from your credential file")

	return cmd
}

// MFAStatus represents the MFA enrollment status
type MFAStatus struct {
	Enabled  bool
	Status   string
	Device   string
	Enrolled time.Time
}

// getMFAStatus retrieves the MFA status for a user
func getMFAStatus(ctx context.Context, client *iam.Client, username *string) (*MFAStatus, error) {
	// List MFA devices
	listInput := &iam.ListMFADevicesInput{
		UserName: username,
	}

	listResult, err := client.ListMFADevices(ctx, listInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list MFA devices: %w", err)
	}

	// Check if any MFA devices exist
	if len(listResult.MFADevices) == 0 {
		return &MFAStatus{
			Enabled: false,
			Status:  "Disabled",
			Device:  "None",
		}, nil
	}

	// Use the first MFA device (assuming only one)
	device := listResult.MFADevices[0]

	return &MFAStatus{
		Enabled:  true,
		Status:   "Enabled",
		Device:   *device.SerialNumber,
		Enrolled: *device.EnableDate,
	}, nil
}
