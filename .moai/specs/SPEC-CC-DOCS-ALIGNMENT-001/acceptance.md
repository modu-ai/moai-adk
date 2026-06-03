# SPEC-CC-DOCS-ALIGNMENT-001 — Acceptance Criteria

## §A — Verification Conventions

- 경로 약어: `R=` 배포 mirror(`.claude/rules/moai/...`), `T=` template source(`internal/template/templates/.claude/rules/moai/...`).
- **Template-managed AC는 source(T)와 mirror(R) 양쪽에 대해 동일 단언을 적용**한다(B.1 부류). 단, neutrality split 파일(orchestration-mode-selection.md / agent-authoring.md / CLAUDE.md)은 byte-parity를 단언하지 않고 "정정된 내용이 양쪽에 각각 반영"되었는지만 검증한다(C3).
- **LOCAL-ONLY AC(REQ-017)는 mirror 단언을 적용하지 않는다**(template 트리에 부재).
- grep 예시의 라인 번호는 plan-phase 검증 시점 값이며, run-phase는 literal 앵커로 위치를 재확인한다(C5).
- 공통 게이트(모든 마일스톤): `make build` 후 `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS + `go test ./internal/template/... -run TestInternalContentLeak`(또는 `internal_content_leak_test.go`의 실제 테스트명) PASS.

---

## §B — M1 workflows ACs

### AC-CDA-001 (REQ-CDA-001, high)
Given `dynamic-workflows.md` § Disabling Workflows,
When 정정 후,
Then `ultracode` 트리거 키워드가 명시되고 "no longer triggers"의 stale 표현이 정정된다.
- grep: `grep -c "no longer triggers a run" R/workflow/dynamic-workflows.md` → 정정 후 해당 stale 문구가 `ultracode` 맥락으로 대체(원 문장 0 또는 ultracode 동반).
- grep: `grep -c "ultracode" R/workflow/dynamic-workflows.md` ≥ 2 (기존 1 + § Disabling Workflows 신규).
- parity: 동일 grep을 `T/workflow/dynamic-workflows.md`에 적용 → 동일 결과(byte-IDENTICAL 파일).

### AC-CDA-002 (REQ-CDA-002, high)
Given workflow 트리거 의미,
When 정정 후,
Then `ultracode`(또는 "use a workflow")가 per-prompt 트리거이며 세션 전역 `/effort ultracode`와 구별됨이 문서화된다.
- grep: `grep -niE "per-prompt|per prompt" R/workflow/dynamic-workflows.md` ≥ 1 AND 같은 문맥에 `ultracode` 포함.
- parity(T): 동일.

### AC-CDA-003 (REQ-CDA-003, med)
Given `/workflows` TUI,
When 정정 후,
Then "Manage runs" 하위 절이 존재하고 list/watch/pause/resume/save + 키 바인딩(p/x/s/r)을 기술.
- grep: `grep -c "Manage runs" R/workflow/dynamic-workflows.md` ≥ 1.
- grep: TUI 동작 키 `grep -niE "pause|resume|watch" R/workflow/dynamic-workflows.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-004 (REQ-CDA-004, med)
Given Saved-workflow `args`,
When 정정 후,
Then § Saved workflows 아래 `args` global input 항목이 존재.
- grep: `grep -niE "\bargs\b" R/workflow/dynamic-workflows.md` ≥ 1 (Saved workflows 문맥).
- parity(T): 동일.

