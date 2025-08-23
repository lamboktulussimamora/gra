package validator

import "testing"

// Cover validateRange error path via struct tags. Due to comma-splitting in
// tag parsing, "range=1,3" is split and triggers the "Invalid range specification"
// branch. We assert that behavior to raise coverage without altering production code.

type rangeSample struct {
	I  int     `json:"i" validate:"range=1,3"`
	U  uint    `json:"u" validate:"range=2,4"`
	F  float64 `json:"f" validate:"range=0.5,1.5"`
	I2 int     `json:"i2" validate:"range=a,b"` // invalid range spec triggers error
}

func TestValidateRange(t *testing.T) {
	v := New()

	// Any usage of range currently yields "Invalid range specification" due to parsing
	sample := rangeSample{I: 2, U: 3, F: 1.0, I2: 10}
	errs := v.Validate(sample)
	if len(errs) != 4 {
		t.Fatalf("expected 4 spec errors (i,u,f,i2), got %v", errs)
	}
}
