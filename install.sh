#!/usr/bin/env bash
# Build and install english-arg-validator (formerly argcheck), the deterministic
# numbered-argument gate CLI. Default target is ~/.local/bin; override with
# BIN_DIR=. The binary reports the commit it was built from. An `argcheck`
# symlink is installed beside it so the old command name keeps working.
set -euo pipefail

cd "$(dirname "$0")"

bin_dir="${BIN_DIR:-$HOME/.local/bin}"
commit="$(git rev-parse --short=7 HEAD)"
if [[ -n "$(git status --porcelain -- cmd/english-arg-validator validation)" ]]; then
  commit="$commit-dirty"
fi

mkdir -p "$bin_dir" ./tmp
go build -trimpath -ldflags="-w -s -X main.commit=$commit" -o ./tmp/english-arg-validator ./cmd/english-arg-validator
install ./tmp/english-arg-validator "$bin_dir/english-arg-validator"
# The old name: replace a previously installed argcheck binary with a symlink.
ln -sf english-arg-validator "$bin_dir/argcheck"

printf 'Installed english-arg-validator (english-arg-validator@%s) to %s\n' "$commit" "$bin_dir/english-arg-validator"
printf 'argcheck -> english-arg-validator (old name, kept as an alias)\n'
