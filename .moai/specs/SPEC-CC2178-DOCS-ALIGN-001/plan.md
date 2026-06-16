# plan.md — SPEC-CC2178-DOCS-ALIGN-001

> **Implementation plan** for CC 2.1.169→2.1.178 Tier 1 docs-only 정합화. Tier S, docs-only, 3 milestones. ZERO Go code.

## §A — Context

- **SPEC**: `.moai/specs/SPEC-CC2178-DOCS-ALIGN-001/spec.md` (212 lines)
- **Research**: `.moai/research/cc-update-2.1.163-to-2.1.178.md` P4 + P5 (lines 192-200) — falsifiable evidence layer
- **Predecessor pattern**: `.moai/specs/SPEC-CC-DOCS-ALIGNMENT-001/` (completed, 33 defects, same approach)
- **Tier**: S (docs-only, < 5 Go files = 0 Go files, 3 milestones)
- **cycle_type**: n/a (docs-only, manager-develop `cycle_type=tdd`은 Go 코드에만 적용; 본 SPEC은 markdown in-place 편집이므로 manager-develop이 markdown 편집으로 수행)
- **mode selection (Phase 0.95)**: Mode 5 (sub-agent, sequential) — coding-heavy가 아닌 docs-heavy 연속 작업이므로 Mode 5 sequential이 안전 기본값(orchestration-mode-selection.md §B)

### Working location
- **main checkout** (plan-phase 규율 per spec-workflow.md § SPEC Phase Discipline Step 1; L2 worktree는 opt-in이며 본 SPEC은 main에서 수행)
- branch: `feat/SPEC-CC2178-DOCS-ALIGN-001` (run-phase 진입 시)

### Baseline (2026-06-16 검증)
- 6개 rules 파일 존재 확인 (`wc -l` 검증 완료)
- 9개 Tier 1 항목 전부 rules 파일에서 0 매치 (grep 검증 — 전부 신규 문서화 대상)
- docs-site 4-locale 디렉터리 구조 동일 (en/ko/ja/zh 각 16개 advanced 페이지 + claude-code 섹션)
- `moai spec lint` 사용 가능 (`/Users/goos/go/bin/moai`)

## §B — Known Issues (자동 주입)

본 SPEC은 docs-only이므로 manager-develop-prompt-template.md B1-B12 중 docs-relevant 항목만 적용:

- **B4 (Frontmatter Canonical Schema)**: N/A — 본 SPEC은 SPEC frontmatter를 수정하지 않는다.
- **B8 (Working Tree Hygiene)**: 무관 untracked files commit 금지. runtime-managed files (`.moai/harness/*`, `.moai/state/*`) 손대지 말 것.
- **B9 (Git Commit + Push)**: manager-develop은 본 SPEC scope 내 commit + push 자체 수행(main 직진 Tier S per git-workflow-doctrine.md). Conventional Commits format 의무(`feat(SPEC-CC2178-DOCS-ALIGN-001): M{N} <subject>`). `--no-verify` 절대 금지.
- **B10 (Untouched Paths PRESERVE)**: 본 SPEC §A.5 PRESERVE list 외 working tree 변경 금지. 특히 sibling MODEL-POLICY-REPAIR-001 scope 파일(`internal/template/model_policy.go` 등) 손대지 말 것.
- **B11 (AskUserQuestion 금지)**: subagent는 사용자와 직접 상호작용 금지. Blocker 시 structured blocker report 반환.
- **B12 (CHANGELOG emission)**: N/A — 본 SPEC은 CHANGELOG를 직접 write하지 않는다(sync-phase 책임).

**Cross-cutting (docs-specific)**:
- **4-locale parity**: ko만 편집하고 en/ja/zh 누락 시 `scripts/docs-i18n-check.sh`가 CI에서 실패. 각 마일스톤 완료 시 4-locale 동시 편집 확인.
- **§25 neutrality**: template rules 본문에 내부 SPEC ID / "per SPEC-CC2178-DOCS-ALIGN-001" / 날짜 / SHA 임베드 금지. generic prose만.
- **Mermaid TD-only + 강조/괄호 공백**: docs-site 편집 시 §17.2 [HARD] 준수.

## §C — Pre-flight Check List (착수 전 의무 검증)

