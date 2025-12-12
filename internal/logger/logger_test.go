package logger

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tmp := t.TempDir()
	_, closeFunc, err := New(tmp, true)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer closeFunc()

	if _, err := os.Stat(LogFile); os.IsNotExist(err) {
		t.Errorf("LogFile not created: %s", LogFile)
	}
}
