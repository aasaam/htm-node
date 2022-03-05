package main

import (
	"testing"
)

func TestExecute(t *testing.T) {
	out, err1 := execute("date", "--rfc-email")
	if err1 != nil || len(out) < 1 {
		t.Error(err1)
	}
	_, err2 := execute("command-ensure-not-exist")
	if err2 == nil {
		t.Errorf("invalid command must return error")
	}

	err3 := executeMany("date --rfc-email", "echo '1'")
	if err3 != nil {
		t.Error(err1)
	}
}
