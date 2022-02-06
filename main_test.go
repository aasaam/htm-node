package main

import (
	"testing"
)

func TestNum(t *testing.T) {
	n1 := num()
	if n1 != 1 {
		t.Errorf("id must be unique")
	}
}
