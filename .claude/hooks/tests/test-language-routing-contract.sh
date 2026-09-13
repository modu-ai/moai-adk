#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
HOOK="$ROOT/.claude/hooks/moai/sync-phase-quality-gate.sh"
RULE="$ROOT/.claude/rules/moai/workflow/language-routing-contract.md"
DOC="$ROOT/.claude/skills/moai/workflows/sync/quality-gates-quality.md"

TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "$TMP_ROOT"' EXIT

set -- ""
source "$HOOK"

mkdir -p "$TMP_ROOT/kotlin"
printf 'plugins { kotlin("jvm") version "2.0.0" }\n' > "$TMP_ROOT/kotlin/build.gradle.kts"
test "$(detect_language "$TMP_ROOT/kotlin")" = kotlin

mkdir -p "$TMP_ROOT/mono"
: > "$TMP_ROOT/mono/go.mod"
: > "$TMP_ROOT/mono/package.json"
: > "$TMP_ROOT/mono/service.go"
: > "$TMP_ROOT/mono/app.ts"
candidates="$(detect_languages "$TMP_ROOT/mono")"
grep -Fxq go <<<"$candidates"
grep -Fxq node <<<"$candidates"

# Card t604: a version-catalog build script names no Kotlin token, so detection
# must rest on the Kotlin source itself — which lives at src/main/kotlin/, below
# the shared suffix probe's depth bound. Before the fix this resolved to java and
# the gate skipped the project without writing a log line.
mkdir -p "$TMP_ROOT/kotlin-catalog/src/main/kotlin/com/example"
printf 'plugins {\n    alias(libs.plugins.jvm)\n}\n' > "$TMP_ROOT/kotlin-catalog/build.gradle.kts"
printf 'fun main() { println("hi") }\n' > "$TMP_ROOT/kotlin-catalog/src/main/kotlin/com/example/Main.kt"
got="$(detect_language "$TMP_ROOT/kotlin-catalog")"
test "$got" = kotlin || {
    printf 'FAIL: version-catalog Kotlin project detected as %s, want kotlin\n' "$got" >&2
    exit 1
}

# The build-script leg, on its own: a catalog alias that DOES name Kotlin, in a
# module that carries no .kt source yet. Only the grep can decide this one, so it
# is what guards the alias pattern — the fixture above cannot, because its alias
# names no Kotlin token at all.
mkdir -p "$TMP_ROOT/kotlin-alias"
printf 'plugins {\n    alias(libs.plugins.kotlin.jvm)\n}\n' > "$TMP_ROOT/kotlin-alias/build.gradle.kts"
got="$(detect_language "$TMP_ROOT/kotlin-alias")"
test "$got" = kotlin || {
    printf 'FAIL: Kotlin catalog alias detected as %s, want kotlin\n' "$got" >&2
    exit 1
}

# Control for the same fix: a Java Gradle project also carries build.gradle.kts
# but no .kt source, and must keep resolving to java. Without this the deeper
# Kotlin probe could be widened (e.g. to *.kts) and pass the case above while
# silently misrouting every Java Gradle project.
mkdir -p "$TMP_ROOT/java-gradle/src/main/java"
printf 'plugins { id("java") }\n' > "$TMP_ROOT/java-gradle/build.gradle.kts"
printf 'class Main {}\n' > "$TMP_ROOT/java-gradle/src/main/java/Main.java"
got="$(detect_language "$TMP_ROOT/java-gradle")"
test "$got" = java || {
    printf 'FAIL: Java Gradle project detected as %s, want java\n' "$got" >&2
    exit 1
}

# Card t604 (derived, lead-approved): the Kotlin code-delta pattern was
# '\.kt|\.kts$' — the first alternative carried no end anchor, so any path
# merely CONTAINING ".kt" counted as a Kotlin code change. Both directions are
# asserted, so a pattern that matches nothing cannot pass as "no false positive".
kotlin_delta="$(code_delta_pattern kotlin)"
printf 'src/main/kotlin/Main.kt\nbuild.gradle.kts\n' | grep -Eq "$kotlin_delta" || {
    printf 'FAIL: kotlin delta pattern %s matches no Kotlin source\n' "$kotlin_delta" >&2
    exit 1
}
if printf 'docs/README.kt.md\n' | grep -Eq "$kotlin_delta"; then
    printf 'FAIL: kotlin delta pattern %s matches a non-Kotlin path\n' "$kotlin_delta" >&2
    exit 1
fi

# Card t664: the shared has_suffix probe was capped at -maxdepth 3, so every
# language whose only evidence is a source suffix went undetected under its own
# conventional layout (src/main/java/<pkg>/, lib/<app>/<area>/, Sources/<Module>/…).
# Detection then returned nothing and the gate exited silently. Kotlin was the
# sole exception, because t604 had already given it a prune-based probe.
# One case per suffix-detectable language, each with NO root manifest so the
# suffix probe is the only leg that can decide.
while IFS='|' read -r lang rel body; do
    [ -n "$lang" ] || continue
    root="$TMP_ROOT/deep-$lang"
    mkdir -p "$root/$(dirname "$rel")"
    printf '%s\n' "$body" > "$root/$rel"
    got="$(detect_languages "$root" | paste -sd, -)"
    case ",$got," in
        *",$lang,"*) ;;
        *)
            printf 'FAIL: %s at %s detected as [%s], want %s\n' \
                "$lang" "$rel" "$got" "$lang" >&2
            exit 1
            ;;
    esac
done <<'FIXTURES'
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
FIXTURES

# The depth cap was replaced by a prune set, not removed — dropping the prune
# would turn every vendored dependency into a detection hit. One case per
# pruned class, asserted from a tree whose ONLY sources live inside them.
for skipdir in node_modules vendor .git dist build target __pycache__ .venv; do
    root="$TMP_ROOT/pruned-${skipdir#.}"
    mkdir -p "$root/$skipdir/pkg/deep"
    printf 'package main\n' > "$root/$skipdir/pkg/deep/main.go"
    printf 'class S {}\n' > "$root/$skipdir/pkg/deep/S.java"
    got="$(detect_languages "$root" | paste -sd, -)"
    test -z "$got" || {
        printf 'FAIL: sources under %s/ detected as [%s], want no candidates\n' \
            "$skipdir" "$got" >&2
        exit 1
    }
done

# The .csproj probe carried a second copy of the same cap; a solution laying
# projects out under src/<Module>/ sits below it.
mkdir -p "$TMP_ROOT/csproj-deep/src/Modules/Billing"
printf '<Project Sdk="Microsoft.NET.Sdk" />\n' \
    > "$TMP_ROOT/csproj-deep/src/Modules/Billing/Billing.csproj"
got="$(detect_language "$TMP_ROOT/csproj-deep")"
test "$got" = csharp || {
    printf 'FAIL: nested .csproj detected as %s, want csharp\n' "$got" >&2
    exit 1
}

grep -Fq 'all candidates' "$RULE"
grep -Fq 'multiple candidates' "$RULE"
grep -Fq 'Do not use first-match wins' "$DOC"

printf 'PASS: language routing distinguishes Kotlin, preserves monorepo candidates, and reaches conventional deep layouts\n'
