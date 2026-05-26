package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tta-lab/boule/internal/db"
	"github.com/tta-lab/boule/internal/store"
)

const (
	testAlice   = "alice"
	testBob     = "bob"
	testCharlie = "charlie"
)

type testEnv struct {
	db     *sql.DB
	q      *db.Queries
	dbPath string
	root   *cobra.Command
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")

	d, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })

	env := &testEnv{db: d, q: db.New(d), dbPath: dbPath}
	env.root = env.buildRoot()
	return env
}

func (e *testEnv) buildRoot() *cobra.Command {
	root := &cobra.Command{Use: "bo"}
	root.AddCommand(
		e.buildSend(),
		e.buildInbox(),
		e.buildFeed(),
		e.buildRead(),
		e.buildEntities(),
	)
	return root
}

func (e *testEnv) buildSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "send [flags] <recipient>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sender, _ := cmd.Flags().GetString("from")
			if sender == "" {
				return fmt.Errorf("--from is required")
			}
			content, err := readAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			if len(content) == 0 {
				return fmt.Errorf("message content is empty")
			}
			msg, err := e.q.SendMessage(cmd.Context(), db.SendMessageParams{
				ID: fmt.Sprintf("test-%d", os.Getpid()), Sender: sender, Recipient: args[0], Content: string(content),
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "sent %s -> %s (id: %s)\n", msg.Sender, msg.Recipient, msg.ID)
			return nil
		},
	}
	cmd.Flags().String("from", "", "sender")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}

func (e *testEnv) buildInbox() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "inbox <recipient>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			asJSON, _ := cmd.Flags().GetBool("json")
			msgs, err := e.q.GetInbox(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(msgs)
			}
			if len(msgs) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "no unread messages for %s\n", args[0])
				return nil
			}
			for _, m := range msgs {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", m.ID, m.Sender, m.Content)
			}
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (e *testEnv) buildFeed() *cobra.Command {
	cmd := &cobra.Command{
		Use: "feed",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			unreadOnly, _ := cmd.Flags().GetBool("unread")
			asJSON, _ := cmd.Flags().GetBool("json")
			var unreadFilter int64
			if unreadOnly {
				unreadFilter = 1
			}
			msgs, err := e.q.GetFeed(cmd.Context(), db.GetFeedParams{
				Column1: from, Sender: from, Column3: "", Recipient: "",
				Column5: unreadFilter, Limit: 100, Offset: 0,
			})
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(msgs)
			}
			if len(msgs) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no messages")
				return nil
			}
			for _, m := range msgs {
				mark := " "
				if m.Read == 1 {
					mark = "x"
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] [%s] %s -> %s: %s\n", m.ID, mark, m.Sender, m.Recipient, m.Content)
			}
			return nil
		},
	}
	cmd.Flags().String("from", "", "filter by sender")
	cmd.Flags().Bool("unread", false, "only unread")
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (e *testEnv) buildRead() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "read <id>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			asJSON, _ := cmd.Flags().GetBool("json")
			if err := e.q.MarkRead(cmd.Context(), args[0]); err != nil {
				return fmt.Errorf("mark read: %w", err)
			}
			msg, err := e.q.GetMessageByID(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get message: %w", err)
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(msg)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(),
				"marked as read: [%s] %s -> %s: %s\n",
				msg.ID, msg.Sender, msg.Recipient, msg.Content)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (e *testEnv) buildEntities() *cobra.Command {
	cmd := &cobra.Command{
		Use: "entities",
		RunE: func(cmd *cobra.Command, args []string) error {
			asJSON, _ := cmd.Flags().GetBool("json")
			ents, err := e.q.ListEntities(cmd.Context())
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(ents)
			}
			if len(ents) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no entities found")
				return nil
			}
			for _, ent := range ents {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), ent)
			}
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (e *testEnv) run(args ...string) (string, error) {
	return e.runWithStdin("", args...)
}

func (e *testEnv) runWithStdin(stdin string, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	e.root.SetArgs(args)
	e.root.SetIn(strings.NewReader(stdin))
	e.root.SetOut(buf)
	e.root.SetErr(buf)
	err := e.root.Execute()
	return buf.String(), err
}

func readAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 256)
	for {
		n, err := r.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return buf, err
		}
	}
	return buf, nil
}

