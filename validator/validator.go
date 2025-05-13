// Package validator provides validation utilities for structs.
package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// Common validation patterns
var (
	// EmailRegex is a regex pattern for validating email addresses
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// regexpCache caches compiled regular expressions to improve performance
var regexpCache = make(map[string]*regexp.Regexp)
var regexpCacheMutex sync.RWMutex

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validator validates structs based on validate tags
type Validator struct {
	errors []ValidationError
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		errors: []ValidationError{},
	}
}

// addError adds a validation error with support for custom message
func (v *Validator) addError(field, defaultMsg, customMsg string) {
	message := defaultMsg
	if customMsg != "" {
		message = customMsg
	}

	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Validate validates a struct using tags
func (v *Validator) Validate(obj any) []ValidationError {
	v.errors = []ValidationError{}
	v.validateStruct("", obj)
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// validateStruct recursively validates a struct using validate tags
func (v *Validator) validateStruct(prefix string, obj any) {
	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if fieldType.Anonymous {
			// Handle embedded struct
			v.validateStruct(prefix, field.Interface())
			continue
		}

		// Get json tag name
		tag := fieldType.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Get the field name from tag
		fieldName := strings.Split(tag, ",")[0]
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		// Get validate tag
		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// Nested struct validation
		if field.Kind() == reflect.Struct {
			v.validateStruct(fieldName, field.Interface())
			continue
		}

		// Slice validation for struct elements
		if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Struct {
			// Validate each item in the slice
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)
				itemFieldName := fmt.Sprintf("%s[%d]", fieldName, j)
				v.validateStruct(itemFieldName, item.Interface())
			}
		}

		// Validate by rules
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			// Check for custom error message
			parts := strings.Split(rule, "|")
			ruleText := parts[0]

			var customMessage string
			if len(parts) > 1 {
				customMessage = parts[1]
			}

			v.validateField(field, fieldName, ruleText, customMessage)
		}
	}
}

// validateField validates a single field against a rule
func (v *Validator) validateField(field reflect.Value, fieldName, rule, customMessage string) {
	// Parse rule and arguments
	parts := strings.SplitN(rule, "=", 2)
	ruleName := parts[0]

	var ruleArg string
	if len(parts) > 1 {
		ruleArg = parts[1]
	}

	// Apply the rule
	switch ruleName {
	case "required":
		v.validateRequired(field, fieldName, customMessage)
	case "email":
		v.validateEmail(field, fieldName, customMessage)
	case "min":
		v.validateMin(field, fieldName, ruleArg, customMessage)
	case "max":
		v.validateMax(field, fieldName, ruleArg, customMessage)
	case "regexp":
		v.validateRegexp(field, fieldName, ruleArg, customMessage)
	case "enum":
		v.validateEnum(field, fieldName, ruleArg, customMessage)
	case "range":
		v.validateRange(field, fieldName, ruleArg, customMessage)
	}
}

// validateRequired checks if a field is not empty
func (v *Validator) validateRequired(field reflect.Value, fieldName, customMessage string) {
	isValid := true

	switch field.Kind() {
	case reflect.String:
		isValid = field.String() != ""
	case reflect.Ptr, reflect.Slice, reflect.Map:
		isValid = !field.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		isValid = field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		isValid = field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		isValid = field.Float() != 0
	case reflect.Bool:
		isValid = field.Bool()
	}

	if !isValid {
		v.addError(fieldName, fieldName+" is required", customMessage)
	}
}

// validateEmail checks if a field is a valid email
func (v *Validator) validateEmail(field reflect.Value, fieldName, customMessage string) {
	if field.Kind() != reflect.String {
		return
	}

	email := field.String()
	if email == "" {
		return
	}

	if !EmailRegex.MatchString(email) {
		v.addError(fieldName, fieldName+" must be a valid email address", customMessage)
	}
}

