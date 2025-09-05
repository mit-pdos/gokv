{
  description = "Gokv, a replicated, verified key-value store";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/a595dde4d0d31606e19dcec73db02279db59d201";
    flake-utils.url = "github:numtide/flake-utils";
    grackle.url = "github:mjschwenne/grackle";
  };

  outputs = {
    nixpkgs,
    flake-utils,
    grackle,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        devShells.default = with pkgs;
          mkShell {
            buildInputs = [
              go
              gopls
              grackle.packages.${system}.default
              grackle.packages.${system}.goose
              protobuf
              protoc-gen-go
              proto-contrib
              protoscope

              # nix helpers
              nix-update
            ];

            shellHook = ''
            '';
          };
      }
    );
}
