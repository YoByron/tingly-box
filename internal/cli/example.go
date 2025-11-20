package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"tingly-box/internal/auth"
	"tingly-box/internal/config"
)

// ExampleCommand represents the example command
func ExampleCommand(appConfig *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example",
		Short: "Generate example token and curl command for testing",
		Long: `Generate a JWT token and show example curl command to test the API.
This is useful for quickly testing your setup and understanding the API format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := appConfig.ListProviders()

			if len(providers) == 0 {
				fmt.Println("⚠️  No providers configured!")
				fmt.Println("First add a provider:")
				fmt.Println("  ./tingly add openai https://api.openai.com/v1 sk-your-token")
				fmt.Println("  ./tingly add anthropic https://api.anthropic.com sk-your-token")
				return nil
			}

			// Generate token
			jwtManager := auth.NewJWTManager(appConfig.GetJWTSecret())
			token, err := jwtManager.GenerateToken("example-client")
			if err != nil {
				return fmt.Errorf("failed to generate token: %w", err)
			}

			fmt.Println("🔑 Generated JWT Token:")
			fmt.Println(token)
			fmt.Println()

			// Show available providers
			fmt.Println("📋 Available Providers:")
			for _, provider := range providers {
				if provider.Enabled {
					fmt.Printf("  • %s (%s)\n", provider.Name, provider.APIBase)
				}
			}
			fmt.Println()

			// Show server info
			port := appConfig.GetServerPort()
			endpoint := fmt.Sprintf("http://localhost:%d/v1/chat/completions", port)

			fmt.Println("🚀 Example API Test:")
			fmt.Println("Make sure your server is running:")
			fmt.Printf("  ./tingly start --port %d\n", port)
			fmt.Println()

			fmt.Println("Then test with curl:")
			fmt.Println("```bash")
			fmt.Printf("curl -X POST %s \\\n", endpoint)
			fmt.Println("  -H \"Content-Type: application/json\" \\")
			fmt.Printf("  -H \"Authorization: Bearer %s\" \\\n", token)
			fmt.Println("  -d '{")
			fmt.Println("    \"model\": \"gpt-3.5-turbo\",")
			fmt.Println("    \"messages\": [")
			fmt.Println("      {\"role\": \"user\", \"content\": \"Hello, how are you?\"}")
			fmt.Println("    ]")
			fmt.Println("  }'")
			fmt.Println("```")
			fmt.Println()

			// Show model suggestions based on providers
			fmt.Println("💡 Model Suggestions:")
			for _, provider := range providers {
				if !provider.Enabled {
					continue
				}

				switch provider.Name {
				case "openai":
					fmt.Println("  • gpt-3.5-turbo, gpt-4, gpt-4-turbo")
				case "anthropic":
					fmt.Println("  • claude-3-sonnet, claude-3-opus, claude-3-haiku")
				default:
					fmt.Printf("  • Check %s documentation for available models\n", provider.Name)
				}
			}
			fmt.Println()

			fmt.Println("📖 More examples:")
			fmt.Println("• With explicit provider selection:")
			fmt.Println("  {\"model\": \"gpt-3.5-turbo\", \"provider\": \"openai\", \"messages\": [...]}")
			fmt.Println("• Streaming response:")
			fmt.Println("  {\"model\": \"gpt-3.5-turbo\", \"stream\": true, \"messages\": [...]}")
			fmt.Println("• System message:")
			fmt.Println("  {\"model\": \"gpt-3.5-turbo\", \"messages\": [{\"role\": \"system\", \"content\": \"You are a helpful assistant.\"}, {\"role\": \"user\", \"content\": \"Hello!\"}]}")

			return nil
		},
	}

	return cmd
}