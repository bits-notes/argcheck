#!/usr/bin/env bash
# Build and install argcheck, the deterministic numbered-argument gate CLI.
# Default target is ~/.local/bin; override with BIN_DIR=. The binary reports
# the commit it was built from.
set -euo pipefail

cd "$(dirname "$0")"

bin_dir="${BIN_DIR:-$HOME/.local/bin}"
commit="$(git rev-parse --short=7 HEAD)"
if [[ -n "$(git status --porcelain -- cmd/argcheck validation)" ]]; then
  commit="$commit-dirty"
fi

mkdir -p "$bin_dir" ./tmp
go build -trimpath -ldflags="-w -s -X main.commit=$commit" -o ./tmp/argcheck ./cmd/argcheck
install ./tmp/argcheck "$bin_dir/argcheck"

printf 'Installed argcheck (argcheck@%s) to %s\n' "$commit" "$bin_dir/argcheck"
