{ pkgs ? import <nixpkgs> { } }:
pkgs.buildGoModule rec {
	meta = {
		description = "FieldSeeker sync";
		homepage = "https://github.com/Gleipnir-Technology/fieldseeker-sync";
	};
	pname = "fieldseeker-sync";
	src = ./.;
	subPackages = [
		"cmd/audio-post-processor"
		"cmd/download-schema"
		"cmd/dump"
		"cmd/full-export"
		"cmd/login"
		"cmd/migrate"
		"cmd/registration"
		"cmd/webserver"
	];
	version = "0.0.22";
	# Needs to be updated after every modification of go.mod/go.sum
	vendorHash = "sha256-0JKxQ+r4ri/BaLVshy47HEGSw5Em6Q2kYY9Z+xNEsZM=";
}
