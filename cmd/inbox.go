package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox [flags]",
	Short: "Show unread messages",
	Long:  "Show unread messages for a recipient.",
	RunE: func(cmd *cobra.Command, args []string) error {
		recipient, _ := cmd.Flags().GetString("to")
		if recipient == "" {
			return fmt.Errorf("--to is required")
		}

		asJSON, _ := cmd.Flags().GetBool("json")

		q := db.New(database)
		msgs, err := q.GetInbox(cmd.Context(), recipient)
		if err != nil {
			return fmt.Errorf("get inbox: %w", err)
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(msgs)
		}

		if len(msgs) == 0 {
			fmt.Fprintf(os.Stdout, "no unread messages for %s\n", recipient)
			return nil
		}

		for _, m := range msgs {
			fmt.Fprintf(os.Stdout, "[%s] %s: %s\n", m.ID, m.Sender, m.Content)
		}
		return nil
	},
}

func init() {
	inboxCmd.Flags().String("to", "", "recipient to check inbox for (required)")
	_ = inboxCmd.MarkFlagRequired("to")
	inboxCmd.Flags().Bool("json", false, "output as JSON")
	rootCmd.AddCommand(inboxCmd)
}
