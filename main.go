package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/conduit/manager"
	log "github.com/sirupsen/logrus"
)

const (
	VERSION = "0.0.2"
)

func run(c *cli.Context) {

	if len(c.StringSlice("repo")) == 0 {
		cli.ShowAppHelp(c)
		log.Fatal("you must specify at least 1 repo")
	}
	m, err := manager.NewManager(c.StringSlice("repo"),
		c.String("docker"), c.String("auth-username"), c.String("auth-password"),
		c.String("auth-email"), c.String("token"), c.Bool("debug"))
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func main() {
	app := cli.NewApp()

	app.Name = "conduit"
	app.Author = "@ehazlett"
	app.Email = ""
	app.Usage = "docker deployment system"
	app.Version = VERSION
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "repo, r",
			Usage: "repo for whitelist",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:   "docker, d",
			Usage:  "URL to Docker",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},
		cli.StringFlag{
			Name:   "auth-username, u",
			Usage:  "docker auth username (optional)",
			Value:  "",
			EnvVar: "DOCKER_AUTH_USERNAME",
		},
		cli.StringFlag{
			Name:   "auth-password, p",
			Usage:  "docker auth password (optional)",
			Value:  "",
			EnvVar: "DOCKER_AUTH_PASSWORD",
		},
		cli.StringFlag{
			Name:   "auth-email, e",
			Usage:  "docker auth email (optional)",
			Value:  "",
			EnvVar: "DOCKER_AUTH_EMAIL",
		},
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "webhook token",
			Value:  "",
			EnvVar: "TOKEN",
		},
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "enable debug",
		},
	}

	app.Run(os.Args)

}