```bash
# 1. 현재 branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. template-managed rules source 존재 확인 (source-first 규율)
ls internal/template/templates/.claude/rules/moai/core/settings-management.md \
   internal/template/templates/.claude/rules/moai/development/skill-authoring.md \
   internal/template/templates/.claude/rules/moai/development/agent-authoring.md \
   internal/template/templates/.claude/rules/moai/core/hooks-system.md \
   internal/template/templates/.claude/rules/moai/workflow/session-handoff.md \
   internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md

# 3. docs-site 4-locale 대상 페이지 존재 확인
for loc in en ko ja zh; do
  ls docs-site/content/$loc/advanced/settings-json.md \
     docs-site/content/$loc/advanced/hooks-reference.md \
     docs-site/content/$loc/advanced/skill-guide.md \
     docs-site/content/$loc/advanced/agent-guide.md
done

# 4. 대상 기능이 정말 미문서화인지 재활 (grep 0 매치 예상)
for term in "Tool(param" "closest-wins" "disableBundledSkills" "post-session" "/cd "; do
  echo "=== $term ==="
  grep -rl "$term" .claude/rules/moai/ 2>/dev/null | wc -l
done

# 5. lint baseline
moai spec lint .moai/specs/SPEC-CC2178-DOCS-ALIGN-001/ 2>&1 | tail -5

# 6. neutrality baseline (template source vs mirror 현재 차이)
diff -q internal/template/templates/.claude/rules/moai/core/settings-management.md \
         .claude/rules/moai/core/settings-management.md
# session-handoff.md는 intentional-DIFFER (§E.4.1) — DIFFER exit 1이 정상
diff -q internal/template/templates/.claude/rules/moai/workflow/session-handoff.md \
         .claude/rules/moai/workflow/session-handoff.md && echo "UNEXPECTED: session-handoff source==mirror" || echo "EXPECTED DIFFER (intentional, §E.4.1)"
grep -c "Diet Constraints\|V0 Abort Gate" .claude/rules/moai/workflow/session-handoff.md  # ≥ 3 기대
```

## §D — Constraints (DO NOT VIOLATE)

- **PRESERVE (수정 금지)**:
  - `internal/**/*.go` (Go 소스 전부 — ZERO Go code 변경)
  - `internal/template/model_policy.go` (sibling MODEL-POLICY-REPAIR-001 scope)
  - `pkg/**/*.go`, `cmd/**/*.go`
  - `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/**` (sibling SPEC 산출물)
  - `CLAUDE.md`, `CLAUDE.local.md` (본 SPEC scope 아님)
  - `.moai/research/cc-update-2.1.163-to-2.1.178.md` (read-only source)
- **금지 명령**: `--no-verify`, `--amend`, force-push to main, `git add -A` (specific path만)
- **의무 명령**: Conventional Commits (`feat(SPEC-CC2178-DOCS-ALIGN-001): M{N} ...`), `🗿 MoAI <email@mo.ai.kr>` trailer
- **4-locale 동시 편집 의무**: 각 docs-site 페이지 편집 시 ko/en/ja/zh 4개 locale 동일 커밋 또는 동일 마일스톤 내 반영
- **template-first**: rules 파일은 `internal/template/templates/` source 우선 편집 → `make build` → mirror 동기화
- **§25 neutrality**: template rules 본문에 내부 토큰(SPEC ID, REQ, 날짜, SHA) 임베드 금지

## §E — Self-Verification Deliverables

manager-develop 완료 보고 시 자체 검증 강제 포함:

**E1. AC Binary PASS/FAIL Matrix** (acceptance.md의 8 AC + 4-locale parity AC + no-Go-code AC)

**E2. No-Go-Code Verification (ZERO Go source changed)**
```bash
git diff --name-only HEAD~<N> HEAD | grep -E '\.go$' || echo "PASS: 0 Go files changed"
```

**E3. 4-Locale Parity Verification**
```bash
# 각 docs-site 페이지가 4 locale에 모두 존재 + 동일 편집
for page in advanced/settings-json.md advanced/hooks-reference.md advanced/skill-guide.md advanced/agent-guide.md advanced/statusline.md; do
  for loc in en ko ja zh; do
    grep -l "<target-feature>" docs-site/content/$loc/$page 2>/dev/null || echo "MISSING: $loc/$page"
  done
done
# 또는 scripts/docs-i18n-check.sh 실행
```

**E4. Template Mirror Parity (neutrality)**
```bash
# source와 mirror가 의도적 split 제외하고 동일한지 (make build 후)
for f in core/settings-management development/skill-authoring development/agent-authoring core/hooks-system workflow/orchestration-mode-selection; do
  diff -q internal/template/templates/.claude/rules/moai/$f.md .claude/rules/moai/$f.md
done
# 의도적 DIFFER 파일 (byte-parity 단언 금지):
#   - orchestration-mode-selection.md / agent-authoring.md / CLAUDE.md (precedent SPEC §B.1 참조)
#   - workflow/session-handoff.md (본 SPEC §E.4.1 intentional-DIFFER + §25 rationale 참조)
```