// validateMin checks if a field meets a minimum constraint
func (v *Validator) validateMin(field reflect.Value, fieldName, arg, customMessage string) {
	switch field.Kind() {
	case reflect.String:
		min := 0
		fmt.Sscanf(arg, "%d", &min)
		if len(field.String()) < min {
			v.addError(fieldName, fmt.Sprintf("%s must be at least %d characters", fieldName, min), customMessage)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min := int64(0)
		fmt.Sscanf(arg, "%d", &min)
		if field.Int() < min {
			v.addError(fieldName, fmt.Sprintf("%s must be at least %d", fieldName, min), customMessage)
		}
	case reflect.Float32, reflect.Float64:
		min := float64(0)
		fmt.Sscanf(arg, "%f", &min)
		if field.Float() < min {
			v.addError(fieldName, fmt.Sprintf("%s must be at least %f", fieldName, min), customMessage)
		}
	}
}

// validateMax checks if a field meets a maximum constraint
func (v *Validator) validateMax(field reflect.Value, fieldName, arg, customMessage string) {
	switch field.Kind() {
	case reflect.String:
		max := 0
		fmt.Sscanf(arg, "%d", &max)
		if len(field.String()) > max {
			v.addError(fieldName, fmt.Sprintf("%s must be at most %d characters", fieldName, max), customMessage)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max := int64(0)
		fmt.Sscanf(arg, "%d", &max)
		if field.Int() > max {
			v.addError(fieldName, fmt.Sprintf("%s must be at most %d", fieldName, max), customMessage)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max := uint64(0)
		fmt.Sscanf(arg, "%d", &max)
		if field.Uint() > max {
			v.addError(fieldName, fmt.Sprintf("%s must be at most %d", fieldName, max), customMessage)
		}
	case reflect.Float32, reflect.Float64:
		max := float64(0)
		fmt.Sscanf(arg, "%f", &max)
		if field.Float() > max {
			v.addError(fieldName, fmt.Sprintf("%s must be at most %f", fieldName, max), customMessage)
		}
	}
}

// validateRegexp checks if a field matches a regular expression pattern
func (v *Validator) validateRegexp(field reflect.Value, fieldName, pattern, customMessage string) {
	if field.Kind() != reflect.String {
		return
	}

	value := field.String()
	if value == "" {
		return
	}

	// Get or compile regex pattern
	var regex *regexp.Regexp
	var err error

	// Use a mutex to safely access the cache
	regexpCacheMutex.Lock()
	regex, exists := regexpCache[pattern]
	if !exists {
		regex, err = regexp.Compile(pattern)
		if err == nil {
			regexpCache[pattern] = regex
		}
	}
	regexpCacheMutex.Unlock()

	if err != nil {
		// If the pattern is invalid, add an error about the validation itself
		v.addError(fieldName, fmt.Sprintf("Invalid validation pattern for %s", fieldName), customMessage)
		return
	}

	if !regex.MatchString(value) {
		v.addError(fieldName, fmt.Sprintf("%s has an invalid format", fieldName), customMessage)
	}
}

// validateEnum checks if a field value is one of the allowed values
func (v *Validator) validateEnum(field reflect.Value, fieldName, allowedValues, customMessage string) {
	// Only apply to string fields
	if field.Kind() != reflect.String {
		return
	}

	value := field.String()
	if value == "" {
		return
	}

	// Split the allowed values by comma
	allowed := strings.Split(allowedValues, ",")

	// Check if the value is in the allowed list
	for _, allowedValue := range allowed {
		if value == strings.TrimSpace(allowedValue) {
			return // Value is allowed
		}
	}

	// Value is not in the allowed list
	v.addError(fieldName, fmt.Sprintf("%s must be one of: %s", fieldName, allowedValues), customMessage)
}

// validateRange checks if a field value falls within a specified numeric range
func (v *Validator) validateRange(field reflect.Value, fieldName, rangeValues, customMessage string) {
	// Parse min,max values
	rangeParts := strings.Split(rangeValues, ",")
	if len(rangeParts) != 2 {
		v.addError(fieldName, fmt.Sprintf("Invalid range specification for %s", fieldName), customMessage)
		return
	}

	minStr, maxStr := strings.TrimSpace(rangeParts[0]), strings.TrimSpace(rangeParts[1])

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min, err1 := parseInt(minStr)
		max, err2 := parseInt(maxStr)

		if err1 != nil || err2 != nil {
			v.addError(fieldName, fmt.Sprintf("Invalid range values for %s", fieldName), customMessage)
			return
		}

		value := field.Int()
		if value < min || value > max {
			v.addError(fieldName, fmt.Sprintf("%s must be between %d and %d", fieldName, min, max), customMessage)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, err1 := parseUint(minStr)
		max, err2 := parseUint(maxStr)

		if err1 != nil || err2 != nil {
			v.addError(fieldName, fmt.Sprintf("Invalid range values for %s", fieldName), customMessage)
			return
		}

		value := field.Uint()
		if value < min || value > max {
			v.addError(fieldName, fmt.Sprintf("%s must be between %d and %d", fieldName, min, max), customMessage)
		}

	case reflect.Float32, reflect.Float64:
		min, err1 := parseFloat(minStr)
		max, err2 := parseFloat(maxStr)

		if err1 != nil || err2 != nil {
			v.addError(fieldName, fmt.Sprintf("Invalid range values for %s", fieldName), customMessage)
			return
		}

		value := field.Float()
		if value < min || value > max {
			v.addError(fieldName, fmt.Sprintf("%s must be between %f and %f", fieldName, min, max), customMessage)
		}
	}
}

// Helper functions for parsing numbers
func parseInt(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseUint(s string) (uint64, error) {
	var result uint64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// BatchResult contains validation results for a batch of objects
type BatchResult struct {
	Index  int               `json:"index"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidateBatch validates a slice of objects and returns validation results
func (v *Validator) ValidateBatch(objects []any) []BatchResult {
	results := make([]BatchResult, len(objects))

	for i, obj := range objects {
		errors := v.Validate(obj)
		results[i] = BatchResult{
			Index:  i,
			Errors: errors,
		}
	}

	return results
}

// HasAnyErrors returns true if any object in the batch has validation errors
func (v *Validator) HasBatchErrors(results []BatchResult) bool {
	for _, result := range results {
		if len(result.Errors) > 0 {
			return true
		}
	}
	return false
}

// FilterInvalid returns only the batch results that have validation errors
func (v *Validator) FilterInvalid(results []BatchResult) []BatchResult {
	invalid := []BatchResult{}
	for _, result := range results {
		if len(result.Errors) > 0 {
			invalid = append(invalid, result)
		}
	}
	return invalid
}

// SchemaField represents a field in a validation schema
type SchemaField struct {
	Type      string // string, number, boolean, array, object
	Required  bool
	MinLength int
	MaxLength int
	Min       float64
	Max       float64
	Pattern   string
	Enum      []string
}

// Schema represents a validation schema
type Schema struct {
	Fields map[string]SchemaField
}

// NewSchema creates a new validation schema
func NewSchema() *Schema {
	return &Schema{
		Fields: make(map[string]SchemaField),
	}
}

// AddField adds a field to the schema
func (s *Schema) AddField(name string, field SchemaField) *Schema {
	s.Fields[name] = field
	return s
}

// Validate validates data against the schema
func (s *Schema) Validate(data map[string]any) []ValidationError {
	errors := []ValidationError{}

	for name, field := range s.Fields {
		value, exists := data[name]

		// Check required fields
		if field.Required && (!exists || value == nil) {
			errors = append(errors, ValidationError{
				Field:   name,
				Message: name + " is required",
			})
			continue
		}

		// Skip validation for non-existent optional fields
		if !exists || value == nil {
			continue
		}

		// Type validation
		if !s.validateType(name, value, field.Type, &errors) {
			continue // Skip further validation if type is wrong
		}

		// Field-specific validations based on type
		switch field.Type {
		case "string":
			s.validateString(name, value.(string), field, &errors)
		case "number":
			s.validateNumber(name, value, field, &errors)
		case "array":
			// Basic array validation, could be extended
			if arr, ok := value.([]any); ok && field.MinLength > 0 && len(arr) < field.MinLength {
				errors = append(errors, ValidationError{
					Field:   name,
					Message: fmt.Sprintf("%s must have at least %d items", name, field.MinLength),
				})
			}
		}
	}

	return errors
}

// validateType checks if a value matches the expected type
func (s *Schema) validateType(name string, value any, expectedType string, errors *[]ValidationError) bool {
	var valid bool

	switch expectedType {
	case "string":
		_, valid = value.(string)
	case "number":
		_, valid = value.(float64)
		if !valid {
			// Try integer types
			_, intValid := value.(int)
			_, int64Valid := value.(int64)
			valid = intValid || int64Valid
		}
	case "boolean":
		_, valid = value.(bool)
	case "object":
		_, valid = value.(map[string]any)
	case "array":
		_, valid = value.([]any)
	default:
		valid = true // Unknown type
	}

	if !valid {
		*errors = append(*errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("%s must be a %s", name, expectedType),
		})
	}

	return valid
}

// validateString validates a string value against string-specific rules
func (s *Schema) validateString(name, value string, field SchemaField, errors *[]ValidationError) {
	// Check min length
	if field.MinLength > 0 && len(value) < field.MinLength {
		*errors = append(*errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("%s must be at least %d characters", name, field.MinLength),
		})
	}

	// Check max length
	if field.MaxLength > 0 && len(value) > field.MaxLength {
		*errors = append(*errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("%s must be at most %d characters", name, field.MaxLength),
		})
	}

	// Check pattern
	if field.Pattern != "" {
		regex, err := regexp.Compile(field.Pattern)
		if err == nil && !regex.MatchString(value) {
			*errors = append(*errors, ValidationError{
				Field:   name,
				Message: fmt.Sprintf("%s has an invalid format", name),
			})
		}
	}

	// Check enum
	if len(field.Enum) > 0 {
		valid := false
		for _, enumValue := range field.Enum {
			if value == enumValue {
				valid = true
				break
			}
		}

		if !valid {
			*errors = append(*errors, ValidationError{
				Field:   name,
				Message: fmt.Sprintf("%s must be one of: %v", name, field.Enum),
			})
		}
	}
}

// validateNumber validates a numeric value against number-specific rules
func (s *Schema) validateNumber(name string, value any, field SchemaField, errors *[]ValidationError) {
	var floatVal float64

	switch v := value.(type) {
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	case float64:
		floatVal = v
	default:
		return // Should never happen as type is already checked
	}

	// Check minimum
	if field.Min != 0 && floatVal < field.Min {
		*errors = append(*errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("%s must be at least %v", name, field.Min),
		})
	}

	// Check maximum
	if field.Max != 0 && floatVal > field.Max {
		*errors = append(*errors, ValidationError{
			Field:   name,
			Message: fmt.Sprintf("%s must be at most %v", name, field.Max),
		})
	}
}
