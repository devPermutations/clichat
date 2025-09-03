package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yourname/clichat/internal/config"
	ctxutil "github.com/yourname/clichat/internal/context"
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
	// Always reset color on exit so user prompt returns to default/white
	defer fmt.Print("\x1b[0m")

	if conversationID == "" {
		return errors.New("conversation id required")
	}
	if _, err := s.store.CreateOrGetConversation(conversationID, conversationID); err != nil {
		return err
	}
	// Persist only the current user message now
	if _, err := s.store.AppendMessage(conversationID, "user", text); err != nil {
		return err
	}
	// Load up to 200 recent messages for context
	messages, err := s.store.ListMessages(conversationID, 200)
	if err != nil {
		return err
	}
	var reqMsgs []litellm.ChatMessage
	if s.cfg.SystemPrompt != "" {
		reqMsgs = append(reqMsgs, litellm.ChatMessage{Role: "system", Content: s.cfg.SystemPrompt})
	}
	// Prefer relevant tail of history. If we have any assistant replies, include from the last assistant onward.
	// If we only have user messages (e.g., due to prior bug or interruptions), include only the last 1-2 user messages
	// to avoid the model re-answering the entire backlog repeatedly.
	lastAssistant := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" {
			lastAssistant = i
			break
		}
	}
	if lastAssistant >= 0 {
		for _, m := range messages[lastAssistant:] {
			reqMsgs = append(reqMsgs, litellm.ChatMessage{Role: m.Role, Content: m.Content})
		}
	} else {
		start := len(messages) - 2
		if start < 0 {
			start = 0
		}
		for _, m := range messages[start:] {
			if m.Role == "user" {
				reqMsgs = append(reqMsgs, litellm.ChatMessage{Role: m.Role, Content: m.Content})
			}
		}
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
	if s.cfg.DebugPrompts {
		fmt.Println("\n[debug] prompt context:")
		for i, m := range req.Messages {
			fmt.Printf("  %02d %s: %.60s\n", i, m.Role, m.Content)
		}
	}

	promptTokens := estimatePromptTokens(reqMsgs)
	deltas, errs := s.prov.StreamChat(ctx, req)
	var assistant string
	saved := false
	saveAssistant := func() int {
		if saved || assistant == "" {
			return 0
		}
		_, _ = s.store.AppendMessage(conversationID, "assistant", assistant)
		tokens := ctxutil.EstimateTokens(assistant)
		_ = s.store.UpdateContextUsage(conversationID, promptTokens, tokens)
		saved = true
		return tokens
	}
	for {
		select {
		case d, ok := <-deltas:
			if !ok {
				answerTokens := saveAssistant()
				if assistant != "" {
					if s.cfg.ModelContextTokens > 0 {
						used := promptTokens + answerTokens
						fmt.Printf("  [context: %d/%d (%s)]\n", used, s.cfg.ModelContextTokens, ctxutil.PercentUsed(used, s.cfg.ModelContextTokens))
					} else {
						fmt.Println()
					}
				}
				return nil
			}
			assistant += d
			_ = s.r.WriteToken(d)
		case err := <-errs:
			if err != nil {
				saveAssistant()
				return err
			}
			// nil error: ignore and continue
		case <-ctx.Done():
			saveAssistant()
			return ctx.Err()
		}
	}
}

func estimatePromptTokens(msgs []litellm.ChatMessage) int {
	contents := make([]string, 0, len(msgs))
	for _, m := range msgs {
		contents = append(contents, m.Content)
	}
	return ctxutil.EstimateTokensForContents(contents)
}