func TestSendCommand(t *testing.T) {
	env := newTestEnv(t)
	output, err := env.runWithStdin("hello world", "send", "--from", testAlice, testBob)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !strings.Contains(output, "sent alice -> bob") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestSendEmptyContent(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.runWithStdin("", "send", "--from", testAlice, testBob)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestInboxCommand(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.q.SendMessage(t.Context(), db.SendMessageParams{
		ID: "t1", Sender: testAlice, Recipient: testBob, Content: "inbox test",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	output, err := env.run("inbox", testBob)
	if err != nil {
		t.Fatalf("inbox failed: %v", err)
	}
	if !strings.Contains(output, "alice: inbox test") {
		t.Fatalf("expected message: %s", output)
	}
}

func TestInboxEmpty(t *testing.T) {
	env := newTestEnv(t)
	output, err := env.run("inbox", "nobody")
	if err != nil {
		t.Fatalf("inbox failed: %v", err)
	}
	if !strings.Contains(output, "no unread messages") {
		t.Fatalf("expected empty: %s", output)
	}
}

func TestInboxJSON(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.q.SendMessage(t.Context(), db.SendMessageParams{
		ID: "tj", Sender: testAlice, Recipient: testBob, Content: "json test",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	output, err := env.run("inbox", testBob, "--json")
	if err != nil {
		t.Fatalf("inbox --json failed: %v", err)
	}
	var msgs []db.Message
	if err := json.Unmarshal([]byte(output), &msgs); err != nil {
		t.Fatalf("invalid JSON: %v: raw=%s", err, output)
	}
	if len(msgs) != 1 || msgs[0].Content != "json test" {
		t.Fatalf("unexpected msgs: %+v", msgs)
	}
}

func TestFeedCommand(t *testing.T) {
	env := newTestEnv(t)
	seeds := []struct{ s, r, c string }{
		{testAlice, testBob, "msg 1"},
		{testBob, testAlice, "msg 2"},
		{testAlice, testCharlie, "msg 3"},
	}
	for i, m := range seeds {
		_, err := env.q.SendMessage(t.Context(), db.SendMessageParams{
			ID: fmt.Sprintf("f%d", i), Sender: m.s, Recipient: m.r, Content: m.c,
		})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	output, err := env.run("feed")
	if err != nil {
		t.Fatalf("feed failed: %v", err)
	}
	for _, want := range []string{"msg 1", "msg 2", "msg 3"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q: %s", want, output)
		}
	}
}

func TestFeedFilterFrom(t *testing.T) {
	env := newTestEnv(t)
	_, _ = env.q.SendMessage(t.Context(), db.SendMessageParams{
		ID: "f1", Sender: testAlice, Recipient: testBob, Content: "from alice",
	})
	_, _ = env.q.SendMessage(t.Context(), db.SendMessageParams{
		ID: "f2", Sender: testBob, Recipient: testAlice, Content: "from bob",
	})
	output, err := env.run("feed", "--from", testAlice)
	if err != nil {
		t.Fatalf("feed --from failed: %v", err)
	}
	if !strings.Contains(output, "from alice") {
		t.Fatalf("expected alice msg: %s", output)
	}
	if strings.Contains(output, "from bob") {
		t.Fatalf("unexpected bob msg: %s", output)
	}
}

func TestReadCommand(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.q.SendMessage(t.Context(), db.SendMessageParams{
		ID: "r1", Sender: testAlice, Recipient: testBob, Content: "read test",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	output, err := env.run("read", "r1")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(output, "marked as read") {
		t.Fatalf("expected marked as read: %s", output)
	}
	msg, err := env.q.GetMessageByID(t.Context(), "r1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if msg.Read != 1 {
		t.Fatalf("expected read=1, got %d", msg.Read)
	}
}

func TestReadNonexistent(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run("read", "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEntitiesCommand(t *testing.T) {
	env := newTestEnv(t)
	seeds := []struct{ s, r string }{
		{testAlice, testBob},
		{testBob, testCharlie},
		{testAlice, testCharlie},
	}
	for i, m := range seeds {
		_, err := env.q.SendMessage(t.Context(), db.SendMessageParams{
			ID: fmt.Sprintf("e%d", i), Sender: m.s, Recipient: m.r, Content: "x",
		})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	output, err := env.run("entities")
	if err != nil {
		t.Fatalf("entities failed: %v", err)
	}
	for _, want := range []string{testAlice, testBob, testCharlie} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q: %s", want, output)
		}
	}
}

func TestEntitiesJSON(t *testing.T) {
	env := newTestEnv(t)
	_, _ = env.q.SendMessage(t.Context(), db.SendMessageParams{
		ID: "ej", Sender: testAlice, Recipient: testBob, Content: "x",
	})
	output, err := env.run("entities", "--json")
	if err != nil {
		t.Fatalf("entities --json failed: %v", err)
	}
	var entities []string
	if err := json.Unmarshal([]byte(output), &entities); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entities) < 2 {
		t.Fatalf("expected at least 2 entities, got %d", len(entities))
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
