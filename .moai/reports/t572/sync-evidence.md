# t572 sync-evidence — SPEC-OWNERSHIP-SILENCE-001 (sync-phase §E.4 증거 원장)

- 트리: `.claude/worktrees/t572`, branch `WT-ownership-lint-silent`
- 시작 HEAD (sync 진입 시): `91c7079dc` (run body-repair 커밋 — 배차문 예상과 일치)
- sync 커밋: `26878d787` — `docs(SPEC-OWNERSHIP-SILENCE-001): sync-phase — 3-phase close (t572)`
- 이 파일은 sync 커밋 **이후**에 작성됐다(커밋은 자신의 SHA를 담을 수 없음) — 따라서 본 파일은 sync 커밋에 포함되지 않고 untracked로 남는다. D3 백필 창에서 `sync_commit_sha` 실측값과 함께 반출 소관은 리드에게 있다.

---

## Claim

단일 sync 커밋 `26878d787`이 3-phase close 전체를 운반한다: CHANGELOG 진입 1건, progress.md §E.4 (placeholder 포함), spec.md frontmatter `in-progress → completed` 전환 (`status` + `updated` 두 필드만), 그리고 `Authored-By-Agent: manager-docs` 트레일러.

## Evidence

### (1) B12 사전 점검 — CHANGELOG 중복 가드

```
$ grep -c 'SPEC-OWNERSHIP-SILENCE-001' CHANGELOG.md
0
exit=1
```

→ 카운트 0 — HALT 조건 아님, 진입 허용.

### (2) B12 사전 점검 — 인용 경로 실재 확인

```
$ ls .moai/reports/t572/
exclusion-fix-evidence.md  golangci-lint.txt  measurement-baseline.md  plan-audit.md
postlint-strict.txt  run-evidence.md  run-preflight.md  unmeasured-per-spec.txt
$ ls .moai/specs/SPEC-OWNERSHIP-SILENCE-001/
acceptance.md  plan.md  progress.md  spec.md
$ diff -q internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md .claude/rules/moai/development/spec-frontmatter-schema.md && echo 'BYTE-TWIN OK'
BYTE-TWIN OK
```

→ CHANGELOG가 인용하는 4개 경로 (`internal/spec/lint_ownership.go`, `.moai/specs/SPEC-OWNERSHIP-SILENCE-001/spec.md`, rule-doc 쌍둥이, `.moai/reports/t572/unmeasured-per-spec.txt`) 전부 실재. AC 개수 = acceptance.md SSOT 기준 8 (AC-OWN-001..008, 배차 AC-1..8과 1:1).

### (3) 커밋 직전 재판독 (staleness 규칙)

```
$ git status --short
 M .moai/specs/SPEC-OWNERSHIP-SILENCE-001/progress.md
 M .moai/specs/SPEC-OWNERSHIP-SILENCE-001/spec.md
 M CHANGELOG.md
$ git rev-parse --short HEAD
91c7079dc
$ git branch --show-current
WT-ownership-lint-silent
```

→ 내 변경 3파일만, 타인의 파일 없음, HEAD 이탈 없음. 명시적 pathspec 3개로 스테이징 (sweep 금지 준수).

### (4) 착지된 CHANGELOG 진입 (as-landed, `git show HEAD -- CHANGELOG.md` 의 +행 전문)

```
- **[SPEC-OWNERSHIP-SILENCE-001](.moai/specs/SPEC-OWNERSHIP-SILENCE-001/spec.md)** — sync-phase close (3-phase plan→run→sync, card t572, Tier M). A trailer-less ownership transition passed `moai spec lint` silently: the rule found the transition and its expected owner, then returned nil when the commit carried no `Authored-By-Agent:` trailer — so every such transition read as a green (the previous subject-prefix fallback rested on the premise that trailer-less commits are legacy/non-MoAI, which stopped being true when trailer-less became the default state of recent MoAI commits). The silent skip is now a loud Info finding — `OwnershipTransitionUnmeasured` — a statement about measurement state (the transition cannot be attributed), not a violation: loud unmeasured over empty green. **`Authored-By-Agent:` is the single WHO signal**: the commit subject is never consulted (subject-prefix classification survives in source only to pin its regression tests), and the ownership rule docs (`spec-frontmatter-schema.md`, template + local byte-twins) are aligned with all three finding codes (`OwnershipTransitionInvalid` Warning / `OwnershipTransitionUnreachable` Info / `OwnershipTransitionUnmeasured` Info) and the three silent-by-design transitions documented. **Corpus effect, intentional**: the repaired rule immediately reports 199 `OwnershipTransitionUnmeasured` Info findings — 199 SPECs, one finding each (per-SPEC list in `.moai/reports/t572/unmeasured-per-spec.txt`) — all advisory. **`spec lint --strict` exit status is unaffected**: Info findings are never strict-promoted, measured rc=1 before and after (the single error on this tree is a pre-existing plan-phase artifact finding, not this change). This card's own commits carry trailers, so its `draft → in-progress` transition measured green — and this sync commit carries `Authored-By-Agent: manager-docs`, making it the first sync transition the repaired rule actually measures (the expected owner for the close transition is manager-docs). Write surface: `internal/spec/lint_ownership.go` (silent-nil branch → Info emission + two stale comment fixes), `internal/spec/lint_ownership_test.go` (RED-first tests + strict-safety test), rule-doc pair. 8 acceptance criteria, all PASS. This sync commit carries the 3-phase close (`spec.md` frontmatter `in-progress → completed`, `status` + `updated` only), the `progress.md` §E.4 signal, and this entry; `sync_commit_sha` is written as the `pending-backfill-sync` placeholder in this commit — a commit cannot cite its own hash — and the resolved value is backfilled in a following commit.
```

