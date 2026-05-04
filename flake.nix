# Grout Nix Flake Configuration
#
# Quick start:
#   nix develop          # Enter development shell with all dependencies
#   nix build            # Build the grout binary (local build)
#   nix run              # Run grout directly
#
# Note: The package build uses the development approach. For Flatpak distribution,
# use `task build:arm64` or the standard Go build in the dev environment.

{
  description = "Grout - A RomM client for retro gaming";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        # Minimal package: builds grout using the dev environment approach
        packages.default = pkgs.stdenv.mkDerivation {
          pname = "grout";
          version = (builtins.fromJSON (builtins.readFile ./pak.json)).version;
          src = ./.;

          nativeBuildInputs = with pkgs; [
            go_1_25
            pkg-config
          ];

          buildInputs = with pkgs; [
            gtk4
            libadwaita
            gobject-introspection
          ];

          buildPhase = ''
            export HOME=$TMP
            go build -o grout ./cmd/grout-desktop
          '';

          installPhase = ''
            mkdir -p $out/bin
            cp grout $out/bin/
          '';

          meta = with pkgs.lib; {
            description = "A RomM client for retro gaming";
            homepage = "https://github.com/rommapp/grout";
            license = licenses.gpl3Plus;
            maintainers = [ ];
            platforms = platforms.linux;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_25

            # Build & task runner
            go-task
            docker

            # Code quality
            golangci-lint
            statix

            # Debugger (optional)
            delve

            # For GTK4 UI development
            pkg-config
            gtk4
            libadwaita
            gobject-introspection
            libx11
            libxxf86vm

            # Helpful tools
            git
            gnumake
          ];

          shellHook = ''
            export CGO_CFLAGS="-Wno-builtin-declaration-mismatch"
            echo "✓ Grout development environment loaded"
            echo ""
            echo "Quick commands:"
            echo "  go run ./cmd/grout-desktop      # Run the UI locally"
            echo "  task code:lint                  # Lint code"
            echo "  go test ./...                   # Run tests"
            echo "  task --list                     # See all build tasks"
            echo ""
            echo "For Flatpak/Linux deployment:"
            echo "  nix build                       # Build binary locally"
            echo "  task build:arm64                # Cross-compile for ARM64 (requires Docker)"
          '';
        };

      }
    );
}
