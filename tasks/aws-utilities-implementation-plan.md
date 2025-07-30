# Phase 2 Implementation Plan: AWS Utilities Layer

## PRODUCT MANAGER & LEAD DEVELOPER PLAN

This document serves as the detailed implementation plan for Phase 2 of the iamctl project, focusing on the AWS utilities layer. This layer is critical for establishing secure connections to AWS services and retrieving user information while maintaining the highest security standards.

## ️ SECURITY PRIME DIRECTIVE

Credentials must NEVER be exposed to memory longer than necessary. Our implementation will follow AWS SDK's built-in credential chain methodology, not custom parsing. This is non-negotiable for security.

## IMPLEMENTATION PLAN

### ✅ File 1: internal/aws/client.go

This file will implement the NewIAMClient function that establishes a secure connection to AWS using the SDK's built-in credential provider chain.

Implementation details:
- Use AWS SDK's session management with SharedConfigEnable
- Handle profile fallback to "default" when none specified
- Implement 15-second timeout context for all operations
- Return appropriate error types for credential failures

### ✅ File 2: internal/aws/utils.go

This file will implement the GetCurrentUser function that retrieves the current IAM user with proper MFA handling.

Implementation details:
- Use the iam.GetUser API call with proper context
- Implement MFA pattern for sensitive operations using X-Amz-Security-Token header
- Classify errors into specific types (credentials_error, permission_error, service_error)
- Never leak sensitive details in error messages

### ✅ File 3: internal/aws/client_test.go

This file will contain comprehensive tests covering all credential scenarios:

Test scenarios:
1. Valid default profile
2. Valid custom profile
3. Missing profile
4. Malformed credentials
5. Invalid credentials
6. MFA-enabled profile

### MEMORY MANAGEMENT REQUIREMENT

All sensitive data will implement secure cleanup:
- Use defer statements to clear sensitive data after use
- Follow AWS SDK's memory management patterns
- Ensure no credentials remain in memory longer than necessary

### VERIFICATION REQUIREMENTS

Before proceeding, we will complete all verification steps:
1. Run go test with verbose output for the aws package
2. Run gosec security static analysis on the aws package
3. Verify no credential leaks in memory profiler
4. Test with AWS_PROFILE set to non-default value

## MODELING SECURITY BEST PRACTICES

Before implementation, we will model security best practices:
1. Review AWS SDK documentation for credential management
2. Examine error handling patterns in AWS SDK
3. Study MFA implementation requirements in AWS APIs
4. Verify memory management techniques for sensitive data

This approach ensures we "model security best practices first, then implement" as recommended.

## NEXT STEPS

After approval of this plan, we will proceed with implementation following the exact specifications provided, ensuring all security requirements are met before considering the implementation complete.