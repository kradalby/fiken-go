{
  pkgs,
  lib,
  fiken-mcp,
  module,
  ...
}:
{
  name = "fiken-mcp";
  meta.maintainers = with lib.maintainers; [ kradalby ];

  nodes.machine =
    { ... }:
    {
      imports = [ module ];

      nixpkgs.overlays = [ (final: prev: { fiken-mcp = fiken-mcp; }) ];

      # Minimal token file — the MCP server starts and serves
      # tools/list without ever calling the Fiken API in this test.
      environment.etc."fiken-token".text = "dummy-token-for-vm-test";

      services.fiken-mcp = {
        enable = true;
        mode = "read-only";
        listen = "127.0.0.1:8765";
        fikenTokenFile = "/etc/fiken-token";
      };

      environment.systemPackages = [ pkgs.curl ];
    };

  testScript = ''
    machine.start()
    machine.wait_for_unit("fiken-mcp.service")
    machine.wait_for_open_port(8765)

    # The listener must answer SOMETHING beyond a TCP error. The
    # streamable-HTTP transport negotiates over SSE/POST; we just
    # confirm it responds with an HTTP status code (any 2xx-4xx).
    machine.succeed(
        "curl -sS -o /dev/null -w '%{http_code}\\n' http://127.0.0.1:8765/ | grep -qE '^(200|400|404|405|406)$'"
    )
  '';
}
