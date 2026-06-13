{
  description = "fiken-go — Go library, CLI, and MCP server for the Fiken API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    flake-checks.url = "github:kradalby/flake-checks";
    flake-checks.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , flake-checks
    ,
    }:
    # Per-system outputs (packages, devShells, formatter) are merged
    # via // with system-agnostic outputs (overlays) below.
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        inherit (pkgs) lib;
        fc = flake-checks.lib;

        # Prefer Go 1.26 if available, else 1.25. go.mod sets the
        # language version separately; this just selects the
        # toolchain in the devShell and package builder.
        go = pkgs.go_1_26 or pkgs.go_1_25;

        # Shared context for the flake-checks Go helpers. i18n/i18n.go
        # embeds locale files via //go:embed, so the locale directory must
        # be part of the fileset-filtered source or the build fails with
        # "no matching files found".
        common = {
          inherit pkgs;
          root = ./.;
          pname = "fiken-go";
          version = "0.0.1";
          vendorHash = "sha256-mtnfJM9FxiEEQvLxoTty+1QZjJy7tf3R7R1DcCnGLq8=";
          goPkg = go;
          embedDirs = [ (./. + "/i18n/locales") ];
        };

        fiken = fc.goBuild common;
      in
      {
        packages = {
          fiken = fiken;
          fiken-mcp = fiken;
          default = fiken;
        };

        formatter = fc.formatter common;

        checks = {
          build = fc.goBuild common;
          gotest = fc.goTest common;
          golangci-lint = fc.goLint common;
          formatting = fc.goFormat common;
        }
        # NixOS VM test for the fiken-mcp module. Plain-HTTP only;
        # tsnet can't reach the control plane inside the sandbox.
        // lib.optionalAttrs pkgs.stdenv.isLinux {
          fiken-mcp-module = pkgs.testers.nixosTest (
            import ./nix/tests/fiken-mcp.nix {
              inherit pkgs;
              inherit (pkgs) lib;
              fiken-mcp = fiken;
              module = ./nix/module.nix;
            }
          );
        };

        devShells.default = pkgs.mkShell {
          packages = [
            go
            pkgs.gopls
            pkgs.gofumpt
            pkgs.gotools # provides goimports
            pkgs.golangci-lint
            pkgs.gotestsum
            pkgs.gotests
            pkgs.difftastic
            pkgs.prek
            pkgs.ogen
            pkgs.prettier
            pkgs.nixpkgs-fmt
            pkgs.git
          ];

          shellHook = ''
            echo "fiken-go devShell — Go $(${go}/bin/go version | cut -d' ' -f3)"
          '';
        };
      }
    ))
    // {
      # System-agnostic outputs live outside eachDefaultSystem so
      # `inputs.fiken-go.overlays.default` resolves correctly for any
      # consumer regardless of their `system`.
      overlays.default = final: prev: {
        fiken = self.packages.${prev.system}.fiken;
        fiken-mcp = self.packages.${prev.system}.fiken-mcp;
      };

      nixosModules = {
        fiken-mcp = ./nix/module.nix;
        default = ./nix/module.nix;
      };
    };
}
