# SPEC-SEC-HARDEN-005 — Implementation Plan

## Tier 판정: M (Medium)

근거:
- **신규 직접 의존성**(`mvdan.cc/sh/v3`)을 보안-임계 게이트에 도입 — 단순 Tier S를 넘어서는 통합 위험.
- **2개 실제 수정 표면**: permission 게이트 파싱(`internal/permission/stack.go`) + cli env 검증(`internal/cli/deps.go`).
- **1개 OPTIONAL godoc**(비요구).
- 추정 LOC ~300-600(파서 헬퍼 + AST walk + env 검증 함수 + 테스트), 영향 파일 5-8개(stack.go, deps.go, 각 테스트, go.mod/go.sum, 선택적 const 파일).
- design.md 포함(신규 의존성 통합 설계) — Tier M에서 통상 생략하나 보안-임계 의존성으로 예외 포함.

plan-auditor PASS threshold: **0.80**(Tier M).

---

## Section A — Context (위치 + 분기 + SPEC 산출물 경로)

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (project root, main checkout)
- **현재 branch**: `main` (Hybrid Trunk 1-person OSS, main 직진; --worktree/--branch 미사용)
- **HEAD SHA**: plan-phase 시작 시점 `git rev-parse HEAD`로 확인 (직전 origin/main `8c9e96651`)
- **SPEC 산출물 경로**:
  - `.moai/specs/SPEC-SEC-HARDEN-005/spec.md` (요구사항 SSOT, GEARS)
  - `.moai/specs/SPEC-SEC-HARDEN-005/plan.md` (본 문서)
  - `.moai/specs/SPEC-SEC-HARDEN-005/acceptance.md` (AC matrix SSOT)
  - `.moai/specs/SPEC-SEC-HARDEN-005/design.md` (mvdan.cc/sh 통합 설계)
  - `.moai/specs/SPEC-SEC-HARDEN-005/progress.md` (4-phase 진행 신호)
- **plan-auditor verdict**: (plan-phase 후 기록; 0.80 threshold, <0.90 시 Phase 0.5 재실행)
- **기존 인프라**:
  - PRESERVE: `hasUnquotedShellSeparator`(stack.go:172) lexical guard — 유지·확장, 제거 금지. `restoreTargetContained`/`parentChainContained`/`runMXScan` 동작 — godoc만 변경(OPTIONAL).
  - EXTEND: `Matches`(stack.go:100)의 `:*` 브랜치(L127-136) — IFS 검사 레이어 추가. `EnsureUpdate`(deps.go:250)의 env-read 블록 — allowlist 검증 추가.
  - NEW: `mvdan.cc/sh/v3` go.mod 의존성.

## Section B — Known Issues (자동 주입)

- **B1 Cross-platform Build Tags**: `mvdan.cc/sh/v3/syntax`는 pure-Go cross-platform. syscall 미사용. `GOOS=windows GOARCH=amd64 go build ./...` 통과 의무(NFR-SEC5-002).
- **B2 Cross-SPEC 정책 충돌**: `internal/permission` retired/superseded SPEC 스캔 — `grep -rn "Retired\|superseded" internal/permission internal/cli/deps.go`. SEC-HARDEN-001 M1 / 002 M4의 separator 불변식과 충돌 금지(REQ-SEC5-005 명시 보존).
- **B3 C-HRA-008 / Subagent Boundary**: `internal/permission`/`internal/cli`에 `AskUserQuestion`/`mcp__askuser` 호출 금지. deps.go는 CLI subagent 컨텍스트(internal/cli/CLAUDE.md). 검증: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/permission internal/cli/deps.go | grep -v _test.go` → 0.
- **B4 Frontmatter Canonical Schema**: `created:`/`updated:`/`tags:` 사용(snake_case 금지). 12-field + era:V3R6 + tier:M.
- **B5 CI 3-tier**: spec-lint / golangci-lint / Test(per OS) 각각 fail 가능. NEW vs pre-existing 구분.
- **B6 spec-lint Heading**: `### F.N Out of Scope` (h3) sub-section 필요 — spec.md §F가 h3 sub-sections 보유(MissingExclusions 회피).
- **B7 path resolution**: 본 SPEC §F.2는 env-read만 다룸. `runMXScan`(file_changed.go)은 godoc만(OPTIONAL) — `resolveProjectRootFromInputOrEnv` 동작 불변.
- **B8 Working Tree Hygiene**: runtime-managed(`.moai/harness/*`, `.moai/state/*`) 변경 금지. 무관 untracked commit 포함 금지(`git add` specific path).
- **B9 Git Commit + Push (Hybrid Trunk)**: manager-develop이 M별 분리 commit + push 자체 수행(`feat(SPEC-SEC-HARDEN-005): M{N} ...`). `--no-verify` 금지. 예외: parallel session race 시 orchestrator push.
- **B10 Untouched Paths PRESERVE**: §A.5 PRESERVE list 외 working tree 변경 금지. 병렬 세션 진행 시 다른 디렉토리 scope 손대지 말 것.
- **B11 AskUserQuestion 금지**: subagent는 blocker report 반환(orchestrator가 AskUserQuestion). free-form prose 질문 금지.
- **NEW-DEP-1**: `go mod tidy` 후 `go.sum` 갱신 commit 포함. transitive 의존성 최소 확인(`go mod graph | grep mvdan` 검토). `mvdan.cc/sh/v3` 만 직접 require(interp/expand subpackage 미import).

