{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.goose
    pkgs.gotools
    pkgs.lefthook
    pkgs.watchexec
  ];
}
