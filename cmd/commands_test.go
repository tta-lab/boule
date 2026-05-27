package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"

	"github.com/tta-lab/boule/internal/store"
)

const (
	testAlice   = "alice"
	testBob     = "bob"
	testCharlie = "charlie"
)

type testEnv struct {
	db     *sql.DB
	dbPath string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	d, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })

	return &testEnv{db: d, dbPath: dbPath}
}

func (e *testEnv) seed(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := e.db.Exec(query, args...); err != nil {
		t.Fatalf("seed %q: %v", query, err)
	}
}

func (e *testEnv) newRoot() *cobra.Command {
	root := &cobra.Command{
		Use: "bo",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			database = e.db
			return nil
		},
	}
	root.AddCommand(sendCmd, inboxCmd, feedCmd, readCmd, entitiesCmd)
	return root
}

func (e *testEnv) run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return e.runWithStdin(t, "", args...)
}

func (e *testEnv) runWithStdin(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	root := e.newRoot()
	root.SetArgs(args)
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(buf)
	root.SetErr(buf)
	err := root.Execute()
	return buf.String(), err
}

func TestSendCommand(t *testing.T) {
	env := newTestEnv(t)
	output, err := env.runWithStdin(t, "hello world", "send", "--from", testAlice, testBob)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !strings.Contains(output, "sent alice -> bob") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestSendEmptyContent(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.runWithStdin(t, "", "send", "--from", testAlice, testBob)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestInboxCommand(t *testing.T) {
	env := newTestEnv(t)
	env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
		"t1", testAlice, testBob, "inbox test")
	output, err := env.run(t, "inbox", testBob)
	if err != nil {
		t.Fatalf("inbox failed: %v", err)
	}
	if !strings.Contains(output, "alice: inbox test") {
		t.Fatalf("expected message: %s", output)
	}
}

func TestInboxEmpty(t *testing.T) {
	env := newTestEnv(t)
	output, err := env.run(t, "inbox", "nobody")
	if err != nil {
		t.Fatalf("inbox failed: %v", err)
	}
	if !strings.Contains(output, "no unread messages") {
		t.Fatalf("expected empty: %s", output)
	}
}

func TestInboxJSON(t *testing.T) {
	env := newTestEnv(t)
	env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
		"tj", testAlice, testBob, "json test")
	output, err := env.run(t, "inbox", testBob, "--json")
	if err != nil {
		t.Fatalf("inbox --json failed: %v", err)
	}
	var msgs []struct {
		ID        string `json:"id"`
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
		Content   string `json:"content"`
	}
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
		env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("f%d", i), m.s, m.r, m.c)
	}
	output, err := env.run(t, "feed")
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
	env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
		"f1", testAlice, testBob, "from alice")
	env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
		"f2", testBob, testAlice, "from bob")
	output, err := env.run(t, "feed", "--from", testAlice)
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
	env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
		"r1", testAlice, testBob, "read test")
	output, err := env.run(t, "read", "r1")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(output, "marked as read") {
		t.Fatalf("expected marked as read: %s", output)
	}
	var read int
	if err := env.db.QueryRow("SELECT read FROM messages WHERE id = ?", "r1").Scan(&read); err != nil {
		t.Fatalf("get: %v", err)
	}
	if read != 1 {
		t.Fatalf("expected read=1, got %d", read)
	}
}

func TestReadNonexistent(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run(t, "read", "nonexistent")
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
		env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("e%d", i), m.s, m.r, "x")
	}
	output, err := env.run(t, "entities")
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
	env.seed(t, "INSERT INTO messages (id, sender, recipient, content) VALUES (?, ?, ?, ?)",
		"ej", testAlice, testBob, "x")
	output, err := env.run(t, "entities", "--json")
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
