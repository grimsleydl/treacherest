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
  lib = pkgs.lib;

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

      # Playwright with priority to resolve conflicts
      nodejs_22 # Match the version playwright expects
      (lib.hiPrio playwright-driver) # Give playwright high priority
      (lib.hiPrio playwright-test)   # Give playwright-test high priority

      # Node/npm already included via nodejs_22 above
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
      {
        name = "test-playwright";
        help = "Run Playwright browser tests";
        command = ''
          cd $PRJ_ROOT/nix/app
          playwright test
        '';
      }
      {
        name = "test-playwright-ui";
        help = "Run Playwright tests with UI mode";
        command = ''
          cd $PRJ_ROOT/nix/app
          playwright test --ui
        '';
      }
      {
        name = "test-playwright-debug";
        help = "Run Playwright tests in debug mode";
        command = ''
          cd $PRJ_ROOT/nix/app
          playwright test --debug
        '';
      }
      {
        name = "download-cards";
        help = "Download all MTG Treachery card images";
        command = ''
          cd $PRJ_ROOT/nix/app/scripts/download
          go run download_cards.go
        '';
      }
      {
        name = "download-cards-sample";
        help = "Download sample MTG Treachery card images for testing";
        command = ''
          cd $PRJ_ROOT/nix/app/scripts/sample
          go run download_sample.go
        '';
      }
      {
        name = "download-cards-info";
        help = "Show information about downloading card images";
        command = ''
          cd $PRJ_ROOT/nix/app/scripts/info
          go run download_cards_info.go
        '';
      }
      {
        name = "setup-css";
        help = "Install CSS dependencies via npm";
        command = ''
          cd $PRJ_ROOT/nix/app
          npm install
          echo "✅ CSS dependencies installed"
        '';
      }
      {
        name = "build-css";
        help = "Build CSS with Tailwind and DaisyUI";
        command = ''
          cd $PRJ_ROOT/nix/app
          npm run build:css
        '';
      }
      {
        name = "build-css-prod";
        help = "Build minified CSS for production";
        command = ''
          cd $PRJ_ROOT/nix/app
          npm run build:css:prod
        '';
      }
      {
        name = "dev-css";
        help = "Start development server with CSS watching";
        command = ''
          cd $PRJ_ROOT/nix/app
          
          # Kill any existing processes
          pkill -f "postcss.*watch" || true
          pkill -f "templ generate.*watch" || true
          
          # Start CSS watcher
          npm run watch:css &
          CSS_PID=$!
          
          # Start templ watcher
          templ generate --watch --proxy="http://localhost:8080" --open-browser=false &
          TEMPL_PID=$!
          
          # Cleanup on exit
          trap "kill $CSS_PID $TEMPL_PID 2>/dev/null || true" EXIT INT TERM
          
          # Start air
          air
        '';
      }
      {
        name = "backup-templates";
        help = "Backup template files before CSS migration";
        command = ''
          cd $PRJ_ROOT/nix/app
          timestamp=$(date +%Y%m%d-%H%M%S)
          cp -r internal/views internal/views.backup-$timestamp
          echo "✅ Templates backed up to internal/views.backup-$timestamp"
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
      # Playwright environment variables for Nix
      {
        name = "PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS";
        value = "true";
      }
      {
        name = "PLAYWRIGHT_BROWSERS_PATH";
        eval = "${pkgs.playwright-driver.browsers}";
      }
      {
        name = "PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD";
        value = "1";
      }
      {
        name = "PLAYWRIGHT_NODEJS_PATH";
        eval = "${pkgs.nodejs_22}/bin/node";
      }
    ];

    devshell.startup.treacherest-startup.text = ''
      mkdir -p "$GOPATH" "$GOCACHE"

      echo "🃏 Treacherest Go Development Environment"
      echo ""
      echo "Available commands:"
      echo "  dev              - Start development server with hot reload"
      echo "  run              - Run the server (builds templates first)"
      echo "  build            - Build the application"
      echo "  test-all         - Run all Go tests"
      echo "  test-coverage    - Run Go tests with coverage report"
      echo "  test-playwright  - Run Playwright browser tests"
      echo "  test-playwright-ui - Run Playwright tests with UI mode"
      echo "  test-playwright-debug - Run Playwright tests in debug mode"
      echo "  fmt              - Format Go and templ code"
      echo "  build-templ      - Generate Go code from templ templates"
      echo "  update-deps      - Update Go dependencies"
      echo "  download-cards   - Download all MTG Treachery card images"
      echo "  download-cards-sample - Download sample card images for testing"
      echo ""
      echo "CSS/DaisyUI Migration commands:"
      echo "  setup-css        - Install CSS dependencies"
      echo "  build-css        - Build CSS with Tailwind and DaisyUI"
      echo "  build-css-prod   - Build minified CSS for production"
      echo "  dev-css          - Start dev server with CSS watching"
      echo "  backup-templates - Backup templates before migration"
      echo ""
      echo "Playwright environment configured:"
      echo "  PLAYWRIGHT_BROWSERS_PATH=$PLAYWRIGHT_BROWSERS_PATH"
      echo "  PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=$PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD"
      echo "  Node.js: $(node --version)"
      echo "  Priority conflict resolved with lib.hiPrio"
      echo ""
    '';
  };
}
