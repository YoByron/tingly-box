# Tingly Box

A CLI tool and server for managing multiple AI model providers with a unified OpenAI-compatible endpoint.

## Features

- **CLI Management**: Add, list, and delete AI provider configurations
- **Unified Endpoint**: Single OpenAI-compatible API endpoint for all providers
- **Dynamic Configuration**: Hot-reload configuration changes without server restart
- **JWT Authentication**: Secure token-based API access
- **Encrypted Storage**: Secure storage of sensitive API tokens

## Quick Start

### 1. Build the application

```bash
go build ./cmd/tingly
```

### 2. Add an AI provider

```bash
./tingly add openai https://api.openai.com/v1 sk-your-openai-token
./tingly add anthropic https://api.anthropic.com sk-your-anthropic-token
```

### 3. List configured providers

```bash
./tingly list
```

### 4. Generate example and test token

```bash
./tingly example
```

The `example` command generates a JWT token and shows a ready-to-use curl command for testing.

### 5. Start the server

```bash
./tingly start --port 8080
```

### 6. Use the unified API endpoint

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello, world!"}]
  }'
```

## CLI Commands

### Provider Management

- `tingly add <name> <api-base> <token>` - Add a new AI provider
- `tingly list` - List all configured providers
- `tingly delete <name>` - Delete a provider configuration
- `tingly token` - Generate a JWT authentication token
- `tingly example` - Generate example token and curl command for testing

### Server Management

- `tingly start [--port <port>]` - Start the server (default port: 8080)
- `tingly stop` - Stop the running server
- `tingly status` - Check server status and configuration

## Configuration

Configuration is stored securely in `~/.tingly-box/config.enc` with encryption based on your hostname.

## API Endpoints

- `GET /health` - Health check
- `POST /token` - Generate JWT token
- `POST /v1/chat/completions` - OpenAI-compatible chat completions (requires authentication)

## Supported Providers

The system is provider-agnostic and works with any OpenAI-compatible API. Provider selection can be:
- Automatic based on model name patterns
- Explicit by adding `provider` parameter to requests

## Development

### Project Structure

```
├── cmd/tingly/          # CLI entry point
├── internal/
│   ├── auth/           # JWT authentication
│   ├── cli/            # CLI command implementations
│   ├── config/         # Configuration management and hot-reload
│   └── server/         # HTTP server and API handlers
├── pkg/utils/          # Server management utilities
└── go.mod              # Go module definition
```

### Build Requirements

- Go 1.25.3+
- See go.mod for full dependency list

### Running Tests

```bash
go test ./...
```

## License

MIT License