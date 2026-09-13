#!/usr/bin/env bash
# t664 measurement harness — idiomatic deep source layouts, no root manifest.
#
# Each fixture places its ONLY language evidence at the depth that language's
# conventional layout actually uses, and carries no root manifest, so the shared
# has_suffix probe is the only leg that can decide. Emits one TSV row per
# language: <lang> <want> <got-or-NONE> <path-depth>.
#
# Usage: measure-depth.sh <path-to-sync-phase-quality-gate.sh>
set -uo pipefail

HOOK="${1:?usage: measure-depth.sh <path-to-sync-phase-quality-gate.sh>}"
FX="$(mktemp -d)"
trap 'rm -rf "$FX"' EXIT

set -- ""
# shellcheck disable=SC1090
source "$HOOK"

# lang | idiomatic source path | file body
FIXTURES='
go|src/cmd/app/internal/main.go|package main
python|src/pkg/core/util/helper.py|def f(): pass
node|src/components/ui/forms/Input.tsx|export const I = 1
rust|src/modules/core/util/lib.rs|pub fn f() {}
java|src/main/java/com/example/Main.java|class Main {}
kotlin|src/main/kotlin/com/example/Main.kt|fun main() {}
csharp|src/App/Services/Impl/Service.cs|class S {}
ruby|lib/app/models/core/user.rb|class User; end
php|src/App/Http/Controllers/Home.php|<?php class Home {}
elixir|lib/app/core/domain/user.ex|defmodule U do end
cpp|src/core/engine/impl/engine.cpp|int main(){return 0;}
scala|src/main/scala/com/example/Main.scala|object Main
r|R/analysis/core/model.R|f <- function() 1
flutter|lib/src/features/home/view.dart|void main() {}
swift|Sources/App/Features/Home/View.swift|struct V {}
'

printf 'lang\twant\tgot\tdepth\n'
printf '%s\n' "$FIXTURES" | while IFS='|' read -r lang rel body; do
    [ -n "$lang" ] || continue
    root="$FX/$lang"
    mkdir -p "$root/$(dirname "$rel")"
    printf '%s\n' "$body" > "$root/$rel"
    depth=$(printf '%s' "$rel" | awk -F/ '{print NF}')
    got="$(detect_languages "$root" | paste -sd, -)"
    [ -n "$got" ] || got="NONE"
    printf '%s\t%s\t%s\t%s\n' "$lang" "$lang" "$got" "$depth"
done
