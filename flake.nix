{
  description = "A Flake for Applying Grackle to gokv";

  inputs = {
    nixpkgs.url = "nixpkgs";
  };

  outputs = {nixpkgs, ...}: let
    system = "x86_64-linux";
  in {
    devShells."${system}".default = let
      pkgs = import nixpkgs {
        inherit system;
      };
      goose = pkgs.buildGoModule {
        name = "goose";
        src = pkgs.fetchFromGitHub {
          owner = "goose-lang";
          repo = "goose";
          rev = "8d13c771b9a80957089f7c5b0ee2ccf58e5eb06f";
          sha256 = "1fbqs75ya4as3my2knkaq4m0crdh3n004grw5g5iczvb5h5k06lz";
        };
        vendorHash = "sha256-HCJ8v3TSv4UrkOsRuENWVz5Z7zQ1UsOygx0Mo7MELzY=";
      };
      grackle = pkgs.buildGoModule {
        name = "grackle";
        src = pkgs.fetchFromGitHub {
          owner = "mjschwenne";
          repo = "grackle";
          rev = "d961e849c7737781b6a09a18c9f3cc669178f98a";
          sha256 = "1385sinsmm1k58blvv2sf951fi7kihka5g7zhg0m8r5ijzy9ac91";
        };
        vendorHash = "sha256-Wk2v0HSAkrzxHJvCfbw6xOn0OQ1xukvYjDxk3c2LmH8=";
        checkPhase = false;
      };
    in
      pkgs.mkShell {
        # create an environment with the required coq libraries
        packages = with pkgs; [
          # Go deps
          go
          gopls
          goose
          grackle

          # Protobuf deps
          protobuf
          protoc-gen-go
          proto-contrib
          protoscope

          # nix tools
          nix-prefetch-git
          nix-prefetch
          update-nix-fetchgit
        ];

        shellHook = ''
        '';
      };
  };
}
