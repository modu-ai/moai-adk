# Design — SPEC-HOOK-DISCIPLINE-WIRING-001

> Technical design for (1) sync-gate language-auto-detect generalization and (2) settings.json.tmpl hook insertion.
> Status: draft. WHAT/WHY lives in spec.md; this is the HOW for run-phase implementation.

## 1. sync-gate Language-Auto-Detect Generalization (the core design)

### 1.1 Current state (Go-hardcoded)

`sync-phase-quality-gate.sh`는 현재 `go vet ./...`, `golangci-lint run`, `go test ./...`, `go test -cover ./...`를 무조건 호출하고, sync-phase 커밋 감지 후 `git diff --name-only ... | grep -cE '\.go$'`로 Go 파일 델타를 검사한다. 모두 Go 전용이다.

### 1.2 Target design — marker-driven language detection

CLAUDE.md §7 canonical 매트릭스를 그대로 채택한다:

| Language | Project marker (root) | Lint/vet | Test |
|----------|----------------------|----------|------|
| Go | `go.mod` | `go vet` → `golangci-lint`(opt) | `go test ./...` |
| Node.js | `package.json` | `eslint`(opt) | `npm test` |
| Python | `pyproject.toml` OR `requirements.txt` | `ruff`(opt) | `pytest` |
| Rust | `Cargo.toml` | `cargo clippy`(opt) | `cargo test` |

**감지 함수 (의사 설계)**:

```
detect_language(project_root):
  if exists "$project_root/go.mod"         → "go"
  elif exists "$project_root/package.json"  → "node"
  elif exists "$project_root/pyproject.toml" or exists "$project_root/requirements.txt" → "python"
  elif exists "$project_root/Cargo.toml"    → "rust"
  else                                       → ""   # no marker → silent pass
```

우선순위는 위 표 순서(go → node → python → rust). 모노레포(복수 마커)의 경우 **첫 매치 단일 언어**를 처리한다 (단순성 우선; 복수-언어 순차 실행은 over-engineering이며 본 SPEC scope 밖). 이 결정은 edge case AC-HDW D.7-3에 반영.

[HARD — D1 testability mandate] `detect_language()`는 **직접 호출 가능한(directly-invocable / source-able) 셸 함수**로 구현한다. 즉 sync-phase-commit git 게이트(아래 §1.8)를 통과하지 않고도 `source sync-phase-quality-gate.sh && detect_language "$dir"` 형태로 독립 호출·단위 검증이 가능해야 한다. 이는 AC-HDW-002(a)의 직접-호출 검증과 AC-HDW-009의 case-block 추출을 가능케 한다. 함수는 부수효과(toolchain 실행, exit) 없이 언어 문자열만 stdout으로 반환한다.

### 1.8 실행 순서 — git 게이트가 detect_language()보다 먼저 (D1 근거)

[HARD] 기존 스크립트의 실행 순서를 명시한다(런타임 AC가 이 순서에 의존):

1. `--skip-hook` 체크 (line 14)
2. **sync-phase-commit 감지** (`git log -1 --format='%s'`, lines 23-32) — 이 게이트가 **가장 먼저** 단락(short-circuit)한다. sync-phase 커밋이 아니면 `exit 0`("not a sync-phase commit")로 즉시 종료하며 **detect_language()에 도달하지 않는다**.
3. 코드 파일 델타 skip (markdown-only sync)
4. **detect_language()** + 언어별 toolchain (← 신규 일반화 영역)

[HARD — D1 검증 함의] `mktemp -d`는 git 저장소가 아니므로 step 2의 `git log -1`이 exit 128 → "not a sync-phase commit"로 단락한다. 따라서 **AC-HDW-002의 런타임 fixture는 반드시 실제 git 저장소**(`git init` + sync-phase subject 커밋 + non-`.go` 델타)여야 detect_language()에 도달한다. 비-git temp dir 호출은 언어 감지를 전혀 실행하지 못하므로 vacuous(이전 acceptance.md 결함). 또는 detect_language()를 직접 호출(§1.2 mandate)하여 게이트를 우회 검증한다.

### 1.3 Graceful skip mechanism

각 toolchain step은 `command -v <tool>` 가드로 감싼다. 도구 부재 시 그 step만 skip(exit code 0 기록, 로그에 `skipped (tool absent)`)하고 다음 step으로 진행. 이미 기존 코드가 `golangci-lint`에 대해 이 패턴을 사용하므로(`if command -v golangci-lint ...`), 모든 언어의 모든 도구로 확장한다.

```
run_step(tool, cmd):
  if command -v "$tool" >/dev/null 2>&1:
    run cmd; record exit
  else:
    record "0" (skipped); log "skipped: $tool absent"
```

CLAUDE.md §7: "Tools that are not installed are skipped gracefully." 이 설계가 그 의미론을 직접 구현한다.

### 1.4 Silent pass (no marker)

