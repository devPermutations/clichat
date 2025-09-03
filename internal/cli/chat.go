package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"github.com/yourname/clichat/internal/chat"
	"github.com/yourname/clichat/internal/config"
	ctxutil "github.com/yourname/clichat/internal/context"
	"github.com/yourname/clichat/internal/memory/sqlite"
	"github.com/yourname/clichat/internal/provider/litellm"
	"github.com/yourname/clichat/internal/stream"
)

func init() {
	rootCmd.AddCommand(chatCmd)
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		store, err := sqlite.Open(cfg.DBPath)
		if err != nil {
			return err
		}
		defer store.Close()
		prov := litellm.NewClient(cfg.LiteLLMBaseURL, cfg.LiteLLMAPIKey)
		r := stream.NewRenderer()
		svc := chat.NewService(cfg, store, prov, r)

		fmt.Println("Enter messages (Ctrl+C to quit). Conversation: default")

		ln := liner.NewLiner()
		defer ln.Close()
		ln.SetCtrlCAborts(true)

		ln.SetCompleter(func(line string) (c []string) {
			trim := strings.TrimSpace(line)
			lower := strings.ToLower(trim)

			// Base command suggestions when user starts typing '/'
			allCmds := []string{"/models", "/model ", "/history", "/clear", "/contextwindow"}
			if cfg.AllowLocalShell {
				allCmds = append(allCmds, "/bash ")
			}
			if strings.HasPrefix(lower, "/") && !strings.HasPrefix(lower, "/model") {
				for _, cmd := range allCmds {
					if strings.HasPrefix(cmd, lower) || lower == "/" {
						c = append(c, cmd)
					}
				}
			}

			// Model completion: support '/model' (no space) and '/model <partial>'
			if strings.HasPrefix(lower, "/model") {
				if lower == "/model" {
					return []string{"/model "}
				}
				if strings.HasPrefix(lower, "/model ") {
					partial := strings.TrimSpace(strings.TrimPrefix(trim, "/model "))
					mods, err := prov.ListModels(context.Background())
					if err != nil {
						return []string{"/models"}
					}
					for _, m := range mods {
						name := m.Name
						if name == "" {
							name = m.ID
						}
						if partial == "" || strings.HasPrefix(strings.ToLower(name), strings.ToLower(partial)) {
							c = append(c, "/model "+name)
						}
					}
					return c
				}
			}

			return c
		})

		for {
			// Plain user prompt (avoid ANSI here to prevent liner errors on Windows)
			line, err := ln.Prompt("you> ")
			if err != nil {
				if err == liner.ErrPromptAborted {
					fmt.Println()
					break
				}
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			ln.AppendHistory(line)

			if strings.HasPrefix(line, "/") {
				if handled, err := handleSlashCommand(context.Background(), prov, cfg, line); err != nil {
					fmt.Println("error:", err)
					continue
				} else if handled {
					continue
				}
			}

			// Blue tag for the model name; streaming stays blue; service resets color at end
			fmt.Printf("\x1b[34m%s> ", currentModelPrompt(cfg))
			if err := svc.HandleUserInput(context.Background(), "default", line); err != nil {
				fmt.Println("\nerror:", err)
			}
			fmt.Println()
		}
		return nil
	},
}

func currentModelPrompt(cfg *config.Config) string {
	name := cfg.Model
	if st, err := config.LoadState(); err == nil && st.Model != "" {
		name = st.Model
	}
	if name == "" {
		name = "assistant"
	}
	return name
}

