package validator

import "testing"

// Validate proper range handling for int, uint, and float, plus an invalid spec.

type rangeSample struct {
	I  int     `json:"i" validate:"range=1,3"`
	U  uint    `json:"u" validate:"range=2,4"`
	F  float64 `json:"f" validate:"range=0.5,1.5"`
	I2 int     `json:"i2" validate:"range=a,b"` // invalid range spec triggers error
}

func TestValidateRange(t *testing.T) {
	v := New()

	// In-range values should pass with no errors for I, U, F.
	sample := rangeSample{I: 2, U: 3, F: 1.0, I2: 10}
	errs := v.Validate(sample)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for invalid spec (i2), got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "i2" {
		t.Fatalf("expected error on field 'i2', got %s", errs[0].Field)
	}

	// Out-of-range should error for each field plus the invalid spec (i2)
	sample2 := rangeSample{I: 0, U: 5, F: 2.0}
	errs2 := v.Validate(sample2)
	if len(errs2) != 4 { // i out (<1), u out (>4), f out (>1.5), and i2 invalid spec
		t.Fatalf("expected 4 errors (i,u,f out-of-range + i2 invalid spec), got %d: %v", len(errs2), errs2)
	}
}
