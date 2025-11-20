package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"tingly-box/internal/config"
	"tingly-box/internal/server"
)

// ServerManager manages the HTTP server lifecycle
type ServerManager struct {
	appConfig *config.AppConfig
	server    *server.Server
	pidFile   string
}

// NewServerManager creates a new server manager
func NewServerManager(appConfig *config.AppConfig) *ServerManager {
	return &ServerManager{
		appConfig: appConfig,
		pidFile:   filepath.Join(os.TempDir(), "tingly-server.pid"),
	}
}

// Start starts the server
func (sm *ServerManager) Start(port int) error {
	// Check if already running
	if sm.IsRunning() {
		return fmt.Errorf("server is already running")
	}

	// Set port if provided
	if port > 0 {
		if err := sm.appConfig.SetServerPort(port); err != nil {
			return fmt.Errorf("failed to set server port: %w", err)
		}
	}

	// Create server
	sm.server = server.NewServer(sm.appConfig)

	// Create PID file
	pid := os.Getpid()
	if err := os.WriteFile(sm.pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	// Start server in goroutine
	go func() {
		if err := sm.server.Start(sm.appConfig.GetServerPort()); err != nil {
			fmt.Printf("Server error: %v\n", err)
			sm.Cleanup()
		}
	}()

	// Setup signal handling for graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nReceived shutdown signal")
		sm.Stop()
	}()

	return nil
}

// Stop stops the server gracefully
func (sm *ServerManager) Stop() error {
	if sm.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sm.server.Stop(ctx); err != nil {
		fmt.Printf("Error stopping server: %v\n", err)
	}

	sm.Cleanup()
	return nil
}

// Cleanup removes PID file
func (sm *ServerManager) Cleanup() {
	if sm.pidFile != "" {
		os.Remove(sm.pidFile)
	}
}

// IsRunning checks if the server is currently running
func (sm *ServerManager) IsRunning() bool {
	if _, err := os.Stat(sm.pidFile); os.IsNotExist(err) {
		return false
	}

	// In a real implementation, we would check if the PID is actually running
	// For now, just check if the file exists
	return true
}