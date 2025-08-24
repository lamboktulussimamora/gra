# Validator range= Tag Guide

The validator supports inclusive numeric ranges using the `range=min,max` tag.

Supported field kinds:
- Signed integers (int, int8, int16, int32, int64)
- Unsigned integers (uint, uint8, uint16, uint32, uint64)
- Floats (float32, float64)

Usage examples:

```go
// Integers
Age int `json:"age" validate:"range=0,150|Age must be between 0 and 150"`

// Unsigned
Count uint `json:"count" validate:"range=1,100"`

// Floats
Price float64 `json:"price" validate:"range=0.01,9999.99|Invalid price"`
```

Notes:
- The range is inclusive: value must satisfy min <= v <= max
- min and max must be valid numbers for the field type
- For invalid specs (e.g., `range=a,b`), an error will be reported for that field

Tips:
- Combine with `required` when the field must be present: `validate:"required,range=1,10"`
- For strings, use `min`/`max` for length checks instead of `range`
