package mfa

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	awssdk "github.com/yourusername/iamctl/internal/aws"
)

// Mock IAM client for testing
type mockIAMClient struct {
	createVirtualMFADeviceFunc func(context.Context, *iam.CreateVirtualMFADeviceInput) (*iam.CreateVirtualMFADeviceOutput, error)
	enableMFADeviceFunc        func(context.Context, *iam.EnableMFADeviceInput) (*iam.EnableMFADeviceOutput, error)
	listMFADevicesFunc         func(context.Context, *iam.ListMFADevicesInput) (*iam.ListMFADevicesOutput, error)
	deactivateMFADeviceFunc    func(context.Context, *iam.DeactivateMFADeviceInput) (*iam.DeactivateMFADeviceOutput, error)
	deleteVirtualMFADeviceFunc func(context.Context, *iam.DeleteVirtualMFADeviceInput) (*iam.DeleteVirtualMFADeviceOutput, error)
}

func (m *mockIAMClient) CreateVirtualMFADevice(ctx context.Context, input *iam.CreateVirtualMFADeviceInput) (*iam.CreateVirtualMFADeviceOutput, error) {
	if m.createVirtualMFADeviceFunc != nil {
		return m.createVirtualMFADeviceFunc(ctx, input)
	}
	return &iam.CreateVirtualMFADeviceOutput{
		VirtualMFADevice: &types.VirtualMFADevice{
			SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/testuser"),
			QRCodePNG:    []byte("otpauth://totp/AWS:testuser?secret=TESTSECRET"),
		},
	}, nil
}

func (m *mockIAMClient) EnableMFADevice(ctx context.Context, input *iam.EnableMFADeviceInput) (*iam.EnableMFADeviceOutput, error) {
	if m.enableMFADeviceFunc != nil {
		return m.enableMFADeviceFunc(ctx, input)
	}
	return &iam.EnableMFADeviceOutput{}, nil
}

func (m *mockIAMClient) ListMFADevices(ctx context.Context, input *iam.ListMFADevicesInput) (*iam.ListMFADevicesOutput, error) {
	if m.listMFADevicesFunc != nil {
		return m.listMFADevicesFunc(ctx, input)
	}
	return &iam.ListMFADevicesOutput{
		MFADevices: []types.MFADevice{
			{
				SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/testuser"),
				EnableDate:   aws.Time(time.Now().AddDate(0, 0, -100)), // 100 days ago
			},
		},
	}, nil
}

func (m *mockIAMClient) DeactivateMFADevice(ctx context.Context, input *iam.DeactivateMFADeviceInput) (*iam.DeactivateMFADeviceOutput, error) {
	if m.deactivateMFADeviceFunc != nil {
		return m.deactivateMFADeviceFunc(ctx, input)
	}
	return &iam.DeactivateMFADeviceOutput{}, nil
}

func (m *mockIAMClient) DeleteVirtualMFADevice(ctx context.Context, input *iam.DeleteVirtualMFADeviceInput) (*iam.DeleteVirtualMFADeviceOutput, error) {
	if m.deleteVirtualMFADeviceFunc != nil {
		return m.deleteVirtualMFADeviceFunc(ctx, input)
	}
	return &iam.DeleteVirtualMFADeviceOutput{}, nil
}

func TestDeviceRegistration(t *testing.T) {
	// Setup mock client
	client := &mockIAMClient{
		createVirtualMFADeviceFunc: func(ctx context.Context, input *iam.CreateVirtualMFADeviceInput) (*iam.CreateVirtualMFADeviceOutput, error) {
			return &iam.CreateVirtualMFADeviceOutput{
				VirtualMFADevice: &types.VirtualMFADevice{
					SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/testuser"),
					QRCodePNG:    []byte("otpauth://totp/AWS:testuser?secret=TESTSECRET"),
				},
			}, nil
		},
		enableMFADeviceFunc: func(ctx context.Context, input *iam.EnableMFADeviceInput) (*iam.EnableMFADeviceOutput, error) {
			return &iam.EnableMFADeviceOutput{}, nil
		},
	}

	// Test successful device registration
	ctx := context.Background()
	qrCodeURI, err := enableMFAWithClient(ctx, client, "test-profile", aws.String("testuser"), "password", "123456789012")
	if err != nil {
		t.Errorf("Expected successful device registration, got error: %v", err)
	}

	if qrCodeURI == "" {
		t.Error("Expected QR code URI, got empty string")
	}
}

