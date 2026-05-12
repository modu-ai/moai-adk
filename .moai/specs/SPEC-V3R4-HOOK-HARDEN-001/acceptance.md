# Acceptance Criteria — SPEC-V3R4-HOOK-HARDEN-001

## Schema

본 문서는 hierarchical AC 스키마 (SPC-001 amendment, max depth 3, suffix `-NN` / `.a-z` / `.i-xxvi`)를 따른다. 모든 leaf AC는 `(maps REQ-...)` tail을 포함한다.

식별자 prefix: `AC-HOOK-HARDEN-001-<NN>`

## Wave 1 — Visibility

### AC-HOOK-HARDEN-001-01: Hook wrapper observability and environment neutrality

**Given** 진단 세션 evidence (22 PreToolUse:Bash cancellations) 및 30 wrapper의 stderr silence 상태가 baseline으로 확인된 상태에서,

- **AC-HOOK-HARDEN-001-01.a**: stderr preservation in wrappers
  - **Given** baseline wrapper template `handle-pre-tool.sh.tmpl`이 `2>/dev/null`로 stderr를 silence하는 상태에서,
  - **When** Wave 1 PR이 머지되어 `make build` + `moai update -t`가 실행된 후,
  - **Then** `grep '2>>"\$MOAI_HOOK_STDERR_LOG"' .claude/hooks/moai/handle-*.sh | wc -l`이 30 이상이며, `grep '2>/dev/null' .claude/hooks/moai/handle-*.sh`의 결과가 0건이다.
  - (maps REQ-HOOK-HARDEN-001-001)

    - **AC-HOOK-HARDEN-001-01.a.i**: log directory bootstrap
      - **Given** 로컬 파일시스템에 `$HOME/.moai/logs/` 디렉토리가 존재하지 않는 상태에서,
      - **When** 어떤 hook wrapper가 처음 실행되어 stderr write를 시도할 때,
      - **Then** wrapper는 `mkdir -p "$(dirname "$MOAI_HOOK_STDERR_LOG")"`를 실행하여 디렉토리를 생성하고 stderr write가 성공한다. 디렉토리 생성 실패는 `|| true`로 silently swallow되어 hook execution은 차단되지 않는다.
      - (maps REQ-HOOK-HARDEN-001-002)

    - **AC-HOOK-HARDEN-001-01.a.ii**: empty stderr no-op
      - **Given** moai hook 실행이 stderr 없이 정상 종료된 상태에서,
      - **When** wrapper가 종료될 때,
      - **Then** `$HOME/.moai/logs/hook-stderr.log` 파일은 생성되지만 (touch 효과) 새로운 entry는 append되지 않거나, 또는 파일이 미존재 상태로 유지된다 (both acceptable — `mkdir -p` does NOT touch the file).
      - (maps REQ-HOOK-HARDEN-001-001 supplementary)

- **AC-HOOK-HARDEN-001-01.b**: hardcoded user path removal
  - **Given** 28 template wrapper에 `/Users/goos/go/bin/moai` 또는 유사 사용자 절대 경로가 박혀 있는 baseline 상태에서,
  - **When** Wave 1의 template patch가 적용된 후,
  - **Then** `grep -rE '/(Users|home)/[a-zA-Z][^/]+/' internal/template/templates/.claude/hooks/moai/`의 결과가 0건이다. `command -v moai` → `$HOME/go/bin/moai` → `$HOME/.local/bin/moai` → silent exit의 4-단계 fallback chain만 잔존한다.
  - (maps REQ-HOOK-HARDEN-001-004)

    - **AC-HOOK-HARDEN-001-01.b.i**: render audit on fresh init
      - **Given** 신규 사용자 머신 (`$HOME=/home/alice`, `GOPATH=/home/alice/go`) 상태에서,
      - **When** `moai init test-project`로 새 프로젝트가 생성된 후,
      - **Then** `.claude/hooks/moai/handle-*.sh` 30개 파일 모두에서 `grep -lE '/(Users|home)/[a-zA-Z][^/]+/'`의 결과가 0건이며, 모든 wrapper가 `$HOME` 또는 `$CLAUDE_PROJECT_DIR`로 시작하는 path만 사용한다.
      - (maps REQ-HOOK-HARDEN-001-005)

