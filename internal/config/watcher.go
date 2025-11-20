package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher monitors configuration changes and triggers reloads
type ConfigWatcher struct {
	config      *AppConfig
	watcher     *fsnotify.Watcher
	callbacks   []func(*Config)
	stopCh      chan struct{}
	mu          sync.RWMutex
	running     bool
	lastModTime time.Time
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(config *AppConfig) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	cw := &ConfigWatcher{
		config:  config,
		watcher: watcher,
		stopCh:  make(chan struct{}),
	}

	return cw, nil
}

// AddCallback adds a callback function to be called when configuration changes
func (cw *ConfigWatcher) AddCallback(callback func(*Config)) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.callbacks = append(cw.callbacks, callback)
}

// Start starts watching for configuration changes
func (cw *ConfigWatcher) Start() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.running {
		return fmt.Errorf("watcher is already running")
	}

	// Get initial modification time
	if stat, err := os.Stat(cw.config.configFile); err == nil {
		cw.lastModTime = stat.ModTime()
	}

	// Add config file to watcher
	if err := cw.watcher.Add(cw.config.configFile); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	// Also watch the directory for file creation/rename
	configDir := filepath.Dir(cw.config.configFile)
	if err := cw.watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	cw.running = true

	// Start watching in goroutine
	go cw.watchLoop()

	return nil
}

// Stop stops the configuration watcher
func (cw *ConfigWatcher) Stop() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.running {
		return nil
	}

	cw.running = false
	close(cw.stopCh)

	return cw.watcher.Close()
}

// watchLoop monitors file system events
func (cw *ConfigWatcher) watchLoop() {
	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // Stop the initial timer

	for {
		select {
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Filter for events related to our config file
			if !cw.isConfigEvent(event) {
				continue
			}

			// Debounce rapid file changes
			debounceTimer.Stop()
			debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
				cw.handleConfigChange(event)
			})

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Config watcher error: %v", err)

		case <-cw.stopCh:
			return
		}
	}
}

// isConfigEvent checks if an event is related to our config file
func (cw *ConfigWatcher) isConfigEvent(event fsnotify.Event) bool {
	// Direct config file events
	if event.Name == cw.config.configFile {
		return event.Op&(fsnotify.Write|fsnotify.Create) != 0
	}

	// Check if it's a create/rename event in the config directory
	configDir := filepath.Dir(cw.config.configFile)
	if filepath.Dir(event.Name) == configDir {
		return event.Op&(fsnotify.Create|fsnotify.Rename) != 0
	}

	return false
}

// handleConfigChange processes configuration changes
func (cw *ConfigWatcher) handleConfigChange(event fsnotify.Event) {
	// Check if file actually changed (avoid reloads on metadata changes)
	if stat, err := os.Stat(cw.config.configFile); err == nil {
		if !stat.ModTime().After(cw.lastModTime) {
			return
		}
		cw.lastModTime = stat.ModTime()
	} else {
		// File doesn't exist, skip reload
		return
	}

	// Reload configuration
	if err := cw.config.Load(); err != nil {
		log.Printf("Failed to reload configuration: %v", err)
		return
	}

	// Notify callbacks
	cw.mu.RLock()
	callbacks := make([]func(*Config), len(cw.callbacks))
	copy(callbacks, cw.callbacks)
	cw.mu.RUnlock()

	for _, callback := range callbacks {
		callback(cw.config.config)
	}

	log.Println("Configuration reloaded successfully")
}

// TriggerReload manually triggers a configuration reload
func (cw *ConfigWatcher) TriggerReload() error {
	return cw.config.Load()
}