func handleSlashCommand(ctx context.Context, prov *litellm.Client, cfg *config.Config, line string) (bool, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false, nil
	}
	switch parts[0] {
	case "/models":
		mods, err := prov.ListModels(ctx)
		if err != nil {
			return true, err
		}
		for _, m := range mods {
			name := m.Name
			if name == "" {
				name = m.ID
			}
			fmt.Println(name)
		}
		return true, nil
	case "/model":
		if len(parts) < 2 {
			fmt.Println("usage: /model <name>")
			return true, nil
		}
		name := strings.TrimSpace(parts[1])
		if name == "" {
			fmt.Println("usage: /model <name>")
			return true, nil
		}
		st, err := config.LoadState()
		if err != nil {
			return true, err
		}
		st.Model = name
		if err := config.SaveState(st); err != nil {
			return true, err
		}
		fmt.Println("default model set to:", name)
		return true, nil
	case "/history":
		store, err := sqlite.Open(cfg.DBPath)
		if err != nil {
			return true, err
		}
		defer store.Close()
		msgs, err := store.ListMessages("default", 200)
		if err != nil {
			return true, err
		}
		for _, m := range msgs {
			role := m.Role
			switch role {
			case "user":
				role = "you"
			case "assistant":
				role = currentModelPrompt(cfg)
			}
			fmt.Printf("%s> %s\n", role, strings.TrimSpace(m.Content))
		}
		return true, nil
	case "/clear":
		store, err := sqlite.Open(cfg.DBPath)
		if err != nil {
			return true, err
		}
		defer store.Close()
		if err := store.ClearConversation("default"); err != nil {
			return true, err
		}
		fmt.Println("history cleared for conversation: default")
		return true, nil
	case "/contextwindow":
		store, err := sqlite.Open(cfg.DBPath)
		if err != nil {
			return true, err
		}
		defer store.Close()
		conv, err := store.CreateOrGetConversation("default", "default")
		if err != nil {
			return true, err
		}
		used := conv.ContextPromptTokens + conv.ContextAnswerTokens
		if used == 0 {
			msgs, err := store.ListMessages("default", 200)
			if err == nil {
				answerTokens := 0
				promptTokens := 0
				promptCount := 0
				answerCount := 0
				if len(msgs) > 0 {
					// last assistant index
					idx := -1
					for i := len(msgs) - 1; i >= 0; i-- {
						if msgs[i].Role == "assistant" {
							idx = i
							break
						}
					}
					for i, m := range msgs {
						if m.Role == "assistant" {
							answerCount++
							if i == idx {
								answerTokens += ctxutil.EstimateTokens(m.Content)
							}
							continue
						}
						// user/system as prompt
						promptCount++
						promptTokens += ctxutil.EstimateTokens(m.Content)
					}
				}
				_ = store.UpdateContextStats("default", promptTokens, answerTokens, promptCount, answerCount)
				conv, _ = store.CreateOrGetConversation("default", "default")
				used = conv.ContextPromptTokens + conv.ContextAnswerTokens
			}
		}
		if cfg.ModelContextTokens > 0 {
			fmt.Printf("context: prompts=%d, answers=%d, tokens %d/%d (%s)\n", conv.PromptMessageCount, conv.AnswerMessageCount, used, cfg.ModelContextTokens, ctxutil.PercentUsed(used, cfg.ModelContextTokens))
		} else {
			fmt.Printf("context: prompts=%d, answers=%d, tokens %d (N/A)\n", conv.PromptMessageCount, conv.AnswerMessageCount, used)
		}
		return true, nil
	case "/bash":
		if !cfg.AllowLocalShell {
			fmt.Println("local shell is disabled (set ALLOW_LOCAL_SHELL=true)")
			return true, nil
		}
		if len(parts) < 2 {
			fmt.Println("usage: /bash <command>")
			return true, nil
		}
		cmdStr := strings.TrimSpace(strings.TrimPrefix(line, "/bash "))
		if cmdStr == "" {
			fmt.Println("usage: /bash <command>")
			return true, nil
		}
		execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		var c *exec.Cmd
		if runtime.GOOS == "windows" {
			// Prefer WSL on Windows; provide a clearer message if unavailable
			if _, err := exec.LookPath("wsl"); err != nil {
				fmt.Println("bash error: WSL is not installed or not in PATH; install WSL or run commands outside Windows")
				return true, nil
			}
			c = exec.CommandContext(execCtx, "wsl", "bash", "-lc", cmdStr)
		} else {
			// Use bash when available; fall back to sh (e.g., Alpine)
			shell := "bash"
			if _, err := exec.LookPath("bash"); err != nil {
				shell = "sh"
			}
			c = exec.CommandContext(execCtx, shell, "-lc", cmdStr)
		}
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			if execCtx.Err() == context.DeadlineExceeded {
				fmt.Println("bash error: command timed out")
			} else {
				fmt.Println("bash error:", err)
			}
		}
		return true, nil
	default:
		return false, nil
	}
}
