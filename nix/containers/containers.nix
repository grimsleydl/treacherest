{
  inputs,
  cell,
}: let
  inherit (inputs) nixpkgs;
  l = nixpkgs.lib // builtins;
  
  # Import nix2container
  n2c = inputs.nix2container.packages.${inputs.nixpkgs.system};
in {
  # Main production container
  default = n2c.nix2container.buildImage {
    name = "docker.io/grimsleydl/treacherest";
    tag = "latest";
    
    # Copy binary, static files, and configs to container
    copyToRoot = nixpkgs.buildEnv {
      name = "image-root";
      paths = [ 
        inputs.cells.app.packages.default
        # Include default config files
        (nixpkgs.runCommand "configs" {} ''
          mkdir -p $out/app/config
          cp ${inputs.self}/configs/server.yaml $out/app/config/server.yaml
          cp ${inputs.self}/configs/production.yaml $out/app/config/production.yaml
        '')
      ];
      pathsToLink = [ "/bin" "/static" "/app" ];  # Copy bin, static, and app directories
    };
    
    # OCI-compliant configuration
    config = {
      # Run the binary directly from root where static files are
      entrypoint = [ "/bin/server" ];
      workingDir = "/";
      
      env = [
        "PATH=/bin"
        "SSL_CERT_FILE=${nixpkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
      ];
      
      exposedPorts = {
        "8080/tcp" = {};
      };
      
      labels = {
        "org.opencontainers.image.title" = "Treacherest";
        "org.opencontainers.image.description" = "Real-time multiplayer MTG Treachery game";
      };
      
      user = "1000:1000";
    };
    
    maxLayers = 25;
  };
  
  # Development container with debugging tools
  dev = n2c.nix2container.buildImage {
    name = "docker.io/grimsleydl/treacherest";
    tag = "dev";
    
    copyToRoot = nixpkgs.buildEnv {
      name = "image-root";
      paths = [
        inputs.cells.app.packages.default
        nixpkgs.bashInteractive
        nixpkgs.coreutils
        nixpkgs.curl
        nixpkgs.jq
        nixpkgs.busybox
        # Include development config
        (nixpkgs.runCommand "configs" {} ''
          mkdir -p $out/app/config
          cp ${inputs.self}/configs/server.yaml $out/app/config/server.yaml
          cp ${inputs.self}/configs/development.yaml $out/app/config/development.yaml
        '')
      ];
      pathsToLink = [ "/bin" "/static" "/app" ];
    };
    
    config = {
      entrypoint = [ "/bin/bash" ];
      cmd = [ "-c" "/bin/server" ];
      
      env = [
        "PATH=/bin"
        "SSL_CERT_FILE=${nixpkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
      ];
      
      exposedPorts = {
        "8080/tcp" = {};
      };
      
      labels = {
        "org.opencontainers.image.title" = "Treacherest Development";
        "org.opencontainers.image.description" = "Development container for Treacherest";
      };
      
      user = "0:0";
      workingDir = "/";
    };
    
    maxLayers = 50;
  };
  
  # Minimal container
  minimal = n2c.nix2container.buildImage {
    name = "docker.io/grimsleydl/treacherest";
    tag = "minimal";
    
    copyToRoot = nixpkgs.buildEnv {
      name = "image-root";
      paths = [ 
        inputs.cells.app.packages.default
        # Include minimal config
        (nixpkgs.runCommand "configs" {} ''
          mkdir -p $out/app/config
          cp ${inputs.self}/configs/server.yaml $out/app/config/server.yaml
        '')
      ];
      pathsToLink = [ "/bin" "/static" "/app" ];
    };
    
    config = {
      entrypoint = [ "/bin/server" ];
      
      env = [ "PATH=/bin" ];
      
      exposedPorts = {
        "8080/tcp" = {};
      };
      
      labels = {
        "org.opencontainers.image.title" = "Treacherest Minimal";
        "org.opencontainers.image.description" = "Minimal container with just the binary";
      };
      
      user = "65534:65534";
      workingDir = "/";
    };
    
    maxLayers = 1;
  };
}