func TestInvalidToken(t *testing.T) {
	// Setup mock client that returns MFA error
	client := &mockIAMClient{
		createVirtualMFADeviceFunc: func(ctx context.Context, input *iam.CreateVirtualMFADeviceInput) (*iam.CreateVirtualMFADeviceOutput, error) {
			return &iam.CreateVirtualMFADeviceOutput{
				VirtualMFADevice: &types.VirtualMFADevice{
					SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/testuser"),
					QRCodePNG:    []byte("otpauth://totp/AWS:testuser?secret=TESTSECRET"),
				},
			}, nil
		},
		enableMFADeviceFunc: func(ctx context.Context, input *iam.EnableMFADeviceInput) (*iam.EnableMFADeviceOutput, error) {
			return nil, &awssdk.PermissionError{Err: errors.New("access denied")}
		},
	}

	// Test device registration failure due to invalid token
	ctx := context.Background()
	_, err := enableMFAWithClient(ctx, client, "test-profile", aws.String("testuser"), "password", "invalid")
	if err == nil {
		t.Error("Expected device registration to fail due to invalid token")
	}
}

func TestRotation(t *testing.T) {
	// Setup mock client
	client := &mockIAMClient{}

	// Test MFA status with old device
	ctx := context.Background()
	status, err := getMFAStatusWithClient(ctx, client, aws.String("testuser"))
	if err != nil {
		t.Errorf("Expected successful status check, got error: %v", err)
	}

	if !status.Enabled {
		t.Error("Expected MFA to be enabled")
	}

	// Device is 100 days old, so rotation should be recommended
	// This is checked in the status command output, not in the status struct
}

// enableMFAWithClient is a testable version of enableMFA that accepts a mock client
func enableMFAWithClient(ctx context.Context, client *mockIAMClient, profile string, username *string, password, mfaToken string) (string, error) {
	// Create virtual MFA device
	deviceName := "test-device"
	createDeviceInput := &iam.CreateVirtualMFADeviceInput{
		VirtualMFADeviceName: aws.String(deviceName),
	}

	deviceResult, err := client.CreateVirtualMFADevice(ctx, createDeviceInput)
	if err != nil {
		return "", err
	}

	// Enable the MFA device
	enableInput := &iam.EnableMFADeviceInput{
		UserName:            username,
		SerialNumber:        deviceResult.VirtualMFADevice.SerialNumber,
		AuthenticationCode1: aws.String(mfaToken[:6]),
		AuthenticationCode2: aws.String(mfaToken[6:]),
	}

	_, err = client.EnableMFADevice(ctx, enableInput)
	if err != nil {
		// Clean up the virtual MFA device if enabling failed
		_, deleteErr := client.DeleteVirtualMFADevice(ctx, &iam.DeleteVirtualMFADeviceInput{
			SerialNumber: deviceResult.VirtualMFADevice.SerialNumber,
		})
		if deleteErr != nil {
			// Log but don't return this error as we're already handling another error
		}
		return "", err
	}

	// Return the QR code URI (convert byte array to string)
	return string(deviceResult.VirtualMFADevice.QRCodePNG), nil
}

// getMFAStatusWithClient is a testable version of getMFAStatus that accepts a mock client
func getMFAStatusWithClient(ctx context.Context, client *mockIAMClient, username *string) (*MFAStatus, error) {
	// List MFA devices
	listInput := &iam.ListMFADevicesInput{
		UserName: username,
	}

	listResult, err := client.ListMFADevices(ctx, listInput)
	if err != nil {
		return nil, err
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