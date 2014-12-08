package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/conduit/manager"
	log "github.com/sirupsen/logrus"
)

const (
	VERSION = "0.0.1"
)

func run(c *cli.Context) {

	m, err := manager.NewManager(c.StringSlice("repo"),
		c.String("docker"), c.String("auth-username"), c.String("auth-password"),
		c.String("auth-email"), c.Bool("debug"))
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func main() {
	app := cli.NewApp()

	app.Name = "conduit"
	app.Usage = "docker deployer"
	app.Version = VERSION
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "repo, r",
			Usage: "repo for whitelist",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "docker, d",
			Usage: "URL to Docker",
			Value: "unix:///var/run/docker.sock",
		},
		cli.StringFlag{
			Name:  "auth-username, u",
			Usage: "docker auth username (optional)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "auth-password, p",
			Usage: "docker auth password (optional)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "auth-email, e",
			Usage: "docker auth email (optional)",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "enable debug",
		},
	}

	app.Run(os.Args)

}
