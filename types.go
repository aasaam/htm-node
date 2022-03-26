package main

import (
	"time"

	cpu "github.com/klauspost/cpuid/v2"
	memory "github.com/pbnjay/memory"
)

const (
	tokenHeader = "X-Token"

	tlsVersionLegacy       = "legacy"
	tlsVersionModern       = "modern"
	tlsVersionIntermediate = "intermediate"
)

type firewallRequest struct {
	IPs   []string `json:"ips"`
	CIDRs []string `json:"cidrs"`
}

type dockerImage struct {
	ID          string    `json:"Id"`
	Created     time.Time `json:"Created"`
	Size        int       `json:"Size"`
	VirtualSize int       `json:"VirtualSize"`
}

type cpuInfo struct {
	BrandName      string `json:"brandname"`
	Vendor         string `json:"vendor"`
	PhysicalCores  int    `json:"physical_cores"`
	ThreadsPerCore int    `json:"threads_per_core"`
	LogicalCores   int    `json:"logical_cores"`
	Family         int    `json:"family"`
	Model          int    `json:"model"`
}

type memoryInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

type nodeInfo struct {
	Kernel                   string      `json:"kernel"`
	Distribution             string      `json:"distribution"`
	DockerVersion            string      `json:"docker_version"`
	CPU                      cpuInfo     `json:"cpu"`
	Memory                   memoryInfo  `json:"memory"`
	ImageWebServer           dockerImage `json:"image_web_server"`
	ImageNginxProtection     dockerImage `json:"image_nginx_protection"`
	ImageRESTCaptcha         dockerImage `json:"image_rest_captcha"`
	ImageNginxErrorLogParser dockerImage `json:"image_nginx_error_log_parser"`
}

var cpuInfoValue cpuInfo
var totalMemory uint64

func init() {
	cpuInfoValue = cpuInfo{
		BrandName:      cpu.CPU.BrandName,
		Family:         cpu.CPU.Family,
		Vendor:         cpu.CPU.VendorString,
		PhysicalCores:  cpu.CPU.PhysicalCores,
		LogicalCores:   cpu.CPU.LogicalCores,
		ThreadsPerCore: cpu.CPU.ThreadsPerCore,
		Model:          cpu.CPU.Model,
	}

	totalMemory = memory.TotalMemory()
}
