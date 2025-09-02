package litellm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListModels(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(struct {
			Data []Model `json:"data"`
		}{Data: []Model{{ID: "m1", Name: "model-one"}}})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	c := NewClient(ts.URL, "")
	mods, err := c.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(mods) != 1 || mods[0].Name != "model-one" {
		t.Fatalf("unexpected mods: %+v", mods)
	}
}

func TestStreamChat(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		io := func(s string) {
			_, _ = w.Write([]byte(s))
			if flusher != nil {
				flusher.Flush()
			}
		}
		io("data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n")
		io("data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n")
		io("data: [DONE]\n")
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	c := NewClient(ts.URL, "")
	deltas, errs := c.StreamChat(context.Background(), ChatRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "hi"}}, Stream: true})
	var sb strings.Builder
	for d := range deltas {
		sb.WriteString(d)
	}
	if err := <-errs; err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if got := sb.String(); got != "Hello" {
		t.Fatalf("got %q", got)
	}
}
