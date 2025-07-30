package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// NewIAMClient creates a new IAM client using AWS SDK v2 patterns
func NewIAMClient(profile string) (*iam.Client, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Config options for v2
	var opts []func(*config.LoadOptions) error

	// Handle profile - use "default" when none specified
	if profile == "" {
		opts = append(opts, config.WithSharedConfigProfile("default"))
	} else {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return iam.NewFromConfig(cfg), nil
}