### AC-CDA-005 (REQ-CDA-005, low)
Given plan/provider 가용성,
When 정정 후,
Then 유료 플랜 + API/Bedrock/Vertex/Foundry + Pro(`/config`) 노트 존재.
- grep: `grep -niE "Bedrock|Vertex|Foundry" R/workflow/dynamic-workflows.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-006 (REQ-CDA-006, low)
Given per-run approval / permission-mode 매트릭스,
When 정정 후,
Then Default/accept-edits=prompt-every-run, Auto=first-launch, Bypass/`-p`/SDK=never 노트 존재(GATE-2 연계).
- grep: `grep -niE "per-run|prompt every run|first.launch" R/workflow/dynamic-workflows.md` ≥ 1.
- parity(T): 동일.

---

## §C — M2 skills ACs

### AC-CDA-007 (REQ-CDA-007, high)
Given `skill-writing-craft.md` schema 테이블 + 4 example 블록,
When 정정 후,
Then `type: skill` / `type` 필드가 schema·example 영역에서 모두 제거된다.
- grep: `grep -c "type: skill" R/development/skill-writing-craft.md` == 0.
- grep: schema 테이블 행 `grep -nE "^\| \`type\`" R/development/skill-writing-craft.md` == 0.
- parity(T): `grep -c "type: skill" T/development/skill-writing-craft.md` == 0 (byte-IDENTICAL).

### AC-CDA-008 (REQ-CDA-008, high)
Given description 길이 한계 상호 모순,
When 정정 후,
Then 1,536 hard cap으로 통일되고 ≤80/under-250 heuristic이 제거 또는 완화된다.
- grep: `grep -c "1,536\|1536" R/development/skill-authoring.md` ≥ 1 (cap 존재).
- grep: `grep -c "≤80\|80 char" R/development/skill-writing-craft.md` == 0 (heuristic 제거).
- grep: `grep -c "under 250\|max 1024" R/development/skill-authoring.md` == 0 (모순 표현 제거).
- parity(T): 두 파일 모두 byte-IDENTICAL → 동일 결과.

### AC-CDA-009 (REQ-CDA-009, med)
Given `skillOverrides` settings 키,
When 정정 후,
Then § Skill Invocation Control에 4 상태(on/name-only/user-invocable-only/off) + plugin skills 미영향 노트 존재.
- grep: `grep -c "skillOverrides" R/development/skill-authoring.md` ≥ 1.
- grep: `grep -niE "name-only|user-invocable-only" R/development/skill-authoring.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-010 (REQ-CDA-010, med)
Given `name` 필드 required 표기,
When 정정 후,
Then `name`이 optional(디렉터리명 기본값)로 재분류되고 `moai-{category}-{name}` 권고 유지.
- grep: skill-authoring.md 라인 16 영역 `name` 항목이 optional 표현 포함 — `grep -niE "name.*optional|optional.*directory name" R/development/skill-authoring.md` ≥ 1.
- grep: `grep -c "moai-{category}-{name}\|moai-{category}" R/development/skill-authoring.md` ≥ 1 (권고 보존).
- parity(T): 동일(byte-IDENTICAL).

