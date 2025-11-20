package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Provider represents an AI model provider configuration
type Provider struct {
	Name    string `json:"name"`
	APIBase string `json:"api_base"`
	Token   string `json:"token"`
	Enabled bool   `json:"enabled"`
}

// Config represents the application configuration
type Config struct {
	Providers  map[string]*Provider `json:"providers"`
	ServerPort int                 `json:"server_port"`
	JWTSecret  string              `json:"jwt_secret"`
	mu         sync.RWMutex        `json:"-"`
}

// AppConfig holds the application configuration with encrypted storage
type AppConfig struct {
	configFile string
	config     *Config
	gcm        cipher.AEAD
	mu         sync.RWMutex
}

// NewAppConfig creates a new application configuration
func NewAppConfig() (*AppConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".tingly-box")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.enc")
	ac := &AppConfig{
		configFile: configFile,
		config: &Config{
			Providers:  make(map[string]*Provider),
			ServerPort: 8080,
			JWTSecret:  generateSecret(),
		},
	}

	// Initialize encryption
	if err := ac.initEncryption(); err != nil {
		return nil, fmt.Errorf("failed to initialize encryption: %w", err)
	}

	// Load existing configuration if exists
	if _, err := os.Stat(configFile); err == nil {
		if err := ac.Load(); err != nil {
			return nil, fmt.Errorf("failed to load existing config: %w", err)
		}
	}

	return ac, nil
}

// initEncryption initializes the encryption cipher
func (ac *AppConfig) initEncryption() error {
	// Use machine-specific key for encryption
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "tingly-box"
	}

	key := sha256.Sum256([]byte(hostname + "tingly-box-encryption-key"))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	ac.gcm = gcm
	return nil
}

// AddProvider adds a new AI provider configuration
func (ac *AppConfig) AddProvider(name, apiBase, token string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.config.mu.Lock()
	defer ac.config.mu.Unlock()

	if name == "" {
		return errors.New("provider name cannot be empty")
	}
	if apiBase == "" {
		return errors.New("API base URL cannot be empty")
	}
	if token == "" {
		return errors.New("API token cannot be empty")
	}

	ac.config.Providers[name] = &Provider{
		Name:    name,
		APIBase: apiBase,
		Token:   token,
		Enabled: true,
	}

	return ac.Save()
}

// GetProvider returns a provider by name
func (ac *AppConfig) GetProvider(name string) (*Provider, error) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	ac.config.mu.RLock()
	defer ac.config.mu.RUnlock()

	provider, exists := ac.config.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return provider, nil
}

// ListProviders returns all providers
func (ac *AppConfig) ListProviders() []*Provider {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	ac.config.mu.RLock()
	defer ac.config.mu.RUnlock()

	providers := make([]*Provider, 0, len(ac.config.Providers))
	for _, provider := range ac.config.Providers {
		providers = append(providers, provider)
	}

	return providers
}

// DeleteProvider removes a provider by name
func (ac *AppConfig) DeleteProvider(name string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.config.mu.Lock()
	defer ac.config.mu.Unlock()

	if _, exists := ac.config.Providers[name]; !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	delete(ac.config.Providers, name)
	return ac.Save()
}

// Save encrypts and saves the configuration to file
func (ac *AppConfig) Save() error {
	data, err := json.Marshal(ac.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	nonce := make([]byte, ac.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := ac.gcm.Seal(nonce, nonce, data, nil)

	// Encode to base64 for storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	if err := os.WriteFile(ac.configFile, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load decrypts and loads the configuration from file
func (ac *AppConfig) Load() error {
	data, err := os.ReadFile(ac.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	nonceSize := ac.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := ac.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(plaintext, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	ac.mu.Lock()
	ac.config = &config
	ac.mu.Unlock()

	return nil
}

// GetServerPort returns the configured server port
func (ac *AppConfig) GetServerPort() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return ac.config.ServerPort
}

// GetJWTSecret returns the JWT secret for token generation
func (ac *AppConfig) GetJWTSecret() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	return ac.config.JWTSecret
}

// SetServerPort updates the server port
func (ac *AppConfig) SetServerPort(port int) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.config.mu.Lock()
	defer ac.config.mu.Unlock()

	ac.config.ServerPort = port
	return ac.Save()
}

// generateSecret generates a random secret for JWT
func generateSecret() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}