# acceptance.md — SPEC-CC2178-DOCS-ALIGN-001

> **Acceptance criteria** for CC 2.1.169→2.1.178 Tier 1 docs-only 정합화. 8 functional AC + 2 cross-cutting AC (4-locale parity + no-Go-code). Tier S.

## §D — AC Matrix

### AC-DA-001 — `Tool(param:value)` wildcard permission syntax documented [MUST]

**Traces to**: REQ-DA-001

**Given** CC 2.1.178이 `Tool(param:value)` 형식의 `*` wildcard permission-rule syntax를 도입했고
**When** run-phase가 `.claude/rules/moai/core/settings-management.md` § Permission Management와 docs-site `advanced/settings-json.md`(ko/en/ja/zh 4-locale)의 해당 영역을 읽으면
**Then** `Tool(param:value)` wildcard 문법이 예시와 함께 문서화되어 있어야 하며, MoAI가 현재 param-scoped 규칙을 emit하지 않으나 "옵션으로 존재함"이 명시되어야 한다.

**Severity**: MUST (high). REQ-DA-001 priority high.

**Verification (observable)**:
```bash
# rules 파일 (template source + mirror)
grep -c "Tool(param:value)" internal/template/templates/.claude/rules/moai/core/settings-management.md
grep -c "Tool(param:value)" .claude/rules/moai/core/settings-management.md
# docs-site 4-locale
for loc in en ko ja zh; do
  grep -c "Tool(param:value)" docs-site/content/$loc/advanced/settings-json.md
done
# 각 카운트 ≥ 1
```

---

### AC-DA-002 — Nested `.claude/` closest-wins precedence documented [MUST]

**Traces to**: REQ-DA-002

**Given** CC 2.1.178이 nested `.claude/` 디렉터리에서 이름 충돌 시 closest-directory-wins 선행 규칙을 도입했고(agent/workflow/output-style)
**When** run-phase가 `skill-authoring.md`, `agent-authoring.md`, docs-site `advanced/skill-guide.md`, `advanced/agent-guide.md`(4-locale)의 해당 절을 읽으면
**Then** closest-directory-wins 선행 규칙이 문서화되어 있어야 한다.

**Severity**: MUST (high). REQ-DA-002 priority high.

**Verification**:
```bash
for f in internal/template/templates/.claude/rules/moai/development/skill-authoring.md \
         internal/template/templates/.claude/rules/moai/development/agent-authoring.md; do
  grep -ciE "closest.wins|closest-directory" "$f"
done
for loc in en ko ja zh; do
  grep -ciE "closest.wins|closest-directory" docs-site/content/$loc/advanced/skill-guide.md
  grep -ciE "closest.wins|closest-directory" docs-site/content/$loc/advanced/agent-guide.md
done
# 각 카운트 ≥ 1
```

---

### AC-DA-003 — Nested `.claude/skills` loading documented [SHOULD]

**Traces to**: REQ-DA-003

**Given** CC 2.1.178이 nested `.claude/skills` 디렉터리 작업 시 skills를 로드함
**When** run-phase가 `skill-authoring.md` § Skill Scope and Priority와 docs-site `advanced/skill-guide.md`(4-locale)를 읽으면
**Then** nested `.claude/skills` 디렉터리 로딩 시맨틱이 발견 노트로 문서화되어 있어야 한다.

**Severity**: SHOULD (med). REQ-DA-003 priority med.

**Verification**:
```bash
grep -ciE "nested.*\.claude/skills|skills.*load.*nested" \
  internal/template/templates/.claude/rules/moai/development/skill-authoring.md
for loc in en ko ja zh; do
  grep -ciE "nested.*\.claude/skills|skills.*load.*nested" docs-site/content/$loc/advanced/skill-guide.md
done
```

---

### AC-DA-004 — `disableBundledSkills` documented [SHOULD]

**Traces to**: REQ-DA-004

