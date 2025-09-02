package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourname/clichat/internal/chat"
	"github.com/yourname/clichat/internal/config"
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
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("you> ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "/") {
				if handled, err := handleSlashCommand(context.Background(), prov, line); err != nil {
					fmt.Println("error:", err)
					continue
				} else if handled {
					continue
				}
			}
			fmt.Print("assistant> ")
			if err := svc.HandleUserInput(context.Background(), "default", line); err != nil {
				fmt.Println("\nerror:", err)
			}
			fmt.Println()
		}
		return scanner.Err()
	},
}

func handleSlashCommand(ctx context.Context, prov *litellm.Client, line string) (bool, error) {
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
	default:
		// Unknown slash command; not handled.
		return false, nil
	}
}
