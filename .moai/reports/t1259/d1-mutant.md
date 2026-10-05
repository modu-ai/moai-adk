# t1259 — plan-audit iter4 D1: moot here, carried to t1290 (with a demonstrated predicate)

Card **t1259** · SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 v0.3.0 · worktree `.claude/worktrees/t1259`
(branch `WT-local-instructions`), measured on HEAD `4441cf1a6`, 2026-09-28.

## 1. Why D1 is moot in this SPEC

iter4 D1 (`.moai/reports/t1259/plan-audit-iter4.md` §2) named exactly two criteria —
`AC-IFU-007` and `AC-IFU-024` — as reading the gitignored working copy of `AGENTS.local.md`.
Both verified only `REQ-IFU-021` / `REQ-IFU-022` (the v0.2.5 §D.2 traceability table maps each to
exactly one of them, one-to-one). The lead's tier-3 decision moved M3 — those two requirements and
those two criteria — to card **t1290**. After the v0.3.0 edit neither criterion is live in
`acceptance.md`: the AC counter reports `live=8 excluded=5 ambiguous=0`, the remaining references
carry the `[RETIRED]` marker, and the §D.2 traceability diff is empty over the reduced set
(verbatim outputs in `progress.md` §E.1, v0.3.0 section). No criterion in this SPEC reads
`AGENTS.local.md` or `CLAUDE.local.md` any more, so there is nothing left here for the mutant to
satisfy.

**D1 is therefore carried to t1290, not closed.** t1290 inherits both criteria at `4441cf1a6`
verbatim; they must not be adopted as written. The predicate below is the replacement, with its
mutant kills demonstrated so that t1290 adopts a criterion whose red has been observed
(`verification-completeness.md` §1.1, §2 mutant probe).

## 2. The ignore facts it rests on (this run, this tree)

```
$ git check-ignore -v AGENTS.local.md CLAUDE.local.md
.gitignore:275:/AGENTS.local.md	AGENTS.local.md
$ git ls-files AGENTS.local.md CLAUDE.local.md
CLAUDE.local.md
$ git check-ignore --no-index -v CLAUDE.local.md
.gitignore:276:/CLAUDE.local.md	CLAUDE.local.md
```

Both names are ignored; `CLAUDE.local.md` is tracked despite its rule, `AGENTS.local.md` would not
be unless force-added.

## 3. The replacement predicate (committed tree, never the working copy)

For the lane's merge commit `<ref>`, all three must hold:

1. `git cat-file -e <ref>:AGENTS.local.md` succeeds — the migrated file is **committed**;
2. `git show <ref>:AGENTS.local.md | wc -m` is at most 39,999;
3. `git cat-file -e <ref>:CLAUDE.local.md` fails — the old file is gone from the committed tree.

The v0.2.5 predicate (`wc -m < AGENTS.local.md`, working copy) is run beside it as the contrast.

## 4. Demonstration — temp git repo, three mutants and one control

Fixture mirrors this repository's geometry: both names ignored, `CLAUDE.local.md` force-tracked at
45,000 characters; "the verb" is simulated as move-to-backup plus a 39,000-character
`AGENTS.local.md` in the working tree. Script (run with `sh`, repo created under `mktemp -d` and
removed at the end):

