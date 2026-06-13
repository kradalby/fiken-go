{ config
, lib
, pkgs
, ...
}:
let
  cfg = config.services.fiken-mcp;
in
{
  options.services.fiken-mcp = {
    enable = lib.mkEnableOption "fiken-mcp, the Fiken Model Context Protocol server";

    package = lib.mkPackageOption pkgs "fiken-mcp" { };

    user = lib.mkOption {
      type = lib.types.str;
      default = "fiken-mcp";
      description = ''
        User account under which fiken-mcp runs.

        If left as the default value this user is created automatically
        on system activation.
      '';
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "fiken-mcp";
      description = ''
        Group under which fiken-mcp runs.

        If left as the default value this group is created automatically
        on system activation.
      '';
    };

    mode = lib.mkOption {
      type = lib.types.enum [
        "read-only"
        "read-write"
      ];
      default = "read-only";
      description = ''
        Tool exposure for the plain HTTP / stdio transports. Ignored when
        `tsnet.enable` is true (tsnet exposes every tool and per-request
        capability grants gate writes).
      '';
    };

    listen = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1:8765";
      description = ''
        Plain-HTTP listen address. Ignored when `tsnet.enable` is true.
      '';
    };

    tsnet = {
      enable = lib.mkEnableOption "exposing fiken-mcp on a Tailscale tailnet (disables plain HTTP)";

      hostname = lib.mkOption {
        type = lib.types.str;
        default = "fiken-mcp";
        description = "Tailscale device name advertised by tsnet.";
      };

      authKeyFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          Path to a file containing a Tailscale pre-auth key. Only required
          for first registration; state in `dataDir/tsnet` persists across
          restarts and subsequent runs do not need a key.
        '';
      };
    };

    fikenTokenFile = lib.mkOption {
      type = lib.types.path;
      description = ''
        Path to a file containing the Fiken API token. Contents are loaded
        into the FIKEN_TOKEN environment variable at service start.
      '';
    };

    environmentFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = ''
        Optional EnvironmentFile sourced verbatim by systemd. Use for
        additional secrets or operator-defined `FIKEN_*` overrides.
      '';
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/fiken-mcp";
      description = ''
        State directory. tsnet keys live under `''${dataDir}/tsnet`.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    assertions = [
      {
        assertion = !(cfg.tsnet.enable && cfg.listen != "127.0.0.1:8765");
        message = "services.fiken-mcp: cannot set both tsnet.enable and a custom `listen` address. tsnet replaces the plain HTTP listener.";
      }
    ];

    users.users = lib.mkIf (cfg.user == "fiken-mcp") {
      fiken-mcp = {
        isSystemUser = true;
        group = cfg.group;
        home = cfg.dataDir;
        description = "fiken-mcp service user";
      };
    };

    users.groups = lib.mkIf (cfg.group == "fiken-mcp") {
      fiken-mcp = { };
    };

    systemd.services.fiken-mcp = {
      description = "Fiken MCP server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network-online.target" ];
      wants = [ "network-online.target" ];

      script = ''
        FIKEN_TOKEN="$(cat ${lib.escapeShellArg cfg.fikenTokenFile})"
        export FIKEN_TOKEN
        ${lib.optionalString (cfg.tsnet.enable && cfg.tsnet.authKeyFile != null) ''
          if [ -r ${lib.escapeShellArg cfg.tsnet.authKeyFile} ]; then
            FIKEN_TSNET_AUTHKEY="$(cat ${lib.escapeShellArg cfg.tsnet.authKeyFile})"
            export FIKEN_TSNET_AUTHKEY
          fi
        ''}
        exec ${cfg.package}/bin/fiken mcp ${
          if cfg.tsnet.enable then
            "--tsnet --tsnet-hostname ${lib.escapeShellArg cfg.tsnet.hostname} --tsnet-state-dir ${lib.escapeShellArg "${cfg.dataDir}/tsnet"}"
          else
            "--transport=http --listen ${lib.escapeShellArg cfg.listen} --mode=${cfg.mode}"
        }
      '';

      serviceConfig = {
        Type = "simple";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = "5s";
        StateDirectory = "fiken-mcp";
        StateDirectoryMode = "0750";
        EnvironmentFile = lib.optional (cfg.environmentFile != null) cfg.environmentFile;
        ReadWritePaths = [ cfg.dataDir ];

        # Hardening
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectKernelLogs = true;
        ProtectControlGroups = true;
        ProtectHostname = true;
        ProtectClock = true;
        ProtectProc = "invisible";
        ProcSubset = "pid";
        RestrictSUIDSGID = true;
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
          "AF_UNIX"
          "AF_NETLINK"
        ];
        LockPersonality = true;
        UMask = "0077";
        CapabilityBoundingSet = [ "" ];
        AmbientCapabilities = [ ];
        SystemCallFilter = [
          "@system-service"
          "~@privileged"
        ];
        SystemCallArchitectures = "native";
      };
    };
  };
}
