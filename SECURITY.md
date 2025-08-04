# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in this project, please report it privately by emailing the project maintainers. Do not create public issues for security vulnerabilities.

## Test API Keys

This repository may contain test API keys in git history that are used only for testing purposes. These keys:

1. **Pass validation**: They match the expected Claude API key format for testing
2. **Fail authentication**: They will not work with the actual Claude API
3. **Are documented**: Known test keys follow patterns like `sk-ant-api03-*`

**These are NOT real API keys** and pose no security risk. They are designed to:
- Test the SDK's authentication handling
- Validate API key format validation
- Ensure CI/CD pipelines work correctly
- Demonstrate proper error handling

## Security Best Practices

When using this SDK in production:

1. **Never commit real API keys** to version control
2. **Use environment variables** for sensitive configuration
3. **Use GitHub secrets** for CI/CD workflows
4. **Follow the principle of least privilege** for API access
5. **Regularly rotate your API keys**

## Protected Information

The `.gitignore` file includes patterns to prevent accidental commitment of:
- Environment files (`.env*`)
- API keys and secrets (`*.key`, `*_secret`, etc.)
- Private keys and certificates (`*.pem`, `*.p12`, etc.)
- Configuration files containing secrets

## Security Scanning

This project uses:
- **GitGuardian**: For secret detection (may flag test keys - this is expected)
- **Gosec**: For Go security analysis
- **Dependency scanning**: For vulnerable dependencies

Test keys that match real API patterns may trigger security scanners. This is expected behavior and does not indicate a real security issue when the keys are properly documented as test keys.

## Secure Development

Contributors should:
- Use `.env` files for local development (excluded from git)
- Never commit real credentials or API keys
- Use the provided test keys for development and testing
- Report any suspicious security patterns or vulnerabilities