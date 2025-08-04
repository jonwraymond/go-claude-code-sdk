# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability within the Go Claude Code SDK, please follow these steps:

1. **Do NOT open a public issue** - This could put users at risk

2. **Email us directly** at: security@[domain].com
   - Include the word "SECURITY" in the subject line
   - Provide a detailed description of the vulnerability
   - Include steps to reproduce if possible
   - Suggest a fix if you have one

3. **What to expect:**
   - You'll receive an acknowledgment within 48 hours
   - We'll work with you to understand and validate the issue
   - We'll prepare a fix and coordinate the disclosure timeline
   - We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

When using the Go Claude Code SDK:

1. **API Keys**: Never hardcode API keys in your source code
   ```go
   // Bad
   options.APIKey = "sk-abc123..."
   
   // Good
   options.APIKey = os.Getenv("CLAUDE_API_KEY")
   ```

2. **Permissions**: Use the least permissive mode necessary
   ```go
   // Only use bypassPermissions in controlled environments
   options.PermissionMode = &claudecode.PermissionModeDefault
   ```

3. **Tool Restrictions**: Limit allowed tools to what you need
   ```go
   options.AllowedTools = []string{"Read", "Write"} // Don't include "Bash" unless needed
   ```

4. **Input Validation**: Always validate user inputs before passing to Claude
   ```go
   if err := validateUserInput(prompt); err != nil {
       return err
   }
   ```

5. **Error Handling**: Don't expose sensitive information in errors
   ```go
   // Don't log full error details that might contain sensitive data
   log.Printf("Query failed: %v", errors.New("internal error"))
   ```

## Security Features

The SDK includes several security features:

- **Permission Modes**: Control what tools Claude can use
- **Tool Restrictions**: Limit available tools
- **Working Directory Isolation**: Restrict file system access
- **No Shell Injection**: Safe command execution
- **Input Sanitization**: Automatic sanitization of inputs

## Vulnerability Disclosure

We follow a coordinated disclosure process:

1. Security issues are assessed and fixed privately
2. A security advisory is prepared
3. The fix is released with the advisory
4. Credit is given to the reporter (with permission)

## Contact

For security concerns, please email: security@[domain].com

For general bugs and issues, use the public issue tracker.