`detect_language`가 빈 문자열 반환 시 어떤 toolchain도 실행하지 않고 즉시 `exit 0`하며 decision JSON에 `{"decision":"skip","reason":"no recognized language marker"}`. CLAUDE.md §7: "projects with no recognized language marker pass the gate silently."

### 1.5 Language-generalized code-delta skip

기존 `\.go$` 델타 검사를 언어별 확장자로 일반화:
- go → `\.go$`
- node → `\.(js|ts|jsx|tsx|mjs|cjs)$`
- python → `\.py$`
- rust → `\.rs$`

코드 파일 델타 0이면 markdown-only sync로 보고 skip(기존 의미 보존).

### 1.6 Warn-first (advisory) decision

decision 산출 로직은 유지하되, 본 SPEC에서는 **exit 2 경로를 등록 기본값에서 활성화하지 않는다**. 구현 선택지 2가지:

- (선호) 스크립트 내 `BLOCKING` 환경/플래그 게이트를 두고, 미설정(기본) 시 `decision="block"`이어도 `exit 0`(advisory 로그만). exit-2는 후속 SPEC에서 게이트 ON.
- (대안) 본 SPEC에서는 decision을 `allow`/`skip`만 산출하고 `block`+exit2 분기를 dormant 주석으로 보존.

어느 쪽이든 §Exclusions item 1(exit-2 deferred) 준수가 핵심. AC-HDW-008이 passing-scenario exit 0로 검증.

### 1.7 Neutrality constraint (central risk)

Go-tool 키워드(`go vet`, `golangci-lint`, `go test` 호출)는 **반드시 Go 분기 case 내부에만** 위치해야 한다. 비-Go 언어 키워드(eslint/ruff/cargo/pytest/npm)도 각 언어 case 내부에만. 어떤 단일 언어 toolchain 키워드도 공통/default 경로에 노출되면 안 된다 — 단, 다중 언어 **마커**(`go.mod`/`package.json`/`pyproject.toml`/`Cargo.toml`)를 `detect_language()` 안에서 **동등하게** 나열하는 것은 중립적이므로 허용 (16개 언어 중 4개를 동등 case로 나열하는 것은 §15 언어 중립성 위반 아님; "한 언어를 PRIMARY로 배치"하지 않으면 됨).

[중요 — D3 CI guard scope 정정] neutrality CI guard(`internal_content_leak_test.go` + `template_neutrality_audit_test.go`)는 **내부 콘텐츠 누출만** 검사한다 — 실제 탐지 클래스는 **C1 macOS-bias-path / C2 V3R-sigil / C4 feedback-memory-ref / C5 CLAUDE.local-ref / C6 PR#-ref / C7 SHA / C8 GOOS-preserve / spec-id-date** (테스트 자신의 subtest 이름으로 실증 확인). **언어 편향(Go-bias / language-bias) 탐지 클래스는 없다**. 증거: 현재 하드코딩된 `go vet` 4회에도 guard는 GREEN. 따라서 본 스크립트의 **Go-bias 부재(언어 중립성)는 CI guard가 보증하지 못하며**, 대신 다음 두 가지로 입증한다:

- **AC-HDW-009 (구조적 자동 guard)**: Go-tool 토큰이 Go case branch 내부에만 존재하는지 `awk` case-block 추출 + `total == inblk` 카운트 비교로 검증. 이것이 central neutrality risk에 대한 실제 자동 guard다.
- **AC-HDW-002 (런타임 guard)**: 실제 git 저장소 non-Go fixture에서 Go toolchain 미실행 + detect_language() 직접 호출 `node` 반환.

본 스크립트가 내부 SPEC-ID/REQ 토큰을 포함하지 않으므로 CI guard(C1-C8)는 별개로 통과한다(AC-HDW-006). 즉 AC-HDW-006(내부누출 부재)과 AC-HDW-009/002(언어중립성)는 **서로 다른 두 속성**을 검증한다 — 혼동 금지.

## 2. settings.json.tmpl Insertion Design

### 2.1 Existing wrapper pattern (verbatim mirror)

모든 기존 엔트리는 다음 platform-conditional 패턴을 사용:

```
{{- if eq .Platform "windows"}}
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/<script>.sh\"",
{{- else}}
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/<script>.sh\"",
{{- end}}
            "timeout": <N>,
            "type": "command"
```

신규 두 엔트리는 이 패턴을 verbatim 미러한다 (REQ-HDW-008).

### 2.2 PostToolUse insertion (status-transition)

기존 PostToolUse 블록(line ~67-82)의 `hooks` 배열에 두 번째 객체로 추가:

```
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          { ...existing handle-post-tool.sh (async)... },
          {
{{- if eq .Platform "windows"}}
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/status-transition-ownership.sh\"",
{{- else}}
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/status-transition-ownership.sh\"",
{{- end}}
            "timeout": 5,
            "type": "command"
          }
        ]
      }
    ]
```

