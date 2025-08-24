package dbcontext

import (
	"reflect"
	"testing"
	"time"
)

func TestSetFieldValue_Primitives(t *testing.T) {
	var s struct {
		S string
		I int64
		U uint64
		F float64
		B bool
		T time.Time
	}

	// string from []byte
	fv := reflect.ValueOf(&s).Elem().FieldByName("S")
	setStringField(fv, []byte("abc"))
	if s.S != "abc" {
		t.Fatalf("string set failed: %q", s.S)
	}

	// int from string
	fi := reflect.ValueOf(&s).Elem().FieldByName("I")
	setIntField(fi, "123")
	if s.I != 123 {
		t.Fatalf("int set failed: %d", s.I)
	}

	// uint from string
	fu := reflect.ValueOf(&s).Elem().FieldByName("U")
	setUintField(fu, "456")
	if s.U != 456 {
		t.Fatalf("uint set failed: %d", s.U)
	}

	// float from string
	ff := reflect.ValueOf(&s).Elem().FieldByName("F")
	setFloatField(ff, "1.5")
	if s.F != 1.5 {
		t.Fatalf("float set failed: %f", s.F)
	}

	// bool from int64
	fb := reflect.ValueOf(&s).Elem().FieldByName("B")
	setBoolField(fb, int64(1))
	if !s.B {
		t.Fatalf("bool set failed")
	}

	// time from string
	ft := reflect.ValueOf(&s).Elem().FieldByName("T")
	setTimeField(ft, "2024-01-02 03:04:05")
	if s.T.IsZero() {
		t.Fatalf("time set failed")
	}
}
