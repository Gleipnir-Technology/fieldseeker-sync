{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.goose
    pkgs.ninja
    pkgs.nodejs
    pkgs.pre-commit
    pkgs.postgresql
    pkgs.python3
  ];
}
