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
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        go = pkgs.go_1_26 or pkgs.go_1_25;

        fiken = pkgs.buildGoModule {
          pname = "fiken-go";
          version = "0.0.0-dev";
          src = ./.;
          inherit go;
          vendorHash = "sha256-sVh7G41xdcoX9YlAHUxqxw52auA2pRKHSuf99ACqoLY=";
          subPackages = [ "./..." ];
        };
      in
      {
        packages = {
          fiken = fiken;
          fiken-mcp = fiken;
          default = fiken;
        };

        devShells.default = pkgs.mkShell {
          packages = [
            go
            pkgs.gopls
            pkgs.gofumpt
            pkgs.gotools
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
      overlays.default = final: prev: {
        fiken = self.packages.${prev.system}.fiken;
        fiken-mcp = self.packages.${prev.system}.fiken-mcp;
      };
    };
}
