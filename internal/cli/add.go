package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"tingly-box/internal/config"
)

// AddCommand represents the add provider command
func AddCommand(appConfig *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new AI provider configuration",
		Long: `Add a new AI provider with name, API base URL, and token through interactive prompts.
This command will guide you through entering the required information step by step.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractiveAdd(appConfig)
		},
	}

	return cmd
}

// runInteractiveAdd handles the interactive provider addition process
func runInteractiveAdd(appConfig *config.AppConfig) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Let's add a new AI provider configuration.")
	fmt.Println()

	// Get provider name
	name, err := promptForInput(reader, "Enter provider name (e.g., openai, anthropic): ", true)
	if err != nil {
		return err
	}

	// Check if provider already exists
	if existingProvider, err := appConfig.GetProvider(name); err == nil && existingProvider != nil {
		fmt.Printf("Provider '%s' already exists. Please use a different name or update the existing provider.\n", name)
		return fmt.Errorf("provider already exists")
	}

	// Get API base URL
	apiBase, err := promptForInput(reader, "Enter API base URL (e.g., https://api.openai.com/v1): ", true)
	if err != nil {
		return err
	}

	// Get API token
	token, err := promptForInput(reader, "Enter API token: ", true)
	if err != nil {
		return err
	}

	// Display summary and get confirmation
	fmt.Println("\n--- Configuration Summary ---")
	fmt.Printf("Provider Name: %s\n", name)
	fmt.Printf("API Base URL: %s\n", apiBase)
	fmt.Printf("Token: %s\n", maskToken(token))
	fmt.Println("---------------------------")

	confirmed, err := promptForConfirmation(reader, "Do you want to save this configuration? (Y/n): ")
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// Add the provider
	if err := appConfig.AddProvider(name, apiBase, token); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}

	fmt.Printf("Successfully added provider '%s'\n", name)
	return nil
}

// promptForInput prompts the user for input and returns the trimmed response
func promptForInput(reader *bufio.Reader, prompt string, required bool) (string, error) {
	for {
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		if required && input == "" {
			fmt.Println("This field is required. Please enter a value.")
			continue
		}

		return input, nil
	}
}

// promptForConfirmation prompts the user for a yes/no confirmation
func promptForConfirmation(reader *bufio.Reader, prompt string) (bool, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	// Default to Yes if user just presses Enter
	return input == "" || input == "y" || input == "yes", nil
}

// maskToken masks the API token for display purposes
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}