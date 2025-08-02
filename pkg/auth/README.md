# Authentication Package

The `auth` package provides secure authentication mechanisms for the Claude SDK with multiple authentication methods and built-in security features.

## Features

- **Multiple Authentication Methods**
  - API Key authentication
  - Session-based authentication with token expiration
  - Extensible interface for custom authentication methods

- **Security Features**
  - Secure credential validation with entropy checking
  - Protection against test/example credentials
  - Thread-safe credential management
  - Automatic session refresh capabilities

- **Credential Management**
  - In-memory and file-based credential storage
  - Secure credential lifecycle management
  - Multi-credential support with unique IDs

## Usage

### API Key Authentication

```go
import "github.com/jraymond/claude-code-go-sdk/pkg/auth"

// Create API key authenticator
apiAuth, err := auth.NewAPIKeyAuthenticator("sk-ant-api03-your-key-here")
if err != nil {
    log.Fatal(err)
}

// Get authentication headers
headers := apiAuth.GetHeaders()
// headers["X-API-Key"] = "sk-ant-api03-your-key-here"
```

### Session Authentication

```go
// Create session authenticator
sessionAuth, err := auth.NewSessionAuthenticator("sess_your_session_key")
if err != nil {
    log.Fatal(err)
}

// Authenticate to establish session
ctx := context.Background()
authInfo, err := sessionAuth.Authenticate(ctx)
if err != nil {
    log.Fatal(err)
}

// Check if authentication is valid
if sessionAuth.IsValid() {
    // Use the session
}

// Refresh session before expiration
err = sessionAuth.Refresh(ctx)
```

### Credential Management

```go
// Create a credential manager with file storage
store, err := auth.NewFileStore("/path/to/credentials")
if err != nil {
    log.Fatal(err)
}

manager := auth.NewManager(store)

// Store API key
err = manager.StoreAPIKey(ctx, "my-api-key", "sk-ant-api03-your-key")
if err != nil {
    log.Fatal(err)
}

// Retrieve and use authenticator
authenticator, err := manager.GetAuthenticator(ctx, "my-api-key")
if err != nil {
    log.Fatal(err)
}
```

## Validation

The package includes comprehensive validation for credentials:

- **API Keys**: Must follow the pattern `sk-ant-*` with sufficient entropy
- **Session Keys**: Must be at least 32 characters with alphanumeric content
- **Security Checks**: Rejects test keys, low entropy patterns, and sequential characters

### Custom Validation

```go
// Create validator with custom rules
validator := auth.NewCustomValidator(
    auth.WithAPIKeyPattern(`^custom-prefix-[a-zA-Z0-9]+$`),
    auth.WithAPIKeyLength(20, 100),
    auth.WithComplexityRequirements(true, true, true, false),
)
```

## Security Best Practices

1. **Never hardcode credentials** - Use environment variables or secure storage
2. **Rotate credentials regularly** - Implement credential rotation policies
3. **Use appropriate storage** - FileStore for persistence, MemoryStore for testing
4. **Monitor authentication failures** - Implement logging and alerting
5. **Implement rate limiting** - Protect against brute force attacks

## Testing

The package includes comprehensive test coverage for all authentication methods, validation rules, and storage mechanisms. Run tests with:

```bash
go test ./pkg/auth -v
```