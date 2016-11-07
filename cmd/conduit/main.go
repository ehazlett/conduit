package main

import (
	"fmt"
	"os"

	"github.com/ehazlett/conduit/cmd/conduit/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
