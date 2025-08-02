package errors

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// NetworkError represents network connectivity and communication errors.
type NetworkError struct {
	*BaseError
	Operation string        // The network operation that failed (dial, read, write, etc.)
	Address   string        // The address being connected to (sanitized)
	Timeout   time.Duration // Timeout value if applicable
}

// NewNetworkError creates a new network error.
func NewNetworkError(operation, address string, cause error) *NetworkError {
	message := fmt.Sprintf("Network %s failed", operation)
	if address != "" {
		// Sanitize the address to avoid exposing sensitive information
		sanitizedAddr := sanitizeAddress(address)
		message = fmt.Sprintf("Network %s failed for %s", operation, sanitizedAddr)
	}

	// Determine if the error is retryable based on the operation and cause
	retryable := isNetworkErrorRetryable(cause)

	err := &NetworkError{
		BaseError: NewBaseError(CategoryNetwork, SeverityHigh, "NETWORK_ERROR", message).
			WithCause(cause).
			WithRetryable(retryable),
		Operation: operation,
		Address:   sanitizeAddress(address),
	}

	err.WithDetail("operation", operation).
		WithDetail("address", sanitizeAddress(address))

	return err
}

// TimeoutError represents timeout errors with context about what timed out.
type TimeoutError struct {
	*BaseError
	Operation string        // The operation that timed out
	Timeout   time.Duration // The timeout duration
	Elapsed   time.Duration // How long the operation ran before timing out
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(operation string, timeout, elapsed time.Duration) *TimeoutError {
	message := fmt.Sprintf("%s timed out after %v", operation, elapsed)
	if timeout > 0 {
		message = fmt.Sprintf("%s timed out after %v (timeout: %v)", operation, elapsed, timeout)
	}

	err := &TimeoutError{
		BaseError: NewBaseError(CategoryNetwork, SeverityMedium, "TIMEOUT_ERROR", message).
			WithRetryable(true), // Timeouts are generally retryable
		Operation: operation,
		Timeout:   timeout,
		Elapsed:   elapsed,
	}

	err.WithDetail("operation", operation).
		WithDetail("timeout_seconds", timeout.Seconds()).
		WithDetail("elapsed_seconds", elapsed.Seconds())

	return err
}

// ConnectionError represents connection establishment failures.
type ConnectionError struct {
	*BaseError
	Address string // The address that failed to connect (sanitized)
	Reason  string // The reason for connection failure
}

// NewConnectionError creates a new connection error.
func NewConnectionError(address, reason string, cause error) *ConnectionError {
	sanitizedAddr := sanitizeAddress(address)
	message := fmt.Sprintf("Failed to connect to %s", sanitizedAddr)
	if reason != "" {
		message = fmt.Sprintf("Failed to connect to %s: %s", sanitizedAddr, reason)
	}

	// Connection errors are generally retryable unless it's a permanent failure
	retryable := !isPermanentConnectionFailure(cause)

	err := &ConnectionError{
		BaseError: NewBaseError(CategoryNetwork, SeverityHigh, "CONNECTION_ERROR", message).
			WithCause(cause).
			WithRetryable(retryable),
		Address: sanitizedAddr,
		Reason:  reason,
	}

	err.WithDetail("address", sanitizedAddr).
		WithDetail("reason", reason)

	return err
}

// TLSError represents TLS/SSL related errors.
type TLSError struct {
	*BaseError
	Address  string // The address where TLS failed (sanitized)
	Reason   string // The reason for TLS failure
	CertInfo string // Certificate information if available (sanitized)
}

// NewTLSError creates a new TLS error.
func NewTLSError(address, reason string, cause error) *TLSError {
	sanitizedAddr := sanitizeAddress(address)
	message := fmt.Sprintf("TLS handshake failed for %s", sanitizedAddr)
	if reason != "" {
		message = fmt.Sprintf("TLS handshake failed for %s: %s", sanitizedAddr, reason)
	}

	// TLS errors are typically not retryable unless it's a timeout
	retryable := strings.Contains(strings.ToLower(reason), "timeout")

	err := &TLSError{
		BaseError: NewBaseError(CategoryNetwork, SeverityCritical, "TLS_ERROR", message).
			WithCause(cause).
			WithRetryable(retryable),
		Address: sanitizedAddr,
		Reason:  reason,
	}

	err.WithDetail("address", sanitizedAddr).
		WithDetail("reason", reason)

	// Extract certificate information if available
	if tlsErr, ok := cause.(*tls.CertificateVerificationError); ok {
		certInfo := sanitizeCertificateInfo(tlsErr)
		err.CertInfo = certInfo
		err.WithDetail("certificate_info", certInfo)
	}

	return err
}

// DNSError represents DNS resolution failures.
type DNSError struct {
	*BaseError
	Hostname string // The hostname that failed to resolve
	DNSType  string // The type of DNS query (A, AAAA, etc.)
}

// NewDNSError creates a new DNS error.
func NewDNSError(hostname, dnsType string, cause error) *DNSError {
	message := fmt.Sprintf("DNS resolution failed for %s", hostname)
	if dnsType != "" {
		message = fmt.Sprintf("DNS %s resolution failed for %s", dnsType, hostname)
	}

	// DNS errors are generally retryable
	retryable := true

	err := &DNSError{
		BaseError: NewBaseError(CategoryNetwork, SeverityMedium, "DNS_ERROR", message).
			WithCause(cause).
			WithRetryable(retryable),
		Hostname: hostname,
		DNSType:  dnsType,
	}

	err.WithDetail("hostname", hostname).
		WithDetail("dns_type", dnsType)

	return err
}

// ProxyError represents proxy-related errors.
type ProxyError struct {
	*BaseError
	ProxyAddress string // The proxy address (sanitized)
	ProxyType    string // The type of proxy (HTTP, SOCKS, etc.)
}

// NewProxyError creates a new proxy error.
func NewProxyError(proxyAddress, proxyType string, cause error) *ProxyError {
	sanitizedAddr := sanitizeAddress(proxyAddress)
	message := fmt.Sprintf("Proxy connection failed")
	if sanitizedAddr != "" {
		message = fmt.Sprintf("Proxy connection failed for %s", sanitizedAddr)
	}

	// Proxy errors are generally retryable
	retryable := true

	err := &ProxyError{
		BaseError: NewBaseError(CategoryNetwork, SeverityHigh, "PROXY_ERROR", message).
			WithCause(cause).
			WithRetryable(retryable),
		ProxyAddress: sanitizedAddr,
		ProxyType:    proxyType,
	}

	err.WithDetail("proxy_address", sanitizedAddr).
		WithDetail("proxy_type", proxyType)

	return err
}

// Helper functions for network error analysis

// ClassifyNetworkError analyzes a generic error and creates an appropriate network error type.
func ClassifyNetworkError(err error) SDKError {
	if err == nil {
		return nil
	}

	// Check for context cancellation/timeout
	if err == context.Canceled {
		return NewBaseError(CategoryNetwork, SeverityMedium, "CONTEXT_CANCELED",
			"Operation was canceled").WithCause(err).WithRetryable(false)
	}
	if err == context.DeadlineExceeded {
		return NewTimeoutError("context", 0, 0).BaseError.WithCause(err)
	}

	// Check for net.Error interface
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return NewTimeoutError("network operation", 0, 0).BaseError.WithCause(err)
		}
		if netErr.Temporary() {
			return NewNetworkError("network operation", "", err)
		}
	}

	// Check for specific error types
	switch e := err.(type) {
	case *net.OpError:
		return classifyOpError(e)
	case *net.DNSError:
		return NewDNSError(e.Name, "", err)
	case *tls.CertificateVerificationError:
		return NewTLSError("", "certificate verification failed", err)
	case *url.Error:
		return classifyURLError(e)
	}

	// Check for syscall errors
	if isConnectionRefused(err) {
		return NewConnectionError("", "connection refused", err)
	}
	if isHostUnreachable(err) {
		return NewConnectionError("", "host unreachable", err)
	}
	if isNetworkUnreachable(err) {
		return NewConnectionError("", "network unreachable", err)
	}

	// Default to generic network error
	return NewNetworkError("unknown", "", err)
}