```sh
#!/bin/sh
# D1 mutant/control demonstration for the committed-tree predicate (card t1259 -> t1290).
set -u
R=$(mktemp -d)/repo; mkdir -p "$R"; cd "$R" || exit 9
git init -q -b develop . && git config user.email t@t && git config user.name t
printf '/AGENTS.local.md\n/CLAUDE.local.md\n' > .gitignore
printf 'x%.0s' $(seq 1 45000) > CLAUDE.local.md
git add .gitignore && git add -f CLAUDE.local.md && git commit -qm base
BASE=$(git rev-parse HEAD)

# NEW predicate: reads the committed tree at <ref>, never the working copy.
committed_pred() {
  ref=$1
  git cat-file -e "$ref:AGENTS.local.md" 2>/dev/null || { echo "  FAIL: $ref carries no AGENTS.local.md"; return 1; }
  n=$(git show "$ref:AGENTS.local.md" | wc -m | tr -d ' ')
  [ "$n" -le 39999 ] || { echo "  FAIL: $ref:AGENTS.local.md is $n chars (> 39999)"; return 1; }
  if git cat-file -e "$ref:CLAUDE.local.md" 2>/dev/null; then echo "  FAIL: $ref still carries CLAUDE.local.md"; return 1; fi
  echo "  PASS: $ref:AGENTS.local.md = $n chars, CLAUDE.local.md absent from $ref"; return 0
}
# OLD predicate (acceptance.md v0.2.5 AC-IFU-007): working copy.
old_pred() { n=$(wc -m < AGENTS.local.md | tr -d ' '); [ "$n" -le 39999 ] && echo "  PASS (old, working copy): $n chars" || echo "  FAIL (old): $n"; }

# simulate the verb + trim, in the working tree
migrate() { mkdir -p .moai/backup && mv CLAUDE.local.md .moai/backup/ && printf 'y%.0s' $(seq 1 39000) > AGENTS.local.md; }

echo "== MUTANT A: verb run, commit WITHOUT git add -f (git add -A sweep) =="
migrate; git add -A; git commit -qm mutantA
echo "git ls-files AGENTS.local.md CLAUDE.local.md -> [$(git ls-files AGENTS.local.md CLAUDE.local.md)]"
old_pred
committed_pred HEAD; echo "  exit=$?"

echo "== MUTANT B: verb run, file left only in working tree, deletion not committed either =="
git reset -q --hard "$BASE"; rm -rf .moai AGENTS.local.md; migrate
old_pred
committed_pred HEAD; echo "  exit=$?"

echo "== MUTANT C: AGENTS.local.md force-added, but CLAUDE.local.md deletion left uncommitted =="
git reset -q --hard "$BASE"; rm -rf .moai AGENTS.local.md; migrate
git add -f AGENTS.local.md; git commit -qm mutantC
echo "git ls-files AGENTS.local.md CLAUDE.local.md -> [$(git ls-files AGENTS.local.md CLAUDE.local.md | tr '\n' ' ')]"
old_pred
committed_pred HEAD; echo "  exit=$?"

echo "== CONTROL: verb run, git add -f AGENTS.local.md, deletion committed =="
git reset -q --hard "$BASE"; rm -rf .moai AGENTS.local.md; migrate
git add -f AGENTS.local.md; git rm -q --cached CLAUDE.local.md; git commit -qm control
echo "git ls-files AGENTS.local.md CLAUDE.local.md -> [$(git ls-files AGENTS.local.md CLAUDE.local.md)]"
old_pred
committed_pred HEAD; echo "  exit=$?"
rm -rf "$(dirname "$R")"
```

Verbatim output (`sh d1.sh` → exit 0):

```
== MUTANT A: verb run, commit WITHOUT git add -f (git add -A sweep) ==
git ls-files AGENTS.local.md CLAUDE.local.md -> []
  PASS (old, working copy): 39000 chars
  FAIL: HEAD carries no AGENTS.local.md
  exit=1
== MUTANT B: verb run, file left only in working tree, deletion not committed either ==
  PASS (old, working copy): 39000 chars
  FAIL: HEAD carries no AGENTS.local.md
  exit=1
== MUTANT C: AGENTS.local.md force-added, but CLAUDE.local.md deletion left uncommitted ==
git ls-files AGENTS.local.md CLAUDE.local.md -> [AGENTS.local.md CLAUDE.local.md ]
  PASS (old, working copy): 39000 chars
  FAIL: HEAD still carries CLAUDE.local.md
  exit=1
== CONTROL: verb run, git add -f AGENTS.local.md, deletion committed ==
git ls-files AGENTS.local.md CLAUDE.local.md -> [AGENTS.local.md]
  PASS (old, working copy): 39000 chars
  PASS: HEAD:AGENTS.local.md = 39000 chars, CLAUDE.local.md absent from HEAD
  exit=0
```

| Case | Old predicate (working copy) | New predicate (committed tree) |
|---|---|---|
| A — verb run, sweep-staged, no `git add -f` | PASS | **FAIL** (exit 1): no `AGENTS.local.md` in HEAD |
| B — file left only in the working tree | PASS | **FAIL** (exit 1): no `AGENTS.local.md` in HEAD |
| C — force-added, but `CLAUDE.local.md` deletion uncommitted | PASS | **FAIL** (exit 1): HEAD still carries `CLAUDE.local.md` |
| Control — `git add -f` + deletion committed | PASS | **PASS** (exit 0) |

The old predicate passes all four, so it cannot tell the mutants from the correct state; the new one
kills each mutant on a different clause and passes the control.

## 5. Gaps

- The fixture is a synthetic repo, not a real migration of this repository's file; the verb itself
  does not exist yet (M1 builds it). What was shown is the predicate's discrimination, not the
  migration.
- iter4 D2 (worktree reception of `AGENTS.local.md`) is not addressed here; it is t1290's
  prerequisite, and nothing above measures it.
