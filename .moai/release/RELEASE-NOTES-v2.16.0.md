# MoAI-ADK v2.16.0 — Pattern Cookbook + V3R2 Backup Restore + V3R3 Phase A 통합 릴리스

**Release Date**: 2026-04-26
**Type**: Minor (consolidated)
**Branch**: `feat/SPEC-V3R3-PATTERNS-001-cookbook` → `main`
**Predecessor**: v2.14.0 (v2.15.0 was prepared but never tagged — fully included here)

---

## TL;DR

v2.16.0은 세 갈래의 작업을 **단일 통합 릴리스**로 묶습니다:

1. **V3R3 Phase B — Pattern Cookbook** (NEW): revfactory/harness Apache 2.0 6 reference docs를
   `.claude/rules/moai/` 하위로 흡수해 권위 있는 패턴 cookbook 확립.
2. **V3R2 Backup Restore**: Plan Audit Gate (Phase 0.5), FROZEN/EVOLVABLE Zone Registry
   (`moai constitution` CLI), Typed Memory Taxonomy (4-type enforcement) 복원.
3. **V3R3 Phase A — Foundation Hardening** (Convention Compliance Sweep, Expert Tool Uplift,
   Token Circuit Breaker, Mobile Native Coverage, Commands Cleanup) — v2.15.0에 준비됐으나
   tag되지 않아 v2.16.0에 합류.

전체 SPEC: 9개 (Phase B 1, V3R2 3, Phase A 5, +Plan Audit Gate). 모든 AC PASS, 회귀 0.

---

## Highlights

### Pattern Cookbook (SPEC-V3R3-PATTERNS-001) — NEW

agent 설계 / skill 작성 / 팀 운영 / QA 경계면 / orchestrator 선택을 위한 권위 있는 cookbook
6 rule files + NOTICE를 추가했습니다. 모든 파일은 frontmatter `paths` (CSV)로 conditional
auto-loading되며, Apache 2.0 attribution을 상단에 보존합니다.

| Path | 핵심 |
|------|------|
| `.claude/rules/moai/development/agent-patterns.md` | 6 architectural patterns + MoAI 어휘 매핑 |
| `.claude/rules/moai/development/orchestrator-templates.md` | 3 templates (Team-/Sub-/Hybrid-orchestrator) |
| `.claude/rules/moai/development/skill-ab-testing.md` | with-skill vs baseline A/B 방법론 |
| `.claude/rules/moai/development/skill-writing-craft.md` | description craft + 3-level disclosure + schema |
| `.claude/rules/moai/quality/boundary-verification.md` | 7 documented bug case studies (NEW dir) |
| `.claude/rules/moai/workflow/team-pattern-cookbook.md` | 5 team patterns (R/I/R/D/D) |
| `.claude/rules/moai/NOTICE.md` | Apache 2.0 attribution + source list |

License: revfactory/harness Apache 2.0 → MoAI-ADK MIT. NOTICE 보존 시 호환.

### Plan Audit Gate (SPEC-WF-AUDIT-GATE-001)

`/moai run` 실행 시 **Phase 0.5 mandatory gate**가 plan-auditor를 호출해 SPEC plan 산출물을
독립 검토합니다. 4 verdicts (PASS / FAIL / BYPASSED / INCONCLUSIVE), 7-day grace window
(FAIL_WARNED downgrade). minimal harness에서도 skip 불가.

리포트: `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md`

### FROZEN/EVOLVABLE Zone Registry (SPEC-V3R2-CON-001)

MoAI rule tree의 모든 HARD 조항에 대한 **단일 진실 공급원**. 68 entries (38 Frozen + 30
Evolvable). 새 CLI 서브커맨드:

- `moai constitution list [--zone frozen|evolvable] [--file <pattern>] [--format table|json]`
- `moai constitution guard <changed-rule-ids>` — CI에서 Frozen 위반 차단
- `moai doctor` Constitution Registry 점검 통합

200-entry cold load 벤치: ~1.85ms (목표 <10ms). 86.5% test coverage. Binary delta +33 KiB.

### Typed Memory Taxonomy (SPEC-V3R2-EXT-001)

`.claude/agent-memory/` 4-type enforcement: `user / feedback / project / reference`. 필수
frontmatter (`name / description / type`), `feedback`/`project` body 구조 (`**Why:**` +
`**How to apply:**`).

- 24h staleness detection → `<system-reminder>` wrap (≥10 동시 stale 시 aggregate warning)
- MEMORY.md 200-line cap (Claude Code memory loader 사일런트 절단 방지)
- PostToolUse audit: non-blocking stderr 경고 (`MEMORY_MISSING_TYPE`,
  `MEMORY_INDEX_OVERFLOW`, `MEMORY_DUPLICATE` 등)

`MOAI_MEMORY_AUDIT=0` 환경 변수로 일괄 마이그레이션 시 일시 비활성 가능.

### V3R3 Phase A 통합

v2.15.0에 준비됐던 Foundation Hardening 5 SPECs가 그대로 포함됩니다:

- **SPEC-V3R3-DEF-007** Convention Compliance Sweep — 11 skills `progressive_disclosure` blocks,
  manager-git Scope Boundaries 추가
