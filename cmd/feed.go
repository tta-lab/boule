package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
)

var feedCmd = &cobra.Command{
	Use:   "feed [flags]",
	Short: "Show all messages with optional filters",
	Long:  "Show all messages. Filter by sender, recipient, or unread status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		unreadOnly, _ := cmd.Flags().GetBool("unread")
		asJSON, _ := cmd.Flags().GetBool("json")

		senderFilter := from
		recipientFilter := to
		var unreadFilter int64
		if unreadOnly {
			unreadFilter = 1
		}

		q := db.New(database)
		msgs, err := q.GetFeed(cmd.Context(), db.GetFeedParams{
			Column1:   senderFilter,
			Sender:    senderFilter,
			Column3:   recipientFilter,
			Recipient: recipientFilter,
			Column5:   unreadFilter,
			Limit:     100,
			Offset:    0,
		})
		if err != nil {
			return fmt.Errorf("get feed: %w", err)
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(msgs)
		}

		if len(msgs) == 0 {
			_, _ = fmt.Fprintln(os.Stdout, "no messages")
			return nil
		}

		for _, m := range msgs {
			readMark := " "
			if m.Read == 1 {
				readMark = "x"
			}
			fmt.Fprintf(os.Stdout, "[%s] [%s] %s -> %s: %s\n", m.ID, readMark, m.Sender, m.Recipient, m.Content)
		}
		return nil
	},
}

func init() {
	feedCmd.Flags().String("from", "", "filter by sender")
	feedCmd.Flags().String("to", "", "filter by recipient")
	feedCmd.Flags().Bool("unread", false, "show only unread messages")
	feedCmd.Flags().Bool("json", false, "output as JSON")
	rootCmd.AddCommand(feedCmd)
}
