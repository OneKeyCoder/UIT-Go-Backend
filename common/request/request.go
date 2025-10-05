package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
	"github.com/go-playground/validator/v10"
)

const maxBodySize = 1048576 // 1MB

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ReadJSON reads and validates JSON from the request body
func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit the size of the request body
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError
		
		switch {
		case errors.As(err, &syntaxError):
			return errors.New("malformed JSON")
		case errors.As(err, &unmarshalTypeError):
			return errors.New("invalid JSON type")
		case errors.As(err, &maxBytesError):
			return errors.New("request body too large")
		case errors.Is(err, io.EOF):
			return errors.New("request body is empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("malformed JSON")
		default:
			return err
		}
	}
	
	// Check for additional JSON values
	if decoder.More() {
		return errors.New("body must contain only a single JSON value")
	}
	
	return nil
}

// ReadAndValidate reads JSON and validates it using struct tags
func ReadAndValidate(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if err := ReadJSON(w, r, dst); err != nil {
		return err
	}
	
	return Validate(dst)
}

// Validate validates a struct using validation tags
func Validate(dst interface{}) error {
	if err := validate.Struct(dst); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			details := make([]response.ErrorDetail, 0, len(validationErrors))
			for _, err := range validationErrors {
				details = append(details, response.ErrorDetail{
					Field:   err.Field(),
					Message: getValidationMessage(err),
					Code:    err.Tag(),
				})
			}
			return &ValidationError{Details: details}
		}
		return err
	}
	
	return nil
}

// ValidationError represents validation errors
type ValidationError struct {
	Details []response.ErrorDetail
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// IsValidationError checks if error is a ValidationError and returns details
func IsValidationError(err error) ([]response.ErrorDetail, bool) {
	if err == nil {
		return nil, false
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		return nil, false
	}
	return valErr.Details, true
}

// HandleError handles common request errors with proper responses
// Returns true if error was handled, false if no error
func HandleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	
	details, ok := IsValidationError(err)
	if ok {
		response.ValidationError(w, details)
		return true
	}
	
	response.BadRequest(w, err.Error())
	return true
}

// getValidationMessage returns a human-readable validation error message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "gte":
		return "Value must be greater than or equal to " + err.Param()
	case "lte":
		return "Value must be less than or equal to " + err.Param()
	default:
		return "Invalid value"
	}
}

// GetValidator returns the validator instance for custom validations
func GetValidator() *validator.Validate {
	return validate
}
