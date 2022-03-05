package main

import (
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
)

func runServer(c *cli.Context) error {

	cnf := nodeConfig{
		id:         c.String("id"),
		tlsVersion: getTLSVersion(c.String("tls-version")),
		port:       uint16(c.Int("port")),
		token:      c.String("token"),
		dockerPath: c.String("docker-path"),

		logLevel: c.String("log-level"),
	}

	if c.String("management-ips") != "" {
		cnf.setMangementIPs(c.String("management-ips"))
	}

	if !c.Bool("no-write-env") {
		err := cnf.writeEnv()
		if err != nil {
			panic(err)
		}
	}

	app := newHTTPServer(&cnf)
	return app.Listen("0.0.0.0:" + strconv.Itoa(int(cnf.port)))
}

func main() {
	app := cli.NewApp()
	app.Usage = "htm-node"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:  "run",
			Usage: "Run protection server",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "ID",
					Required: true,
					EnvVars:  []string{"ASM_HTM_NODE_ID"},
				},
				&cli.StringFlag{
					Name:     "token",
					Usage:    "Token",
					Required: true,
					EnvVars:  []string{"ASM_HTM_NODE_TOKEN"},
				},
				&cli.StringFlag{
					Name:    "tls-version",
					Usage:   "TLS Version",
					Value:   tlsVersionIntermediate,
					EnvVars: []string{"ASM_HTM_NODE_TLS_VERSION"},
				},
				&cli.StringFlag{
					Name:    "management-ips",
					Usage:   "List management server IP addressess",
					Value:   "127.0.0.1",
					EnvVars: []string{"ASM_HTM_MANAGEMENT_IP"},
				},
				&cli.IntFlag{
					Name:    "port",
					Usage:   "Port",
					Value:   9199,
					EnvVars: []string{"ASM_HTM_NODE_PORT"},
				},
				&cli.StringFlag{
					Name:    "docker-path",
					Usage:   "HTM docker path",
					Value:   "/opt/htm-docker",
					EnvVars: []string{"ASM_HTM_NODE_DOCKER_PATH"},
				},
				&cli.BoolFlag{
					Name:    "no-write-env",
					Usage:   "Do not write environment variable on start",
					Value:   false,
					EnvVars: []string{"ASM_HTM_WRITE_ENV"},
				},
				&cli.StringFlag{
					Name:    "log-level",
					Usage:   "Could be one of `panic`, `fatal`, `error`, `warn`, `info`, `debug` or `trace`",
					Value:   "warn",
					EnvVars: []string{"ASM_HTM_NODE_LOG_LEVEL"},
				},
			},
			Action: runServer,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
