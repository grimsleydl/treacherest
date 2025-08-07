/*
This file holds reproducible shells with commands in them.

They conveniently also generate config files in their startup hook.
*/
{
  inputs,
  cell,
}: let
  inherit (cell) config;
  inherit (inputs.std.std) cli;
  inherit (inputs.std.lib) dev cfg;
  pkgs = import inputs.nixpkgs {
    inherit (inputs.nixpkgs) system;
    overlays = [
      inputs.gomod2nix.overlays.default
    ];
  };

  # Create Go environment for the app
  goEnv = pkgs.mkGoEnv {
    pwd = inputs.self + /nix/app;
  };
in {
  default = dev.mkShell {
    name = "treacherest";

    nixago = [
      (dev.mkNixago cfg.conform)
      (dev.mkNixago cfg.treefmt config.treefmt)
      (dev.mkNixago cfg.editorconfig config.editorconfig)
      (dev.mkNixago cfg.lefthook config.lefthook)
    ];

    packages = with pkgs; [
      # Use the Go environment from gomod2nix
      # goEnv

      # Base go compiler needed for air
      go

      gopls
      gotools
      golangci-lint
      delve
      air # for hot reloading during development

      # gomod2nix tools
      gomod2nix

      # templ for template compilation
      templ
      
      # Lightweight alternative - just chromedriver for basic browser automation
      chromedriver
    ];

    commands = [
      {package = cli.std;}
      {package = pkgs.nvfetcher;}

      {
        name = "dev";
        help = "Start development server with hot reload";
        command = ''
          cd $PRJ_ROOT/nix/app
          templ generate --watch --proxy="http://localhost:8080" --open-browser=false &
          air
        '';
      }
      {
        name = "build-templ";
        help = "Generate Go code from templ templates";
        command = "cd $PRJ_ROOT/nix/app && templ generate";
      }
      {
        name = "update-deps";
        help = "Update Go dependencies and regenerate gomod2nix.toml";
        command = ''
          cd $PRJ_ROOT/nix/app
          go mod tidy
          gomod2nix generate
          echo "Dependencies updated!"
        '';
      }
      {
        name = "import-deps";
        help = "Import Go dependencies from cache to speed up builds";
        command = ''
          cd $PRJ_ROOT/nix/app
          gomod2nix import
          echo "Dependencies imported from Go cache!"
        '';
      }
      {
        name = "run";
        help = "Run the server (builds templates first)";
        command = ''
          cd $PRJ_ROOT/nix/app
          templ generate
          go run cmd/server/main.go
        '';
      }
      {
        name = "build";
        help = "Build the application";
        command = ''
          cd $PRJ_ROOT/nix/app
          templ generate
          go build -o bin/server cmd/server/main.go
          echo "Built server at nix/app/bin/server"
        '';
      }
      {
        name = "test-all";
        help = "Run all tests";
        command = ''
          cd $PRJ_ROOT/nix/app
          go test ./...
        '';
      }
      {
        name = "test-coverage";
        help = "Run all tests with coverage";
        command = ''
          cd $PRJ_ROOT/nix/app
          mkdir -p build/coverage
          go test -v -coverprofile=build/coverage/coverage.out ./...
          go tool cover -html=build/coverage/coverage.out -o build/coverage/coverage.html
          go tool cover -func=build/coverage/coverage.out
          echo "Coverage report generated at build/coverage/coverage.html"
        '';
      }
      {
        name = "fmt";
        help = "Format Go and templ code";
        command = ''
          cd $PRJ_ROOT/nix/app
          go fmt ./...
          templ fmt .
          echo "Code formatted!"
        '';
      }

    ];

    env = [
      {
        name = "GOPATH";
        eval = "$(pwd)/.go";
      }
      {
        name = "GOCACHE";
        eval = "$(pwd)/.go/cache";
      }
    ];

    devshell.startup.treacherest-startup.text = ''
      mkdir -p "$GOPATH" "$GOCACHE"

      echo "🃏 Treacherest Go Development Environment"
      echo ""
      echo "Available commands:"
      echo "  dev         - Start development server with hot reload"
      echo "  run         - Run the server (builds templates first)"
      echo "  build       - Build the application"
      echo "  test-all    - Run all tests"
      echo "  test-coverage - Run tests with coverage report"
      echo "  fmt         - Format Go and templ code"
      echo "  build-templ - Generate Go code from templ templates"
      echo "  update-deps - Update Go dependencies"
      echo ""
    '';
  };
}
