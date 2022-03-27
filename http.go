package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	memory "github.com/pbnjay/memory"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func errorResponse(c *fiber.Ctx, message string, code int) error {
	c.Status(code)
	return c.JSON(message)
}

func successResponse(c *fiber.Ctx, message string) error {
	c.Status(200)
	return c.JSON(message)
}

func successResponseInterface(c *fiber.Ctx, data interface{}) error {
	c.Status(200)
	return c.JSON(data)
}

func healthCheck() (bool, string) {
	out1, err1 := execute("docker", "inspect", "aasaam-web-server", "--format", "{{.State.Running}}")
	if err1 != nil {
		return false, "Web server container not found"
	}
	var stateRunning bool
	err2 := json.Unmarshal(out1, &stateRunning)
	if err2 != nil || !stateRunning {
		return false, "Web server container not running"
	}

	out3, err3 := execute("docker", "exec", "-t", "aasaam-web-server", "openresty", "-t")
	if err3 != nil {
		return false, string(err3.Error())
	}

	return true, "htm-node: healthy\n" + normalizeStd(out3)
}

func newHTTPServer(config *nodeConfig) *fiber.App {
	app := fiber.New(fiber.Config{

		DisableStartupMessage: true,
		Prefork:               false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			defer config.getLogger().
				Error().
				Str("error", err.Error()).
				Str("ip", c.IP()).
				Str("method", c.Method()).
				Str("url", c.Request().URI().String()).
				Int("status_code", code).
				Send()

			return errorResponse(c, "Internal Server Error", code)
		},
	})

	app.Use(recover.New())

	// middle ware
	app.Use(func(c *fiber.Ctx) error {
		token := c.Get(tokenHeader, "")
		if token != config.token {
			defer config.getLogger().
				Error().
				Str("ip", c.IP()).
				Str("method", c.Method()).
				Str("url", c.Request().URI().String()).
				Msg("Forbidden")
			return errorResponse(c, "Forbidden", 403)
		}

		defer config.getLogger().
			Info().
			Str("ip", c.IP()).
			Str("method", c.Method()).
			Str("url", c.Request().URI().String()).
			Msg("Access granted")

		return c.Next()
	})

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON("pong")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		healthy, message := healthCheck()
		if !healthy {
			return errorResponse(c, message, 500)
		}

		return successResponse(c, message)
	})

	app.Get("/info", func(c *fiber.Ctx) error {
		memoryInfoValue := memoryInfo{
			Total: totalMemory,
			Free:  memory.FreeMemory(),
		}
		info := nodeInfo{
			CPU:    cpuInfoValue,
			Memory: memoryInfoValue,
		}

		out1, err1 := executeString("uname", "-r")
		out2, err2 := executeString("lsb_release", "-sd")
		out3, err3 := execute("docker", "image", "inspect", "ghcr.io/aasaam/web-server:latest", "--format", "{{json .}}")
		out4, err4 := execute("docker", "image", "inspect", "ghcr.io/aasaam/nginx-protection:latest", "--format", "{{json .}}")
		out5, err5 := execute("docker", "image", "inspect", "ghcr.io/aasaam/rest-captcha:latest", "--format", "{{json .}}")
		out6, err6 := execute("docker", "image", "inspect", "ghcr.io/aasaam/nginx-error-log-parser:latest", "--format", "{{json .}}")

		if err1 != nil || err2 != nil {
			return errorResponse(c, "cannot get os information", 500)
		}

		if err3 != nil {
			return errorResponse(c, "cannot image 'web-server' information", 500)
		}

		if err4 != nil {
			return errorResponse(c, "cannot image 'nginx-protection' information", 500)
		}

		if err5 != nil {
			return errorResponse(c, "cannot image 'rest-captcha' information", 500)
		}

		if err6 != nil {
			return errorResponse(c, "cannot image 'nginx-error-log-parser' information", 500)
		}

		var imageWebServer dockerImage
		err7 := json.Unmarshal(out3, &imageWebServer)
		if err7 != nil {
			return errorResponse(c, "invalid image data for 'web-server'", 500)
		}

		var imageNginxProtection dockerImage
		err8 := json.Unmarshal(out4, &imageNginxProtection)
		if err8 != nil {
			return errorResponse(c, "invalid image data for 'nginx-protection'", 500)
		}

		var imageRESTCaptcha dockerImage
		err9 := json.Unmarshal(out5, &imageRESTCaptcha)
		if err9 != nil {
			return errorResponse(c, "invalid image data for 'rest-captcha'", 500)
		}

		var imageNginxErrorLogParser dockerImage
		err10 := json.Unmarshal(out6, &imageNginxErrorLogParser)
		if err10 != nil {
			return errorResponse(c, "invalid image data for 'nginx-error-log-parser'", 500)
		}

		info.Kernel = out1
		info.Distribution = out2
		info.ImageWebServer = imageWebServer
		info.ImageNginxProtection = imageNginxProtection
		info.ImageRESTCaptcha = imageRESTCaptcha
		info.ImageNginxErrorLogParser = imageNginxErrorLogParser

		return successResponseInterface(c, info)
	})

	app.Post("/firewall", func(c *fiber.Ctx) error {
		var firewallRequest firewallRequest
		if err1 := c.BodyParser(firewallRequest); err1 != nil {
			return errorResponse(c, "invalid firewall data", 422)
		}

		result := []string{}

		for _, ipString := range firewallRequest.IPs {
			if config.canBlockIP(ipString) {
				result = append(result, ipString)
			}
		}

		for _, cidrString := range firewallRequest.CIDRs {
			if config.canBlockCIDR(cidrString) {
				result = append(result, cidrString)
			}
		}

		err2 := ioutil.WriteFile(config.dockerPath+"/var/ufw_block", []byte(strings.Join(result, "\n")), 0644)
		if err2 != nil {
			return errorResponse(c, "failed to write to file: "+err2.Error(), 500)
		}

		_, err3 := execute("/usr/local/bin/firewall")
		if err3 != nil {
			return errorResponse(c, "failed to run firewall: "+err3.Error(), 500)
		}

		out, err4 := execute("ufw", "status", "numbered")
		if err4 != nil {
			return errorResponse(c, "failed to check ufw status: "+err4.Error(), 500)
		}

		return successResponse(c, string(out))
	})

	app.Post("/restart", func(c *fiber.Ctx) error {
		err1 := executeMany("cd "+config.dockerPath, "docker-compose up -d")
		if err1 != nil {
			return errorResponse(c, err1.Error(), 500)
		}

		healthy2, message2 := healthCheck()
		if !healthy2 {
			return errorResponse(c, message2, 500)
		}

		return c.JSON(message2)
	})

	app.Post("/update", func(c *fiber.Ctx) error {

		file, err1 := c.FormFile("addon.tgz")

		if err1 != nil {
			return errorResponse(c, "file 'addon.tgz' not present: "+err1.Error(), 400)
		}

		executeMany("rm -rf /tmp/addon.tgz")
		err2 := c.SaveFile(file, "/tmp/addon.tgz")
		if err2 != nil {
			return errorResponse(c, "cannot save 'addon.tgz': "+err2.Error(), 400)
		}

		executeMany("/usr/local/bin/htm-addon-backup")

		err3 := executeMany(
			"cd "+config.dockerPath,
			"cp /tmp/addon.tgz ./addon.tgz",
			"rm -rf addon/",
			"tar -xf addon.tgz",
			"rm -rf addon.tgz",
		)

		if err3 != nil {
			return errorResponse(c, "cannot extract 'addon.tgz': "+err3.Error(), 500)
		}

		healthy, message := healthCheck()
		if !healthy {
			return errorResponse(c, message, 500)
		}

		return successResponse(c, message)
	})

	return app
}