- **SPEC-V3R3-ARCH-003** Expert Agent Tool Uplift — 7 expert agents `Agent` tool (max depth 2),
  expert-debug/performance에 Write/Edit, Escalation Protocol
- **SPEC-V3R3-ARCH-007** Token Circuit Breaker — `runtime.yaml` per-agent budgets, 75/90%
  threshold warning, /clear는 MANUAL 유지
- **SPEC-V3R3-COV-001** Mobile Native Coverage — expert-mobile (4-strategy router) +
  moai-domain-mobile, react-native deep, flutter deep skills
- **SPEC-V3R3-CMD-CLEANUP-001** Commands Cleanup — `/moai gate` 명령 wrapper 추가, review.md
  Phase 4 보안 강화, sync.md Phase 0.55 강화, `/moai context` 제거

---

## Breaking Changes

이번 릴리스의 BC 항목은 모두 v2.15.0에 누적됐던 V3R3 Phase A 변경입니다:

| ID | 영향 | 마이그레이션 |
|----|------|------------|
| **BC-V3R3-001** | 7 expert agents에 `Agent` tool 추가 (T2 → T2 escalation, max depth 2) | 자동 — 기존 호출 그대로 동작 |
| **BC-V3R3-002** | expert-debug, expert-performance에 `Write`, `Edit` 추가 | 자동 |
| **BC-V3R3-006** | `runtime.yaml` 신설 (token budget). 기본값 보수적, 75/90% threshold만 경고 emission | 자동 — 미설정 시 default 적용. 사용자 정의 시 `.moai/config/sections/runtime.yaml` 편집 |

`/moai context` 제거는 V3R3 Phase A에서 BC가 아닌 deprecation으로 처리됨 (`@MX` annotations
+ auto-memory로 대체).

---

## Verification

| Check | Result |
|-------|--------|
| `make build` | ✅ green (go:embed 정상 빌드) |
| `go test ./...` | ✅ all PASS |
| `go test ./internal/template/...` | ✅ PASS |
| `go test ./internal/constitution/...` | ✅ PASS, 86.5% coverage |
| `go test ./internal/hook/memo/...` | ✅ PASS, 91.7% coverage |
| `golangci-lint run` | ✅ 0 issues |
| Mirror diff (Pattern Cookbook 7 files) | ✅ 7/7 byte-identical |
| AC PASS rate | ✅ 100% (PATTERNS-001 5/5, CON-001 17/17, EXT-001 전체, WF-AUDIT-GATE-001 전체, Phase A 5 SPEC 전체) |
| Apache 2.0 NOTICE | ✅ all 6 cookbook files attributed |
| 16-language neutrality (cookbook) | ✅ no standalone install commands; pseudo-code where examples needed |

---

## Note on Versioning

v2.15.0은 system.yaml + CHANGELOG에 준비됐으나 **git tag가 push되지 않은 상태**에서 V3R2
backup restore가 추가됐습니다. 사용자 결정 (2026-04-26)에 따라 v2.15.0 tag는 건너뛰고 모든
누적 콘텐츠를 v2.16.0 단일 릴리스로 통합합니다. CHANGELOG의 v2.15.0 섹션은 historical
artifact로 보존됩니다.

---

## Upgrade Path

```bash
# Homebrew / direct binary download
brew upgrade moai-adk

# 또는 Go install
go install github.com/modu-ai/moai-adk/cmd/moai@v2.16.0

# Project upgrade
moai update
```

`moai update` 실행 시:
- 신규 6 cookbook rule 파일이 `.claude/rules/moai/development|quality|workflow/`로 자동 배포
- 기존 user customization (`.moai/project/`, `.moai/specs/`) 보호
- `.moai/config/sections/runtime.yaml` 신규 (Token Circuit Breaker)
- `.moai/config/sections/harness.yaml` 신규 (V3R3 Phase A에서 도입)
- `.claude/agent-memory/` 4-type taxonomy 점검 (warning만, 비차단)

업그레이드 후 권장 검증:
```bash
moai doctor                              # Zone Registry + 신규 항목 종합 점검
moai constitution list --zone frozen     # 38 Frozen entries 확인
go test ./internal/template/...          # 회귀 테스트 (개발 중인 경우)
```

---

## Next

- **v2.17.0** (Phase C — Extreme Aggressive 핵심): meta-harness skill + 16 정적 skills 제거
  (BC-V3R3-007), Vibe Design (DTCG 2025.10) Path A/B1/B2, /moai project 소크라테스 인터뷰
  + harness 자동 분기 (5-layer 통합 장치), Self-Learning Harness (4-tier 학습 + 5-layer safety).
- 4 SPECs: SPEC-V3R3-HARNESS-001, DESIGN-PIPELINE-001, PROJECT-HARNESS-001, **HARNESS-LEARNING-001**

---

## Credits

- **Pattern Cookbook source**: [revfactory/harness](https://github.com/revfactory/harness) — Apache License 2.0.
  Attribution preserved in `.claude/rules/moai/NOTICE.md`. 6 reference docs (agent-design-patterns,
  qa-agent-guide, skill-testing-guide, team-examples, orchestrator-template, skill-writing-guide).
- **Author**: GOOS행님 (Goos Kim)
- **Orchestrator**: MoAI (Claude Opus 4.7)

---

🗿 MoAI <email@mo.ai.kr>
