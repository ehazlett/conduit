package commands

import (
	"fmt"

	"github.com/ehazlett/conduit/version"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version",
	Long:  "Show version of the application and exit.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.FullVersion())
	},
}
