package main

import (
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type nodeConfig struct {
	id           string
	tlsVersion   string
	token        string
	port         uint16
	logLevel     string
	dockerPath   string
	mangementIPs []net.IP

	logger *zerolog.Logger
}

func (c *nodeConfig) getLogger() *zerolog.Logger {
	if c.logger == nil {
		// logger config
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.SetGlobalLevel(zerolog.WarnLevel)

		logConfigLevel, errLogLevel := zerolog.ParseLevel(c.logLevel)
		if errLogLevel == nil {
			zerolog.SetGlobalLevel(logConfigLevel)
		}
		logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		c.logger = &logger
	}
	return c.logger
}

func (c *nodeConfig) setMangementIPs(ips string) {
	var lst []net.IP
	for _, ip := range strings.Split(ips, ",") {
		i := net.ParseIP(strings.TrimSpace(ip))
		if i != nil {
			lst = append(lst, net.ParseIP(ip))
		}
	}
	c.mangementIPs = lst
}

func (c *nodeConfig) canBlockIP(ipString string) bool {
	ip := net.ParseIP(ipString)
	if ip != nil {
		for _, mip := range c.mangementIPs {
			if mip.Equal(ip) {
				return false
			}
		}
	}
	return ip != nil
}

func (c *nodeConfig) canBlockCIDR(cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	for _, mip := range c.mangementIPs {
		if ipNet.Contains(mip) {
			return false
		}
	}
	return true
}

func (c *nodeConfig) writeEnv() error {
	envSample, err1 := ioutil.ReadFile(c.dockerPath + "/.env.2.ready")
	if err1 != nil {
		return err1
	}

	envSampleString := string(envSample)

	addons := []string{
		"ASM_NODE_ID=" + c.id,
		"ASM_SSL_PROFILE=" + c.tlsVersion,
		"",
	}

	envSampleString += "\n" + strings.Join(addons, "\n")

	return ioutil.WriteFile(c.dockerPath+"/.env", []byte(envSampleString), 0644)
}
