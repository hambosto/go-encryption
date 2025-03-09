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
      pname = "go-encryption";
      version = "1.0";

      nixosModule =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.programs.${pname};
        in
        {
          options.programs.${pname}.enable = lib.mkEnableOption "Enable the ${pname} application";

          config = lib.mkIf cfg.enable {
            home.packages = [ self.packages.${pkgs.system}.default ];
          };
        };

      perSystem =
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          packages.default = pkgs.buildGoModule {
            inherit pname version;
            src = ./.;
            vendorHash = "sha256-nxRo/spwxVE+B41znEJEWuHozxJ6dc/BAAFRU5TIYuk=";
            env.CGO_ENABLED = 0;
            ldflags = [
              "-extldflags '-static'"
              "-s -w"
            ];
          };

          apps.default = flake-utils.lib.mkApp {
            drv = self.packages.${system}.default;
            name = pname;
          };
        };
    in
    flake-utils.lib.eachDefaultSystem perSystem
    // {
      nixosModules.${pname} = nixosModule;
    };
}