**E4.1 — Intentional-DIFFER manifest + §25 rationale (session-handoff.md)**

[HARD] `workflow/session-handoff.md`는 **intentional-DIFFER** 파일이다 — template source(`internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`, 209L)와 배포 mirror(`.claude/rules/moai/workflow/session-handoff.md`, 310L)가 의도적으로 다르며, mirror-only doctrine 101줄이 존재한다.

**§25 rationale (forbidden content class)**: mirror-only doctrine은 다음 internal SPEC ID들을 참조한다 — `LIFECYCLE-SYNC-GATE-001`, `HARNESS-NAMESPACE`, `SESSION-AUTO-RESUME-001` (mirror grep: 6 매치). 이들은 CLAUDE.local.md §25 [HARD] forbidden content class("내부 SPEC ID / REQ 토큰 / 내부 작업 날짜 / commit SHA / archive·memory 경로")에 해당하므로, template source로 이식(port)할 수 없다. source에 이식 시 `internal/template/internal_content_leak_test.go` + `template-neutrality-check.yaml` CI가 실패한다.

**Mirror-only doctrine sections** (source에는 부재, run-phase는 절대 source에서 삭제/이식 시도 금지):
- `## Diet Constraints` (L127 영역) — paste-ready resume message 다이어트 규칙
- `## V0 Abort Gate Doctrine` (L185 영역) — lsof + cwd 교차 검증 abort gate

**Run-phase 편집 규칙 (mirror-direct, source-first BYPASS)**:
1. `session-handoff.md`에 대한 REQ-DA-008(`/cd` cache-preserving 노트) 편집은 **mirror(`.claude/rules/moai/workflow/session-handoff.md`)를 직접 편집**한다.
2. source-first 규칙(C2)을 이 파일 한정으로 **bypass**한다 — `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`를 편집하지 않는다.
3. `make build`를 실행하더라도 source(209L)가 mirror(310L)을 덮어쓰는 현상이 발생하면 안 된다 — 실제로는 source가 SSOT이므로 `make build` 후 mirror가 source 기반으로 재생성된다. 따라서 run-phase는 **`make build`를 session-handoff.md에 대해 실행하지 않거나, 실행 후 mirror-only doctrine 101줄을 수동으로 복원**해야 한다. 안전 경로: session-handoff.md 편집 후 `make build`를 실행하지 말 것 (본 SPEC의 다른 5개 rules 파일은 source-first → `make build` → mirror 동기화 정상 경로를 따른다; session-handoff.md만 예외).
4. 편집 완료 후 검증: `diff -q source mirror`가 여전히 DIFFER (exit 1)이어야 하며, mirror-only doctrine 2개 섹션(`Diet Constraints`, `V0 Abort Gate`)과 6개 internal SPEC ID 참조가 보존되어야 한다 (`grep -c "Diet Constraints\|V0 Abort Gate" mirror ≥ 3`, `grep -c "LIFECYCLE-SYNC-GATE-001\|HARNESS-NAMESPACE\|SESSION-AUTO-RESUME-001" mirror ≥ 6`).

**E4.2 — template-first 규칙 적용 파일 (session-handoff.md 제외)**

나머지 5개 rules 파일은 정상 source-first 경로를 따른다:
- `core/settings-management.md` (REQ-DA-001, REQ-DA-004)
- `development/skill-authoring.md` (REQ-DA-002, REQ-DA-003, REQ-DA-004)
- `development/agent-authoring.md` (REQ-DA-002, REQ-DA-006)
- `core/hooks-system.md` (REQ-DA-005)
- `workflow/orchestration-mode-selection.md` (REQ-DA-007)

이 5개 파일은 source(`internal/template/templates/...`) 우선 편집 → `make build` → mirror 동기화 → neutrality 통과 정상 경로를 따른다 (C2).

**E5. Lint Status**
```bash
moai spec lint .moai/specs/SPEC-CC2178-DOCS-ALIGN-001/
go test ./internal/template/... -run TestTemplateNeutralityAudit
```

**E6. Branch HEAD + Push 상태**
- 새 commits SHA 리스트
- `git push origin feat/SPEC-CC2178-DOCS-ALIGN-001` 결과

