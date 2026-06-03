# Template Internal-Content Isolation Doctrine — extracted from CLAUDE.local.md §25

> Maintainer-local doctrine extracted from CLAUDE.local.md to cut session-launch context (CLAUDE.local.md loads in full at every launch). The matching CLAUDE.local.md section now carries a short stub pointing here. This file is NOT loaded at launch — read it when the topic applies. Subsection numbering is preserved so existing cross-references still resolve.

## 25. Template Internal-Content Isolation

[HARD] `internal/template/templates/` 하위 산출물은 **moai-adk를 사용하는 외부 사용자에게 deploy되는 범용 자산**이며, **moai-adk 내부 개발 과정의 흔적**을 포함해서는 안 된다. 사용자 프로젝트에 deploy되었을 때 의미를 가지지 않거나 (예: 다른 사용자의 SPEC ID), 잘못된 도메인 신호를 주거나 (예: REQ-ATR-* 등 moai-adk 내부 추적 ID), audit trail을 노출하면 (예: "Audit 3 Findings A1-A6") template의 본분이 깨진다. 본 §25는 **§15 (16-language neutrality)** + **§21 (97/98/99 dev-only commands isolation)** + **§24 (harness namespace separation)** 와 동일한 isolation doctrine 계열에 위치하며, 각각 다른 차원의 "내부 ≠ 배포" 경계를 정의한다.

### §25.1 정의 — Allowed vs Forbidden Content Classes

[HARD] template에 **포함 가능한** content classes (per REQ-TII-013):

1. **Generic prose**: 범용 정책 설명 (예: "범용 배포 vs 사용자 생성 분리", "byte-for-byte mirror parity")
2. **Generic mechanism descriptions**: 메커니즘 설명 (예: "predecessor SPEC supersession via frontmatter status: superseded")
3. **Generic examples**: 도메인-중립적 예시 (예: "When user runs `moai update`, the system shall ...")
4. **External public references**: 공개 자료 인용 (예: "Per Anthropic Claude Code documentation at claude.com/docs/en/sub-agents, subagents cannot spawn other subagents")
5. **Permanent rule citations**: 영구적 규칙 인용 (예: ".claude/rules/moai/core/agent-common-protocol.md § User Interaction Boundary")
6. **MoAI-ADK system identifiers**: 시스템 정체성 자체 (예: "MoAI-ADK", "MoAI orchestrator", "MoAI agent catalog")

[HARD] template에 **절대 포함 금지** content classes (per REQ-TII-001):

1. **moai-adk internal SPEC IDs**: 본 프로젝트의 SPEC 식별자 (예: `SPEC-V3R6-AGENT-TEAM-REBUILD-001`, `SPEC-V3R6-WORKFLOW-OPT-001`)
2. **moai-adk internal REQ tokens**: 본 프로젝트의 requirement 식별자 (예: `REQ-ATR-007`, `REQ-WO-013`, `REQ-COORD-018`)
3. **moai-adk internal AC tokens**: 본 프로젝트의 acceptance criterion 식별자 (예: `AC-ATR-022`, `AC-WO-007`)
4. **Audit/post-mortem citations**: 본 프로젝트의 audit 인용 (예: "Audit 3 Findings A1-A6", "verbatim Anthropic 2026 sources cited in spec.md §B.1")
5. **Internal session dates**: 본 프로젝트 작업 날짜 (예: "2026-05-25", "2026-05-23", "added 2026-05-22") — 단, generic prose 안의 "today" / "yesterday" 같은 상대 표현은 OK
6. **Internal archive paths**: 본 프로젝트 archive 경로 (예: `.moai/backups/agent-archive-2026-05-25/`)
7. **Internal commit SHAs**: 본 프로젝트 commit hash (예: `b957a4d04`, `d9838995d`) — 단, generic prose 안의 short-sha mention `abc1234 `로 끝나는 trailing space 패턴은 grep 노이즈로 허용 (D-007 inline 해소; M3 lint test 참조)
8. **Internal memory file references**: 본 프로젝트 memory hash 경로 (예: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)

### §25.2 Forbidden / Allowed Worked Examples

