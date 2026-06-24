# Research — SPEC-HOOK-DISCIPLINE-WIRING-001

> Source citations and ground-truth evidence backing the spec/plan/design/acceptance artifacts.
> Status: draft. All facts below were verified by read-only investigation prior to authoring; cited verbatim so the run-phase agent does not re-discover them.

## R1. CLAUDE.md §7 — Language-Auto-Detect Matrix (canonical for REQ-HDW-003..005)

CLAUDE.md §7 "Language-Specific Guidelines"는 품질 게이트가 프로젝트 언어를 자동 감지하고 해당 toolchain을 실행하도록 규정한다 (canonical 매트릭스):

- **Go**: `go vet` → `golangci-lint` → `go test`
- **Node.js**: `eslint` → `npm test`
- **Python**: `ruff` → `pytest`
- **Rust**: `cargo clippy` → `cargo test`

핵심 의미론 2가지 (REQ-HDW-004/005의 근거):
- "Tools that are not installed are skipped gracefully." → graceful skip (도구 부재 시 exit 0).
- "Projects with no recognized language marker pass the gate silently." → silent pass (마커 없으면 무음 통과).

design.md §1.2의 marker→toolchain 표는 이 매트릭스를 그대로 채택한다.

## R2. Existing settings.json.tmpl Wrapper Pattern (canonical for REQ-HDW-008, design §2)

`internal/template/templates/.claude/settings.json.tmpl`은 모든 hook 엔트리에 platform-conditional quoting 패턴을 사용한다 (실측 lines 9-13, 67-104, 201-211):

```
{{- if eq .Platform "windows"}}
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/<script>.sh\"",
{{- else}}
            "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/<script>.sh\"",
{{- end}}
            "timeout": <N>,
            "type": "command"
```

실측된 기존 블록:
- **PostToolUse** (line 67): `handle-post-tool.sh`, matcher `Write|Edit`, `"async": true`.
- **Stop** (line 84): `handle-stop.sh` + 조건부 `{{ if .HookOptIn.Enabled }} handle-harness-observe-stop.sh {{ end }}`.
- **TaskCompleted** (line 201): `handle-task-completed.sh`, timeout 5s.

→ 신규 두 엔트리는 이 패턴을 verbatim 미러하며 기존 핸들러 배열에 ADD한다 (충돌 없음).

## R3. AGENT-TEAM-REBUILD Deferral Statement (canonical for spec §A "deferred-wiring, NOT bug-fix")

선행 SPEC(`SPEC-V3R6-AGENT-TEAM-REBUILD-001`)의 acceptance.md AC-ATR-009(실측 lines 119-129)는 sync-gate 훅을 **파일 존재 + syntax + 키워드 존재**에만 결합했다:

> Pass criterion: Exit code is 2 (if test fixture present); OR (fallback) `bash -n .claude/hooks/moai/sync-phase-quality-gate.sh; echo "syntax:$?"` returns `syntax:0` AND the script body contains `golangci-lint`, `go test`, and coverage-delta verification keywords.

→ settings.json 등록은 검증 대상이 **아니었다**. 더불어 AC-ATR-014(lines 184-189)는 dependency manifest audit 로직의 **본문 존재**(grep ≥ 2)만 요구했다. 즉 wiring은 의도적으로 미포함.

`--skip-hook` opt-out 플래그는 세 훅 모두에 이미 구현되어 있다 (실측: 각 스크립트 line 10-22 영역) — 이는 훅이 wiring된 이후를 위한 런타임 opt-out 설계로, wiring이 사전 예고되었음을 뒷받침한다.

대조: DB-sync 계열 SPEC은 "SHALL register a PostToolUse matcher in settings.json.tmpl"을 명시 요구했다 (background 제공). 본 영역에서 그 요구의 부재 = 의도적 연기.

→ 결론(spec §A): 본 SPEC은 결함 수정이 아니라 그 예고된 Phase-2 wiring의 실현이다.

## R4. The 3 Discipline Scripts are Self-Contained (canonical for "no new wrapper" — Exclusion 4, design §2.4)

세 스크립트 본문 실측:
- `status-transition-ownership.sh` (87 lines): jq로 stdin 파싱, SPEC-artifact 경로 case 필터, `grep` status 추출, advisory JSON 출력 + audit 로그. **모든 비-`--skip-hook` 경로가 exit 0** (line 86). `moai hook` 미호출.
- `sync-phase-quality-gate.sh` (126 lines): git log로 sync-phase 커밋 감지, `go vet`/`golangci-lint`/`go test`/`go test -cover` 병렬 실행, decision 산출, block 시 exit 2 (line 122-124). Go-하드코딩. `moai hook` 미호출.
- `team-ac-verify.sh` (82 lines): workflow.yaml `team.enabled` capability gate(미설정 시 dormant exit 0), jq AC-ref 추출, advisory 로그. **모든 경로 exit 0** (active 모드도 advisory). `moai hook` 미호출.

→ 세 스크립트는 위임 wrapper가 아닌 최종 핸들러. wiring = `.sh` 직접 가리키는 command 엔트리.

## R5. team-ac-verify Redundancy with Go task_completed.go (canonical for REQ-HDW-007, Exclusion 2)

Go `task_completed.go` 핸들러는 SPEC-AC 검증을 exit-2로 수행한다 (background 제공 사실; hooks-system.md TaskCompleted "Exit 2 = reject completion"). `team-ac-verify.sh`는 (a) team 모드에서만 활성, (b) 활성 시에도 advisory(모든 경로 exit 0, line 60-80), (c) "active verification logic deferred to follow-up SPEC" 주석(line 78) — 즉 stub. → Go 핸들러가 더 강력하고 team-ac-verify.sh는 기능적 중복. 파일 보존 + 미등록 결정의 근거.

