{
  description = "fiken-go — Go library, CLI, and MCP server for the Fiken API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    flake-checks.url = "github:kradalby/flake-checks";
    flake-checks.inputs.nixpkgs.follows = "nixpkgs";
    flake-checks.inputs.flake-utils.follows = "flake-utils";
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
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            (_final: prev: {
              # goimports (shipped in gotools) shells out to `go` for type
              # information, and nixpkgs wraps it with `pkgs.go` pinned onto
              # PATH. `pkgs.go` still trails the release, so a 1.26 `go` meets
              # `go 1.27.0` in go.mod and tries to fetch the newer toolchain.
              # The formatting check runs in the network-less nix sandbox,
              # where that fetch cannot succeed. `go` fixes the wrapper's PATH;
              # `buildGoModule` compiles the tool with the matching toolchain.
              gotools = prev.gotools.override {
                buildGoModule = prev.buildGoLatestModule;
                go = prev.go_latest;
              };
            })
          ];
        };
        inherit (pkgs) lib;
        fc = flake-checks.lib;

        # Track the newest Go nixpkgs ships rather than pinning a
        # version that goes stale. `pkgs.go` still resolves to the
        # previous release, so the `_latest` attribute is required.
        # flake-checks feeds this to `buildGoModule.override { go = ...; }`,
        # which is what `buildGoLatestModule` is. go.mod sets the language
        # version separately; this selects the toolchain for the devShell,
        # the package builder, and the checks.
        go = pkgs.go_latest;

        # Shared context for the flake-checks Go helpers. i18n/i18n.go
        # embeds locale files via //go:embed, so the locale directory must
        # be part of the fileset-filtered source or the build fails with
        # "no matching files found".
        common = {
          inherit pkgs;
          root = ./.;
          pname = "fiken-go";
          version = "0.0.1";
          vendorHash = "sha256-btXoAy+wOeVxwp5QLrhOp3+Wu8YRtcOulKZT5rg6gNA=";
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

          # `go generate ./...` drift. Three generators run: ogen over
          # api/fiken-openapi.yaml, plus cmd/fiken-mutating-gen and
          # cmd/fiken-i18n-scaffold over the same spec. The spec is not a Go
          # file, so the fileset-filtered source has to opt it in explicitly
          # or the generators fail on a missing input.
          #
          # ogen comes from the `tool` directive in go.mod, so `go run
          # github.com/ogen-go/ogen/cmd/ogen` resolves out of the vendored
          # tree and regenerates with exactly the version go.mod pins rather
          # than whatever nixpkgs happens to ship. gofumpt and goimports are
          # the last two //go:generate lines in fiken/doc.go and need to be
          # on PATH.
          #
          # Until now this was only guarded by the local `codegen-clean`
          # pre-commit hook, which does nothing for a commit pushed from a
          # machine without the hook installed.
          generate = fc.goGenerate (
            common
            // {
              extraSrc = [ (./. + "/api/fiken-openapi.yaml") ];
              nativeCheckInputs = [ pkgs.gofumpt pkgs.gotools ];
            }
          );
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
