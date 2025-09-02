package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yourname/clichat/internal/config"
	"github.com/yourname/clichat/internal/memory/sqlite"
	"github.com/yourname/clichat/internal/provider/litellm"
	"github.com/yourname/clichat/internal/stream"
)

type Service struct {
	cfg   *config.Config
	store *sqlite.Store
	prov  *litellm.Client
	r     *stream.Renderer
}

func NewService(cfg *config.Config, store *sqlite.Store, prov *litellm.Client, r *stream.Renderer) *Service {
	return &Service{cfg: cfg, store: store, prov: prov, r: r}
}

func (s *Service) HandleUserInput(ctx context.Context, conversationID string, text string) error {
	if conversationID == "" {
		return errors.New("conversation id required")
	}
	if _, err := s.store.CreateOrGetConversation(conversationID, conversationID); err != nil {
		return err
	}
	if _, err := s.store.AppendMessage(conversationID, "user", text); err != nil {
		return err
	}
	messages, err := s.store.ListMessages(conversationID, 200)
	if err != nil {
		return err
	}
	var reqMsgs []litellm.ChatMessage
	if s.cfg.SystemPrompt != "" {
		reqMsgs = append(reqMsgs, litellm.ChatMessage{Role: "system", Content: s.cfg.SystemPrompt})
	}
	for _, m := range messages {
		reqMsgs = append(reqMsgs, litellm.ChatMessage{Role: m.Role, Content: m.Content})
	}
	// resolve model: state overrides env if present
	model := s.cfg.Model
	if st, err := config.LoadState(); err == nil {
		if st.Model != "" {
			model = st.Model
		}
	}
	tools := []litellm.Tool{}
	if s.cfg.EnableProviderWebsearch {
		tools = append(tools, litellm.Tool{Type: "web_search"})
	}
	// Build request with conditional sampling params
	req := litellm.ChatRequest{
		Model:    model,
		Messages: reqMsgs,
		Stream:   true,
		Tools:    tools,
	}
	if !(s.cfg.DropSamplingParams || strings.HasPrefix(model, "gpt-5")) {
		req.Temperature = s.cfg.Temperature
		req.TopP = s.cfg.TopP
	}
	deltas, errs := s.prov.StreamChat(ctx, req)
	var assistant string
	for {
		select {
		case d, ok := <-deltas:
			if !ok {
				if assistant != "" {
					_, _ = s.store.AppendMessage(conversationID, "assistant", assistant)
					// context usage suffix (placeholder as provider usage not parsed yet)
					if s.cfg.ModelContextTokens > 0 {
						fmt.Printf("  [context: N/A/%d]\n", s.cfg.ModelContextTokens)
					} else {
						fmt.Println()
					}
				}
				return nil
			}
			assistant += d
			_ = s.r.WriteToken(d)
		case err := <-errs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
