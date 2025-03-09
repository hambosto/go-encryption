{
  description = "Secure file encryption and decryption CLI tool built with Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    let

      nixosModule =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.programs.go-encryption;
        in
        {
          options.programs.go-encryption = {
            enable = lib.mkEnableOption "Enable the Go Encryption CLI tool";
          };

          config = lib.mkIf cfg.enable {
            home.packages = [ self.packages.${pkgs.system}.default ];
          };
        };

      perSystem =
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};

          package = pkgs.buildGoModule {
            pname = "go-encryption";
            version = "1.0";

            src = ./.;

            vendorHash = "sha256-nxRo/spwxVE+B41znEJEWuHozxJ6dc/BAAFRU5TIYuk=";

            env.CGO_ENABLED = 0;

            ldflags = [
              "-extldflags '-static'"
              "-s -w"
            ];
          };
        in
        {
          packages = {
            default = package;
            go-encryption = package;
          };
        };
    in
    flake-utils.lib.eachDefaultSystem perSystem
    // {
      nixosModules = {
        default = nixosModule;
        go-encryption = nixosModule;
      };
    };
}