**Given** CC 2.1.169가 `disableBundledSkills` 설정 + 환경변수(번들 skills/workflows 숨김)를 도입했고
**When** run-phase가 `settings-management.md`, `skill-authoring.md`, docs-site `advanced/settings-json.md`, `advanced/skill-guide.md`(4-locale)를 읽으면
**Then** `disableBundledSkills` 토글이 문서화되어 있어야 하며, 번들 `/deep-research` 등을 비활성화함이 명시되고, MoAI가 이를 emit하지 않음이 옵션 노트로 기록되어야 한다.

**Severity**: SHOULD (med). REQ-DA-004 priority med. Accuracy-over-completeness: "MoAI 미연결" 명시 의무.

**Verification**:
```bash
# REQ-DA-004는 두 토글을 모두 커버 (CC 2.1.169 governance/debug toggles folded together):
#   (a) disableBundledSkills  (b) --safe-mode
# 두 토글 모두 각 파일에서 ≥ 1 매치여야 AC PASS.
for f in internal/template/templates/.claude/rules/moai/core/settings-management.md \
         internal/template/templates/.claude/rules/moai/development/skill-authoring.md; do
  echo "== $f =="
  grep -c "disableBundledSkills" "$f"        # ≥ 1
  grep -ciE "safe-mode|--safe-mode" "$f"     # ≥ 1
done
for loc in en ko ja zh; do
  echo "== $loc/settings-json.md =="
  grep -c "disableBundledSkills" docs-site/content/$loc/advanced/settings-json.md    # ≥ 1
  grep -ciE "safe-mode|--safe-mode" docs-site/content/$loc/advanced/settings-json.md # ≥ 1
  echo "== $loc/skill-guide.md =="
  grep -c "disableBundledSkills" docs-site/content/$loc/advanced/skill-guide.md      # ≥ 1
  grep -ciE "safe-mode|--safe-mode" docs-site/content/$loc/advanced/skill-guide.md   # ≥ 1
done
```

---

### AC-DA-005 — `post-session` lifecycle hook documented [SHOULD]

**Traces to**: REQ-DA-005

**Given** CC 2.1.169가 `post-session` 라이프사이클 훅(self-hosted runner, 세션 종료 후 발화)을 도입했고
**When** run-phase가 `hooks-system.md` § Hook Events와 docs-site `advanced/hooks-reference.md`, `advanced/hooks-guide.md`(4-locale)를 읽으면
**Then** `post-session` 이벤트가 문서화되어 있어야 하며, MoAI가 현재 이 훅을 wiring하지 않음이 "존재와 옵션" framing으로 명시되어야 한다.

**Severity**: SHOULD (med). REQ-DA-005 priority med. Accuracy-over-completeness 핵심 케이스.

**Verification**:
```bash
grep -c "post-session" internal/template/templates/.claude/rules/moai/core/hooks-system.md
for loc in en ko ja zh; do
  grep -c "post-session" docs-site/content/$loc/advanced/hooks-reference.md
  grep -c "post-session" docs-site/content/$loc/advanced/hooks-guide.md
done
# 각 카운트 ≥ 1
```

---

### AC-DA-006 — subagent `disallowedTools` MCP enforcement documented [SHOULD]

**Traces to**: REQ-DA-006

**Given** CC 2.1.178이 subagent `disallowedTools`의 MCP server-level specs을 강제로 변경했고(이전 silent-ignore)
**When** run-phase가 `agent-authoring.md` § Frontmatter Format Rules와 docs-site `advanced/agent-guide.md`(4-locale)를 읽으면
**Then** `disallowedTools`가 MCP 서버 수준에서 강제됨이 노트로 문서화되어 있어야 한다.

**Severity**: SHOULD (low). REQ-DA-006 priority low.

**Verification**:
```bash
grep -ciE "disallowedTools.*MCP|MCP.*disallowedTools|disallowedTools.*enforc" \
  internal/template/templates/.claude/rules/moai/development/agent-authoring.md
for loc in en ko ja zh; do
  grep -ciE "disallowedTools.*MCP|MCP.*disallowedTools|disallowedTools.*enforc" \
    docs-site/content/$loc/advanced/agent-guide.md
done
```

