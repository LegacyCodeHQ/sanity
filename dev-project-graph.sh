#!/usr/bin/env bash
set -euo pipefail

./sanity graph -u --include "$(find . -name '*.go' | tr '\n' ',' | sed 's/,$//')"
