package main

import (
	"regexp"
	"strings"
)

var returnNewLineRegex = regexp.MustCompile(`[\n\r]+`)

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

func normalizeStd(by []byte) string {
	s := returnNewLineRegex.ReplaceAllString(string(by), "\n")
	return strings.TrimSpace(s) + "\n"
}
