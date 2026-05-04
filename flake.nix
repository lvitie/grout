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
            SDL2
            SDL2_image
            SDL2_ttf
            SDL2_gfx
          ];

          buildPhase = ''
            export HOME=$TMP
            go build -o grout ./cmd/grout-gui
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

            # For local Fyne UI development
            pkg-config
            SDL2
            SDL2_image
            SDL2_ttf
            SDL2_gfx
            libx11
            libxxf86vm

            # Helpful tools
            git
            gnumake
          ];

          shellHook = ''
            echo "✓ Grout development environment loaded"
            echo ""
            echo "Quick commands:"
            echo "  go run ./cmd/grout-gui          # Run the UI locally"
            echo "  task code:lint                  # Lint code"
            echo "  go test ./...                   # Run tests"
            echo "  task --list                     # See all build tasks"
            echo ""
            echo "For Flatpak/Linux deployment:"
            echo "  nix build                       # Build binary locally"
            echo "  task build:arm64                # Cross-compile for ARM64 (requires Docker)"
          '';
        };

        # Development shell with cross-compilation tools
        devShells.cross = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_25
            go-task
            docker
            golangci-lint
            statix
            pkg-config
            SDL2
            SDL2_image
            SDL2_ttf
            SDL2_gfx
            qemu
          ];

          shellHook = ''
            echo "✓ Grout cross-compilation environment loaded"
            echo ""
            echo "Use these tasks:"
            echo "  task build:arm64                # Cross-compile for ARM64"
            echo "  task build:arm32                # Cross-compile for ARM32"
            echo "  task all-arm64                  # Build and package for all ARM64 platforms"
            echo ""
            echo "Note: Docker must be running for cross-compilation"
          '';
        };
      }
    );
}