| Class | Forbidden 예시 | Allowed Substitution 예시 |
|-------|---------------|--------------------------|
| SPEC ID literal | `Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 (2026-05-25), the agent catalog ...` | `Per the canonical MoAI agent catalog policy, the agent catalog ...` |
| REQ token | `REQ-ATR-008 specifies orchestration mode selection` | `The orchestration mode selection rule (.claude/rules/moai/workflow/orchestration-mode-selection.md) specifies ...` |
| AC token | `AC-ATR-022 verifies hook subagent boundary` | `The hook subagent boundary verification rule verifies ...` |
| Audit citation | `Per Audit 3 Finding A1 (verbatim Anthropic)` | `Per the canonical principle from Anthropic Claude Code documentation: "Subagents cannot spawn other subagents." (Source: https://claude.com/docs/en/sub-agents)` |
| Date | `Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 plan-phase commit (2026-05-25)` | `Per the canonical agent catalog policy` |
| Archive path | `12 agents archived at .moai/backups/agent-archive-2026-05-25/` | (omit entirely, or rephrase: `12 archived agents are preserved offline for reference`) |
| Commit SHA | `commit b957a4d04 introduced ...` | (omit entirely, or rephrase: `The canonical rule introduced ...`) |
| Memory ref | `Per ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_template_internal_content_isolation.md` | (omit; memory is per-user, never reference in template) |

### §25.3 Pre-commit Self-Check (5-item Mandatory Checklist)

[HARD] template 영역 (`internal/template/templates/**`) 의 staged changes commit 전에 다음 5개 항목을 **체크리스트로** 수동 통과시켜야 한다. 자동화는 §25.4 의 Go lint test (`TestTemplateNoInternalContentLeak`) 가 수행하며, 본 self-check는 그 lint test가 catch하지 못하는 의미적 누출까지 차단하기 위한 사람-인-더-루프(HITL) layer다.

- [ ] **C1 — SPEC ID literal**: staged diff 안에 `SPEC-V3R6-` / `SPEC-AGENCY-` / 기타 본 프로젝트 SPEC ID prefix가 등장하지 않는다. (단, `internal/template/templates/.moai/specs/.example/` 같은 example fixture 안의 `SPEC-XXX-001` 등 generic placeholder는 OK)
- [ ] **C2 — REQ/AC token**: staged diff 안에 `REQ-XXX-NNN` / `AC-XXX-NNN` 형태의 본 프로젝트 추적 token이 등장하지 않는다. (단, EARS/GEARS 본문 안에 `REQ-EXAMPLE-NNN` 등 generic placeholder는 OK)
- [ ] **C3 — Audit citation**: "Audit N Finding AX" / "verbatim Anthropic 2026 sources" / "spec.md §X.Y" 같은 audit 인용 패턴이 등장하지 않는다.
- [ ] **C4 — Date / short-sha**: 본 프로젝트 작업 날짜 (`2026-MM-DD` ISO 형식) 또는 commit short-sha (`[0-9a-f]{7,8}`) 가 prose 안에 등장하지 않는다. (단, `version: vX.Y.Z` 같은 SemVer는 OK. 또한 short-sha sentence-final 패턴은 D-007 inline 해소 후 lint test가 허용)
- [ ] **C5 — Memory/archive path**: `~/.claude/projects/-Users-goos-` / `.moai/backups/` 같은 본 프로젝트 maintainer-only 경로가 등장하지 않는다.

체크리스트 위반 발견 시: substitution dictionary (`.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md §B`) 의 generic-prose 치환 패턴을 적용한 후 다시 self-check 통과시키고 commit 진행.

### §25.4 Anti-pattern Catalogue (predecessor cleanup 사례 기반)

predecessor cleanup history (`research.md` §B 참조: chore commits `20a66df85` pass 1 + `40dc43f5b` pass 2) 에서 surface된 3개 anti-pattern:

#### Anti-pattern AP-25.1 — Audit/citation 본문 leak (Audit 3 Findings citation 패턴)

원천: M5/M7 작업이 NOTICE.md (`.claude/rules/moai/NOTICE.md` 의 template mirror) 에 "Audit 3 Findings A1-A6" 6개 verbatim Anthropic citation을 통째로 mirror.

위반 양상: template `.claude/rules/moai/NOTICE.md` 가 사용자 프로젝트에 deploy되면, 사용자는 "Audit 3"이 자기 프로젝트의 audit인지 moai-adk의 audit인지 알 수 없다. citation 자체는 공개 자료이지만 "Audit 3" 식별자는 moai-adk 내부 추적 번호.

