package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	memory "github.com/pbnjay/memory"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var sleepTest = time.Duration(3) * time.Second

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

func restoreLastSuccess() {
	executeMany(
		"rm -rf addon/",
		"tar -xf addon.success.tgz",
	)
}

func backupLastSuccess() {
	executeMany(
		"rm -rf addon.success.tgz",
		"tar -czf addon.success.tgz addon",
	)
}

func healthCheck(dockerPath string) (bool, string) {
	_, testError := execute("docker", "run", "--rm", "--name", "test", "-v", dockerPath+"/addon:/usr/local/openresty/nginx/addon", "ghcr.io/aasaam/web-server", "openresty", "-t")
	if testError != nil {
		return false, testError.Error()
	}

	execute("docker-compose", "down", "-d")

	_, composeUpErr := execute("docker-compose", "up", "-d")
	if composeUpErr != nil {
		return false, composeUpErr.Error()
	}

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
		return false, err3.Error()
	}

	return true, "htm-node: healthy\n" + normalizeStd(out3)
}

func newHTTPServer(config *nodeConfig) *fiber.App {

	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		panic(cwdErr)
	} else if cwd != config.dockerPath {
		panic(errors.New("invlaid path for htm: " + cwd))
	}

	app := fiber.New(fiber.Config{
		BodyLimit:             1024 * 1024 * 64,
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
				Str("url", c.Path()).
				Int("status_code", code).
				Send()

			return errorResponse(c, "Internal Server Error", code)
		},
	})

	app.Use(recover.New())

	// middle ware
	app.Use(func(c *fiber.Ctx) error {
		tokenString := c.Get(tokenHeader, "")
		token, tokenErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.token), nil
		})

		validationError := "invalid token"

		if tokenErr != nil {
			validationError = tokenErr.Error()
		} else {
			_, validationOK := token.Claims.(jwt.MapClaims)
			if validationOK && token.Valid {
				defer config.getLogger().
					Info().
					Str("ip", c.IP()).
					Str("method", c.Method()).
					Str("url", c.Path()).
					Msg("Access granted")
				return c.Next()
			}
		}

		defer config.getLogger().
			Error().
			Str("ip", c.IP()).
			Str("error", validationError).
			Str("method", c.Method()).
			Str("url", c.Path()).
			Msg("Forbidden")
		return errorResponse(c, "Forbidden: "+validationError, 403)
	})

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON("pong")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		isHealthy, isHealthyMessage := healthCheck(config.dockerPath)
		if !isHealthy {
			return errorResponse(c, isHealthyMessage, 500)
		}

		return c.JSON(isHealthyMessage)
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

		uname, unameErr := executeString("uname", "-r")
		lsb, lsbErr := executeString("lsb_release", "-sd")
		if unameErr != nil || lsbErr != nil {
			return errorResponse(c, "cannot get os information", 500)
		}

		diWebServer, diWebServerErr := execute("docker", "image", "inspect", "ghcr.io/aasaam/web-server:latest", "--format", "{{json .}}")
		if diWebServerErr != nil {
			return errorResponse(c, "cannot image 'web-server' information", 500)
		}
		var imageWebServer dockerImage
		diWebServerJSONErr := json.Unmarshal(diWebServer, &imageWebServer)
		if diWebServerJSONErr != nil {
			return errorResponse(c, "invalid image data for 'web-server'", 500)
		}

		nginxProtection, nginxProtectionErr := execute("docker", "image", "inspect", "ghcr.io/aasaam/nginx-protection:latest", "--format", "{{json .}}")
		if nginxProtectionErr != nil {
			return errorResponse(c, "cannot image 'nginx-protection' information", 500)
		}
		var imageNginxProtection dockerImage
		nginxProtectionJSONErr := json.Unmarshal(nginxProtection, &imageNginxProtection)
		if nginxProtectionJSONErr != nil {
			return errorResponse(c, "invalid image data for 'nginx-protection'", 500)
		}

		restCaptcha, restCaptchaErr := execute("docker", "image", "inspect", "ghcr.io/aasaam/rest-captcha:latest", "--format", "{{json .}}")
		if restCaptchaErr != nil {
			return errorResponse(c, "cannot image 'rest-captcha' information", 500)
		}
		var imageRESTCaptcha dockerImage
		restCaptchaJSONErr := json.Unmarshal(restCaptcha, &imageRESTCaptcha)
		if restCaptchaJSONErr != nil {
			return errorResponse(c, "invalid image data for 'rest-captcha'", 500)
		}

		nginxErrorLogParser, nginxErrorLogParserErr := execute("docker", "image", "inspect", "ghcr.io/aasaam/nginx-error-log-parser:latest", "--format", "{{json .}}")
		if nginxErrorLogParserErr != nil {
			return errorResponse(c, "cannot image 'nginx-error-log-parser' information", 500)
		}

		var imageNginxErrorLogParser dockerImage
		nginxErrorLogParserJSONErr := json.Unmarshal(nginxErrorLogParser, &imageNginxErrorLogParser)
		if nginxErrorLogParserJSONErr != nil {
			return errorResponse(c, "invalid image data for 'nginx-error-log-parser'", 500)
		}

		info.Kernel = uname
		info.Distribution = lsb
		info.ImageWebServer = imageWebServer
		info.ImageNginxProtection = imageNginxProtection
		info.ImageRESTCaptcha = imageRESTCaptcha
		info.ImageNginxErrorLogParser = imageNginxErrorLogParser

		return successResponseInterface(c, info)
	})

	app.Post("/firewall", func(c *fiber.Ctx) error {
		var firewallRequest firewallRequest
		if jsonError := c.BodyParser(firewallRequest); jsonError != nil {
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

		writeFileErr := ioutil.WriteFile(config.dockerPath+"/var/ufw_block", []byte(strings.Join(result, "\n")), 0644)
		if writeFileErr != nil {
			return errorResponse(c, "failed to write to file: "+writeFileErr.Error(), 500)
		}

		_, runFireWallErr := execute("/usr/local/bin/firewall")
		if runFireWallErr != nil {
			return errorResponse(c, "failed to run firewall: "+runFireWallErr.Error(), 500)
		}

		ufw, ufwErr := execute("ufw", "status", "numbered")
		if ufwErr != nil {
			return errorResponse(c, "failed to check ufw status: "+ufwErr.Error(), 500)
		}

		return successResponse(c, string(ufw))
	})

	app.Post("/restart", func(c *fiber.Ctx) error {
		_, composeDownErr := execute("docker-compose", "down")
		if composeDownErr != nil {
			restoreLastSuccess()
			return errorResponse(c, composeDownErr.Error(), 500)
		}

		_, composeUpErr := execute("docker-compose", "up", "-d")
		if composeUpErr != nil {
			restoreLastSuccess()
			return errorResponse(c, composeUpErr.Error(), 500)
		}

		isHealthy, isHealthyMessage := healthCheck(config.dockerPath)
		if !isHealthy {
			restoreLastSuccess()
			return errorResponse(c, isHealthyMessage, 500)
		}

		backupLastSuccess()

		return c.JSON(isHealthyMessage)
	})

	app.Post("/update", func(c *fiber.Ctx) error {
		file, fileErr := c.FormFile("addon.tgz")
		if fileErr != nil {
			return errorResponse(c, "file 'addon.tgz' not present: "+fileErr.Error(), fiber.StatusFailedDependency)
		}

		executeMany("rm -rf /tmp/addon.tgz")
		saveErr := c.SaveFile(file, "/tmp/addon.tgz")
		if saveErr != nil {
			return errorResponse(c, "cannot save 'addon.tgz': "+saveErr.Error(), fiber.StatusLoopDetected)
		}

		executeMany("/usr/local/bin/htm-addon-backup")
		extractErr := executeMany(
			"cd "+config.dockerPath,
			"cp /tmp/addon.tgz ./addon.tgz",
			"rm -rf addon/",
			"tar -xf addon.tgz",
			"rm -rf addon.tgz",
		)

		if extractErr != nil {
			return errorResponse(c, "cannot extract 'addon.tgz': "+extractErr.Error(), fiber.StatusInternalServerError)
		}

		isHealthy, isHealthyMessage := healthCheck(config.dockerPath)
		if !isHealthy {
			return errorResponse(c, isHealthyMessage, 500)
		}

		return c.JSON(isHealthyMessage)
	})

	return app
}
