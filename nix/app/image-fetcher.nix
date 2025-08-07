{ pkgs, go, treachery-cards-json }:

pkgs.stdenv.mkDerivation {
  name = "mtg-treachery-card-images";
  
  src = ./scripts/download;

  nativeBuildInputs = [ go ];

  # The go script needs the json file to know what to download.
  # We'll copy it into the build environment.
  # The script looks in ../../docs/external, so we'll create that structure.
  buildPhase = ''
    runHook preBuild
    
    # Set HOME and GOCACHE to writable, absolute paths within the sandbox.
    export HOME=$(pwd)
    export GOCACHE=$(pwd)/.go-cache
    mkdir -p $GOCACHE

    cp ${treachery-cards-json} ./treachery-cards.json
    
    go run download_cards.go
    
    runHook postBuild
  '';

  # The script downloads files to a 'cards' directory. We need to install it.
  installPhase = ''
    runHook preInstall
    
    mkdir -p $out/static/images
    mv cards $out/static/images/
    
    runHook postInstall
  '';

  # This is a fixed-output derivation
  outputHash = "sha256-4YBITZeMVZmEJQXwUg1RG/9BFI3LUsxHh7FaczBcypg=";
  outputHashAlgo = "sha256";
  outputHashMode = "recursive";
}
