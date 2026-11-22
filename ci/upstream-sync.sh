#!/usr/bin/env bash
#
# Holt sipeed/NanoKVM main und merged es in einen neuen Branch
# sync/upstream-<datum>. Danach: ./ci/gate.sh all und ein Pull Request.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

UPSTREAM_URL="${UPSTREAM_URL:-https://github.com/sipeed/NanoKVM.git}"
BASE="${BASE:-origin/main}"

if ! git remote get-url upstream >/dev/null 2>&1; then
    git remote add upstream "$UPSTREAM_URL"
fi
git fetch upstream main
git fetch origin main

branch="sync/upstream-$(date +%Y-%m-%d)"
git switch -c "$branch" "$BASE"

if git merge --no-edit upstream/main; then
    echo "Merge ohne Konflikt auf $branch. Weiter mit ./ci/gate.sh all"
else
    echo "Konflikte auf $branch:" >&2
    git diff --name-only --diff-filter=U >&2
    exit 1
fi
