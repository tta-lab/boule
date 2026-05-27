package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
)

var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List all known sender/receiver identities",
	Long:  "List all unique senders and recipients found in messages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")

		q := db.New(database)
		entities, err := q.ListEntities(cmd.Context())
		if err != nil {
			return fmt.Errorf("list entities: %w", err)
		}

		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(entities)
		}

		if len(entities) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no entities found")
			return nil
		}

		for _, e := range entities {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), e)
		}
		return nil
	},
}

func init() {
	entitiesCmd.Flags().Bool("json", false, "output as JSON")
	rootCmd.AddCommand(entitiesCmd)
}
