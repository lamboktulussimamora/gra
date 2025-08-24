package validator

import (
	"reflect"
	"testing"
)

// Covers: getCompiledRegexp cache hit/miss, invalid pattern path,
// fixPattern/addAnchorsIfNeeded, validateEnum with custom message,
// and direct range helpers for int/uint/float including parse errors.

func TestGetCompiledRegexp_CacheAndError(t *testing.T) {
	// miss -> compile
	r1, err := getCompiledRegexp("^abc$")
	if err != nil || r1 == nil {
		t.Fatalf("expected compiled regex, got err=%v r1=%v", err, r1)
	}
	// hit -> same pointer
	r2, err := getCompiledRegexp("^abc$")
	if err != nil || r2 == nil || r1 != r2 {
		t.Fatalf("expected cache hit with same pointer, err=%v r2=%v", err, r2)
	}

	// invalid pattern
	if _, err := getCompiledRegexp("["); err == nil {
		t.Fatalf("expected error for invalid regex pattern")
	}
}

func TestFixPattern_AndAnchors(t *testing.T) {
	// Known truncated username -> fixed
	if got := fixPattern(UsernamePatternPrefix); got != UsernamePattern {
		t.Fatalf("expected %q, got %q", UsernamePattern, got)
	}
	// Lowercase username without anchors -> add
	if got := fixPattern(LowercaseUsernamePattern); got != "^[a-z0-9_]{3,16}$" {
		t.Fatalf("expected anchored lowercase username, got %q", got)
	}
	// Phone number without anchors
	if got := fixPattern(PhoneNumberPattern); got != "^[0-9]{10}$" {
		t.Fatalf("expected anchored phone number, got %q", got)
	}
	// Generic char class gets anchors added
	if got := fixPattern("[A-Za-z]{2,4}"); got != "^[A-Za-z]{2,4}$" {
		t.Fatalf("expected anchors added, got %q", got)
	}
}

func TestValidateEnum_WithCustomMessage(t *testing.T) {
	v := New()
	type S struct {
		Code string `json:"code" validate:"enum=A,B|custom"`
	}
	errs := v.Validate(S{Code: "C"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %v", errs)
	}
	if errs[0].Message != "custom" || errs[0].Field != "code" {
		t.Fatalf("expected custom message on field code, got %+v", errs[0])
	}
}

func TestRangeHelpers_DirectCalls(t *testing.T) {
	v := New()

	// int in range
	type I struct {
		V int `json:"v"`
	}
	si := I{V: 5}
	rv := reflect.ValueOf(si)
	rf := rv.FieldByName("V")
	v.validateIntRange(rf, "v", "1", "10", "")
	if v.HasErrors() {
		t.Fatalf("unexpected errors: %v", v.errors)
	}

	// int out of range
	v.validateIntRange(rf, "v", "6", "10", "")
	if !v.HasErrors() {
		t.Fatalf("expected error for int range")
	}

	// uint in range
	v.errors = nil
	type U struct {
		V uint `json:"v"`
	}
	su := U{V: 3}
	rv = reflect.ValueOf(su)
	rf = rv.FieldByName("V")
	v.validateUintRange(rf, "v", "1", "5", "")
	if v.HasErrors() {
		t.Fatalf("unexpected errors: %v", v.errors)
	}

	// float in range
	v.errors = nil
	type F struct {
		V float64 `json:"v"`
	}
	sf := F{V: 1.5}
	rv = reflect.ValueOf(sf)
	rf = rv.FieldByName("V")
	v.validateFloatRange(rf, "v", "1.0", "2.0", "")
	if v.HasErrors() {
		t.Fatalf("unexpected errors: %v", v.errors)
	}

	// parse errors -> Invalid range values
	v.errors = nil
	v.validateIntRange(reflect.ValueOf(si), "v", "a", "b", "")
	if !v.HasErrors() {
		t.Fatalf("expected parse error for int range")
	}
}