→ `[Unreleased] → ### Added` 최상단(기존 최신 진입 t518 위)에 삽입. 199건 증가의 의도성과 strict 불변성 명기 (§E.3 sync 이관 메모·acceptance §D.2.1 이행).

### (5) `git show --stat HEAD`

```
commit 26878d7870545b5e4178f009dc194634458e4bbd
Author: t <t@t.t>
Date:   Wed Sep 9 00:24:55 2026 +0900

    docs(SPEC-OWNERSHIP-SILENCE-001): sync-phase — 3-phase close (t572)
    [...body...]
 .moai/specs/SPEC-OWNERSHIP-SILENCE-001/progress.md | 11 +++++++++++
 .moai/specs/SPEC-OWNERSHIP-SILENCE-001/spec.md     |  4 ++--
 CHANGELOG.md                                       |  1 +
 3 files changed, 14 insertions(+), 2 deletions(-)
```

### (6) 트레일러 검증 (핵심 판정)

```
$ git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)'
manager-docs

trailer_rc=0
```

→ `manager-docs` 출력 — 트레일러 블록 생존 (커밋 본문에서 `🗿 MoAI` 마커와 트레일러를 별도 paragraph로 분리한 형식이 git trailer 파싱을 통과). 수리된 OwnershipTransitionRule이 실제로 측정하는 첫 sync 전환 — 기대 소유자 manager-docs와 일치 (REQ-OWN-010 도그푸드 관측점).

### (7) `git log -1 --format='%h %s'`

```
26878d787 docs(SPEC-OWNERSHIP-SILENCE-001): sync-phase — 3-phase close (t572)
```

### (8) spec.md frontmatter 전환 diff (`git show HEAD -- .moai/specs/SPEC-OWNERSHIP-SILENCE-001/spec.md`)

```diff
-status: in-progress
+status: completed
 created: 2026-09-08
-updated: 2026-09-08
+updated: 2026-09-09
```

→ `status` + `updated` 정확히 두 필드. 본문 무접촉 (manager-docs 소유 경계 준수 — 스키마 행렬의 단일 sync 커밋 이중 홉 `in-progress → implemented → completed`).

## Baseline-attribution

모든 측정은 본 트리(`.claude/worktrees/t572`, branch `WT-ownership-lint-silent`)에서, sync 커밋 전후 직접 실행으로 관측. 인용된 모든 커맨드·출력은 이 원장에 verbatim 기록. run-phase 근거는 `.moai/reports/t572/run-evidence.md` (AC-OWN-001..008 원장) 인용.

## Gaps

- `sync_commit_sha`는 placeholder `pending-backfill-sync` 상태 — 실측 SHA `26878d787` 백필은 후속 커밋 소관 (D3 자기참조 규약; 본 파일과 마찬가지로 커밋이 자신의 SHA를 담을 수 없는 물리적 한계)
- `go test` 미실행 (sync-phase) — 이번 커밋이 건드린 것은 markdown 3파일뿐이고, 피시험 Go 코드(run M1-M3 착지분)의 판정은 run-phase §E.2 원장 + CI 몫. 선택적 sanity(`go test ./internal/spec/ -run TestOwnershipTransition -count=1`)도 실행하지 않았다 — markdown-only 커밋에 Go 테스트 근거는 불요 판단.
- `moai spec lint` 재실행 안 함 — §E.4 기록의 소유 전환 판정(기대 소유자 manager-docs 일치)은 트레일러 파싱 관측(위 (6))으로 성립하며, lint 재실행은 백필 이후 트리에서 의미가 있다(placeholder 상태에서는 slot-format 판정이 중간 상태를 가리킨다).

## Residual-risk

- CHANGELOG 진입이 기술한 "199건" "rc=1 before/after" 수치는 run-evidence.md의 run-phase 실측을 인용한 것 — sync-phase에서 재측정하지 않았다(같은 트리, 코드 불변이므로 인용 유효).
- 본 sync 커밋의 소유 전환 판정("기대 소유자 manager-docs와 일치")은 트레일러 존재(위 (6))와 스키마 행렬 매핑에 근거하며, 룰의 실제 런타임 판정 출력은 관측하지 않았다 — 백필 후 코퍼스 lint에서 본 SPEC의 OwnershipTransitionUnmeasured = 0건이 확인되면 닫힌다.
