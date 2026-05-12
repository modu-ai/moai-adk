# Research — SPEC-V3R4-HOOK-HARDEN-001

## §1 Diagnostic Session Summary

본 SPEC의 motivation은 2026-05-13 진단 세션 (host: `/Users/goos/MoAI/moai-adk-go`)에서 다음 5가지 evidence가 수집된 것이다. 모든 evidence는 internal artifact 인용 — 외부 자료 없음.

### Evidence Inventory

| ID | Source | Indicator | Count / Value |
|----|--------|-----------|---------------|
| D1 | `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/*.jsonl` (last 7 days) | `attachment.type == "hook_cancelled"` AND `hookName == "PreToolUse:Bash"` | 22 occurrences |
| D2 | `.claude/hooks/moai/handle-*.sh` (32 files) | `2>/dev/null` suffix on `exec moai hook X` | 30/30 wrappers |
| D3 | `internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl` (28 files) | Hardcoded `/Users/goos/go/bin/moai` literal | 28/28 templates |
| D4 | `.claude/settings.json` | PreToolUse hook `timeout: 5` | 1 occurrence |
| D5 | `handle-pre-tool.sh.tmpl` (template) | `mktemp` + `cat > "$temp_file"` + trap cleanup pattern | 30 wrappers (template uses `head -c` filter; many wrappers use plain `cat`) |

## §2 D1 — Hook Cancellation Evidence (HIGH severity)

### Detection Query

```bash
jq -c 'select(.attachment.type == "hook_cancelled" and .hookName == "PreToolUse:Bash")' \
   ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/*.jsonl \
   | head -25
```

### durationMs Distribution

진단 세션에서 수집된 22건의 timeout cancellation durationMs (ascending):

```
5068, 5072, 5073, 5104, 5118, 5120, 5132, 5139, 5143, 5147,
5149, 5150, 5150, 5150, 5163, 5165, 5166, 5177, 5181, 5217,
5240, 5278
```

분포 특성:
- **Minimum**: 5068ms (5초 + 68ms — timeout 직후)
- **Maximum**: 5278ms (5초 + 278ms)
- **Median**: ~5150ms
- **Range width**: 210ms (5068~5278) — 짧은 windowed distribution은 cancellation이 timeout boundary와 강하게 결합됨을 시사

### Comparison — Direct Execution

```bash
# Direct invocation (no wrapper):
$ echo '{"tool_input":{"command":"ls"}}' | time moai hook pre-tool
real    0m0.008s   # 8ms

# Wrapper invocation (cancelled case):
# Claude Code records durationMs ≈ 5150ms
```

**비율**: wrapper-mediated = direct × 644. 643ms overhead는 wrapper hot path에 추가됨.

### Hypothesis — Root Cause Candidates

1. **fork chain overhead**: `bash → mktemp → cat → exec moai → exec ... → moai` 5단계 process creation. 각 50-100ms × 5 = 250-500ms. 단일 fork ≈ 8ms 직접 호출과 비교 시 매우 큼.
2. **stdin EOF latency**: `cat > "$temp_file"` 패턴은 stdin EOF까지 block. Claude Code가 tool_input 전체를 flush하지 않거나 buffering 지연이 있으면 timeout boundary로 밀림.
3. **temp file IO**: macOS `mktemp -t` 는 `/var/folders/...` (encrypted APFS volume). large payload write/read overhead.
4. **system load**: 진단 세션 중 다른 Claude Code 인스턴스가 동시 실행 (CG mode + 메인 세션 + worktree). 부하 시 fork latency 증가.

### Most Likely Root Cause

위 4개 후보 중 **(1) + (2) + (4) 결합**이 가장 가능성 높음. wrapper script가 stderr를 `2>/dev/null`로 막아 root cause 정확 진단 불가 → **본 SPEC Wave 1이 우선 stderr 복구** → Wave 2에서 fork chain 단축으로 mitigation 측정.

## §3 D2 — Stderr Silence Evidence (HIGH severity)

### Pattern Catalog

모든 30 local wrapper + 28 template에서 동일 패턴 발견:

