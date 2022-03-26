package main

import (
	"strings"
)

func getTLSVersion(inp string) string {
	s := strings.ToLower(inp)
	if s == tlsVersionLegacy {
		return tlsVersionLegacy
	}
	if s == tlsVersionModern {
		return tlsVersionModern
	}
	return tlsVersionIntermediate
}
