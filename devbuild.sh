#!/usr/bin/env bash
# ScraperGo-style local build counter. No tags, releases, or remote mutations.
set -Eeuo pipefail
cd "$(dirname "$0")"
command -v git >/dev/null || { echo 'Git required' >&2; exit 1; }
command -v go >/dev/null || { echo 'Go required' >&2; exit 1; }
[[ $# -eq 0 ]] || { echo 'Usage: ./devbuild.sh' >&2; exit 2; }
printf 'Running Go tests...\n'
go test -count=1 ./...
# Counter is local to the Git checkout and never committed or reset by branch changes.
counter="$(git rev-parse --git-path mauristocks-build-counter)"
mkdir -p "$(dirname "$counter")"
# Prior published/development UI build was v40. Seed there ONLY when there is
# no existing local counter, to avoid accidentally reusing an older displayed ID.
previous=40
if [[ -f "$counter" ]]; then
 previous="$(cat "$counter")"
 [[ "$previous" =~ ^[0-9]+$ ]] || { echo 'Corrupt build counter' >&2; exit 1; }
fi
next=$((10#$previous + 1))
[[ "$next" -le 999999 ]] || { echo 'Build number exhausted' >&2; exit 1; }
printf '%d\n' "$next" > "${counter}.tmp"
mv "${counter}.tmp" "$counter"
printf -v build_label 'Build %06d' "$next"
printf 'Building MauriStocks %s...\n' "$build_label"
go build -trimpath -buildvcs=false -ldflags="-X 'main.appBuild=$build_label'" -o mauristocks .
./mauristocks -version
