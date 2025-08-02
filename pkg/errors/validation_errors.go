package errors

import (
	"fmt"
	"strings"
)

// ValidationError represents input validation errors with field-specific details.
type ValidationError struct {
	*BaseError
	Field      string                // The field that failed validation
	Value      string                // The invalid value (sanitized)
	Constraint string                // The validation constraint that was violated
	Violations []ValidationViolation // Detailed validation violations
}

// ValidationViolation represents a specific validation rule violation.
type ValidationViolation struct {
	Field      string      `json:"field"`      // Field path (e.g., "messages.0.content")
	Code       string      `json:"code"`       // Violation code (e.g., "required", "max_length")
	Message    string      `json:"message"`    // Human-readable message
	Value      any `json:"value"`      // The invalid value (sanitized)
	Constraint any `json:"constraint"` // The constraint that was violated
}

// NewValidationError creates a new validation error.
func NewValidationError(field, value, constraint, message string) *ValidationError {
	if message == "" {
		message = fmt.Sprintf("Validation failed for field '%s'", field)
		if constraint != "" {
			message += fmt.Sprintf(": %s", constraint)
		}
	}

	// Sanitize the value to prevent exposing sensitive data
	sanitizedValue := sanitizeValidationValue(value)

	err := &ValidationError{
		BaseError: NewBaseError(CategoryValidation, SeverityMedium, "VALIDATION_ERROR", message).
			WithRetryable(false), // Validation errors are not retryable without fixing the input
		Field:      field,
		Value:      sanitizedValue,
		Constraint: constraint,
		Violations: []ValidationViolation{},
	}

	err.WithDetail("field", field).
		WithDetail("value", sanitizedValue).
		WithDetail("constraint", constraint)

	return err
}

// NewValidationErrorWithViolations creates a validation error with multiple violations.
func NewValidationErrorWithViolations(violations []ValidationViolation) *ValidationError {
	message := "Request validation failed"
	if len(violations) == 1 {
		message = fmt.Sprintf("Validation failed for field '%s': %s",
			violations[0].Field, violations[0].Message)
	} else if len(violations) > 1 {
		message = fmt.Sprintf("Validation failed for %d fields", len(violations))
	}

	err := &ValidationError{
		BaseError: NewBaseError(CategoryValidation, SeverityMedium, "VALIDATION_ERROR", message).
			WithRetryable(false),
		Field:      "", // Multiple fields
		Value:      "", // Multiple values
		Constraint: "", // Multiple constraints
		Violations: violations,
	}

	// Add violations to details
	violationDetails := make([]map[string]any, len(violations))
	for i, v := range violations {
		violationDetails[i] = map[string]any{
			"field":      v.Field,
			"code":       v.Code,
			"message":    v.Message,
			"value":      sanitizeValidationValue(fmt.Sprintf("%v", v.Value)),
			"constraint": v.Constraint,
		}
	}
	err.WithDetail("violations", violationDetails)

	return err
}

// AddViolation adds a validation violation to the error.
func (e *ValidationError) AddViolation(field, code, message string, value, constraint any) {
	violation := ValidationViolation{
		Field:      field,
		Code:       code,
		Message:    message,
		Value:      sanitizeValidationValue(fmt.Sprintf("%v", value)),
		Constraint: constraint,
	}
	e.Violations = append(e.Violations, violation)

	// Update the error message for multiple violations
	if len(e.Violations) > 1 {
		e.message = fmt.Sprintf("Validation failed for %d fields", len(e.Violations))
	}
}

// RequestValidationError represents validation errors for API request parameters.
type RequestValidationError struct {
	*ValidationError
	RequestType string // The type of request (e.g., "chat_completion", "text_completion")
	Method      string // HTTP method
	Endpoint    string // API endpoint
}

// NewRequestValidationError creates a new request validation error.
func NewRequestValidationError(requestType, method, endpoint string, violations []ValidationViolation) *RequestValidationError {
	message := fmt.Sprintf("Invalid %s request", requestType)
	if len(violations) > 0 {
		if len(violations) == 1 {
			message = fmt.Sprintf("Invalid %s request: %s", requestType, violations[0].Message)
		} else {
			message = fmt.Sprintf("Invalid %s request: %d validation errors", requestType, len(violations))
		}
	}

	validationErr := NewValidationErrorWithViolations(violations)
	validationErr.message = message
	validationErr.code = "REQUEST_VALIDATION_ERROR"

	err := &RequestValidationError{
		ValidationError: validationErr,
		RequestType:     requestType,
		Method:          method,
		Endpoint:        endpoint,
	}

	err.WithDetail("request_type", requestType).
		WithDetail("method", method).
		WithDetail("endpoint", endpoint)

	return err
}

