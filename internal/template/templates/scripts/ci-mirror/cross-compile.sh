#!/bin/sh
# scripts/ci-mirror/cross-compile.sh — Go cross-compilation matrix
# Verifies the project builds for all 6 required OS/ARCH combinations
# (REQ-CIAUT-003: linux/darwin/windows × amd64/arm64).
# Exit codes: 0 success, 3 build failure.
set -eu

log() { printf '[cross-compile] %s\n' "$1" >&2; }

log "linux/amd64"
GOOS=linux GOARCH=amd64 go build -o /dev/null ./... || exit 3

log "linux/arm64"
GOOS=linux GOARCH=arm64 go build -o /dev/null ./... || exit 3

log "darwin/amd64"
GOOS=darwin GOARCH=amd64 go build -o /dev/null ./... || exit 3

log "darwin/arm64"
GOOS=darwin GOARCH=arm64 go build -o /dev/null ./... || exit 3

log "windows/amd64"
GOOS=windows GOARCH=amd64 go build -o /dev/null ./... || exit 3

log "windows/arm64"
GOOS=windows GOARCH=arm64 go build -o /dev/null ./... || exit 3

log "all 6 targets built successfully"
