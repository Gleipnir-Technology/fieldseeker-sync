#! /usr/bin/env nix-shell
#! nix-shell -i bash -p bash

# Strict settings
set -o errexit
set -o pipefail
set -o nounset

export PSQL_DSN="postgresql://fieldseeker-sync:@?host=/var/run/postgresql&sslmode=disable"
pushd database
go run github.com/stephenafamo/bob/gen/bobgen-psql@latest
popd