- **AC-HOOK-HARDEN-001-01.c**: template-local sync verification
  - **Given** template 28개 vs local 30개의 4-file divergence가 baseline으로 존재하는 상태에서,
  - **When** Wave 1의 T1.6 task가 수행되어 분류가 완료된 후,
  - **Then** divergence의 각 4 file은 다음 중 하나로 분류된다: (a) template에 추가됨 (counterpart 생성), 또는 (b) `.gitignore`에 추가되며 rationale이 commit message에 명시됨. `ls .claude/hooks/moai/*.sh | wc -l`과 `ls internal/template/templates/.claude/hooks/moai/*.sh.tmpl | wc -l`의 차이가 (a) 0이 되거나 (b) `.gitignore` 처리된 항목 수만큼만 남는다.
  - (maps REQ-HOOK-HARDEN-001-010)

## Wave 2 — Mechanism

### AC-HOOK-HARDEN-001-02: Timeout uplift, direct pipe, and log rotation

**Given** Wave 1이 merge되어 stderr가 캡처되기 시작한 상태에서, 5초 PreToolUse timeout과 mktemp + cat 패턴이 여전히 baseline인 상황에서,

- **AC-HOOK-HARDEN-001-02.a**: PreToolUse timeout uplift
  - **Given** `internal/template/templates/.claude/settings.json`의 PreToolUse hook `timeout`이 `5`인 상태에서,
  - **When** Wave 2의 T2.1 patch가 적용된 후 `make build`가 실행된 후,
  - **Then** 새로 생성된 `.claude/settings.json` (via `moai init` on fresh dir)의 PreToolUse `timeout` 값이 `10`이며, Notification / PostToolUse-failure / SessionStart 등 다른 hook의 timeout은 변경되지 않는다 (각각 5 / 5 / 30 유지).
  - (maps REQ-HOOK-HARDEN-001-006)

    - **AC-HOOK-HARDEN-001-02.a.i**: drift warning on update
      - **Given** 기존 프로젝트의 `.claude/settings.json`이 `PreToolUse.timeout: 5`인 상태에서,
      - **When** 사용자가 `moai update -t`를 실행할 때,
      - **Then** stdout 또는 stderr에 `[update-warn]` prefix와 함께 `settings.json PreToolUse timeout is 5s; template recommends 10s` 메시지가 출력된다. 그러나 자동 수정은 적용되지 않으며 (REQ-007 자동 적용 금지) 사용자가 수동 편집 시점을 결정한다.
      - (maps REQ-HOOK-HARDEN-001-007)

- **AC-HOOK-HARDEN-001-02.b**: direct pipe mechanism
  - **Given** baseline wrapper에 `mktemp` + `cat > "$temp_file"` + `trap 'rm -f' EXIT` 패턴이 존재하는 상태에서,
  - **When** Wave 2의 T2.3/T2.4 patch가 적용된 후,
  - **Then** `grep -l 'mktemp' .claude/hooks/moai/handle-*.sh`의 결과가 0건이며, 모든 wrapper가 `exec ... | moai hook X` 또는 `exec moai hook X` 패턴을 사용한다 (직접 pipe). `head -c 65536` 패턴은 64KB-limited wrapper (handle-pre-tool.sh)에서만 잔존한다.
  - (maps REQ-HOOK-HARDEN-001-008)

    - **AC-HOOK-HARDEN-001-02.b.i**: 64KB size limit preserved
      - **Given** handle-pre-tool.sh wrapper가 64KB tool_input limit을 enforce해야 하는 상태에서,
      - **When** 65536 bytes (정확히 64KB)를 초과하는 stdin이 wrapper에 전달될 때,
      - **Then** wrapper는 stdout으로 `{"decision":"block","reason":"stdin exceeds 64KB limit"}` JSON을 출력하고 exit 0으로 종료한다 (refactor 이전과 동일 동작 + 동일 decision 효과).
      - (maps REQ-HOOK-HARDEN-001-009)

