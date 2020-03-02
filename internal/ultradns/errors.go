package ultradns

import "fmt"

// ErrorResponse is a representation of the UltraDNS' API JSON error messages.
// API calls can return this type as an error.
// UltraDNS' API return values vary between snake and camel case. This attempts to handle that.
// TODO: move to common code.
type ErrorResponse struct {
	ErrorResponseI
	// Numerical code
	ErrorCodeCC int `json:"errorCode"`
	ErrorCodeSC int `json:"error_code"`
	// human-readable error message
	ErrorMessageCC string `json:"errorMessage"`
	ErrorMessageSC string `json:"error_message"`
	// Specific error type, e.g. 'unsupported_grant_type'
	ErrorTypeValue string `json:"error"`
	// ErrorCode + ErrorMessage
	ErrorDescriptionCC string `json:"errorDescription"`
	ErrorDescriptionSC string `json:"error_description"`
}

// ErrorResponseI is the error reponse interface
type ErrorResponseI interface {
	ErrorCode() int
	ErrorMessage() string
	ErrorType() string
	ErrorDescription() string
}

// ErrorCode returns the error code from the error response
// The code is a numerical representation. `0` means no error.
func (e *ErrorResponse) ErrorCode() int {
	if e.ErrorCodeCC > 0 {
		return e.ErrorCodeCC
	}
	return e.ErrorCodeSC
}

// ErrorMessage returns the error message from the error response
// The error message is a human-readable message
func (e *ErrorResponse) ErrorMessage() string {
	if e.ErrorMessageCC != "" {
		return e.ErrorMessageCC
	}
	return e.ErrorMessageSC
}

// ErrorType returns a string error type. This is useful for being able to
// get a more concise string to do switching on for custom error handling.
// This just returns the ErrorTypeValue field to provide a consistent API for this error struct.
func (e *ErrorResponse) ErrorType() string {
	return e.ErrorTypeValue
}

// ErrorDescription returns the error description from the error response
// The error description is typically a combination of the error code and the error message.
func (e *ErrorResponse) ErrorDescription() string {
	if e.ErrorDescriptionCC != "" {
		return e.ErrorDescriptionCC
	}
	return e.ErrorDescriptionSC
}

// Error is the interface for the error type.
func (e *ErrorResponse) Error() string {
	switch {
	case e.ErrorDescription() != "":
		return e.ErrorDescription()
	case e.ErrorMessage() != "":
		return fmt.Sprintf("%d: %s", e.ErrorCode(), e.ErrorMessage())
	default:
		panic(e)
	}
}
