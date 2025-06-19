# MTG Treacherest

A real-time multiplayer game of deception and hidden roles, built with Go, Templ, and Datastar.

## Development

This project uses Nix for development environment management and gomod2nix for Go dependency management.

### Prerequisites

- Nix with flakes enabled

### Getting Started

1. Enter the development shell:
   ```bash
   nix develop
   ```

2. Start the development server with hot reload:
   ```bash
   dev
   ```

3. Visit http://localhost:8080

### Available Commands

- `dev` - Start development server with hot reload
- `run` - Run the server (builds templates first)
- `build` - Build the application
- `test` - Run all tests
- `test-all` - Run tests with coverage report
- `fmt` - Format Go and templ code
- `build-templ` - Generate Go code from templ templates
- `update-deps` - Update Go dependencies and regenerate gomod2nix.toml

### Project Structure

```
.
├── cmd/server/         # Application entry point
├── internal/
│   ├── game/          # Core game logic
│   ├── handlers/      # HTTP and SSE handlers
│   ├── store/         # In-memory game storage
│   └── views/         # Templ templates
├── static/            # Static assets
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
└── gomod2nix.toml     # Nix dependency management
```

### Architecture

The application uses:
- **Server-side rendering** with Templ templates
- **Real-time updates** via Server-Sent Events (SSE)
- **Full page morphing** with Datastar's idiomorph algorithm
- **In-memory storage** for game state

### Game Flow

1. Players create or join rooms with 5-character codes
2. 4-8 players required per game
3. Roles are randomly assigned when game starts
4. 5-second countdown before role reveal
5. Players work toward their role's win condition

### Dependencies

All Go dependencies are managed through gomod2nix. To add or update dependencies:

1. Use standard Go commands:
   ```bash
   go get package@version
   ```

2. Regenerate gomod2nix.toml:
   ```bash
   update-deps
   ```

This ensures reproducible builds in the Nix environment.