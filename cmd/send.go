package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
)

var sendCmd = &cobra.Command{
	Use:   "send [flags] <recipient>",
	Short: "Send a message (content from stdin)",
	Long:  "Send a message to a recipient. Content is read from stdin.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sender, _ := cmd.Flags().GetString("from")
		if sender == "" {
			return fmt.Errorf("--from is required")
		}

		recipient := args[0]

		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}

		if len(content) == 0 {
			return fmt.Errorf("message content is empty")
		}

		q := db.New(database)
		msg, err := q.SendMessage(cmd.Context(), db.SendMessageParams{
			ID:        uuid.New().String(),
			Sender:    sender,
			Recipient: recipient,
			Content:   string(content),
		})
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		fmt.Fprintf(os.Stdout, "sent %s -> %s (id: %s)\n", msg.Sender, msg.Recipient, msg.ID)
		return nil
	},
}

func init() {
	sendCmd.Flags().String("from", "", "sender identity (required)")
	sendCmd.MarkFlagRequired("from")
	rootCmd.AddCommand(sendCmd)
}
