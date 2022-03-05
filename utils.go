package main

import (
	"net"
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

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPNets {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isValidPublicIP(str string) bool {
	ipAddress := net.ParseIP(str)
	return !isPrivateIP(ipAddress)
}

func isValidPublicCIDR(str string) bool {
	_, subnet, err := net.ParseCIDR(str)
	if err != nil {
		return false
	}
	return !isPrivateIP(subnet.IP)
}
