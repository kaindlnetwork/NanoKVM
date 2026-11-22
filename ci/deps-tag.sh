#!/usr/bin/env bash
#
# Inhaltsadressierter Tag des CI-Images: gleicher Dockerfile-Inhalt,
# gleicher Tag. Die Pipeline baut nur, wenn es den Tag noch nicht gibt.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
sha256sum ci/Dockerfile | cut -c1-16
