#!/bin/sh
# scripts/ci-mirror/lib/_common.sh — POSIX sh helpers for ci-mirror
# Do NOT use bash-isms: no [[ ]], no ${var//}, no local, no set -o pipefail.

# log_step prints a timestamped step message to stderr.
log_step() { printf '[ci-mirror] %s\n' "$1" >&2; }

# abort prints a fatal error message and exits with the given code (default 99).
abort() { printf '[ci-mirror][FATAL] %s\n' "$1" >&2; exit "${2:-99}"; }

# posix_now returns the current Unix timestamp (seconds since epoch).
posix_now() { date +%s; }

# detect_languages inspects the repository root for language marker files
# and prints each detected language on its own line.
# REPO_ROOT must be set before calling this function.
detect_languages() {
    REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
    [ -f "$REPO_ROOT/go.mod" ] && printf 'go\n'
    [ -f "$REPO_ROOT/package.json" ] && printf 'node\n'
    { [ -f "$REPO_ROOT/pyproject.toml" ] || [ -f "$REPO_ROOT/setup.py" ]; } && printf 'python\n'
    [ -f "$REPO_ROOT/Cargo.toml" ] && printf 'rust\n'
    [ -f "$REPO_ROOT/build.gradle.kts" ] && printf 'kotlin\n'
    { [ -f "$REPO_ROOT/pom.xml" ] || [ -f "$REPO_ROOT/build.gradle" ]; } && printf 'java\n'
    ls "$REPO_ROOT"/*.csproj "$REPO_ROOT"/*.sln 2>/dev/null | head -1 | grep -q . && printf 'csharp\n'
    [ -f "$REPO_ROOT/Gemfile" ] && printf 'ruby\n'
    [ -f "$REPO_ROOT/composer.json" ] && printf 'php\n'
    [ -f "$REPO_ROOT/mix.exs" ] && printf 'elixir\n'
    [ -f "$REPO_ROOT/CMakeLists.txt" ] && printf 'cpp\n'
    [ -f "$REPO_ROOT/build.sbt" ] && printf 'scala\n'
    [ -f "$REPO_ROOT/DESCRIPTION" ] && printf 'r\n'
    [ -f "$REPO_ROOT/pubspec.yaml" ] && printf 'flutter\n'
    [ -f "$REPO_ROOT/Package.swift" ] && printf 'swift\n'
    return 0
}