// ResponseValidationError represents validation errors for API responses.
type ResponseValidationError struct {
	*ValidationError
	ResponseType string // The expected response type
	StatusCode   int    // HTTP status code of the response
}

// NewResponseValidationError creates a new response validation error.
func NewResponseValidationError(responseType string, statusCode int, violations []ValidationViolation) *ResponseValidationError {
	message := fmt.Sprintf("Invalid %s response", responseType)
	if len(violations) > 0 {
		if len(violations) == 1 {
			message = fmt.Sprintf("Invalid %s response: %s", responseType, violations[0].Message)
		} else {
			message = fmt.Sprintf("Invalid %s response: %d validation errors", responseType, len(violations))
		}
	}

	validationErr := NewValidationErrorWithViolations(violations)
	validationErr.message = message
	validationErr.code = "RESPONSE_VALIDATION_ERROR"
	validationErr.severity = SeverityHigh // Response validation failures are more serious

	err := &ResponseValidationError{
		ValidationError: validationErr,
		ResponseType:    responseType,
		StatusCode:      statusCode,
	}

	err.WithDetail("response_type", responseType).
		WithDetail("status_code", statusCode)

	return err
}

// ParameterValidationError represents validation errors for specific parameters.
type ParameterValidationError struct {
	*ValidationError
	ParameterName string      // Name of the parameter
	ParameterType string      // Expected type of the parameter
	MinValue      any // Minimum allowed value (if applicable)
	MaxValue      any // Maximum allowed value (if applicable)
	AllowedValues []string    // List of allowed values (if applicable)
}

// NewParameterValidationError creates a new parameter validation error.
func NewParameterValidationError(paramName, paramType string, value any, constraint string) *ParameterValidationError {
	sanitizedValue := sanitizeValidationValue(fmt.Sprintf("%v", value))
	message := fmt.Sprintf("Invalid parameter '%s': %s", paramName, constraint)

	validationErr := NewValidationError(paramName, sanitizedValue, constraint, message)
	validationErr.code = "PARAMETER_VALIDATION_ERROR"

	err := &ParameterValidationError{
		ValidationError: validationErr,
		ParameterName:   paramName,
		ParameterType:   paramType,
	}

	err.WithDetail("parameter_name", paramName).
		WithDetail("parameter_type", paramType)

	return err
}

// WithMinValue sets the minimum allowed value constraint.
func (e *ParameterValidationError) WithMinValue(min any) *ParameterValidationError {
	e.MinValue = min
	e.WithDetail("min_value", min)
	return e
}

// WithMaxValue sets the maximum allowed value constraint.
func (e *ParameterValidationError) WithMaxValue(max any) *ParameterValidationError {
	e.MaxValue = max
	e.WithDetail("max_value", max)
	return e
}

// WithAllowedValues sets the list of allowed values.
func (e *ParameterValidationError) WithAllowedValues(values []string) *ParameterValidationError {
	e.AllowedValues = values
	e.WithDetail("allowed_values", values)
	return e
}

// SchemaValidationError represents JSON schema validation errors.
type SchemaValidationError struct {
	*ValidationError
	SchemaPath string // JSON schema path that failed
	SchemaRule string // The schema rule that was violated
}

// NewSchemaValidationError creates a new schema validation error.
func NewSchemaValidationError(schemaPath, schemaRule, message string) *SchemaValidationError {
	if message == "" {
		message = fmt.Sprintf("Schema validation failed at '%s': %s", schemaPath, schemaRule)
	}

	validationErr := NewValidationError(schemaPath, "", schemaRule, message)
	validationErr.code = "SCHEMA_VALIDATION_ERROR"

	err := &SchemaValidationError{
		ValidationError: validationErr,
		SchemaPath:      schemaPath,
		SchemaRule:      schemaRule,
	}

	err.WithDetail("schema_path", schemaPath).
		WithDetail("schema_rule", schemaRule)

	return err
}

