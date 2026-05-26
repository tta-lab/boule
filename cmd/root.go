package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/store"
)

var (
	db   *sql.DB
	dbPath string
)

var rootCmd = &cobra.Command{
	Use:   "bo",
	Short: "Boule — local message bus",
	Long:  "Boule is a CLI message bus backed by SQLite for agent-to-agent and agent-to-human messaging.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dbPath == "" {
			dbPath = store.DefaultPath()
		}

		var err error
		db, err = store.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database path (default: ~/.boule/boule.db)")
}
