# Security Policy

## Supported Versions

Currently, we support the following versions of the Claude Code Go SDK with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of the Claude Code Go SDK seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to the project maintainers. You can find contact information in the repository's commit history or profile pages.

Please include:
- Type of vulnerability (e.g., authentication bypass, arbitrary code execution)
- Full paths of source file(s) related to the vulnerability
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Initial Assessment**: Within 5 business days, we'll provide an initial assessment
- **Resolution Timeline**: We aim to release patches for critical vulnerabilities within 7-14 days
- **Credit**: We'll credit you for the discovery (unless you prefer to remain anonymous)

### Security Best Practices for SDK Users

1. **API Key Security**:
   - Never hard-code API keys in your source code
   - Use environment variables or secure credential storage
   - Rotate API keys regularly
   - Use separate API keys for development and production

2. **Input Validation**:
   - Always validate and sanitize user inputs before passing to the SDK
   - Be cautious with file paths and system commands
   - Use the SDK's permission modes appropriately

3. **Dependency Management**:
   - Keep the SDK updated to the latest version
   - Regularly run `go mod tidy` and check for vulnerabilities
   - Use tools like `govulncheck` to scan for known vulnerabilities

4. **Process Security**:
   - The SDK uses subprocess execution - ensure proper process isolation
   - Monitor subprocess resource usage
   - Implement appropriate timeouts

5. **Error Handling**:
   - Never expose detailed error messages to end users
   - Log errors securely without including sensitive information
   - Implement proper error recovery mechanisms

### Security Features

The Claude Code Go SDK includes several security features:

- Secure subprocess management with proper cleanup
- API key sanitization in logs and error messages
- Input validation for all user-provided data
- Timeout controls to prevent resource exhaustion
- Permission modes for tool execution control

## Vulnerability Disclosure Policy

We follow a coordinated disclosure policy:

1. Security issues are assessed and fixed privately
2. A new version is released with the fix
3. The vulnerability is publicly disclosed after users have had time to upgrade

Thank you for helping keep the Claude Code Go SDK secure!