// classifyOpError analyzes a net.OpError and creates an appropriate error type.
func classifyOpError(opErr *net.OpError) SDKError {
	op := opErr.Op
	addr := ""
	if opErr.Addr != nil {
		addr = opErr.Addr.String()
	}

	if opErr.Timeout() {
		return NewTimeoutError(op, 0, 0).BaseError.WithCause(opErr)
	}

	// Check the underlying error
	if opErr.Err != nil {
		if strings.Contains(opErr.Err.Error(), "connection refused") {
			return NewConnectionError(addr, "connection refused", opErr)
		}
		if strings.Contains(opErr.Err.Error(), "no such host") {
			return NewDNSError(addr, "", opErr)
		}
		if strings.Contains(opErr.Err.Error(), "network is unreachable") {
			return NewConnectionError(addr, "network unreachable", opErr)
		}
	}

	return NewNetworkError(op, addr, opErr)
}

// classifyURLError analyzes a url.Error and creates an appropriate error type.
func classifyURLError(urlErr *url.Error) SDKError {
	if urlErr.Timeout() {
		return NewTimeoutError(urlErr.Op, 0, 0).BaseError.WithCause(urlErr)
	}

	// Recursively classify the underlying error
	if underlying := ClassifyNetworkError(urlErr.Err); underlying != nil {
		return underlying
	}

	return NewNetworkError(urlErr.Op, urlErr.URL, urlErr)
}

