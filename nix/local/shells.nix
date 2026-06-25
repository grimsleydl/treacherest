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

      # Task runner for container builds
      just

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

      echo "Treacherest compatibility Nix shell"
      echo ""
      echo "Primary shell: devenv shell"
      echo "Primary commands: just --list"
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
