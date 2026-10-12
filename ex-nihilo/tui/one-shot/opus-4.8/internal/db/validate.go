package db

import (
	"fmt"
	"regexp"
)

var alnum = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// ValidateName checks a proposed DB name.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !alnum.MatchString(name) {
		return fmt.Errorf("name must be alphanumeric only")
	}
	if Exists(name) {
		return fmt.Errorf("a database named %q already exists", name)
	}
	return nil
}
