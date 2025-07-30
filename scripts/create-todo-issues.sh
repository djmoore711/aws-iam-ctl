#!/bin/bash

# Script to create TODO issues for placeholder implementations

# Create issues directory if it doesn't exist
mkdir -p issues

# Create issue for enforce test placeholders
cat > issues/enforce-test-placeholders.md << 'EOF'
---
title: "Implement proper mocking for enforce command tests"
labels: ["enhancement", "testing"]
---
The enforce command tests currently use placeholder implementations because we can't easily mock the IAM client interface. We should implement proper mocking to have more comprehensive tests.
EOF

# Create issue for status command MFA check
cat > issues/status-mfa-implementation.md << 'EOF'
---
title: "Implement proper MFA status check in status command"
labels: ["enhancement", "feature"]
---
The status command has a placeholder implementation for checking MFA status. We need to implement the actual logic to check if MFA is enabled for the current user.
EOF

echo "TODO issues created in the 'issues' directory"