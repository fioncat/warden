package debug

import (
	"errors"
	"testing"
)

func TestDebug(t *testing.T) {
	Enable = true
	Info("Hello debug")
	Infof("Hello, %s", "warden")
	Error(errors.New("Bad happened"), "Failed to exec")
	Errorf("Something bad happened")
	Exec("sleep 20")
}