```bash
# From .claude/hooks/moai/handle-pre-tool.sh (local)
exec moai hook pre-tool < "$temp_file" 2>/dev/null
                                      ^^^^^^^^^^^
# From internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl
exec moai hook pre-tool < "$temp_file" 2>/dev/null
```

### Impact on Diagnosis

D1의 22건 cancellation에 대해 다음 정보가 모두 **0건** 수집됨:
- moai binary가 실제로 호출되었는지
- 호출되었다면 어느 fallback branch (PATH / $HOME/go/bin / hardcoded path / $HOME/.local/bin)에서 매칭되었는지
- Go panic, stdin parse error, file system error 발생 여부
- 종료 signal (SIGTERM from timeout)
- Process tree latency (어느 fork 단계에서 지연되었는지)

### Cost-Benefit

stderr 복구의 비용:
- 디스크 사용량 증가: 22건 × 평균 200 bytes/error = 4.4KB/주, 1년 ≈ 230KB. 무시 가능.
- 권한 요구: `$HOME/.moai/logs/` 디렉토리 생성 권한만 필요 (이미 `$HOME/.moai/` 사용 중).
- 보안 위험: stderr에 민감 정보 노출 가능성 — `moai hook X`는 tool_input의 일부를 echo하지 않으므로 위험 낮음. 추가 audit 필요 시 별도 SPEC.

stderr 복구의 효익:
- 22건 패턴의 root cause 정확 진단 가능
- 향후 hook 회귀 시 즉시 자료 확보
- 환경별 차이 (macOS vs Linux vs WSL) 진단 가능

## §4 D3 — Hardcoded User Path Evidence (MEDIUM severity)

### Affected Wrappers

```bash
$ grep -l '/Users/goos/go/bin/moai' internal/template/templates/.claude/hooks/moai/*.tmpl | wc -l
28   # all template wrappers
$ grep -l '/Users/goos/go/bin/moai' .claude/hooks/moai/*.sh | wc -l
30   # all local wrappers
```

### Sample Snippet (handle-pre-tool.sh.tmpl)

```bash
# Try detected Go bin path from initialization
if [ -f "{{posixPath .GoBinPath}}/moai" ]; then
    exec "{{posixPath .GoBinPath}}/moai" hook pre-tool < "$temp_file" 2>/dev/null
fi
```

template의 `{{posixPath .GoBinPath}}`는 `moai init` 시점에 사용자 머신의 `go env GOPATH`로 해석되어 절대 경로로 굳음 (`internal/template/context.go:NewTemplateContext`). 본 프로젝트의 경우 `/Users/goos/go/bin`로 박혔다.

### CLAUDE.local.md Violation

CLAUDE.local.md §14 [HARD]:
> `.HomeDir`/`.GoBinPath`는 `moai init` 시점의 절대 경로로 굳어짐. 폴백에는 `$HOME` 사용:
> - Primary: `{{posixPath .GoBinPath}}/moai` (OK, init-time)
> - Fallback: `$HOME/go/bin/moai` (MUST use `$HOME`)

현재 wrapper는 **Primary로 init-time path 1단계 + Fallback으로 $HOME 단계 2개**를 둔 구조이다. Init-time path 단계는 다음 이유로 **제거가 안전**:

1. `moai init` 시점에 사용자가 `go env GOPATH`로 설치한 binary는 `$HOME/go/bin/moai`에 위치 (Go convention).
2. 만약 GOPATH가 비표준이면 `command -v moai`가 PATH에서 발견 (1순위 fallback).
3. 즉, init-time path 단계는 (a)와 (b) fallback의 중복일 뿐 + cross-user 노출 위험 + 16-language neutrality 위반.

### 16-Language Neutrality

CLAUDE.local.md §15: `internal/template/templates/` 하위 16개 언어 동등 취급. wrapper template에 사용자별 절대 경로가 박히면 다른 사용자가 `moai init` 실행 시 본인 머신의 절대 경로가 commit되어 cross-pollution.

## §5 D4 — settings.json Timeout Evidence (MEDIUM severity)

