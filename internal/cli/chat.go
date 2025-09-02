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
			fmt.Print("assistant> ")
			if err := svc.HandleUserInput(context.Background(), "default", line); err != nil {
				fmt.Println("\nerror:", err)
			}
			fmt.Println()
		}
		return scanner.Err()
	},
}