올바른 패턴: NOTICE.md 본문에 verbatim citation은 유지 (license-required attribution) 하되, "Audit 3 Findings A1-A6" wrapper는 제거하고 "Anthropic 2026 Alignment" 같은 generic header 아래 직접 인용. SPEC ID, plan-phase commit hash, date 도 모두 제거.

탐지 패턴: `grep -rn 'Audit [0-9]\|Finding A[0-9]' internal/template/templates/`

#### Anti-pattern AP-25.2 — SPEC ID 본문 cross-reference leak

원천: 30+ template files (`agents/`, `rules/`, `skills/` 산하) 가 SPEC 본문에서 derive된 정책 변경 후 cross-reference로 "Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 (M1-M8 milestones)" 형태로 SPEC ID 인용을 추가.

위반 양상: 사용자 프로젝트에 deploy된 template에서 사용자가 `SPEC-V3R6-AGENT-TEAM-REBUILD-001`을 검색하면 자기 SPEC 디렉토리에 그런 SPEC이 없어 confusion. SPEC ID 자체는 본 프로젝트 추적용이지 정책의 본질적 일부가 아님.

올바른 패턴: SPEC ID cross-reference를 generic-prose 정책 인용으로 대체. 예: `Per SPEC-V3R6-AGENT-TEAM-REBUILD-001 REQ-ATR-009 (lint + test + coverage delta gate)` → `Per the canonical sync-phase quality gate policy (lint + test + coverage delta)`. 추가 detail이 필요하면 `.claude/rules/moai/...` 파일 자체를 인용 (rule 파일은 template-mirror 형태로 같이 deploy 되므로 reference 유효).

탐지 패턴: `grep -rn 'SPEC-V3R6-\|SPEC-AGENCY-\|SPEC-WORKTREE-' internal/template/templates/`

#### Anti-pattern AP-25.3 — REQ token + date 본문 leak (canonical rule citation 변형)

원천: agent body 파일 (특히 `.claude/agents/moai/manager-*.md` template mirrors) 이 정책 변경 사유 footnote로 "Per REQ-ATR-008 + REQ-ATR-014 (added 2026-05-25)" 같이 REQ token + date 조합을 추가.

위반 양상: REQ token은 본 프로젝트 acceptance.md row 식별자. 사용자 프로젝트에는 acceptance.md 가 없어 token 자체가 의미 없음. date는 본 프로젝트 작업 history 노출.

올바른 패턴: REQ token + date를 모두 제거하고 정책의 essence만 prose로 유지. 정말로 추적이 필요하면 maintainer-only `CLAUDE.local.md` (현재 §25 같은 doctrine 섹션 안) 에 옮기고, template body는 generic 유지.

탐지 패턴: `grep -rn 'REQ-[A-Z][A-Z]\+-[0-9]\{3\}\|AC-[A-Z][A-Z]\+-[0-9]\{3\}\|20[2-9][0-9]-[0-1][0-9]-[0-3][0-9]' internal/template/templates/`

### §25.5 운영 원칙 + Cross-references

- [HARD] template 영역에 staged 변경이 있을 때 §25.3 5-item self-check **수동 통과 의무**. 자동화 (Go lint test `TestTemplateNoInternalContentLeak`) 와 병행하여 의미적 누출 방어막 (in-depth defense).
- [HARD] §25.1 forbidden classes 가 audit-time에 발견되면 즉시 cleanup chore commit + 본 §25 cross-reference. predecessor cleanup history pattern (`20a66df85` + `40dc43f5b`) 답습.
- [SHOULD] 새로운 forbidden class 등장 시 (예: 향후 "Phase 6 Findings" 같은 새 audit 형식) §25.1 forbidden list 확장 + §25.4 anti-pattern 사례 추가.
- Cross-references:
  - §15 — Template language neutrality (16-language equal treatment)
  - §21 — 97/98/99 dev-only commands isolation
  - §24 — Harness namespace separation (`harness-*` user-owned vs `moai-harness-*` template-managed)
  - SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 §B (Substitution Dictionary) — generic-prose 치환 패턴 SSOT
  - `internal/template/internal_content_leak_test.go` — automated regression guard (M3 deliverable)
