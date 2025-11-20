package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// ModelConfig represents the model configuration structure
type ModelConfig struct {
	Models []ModelDefinition `yaml:"models"`
}

// ModelDefinition defines a model with its mapping and aliases
type ModelDefinition struct {
	Name        string   `yaml:"name"`        // Default name for the model
	Provider    string   `yaml:"provider"`    // Provider name (e.g., "openai", "alibaba")
	APIBase     string   `yaml:"api_base"`    // API base URL for this provider
	Model       string   `yaml:"model"`       // Actual model name for API calls
	Aliases     []string `yaml:"aliases"`     // Alternative names for this model
	Description string   `yaml:"description"` // Human-readable description
	Category    string   `yaml:"category"`    // Category (e.g., "chat", "completion", "embedding")
}

// ModelManager manages model configuration and matching
type ModelManager struct {
	config    ModelConfig
	modelMap  map[string]*ModelDefinition // name -> model definition
	aliasMap  map[string]*ModelDefinition // alias -> model definition
	configFile string
}

// NewModelManager creates a new model manager
func NewModelManager() (*ModelManager, error) {
	mm := &ModelManager{
		configFile: filepath.Join("config", "model.yaml"),
		modelMap:  make(map[string]*ModelDefinition),
		aliasMap:  make(map[string]*ModelDefinition),
	}

	if err := mm.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load model config: %w", err)
	}

	return mm, nil
}

// loadConfig loads the model configuration from YAML file
func (mm *ModelManager) loadConfig() error {
	// Create default config if file doesn't exist
	if _, err := os.Stat(mm.configFile); os.IsNotExist(err) {
		if err := mm.createDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	data, err := os.ReadFile(mm.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &mm.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Build maps for fast lookup
	mm.buildMaps()

	return nil
}

// createDefaultConfig creates a default model configuration file
func (mm *ModelManager) createDefaultConfig() error {
	// Create config directory
	configDir := filepath.Dir(mm.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Default models configuration
	defaultConfig := ModelConfig{
		Models: []ModelDefinition{
			{
				Name:        "gpt-3.5-turbo",
				Provider:    "openai",
				APIBase:     "https://api.openai.com/v1",
				Model:       "gpt-3.5-turbo",
				Aliases:     []string{"chatgpt", "gpt35"},
				Description: "OpenAI GPT-3.5 Turbo model",
				Category:    "chat",
			},
			{
				Name:        "gpt-4",
				Provider:    "openai",
				APIBase:     "https://api.openai.com/v1",
				Model:       "gpt-4",
				Aliases:     []string{"gpt4"},
				Description: "OpenAI GPT-4 model",
				Category:    "chat",
			},
			{
				Name:        "qwen-plus",
				Provider:    "alibaba",
				APIBase:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
				Model:       "qwen-plus",
				Aliases:     []string{"qwen", "tongyi"},
				Description: "Alibaba Qwen Plus model",
				Category:    "chat",
			},
			{
				Name:        "qwen-turbo",
				Provider:    "alibaba",
				APIBase:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
				Model:       "qwen-turbo",
				Aliases:     []string{"qwen-fast"},
				Description: "Alibaba Qwen Turbo model",
				Category:    "chat",
			},
			{
				Name:        "claude-3-sonnet",
				Provider:    "anthropic",
				APIBase:     "https://api.anthropic.com",
				Model:       "claude-3-sonnet-20240229",
				Aliases:     []string{"claude-sonnet", "claude3"},
				Description: "Anthropic Claude 3 Sonnet model",
				Category:    "chat",
			},
		},
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	return os.WriteFile(mm.configFile, data, 0644)
}

// buildMaps creates lookup maps for model names and aliases
func (mm *ModelManager) buildMaps() {
	mm.modelMap = make(map[string]*ModelDefinition)
	mm.aliasMap = make(map[string]*ModelDefinition)

	for i := range mm.config.Models {
		model := &mm.config.Models[i]

		// Add primary name
		mm.modelMap[strings.ToLower(model.Name)] = model

		// Add aliases
		for _, alias := range model.Aliases {
			mm.aliasMap[strings.ToLower(alias)] = model
		}
	}
}

// FindModel finds a model definition by name or alias
func (mm *ModelManager) FindModel(nameOrAlias string) (*ModelDefinition, error) {
	if nameOrAlias == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	// Try exact match first (case-insensitive)
	lowerName := strings.ToLower(nameOrAlias)

	// Check primary names
	if model, exists := mm.modelMap[lowerName]; exists {
		return model, nil
	}

	// Check aliases
	if model, exists := mm.aliasMap[lowerName]; exists {
		return model, nil
	}

	// Try partial match for fuzzy matching
	return mm.fuzzyFindModel(lowerName)
}

// fuzzyFindModel tries to find a model using partial matching
func (mm *ModelManager) fuzzyFindModel(name string) (*ModelDefinition, error) {
	var matches []ModelDefinition

	// Search in primary names
	for _, model := range mm.modelMap {
		if strings.Contains(strings.ToLower(model.Name), name) {
			matches = append(matches, *model)
		}
	}

	// Search in aliases
	for alias, model := range mm.aliasMap {
		if strings.Contains(alias, name) {
			matches = append(matches, *model)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("model '%s' not found", name)
	}

	if len(matches) > 1 {
		var modelNames []string
		for _, match := range matches {
			modelNames = append(modelNames, match.Name)
		}
		return nil, fmt.Errorf("ambiguous model name '%s', possible matches: %v", name, modelNames)
	}

	return &matches[0], nil
}

// GetAllModels returns all available models
func (mm *ModelManager) GetAllModels() []ModelDefinition {
	return mm.config.Models
}

// GetModelsByProvider returns models filtered by provider
func (mm *ModelManager) GetModelsByProvider(provider string) []ModelDefinition {
	var models []ModelDefinition
	provider = strings.ToLower(provider)

	for _, model := range mm.config.Models {
		if strings.ToLower(model.Provider) == provider {
			models = append(models, model)
		}
	}

	return models
}

// GetModelsByCategory returns models filtered by category
func (mm *ModelManager) GetModelsByCategory(category string) []ModelDefinition {
	var models []ModelDefinition
	category = strings.ToLower(category)

	for _, model := range mm.config.Models {
		if strings.ToLower(model.Category) == category {
			models = append(models, model)
		}
	}

	return models
}

// ReloadConfig reloads the configuration from file
func (mm *ModelManager) ReloadConfig() error {
	return mm.loadConfig()
}