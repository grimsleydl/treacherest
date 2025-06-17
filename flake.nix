{
  description = "treacherest";
  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";

    std.url = "github:divnix/std";
    std.inputs.nixpkgs.follows = "nixpkgs";

    std.inputs.devshell.url = "github:numtide/devshell";
    std.inputs.nixago.url = "github:nix-community/nixago";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {
    self,
    std,
    ...
  }:
    std.growOn {
      inherit inputs;
      nixpkgsConfig = {allowUnfree = true;};
      systems = ["x86_64-linux"];
      cellsFrom = ./nix;
      cellBlocks = with std.blockTypes; [
        (installables "packages")
        # Development Environments
        (nixago "config")
        (devshells "shells")
        # (functions "extensions")
        (functions "env")
        (pkgs "pkgs")
        (containers "containers")
      ];
    } {
      devShells = std.harvest self ["local" "shells"];
      packages = std.harvest self ["app" "packages"];
    };

  nixConfig = {
    extra-substituters = [
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];
  };
}