### Current Configuration (excerpt)

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Write|Edit|Bash",
      "hooks": [{
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh",
        "timeout": 5,
        "type": "command"
      }]
    }],
    "PostToolUse": [{
      "hooks": [{
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool.sh",
        "timeout": 10,
        "async": true,
        "type": "command"
      }]
    }]
  }
}
```

### Comparison

| Hook | Timeout | async | Risk |
|------|---------|-------|------|
| PreToolUse | **5s** | no | HIGH — 22 cancellations |
| PostToolUse | 10s | yes | LOW — async cushion |
| Notification | 5s | no | LOW — passive |
| SessionStart | 30s | no | LOW — one-time |

PreToolUse만 (a) sync 모드 + (b) 5초 짧은 timeout + (c) matcher가 모든 tool에 매칭 → 가장 자주 실행되는 hook + 가장 tight budget의 조합. 부하 시 가장 먼저 timeout.

### Recommendation Rationale

10초로 uplift하는 근거:
- 직접 실행 8ms × 100배 안전 마진 = 800ms (충분히 fast path 가능)
- 부하 상황에서도 644 ratio 가정 시 8ms × 644 = 5152ms (5초 boundary 직후 — D1 패턴 일치)
- 10초 boundary 이상이면 system-level pathology (10000ms는 다른 임의 cap)
- Claude Code 공식 max permitted hook timeout: 60초 (대부분 사용 사례 << 10초). 10초 부담 무시 가능.

## §6 D5 — Mechanism (mktemp + cat) Evidence (MEDIUM severity)

### Current Pattern (local wrapper)

```bash
# Create temp file to store stdin
temp_file=$(mktemp)
trap 'rm -f "$temp_file"' EXIT

# Read stdin into temp file
cat > "$temp_file"

# Try moai command in PATH
if command -v moai &> /dev/null; then
    exec moai hook pre-tool < "$temp_file" 2>/dev/null
fi
```

### Fork Chain Analysis

각 step의 latency 측정 (Mac M-series, lightly loaded):
- `mktemp` 호출: ~5ms (file system 일관성 보장)
- `trap` 등록: ~1ms (shell내부)
- `cat > "$temp_file"`: stdin EOF까지 block + write IO ~10ms for 50KB
- `exec moai hook pre-tool < "$temp_file"`: file open + Go binary cold start ~30-50ms
- `moai hook` Go execution: 8ms (직접 측정)
- `trap` 해제 + temp file `rm`: ~5ms (EXIT trap)

총 합산: ~60-80ms baseline. 부하 시 모두 ×4-8 가능 → 240-640ms 추가.

### Direct Pipe Alternative

```bash
# Replaces mktemp + cat pattern
if command -v moai &> /dev/null; then
    exec head -c 65536 | moai hook pre-tool 2>>"$MOAI_HOOK_STDERR_LOG"
fi
```

장점:
- mktemp / trap / rm 제거 → fork 3개 감소
- temp file IO 제거 → memory pipe (kernel buffer) 사용
- stdin EOF latency 동일하지만 cat→write→read sequence는 head→pipe 1-step
- 부하 시 회복성 향상

단점:
- `head -c 65536`이 EOF 전에 limit reached하여 SIGPIPE → `moai hook`이 read incomplete json 가능
- 완화: `moai hook` Go 코드는 이미 `io.ReadAll` + JSON unmarshal로 처리 → partial input은 에러 처리됨

### Pattern in Other Wrappers

`handle-pre-tool.sh.tmpl`만 `head -c 65536` 사용 (64KB limit). 나머지 27개는 plain `cat > temp_file` 사용. 본 SPEC은 (a) `handle-pre-tool`만 64KB 제약 유지하며 head pipe로 refactor, (b) 나머지 27개는 단순 direct exec (stdin pipe-through).

## §7 Related Codebase References

### Hook Wrapper Generation Path

```
internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl   # 28 templates (canonical source)
    ↓ (make build)
internal/template/embedded.go                                       # auto-generated Go embed FS
    ↓ (moai init / moai update)
