# t503 — 판정서 (verdict)

카드: `/moai` 커맨드 16본을 codex 스킬로 발행하는 emitter · Class C · Tier M
SPEC: SPEC-CODEX-COMMAND-SKILLS-001 (completed)
레인: lane-8 · 워크트리 `.claude/worktrees/t503` · 브랜치 `WT-codex-command-skills` · base origin/develop `ace1c5440`

---

## Claim

1. 카드 t503 완수 — `.claude/commands/moai/` 16본(15 `.md.tmpl` + `todo.md`)을 codex 스킬 모양(`SKILL.md` + `name`/`description` 프론트매터)으로 발행하는 emitter가 구현돼 있고, 발행물 16본이 `internal/template/templates/.agents/skills/moai-<command>/SKILL.md`로 커밋돼 있다.
2. Claude 쪽 슬래시 표면은 무변경(발행이지 변환 아님 — 카드 [HARD] 경계)이며, 커맨드 본문은 verbatim 발행(t497 중립성 수리 미실행 — 경계 준수).
3. SPEC-CODEX-COMMAND-SKILLS-001의 3-phase close가 완료됐다 (status: completed, sync 커밋 `9da07ac65`, 그 이후 Go 코드 변경 0).
4. plan-audit iter-1 FAIL 0.75(결함 4건) → 수리 → iter-2 PASS 1.0(결함 0건).

## Evidence (모두 이 턴, 이 트리에서 레인이 직접 재측정)

| 측정 | 명령 | 관측 출력 |
|---|---|---|
| 발행물 수 | `find internal/template/templates/.agents/skills -name SKILL.md \| wc -l` | `16` |
| affected 패키지 테스트 | `go test ./internal/template/...` | `ok internal/template 25.038s` · `ok agentemit 0.821s` · `ok commandemit 0.399s` (전부 ok) |
| 신선도 게이트 | `make commands-emit-check` | rc=0 · 게이트 후 `git status --porcelain …/.agents/` 0행(쓰기 없음 확인) |
| vet / 포맷 | `go vet ./internal/template/...` · `gofmt -l internal/template/` | 양쪽 rc=0, 출력 0행 |
| 컴파일 판별 | `go build ./internal/template/...` | rc=0 (아래 LSP 거짓 진단 판별) |
| SPEC 상태 | `grep '^status:' …/spec.md` | `status: completed` |
| CHANGELOG | `grep -c 'SPEC-CODEX-COMMAND-SKILLS-001' CHANGELOG.md` | `1` |
| sync 마지막 쓰기 | `git diff --name-only 9da07ac65..HEAD -- '*.go' \| wc -l` | `0` |
| 미푸시 커밋 | `git rev-list --count origin/develop..HEAD` | `8` (직접 측정 — 산술 추정 아님) |
| 트리 청결 | `git status --short \| wc -l` | `0` |

**RED→GREEN (E8, run-phase에서 관측 — verbatim 출력은 `.moai/reports/t503/run-evidence.md`)**: AC-005 충돌 거부(가드 제거 시 `TestCollisionRefused` FAIL 2건 → 가드 복원 후 PASS) · AC-010 드리프트(발행물 수동 변조 시 `commands-emit-check` 비영(非zero) + 진단 → `make commands-emit` 후 rc=0). AC-002·AC-008은 감사 인계에 따라 regression-guard로 E8 의무 없음.

**LSP 거짓 진단 판별** (sync 중 발생): deployer.go:245·253 `recordProtectedSkip undefined` 컴파일급 진단 → `go build ./internal/template/...` rc=0 + 정의 실재(`published_skills.go:58`)로 스테일 진단 확정. 워크트리 LSP 진단 거짓 교훈 재확인. info/warning 수준 진단(unused func, SplitSeq, WriteString)은 `golangci-lint run ./internal/template/...` 0 issues(run-phase 측정)와 상충하지 않는다.