// Validation helper functions

// ValidateRequired checks if a required field is present and not empty.
func ValidateRequired(field string, value any) *ValidationViolation {
	if value == nil {
		return &ValidationViolation{
			Field:   field,
			Code:    "required",
			Message: fmt.Sprintf("Field '%s' is required", field),
			Value:   nil,
		}
	}

	// Check for empty strings
	if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
		return &ValidationViolation{
			Field:   field,
			Code:    "required",
			Message: fmt.Sprintf("Field '%s' cannot be empty", field),
			Value:   sanitizeValidationValue(str),
		}
	}

	return nil
}

// ValidateStringLength validates string length constraints.
func ValidateStringLength(field string, value string, minLength, maxLength int) *ValidationViolation {
	length := len(value)

	if minLength > 0 && length < minLength {
		return &ValidationViolation{
			Field:      field,
			Code:       "min_length",
			Message:    fmt.Sprintf("Field '%s' must be at least %d characters long", field, minLength),
			Value:      sanitizeValidationValue(value),
			Constraint: minLength,
		}
	}

	if maxLength > 0 && length > maxLength {
		return &ValidationViolation{
			Field:      field,
			Code:       "max_length",
			Message:    fmt.Sprintf("Field '%s' must be at most %d characters long", field, maxLength),
			Value:      sanitizeValidationValue(value),
			Constraint: maxLength,
		}
	}

	return nil
}

// ValidateNumericRange validates numeric range constraints.
func ValidateNumericRange(field string, value, min, max float64) *ValidationViolation {
	if value < min {
		return &ValidationViolation{
			Field:      field,
			Code:       "min_value",
			Message:    fmt.Sprintf("Field '%s' must be at least %v", field, min),
			Value:      value,
			Constraint: min,
		}
	}

	if value > max {
		return &ValidationViolation{
			Field:      field,
			Code:       "max_value",
			Message:    fmt.Sprintf("Field '%s' must be at most %v", field, max),
			Value:      value,
			Constraint: max,
		}
	}

	return nil
}

// ValidateEnum validates that a value is in a list of allowed values.
func ValidateEnum(field string, value string, allowedValues []string) *ValidationViolation {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &ValidationViolation{
		Field:      field,
		Code:       "invalid_enum",
		Message:    fmt.Sprintf("Field '%s' must be one of: %s", field, strings.Join(allowedValues, ", ")),
		Value:      sanitizeValidationValue(value),
		Constraint: allowedValues,
	}
}

// ValidateSliceLength validates slice/array length constraints.
func ValidateSliceLength(field string, slice []any, minLength, maxLength int) *ValidationViolation {
	length := len(slice)

	if minLength > 0 && length < minLength {
		return &ValidationViolation{
			Field:      field,
			Code:       "min_items",
			Message:    fmt.Sprintf("Field '%s' must have at least %d items", field, minLength),
			Value:      length,
			Constraint: minLength,
		}
	}

	if maxLength > 0 && length > maxLength {
		return &ValidationViolation{
			Field:      field,
			Code:       "max_items",
			Message:    fmt.Sprintf("Field '%s' must have at most %d items", field, maxLength),
			Value:      length,
			Constraint: maxLength,
		}
	}

	return nil
}

// sanitizeValidationValue sanitizes validation values to prevent exposing sensitive data.
func sanitizeValidationValue(value string) string {
	if value == "" {
		return "(empty)"
	}

	// Check if it looks like a secret
	if isLikelySecret(value) {
		if len(value) <= 8 {
			return "[REDACTED]"
		}
		return fmt.Sprintf("%s...[REDACTED]", value[:4])
	}

	// Truncate very long values
	if len(value) > 200 {
		return fmt.Sprintf("%s...(truncated)", value[:197])
	}

	return value
}

// CollectValidationViolations collects multiple validation violations into a single error.
func CollectValidationViolations(violations ...*ValidationViolation) *ValidationError {
	if len(violations) == 0 {
		return nil
	}

	// Filter out nil violations
	validViolations := make([]ValidationViolation, 0, len(violations))
	for _, v := range violations {
		if v != nil {
			validViolations = append(validViolations, *v)
		}
	}

	if len(validViolations) == 0 {
		return nil
	}

	return NewValidationErrorWithViolations(validViolations)
}