### AC-CDA-011 (REQ-CDA-011, med)
Given nested/monorepo discovery + `--add-dir`,
When 정정 후,
Then § Skill Scope and Priority에 discovery 노트(parent-walk, nested, `--add-dir` vs `permissions.additionalDirectories`) 존재.
- grep: `grep -c -- "--add-dir" R/development/skill-authoring.md` ≥ 1.
- grep: `grep -niE "nested|parent-walk|monorepo" R/development/skill-authoring.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-012 (REQ-CDA-012, low)
Given reference mirror `claude-code-skills-official.md`,
When 정정 후,
Then `/run` `/verify` `/run-skill-generator` 번들 skills(v2.1.145+)가 포함된다.
- grep: `grep -c "run-skill-generator" R/../skills/moai-foundation-cc/reference/claude-code-skills-official.md` ≥ 1.
- parity(T): 동일(byte-IDENTICAL — 편집 후에도 양쪽 동일).

### AC-CDA-013 (REQ-CDA-013, low)
Given `coding-standards.md` § Thin Command Pattern,
When 정정 후,
Then "custom commands가 skills로 병합(skill 우선)" 1줄 노트 존재.
- grep: `grep -niE "merged into skill|command.*skill.*win|skill wins" R/development/coding-standards.md` ≥ 1.
- parity(T): 동일(byte-IDENTICAL).

### AC-CDA-014 (REQ-CDA-014, low)
Given `CLAUDE.md`(427-434) + `skill-authoring.md`(131-139) §13 Level 2,
When 정정 후,
Then "description은 항상 listing" vs "body는 invocation 시 로드·유지"로 구분된다.
- grep: `grep -niE "always listed|description.*listed" R(=CLAUDE.md)` ≥ 1 AND skill-authoring.md Level 2 영역에 invocation/listing 구분 표현.
- parity: CLAUDE.md는 split(C3) — source(T)에도 정정 반영되되 byte-parity 강제 안 함. skill-authoring.md는 byte-IDENTICAL → 양쪽 동일.

---

## §D — M3 hooks DOC ACs

### AC-CDA-015 (REQ-CDA-015, med)
Given `hooks-system.md` `if` 필드 버전 자기모순(라인 191 v2.1.84 vs 라인 170 v2.1.85),
When 정정 후,
Then `if` 필드 버전이 v2.1.85+로 통일되고 파일 내부가 일관된다.
- grep: `if` 필드/permission-rule filter 영역에서 `grep -nE "v2\.1\.84" R/core/hooks-system.md` → `if` 문맥에서 0 (v2.1.85로 정정).
- grep: `grep -c "v2.1.85" R/core/hooks-system.md` ≥ 1.
- parity(T): 동일(byte-IDENTICAL).

### AC-CDA-016 (REQ-CDA-016, low)
Given Stop-hook block-cap,
When 정정 후,
Then Stop 절에 8-연속-block override + `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 1줄 존재.
- grep: `grep -c "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP" R/core/hooks-system.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-017 (REQ-CDA-017, low) — LOCAL-ONLY (mirror 단언 없음)
Given `CLAUDE.local.md`(라인 386) + `internal/hook/CLAUDE.md`(라인 7/16)의 "default 5s",
When 정정 후,
Then "MoAI policy default 5s (platform default 10 min)"으로 재기술된다.
- grep: `grep -niE "platform default.*10 min|10 min" CLAUDE.local.md` ≥ 1.
- grep: `grep -niE "platform default.*10 min|10 min" internal/hook/CLAUDE.md` ≥ 1.
- **mirror 단언 없음**: 두 파일은 template 트리에 부재 → `T` 경로 검증 SKIP(정상).

### AC-CDA-018 (REQ-CDA-018, low)
Given exec form troubleshooting,
When 정정 후,
Then `hooks-system.md`에 `"args": []` exec form 노트 + `if [[ $- == *i* ]]` interactive-shell guard 존재.
- grep: `grep -niE "exec form|\"args\": \[\]" R/core/hooks-system.md` ≥ 1.
- grep: `grep -c -- 'if \[\[ $- == \*i\* \]\]' R/core/hooks-system.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-019 (REQ-CDA-019, low)
Given `agent-common-protocol.md` Stop==sync 가정,
When 정정 후,
Then Stop self-gate caveat("매 turn-end 발화, task 완료 시점만 아님, user interrupt 시 미발화") 존재.
- grep: `grep -niE "every turn.end|not only.*completion|user interrupt" R/core/agent-common-protocol.md` ≥ 1.
- parity(T): 동일(편집 영역 중립 → 양쪽 동일).

---

## §E — M4 goal ACs

### AC-CDA-020 (REQ-CDA-020, med)
Given 비교 테이블 2행이 native `/loop`을 `/moai loop`으로 치환,
When 정정 후,
Then native `/loop`(time-interval)이 별도 행/노트로 추가되고 `/moai loop`(diagnostic) 행은 유지된다.
- grep: `grep -c "/moai loop" R/workflow/goal-directive.md` ≥ 1 (유지).
- grep: time-interval native loop 표현 `grep -niE "time interval|time-interval" R/workflow/goal-directive.md` ≥ 1.
- parity(T): 동일(byte-IDENTICAL).

