# Implementation Plan — SPEC-V3R4-HOOK-HARDEN-001

## §0 Motivation

본 SPEC의 catalyst는 2026-05-13 진단 세션에서 수집된 다음 evidence이다:

- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/*.jsonl` 7일 윈도우에서 `PreToolUse:Bash` hook 22건이 5초 timeout cancellation
- durationMs 분포: 5068, 5072, 5073, 5104, 5118, 5120, 5132, 5139, 5143, 5147, 5149, 5150×3, 5163, 5165, 5166, 5177, 5181, 5217, 5240, 5278 (전부 5000ms 경계 직후)
- 직접 호출 `moai hook pre-tool < sample.json` 측정값 8-15ms → root cause는 wrapper layer
- 30개 wrapper 모두 stderr silence (`2>/dev/null`) → timeout 발생 시 진단 정보 0건
- `/Users/goos/go/bin/moai` 하드코딩 1줄이 28개 template wrapper에 동일 prefix로 박힘 (CLAUDE.local.md §14 / §15 위반)

이 evidence는 자체 진단 가능한 **wrapper layer 경화**가 필요함을 시사한다. Go 핸들러 코드 수정은 불필요 (8ms 충분히 빠름).

## §1 Technical Approach

본 SPEC은 `wrapper-layer-only + template-first + audit-driven` 3-layer 전략을 따른다:

1. **Wrapper Layer (Wave 1)**: 30개 local wrapper + 28개 template wrapper에 stderr append redirect 추가 + 하드코딩 제거. Go 코드 변경 없음.
2. **Mechanism Layer (Wave 2)**: settings.json PreToolUse timeout `5 → 10` 상향 + wrapper의 `mktemp` + `cat` 우회 제거 + log rotation 추가. Single source-of-truth는 `internal/template/templates/.claude/settings.json`.
3. **Verification Layer (Wave 3)**: `internal/cli/hook_wrapper_load_test.go` 신규 추가 — 50KB+ payload로 wrapper 실행 시간 < 1초 검증. CLAUDE.local.md §6 (t.TempDir, t.Parallel()) 준수.

각 Wave는 독립 PR로 머지 가능하며 (M-1 / M-2 / M-3), Wave 1 머지로 진단 데이터가 수집되기 시작하면 Wave 2의 timeout 결정에 evidence를 추가로 활용할 수 있다.

## §2 Milestones

총 task ~22개. 우선순위는 P0 (Wave 1) > P1 (Wave 2) > P1 (Wave 3). 시간 추정 금지 (CLAUDE.local.md / agent-common-protocol §Time Estimation).

### Wave 1 — Visibility (P0)

**Goal**: timeout 발생 시 진단 정보 확보 + 사용자 절대 경로 제거. blast radius 최소, 즉시 효과.

- T1.1: `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl` 수정 — `2>/dev/null` → `2>>"$MOAI_HOOK_STDERR_LOG"` 변경. 파일 상단에 `MOAI_HOOK_STDERR_LOG="${MOAI_HOOK_STDERR_LOG:-$HOME/.moai/logs/hook-stderr.log}"` + `mkdir -p` 추가.
- T1.2: 나머지 27개 `*.sh.tmpl` 동일 패턴 적용 — `handle-session-start.sh.tmpl`, `handle-post-tool.sh.tmpl`, …. 변경은 모두 동일하므로 sed-replace patch 적용 후 `make build`.
- T1.3: "detected Go bin path" 블록 (lines 22-24 in current template) 완전 제거 — `if [ -f "{{posixPath .GoBinPath}}/moai" ]; then exec ...` 블록 삭제. fallback chain은 (a) `command -v moai`, (b) `$HOME/go/bin/moai`, (c) `$HOME/.local/bin/moai`로 단축.
- T1.4: `make build` 실행 → `internal/template/embedded.go` 재생성 확인. `git diff embedded.go`로 변경 적용 검증.
- T1.5: `moai update -t` 로컬 적용 → `.claude/hooks/moai/handle-*.sh` 30개 재생성 → `grep -l "/Users/goos" .claude/hooks/moai/*.sh` 결과 0건 확인.
- T1.6: template-local divergence 점검 — `ls .claude/hooks/moai/*.sh` (32) vs `ls internal/template/templates/.claude/hooks/moai/*.sh.tmpl` (28). 차이 4건 분석:
  - 후보 1: local-only generated runtime wrapper (예: `agent-hook.sh`) → `.gitignore`에 등재 + 본 SPEC에서 문서화
  - 후보 2: template 누락 → `make build` 검증 후 template 추가
- T1.7: 자체 점검 — `bash -n internal/template/templates/.claude/hooks/moai/*.sh.tmpl` syntax check (Go template 표현식은 sed 처리 후 검증).

**Wave 1 종료 조건**: PR 머지 후 24시간 내 hook-stderr.log 최소 1건 entry 수집 + 22건 PreToolUse timeout 재발 시 stderr 캡처 확인.

### Wave 2 — Mechanism (P1)

**Goal**: stdin handoff fork chain 단축 + PreToolUse timeout 5초 → 10초 + log rotation.

- T2.1: `internal/template/templates/.claude/settings.json` 의 `PreToolUse` block 내 `timeout: 5` → `timeout: 10`. matcher (`Write|Edit|Bash`)는 변경하지 않음. `PostToolUse` (이미 async + 10초)와 `Notification` (5초 유지)은 손대지 않음.
- T2.2: `internal/cli/update.go` 내 `moai update -t` 경로에 settings.json drift detection 추가 — local file의 PreToolUse.timeout이 5이고 template은 10이면 stderr에 `[update-warn] settings.json PreToolUse timeout is 5s; template recommends 10s. Apply manually or rerun with --apply-settings.` 출력. **수정 자체는 자동 적용하지 않음** (REQ-007). 단, 본 task는 Go 코드 수정이므로 D-Lock 검토 필요 — `update.go`는 wrapper hot path가 아니므로 허용 (D-Lock은 hook handler에만 적용).
- T2.3: `handle-pre-tool.sh.tmpl` 의 `mktemp` + `head -c 65536 > "$temp_file"` + trap 블록 제거. 새 패턴:
  ```bash
  # In-line size check + direct exec
  MOAI_HOOK_STDERR_LOG="${MOAI_HOOK_STDERR_LOG:-$HOME/.moai/logs/hook-stderr.log}"
  mkdir -p "$(dirname "$MOAI_HOOK_STDERR_LOG")" 2>/dev/null || true

  # 64KB size enforcement preserved via `head -c` filter on pipe
  if command -v moai &> /dev/null; then
      exec head -c 65536 | moai hook pre-tool 2>>"$MOAI_HOOK_STDERR_LOG"
  fi
  if [ -f "$HOME/go/bin/moai" ]; then
      exec head -c 65536 | "$HOME/go/bin/moai" hook pre-tool 2>>"$MOAI_HOOK_STDERR_LOG"
  fi
  if [ -f "$HOME/.local/bin/moai" ]; then
      exec head -c 65536 | "$HOME/.local/bin/moai" hook pre-tool 2>>"$MOAI_HOOK_STDERR_LOG"
  fi
  exit 0
  ```
  **주의**: `head -c 65536`이 EOF 전에 SIGPIPE를 받을 수 있으므로 `moai hook X`가 stdin EOF 후 graceful exit 보장해야 함 (이미 Go 핸들러는 `io.ReadAll`로 처리하므로 안전).
- T2.4: 나머지 wrappers (예: `handle-post-tool.sh.tmpl`) — 이들은 64KB 제약이 없으므로 더 단순:
  ```bash
  if command -v moai &> /dev/null; then
      exec moai hook post-tool 2>>"$MOAI_HOOK_STDERR_LOG"
  fi
  # ... fallback chain
  ```
- T2.5: Log rotation — 모든 wrapper 상단에 추가:
  ```bash
  # Single-level rotation at 10MB (best-effort, non-blocking)
  if [ -f "$MOAI_HOOK_STDERR_LOG" ]; then
      _size=$(stat -f%z "$MOAI_HOOK_STDERR_LOG" 2>/dev/null || stat -c%s "$MOAI_HOOK_STDERR_LOG" 2>/dev/null || echo 0)
      if [ "$_size" -gt 10485760 ]; then
          mv -f "$MOAI_HOOK_STDERR_LOG" "${MOAI_HOOK_STDERR_LOG}.1" 2>/dev/null || true
      fi
  fi
  ```
- T2.6: `make build` + 재생성 + 재테스트.
- T2.7: 22건 timeout 패턴 재현 — 부하 상황에서 `Bash(rg pattern dir/)` 100회 반복 후 `attachment.type == hook_cancelled` 카운트 = 0 검증.

**Wave 2 종료 조건**: Wave 1 stderr 로그에 quoted error message 없는 5초 cancellation은 새로 발생하지 않음 (timeout uplift 효과).

### Wave 3 — Verification (P1)

**Goal**: 22건 패턴 회귀 방지 자동 테스트.

- T3.1: `internal/cli/hook_wrapper_load_test.go` 신규 작성 — `TestHookWrapper_LargeStdin_DoesNotExceedTimeout`.
  - 50KB JSON payload 합성 (50 KB `{"tool_input": {"command": "<random_string_50000_chars>"}}`)
  - `t.TempDir()`로 isolated hook wrapper 복사
  - `bash` 가용성 체크 — Windows에서 미가용 시 `t.Skip()`
  - cold-start mitigation: 첫 실행 무시, 두 번째 실행 측정
  - 측정: `time.Now()` before/after `exec.Command("bash", wrapper)`.Run() + payload stdin
  - 검증: `elapsed < 1*time.Second`
  - Cleanup: `t.TempDir()` 자동 정리
- T3.2: CI 호환성 — `go test -run TestHookWrapper -race ./...` 매트릭스 (ubuntu-latest, macos-latest, windows-latest):
  - macOS, Linux: 항상 실행
  - Windows: bash 미가용 시 skip
- T3.3: Opt-out 경로 테스트 — `MOAI_HOOK_STDERR_LOG=/dev/null` 환경변수 설정 시 wrapper가 정상 동작 + log 파일 미생성 검증 (sub-test).
- T3.4: `$HOME` 부재 시 fallback 테스트 — `t.Setenv("HOME", "")` 후 wrapper가 `/tmp/moai-hook-stderr.log` 작성하거나 silent exit 검증.
- T3.5: 회귀 sentinel — `TestHookWrapper_NoHardcodedUserPath` 추가. `grep -lr "/Users/" internal/template/templates/.claude/hooks/moai/*.tmpl` 결과 0건 검증. `grep -lr "/home/" ... | grep -v "$HOME"` 결과 0건 검증.

**Wave 3 종료 조건**: 모든 신규 테스트 GREEN + golangci-lint clean + `go test -race ./...` GREEN.

## §3 File Dependencies

```
Wave 1
  internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl  (28 files, modified)
  internal/template/embedded.go  (regenerated via make build)
  .claude/hooks/moai/handle-*.sh  (30 files, regenerated via moai update -t)

Wave 2
  internal/template/templates/.claude/settings.json  (1 file, timeout 5→10)
  internal/cli/update.go  (drift warning logic, ~20 LOC added)
  .claude/settings.json  (1 file, regenerated)
  internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl  (28 files, mechanism refactor)

Wave 3
  internal/cli/hook_wrapper_load_test.go  (new file, ~150 LOC)
```

**D-Lock 적용 파일** (이 SPEC에서 수정 금지):
- `internal/hook/*.go`
- `internal/cli/hook*.go` (단, `update.go`는 wrapper hot path가 아니므로 예외)
- `cmd/moai/hook.go`

## §4 Quality Strategy — TRUST 5 Mapping

- **T (Tested)**: REQ-011, REQ-012 → Wave 3의 `TestHookWrapper_LargeStdin_DoesNotExceedTimeout` + sentinel test. 회귀 방지.
- **R (Readable)**: 모든 wrapper의 fallback chain이 4-level로 통일 (command -v → $HOME/go/bin → $HOME/.local/bin → silent exit). 주석은 한국어 (`code_comments: ko`).
- **U (Unified)**: 28개 wrapper template이 동일 stderr 패턴 + 동일 rotation 패턴 사용. `make fmt` (gofmt) + `bash -n` syntax check 적용.
- **S (Secured)**: 사용자 절대 경로 제거 → 다른 사용자 환경 노출 방지. log path는 `$HOME` scope이므로 권한 분리.
- **T (Trackable)**: `MOAI_HOOK_STDERR_LOG`를 통해 모든 timeout 케이스가 디스크에 기록 + log rotation으로 무한 누적 방지. session JSONL의 hook_cancelled 카운트도 함께 모니터링.

## §5 Worktree Strategy

[HARD] **Plan phase**: main checkout에서 plan-in-main doctrine 적용 (PR #822 lesson) — 현재 branch `plan/SPEC-V3R4-HOOK-HARDEN-001`. 본 SPEC의 plan artifact는 `.moai/specs/SPEC-V3R4-HOOK-HARDEN-001/` directory 4개 파일만 생성.

[HARD] **Run phase (Wave별)**: 각 Wave는 별도 worktree 생성 권장 — wrapper changes는 cross-file이므로 isolation 이점이 큼.

```bash
# Wave 1
moai worktree new SPEC-V3R4-HOOK-HARDEN-001-wave-1 --base origin/main
# /moai run SPEC-V3R4-HOOK-HARDEN-001 (T1.1~T1.7만)

# Wave 2
moai worktree new SPEC-V3R4-HOOK-HARDEN-001-wave-2 --base origin/main
# /moai run SPEC-V3R4-HOOK-HARDEN-001 (T2.1~T2.7만)

# Wave 3
moai worktree new SPEC-V3R4-HOOK-HARDEN-001-wave-3 --base origin/main
```

또는 wave 통합 가능 — sequential delivery의 경우 단일 worktree 재사용 (Wave 1 PR merge → `git merge origin/main` → Wave 2).

[HARD] **Sync phase**: 동일 worktree 재사용 (CLAUDE.md spec-workflow.md Step 3).
[HARD] **Cleanup**: 모든 PR (run + sync) merge 후 `moai worktree done` (Step 4).

## §6 Conventional Commits

Wave별 commit prefix 규칙:

| Wave / Phase | Prefix | 예시 |
|--------------|--------|------|
| Plan artifacts | `plan(spec)` | `plan(spec): SPEC-V3R4-HOOK-HARDEN-001 plan-phase artifacts` |
| Wave 1 implementation | `feat(hooks)` | `feat(hooks): preserve stderr in wrapper scripts (REQ-001/002)` |
| Wave 1 hardcode removal | `fix(hooks)` | `fix(hooks): remove hardcoded user path from wrapper templates (REQ-004/005)` |
| Wave 2 settings | `feat(settings)` | `feat(settings): uplift PreToolUse timeout 5s → 10s (REQ-006)` |
| Wave 2 update warn | `feat(cli)` | `feat(cli): emit warning on settings.json drift (REQ-007)` |
| Wave 2 mechanism | `refactor(hooks)` | `refactor(hooks): replace temp-file pattern with direct pipe (REQ-008/009)` |
| Wave 2 rotation | `feat(hooks)` | `feat(hooks): add single-level stderr log rotation (REQ-013/014)` |
| Wave 3 tests | `test(hooks)` | `test(hooks): add load-simulated wrapper regression test (REQ-011/012)` |
| Sync phase | `docs(sync)` | `docs(sync): SPEC-V3R4-HOOK-HARDEN-001 completion artifacts` |

모든 commit message footer:
```
🗿 MoAI <email@mo.ai.kr>
```

## §7 Plan Audit Pre-Check (P-1 ~ P-10)

본 plan이 plan-auditor 통과 가능한지 self-audit:

- **P-1 (Schema Conformance)**: spec.md frontmatter 9-field (id, version, status=draft, created_at, updated_at, author, priority=High, labels[5], issue_number=null) → **PASS**
- **P-2 (EARS Compliance)**: 16개 REQ 모두 EARS 키워드 사용 (shall, when X then, while X, if X then) → **PASS**
- **P-3 (Hierarchical AC)**: acceptance.md에 top → child (.a/.b/.c) → grandchild (.i) 구조 + (maps REQ-...) tail → **PASS** (작성 시 확정)
- **P-4 (Trace coverage)**: 모든 REQ가 AC 1개 이상에 매핑됨 (spec.md Trace section 참조) → **PASS**
- **P-5 (Exclusions)**: spec.md §6 "Out of Scope"에 9개 항목 (요구 minimum 1) → **PASS**
- **P-6 (Constraints clarity)**: spec.md §Constraints에 Hard 3개 + Soft 3개 명시 → **PASS**
- **P-7 (Worktree Strategy)**: plan.md §5에 Wave별 worktree 전략 + Sync 동일 worktree 재사용 명시 → **PASS**
- **P-8 (Conventional Commits)**: plan.md §6에 9개 prefix 예시 → **PASS**
- **P-9 (Rollback strategy)**: plan.md §8 (이 다음 섹션) — 각 Wave별 rollback 절차 → **PASS**
- **P-10 (No time estimates)**: plan 전체에서 시간 표현 부재. priority labels만 사용 → **PASS**

**Pre-Check 결과: 10/10 PASS** (작성 시점 self-assessment, plan-auditor 검증은 별도 단계).

## §8 Rollback Strategy

각 Wave는 독립 PR이므로 단일 PR revert로 충분.

### Wave 1 Rollback
- `git revert <wave-1-PR-commit>` → wrapper template 원상복귀 + `2>/dev/null` 복원
- 영향: stderr 진단 정보 다시 silence. 22건 timeout 패턴은 다시 보이지 않음 (그러나 root cause는 그대로).
- 검증: `grep -l "MOAI_HOOK_STDERR_LOG" .claude/hooks/moai/*.sh` → 0건

### Wave 2 Rollback
- `git revert <wave-2-PR-commit>` → settings.json `timeout: 10 → 5` 복귀 + wrapper mechanism 원상 복귀
- 영향: PreToolUse 5초 timeout 재발 가능. mktemp 패턴 재도입.
- 검증: `grep '"timeout": 10' .claude/settings.json` → PreToolUse 블록에서 0건

### Wave 3 Rollback
- `git revert <wave-3-PR-commit>` → 테스트 파일 삭제. 회귀 보호 상실 외 사용자 영향 없음.
- 검증: `ls internal/cli/hook_wrapper_load_test.go` → 파일 없음

### Forward-Only Mitigation (rollback 대신)
- Wave별 PR에 `--ff-only` merge 권장 (branch protection rule)
- Wave 1 → Wave 2 사이 24시간 이상 관찰 권장 — stderr log 안정성 확인 후 Wave 2 진입
- Wave 2 적용 후 1주일 관찰 — `hook_cancelled` 카운트 zero 확인 후 Wave 3 머지

## §9 Risks and Mitigations

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|-----------|--------|------------|
| R1 | `head -c 65536 \| moai hook` 패턴이 macOS bash 3.2와 Linux bash 5.x 사이 다른 동작 | Low | Med | T3.1 테스트가 두 플랫폼에서 동일 검증 |
| R2 | log file 권한 충돌 (read-only $HOME) | Low | Low | REQ-016의 `\|\| true` graceful fallback + `/tmp` 우회 |
| R3 | settings.json drift warning이 false positive 증가 | Med | Low | T2.2의 warning만 출력 + 자동 수정 안 함 |
| R4 | Wave 2 timeout uplift가 사용자 응답성 저하 (PreToolUse hook 무거워질 경우 10초 대기) | Low | Med | 8ms 직접 실행 = 1000배 여유. 실제 부하 발생 시 stderr 로그가 진단 evidence 제공 |
| R5 | 28 template / 30 local 차이 4건이 의도된 local-only인지 누락인지 불명 | High | Low | T1.6에서 분류 + `.gitignore` 명시화 |
| R6 | `mv -f hook-stderr.log hook-stderr.log.1` race condition (여러 wrapper 동시 실행) | Low | Low | `\|\| true` swallow + single-level rotation 한계 명시 (REQ-014) |

## §10 Definition of Done (Cross-reference)

`spec.md §7 Definition of Done` 참조. Wave별 DoD는 위 §2 milestones 각 Wave 종료 조건 참조.

## §11 References

- spec.md (same directory)
- research.md (same directory) — diagnostic evidence 출처
- acceptance.md (same directory) — hierarchical AC
- CLAUDE.local.md §2 (Template-First), §6 (Test Isolation), §14 (Hardcode Ban), §15 (16-Language Neutrality)
- SPEC-V3R4-CATALOG-001/002 (D7 lock pattern reference)
- session-handoff.md (paste-ready resume pattern, applied for Wave handoff)