- **AC-HOOK-HARDEN-001-02.c**: log rotation at 10MB
  - **Given** `$HOME/.moai/logs/hook-stderr.log` 파일이 10485760 bytes (10MB) 이상으로 누적된 상태에서,
  - **When** 어떤 wrapper가 새 invocation에서 rotation check를 수행할 때,
  - **Then** wrapper는 `mv -f $MOAI_HOOK_STDERR_LOG ${MOAI_HOOK_STDERR_LOG}.1`을 실행하여 rename하고, 새 stderr write가 빈 파일에서 시작된다. `mv` 실패는 `|| true`로 swallow되어 hook execution을 차단하지 않는다.
  - (maps REQ-HOOK-HARDEN-001-013)

    - **AC-HOOK-HARDEN-001-02.c.i**: single-level rotation idempotency
      - **Given** `hook-stderr.log.1`이 이전 rotation에서 이미 존재하는 상태에서,
      - **When** 새 rotation이 trigger될 때,
      - **Then** `mv -f`는 기존 `.1`을 overwrite하며, 두 단계 이상의 historical log (예: `.2`, `.3`)는 생성되지 않는다 (단일 레벨 rotation).
      - (maps REQ-HOOK-HARDEN-001-014)

## Wave 3 — Verification

### AC-HOOK-HARDEN-001-03: Reproducibility, opt-out, and absence-resilience tests

**Given** Wave 1 + Wave 2가 merge되어 wrapper layer가 강화된 상태에서, 회귀 방지 자동화 필요 시점에서,

- **AC-HOOK-HARDEN-001-03.a**: load-simulated reproduction test
  - **Given** 신규 테스트 파일 `internal/cli/hook_wrapper_load_test.go`이 추가된 상태에서,
  - **When** `go test -run TestHookWrapper_LargeStdin_DoesNotExceedTimeout ./internal/cli/`이 실행될 때,
  - **Then** 테스트는 50KB+ JSON payload (예: `{"tool_input":{"command":"<random_50000_chars>"}}`)를 `handle-pre-tool.sh` wrapper의 stdin으로 전달하고, 첫 invocation은 warm-up으로 polling, 두 번째 invocation의 end-to-end wall-clock duration이 1000ms 미만임을 `assert`한다.
  - (maps REQ-HOOK-HARDEN-001-011)

    - **AC-HOOK-HARDEN-001-03.a.i**: CI cross-platform compatibility
      - **Given** GitHub Actions matrix (ubuntu-latest, macos-latest, windows-latest)에서 테스트가 실행될 때,
      - **When** Windows runner에서 `bash`가 미가용한 환경이라면,
      - **Then** 테스트는 `t.Skip("bash unavailable on this platform")`로 skip된다. macOS / Linux runners에서는 항상 실행되며 PASS한다.
      - (maps REQ-HOOK-HARDEN-001-012)

    - **AC-HOOK-HARDEN-001-03.a.ii**: hardcode regression sentinel
      - **Given** future PR이 wrapper template에 새로운 절대 경로를 도입하려는 상태에서,
      - **When** `go test -run TestHookWrapper_NoHardcodedUserPath ./internal/template/`이 실행될 때,
      - **Then** sentinel test는 `grep -rE '/(Users|home)/[a-zA-Z][^/]+/'` equivalent regex를 template tree에 적용하여 매칭되는 path가 0건이 아닐 경우 `t.Errorf("HOOK_WRAPPER_HARDCODE: <path>")`로 실패한다.
      - (maps REQ-HOOK-HARDEN-001-005 supplementary)

