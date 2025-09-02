package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clichat",
	Short: "CLI LLM chat bot",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("clichat: use the chat command or run interactively")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