// isNetworkErrorRetryable determines if a network error is retryable.
func isNetworkErrorRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for permanent failures
	if isPermanentConnectionFailure(err) {
		return false
	}

	// Context cancellation is not retryable
	if err == context.Canceled {
		return false
	}

	// Most network errors are retryable
	return true
}

// isPermanentConnectionFailure checks if a connection failure is permanent.
func isPermanentConnectionFailure(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// These are typically permanent failures
	permanentPatterns := []string{
		"no such host",
		"invalid hostname",
		"invalid port",
		"address family not supported",
		"protocol not supported",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// isConnectionRefused checks if the error indicates connection refused.
func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}

	// Check for syscall.ECONNREFUSED
	if errno, ok := err.(syscall.Errno); ok {
		return errno == syscall.ECONNREFUSED
	}

	// Check error string
	return strings.Contains(strings.ToLower(err.Error()), "connection refused")
}

// isHostUnreachable checks if the error indicates host unreachable.
func isHostUnreachable(err error) bool {
	if err == nil {
		return false
	}

	// Check for syscall.EHOSTUNREACH
	if errno, ok := err.(syscall.Errno); ok {
		return errno == syscall.EHOSTUNREACH
	}

	// Check error string
	return strings.Contains(strings.ToLower(err.Error()), "host unreachable")
}

// isNetworkUnreachable checks if the error indicates network unreachable.
func isNetworkUnreachable(err error) bool {
	if err == nil {
		return false
	}

	// Check for syscall.ENETUNREACH
	if errno, ok := err.(syscall.Errno); ok {
		return errno == syscall.ENETUNREACH
	}

	// Check error string
	return strings.Contains(strings.ToLower(err.Error()), "network unreachable")
}

// sanitizeAddress sanitizes a network address to avoid exposing sensitive information.
func sanitizeAddress(address string) string {
	if address == "" {
		return ""
	}

	// Try to parse as URL first (only if it looks like a URL with a scheme)
	if strings.Contains(address, "://") {
		if u, err := url.Parse(address); err == nil {
			scheme := u.Scheme
			host := u.Host

			// Redact userinfo if present
			if u.User != nil {
				host = "[REDACTED]@" + u.Hostname()
				if port := u.Port(); port != "" {
					host += ":" + port
				}
			}

			return fmt.Sprintf("%s://%s", scheme, host)
		}
	}

	// If it looks like host:port, preserve that format
	if strings.Contains(address, ":") {
		if host, port, err := net.SplitHostPort(address); err == nil {
			return fmt.Sprintf("%s:%s", host, port)
		}
	}

	// Return as-is if it doesn't look sensitive
	return address
}

// sanitizeCertificateInfo extracts safe certificate information for debugging.
func sanitizeCertificateInfo(certErr *tls.CertificateVerificationError) string {
	if certErr == nil {
		return ""
	}

	// Extract non-sensitive certificate information
	info := []string{}

	// Add error details without exposing certificate contents
	info = append(info, fmt.Sprintf("verification_error: %s", certErr.Error()))

	return strings.Join(info, ", ")
}
