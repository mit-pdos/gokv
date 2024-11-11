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
          rev = "8352f2a82040a814b60e0dd9ac75f09cec3dd619";
          hash = "sha256-jrcV/J70yzkmbXUiBsYYswqIph1mNQCC2LwvsxFnSSs=";
        };
        vendorHash = "sha256-HCJ8v3TSv4UrkOsRuENWVz5Z7zQ1UsOygx0Mo7MELzY=";
      };
      grackle = pkgs.buildGoModule {
        name = "grackle";
        src = pkgs.fetchFromGitHub {
          owner = "mjschwenne";
          repo = "grackle";
          rev = "18dbbd313c299d74895f566aae028eeba62a699a";
          hash = "sha256-GDo3c3VbRl0FR7sJ2Cvmdau+rBRcxwb/sGAEwV5seFQ=";
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
        ];

        shellHook = ''
        '';
      };
  };
}
