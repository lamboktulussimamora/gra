// Package validator provides validation utilities for structs.
package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

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

// Validate validates a struct using tags
func (v *Validator) Validate(obj interface{}) []ValidationError {
	v.errors = []ValidationError{}
	v.validateStruct("", obj)
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// validateStruct recursively validates a struct using validate tags
func (v *Validator) validateStruct(prefix string, obj interface{}) {
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
			v.validateField(field, fieldName, rule)
		}
	}
}

// validateField validates a single field against a rule
func (v *Validator) validateField(field reflect.Value, fieldName, rule string) {
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
		v.validateRequired(field, fieldName)
	case "email":
		v.validateEmail(field, fieldName)
	case "min":
		v.validateMin(field, fieldName, ruleArg)
	case "max":
		v.validateMax(field, fieldName, ruleArg)
	}
}

// validateRequired checks if a field is not empty
func (v *Validator) validateRequired(field reflect.Value, fieldName string) {
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
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
		})
	}
}

// validateEmail checks if a field is a valid email
func (v *Validator) validateEmail(field reflect.Value, fieldName string) {
	if field.Kind() != reflect.String {
		return
	}

	email := field.String()
	if email == "" {
		return
	}

	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be a valid email address",
		})
	}
}

// validateMin checks if a field meets a minimum constraint
func (v *Validator) validateMin(field reflect.Value, fieldName, arg string) {
	switch field.Kind() {
	case reflect.String:
		min := 0
		fmt.Sscanf(arg, "%d", &min)
		if len(field.String()) < min {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at least %d characters", fieldName, min),
			})
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min := int64(0)
		fmt.Sscanf(arg, "%d", &min)
		if field.Int() < min {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at least %d", fieldName, min),
			})
		}
	case reflect.Float32, reflect.Float64:
		min := float64(0)
		fmt.Sscanf(arg, "%f", &min)
		if field.Float() < min {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at least %f", fieldName, min),
			})
		}
	}
}

// validateMax checks if a field meets a maximum constraint
func (v *Validator) validateMax(field reflect.Value, fieldName, arg string) {
	switch field.Kind() {
	case reflect.String:
		max := 0
		fmt.Sscanf(arg, "%d", &max)
		if len(field.String()) > max {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at most %d characters", fieldName, max),
			})
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max := int64(0)
		fmt.Sscanf(arg, "%d", &max)
		if field.Int() > max {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at most %d", fieldName, max),
			})
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max := uint64(0)
		fmt.Sscanf(arg, "%d", &max)
		if field.Uint() > max {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at most %d", fieldName, max),
			})
		}
	case reflect.Float32, reflect.Float64:
		max := float64(0)
		fmt.Sscanf(arg, "%f", &max)
		if field.Float() > max {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s must be at most %f", fieldName, max),
			})
		}
	}
}
