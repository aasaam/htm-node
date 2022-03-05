package main

import (
	"strings"
	"testing"
)

func TestNode01(t *testing.T) {
	if getTLSVersion("a") != tlsVersionIntermediate {
		t.Errorf("invalid default")
	}
	if getTLSVersion(strings.ToUpper(tlsVersionLegacy)) != tlsVersionLegacy {
		t.Errorf("invalid string")
	}
	if getTLSVersion(strings.ToUpper(tlsVersionModern)) != tlsVersionModern {
		t.Errorf("invalid string")
	}

	c := nodeConfig{
		id:         "0",
		tlsVersion: tlsVersionIntermediate,
		dockerPath: "./test",
	}
	c.getLogger().Debug().Msg("ok")
	c.writeEnv()
}
