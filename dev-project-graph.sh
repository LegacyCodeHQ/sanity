#!/usr/bin/env bash
set -euo pipefail

./sanity graph -u --input "$(find . -name '*.go' | tr '\n' ',' | sed 's/,$//')"
