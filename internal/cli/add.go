package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"tingly-box/internal/config"
)

// AddCommand represents the add provider command
func AddCommand(appConfig *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new AI provider configuration",
		Long: `Add a new AI provider with name, API base URL, and token.
Example: tingly add openai https://api.openai.com/v1 sk-...`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			apiBase := strings.TrimSpace(args[1])
			token := strings.TrimSpace(args[2])

			if err := appConfig.AddProvider(name, apiBase, token); err != nil {
				return fmt.Errorf("failed to add provider: %w", err)
			}

			fmt.Printf("Successfully added provider '%s'\n", name)
			return nil
		},
	}

	return cmd
}