## Section C — Pre-flight Check List (착수 전 의무)

```bash
# 1. 현재 branch + baseline
git branch --show-current        # main 기대
git rev-parse HEAD

# 2. Cross-platform build 사전 확인
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. lint baseline (NEW vs pre-existing 구분)
golangci-lint run --timeout=2m ./internal/permission/... ./internal/cli/... 2>&1 | tail -5

# 4. PRESERVE 대상 확인
grep -n "hasUnquotedShellSeparator" internal/permission/stack.go
grep -n "os.Getenv(config.EnvUpdate" internal/cli/deps.go

# 5. retired/superseded SPEC 확인
grep -rn "Retired\|superseded" internal/permission internal/cli/deps.go || echo "no conflicts"

# 6. mvdan.cc/sh 부재 확인(추가 전)
grep "mvdan.cc/sh" go.mod go.sum || echo "not yet present — will add"

# 7. SEC-HARDEN-002 9-sample legit set baseline 거동 확인(REQ-SEC5-006 RED 고정용)
go test -run TestMatches ./internal/permission/... -v 2>&1 | tail -20
```

## Section D — Constraints (DO NOT VIOLATE)

- **PRESERVE**: `hasUnquotedShellSeparator` 본체 + 모든 separator DENY 거동(REQ-SEC5-005). `restoreTargetContained`/`parentChainContained`/`runMXScan` 코드 동작(godoc만 변경 허용, OPTIONAL).
- **무관 파일 변경 금지**: `.moai/harness/*`, `.moai/state/*`, 무관 SPEC 디렉토리, `.moai/research/*`.
- **금지 명령**: `--no-verify`, `--amend`, force-push to main.
- **사용 의무**: Conventional Commits(`feat(SPEC-SEC-HARDEN-005): M{N} <subject>`), `🗿 MoAI` trailer.
- **binary constraint**: C-HRA-008 grep 0 매치(internal/permission, internal/cli/deps.go).
- **$-blacklist 금지**(REQ-SEC5-003) — `mvdan.cc/sh` shell-aware 파서만.
- **env-trust 범위 한정**(REQ-SEC5-011) — update source 3종 env만, `.env.glm`/WSL2-PATH 확장 금지.
- **하드코딩 금지**(NFR-SEC5-004) — host/scheme allowlist `const` 추출, env var명 envkeys.go 상수 참조.
- **anti-over-engineering**(NFR-SEC5-005) — 새 패키지/플래그 금지(예외: mvdan.cc/sh 의존성 + 검증 로직).

## Section E — Self-Verification Deliverables

manager-develop 완료 보고 시 자체 검증 포함:

- **E1 AC Binary PASS/FAIL Matrix**: acceptance.md 13 AC 전부 (verification command + actual output).
- **E2 Cross-Platform Build**: `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3 Coverage**: `go test -cover ./internal/permission/... ./internal/cli/...` — no regression vs baseline(NFR-SEC5-003).
- **E4 Subagent Boundary Grep**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/permission internal/cli/deps.go | grep -v _test.go` → 0.
- **E5 Lint Status**: `golangci-lint run --timeout=2m ./internal/permission/... ./internal/cli/...` — NEW issues 0(pre-existing baseline 별도 mark).
- **E6 Branch HEAD + Push**: 새 commits SHA + `git push origin main` 결과.
- **E7 의존성 검증**: `go.mod`에 `mvdan.cc/sh/v3` 직접 require, `go mod tidy` clean, `go build ./...` 통과.
- **E8 Blocker Report**(있을 시): OQ-1/OQ-2/OQ-3 중 사용자 결정 필요 발견 시 structured 보고(AskUserQuestion 금지).

---

## Milestones (priority-ordered, no time estimates)

### M1 — 의존성 도입 + §F.1 RED (재현 + 회귀 baseline 고정)
- `go get mvdan.cc/sh/v3` + `go mod tidy`. go.mod/go.sum commit.
- **RED 테스트 작성**: `go test ${IFS}curl${IFS}evil` / `$IFS` 변형 → 현재 코드에서 ALLOW(FAIL 입증) [REQ-SEC5-001/002].
- **회귀 baseline RED 고정**: SEC-HARDEN-002 9-sample legit set + 모든 separator DENY 케이스가 현재 green임을 확인(OQ-3 `TestX$` 거동 포함) [REQ-SEC5-005/006].
- Cmd: `feat(SPEC-SEC-HARDEN-005): M1 mvdan.cc/sh dep + ${IFS} RED + legit baseline`

