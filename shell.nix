{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
	buildInputs = [
		pkgs.air
		pkgs.go
		pkgs.goose
		pkgs.gotools
		pkgs.lefthook
	];
}