### AC-CDA-021 (REQ-CDA-021, med)
Given auto mode(per-tool) 상보성,
When 정정 후,
Then auto mode가 `/goal`(per-turn)과 unattended `ac_converge` loop로 짝을 이룸이 MoAI Integration 노트로 추가.
- grep: `grep -niE "auto mode|ac_converge" R/workflow/goal-directive.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-022 (REQ-CDA-022, low)
Given non-interactive surfaces,
When 정정 후,
Then desktop app + Remote Control이 `-p`에 추가된다.
- grep: `grep -niE "Remote Control|desktop app" R/workflow/goal-directive.md` ≥ 1.
- parity(T): 동일.

### AC-CDA-023 (REQ-CDA-023, low)
Given `◎ /goal active` 인디케이터,
When 정정 후,
Then 인디케이터 + per-turn evaluator reason + bare-`/goal` turns/tokens 표시 1문장 존재.
- grep: `grep -c "/goal active" R/workflow/goal-directive.md` ≥ 1 (또는 indicator 표현).
- parity(T): 동일.

### AC-CDA-024 (REQ-CDA-024, low)
Given disable scope 부정확,
When 정정 후,
Then `disableAllHooks`(any level) / `allowManagedHooksOnly`(managed only) per-flag scope 존재.
- grep: `grep -c "allowManagedHooksOnly" R/workflow/goal-directive.md` ≥ 1 AND `disableAllHooks` 동반.
- parity(T): 동일.

### AC-CDA-025 (REQ-CDA-025, low)
Given evaluator token cost,
When 정정 후,
Then small fast model + negligible + 세션 provider(GLM 포함) 실행 cost 1줄 노트 존재.
- grep: `grep -niE "evaluator.*cost|negligible|fast model" R/workflow/goal-directive.md` ≥ 1.
- parity(T): 동일.

---

## §F — M5 sub-agents ACs

### AC-CDA-026 (REQ-CDA-026, high)
Given `agent-authoring.md` 라인 54 maxTurns deprecated + 부재 필드 maxContextSize,
When 정정 후,
Then `maxContextSize` 토큰이 제거되고 `maxTurns`가 current optional 필드로 복원된다.
- grep: `grep -c "maxContextSize" R/development/agent-authoring.md` == 0.
- grep: `grep -niE "maxTurns.*deprecat" R/development/agent-authoring.md` == 0 (deprecation 노트 제거).
- grep: `grep -c "maxTurns" R/development/agent-authoring.md` ≥ 1 (current 필드 유지).
- parity: agent-authoring.md는 split(C3) → source(T)에도 `grep -c "maxContextSize" T/...` == 0 (정정 양쪽 반영, byte-parity 강제 안 함).

### AC-CDA-027 (REQ-CDA-027, med)
Given `SubagentStart` hook event 미문서화,
When 정정 후,
Then `agent-authoring.md` hook-event 리스트 + `agent-hooks.md` 테이블에 `SubagentStart`(agent-type 매칭) 추가.
- grep: `grep -c "SubagentStart" R/development/agent-authoring.md` ≥ 1.
- grep: `grep -c "SubagentStart" R/core/agent-hooks.md` ≥ 1.
- parity: agent-hooks.md byte-IDENTICAL → T 동일; agent-authoring.md split → T에 `SubagentStart` ≥ 1.

### AC-CDA-028 (REQ-CDA-028, med)
Given fork subagents 전면 미문서화,
When 정정 후,
Then fork 노트(`/fork`, `CLAUDE_CODE_FORK_SUBAGENT`, v2.1.117+, experimental, parent context 상속, nest 불가)가 agent-authoring.md 또는 agent-patterns.md에 추가.
- grep: `grep -rc "CLAUDE_CODE_FORK_SUBAGENT" R/development/agent-authoring.md R/development/agent-patterns.md` 합 ≥ 1.
- grep: `grep -rniE "cannot nest|nest" <편집 대상 파일>` ≥ 1 (fork nest 제약).
- parity(T): 편집한 파일의 source에 동일 grep ≥ 1.

### AC-CDA-029 (REQ-CDA-029, med)
Given `claude-code-guide`를 archived로 취급(archived-agent-rejection.md 84/131 + CLAUDE.md 127),
When 정정 후,
Then MoAI-archived custom 파일과 동명 공식 built-in을 구분하는 disambiguation 노트가 존재하고, built-in이 rejection 미트리거임이 명시된다.
- grep: `grep -niE "built-in|built in" R/workflow/archived-agent-rejection.md` 문맥에 `claude-code-guide` 동반 ≥ 1.
- 회귀 방지: `grep -c "claude-code-guide" R/workflow/archived-agent-rejection.md` ≥ 2 (기존 archived 이력 보존 + 구분 노트 추가, 삭제 아님).
- parity: archived-agent-rejection.md byte-IDENTICAL → T 동일; CLAUDE.md split → 구분 노트가 양쪽(R/T)에 반영(byte-parity 강제 안 함).

### AC-CDA-030 (REQ-CDA-030, low)
Given `model` full-ID form 공식 허용 vs MoAI 금지,
When 정정 후,
Then `model-policy.md`에 full-ID form이 "official-but-intentionally-disallowed"임이 명시(([1m] entitlement 근거 인용)).
- grep: `grep -niE "official.but.intentionally|intentionally.disallow" R/development/model-policy.md` ≥ 1.
- grep: `grep -c "claude-opus-4-8\|full-ID\|full ID" R/development/model-policy.md` ≥ 1.
- parity(T): 동일(byte-IDENTICAL).

### AC-CDA-031 (REQ-CDA-031, low)
Given `agent-authoring.md`(105/185) Permission Modes에 비공식 `delegate`,
When 정정 후,
Then `delegate`가 canonical Frontmatter enum(라인 185)에서 빠지고 "MoAI experimental extension" 노트로 표기된다.
- grep: 라인 185 enum 영역 `grep -nE "permissionMode.*delegate|delegate.*bypassPermissions" R/development/agent-authoring.md` → canonical enum 행에서 `delegate` 부재.
- grep: `grep -niE "experimental.*extension|MoAI.*experimental" R/development/agent-authoring.md` ≥ 1 (delegate 문맥).
- parity: split → T에도 동일 정정 반영.

### AC-CDA-032 (REQ-CDA-032, low)
Given `name` 필드 의미 미포착(라인 48),
When 정정 후,
Then filename 불일치 허용 + hook `agent_type` 수신 두 명확화가 추가된다.
- grep: `grep -niE "agent_type|filename.*not.*match|need not match" R/development/agent-authoring.md` ≥ 1.
- parity: split → T 동일 정정 반영.

### AC-CDA-033 (REQ-CDA-033, low)
Given managed-settings scope 부재(라인 11-25),
When 정정 후,
Then org-wide priority 1 + `.claude/agents/` override precedence 1줄 노트 존재.
- grep: `grep -niE "managed.settings|org-wide|priority 1|precedence" R/development/agent-authoring.md` ≥ 1.
- parity: split → T 동일 정정 반영.

---

## §G — Quality Gate Criteria (전체)

- [ ] 33개 AC 전부 grep 단언 PASS.
- [ ] Template-managed 파일(B.1): source(T) + mirror(R) 양쪽에 정정 반영(byte-IDENTICAL 파일은 동일, split 파일은 각각).
- [ ] LOCAL-ONLY 파일(REQ-017): mirror 단언 미적용, 직접 편집만 검증.
- [ ] `make build` 실행 완료(template 변경 후).
- [ ] `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS.
- [ ] internal content leak 테스트 PASS(`internal_content_leak_test.go`).
- [ ] 규칙 본문에 내부 SPEC ID/REQ 토큰 미임베드(§25/C2) — `grep -rc "SPEC-CC-DOCS-ALIGNMENT" T/` == 0.
- [ ] C3 split 보존: orchestration-mode-selection.md / agent-authoring.md / CLAUDE.md의 source는 중립, mirror는 dev-local 표현(둘 다 정정 반영, byte-identical 아님).
- [ ] Go 소스 무변경(C4/Exclusion §E.1): `git diff --name-only`에 `internal/**/*.go`(embedded 재생성 제외) 미포함.

## §H — Definition of Done

1. §B~§F 33개 AC 전부 PASS.
2. §G Quality Gate 전체 충족.
3. spec.md §E Exclusions 준수(Go hook 결함 미포함, 공식 문서 미수정, 신규 규칙 파일 미생성, CHANGELOG/README/docs-site 미변경).
4. sibling SPEC(SPEC-HOOK-EVENT-REGISTRY-001)과 범위 충돌 없음(Go EventType/CoverageTable 미터치).
