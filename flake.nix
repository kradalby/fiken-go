{
  description = "fiken-go — Go library, CLI, and MCP server for the Fiken API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    # Per-system outputs (packages, devShells, formatter) are merged
    # via // with system-agnostic outputs (overlays) below.
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        inherit (pkgs) lib;

        # Prefer Go 1.26 if available, else 1.25. go.mod sets the
        # language version separately; this just selects the
        # toolchain in the devShell and package builder.
        go = pkgs.go_1_26 or pkgs.go_1_25;

        # Build the fiken-go binary. The vendorHash placeholder must be
        # updated after the first `go mod tidy` produces a vendor tree.
        fiken = pkgs.buildGoModule {
          pname = "fiken-go";
          version = "0.0.0-dev";
          src = ./.;
          inherit go;
          # First-time setup: nix build will fail with a hash mismatch and
          # print the correct value — replace this placeholder with that
          # output. Until cmd/fiken/main.go exists in Plan B this also
          # has no entrypoint, so expect `nix build` to error on a
          # different message until then.
          vendorHash = "sha256-mtnfJM9FxiEEQvLxoTty+1QZjJy7tf3R7R1DcCnGLq8=";
          subPackages = [ "./..." ];
        };
      in
      {
        packages = {
          fiken = fiken;
          fiken-mcp = fiken;
          default = fiken;
        };

        # NixOS VM test for the fiken-mcp module. Plain-HTTP only;
        # tsnet can't reach the control plane inside the sandbox.
        checks = lib.optionalAttrs pkgs.stdenv.isLinux {
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
            pkgs.nixfmt-rfc-style
            pkgs.git
          ];

          shellHook = ''
            echo "fiken-go devShell — Go $(${go}/bin/go version | cut -d' ' -f3)"
          '';
        };

        formatter = pkgs.nixfmt-rfc-style;
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