- **AC-HOOK-HARDEN-001-03.b**: opt-out path honored
  - **Given** 사용자가 read-only filesystem 또는 memory-only filesystem 환경에서 hook 실행 중인 상태에서,
  - **When** `MOAI_HOOK_STDERR_LOG=/dev/null` 환경변수가 설정되고 wrapper가 invoke될 때,
  - **Then** wrapper는 stderr를 `/dev/null`로 redirect하여 디스크 IO 0건 발생하며 hook execution은 정상 완료된다. log file 생성 시도도 없다.
  - (maps REQ-HOOK-HARDEN-001-015)

    - **AC-HOOK-HARDEN-001-03.b.i**: $HOME absence fallback
      - **Given** `$HOME` 환경변수가 unset이거나 빈 문자열인 container / CI 환경에서,
      - **When** wrapper가 stderr redirect를 시도할 때,
      - **Then** wrapper는 `/tmp/moai-hook-stderr.log`로 fallback하거나, 만약 `/tmp`도 쓰기 불가하면 silent exit (exit 0)로 종료한다. hook 자체는 실패하지 않는다.
      - (maps REQ-HOOK-HARDEN-001-016)

- **AC-HOOK-HARDEN-001-03.c**: 80%-budget warning emission (OPT — deferred)
  - **Given** hook execution duration이 80% of configured timeout (예: PreToolUse 10초의 80% = 8초)을 초과한 상태에서,
  - **When** `moai hook X` Go handler가 자체 elapsed time 측정 후 boundary 직전 stderr emit을 수행할 때,
  - **Then** stderr에 `[hook-warn] <hookName> elapsed=<ms>ms threshold=<budget_ms>ms` 라인이 emit되어 captured log에 기록된다.
  - **Note**: 이 AC는 OPT (deferred) — Go 핸들러 코드 수정이 필요하므로 본 SPEC의 D-Lock에 의해 별도 follow-up SPEC으로 분리됨. 본 SPEC의 Definition of Done에 영향 없음.
  - (maps REQ-HOOK-HARDEN-001-003 — deferred)

## Cross-Cutting Acceptance

### AC-HOOK-HARDEN-001-04: Backward compatibility

**Given** 본 SPEC의 모든 Wave가 merge되어 production traffic을 받는 상태에서,

- **AC-HOOK-HARDEN-001-04.a**: existing hook events unaffected
  - **Given** Claude Code 27-event hook system (SPEC-CC2122-HOOK-001/-002 기반)이 active한 상태에서,
  - **When** SessionStart / PostToolUse / Notification 등 PreToolUse 이외의 27개 hook 이벤트가 발생할 때,
  - **Then** 모든 event는 본 SPEC 변경 전과 동일한 timing characteristic (timeout 미변경) + 새로운 stderr capture 이점을 받는다. session JSONL에서 PreToolUse 이외의 hook_cancelled 카운트는 본 SPEC merge 전후 동일하거나 감소한다.
  - (maps REQ-HOOK-HARDEN-001-006 negative case)

- **AC-HOOK-HARDEN-001-04.b**: D-Lock invariant preservation
  - **Given** 본 SPEC의 PR diff가 review되는 상태에서,
  - **When** `git diff --name-only origin/main...HEAD | grep -E '^internal/(hook|cli/hook[^/]*\.go|cmd/moai/hook\.go)'`가 실행될 때,
  - **Then** 결과가 0건이다 (D-Lock 적용 영역의 Go 핸들러 코드는 일절 변경되지 않음). 단, `internal/cli/update.go`는 wrapper hot path가 아니므로 예외이며 본 검사에서 제외된다 (plan.md §3 명시).
  - (maps Constraints — D-Lock)

### AC-HOOK-HARDEN-001-05: Diagnostic effectiveness (post-merge observation)

**Given** Wave 1 + Wave 2가 merge되어 1주일 이상 운영된 상태에서,

- **AC-HOOK-HARDEN-001-05.a**: stderr log content non-empty when timeout occurs
  - **Given** PreToolUse hook이 timeout cancellation이 발생한 상태에서,
  - **When** 사용자가 `tail -n 50 $HOME/.moai/logs/hook-stderr.log`을 실행할 때,
  - **Then** timestamp + hookName + error context가 포함된 entries가 1건 이상 존재한다. 단순히 hook가 timeout만 되고 stderr가 empty인 경우는 없다 (root cause 진단 가능).
  - (maps REQ-HOOK-HARDEN-001-001 effectiveness)

