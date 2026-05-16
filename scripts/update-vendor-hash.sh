#!/usr/bin/env bash
# Regenerate flake.nix vendorHash when go.mod/go.sum change.
#
# Only runs if go.mod or go.sum are staged in the index. Replaces the
# current sha256 hash with pkgs.lib.fakeHash, runs `nix build .#fiken`
# to provoke the mismatch error, captures the actual hash from the
# error output, patches flake.nix back in, and re-stages it.
set -euo pipefail

if ! git diff --cached --name-only | grep -qE '^(go\.mod|go\.sum)$'; then
    exit 0
fi

echo "vendor-hash: go.mod/go.sum changed, regenerating flake vendorHash..."

# Replace the current hash with fakeHash.
sed -i 's|vendorHash = "sha256-[^"]*"|vendorHash = pkgs.lib.fakeHash|' flake.nix

# Run nix build and capture the actual hash from the error output.
# The build is *expected* to fail (hash mismatch with fakeHash) — disable
# pipefail locally so `set -e` doesn't trip on it.
set +o pipefail
got=$(nix --extra-experimental-features 'nix-command flakes' build .#fiken 2>&1 | awk '/got:/{print $2; exit}')
set -o pipefail

if [ -z "$got" ]; then
    echo "vendor-hash: could not capture new hash from nix build output" >&2
    exit 1
fi

# Patch the captured hash in.
sed -i "s|vendorHash = pkgs.lib.fakeHash|vendorHash = \"$got\"|" flake.nix

# Re-stage so the commit includes the updated hash.
git add flake.nix

echo "vendor-hash: updated to $got"
