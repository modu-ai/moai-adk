# t371 — StatusGitConsistency 18건 전수 분류

Baseline: develop `b9149857c` (working tree 동일, `main` = `48239c7dc`).
측정 명령: `go run ./cmd/moai spec lint --strict` — 원본은 `deep-mainref-b9149857c.txt`.
분류 근거: `internal/spec/drift.go` getGitImpliedStatus + `internal/spec/transitions.go` ClassifyPRTitle 을
`statusgit-18-walker-input.txt` 의 커밋 목록에 적용.

## 요약

| 부류 | 건수 | 성격 |
|---|---|---|
| A. 진짜 frontmatter 드리프트 | 2 | 작업은 착지했는데 frontmatter 가 뒤처짐 |
| B. close 상태 불일치 | 2 | 4-phase close 커밋이 있는데 frontmatter 는 implemented |
| C. 워커 휴리스틱 산물 | 14 | frontmatter 가 맞고, 워커가 최신 cosmetic 커밋을 물었거나 규약을 못 읽음 |

## A. 진짜 드리프트 (2)

| SPEC | frontmatter | git-implied | 워커가 문 커밋 |
|---|---|---|---|
| SPEC-KANBAN-TODO-CLI-001 | in-progress | implemented | `a84e48961 feat(SPEC-KANBAN-TODO-CLI-001): moai todo CLI with lock-guarded backlog store (#1529)` |
| SPEC-UPDATE-DOC-DRIFT-001 | draft | implemented | `ddfe2253f feat(SPEC-UPDATE-DOC-DRIFT-001): v0.3.0 staleness rewrite (#1515)` |

## B. close 상태 불일치 (2)

close-infix(`4-phase close`)가 completed 를 함의하는데 frontmatter 는 implemented.

| SPEC | frontmatter | git-implied | 워커가 문 커밋 |
|---|---|---|---|
| SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 | implemented | completed | `7cd25e386 chore(...): sync-phase 4-phase close` |
| SPEC-V3R6-SKILL-GEARS-ALIGN-001 | implemented | completed | `7b8f939b7 chore(...): sync-phase 4-phase close + CHANGELOG v0.2.0` |

## C. 워커 휴리스틱 산물 (14)

전부 frontmatter `implemented`, git-implied `in-progress`. 세 가지 모양으로 갈린다.

### C-1. sync 커밋 주제가 `sync-phase` 리터럴을 안 담음 (5)

`syncPhaseDocsPattern` = `^docs\(spec-[a-z0-9-]+-[0-9]+\):.*sync-phase`. 실제 주제는 `docs(SPEC-X): sync — ...`
라서 generic `docs` 규칙(`in-progress`)으로 떨어진다.

SPEC-V3R6-CHANGELOG-CLEANUP-001 · SPEC-V3R6-CLI-AUDIT-001 · SPEC-V3R6-LEGACY-CLEANUP-001 ·
SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 · SPEC-GOAL-STOPFAILURE-CLEAR-001

### C-2. sync 이후 cosmetic docs/chore 가 창의 최신을 차지 (6)

워커는 첫 분류가능 커밋에서 멈추므로, sync 뒤에 붙은 progress 스탬프·word count·stale link 정리가
`in-progress` 를 만든다. SPEC-V3R6-SKILL-COMPRESS-001 은 `status implemented` 를 실제로 적용한 그 커밋이
`chore` 접두라 in-progress 로 읽히는 자기모순.

SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 · SPEC-V3R6-RULES-COMPRESS-001 · SPEC-V3R6-SKILL-COMPRESS-001 ·
SPEC-V3R6-LEGACY-CLEANUP-002 · SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 · SPEC-V3R6-DOCS-CMD-CATALOG-001

### C-3. SPEC-ID 를 주제에 담은 lifecycle 커밋이 아예 없음 (3)

워커가 무는 것은 그 SPEC 을 스쳐 지나간 제3자 커밋이다.

- SPEC-FACTORY-WORKER-FANOUT-001 — `docs(specs): add Out of Scope to ...` 하나뿐
- SPEC-V3R6-GEARS-MIGRATION-001 — `plan(SPEC-...)` 는 접두 표(`plan(spec)`)와 안 맞아 unknown, 그 아래 `docs(research): ...` 가 채택됨
- SPEC-V3R6-DOCS-USER-DRIFT-001 — `docs(SPEC-...): 좌초 run-phase 구현 회수`

## 판정에 필요한 사실

- 두 규칙 모두 `--strict` 에서 **error 로 승격되지 않는다**. StatusGitConsistency 는 emission 에서
  `Advisory: true`(internal/spec/lint.go:1316), OwnershipTransitionInvalid 는 Advisory 를 안 달지만
  현재 유일 대상 SPEC-LSPMCP-001 이 terminal status 라 `applyEraDemotion`(lint.go:293-296)이 advisory 로 낮춘다.
- 실측: 19건이 전부 보이는 상태에서 `spec lint --strict` **rc=0**(`deep-mainref-b9149857c.txt`).
  즉 눈을 뜨게 해도 지금 코퍼스에서는 SPEC Lint 가 붉어지지 않는다.
- 다만 앞으로 grandfather 아닌 SPEC 이 OwnershipTransitionInvalid 를 맞으면 승격돼 붉어진다 — 잠재 위험은 남는다.
