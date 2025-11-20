package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"tingly-box/internal/config"
	"tingly-box/pkg/utils"
)

// PidFile stores the server process ID
const PidFile = "tingly-server.pid"

// StartCommand represents the start server command
func StartCommand(appConfig *config.AppConfig) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Tingly Box server",
		Long: `Start the Tingly Box HTTP server that provides the unified API endpoint.
The server will handle request routing to configured AI providers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverManager := utils.NewServerManager(appConfig)

			// Check if server is already running
			if serverManager.IsRunning() {
				fmt.Println("Server is already running")
				return nil
			}

			// Start server
			if err := serverManager.Start(port); err != nil {
				return fmt.Errorf("failed to start server: %w", err)
			}

			fmt.Printf("Server started successfully on port %d\n", appConfig.GetServerPort())
			fmt.Printf("API endpoint: http://localhost:%d/v1/chat/completions\n", appConfig.GetServerPort())
			fmt.Println("Use 'tingly status' to check server status")
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Server port (default: 8080)")
	return cmd
}

// StopCommand represents the stop server command
func StopCommand(appConfig *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the Tingly Box server",
		Long: `Stop the running Tingly Box HTTP server gracefully.
All ongoing requests will be completed before shutdown.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverManager := utils.NewServerManager(appConfig)

			if !serverManager.IsRunning() {
				fmt.Println("Server is not running")
				return nil
			}

			fmt.Println("Stopping server...")
			if err := serverManager.Stop(); err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}

			fmt.Println("Server stopped successfully")
			return nil
		},
	}

	return cmd
}

// StatusCommand represents the status command
func StatusCommand(appConfig *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check server status and configuration",
		Long: `Display the current status of the Tingly Box server and
show configuration information including number of providers and server port.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := appConfig.ListProviders()
			serverManager := utils.NewServerManager(appConfig)
			serverRunning := serverManager.IsRunning()

			fmt.Println("=== Tingly Box Status ===")
			fmt.Printf("Server Status: ")
			if serverRunning {
				fmt.Printf("Running\n")
				fmt.Printf("Port: %d\n", appConfig.GetServerPort())
				fmt.Printf("Endpoint: http://localhost:%d/v1/chat/completions\n", appConfig.GetServerPort())
			} else {
				fmt.Printf("Stopped\n")
			}

			fmt.Printf("\nConfigured Providers: %d\n", len(providers))
			if len(providers) > 0 {
				fmt.Println("Providers:")
				for _, provider := range providers {
					status := "Disabled"
					if provider.Enabled {
						status = "Enabled"
					}
					fmt.Printf("  - %s (%s): %s\n", provider.Name, provider.APIBase, status)
				}
			}

			return nil
		},
	}

	return cmd
}