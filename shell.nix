{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.goose
    pkgs.pre-commit
    pkgs.python3
  ];
}