- **timeout 5s**: status-transition은 jq + grep + 단일 파일 읽기로 가볍다.
- **matcher**: 기존 PostToolUse matcher(`Write|Edit`) 공유. 훅 자체가 SPEC-artifact 경로(`*.moai/specs/SPEC-*/spec.md` 등)로 self-filter하므로 비-SPEC 파일은 즉시 allow exit 0 — 광범위 matcher여도 안전.
- **async 미사용**: status-transition은 동기 advisory(빠름). 기존 handle-post-tool.sh의 `async: true`는 그대로 두되, 신규 엔트리는 async 불필요.

### 2.3 Stop insertion (sync-gate)

기존 Stop 블록(line ~84-106)의 `hooks` 배열에 추가. 기존 구조는 `handle-stop.sh` + 조건부 `{{ if .HookOptIn.Enabled }} handle-harness-observe-stop.sh {{ end }}`. 신규 sync-gate 엔트리는 무조건(항상) 추가하되 기존 둘과 공존:

```
    "Stop": [
      {
        "hooks": [
          { ...existing handle-stop.sh... },
          {
{{- if eq .Platform "windows"}}
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/sync-phase-quality-gate.sh\"",
{{- else}}
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/sync-phase-quality-gate.sh\"",
{{- end}}
            "timeout": 10,
            "type": "command"
          }{{ if .HookOptIn.Enabled }},
          { ...existing handle-harness-observe-stop.sh... }{{ end }}
        ]
      }
    ]
```

- **timeout 10s**: sync-gate는 test 실행 가능(go test 등) → 10s 부여 (lifecycle 표 Stop=10s 정합).
- **JSON 콤마 주의**: 조건부 `handle-harness-observe-stop.sh`가 `{{ if }},`로 선행 콤마를 붙이는 패턴이므로, sync-gate 엔트리를 그 앞에 둘 때 콤마 위치를 정확히 유지 (sync-gate 객체 뒤 `{{ if .HookOptIn.Enabled }},` → harness-observe). 렌더 후 유효 JSON 검증 필수(AC-HDW-005).

### 2.4 TaskCompleted — no change

`team-ac-verify.sh`는 등록하지 않으므로 TaskCompleted 블록은 수정하지 않는다 (REQ-HDW-007). 기존 `handle-task-completed.sh`만 유지.

### 2.5 Local settings.json mirror

local `.claude/settings.json`(non-templated, git-tracked)에 동일 두 엔트리를 platform-conditional 없이 직접(Unix 형태) 추가 — local은 dev machine(darwin)이므로 `"$CLAUDE_PROJECT_DIR/..."` 직접 형태. dev-intent 키(defaultMode/env.PATH/teammateMode)는 미접촉 (REQ-HDW-010). hook 배열 엔트리만 ADD.

## 3. Build & Sync Flow

1. template source 편집 (sync-gate.sh + settings.json.tmpl) — Template-First (CLAUDE.local.md §2).
2. `make build` → embedded 재생성 + go:embed 자산 갱신.
3. local mirror 동기화: sync-gate.sh local ← template (byte-identical), local settings.json 수동 ADD.
4. 검증: `moai init` smoke (tmpl 렌더 유효 JSON) + neutrality guard + parity diff.

## 4. Design Decisions / Trade-offs

| Decision | Rationale | Alternative rejected |
|----------|-----------|----------------------|
| 첫-매치 단일 언어 감지 | 단순성; 모노레포는 드물고 복수-언어 순차는 over-engineering | 모든 감지 언어 순차 실행 (복잡도↑, scope creep) |
| sync-gate 직접 등록 (wrapper 없음) | 스크립트가 self-contained 최종 핸들러 | 신규 handle-sync-gate.sh wrapper (불필요한 간접) |
| warn-first (exit-2 deferred) | 점진적 도입, sync 차단 위험 회피 | exit-2 즉시 활성 (예기치 않은 Stop 차단 위험) |
| team-ac 파일 보존 + 미등록 | Go task_completed.go가 더 강력; 재활성 여지 보존 | 파일 삭제 (재활성 불가, 회귀 위험) |
| status-transition timeout 5s / sync-gate 10s | lifecycle 표 정합 (가벼운 검사 5s, 테스트 포함 10s) | 균일 timeout (테스트 시간 부족 또는 과다) |

## 5. Cross-References

- spec.md §D (REQ-HDW-001..010)
- acceptance.md (AC-HDW-001..008, D.7 edge cases)
- CLAUDE.md §7 (언어 매트릭스), §settings.json.tmpl 기존 패턴 (lines 67-128, 201-211)
- CLAUDE.local.md §2 (Template-First), §15 (언어 중립성), §22 (dev intent), §25 (neutrality isolation)
- `.claude/rules/moai/core/hooks-system.md` (이벤트/timeout 표)
