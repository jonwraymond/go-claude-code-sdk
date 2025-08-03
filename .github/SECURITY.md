# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in the Go Claude Code SDK, please report it by emailing the maintainers. Do not open a public issue.

## Security Scanning

This repository uses multiple security scanning tools:

### GitGuardian
- Scans for secrets and credentials
- Configuration: `.gitguardian.yml`
- False positives are documented in `SECURITY_FALSE_POSITIVES.md`

### CodeQL
- Performs semantic code analysis
- Configuration: `.github/codeql/codeql-config.yml`
- Runs on all PRs and pushes to main

### gosec
- Go-specific security scanner
- Integrated into CI pipeline
- Suppressions use `#nosec` comments with justification

## Test Credentials

All credentials in this repository are test fixtures:
- `test-key-for-unit-tests` - Unit test placeholder
- `test-api-key-not-real` - Test fixture
- Any keys in `/tests/fixtures/` are not real

These are explicitly marked and excluded from security scanning.

## Best Practices

1. Never commit real API keys or secrets
2. Use environment variables for sensitive data
3. Test fixtures should use obviously fake values
4. Document any security scanner suppressions