**E7. Blocker Report (있을 시)** — structured 보고 (AskUserQuestion 절대 호출 금지)

## §F — Milestones (Tier S, priority-based, no time estimates)

### M1 — permissions + skills discovery (REQ-DA-001 ~ REQ-DA-004)

**우선순위**: High (REQ-DA-001, REQ-DA-002) / Medium (REQ-DA-003, REQ-DA-004)

**파일별 변경 맵**:

| 파일 | source 경로(template-first) | 변경 내용 |
|------|------------------------------|-----------|
| `.claude/rules/moai/core/settings-management.md` | `internal/template/templates/.claude/rules/moai/core/settings-management.md` | § Permission Management (L245 영역)에 `Tool(param:value)` wildcard 문법 + 예시 추가; § Claude Code Settings (L11 영역)에 `disableBundledSkills` + `--safe-mode` 노트 추가 |
| `.claude/rules/moai/development/skill-authoring.md` | `.../development/skill-authoring.md` | § Skill Scope and Priority에 nested `.claude/skills` 로딩 + closest-wins 선행 규칙 + `disableBundledSkills` 노트 추가 |
| `.claude/rules/moai/development/agent-authoring.md` | `.../development/agent-authoring.md` | nested `.claude/` closest-wins 선행 규칙 노트 추가 (agent/workflow/output-style 충돌) |
| `docs-site/content/{en,ko,ja,zh}/advanced/settings-json.md` | (docs-site, 4-locale 직접) | L362 영역 `Tool(specifier)` 아래에 `Tool(param:value)` wildcard subsection 추가; `disableBundledSkills` + `--safe-mode` 설정 노트 추가 |
| `docs-site/content/{en,ko,ja,zh}/advanced/skill-guide.md` | (docs-site, 4-locale) | nested skills 로딩 + closest-wins + `disableBundledSkills` 노트 추가 |
| `docs-site/content/{en,ko,ja,zh}/advanced/agent-guide.md` | (docs-site, 4-locale) | nested closest-wins 선행 규칙 노트 추가 |

**완료 기준**: REQ-DA-001~004의 관찰 가능한 본문 상태가 4-locale + rules 양쪽에 존재. `make build` 후 mirror 동기화. neutrality 통과.

### M2 — hooks + agent governance (REQ-DA-005 ~ REQ-DA-007)

**우선순위**: Medium (REQ-DA-005) / Low (REQ-DA-006, REQ-DA-007)

**파일별 변경 맵**:

| 파일 | source 경로 | 변경 내용 |
|------|-------------|-----------|
| `.claude/rules/moai/core/hooks-system.md` | `.../core/hooks-system.md` | § Hook Events에 `post-session` 라이프사이클 이벤트 추가 (self-hosted runner, 세션 종료 후 발화; MoAI wiring 안 함 명시) |
| `.claude/rules/moai/development/agent-authoring.md` | `.../development/agent-authoring.md` | § Frontmatter Format Rules `disallowedTools` 영역에 MCP server-level 강제 노트 추가 (CC 2.1.178 behavior fix) |
| `.claude/rules/moai/workflow/orchestration-mode-selection.md` | `.../workflow/orchestration-mode-selection.md` | §B 결정 트리 auto-mode 분기에 pre-launch classifier 노트 추가 |
| `docs-site/content/{en,ko,ja,zh}/advanced/hooks-reference.md` | (4-locale) | `post-session` 이벤트 추가 |
| `docs-site/content/{en,ko,ja,zh}/advanced/hooks-guide.md` | (4-locale) | `post-session` 라이프사이클 가이드 추가 |
| `docs-site/content/{en,ko,ja,zh}/advanced/agent-guide.md` | (4-locale) | `disallowedTools` MCP 강제 노트 추가 |

**완료 기준**: REQ-DA-005~007 관찰 가능 상태 존재. MoAI 미연결 항목은 "존재와 옵션" framing 유지.

### M3 — session resume / `/cd` (REQ-DA-008)

**우선순위**: Medium

**파일별 변경 맵**:

| 파일 | 편집 경로 | 변경 내용 |
|------|-----------|-----------|
| `.claude/rules/moai/workflow/session-handoff.md` | **mirror-direct** (`.claude/rules/moai/workflow/session-handoff.md` 직접 편집; source-first BYPASS — §E.4.1 참조) | § Worktree-Anchored Resume Pattern Block 0 영역에 `/cd` cache-preserving 대안 노트 추가. 기존 new-terminal Block 0를 대체하지 않고 보완. |
| `docs-site/content/{en,ko,ja,zh}/advanced/statusline.md` | (4-locale) | `/cd` 캐시 보존 노트 (이미 grep에서 `/cd` 매치되는 페이지) |
| `docs-site/content/{en,ko,ja,zh}/workflow-commands/moai-sync.md` | (4-locale) | `/cd` 재개 대안 노트 |

