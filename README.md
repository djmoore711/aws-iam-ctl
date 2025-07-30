# iamctl - AWS IAM CLI Utility

A cross-platform CLI tool for managing AWS IAM credentials directly from the terminal.

## Features

- `iamctl configure` - Set up AWS profile (like `aws configure`)
- `iamctl keys rotate` - Rotate access keys securely
- `iamctl password reset` - Change IAM user password
- `iamctl mfa enable` - Enable virtual MFA (TOTP)
- `iamctl mfa disable` - Disable MFA
- `iamctl status` - Show current IAM user, key age, MFA status (with optional CSV output)

## Installation

Download the appropriate binary for your platform from the releases page.

## Usage

```bash
# Configure AWS credentials
iamctl configure

# Check current status (text output)
iamctl status

# Check current status (CSV output)
iamctl status -o csv

# Rotate access keys
iamctl keys rotate

# Reset password (requires MFA)
iamctl password reset

# Enable MFA
iamctl mfa enable

# Disable MFA
iamctl mfa disable
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/iamctl.git
cd iamctl

# Build for your platform
go build -o iamctl .

# Or build for all supported platforms
./scripts/build-release.sh
```

## Security

All sensitive IAM actions require MFA token input. The tool never logs or prints secrets and uses secure credential storage compatible with AWS CLI.