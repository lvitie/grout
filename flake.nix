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
        # Standard Go package build
        packages.default = pkgs.buildGoModule {
          pname = "grout";
          version = (builtins.fromJSON (builtins.readFile ./pak.json)).version;
          src = ./.;

          vendorHash = "sha256-bZUUsUSBOX56hlnipS0jexqpwNMBGvKMjxQMw8rHi08=";

          nativeBuildInputs = with pkgs; [
            pkg-config
          ];

          buildInputs = with pkgs; [
            gtk4
            libadwaita
            gobject-introspection
            librsvg
          ];

          subPackages = [ "cmd/grout-desktop" ];

          # Handle CGo and GTK4 dependencies
          CGO_CFLAGS = "-Wno-builtin-declaration-mismatch";

          postInstall = ''
            mv $out/bin/grout-desktop $out/bin/grout
            mkdir -p $out/share/applications
            mkdir -p $out/share/icons/hicolor/512x512/apps
            cp resources/app.romm.Grout.desktop $out/share/applications/app.romm.Grout.desktop
            cp resources/app.romm.Grout.png $out/share/icons/hicolor/512x512/apps/app.romm.Grout.png
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
            librsvg
            libx11
            libxxf86vm

            # Flatpak building
            flatpak-builder

            # Helpful tools
            git
            gnumake
          ];

          shellHook = ''
            export CGO_CFLAGS="-Wno-builtin-declaration-mismatch"
            
            # Fix GDK pixbuf loaders for SVG support during 'go run'
            export GDK_PIXBUF_MODULE_FILE=$(echo ${pkgs.librsvg}/lib/gdk-pixbuf-2.0/*/loaders.cache)
            
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
