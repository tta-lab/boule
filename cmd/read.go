package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
)

var readCmd = &cobra.Command{
	Use:   "read [flags] <message-id>",
	Short: "Mark a message as read",
	Long:  "Mark a message as read by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		asJSON, _ := cmd.Flags().GetBool("json")

		q := db.New(database)
		if err := q.MarkRead(cmd.Context(), id); err != nil {
			return fmt.Errorf("mark read: %w", err)
		}

		msg, err := q.GetMessageByID(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("get message: %w", err)
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(msg)
		}

		fmt.Fprintf(os.Stdout, "marked as read: [%s] %s -> %s: %s\n", msg.ID, msg.Sender, msg.Recipient, msg.Content)
		return nil
	},
}

func init() {
	readCmd.Flags().Bool("json", false, "output as JSON")
	rootCmd.AddCommand(readCmd)
}
