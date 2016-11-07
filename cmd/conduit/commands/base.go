package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/conduit/handler"
	"github.com/spf13/cobra"
)

var (
	debug        bool
	repositories []string
	listenAddr   string
	dockerURL    string
	token        string
)

func init() {
	//logrus.SetFormatter(&simplelog.SimpleFormatter{})
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")
	RootCmd.PersistentFlags().StringVarP(&listenAddr, "listen", "l", ":8080", "Listen address")
	RootCmd.PersistentFlags().StringSliceVarP(&repositories, "repository", "r", []string{}, "Enable deployment for Docker repository (i.e. ehazlett/conduit)")
	RootCmd.PersistentFlags().StringVar(&dockerURL, "docker", "unix:///run/docker.sock", "Docker Engine URL")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "Token for hooks")
}

var RootCmd = &cobra.Command{
	Use:   "conduit",
	Short: "Docker Container Deployment",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debug("debug enabled")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(repositories) == 0 {
			cmd.Help()
			logrus.Fatal("you must specify at least one repository")
		}

		cfg := &handler.HandlerConfig{
			ListenAddr:   listenAddr,
			Repositories: repositories,
			Token:        token,
		}
		h, err := handler.New(cfg)
		if err != nil {
			logrus.Fatal(err)
		}

		if err := h.Run(); err != nil {
			logrus.Fatal(err)
		}
	},
}
