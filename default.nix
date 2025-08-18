{ pkgs ? import <nixpkgs> { } }:
pkgs.buildGoModule rec {
        meta = {
                description = "FieldSeeker sync";
                homepage = "https://github.com/Gleipnir-Technology/fieldseeker-sync";
        };
        pname = "fieldseeker-sync";
        src = ./.;
        subPackages = [
                "cmd/download-schema"
                "cmd/dump"
                "cmd/full-export"
                "cmd/login"
                "cmd/registration"
                "cmd/schema"
                "cmd/webserver"
        ];
        version = "0.0.1";
        # Needs to be updated after every modification of go.mod/go.sum
        vendorHash = "sha256-mZk70L1JYzLmDSJi1kjVCBxOuMmRGHALR3WFyGiLzJQ=";
}
