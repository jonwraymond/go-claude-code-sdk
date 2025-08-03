# Security Tool False Positives

This document explains the false positives from security scanning tools and how they are addressed.

## GitGuardian

GitGuardian flags the following as potential secrets, but they are all test fixtures or examples:

1. **Test API Key in fixtures**: `sk-ant-api03-test-key` in `/tests/fixtures/test_data.go`
   - This is a test fixture, not a real API key
   - Marked with `// ggignore` comment
   - Excluded via `.gitguardian.yml` configuration

### Configuration

The `.gitguardian.yml` file is configured to:
- Exclude all test files (`**/*_test.go`, `**/fixtures/**`)
- Exclude documentation and examples
- Ignore specific known test patterns

### Manual Actions Required

If GitGuardian continues to flag these after the configuration:
1. Visit the GitGuardian dashboard
2. Mark the alerts as false positives
3. The test key pattern `sk-ant-api03-test-key` is intentionally formatted like a real key for testing purposes

## CodeQL

CodeQL may flag certain patterns as security issues. We've configured it via `.github/codeql/codeql-config.yml` to:
- Focus on high-confidence security issues
- Exclude experimental and low-severity warnings
- Skip test files from certain analyses

### Known False Positives

1. **Hard-coded credentials in test fixtures**: These are intentional for testing
2. **Command injection in mock files**: The mock CLI intentionally simulates command execution for testing

## gosec

All gosec issues have been resolved:
- File permission 0700 for test scripts is marked with `#nosec G302`
- This is required for executable test scripts and is secure (owner-only permissions)

## Summary

All flagged "secrets" in this repository are:
- Test fixtures for SDK testing
- Example code showing usage patterns
- Mock data for unit tests

No real secrets or credentials are stored in this repository.