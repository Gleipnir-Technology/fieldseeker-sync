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
                "cmd/webserver"
        ];
        version = "0.0.12";
        # Needs to be updated after every modification of go.mod/go.sum
        vendorHash = "sha256-i44ge0/x8++tpUGUB+94klL2TR4fxUkmC1ZabmPRBX0=";
}
