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
          rev = "a4f2f84193d34f56dd84fc623adc43a6441da1eb";
          sha256 = "1b1dfa1qsv2h7hy5x20zhic2npr5gz1zp76m1lab4v490adxj2rx";
        };
        vendorHash = "sha256-HCJ8v3TSv4UrkOsRuENWVz5Z7zQ1UsOygx0Mo7MELzY=";
      };
      grackle = pkgs.buildGoModule {
        name = "grackle";
        src = pkgs.fetchFromGitHub {
          owner = "mjschwenne";
          repo = "grackle";
          rev = "101412356cdfbcad78f8aaa724101312928c4978";
          sha256 = "06zf2bvrbbjhgrd6994h3wcaml7m83m6f9r61pj7y09xq9nw10br";
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