---

### AC-DA-007 — auto-mode pre-launch classifier documented [SHOULD]

**Traces to**: REQ-DA-007

**Given** CC 2.1.178이 auto-mode의 subagent spawn을 classifier가 launch 전에 평가하도록 개선했고
**When** run-phase가 `orchestration-mode-selection.md` §B 결정 트리를 읽으면
**Then** auto-mode의 pre-launch classifier 게이트가 노트로 문서화되어 있어야 하며, `/goal` unattended 루프와의 상보성이 언급되어야 한다.

**Severity**: SHOULD (low). REQ-DA-007 priority low. rules-only (docs-site auto-mode는 이미 goal-directive.md에 연결됨).

**Verification**:
```bash
grep -ciE "pre.launch.classifier|classifier.*launch|auto.mode.*classifier" \
  internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md
# 카운트 ≥ 1
```

---

### AC-DA-008 — `/cd` cache-preserving resume documented [MUST]

**Traces to**: REQ-DA-008

**Given** CC 2.1.169가 `/cd` 명령(프롬프트 캐시 보존하며 cwd 이동)을 도입했고, 이는 새 터미널(Block 0)의 cold-start 대안이며
**When** run-phase가 `session-handoff.md` § Worktree-Anchored Resume Pattern Block 0과 docs-site `advanced/statusline.md`, `workflow-commands/moai-sync.md`(4-locale)를 읽으면
**Then** `/cd`가 cache-preserving 재개 대안으로 노트되어 있어야 하며, 기존 Block 0 new-terminal 경로를 대체하지 않고 보완함이 명시되어야 한다.

**Severity**: MUST (med). REQ-DA-008 priority med. Block 0 기존 계약 보존이 핵심.

**Verification**:
```bash
grep -ciE "/cd.*cache|cache.preserv.*/cd|/cd.*prompt cache" \
  internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
for loc in en ko ja zh; do
  grep -ciE "/cd.*cache|cache.preserv.*/cd|/cd.*prompt cache" \
    docs-site/content/$loc/advanced/statusline.md
  grep -ciE "/cd.*cache|cache.preserv.*/cd|/cd.*prompt cache" \
    docs-site/content/$loc/workflow-commands/moai-sync.md
done
# session-handoff.md 카운트 ≥ 1; docs-site는 페이지에 따라 ≥ 0 허용(statusline/moai-sync 중 최소 1곳)
```

---

### AC-DA-009 — 4-locale parity MANDATORY [MUST]

**Traces to**: Constraint C1, `.moai/docs/docs-site-i18n-rules.md` §17.3 [HARD]

**Given** docs-site content 변경 시 ko/en/ja/zh 4-locale 동시 반영이 의무이며
**When** run-phase가 4-locale 편집을 완료하면
**Then** `scripts/docs-i18n-check.sh`가 exit 0을 반환해야 하며, 4-locale 간 파일 수/경로 일치가 확인되어야 한다.

**Severity**: MUST. CI 강제 게이트.

**Verification**:
```bash
# pre-commit + CI 게이트
bash scripts/docs-i18n-check.sh && echo "PASS: 4-locale parity"

# 추가 수동 검증: 각 대상 페이지의 target feature가 4 locale에 모두 존재
for page in advanced/settings-json.md advanced/hooks-reference.md advanced/hooks-guide.md \
            advanced/skill-guide.md advanced/agent-guide.md advanced/statusline.md; do
  for loc in en ko ja zh; do
    test -f docs-site/content/$loc/$page || echo "MISSING: $loc/$page"
  done
done
```

---

### AC-DA-010 — No Go code changed (ZERO) [MUST]

**Traces to**: Constraint C4, spec.md §E #1

**Given** 본 SPEC은 순수 docs/rules markdown 정합화이며 Go 소스를 일절 수정하지 않으며
**When** run-phase가 모든 커밋의 diff를 검토하면
**Then** 어떤 `.go` 파일도 변경되지 않아야 한다(`internal/template/embedded.go` 재생성은 `make build` 산물이며 소스 변경 아님).

