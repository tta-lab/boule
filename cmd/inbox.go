package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox [flags] <recipient>",
	Short: "Show unread messages",
	Long:  "Show unread messages for a recipient.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		recipient := args[0]
		asJSON, _ := cmd.Flags().GetBool("json")

		q := db.New(database)
		msgs, err := q.GetInbox(cmd.Context(), recipient)
		if err != nil {
			return fmt.Errorf("get inbox: %w", err)
		}

		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(msgs)
		}

		if len(msgs) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "no unread messages for %s\n", recipient)
			return nil
		}

		for _, m := range msgs {
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", m.ID, m.Sender, m.Content)
		}
		return nil
	},
}

func init() {
	inboxCmd.Flags().Bool("json", false, "output as JSON")
	rootCmd.AddCommand(inboxCmd)
}
