package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bo",
	Short: "Boule — local message bus",
	Long:  "Boule is a CLI message bus backed by SQLite for agent-to-agent and agent-to-human messaging.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
