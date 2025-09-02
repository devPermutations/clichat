package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreCRUD(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.db")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("db file: %v", err)
	}

	conv, err := st.CreateOrGetConversation("conv1", "Title")
	if err != nil {
		t.Fatalf("CreateOrGetConversation: %v", err)
	}
	if conv.ID != "conv1" {
		t.Fatalf("got id %q", conv.ID)
	}

	if _, err := st.AppendMessage("conv1", "user", "hi"); err != nil {
		t.Fatalf("append user: %v", err)
	}
	if _, err := st.AppendMessage("conv1", "assistant", "hello"); err != nil {
		t.Fatalf("append assistant: %v", err)
	}

	msgs, err := st.ListMessages("conv1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "hi" {
		t.Fatalf("unexpected first message: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "hello" {
		t.Fatalf("unexpected second message: %+v", msgs[1])
	}
}