**주의 (session-handoff.md mirror-direct 규칙)**: 본 파일은 intentional-DIFFER다 (plan.md §E.4.1). run-phase는 source(`internal/template/templates/.../session-handoff.md`)를 편집하지 않고 mirror를 직접 편집한다. `make build`를 session-handoff.md에 대해 실행하지 말 것 (source 209L이 mirror 310L + 101줄 mirror-only doctrine을 덮어씀). 다른 5개 rules 파일은 정상 source-first 경로를 따른다.

**완료 기준**: REQ-DA-008 관찰 가능 상태 존재. Block 0 기존 계약 보존. mirror-only doctrine(`Diet Constraints`, `V0 Abort Gate`, 6개 internal SPEC ID 참조) 보존 — `grep -c "Diet Constraints\|V0 Abort Gate" .claude/rules/moai/workflow/session-handoff.md ≥ 3`.

## §G — Anti-Patterns

- **AP-DA-001 — ko-only docs-site 편집**: ko만 수정하고 en/ja/zh 누락. `scripts/docs-i18n-check.sh` CI 실패. 각 페이지 4-locale 동시 편집 의무.
- **AP-DA-002 — Go 코드 실수 수정**: "의존성 없는 util 함수"라며 `internal/` Go 소스 손대기. 본 SPEC은 ZERO Go code. 발견 시 즉시 revert + blocker report.
- **AP-DA-003 — template source 생략**: `.claude/rules/moai/` mirror만 편집하고 `internal/template/templates/` source를 안 고침. 다음 `make build` 시 덮어씌워짐. source 우선 의무. **예외: `session-handoff.md`는 mirror-direct 규칙(§E.4.1) 적용 — source-first를 의도적 bypass.**
- **AP-DA-003b — session-handoff.md source-first 위반**: `session-handoff.md`를 source(`internal/template/templates/...`)에서 편집하거나, `make build`로 source(209L)가 mirror(310L + 101줄 mirror-only doctrine)을 덮어쓰게 둠. §25 forbidden content class(internal SPEC ID 참조)가 source로 이식되어 neutrality CI 실패 + mirror-only doctrine 101줄 파괴. run-phase는 mirror를 직접 편집하고 `make build`를 session-handoff.md에 대해 실행 금지 (§E.4.1).
- **AP-DA-004 — §25 neutrality 위반**: template rules 본문에 "per SPEC-CC2178-DOCS-ALIGN-001" 또는 내부 날짜/SHA 임베드. neutrality CI 실패. generic prose만.
- **AP-DA-005 — MoAI 미연결 기능 과대 광고**: `post-session`을 "MoAI가 지원한다"고 서술. accuracy over completeness — "CC 기능으로 존재하며, MoAI는 현재 wiring하지 않는다"로 framing.
- **AP-DA-006 — 새 rules 파일 생성**: "전용 권한 규칙 파일을 만들자"며 새 `.claude/rules/...` 생성. 기존 파일 in-place 정합화만(C6).
- **AP-DA-007 — sibling MODEL-POLICY-REPAIR-001 scope 침범**: `model_policy.go` 또는 `availableModels` 문서를 본 SPEC에서 손대기. 명시적 exclusion(§E #2, #3).

## §H — Cross-References

- **Predecessor**: `.moai/specs/SPEC-CC-DOCS-ALIGNMENT-001/` (33 defects, completed, 동일 패턴)
- **Sibling**: `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/` (같은 CC 창, 비용 레버 분리)
- **Source research**: `.moai/research/cc-update-2.1.163-to-2.1.178.md` P4 (L192-196) + P5 (L198-200) + Tier 1 테이블 (L28-42) + C/Q/S-4 `/cd` (L125-134)
- **i18n 규칙**: `.moai/docs/docs-site-i18n-rules.md` §17.2 (Markdown/Hextra 규칙) + §17.3 (4-locale 동기화)
- **Template 규율**: `CLAUDE.local.md` §2 (Template-First) + §25 (Template Internal-Content Isolation)
- **Mode selection**: `.claude/rules/moai/workflow/orchestration-mode-selection.md` §B (Mode 5 default for docs-heavy sequential)