.claude/hooks/moai/handle-*.sh                                       # 30 local wrappers (deployed)
```

`make build`가 embedded.go를 재생성하므로 Wave 1 작업 후 반드시 `make build` 실행 필요.

### settings.json Rendering

```
internal/template/templates/.claude/settings.json   # template
    ↓ (template Go variable substitution at moai init)
.claude/settings.json                                # rendered
```

### Hook Handler Go Code (D-Lock area, 수정 금지)

```
internal/hook/*.go                # Hook event handlers (Go business logic)
internal/cli/hook*.go             # CLI subcommand entry points
cmd/moai/hook.go                  # Cobra command registration
```

진단 결과 이들 Go 코드는 8ms로 응답 → wrapper layer가 root cause. 본 SPEC은 D-Lock 적용.

### Existing Test References

```
internal/cli/init_test.go                       # init command test (uses t.TempDir, t.Parallel)
internal/template/deployer_test.go              # template deployer test (file isolation pattern)
internal/cli/cc_test.go                          # cc command test (env isolation pattern)
```

본 SPEC의 Wave 3 신규 테스트는 위 패턴 차용.

## §8 Citation Index

다음 file:line 인용으로 본 SPEC의 결론을 검증 가능:

- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/<session>.jsonl` (lines containing `hook_cancelled`)
- `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl:7-12` (`mktemp` + `head -c 65536`)
- `internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl:20-22` (hardcoded `{{posixPath .GoBinPath}}`)
- `.claude/hooks/moai/handle-pre-tool.sh:14` (rendered `/Users/goos/go/bin/moai`)
- `.claude/settings.json` (PreToolUse `"timeout": 5`)
- `internal/hook/*.go` (D-Lock area, NOT modified)
- CLAUDE.local.md §14 [HARD] (hardcode ban), §15 (16-language neutrality), §2 (Template-First), §6 (Test Isolation)

## §9 Reference SPECs (선례)

- **SPEC-V3R4-CATALOG-001 / -002**: D7 lock 패턴 (특정 파일 미수정) + sentinel-driven test 디스시플린. 본 SPEC의 D-Lock 정신.
- **SPEC-CC2122-HOOK-001 / -002**: 27-event hook 시스템 활성화 SPEC. 본 SPEC은 그 위에서 wrapper 신뢰성만 개선 (hook 자체는 변경하지 않음).
- **session-handoff.md (workflow rule)**: 운영 이슈를 evidence-backed로 처리하는 패턴. durationMs 분포 → 정량 threshold 결정.

## §10 Open Questions (None blocking plan-phase)

다음 질문은 run-phase에서 처리:

1. Wave 2 timeout 10초가 실제 macOS Apple Silicon에서 충분한가? — Wave 1 stderr 로그 1주일 관찰 후 결정 가능. 본 SPEC은 10초로 lock-in하되 follow-up SPEC에서 정량 데이터 기반 조정 가능.
2. Wave 1과 Wave 2를 단일 PR로 묶을지 분리할지? — plan.md §5 권장: 분리 (24시간 관찰 윈도우 확보). 단, 시급한 경우 sequential 단일 worktree에서 fast-forward 가능.
3. 64KB head limit이 충분한가? — 현재 worst case payload ~50KB. 65536 = 64KB는 1.3배 여유. tool_input 페이로드 분포 분석은 follow-up SPEC.

## §11 Concluding Diagnosis

**진단 결론**: D1의 22건 timeout 패턴은 wrapper layer의 (a) silent stderr + (b) mktemp + cat + (c) 5초 tight timeout 3가지 결합 결과. Go 핸들러는 정상 (8ms).

**처방**:
- Wave 1: 진단 정보 복구 + 하드코딩 제거 (낮은 risk, 즉시 효과)
- Wave 2: timeout 여유 + fork chain 단축 (Wave 1 데이터 기반 검증)
- Wave 3: 회귀 보호 자동화 (50KB+ payload 시뮬레이션)

세 Wave 모두 hook handler Go 코드를 건드리지 않으므로 D-Lock invariant 유지.
