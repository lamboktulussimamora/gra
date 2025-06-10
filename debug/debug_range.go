// Package main provides debug utilities for testing range validation functionality.
package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lamboktulussimamora/gra/validator"
)

func main() {
	type Simple struct {
		Age int `json:"age" validate:"range=10,20"`
	}

	v := validator.New()
	s := Simple{Age: 15}

	fmt.Printf("Testing struct: %+v\n", s)

	errors := v.Validate(s)
	fmt.Printf("Errors: %v\n", errors)

	// Let's also test the basic structure
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		validateTag := fieldType.Tag.Get("validate")

		fmt.Printf("Field %s: json='%s', validate='%s', value=%v, kind=%s\n",
			fieldType.Name, jsonTag, validateTag, field.Interface(), field.Kind())

		// Test manual range parsing
		if validateTag != "" {
			rules := strings.Split(validateTag, ",")
			for _, rule := range rules {
				if strings.HasPrefix(rule, "range=") {
					rangeValues := strings.TrimPrefix(rule, "range=")
					fmt.Printf("Range values: '%s'\n", rangeValues)

					rangeParts := strings.Split(rangeValues, ",")
					fmt.Printf("Range parts: %v (length: %d)\n", rangeParts, len(rangeParts))
				}
			}
		}
	}
}
