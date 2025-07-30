package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// GetMFAEnabledClient creates a client with MFA token for sensitive operations
func GetMFAEnabledClient(ctx context.Context, baseProfile, roleARN, serialNumber, mfaToken string) (*iam.Client, error) {
	// 1. Load base configuration
	baseCfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(baseProfile))
	if err != nil {
		return nil, err
	}

	// 2. Create STS client from base configuration
	stsClient := sts.NewFromConfig(baseCfg)

	// 3. Configure AssumeRole with MFA
	creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN,
		func(o *stscreds.AssumeRoleOptions) {
			o.SerialNumber = aws.String(serialNumber)
			o.TokenProvider = func() (string, error) {
				return mfaToken, nil
			}
		})

	// 4. Create new configuration with MFA credentials
	mfaCfg := aws.NewConfig()
	*mfaCfg = baseCfg.Copy()
	mfaCfg.Credentials = aws.NewCredentialsCache(creds)

	return iam.NewFromConfig(*mfaCfg), nil
}