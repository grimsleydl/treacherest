{ config, inputs, lib, pkgs, ... }:

let
  system = pkgs.stdenv.hostPlatform.system;
  gomod2nixPackage = inputs.gomod2nix.packages.${system}.default;
  templPackage = inputs.templ.packages.${system}.templ;
  devPortBase = 8888;
  devPortMax = 8899;
  devPort = config.processes.treacherest.ports.http.value;
  devPortString = toString devPort;
  appDir = "${config.git.root}/nix/app";
in
{
  name = "treacherest";

  cachix.enable = false;

  packages = with pkgs; [
    air
    bashInteractive
    chromedriver
    coreutils
    curl
    delve
    gcc
    git
    gnugrep
    gnused
    go
    gomod2nixPackage
    golangci-lint
    gopls
    gotools
    jq
    just
    nix
    nodejs_22
    podman
    skopeo
    templPackage
    (lib.hiPrio playwright-driver)
    (lib.hiPrio playwright-test)
  ];

  env = {
    CGO_ENABLED = "0";
    PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS = "true";
    PLAYWRIGHT_BROWSERS_PATH = "${pkgs.playwright-driver.browsers}";
    PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD = "1";
    PLAYWRIGHT_NODEJS_PATH = "${pkgs.nodejs_22}/bin/node";
    TREACHEREST_DEV_PORT_BASE = toString devPortBase;
    TREACHEREST_DEV_PORT_MAX = toString devPortMax;
  };

  enterShell = ''
    export GOPATH="$DEVENV_ROOT/.go"
    export GOCACHE="$DEVENV_ROOT/.go/cache"
    mkdir -p "$GOPATH" "$GOCACHE"

    echo "Treacherest devenv shell"
    echo "Primary commands: just dev, just test, just check-known-green, just check, just build, just image, just image-run"
  '';

  process.manager.implementation = "native";

  processes.css = {
    cwd = appDir;
    exec = "npm run watch:css";
    restart.on = "on_failure";
  };

  processes.templ = {
    cwd = appDir;
    exec = ''
      exec templ generate --watch --proxy="http://localhost:${devPortString}" --open-browser=false
    '';
    restart.on = "on_failure";
  };

  processes.treacherest = {
    cwd = appDir;
    ports.http.allocate = devPortBase;
    env = {
      HOST = "localhost";
      PORT = devPortString;
      CONFIG_PATH = "../../configs/server-development.yaml";
      SHUTDOWN_TIMEOUT = "250ms";
      CGO_ENABLED = "0";
    };
    exec = ''
      set -eu
      port="${devPortString}"
      if [ "$port" -gt "${toString devPortMax}" ]; then
        echo "Treacherest dev port $port is outside the project range ${toString devPortBase}-${toString devPortMax}." >&2
        echo "Stop another local service or set a different project port range before retrying." >&2
        exit 1
      fi

      echo "Treacherest dev server: http://localhost:$port"
      exec air
    '';
    ready = {
      http.get = {
        host = "127.0.0.1";
        port = devPort;
        path = "/health/ready";
      };
      initial_delay = 1;
      period = 2;
      probe_timeout = 1;
      failure_threshold = 30;
      timeout = 90;
    };
    restart.on = "on_failure";
  };
}
