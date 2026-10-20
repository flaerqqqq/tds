#!/usr/bin/env bash
# Builds tds and installs it to ~/.local/bin. Requires Go.
# Usage: install.sh [--uninstall]
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bin_dir="${HOME}/.local/bin"
target="$bin_dir/tds"
marker="${XDG_DATA_HOME:-$HOME/.local/share}/tds/installed-from"

case "${1:-}" in
--uninstall)
    if [[ -f "$marker" && "$(cat "$marker")" == "$root" ]]; then
        rm -f "$target" "$marker"
        echo "Removed $target"
    elif [[ -e "$target" || -L "$target" ]]; then
        echo "Not removed: $target was not installed from $root" >&2
        exit 1
    else
        echo "Nothing to remove: $target does not exist"
    fi
    exit 0
    ;;
"") ;;
*)
    echo "Usage: install.sh [--uninstall]" >&2
    exit 1
    ;;
esac

if ! command -v go &>/dev/null; then
    echo "Error: Go is required to build tds." >&2
    exit 1
fi

mkdir -p "$bin_dir"
# Remove a leftover symlink first: go build would otherwise write through it.
if [[ -L "$target" ]]; then
    rm "$target"
fi
(cd "$root" && go build -o "$target" ./cmd/tds)
mkdir -p "$(dirname "$marker")"
echo "$root" >"$marker"
echo "Installed $target"

case ":$PATH:" in
*":$bin_dir:"*) ;;
*) echo "Note: $bin_dir is not on PATH." ;;
esac