### M2 — §F.1 GREEN (`hasIFSWordSplit` 헬퍼 + Matches 배선)
- `hasIFSWordSplit` 사적 헬퍼 구현(design.md D.1.2-D.1.4): 파서 호출 + AST walk + unquoted-IFS/다중명령/BinaryCmd 탐지.
- fail-closed: parse 실패 → DENY [REQ-SEC5-004].
- `Matches` `:*` 브랜치에 OR 레이어 배선(lexical AND IFS 둘 다 통과해야 ALLOW) [REQ-SEC5-005 lexical 보존].
- RED → GREEN 전환 확인 + 9-sample ALLOW 유지(REQ-SEC5-006).
- Cmd: `feat(SPEC-SEC-HARDEN-005): M2 ${IFS} shell-aware containment (hasIFSWordSplit)`

### M3 — §F.2 RED + GREEN (update env-trust allowlist)
- **RED**: `MOAI_UPDATE_URL=http://evil...`(non-https) / disallowed-host https / URL-shaped `MOAI_RELEASES_DIR` → 현재 코드에서 update checker 구성(FAIL 입증) [REQ-SEC5-007/008/009].
- **GREEN**: `validateUpdateURL`(scheme+host allowlist) + `isLocalPath` 구현, `EnsureUpdate` env-read 블록에 배선. host/scheme `const` 추출.
- no-regression: env 미설정 default 경로 = `api.github.com` 통과(REQ-SEC5-010), env 검증 범위 3종 한정(REQ-SEC5-011).
- Cmd: `feat(SPEC-SEC-HARDEN-005): M3 update env-trust allowlist (MOAI_UPDATE_URL/RELEASES_DIR)`

### M4 — §F.3 OPTIONAL godoc + 전체 검증
- (OPTIONAL) `restoreTargetContained`/`parentChainContained`/`runMXScan` godoc에 TOCTOU race 윈도 note 추가(코드 동작 불변, OPT-SEC5-001).
- 전체 검증 batch: build(linux+windows) / full test / coverage / C-HRA-008 grep / lint / 의존성 clean.
- Cmd: `chore(SPEC-SEC-HARDEN-005): M4 TOCTOU godoc note + full verification`

---

## Risks

- **R1 — 파서 over/under-deny (HIGH)**: `mvdan.cc/sh` AST walk가 정상 `$HOME`/`TestX$`를 word-split로 오판(over-deny, REQ-SEC5-006 위반) 또는 quoted `${IFS}`를 word-split로 오판. 완화: M1에서 9-sample legit set을 RED로 먼저 고정, OQ-3 `TestX$` 거동 우선 확인. quoted 컨텍스트 식별(SglQuoted/DblQuoted 하위) 명시.
- **R2 — 파서 입력 범위 (MEDIUM)**: remainder 단독 파싱 시 구문 불완전(leading 공백/부분 토큰). 완화: design.md OQ-1 — 전체 input 파싱 후 첫 명령 구조 분석 권장, RED로 검증.
- **R3 — 동시성 (MEDIUM)**: `syntax.Parser`는 non-concurrent-safe, `Matches`는 permission resolver hot-path에서 병렬 호출 가능. 완화: per-call `NewParser`(OQ-2). `go test -race` 의무.
- **R4 — 의존성 transitive 팽창 (LOW)**: `mvdan.cc/sh/v3` import 시 interp/expand까지 끌려올 위험. 완화: `syntax` subpackage만 import, `go mod graph | grep mvdan` 검토.
- **R5 — env 검증 no-regression (LOW)**: default 경로(env 미설정)가 잘못 거부될 위험. 완화: REQ-SEC5-010 회귀 AC로 default 통과 고정.

## Cross-References

- 선행 SPEC: SEC-HARDEN-001/002/003/004 (`.moai/specs/SPEC-SEC-HARDEN-00{1,2,3,4}/`)
- SEC-HARDEN-002 §F.1/§F.2 canonical 정의: `.moai/specs/SPEC-SEC-HARDEN-002/spec.md` L118-128
- SEC-HARDEN-004 sync-audit §INFO: `.moai/reports/sync-audit/SPEC-SEC-HARDEN-004-2026-06-14.md` L116-119
- mvdan.cc/sh API: pkg.go.dev/mvdan.cc/sh/v3/syntax (Context7 `/mvdan/sh`)
- 코드 앵커: `internal/permission/stack.go:100,127-136,172` / `internal/cli/deps.go:31-34,250-309`