- **AC-HOOK-HARDEN-001-05.b**: cancellation rate reduction
  - **Given** Wave 2 merge 후 1주일 모니터링 window에서,
  - **When** session JSONL의 `attachment.type == "hook_cancelled"` AND `hookName == "PreToolUse:Bash"` 카운트가 측정될 때,
  - **Then** 동일 traffic 패턴 가정 시 카운트가 baseline (22건/주) 대비 75% 이상 감소한다. 단, 본 AC는 실측 데이터 기반이므로 PR merge 시점이 아닌 1주일 후 follow-up 검증으로 평가된다 (CI-blocking이 아닌 monitoring AC).
  - (maps REQ-HOOK-HARDEN-001-006 effectiveness)

## Definition of Done (Cross-reference)

본 acceptance 문서의 모든 AC가 PASS 상태이며:

- AC-HOOK-HARDEN-001-01 (Wave 1): a / a.i / a.ii / b / b.i / c PASS
- AC-HOOK-HARDEN-001-02 (Wave 2): a / a.i / b / b.i / c / c.i PASS
- AC-HOOK-HARDEN-001-03 (Wave 3): a / a.i / a.ii / b / b.i PASS (c는 OPT, deferred)
- AC-HOOK-HARDEN-001-04 (Cross-cutting): a / b PASS
- AC-HOOK-HARDEN-001-05 (Post-merge): a PASS 즉시, b는 1주일 follow-up

총 leaf AC 수: **18 PASS-required** (a.ii의 supplementary 1건 + c deferred 1건 + b 1주일 후 1건 = 본 머지에서 18 + post-merge 1 follow-up = 19 total).

## Trace Summary

| AC ID | Maps REQ | Wave |
|-------|----------|------|
| AC-HOOK-HARDEN-001-01.a | REQ-001 | 1 |
| AC-HOOK-HARDEN-001-01.a.i | REQ-002 | 1 |
| AC-HOOK-HARDEN-001-01.a.ii | REQ-001 (suppl.) | 1 |
| AC-HOOK-HARDEN-001-01.b | REQ-004 | 1 |
| AC-HOOK-HARDEN-001-01.b.i | REQ-005 | 1 |
| AC-HOOK-HARDEN-001-01.c | REQ-010 | 1 |
| AC-HOOK-HARDEN-001-02.a | REQ-006 | 2 |
| AC-HOOK-HARDEN-001-02.a.i | REQ-007 | 2 |
| AC-HOOK-HARDEN-001-02.b | REQ-008 | 2 |
| AC-HOOK-HARDEN-001-02.b.i | REQ-009 | 2 |
| AC-HOOK-HARDEN-001-02.c | REQ-013 | 2 |
| AC-HOOK-HARDEN-001-02.c.i | REQ-014 | 2 |
| AC-HOOK-HARDEN-001-03.a | REQ-011 | 3 |
| AC-HOOK-HARDEN-001-03.a.i | REQ-012 | 3 |
| AC-HOOK-HARDEN-001-03.a.ii | REQ-005 (suppl.) | 3 |
| AC-HOOK-HARDEN-001-03.b | REQ-015 | 3 |
| AC-HOOK-HARDEN-001-03.b.i | REQ-016 | 3 |
| AC-HOOK-HARDEN-001-03.c | REQ-003 (deferred) | OPT |
| AC-HOOK-HARDEN-001-04.a | REQ-006 (negative) | cross |
| AC-HOOK-HARDEN-001-04.b | Constraints D-Lock | cross |
| AC-HOOK-HARDEN-001-05.a | REQ-001 (effectiveness) | post-merge |
| AC-HOOK-HARDEN-001-05.b | REQ-006 (effectiveness) | post-merge (1-week) |

REQ coverage: 16/16 REQs traced (REQ-003은 deferred OPT 명시). REQ 중복 mapping은 supplementary, negative, effectiveness 변형으로 표시.
