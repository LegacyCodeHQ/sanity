#!/usr/bin/env bash
set -euo pipefail

./sanity graph -u $(find . -name "*.go")
