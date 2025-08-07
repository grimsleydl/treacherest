{
  inputs,
  cell,
}: let
  inherit (inputs.nixpkgs) lib;
  pkgs = import inputs.nixpkgs {
    inherit (inputs.nixpkgs) system;
    overlays = [
      inputs.gomod2nix.overlays.default
    ];
  };
  
  # Get templ from the flake input
  templ = inputs.templ.packages.${pkgs.system}.templ;
  
  # Filter source using gitignore
  src = pkgs.nix-gitignore.gitignoreSource [] (inputs.self + /nix/app);
  
  # Card images fetcher - needed for go:embed
  cardImages = import ./image-fetcher.nix {
    inherit pkgs;
    go = pkgs.go;
    treachery-cards-json = inputs.self + /nix/app/static/treachery-cards.json;
  };
in {
  default = pkgs.buildGoApplication {
    pname = "mtg-treacherest";
    version = "0.1.0";
    
    inherit src;
    
    # Go module dependencies
    modules = ./gomod2nix.toml;
    
    # Set the main package path
    subPackages = ["cmd/server"];
    
    nativeBuildInputs = [ templ ];
    
    # Disable tests
    doCheck = false;
    
    # Generate templ and copy assets for go:embed
    preBuild = let
      # Build CSS using buildNpmPackage inline
      outputCSS = (pkgs.buildNpmPackage {
        pname = "treacherest-css-builder";
        version = "1.0.0";
        src = src;
        npmDepsHash = "sha256-AN3Hj9Lk1JXAGau3BH1g3zPdZyDw/DwrYxJIP2nEDow=";
        nativeBuildInputs = [ templ ];
        buildPhase = ''
          echo "Starting CSS build at $(date)"
          # Generate templ files for Tailwind to scan
          echo "Generating templ files..."
          templ generate
          echo "Templ generation done at $(date)"
          # Build minified CSS
          echo "Building CSS with PostCSS..."
          NODE_ENV=production npx postcss static/css/input.css -o build/output.css
          echo "CSS build done at $(date)"
        '';
        installPhase = ''
          cp build/output.css $out
        '';
      });
    in ''
      # Generate templ files
      ${templ}/bin/templ generate
      
      # Copy card images for go:embed
      mkdir -p ./static/images
      cp -r ${cardImages}/static/images/cards ./static/images/
      
      # Copy pre-built CSS
      mkdir -p ./static/css
      cp ${outputCSS} ./static/css/output.css
    '';
    
    # Install static files (CSS, JS, JSON)
    # Note: Images are embedded in the binary via go:embed, no need to install them
    postInstall = ''
      # Copy only the files we need (not images)
      mkdir -p $out/static/css
      cp ./static/css/output.css $out/static/css/
      cp ./static/css/input.css $out/static/css/  # for debugging
      
      # Copy JSON files
      cp ./static/*.json $out/static/ || true
      
      # Copy JS files if any
      mkdir -p $out/static/js
      cp ./static/js/* $out/static/js/ 2>/dev/null || true
    '';
    
    meta = with lib; {
      description = "MTG Treachery game implementation";
      license = licenses.mit;
      platforms = platforms.linux;
      mainProgram = "server";
    };
  };
}