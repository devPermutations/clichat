package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourname/clichat/internal/config"
	"github.com/yourname/clichat/internal/provider/litellm"
)

func init() {
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(modelCmd)
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models from LiteLLM",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		client := litellm.NewClient(cfg.LiteLLMBaseURL, cfg.LiteLLMAPIKey)
		mods, err := client.ListModels(context.Background())
		if err != nil {
			return err
		}
		for _, m := range mods {
			name := m.Name
			if name == "" {
				name = m.ID
			}
			fmt.Println(name)
		}
		return nil
	},
}

var modelCmd = &cobra.Command{
	Use:   "model <name>",
	Short: "Set default model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("model name required")
		}
		st, err := config.LoadState()
		if err != nil {
			return err
		}
		st.Model = name
		if err := config.SaveState(st); err != nil {
			return err
		}
		fmt.Println("default model set to:", name)
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		client := litellm.NewClient(cfg.LiteLLMBaseURL, cfg.LiteLLMAPIKey)
		mods, err := client.ListModels(context.Background())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var out []string
		for _, m := range mods {
			name := m.Name
			if name == "" {
				name = m.ID
			}
			if toComplete == "" || strings.Contains(strings.ToLower(name), strings.ToLower(toComplete)) {
				out = append(out, name)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	},
}
