package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"tingly-box/internal/config"
)

// ChatCompletionRequest represents OpenAI chat completion request
type ChatCompletionRequest struct {
	Model    string                    `json:"model"`
	Messages []ChatCompletionMessage   `json:"messages"`
	Stream   bool                      `json:"stream,omitempty"`
	Provider string                    `json:"provider,omitempty"`
}

// ChatCompletionMessage represents a chat message
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents OpenAI chat completion response
type ChatCompletionResponse struct {
	ID      string                    `json:"id"`
	Object  string                    `json:"object"`
	Created int64                     `json:"created"`
	Model   string                    `json:"model"`
	Choices []ChatCompletionChoice    `json:"choices"`
	Usage   ChatCompletionUsage       `json:"usage"`
}

// ChatCompletionChoice represents a completion choice
type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      ChatCompletionMessage   `json:"message"`
	FinishReason string                 `json:"finish_reason"`
}

// ChatCompletionUsage represents token usage information
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "tingly-box",
	})
}

// generateToken handles token generation requests
func (s *Server) generateToken(c *gin.Context) {
	var req struct {
		ClientID string `json:"client_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid request body: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	token, err := s.jwtManager.GenerateToken(req.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Message: "Failed to generate token: " + err.Error(),
				Type:    "api_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"type":  "Bearer",
	})
}

// chatCompletions handles OpenAI-compatible chat completion requests
func (s *Server) chatCompletions(c *gin.Context) {
	var req ChatCompletionRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid request body: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Validate required fields
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "Model is required",
				Type:    "invalid_request_error",
			},
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: "At least one message is required",
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Determine provider based on model or explicit provider selection
	provider, err := s.determineProvider(req.Model, req.Provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Message: err.Error(),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Forward request to provider
	response, err := s.forwardRequest(provider, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Message: "Failed to forward request: " + err.Error(),
				Type:    "api_error",
			},
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// determineProvider selects the appropriate provider based on model or explicit provider name
func (s *Server) determineProvider(model, explicitProvider string) (*config.Provider, error) {
	providers := s.config.ListProviders()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	// If explicit provider is specified, use it
	if explicitProvider != "" {
		for _, provider := range providers {
			if provider.Name == explicitProvider && provider.Enabled {
				return provider, nil
			}
		}
		return nil, fmt.Errorf("provider '%s' not found or disabled", explicitProvider)
	}

	// Otherwise, try to determine provider based on model name
	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}

		// Simple model name matching - can be enhanced
		if strings.Contains(strings.ToLower(provider.APIBase), "openai") &&
		   (strings.HasPrefix(strings.ToLower(model), "gpt") || strings.Contains(strings.ToLower(model), "openai")) {
			return provider, nil
		}
		if strings.Contains(strings.ToLower(provider.APIBase), "anthropic") &&
		   strings.HasPrefix(strings.ToLower(model), "claude") {
			return provider, nil
		}
	}

	// If no specific match, return first enabled provider
	for _, provider := range providers {
		if provider.Enabled {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no enabled providers available")
}

// forwardRequest forwards the request to the selected provider
func (s *Server) forwardRequest(provider *config.Provider, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Convert request to provider format (simplified)
	providerReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	reqBody, err := json.Marshal(providerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request to provider
	endpoint := provider.APIBase + "/chat/completions"
	if !strings.HasSuffix(provider.APIBase, "/") {
		endpoint = provider.APIBase + "/v1/chat/completions"
	} else {
		endpoint = provider.APIBase + "v1/chat/completions"
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+provider.Token)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf("provider returned error %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("provider error: %s", errorResp.Error.Message)
	}

	// Parse successful response
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &chatResp, nil
}