**Severity**: MUST. 본 SPEC의 존재 이유.

**Verification**:
```bash
# 본 SPEC의 모든 run-phase 커밋에서 .go 파일이 0개
# <plan-phase-commit>은 run-phase fill-in 토큰이다 — run-phase 진입 시
# plan-phase 마지막 커밋 SHA로 치환 (예: abc1234). plan-phase 시점에는
# 아직 run-phase 커밋이 없으므로 placeholder로 둔다.
git log --name-only --format="" <plan-phase-commit>..HEAD \
  | grep -E '\.go$' \
  | grep -v '^internal/template/embedded.go$' \
  | grep -v '^internal/template/templates_embedded\.go$' \
  || echo "PASS: 0 Go source files changed (embedded.go regeneration excluded as build artifact)"
```

---

## §D.1 — Severity Summary

| AC | Severity | Milestone | REQ |
|----|----------|-----------|-----|
| AC-DA-001 | MUST (high) | M1 | REQ-DA-001 |
| AC-DA-002 | MUST (high) | M1 | REQ-DA-002 |
| AC-DA-003 | SHOULD (med) | M1 | REQ-DA-003 |
| AC-DA-004 | SHOULD (med) | M1 | REQ-DA-004 |
| AC-DA-005 | SHOULD (med) | M2 | REQ-DA-005 |
| AC-DA-006 | SHOULD (low) | M2 | REQ-DA-006 |
| AC-DA-007 | SHOULD (low) | M2 | REQ-DA-007 |
| AC-DA-008 | MUST (med) | M3 | REQ-DA-008 |
| AC-DA-009 | MUST | cross | C1 (4-locale parity) |
| AC-DA-010 | MUST | cross | C4 (no Go code) |

**MUST count**: 5 (AC-DA-001, 002, 008, 009, 010). **SHOULD count**: 5.

## §D.2 — Indirect Verification (run-phase evidence layer)

run-phase(manager-develop) §E self-verification은 다음을 포함해야 한다:

1. **AC binary matrix** — 10 AC 각각 PASS/FAIL
2. **4-locale parity screenshot/grep evidence** — `scripts/docs-i18n-check.sh` exit 0 출력
3. **Zero Go code evidence** — `git diff --name-only | grep '\.go$'` 빈 출력
4. **Template mirror parity** — `make build` 후 `diff -q source mirror` 결과 (의도적 DIFFER 파일 제외)
5. **Neutrality CI** — `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS
6. **spec-lint** — `moai spec lint .moai/specs/SPEC-CC2178-DOCS-ALIGN-001/` clean

## §D.3 — Closure Gates

**Definition of Done (Tier S docs-only)**:
- [ ] 10 AC 전부 PASS (5 MUST + 5 SHOULD; SHOULD는 PASS 권장, 단 blocker 시 defer 가능)
- [ ] 4-locale parity CI 통과 (AC-DA-009)
- [ ] Zero Go code changed (AC-DA-010)
- [ ] `make build` 성공 + template mirror 동기화
- [ ] neutrality CI (`template-neutrality-check.yaml`) 통과
- [ ] spec-lint clean
- [ ] Conventional Commits 준수 (`feat(SPEC-CC2178-DOCS-ALIGN-001): M{N} ...`)

## §D.4 — Forward-Looking Checks (sync-phase + Mx-phase)

- **CHANGELOG**: sync-phase(manager-docs)가 CHANGELOG.md에 본 SPEC의 docs-site 4-locale + rules 정합화 내용을 추가. run-phase에서는 다루지 않음.
- **status transition**: `draft → in-progress`(M1 첫 커밋 시 manager-develop) → `implemented`(sync 커밋 시 manager-docs) → `completed`(Mx chore).
- **sibling 동기화**: MODEL-POLICY-REPAIR-001과 동일 CC 창에서 분리되었으나 독립 진행. 본 SPEC이 먼저 완료되어도 MODEL-POLICY-REPAIR-001에 dependency 없음.
