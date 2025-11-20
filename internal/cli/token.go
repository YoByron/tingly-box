package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"tingly-box/internal/auth"
	"tingly-box/internal/config"
)

// TokenCommand represents the generate token command
func TokenCommand(appConfig *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Generate a JWT token for API authentication",
		Long: `Generate a JWT token that can be used to authenticate requests
to the Tingly Box API endpoint. Include this token in the Authorization header
as 'Bearer <token>'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jwtManager := auth.NewJWTManager(appConfig.GetJWTSecret())

			token, err := jwtManager.GenerateToken("client")
			if err != nil {
				return fmt.Errorf("failed to generate token: %w", err)
			}

			fmt.Println("Generated JWT Token:")
			fmt.Println(token)
			fmt.Println()
			fmt.Println("Usage in API requests:")
			fmt.Println("Authorization: Bearer", token)

			return nil
		},
	}

	return cmd
}