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
		"cmd/convert-completed-tasks"
		"cmd/download-schema"
		"cmd/dump"
		"cmd/full-export"
		"cmd/label-task"
		"cmd/login"
		"cmd/migrate"
		"cmd/registration"
		"cmd/webserver"
	];
	version = "0.0.27";
	# Needs to be updated after every modification of go.mod/go.sum
	vendorHash = "sha256-Z1pIVdRtl13RxbsSKyJKM3bvE2MFPj7KXZjt5jiEPDE=";
}