**경계 준수 실측**: `templates/.claude/commands/moai/**` 무접촡(PRESERVE) · 본문 16본 verbatim (`TestBodiesByteEqual`, diff exit 0) · `Use Skill("moai")` 경계 플래그 16/16 기록(수리 아님) · 충돌 0/16(run 전 실측, plan-auditor 독립 재측정 일치).

## Baseline-attribution

- 트리: `.claude/worktrees/t503`, 재측정 시점 HEAD `223b30f0d` (verdict 커밋 **전** tip — t470 교훈 준수), base origin/develop `ace1c5440`.
- 위 표의 모든 수치는 이 레인이 이 턴에서 실행한 명령의 출력이다. manager-develop·manager-docs 보고의 수치는 인용하지 않고 별도 재측정으로 교차 확인했다.
- 이전 세계 근거: t494 판정서 §2·§4 (`.claude/worktrees/t494/.moai/reports/t494/codex-doc-survey.md`, HEAD ace1c5440) — 공식 문서 인용(deprecated 프롬프트 → 스킬, `$` sigil, `.agents/skills` 스캔 경로, SKILL.md 규약).

## Gaps — 관측하지 않은 것

1. 전체 테스트 스위트 미실행 — 로컬 전수 금지 규율(§4·CLAUDE.local.md §4). 전 패키지 판정은 develop push 이후 CI 몫.
2. codex CLI 실측(발행 스킬의 실제 로드 실증, `codex debug prompt-input`류) 미실행 — 발행물의 로드 가능성은 기존 경로 관례 근거(SPEC-CODEX-SKILL-LOADER-001 분기 A: `<repo>/.agents/skills/<name>/SKILL.md` 실제 로드 확인)에 기대며, t497 실측 축(S1 계열)으로 위임.
3. t494 판정서의 공식 문서 인용은 이 트리에서 재검증하지 않았다 — 워크트리 t494의 판정서(HEAD ace1c5440, 2026-09-07 조회)를 읽기 전용 참조로 사용.
4. docs-site / user-guide 사용자 문서화는 지연 처리로 기록됨 (progress.md §E.4 `user_facing_docs_note`).

## Residual-risk

- codex 공식 문서에 버전 스탬프가 없어(t494 유지) `.agents/skills` 스캔 규약이 0.153.4 이후 변하면 D1 레이아웃 재판정이 필요하다.
- 발행 본문이 verbatim이라 codex에서 `Use Skill("moai")` 호출은 실패한다 — 이것은 카드 경계상 의도된 상태이며 t497(중립성)의 본건이다. 발행물 설명의 탐색 품질도 t497의 재작성 결정에 종속된다.
- `moai update`의 `.agents/` 관리 뿌리 소속은 plan-auditor 실측 시점(`ManagedCleanTargets`에 `.agents` 없음, deploy.go:56-86) 기준 — 이후 deploy.go 변화 시 재측정 필요.
- `[[skills.config]]` 미사용이므로 S1(path 값 모양 실측)과 무관하나, t494 판정서 §7이 C1의 선행으로 S1을 꼽은 것과 이 카드의 S1-비종속 배차 사이에는 운영자 판단이 개입돼 있다 — 본 카드는 경로 관례 축만 사용했다는 점에서 정합.

## 커밋 궤적 (미푸시 8)

| 커밋 | 내용 |
|---|---|
| `a205c181d` | plan: SPEC 초안 |
| `74f3c506e` | plan 수리 round 1 + plan-audit iter-2 PASS |
| `e7d2a1658` | M1 — emitter core + golden 16본 |
| `6ae337e60` | M2 — freshness (`make commands-emit`/`-check`) |
| `082c2c04f` | M3 — deploy 공존 + update-mode 보호 |
| `c657fd2cc` | M4 — 경계 플래그 + 문서 + neutrality |
| `9da07ac65` | sync 3-phase close (CHANGELOG + SPEC completed) |
| `223b30f0d` | sync backfill (sync_commit_sha placeholder) |