## R6. Template-Neutrality CI Guard (canonical for REQ-HDW-009, AC-HDW-006)

`internal/template/internal_content_leak_test.go` + `template_neutrality_audit_test.go`가 5개 forbidden-class를 정규식으로 검사한다 (실측 lines 114-163):

- **C1 (SPEC ID)**: `\bSPEC-(V3R6|AGENCY|WORKTREE)-[A-Z0-9-]+\b`
- **C2 (REQ/AC token)**: `\b(REQ|AC)-(ATR|WO|COORD|UNP|LNC|TII)-[0-9]{3}\b`
- **C3 (Audit citation)**: `Audit [0-9]+ Finding|Audit 3\b`
- **C4 (Finding/archive-date)**: `Finding A[1-6]|archive-202[6-9]-...`
- **C5 (internal path)**: `~/\.claude/projects/-Users-|\.moai/backups/agent-archive-`
- strict 추가: bare ISO date, short-sha, broad REQ/AC.

`template_neutrality_audit_test.go`의 subtest 이름으로 실증 확인한 클래스(런타임 `go test -v` 출력): **C1-macos-bias-path / C2-bare-narrative-v3r / C4-feedback-memory-ref / C5-claude-local-ref / C6-pr-number-ref** (+ C7 SHA / C8 GOOS-preserve / spec-id-date).

→ 편집된 settings.json.tmpl과 sync-gate.sh는 이 **내부-콘텐츠 토큰들**을 포함하지 않아야 한다. 본 SPEC은 내부 SPEC-ID/REQ를 본문에 넣지 않으므로 통과 가능. CLAUDE.local.md §15/§25 정합.

[중요 — D3 정정] 이 CI guard는 **내부 콘텐츠 누출만** 탐지하며 **언어 편향(Go-bias / language-bias) 탐지 클래스는 없다**. 위 클래스 목록(C1-C8)에 language-bias 항목이 부재함이 그 증거이고, 결정적으로 현재 `go vet` 4회 하드코딩에도 guard가 GREEN으로 통과한다. 따라서 **이 guard 통과(AC-HDW-006)는 16개 언어 중립성을 증명하지 않는다**. 이전 표현 "다중-언어 동등 나열(Go-bias 해소)이므로 통과"는 **두 개의 서로 다른 속성을 한 AC로 혼동**한 것이었다. 정정된 검증 분담:
- **AC-HDW-006** = 내부-콘텐츠 누출 부재 (이 CI guard).
- **AC-HDW-009** = Go-bias boundedness (Go-tool 토큰이 Go case branch 내부에만 — `awk` case-block 추출 + `total == inblk`). ← central neutrality risk의 실제 자동 guard.
- **AC-HDW-002** = 런타임 언어-중립성 (실 git-repo non-Go fixture에서 Go toolchain 미실행 + detect_language() 직접호출 `node` 반환).

`.github/workflows/template-neutrality-check.yaml`이 `internal/template/templates/**` 경로 변경 시 이 가드를 트리거한다.

## R7. dev-intent settings.json keys (canonical for REQ-HDW-010, AC-HDW-007)

CLAUDE.local.md §22는 local git-tracked `.claude/settings.json`이 의도적 dev-intent 키를 보유함을 명문화한다: `defaultMode`(bypassPermissions, §22.1), `enableAllProjectMcpServers`(§22.2), `teammateMode`(runtime-managed, §22.3 — settings.local.json에 위치), `env.PATH`(머신 절대경로, §22.4). → wiring은 hook 배열 엔트리만 ADD해야 하며 이 키들을 교란하면 안 된다. git diff로 ADD-only 검증.

## R8. Hook System Reference (canonical for timeouts, design §2.2-2.3)

`.claude/rules/moai/core/hooks-system.md` timeout 표: SessionStart/PreCompact/PreToolUse/PostToolUse=5s, PostCompact/Stop=10s, max 600s. → status-transition(PostToolUse, 가벼움)=5s, sync-gate(Stop, 테스트 포함)=10s 정합. PostToolUse는 "Can Block: No" — status-transition advisory와 정합. Stop은 "Can Block: Yes"이나 본 SPEC은 warn-first로 차단 미활성.

`.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface는 세 discipline 훅의 trigger/owning-REQ/exit-code 의미론을 표로 정의 — status-transition(PostToolUse, exit0=continue/exit2=block), sync-gate(Stop, exit0/exit2), team-ac(TaskCompleted, dormant). 이 표는 본 SPEC의 wiring 대상과 advisory/warn-first 결정의 상위 근거.

## R9. Verified File Inventory

| File | Local | Template mirror | Size | Note |
|------|-------|-----------------|------|------|
| status-transition-ownership.sh | ✓ | ✓ | 3625 B | byte-identical, static .sh |
| sync-phase-quality-gate.sh | ✓ | ✓ | 4379 B | byte-identical, Go-hardcoded (M1 target) |
| team-ac-verify.sh | ✓ | ✓ | 3066 B | byte-identical, EXCLUDED from wiring |
| settings.json.tmpl | — | ✓ | — | has PostToolUse/Stop/TaskCompleted blocks (M2 target) |
| .claude/settings.json | ✓ (git-tracked) | — | — | dev-intent keys; local mirror target |

## R10. Cross-References

- CLAUDE.md §7 (canonical language matrix)
- `.claude/rules/moai/core/hooks-system.md`, `agent-common-protocol.md` § Hook Invocation Surface
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (status-transition 훅이 참조)
- CLAUDE.local.md §2/§15/§22/§25
- `internal/template/internal_content_leak_test.go`, `template_neutrality_audit_test.go`, `.github/workflows/template-neutrality-check.yaml`
