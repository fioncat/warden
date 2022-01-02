package pattern

import (
	"fmt"
	"path/filepath"
)

type Ignore []string

func (i Ignore) Validate() error {
	for _, pattern := range i {
		_, err := filepath.Match(pattern, "")
		if err != nil {
			return fmt.Errorf("validate ignore %s: %v", pattern, err)
		}
	}
	return nil
}

func (i Ignore) OneMatch(name string) bool {
	for _, pattern := range i {
		matched, _ := filepath.Match(pattern, name)
		if matched {
			return true
		}
	}
	return false
}
