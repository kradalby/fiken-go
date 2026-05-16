# Example NixOS configurations for services.fiken-mcp.
#
# Pick one of the two profiles below, copy it into your host config,
# and provide the secret files referenced by `fikenTokenFile` and
# (for the tsnet profile) `tsnet.authKeyFile`.
{ inputs, ... }:
{
  imports = [ inputs.fiken-go.nixosModules.fiken-mcp ];

  nixpkgs.overlays = [ inputs.fiken-go.overlays.default ];

  # ---------------------------------------------------------------
  # Profile A: plain HTTP, bound to loopback. Pair with an SSH tunnel
  # or reverse proxy for remote access.
  # ---------------------------------------------------------------
  services.fiken-mcp = {
    enable = true;
    mode = "read-only";
    listen = "127.0.0.1:8765";
    fikenTokenFile = "/run/secrets/fiken-token";
  };

  # ---------------------------------------------------------------
  # Profile B: Tailscale tsnet. Listens only on the tailnet interface;
  # the plain HTTP listener is disabled. Read access is implicit for
  # any tailnet peer; write access requires a capability grant under
  # `kradalby.no/cap/fiken-mcp` in your tailnet ACL policy:
  #
  # {
  #   "grants": [{
  #     "src": ["autogroup:admin"],
  #     "dst": ["tag:fiken-mcp"],
  #     "app": { "kradalby.no/cap/fiken-mcp": [{"write": true}] }
  #   }]
  # }
  # ---------------------------------------------------------------
  #
  # services.fiken-mcp = {
  #   enable = true;
  #   tsnet.enable = true;
  #   tsnet.hostname = "fiken-mcp";
  #   tsnet.authKeyFile = "/run/secrets/tailscale-authkey";
  #   fikenTokenFile = "/run/secrets/fiken-token";
  # };
}
