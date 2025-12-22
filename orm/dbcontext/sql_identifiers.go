package dbcontext

import "fmt"

// validateSQLIdentifierPath validates a dot-separated identifier path (e.g. "col" or "t.col").
// It intentionally rejects spaces, quotes, commas, parentheses, and other SQL fragments.
func validateSQLIdentifierPath(value string) error {
	if value == "" {
		return fmt.Errorf("identifier must not be empty")
	}

	start := 0
	for {
		// Find next '.' separator
		end := start
		for end < len(value) && value[end] != '.' {
			end++
		}

		part := value[start:end]
		if part == "" {
			return fmt.Errorf("identifier contains empty segment")
		}
		if err := validateSQLIdentifier(part); err != nil {
			return err
		}

		if end == len(value) {
			break
		}
		start = end + 1
		if start >= len(value) {
			return fmt.Errorf("identifier ends with '.'")
		}
	}

	return nil
}

func validateSQLIdentifier(value string) error {
	if value == "" {
		return fmt.Errorf("identifier must not be empty")
	}

	for i := 0; i < len(value); i++ {
		b := value[i]
		isAlpha := (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
		isDigit := b >= '0' && b <= '9'
		isUnderscore := b == '_'

		if i == 0 {
			if !(isAlpha || isUnderscore) {
				return fmt.Errorf("identifier must start with a letter or underscore")
			}
			continue
		}

		if !(isAlpha || isDigit || isUnderscore) {
			return fmt.Errorf("identifier contains invalid character")
		}
	}

	return nil
}
