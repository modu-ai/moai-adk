# t1257 역할 명칭 인벤토리 — 문서 계층 (`leader` · `lane` 통일 준비)

- 카드: t1257 (Tier L · 클래스 C · plan 단계 전용)
- 측정 트리: 워크트리 `.claude/worktrees/t1257`, 브랜치 `WT-role-naming-docs`, HEAD `e62c3e183` (develop 기준점과 동일)
- 측정 시각: 2026-09-26, 이 트리에서 직접 측정. 리드가 준 기준선은 인용하지 않고 §2.3 에서 같은 명령으로 재현했다.
- 이 문서는 **치환을 하지 않는다.** 본문 치환은 코드 계층 카드 t1256 의 결론이 확정된 뒤 `SPEC-ROLE-NAMING-DOCS-001` 의 run 단계에서만 일어난다.

## 0. 측정 방법

모든 표는 아래 절차로 이 트리에서 생성했다. 스크립트 원본은 `.moai/reports/t1257/raw/scripts/` 에 있다.

```bash
git ls-files > "$SP/files.txt"                                   # 12056 행
python3 raw/scripts/inv.py     "$SP" .moai/reports/t1257/raw     # per-file-class.tsv, per-file-pivot.tsv, hard-lines.tsv, headings.tsv, mirror-pairs.tsv, docs-locale-parity.tsv
python3 raw/scripts/anchors.py "$SP" .moai/reports/t1257/raw     # anchors.tsv
python3 raw/scripts/cjk2.py    "$SP" .moai/reports/t1257/raw     # cjk-per-file.tsv, cjk-summary.txt
python3 raw/scripts/table.py   .moai/reports/t1257/raw           # 아래 §1 표
```

- **대상 표면(문서 계층)**: `internal/template/templates/**` 의 텍스트 파일(`.md .tmpl .toml .yaml .yml .json .txt .sh`), 로컬 `.claude/{rules,agents,skills,output-styles,commands}/**`, 루트 `CLAUDE.md` · `AGENTS.md` · `CLAUDE.local.md`, `.moai/config/**`, `docs-site/content/{en,ko,ja,zh}/**`, `README*.md` 4개. `.go` 는 전부 제외(코드 계층 = t1256). `.claude/loop.md`(로컬)는 표면 분류에서 빠졌다 — 템플릿 짝 `T:.claude/loop.md` 로만 잡혔다(§9 Gaps).
- **토큰(영문, 대소문자 무시, 단어 경계)**: `lead`(lead/leads/lead's) · `leader` · `lane` · `worker` · `companion` · `foreman` · `deputy` · `coordinator`. `coordinator` 는 측정 중 발견한 추가 별칭이다(`manager-lead` 의 문서 제목 "Lead Coordinator" 등).
- **분류(휴리스틱, 매치 단위)**: 규칙은 `inv.py` `classify()` 에 있다. 순서대로 판정한다.
  - `ident:agent-name` — 앞이 `manager-`/`manager_` (예: `manager-lead`)
  - `ident:sentinel/env` — 전부 대문자 (예: `LEAD-MERGE-APPROVED`, `MOAI_KANBAN_LEAD_ADDR`)
  - `ident:snake` — `_` 인접 (예: `lead_session_id`)
  - `ident:notation` — 뒤가 `-숫자`/`-N`/`-<n>` (예: `worker-3`, `lane-<n>`)
  - `ident:cli-token` — 앞이 `-f`/`--name`/`--role`/`role=` (예: `-f worker`)
  - `compound(slug-or-prose)` — 하이픈 합성어 (예: `moai-kanban-foreman`, `lane-local`, `leaf-worker`). 식별자와 산문이 섞여 있어 줄 단위 판정이 필요하다.
  - `plain-english` — 일반 영어 (`lead to`, `service worker`, `swim lane` 등)
  - `other-meaning:doc-companion` — 「detail companion」 문서 짝 뜻
  - `role:subagent-worker` — `manager-lead` 의 leaf worker(서브에이전트) 뜻
  - `other-meaning:agent-teams/cg` — Agent Teams 의 team lead / `moai cg` 의 leader pane 뜻(칸반·팩토리 문맥 단서가 없는 줄)
  - `frozen/historical` — HISTORY 절, CHANGELOG, `deprecated`/`legacy`/`retired`/`SUPERSEDED` 줄
  - `role` — 나머지(칸반·팩토리 역할 뜻으로 추정)
- **HARD** 열: 해당 파일에서 `[HARD]` 와 그 토큰이 같은 줄에 있는 줄 수. 제목 줄에만 `[HARD]` 가 있고 본문에 토큰이 있는 경우는 잡히지 않는다(§9).
- **anchor** 열: 그 토큰을 담은 제목이 다른 파일에서 `§ <제목>` 또는 `#<slug>` 로 참조되면 `Y`.

---

## 1. 전체 인벤토리 표 (파일 × 토큰 × 의미 분류)

631 행. 순서는 템플릿 → 로컬 → docs-site(en, ko, ja, zh) → README. 행 단위 원자료는 `raw/per-file-class.tsv`(파일 × 토큰 × 분류 × 개수)다.

| # | file | surface | token | total | role | ident | compound | other-meaning | plain-en | frozen | HARD | anchor |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md` | T:.claude/rules | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 2 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md` | T:.claude/rules | worker | 2 | 2 (sub 2) | 0 | 0 | 0 | 0 | 0 |  |  |
| 3 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md` | T:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 4 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | T:.claude/rules | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 5 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | T:.claude/rules | lane | 15 | 14 | 0 | 1 | 0 | 0 | 0 | 1 |  |
| 6 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | T:.claude/rules | companion | 2 | 1 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 7 | `internal/template/templates/.claude/rules/moai/core/askuser-protocol-reference.md` | T:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 8 | `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` | T:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 9 | `internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md` | T:.claude/rules | leader | 3 | 0 | 0 | 2 | 1 (agent-teams/cg 1) | 0 | 0 | 1 |  |
| 10 | `internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` | T:.claude/rules | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 11 | `internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` | T:.claude/rules | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 12 | `internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 13 | `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` | T:.claude/rules | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 14 | `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 15 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | T:.claude/rules | lead | 10 | 6 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 16 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | T:.claude/rules | lane | 6 | 6 | 0 | 0 | 0 | 0 | 0 |  |  |
| 17 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | T:.claude/rules | worker | 5 | 5 | 0 | 0 | 0 | 0 | 0 |  |  |
| 18 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 19 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md` | T:.claude/rules | lead | 2 | 1 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 20 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md` | T:.claude/rules | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 21 | `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md` | T:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 22 | `internal/template/templates/.claude/rules/moai/core/native-idiom-and-register-detail.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 23 | `internal/template/templates/.claude/rules/moai/core/output-style-localization-catalogue.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 | 1 |  |
| 24 | `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity-detail.md` | T:.claude/rules | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 25 | `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity-detail.md` | T:.claude/rules | companion | 6 | 0 | 0 | 2 | 4 (doc-companion 4) | 0 | 0 | 2 |  |
| 26 | `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` | T:.claude/rules | companion | 2 | 1 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 | 1 |  |
| 27 | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | T:.claude/rules | lead | 8 | 2 | 3 (agent-name 3) | 1 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 28 | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | T:.claude/rules | worker | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 29 | `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | T:.claude/rules | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 30 | `internal/template/templates/.claude/rules/moai/development/agent-patterns.md` | T:.claude/rules | lead | 7 | 0 | 1 (agent-name 1) | 0 | 6 (agent-teams/cg 6) | 0 | 0 |  |  |
| 31 | `internal/template/templates/.claude/rules/moai/development/agent-patterns.md` | T:.claude/rules | worker | 8 | 7 | 0 | 0 | 0 | 0 | 1 |  |  |
| 32 | `internal/template/templates/.claude/rules/moai/development/agent-patterns.md` | T:.claude/rules | coordinator | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 33 | `internal/template/templates/.claude/rules/moai/development/model-policy.md` | T:.claude/rules | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 34 | `internal/template/templates/.claude/rules/moai/development/model-policy.md` | T:.claude/rules | leader | 3 | 0 | 0 | 0 | 3 (agent-teams/cg 3) | 0 | 0 |  |  |
| 35 | `internal/template/templates/.claude/rules/moai/development/model-policy.md` | T:.claude/rules | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 36 | `internal/template/templates/.claude/rules/moai/development/model-policy.md` | T:.claude/rules | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 37 | `internal/template/templates/.claude/rules/moai/development/orchestrator-templates.md` | T:.claude/rules | lead | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 38 | `internal/template/templates/.claude/rules/moai/development/rule-authoring.md` | T:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 39 | `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 40 | `internal/template/templates/.claude/rules/moai/development/sprint-round-naming.md` | T:.claude/rules | lane | 15 | 12 | 0 | 0 | 0 | 0 | 3 |  |  |
| 41 | `internal/template/templates/.claude/rules/moai/languages/cpp.md` | T:.claude/rules | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 42 | `internal/template/templates/.claude/rules/moai/languages/go.md` | T:.claude/rules | worker | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 43 | `internal/template/templates/.claude/rules/moai/languages/kotlin.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 44 | `internal/template/templates/.claude/rules/moai/languages/rust.md` | T:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 45 | `internal/template/templates/.claude/rules/moai/quality/boundary-verification.md` | T:.claude/rules | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 46 | `internal/template/templates/.claude/rules/moai/workflow/archived-agent-rejection.md` | T:.claude/rules | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 47 | `internal/template/templates/.claude/rules/moai/workflow/archived-agent-rejection.md` | T:.claude/rules | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 48 | `internal/template/templates/.claude/rules/moai/workflow/cache-aware-execution-reference.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 49 | `internal/template/templates/.claude/rules/moai/workflow/context-window-management-detail.md` | T:.claude/rules | companion | 5 | 1 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 50 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md` | T:.claude/rules | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 51 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md` | T:.claude/rules | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 52 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md` | T:.claude/rules | worker | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 53 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging-detail.md` | T:.claude/rules | companion | 5 | 1 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 54 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` | T:.claude/rules | lead | 4 | 2 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 55 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` | T:.claude/rules | worker | 5 | 4 | 0 | 1 | 0 | 0 | 0 |  |  |
| 56 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 57 | `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` | T:.claude/rules | coordinator | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 58 | `internal/template/templates/.claude/rules/moai/workflow/goal-directive-detail.md` | T:.claude/rules | companion | 3 | 0 | 0 | 0 | 3 (doc-companion 3) | 0 | 0 |  |  |
| 59 | `internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` | T:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 60 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | T:.claude/rules | lead | 79 | 54 | 14 (agent-name 13,sentinel/env 1) | 2 | 2 (agent-teams/cg 2) | 4 | 3 | 2 | Y |
| 61 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | T:.claude/rules | lane | 57 | 44 | 0 | 8 | 0 | 4 | 1 |  | Y |
| 62 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | T:.claude/rules | worker | 9 | 5 | 4 (notation 4) | 0 | 0 | 0 | 0 |  |  |
| 63 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | T:.claude/rules | companion | 20 | 17 | 0 | 0 | 3 (doc-companion 3) | 0 | 0 | 1 |  |
| 64 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | T:.claude/rules | deputy | 29 | 27 | 1 (sentinel/env 1) | 1 | 0 | 0 | 0 | 1 |  |
| 65 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | T:.claude/rules | lead | 87 | 76 | 8 (agent-name 7,sentinel/env 1) | 0 | 0 | 3 | 0 | 20 |  |
| 66 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | T:.claude/rules | lane | 57 | 53 | 0 | 4 | 0 | 0 | 0 | 8 | Y |
| 67 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | T:.claude/rules | worker | 6 | 2 (sub 1) | 4 (cli-token 1,notation 3) | 0 | 0 | 0 | 0 |  |  |
| 68 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | T:.claude/rules | companion | 18 | 13 | 0 | 1 | 4 (doc-companion 4) | 0 | 0 | 3 |  |
| 69 | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | T:.claude/rules | deputy | 10 | 10 | 0 | 0 | 0 | 0 | 0 | 5 | Y |
| 70 | `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 71 | `internal/template/templates/.claude/rules/moai/workflow/moai-memory.md` | T:.claude/rules | lead | 2 | 0 | 0 | 0 | 0 | 2 | 0 |  |  |
| 72 | `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` | T:.claude/rules | lead | 16 | 3 | 10 (agent-name 10) | 1 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 73 | `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` | T:.claude/rules | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 74 | `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` | T:.claude/rules | worker | 10 | 6 (sub 1) | 1 (cli-token 1) | 2 | 0 | 0 | 1 |  |  |
| 75 | `internal/template/templates/.claude/rules/moai/workflow/resource-slot-lease.md` | T:.claude/rules | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 76 | `internal/template/templates/.claude/rules/moai/workflow/runtime-recovery-doctrine.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 77 | `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` | T:.claude/rules | leader | 2 | 1 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 78 | `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` | T:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 79 | `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 80 | `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` | T:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 81 | `internal/template/templates/.claude/rules/moai/workflow/skill-routing-detail.md` | T:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 82 | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | T:.claude/rules | worker | 1 | 0 | 0 | 0 | 0 | 0 | 1 |  |  |
| 83 | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | T:.claude/rules | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 | 1 |  |
| 84 | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | T:.claude/rules | lane | 5 | 4 | 0 | 1 | 0 | 0 | 0 |  |  |
| 85 | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | T:.claude/rules | worker | 2 | 2 (sub 1) | 0 | 0 | 0 | 0 | 0 | 1 |  |
| 86 | `internal/template/templates/.claude/agents/moai/e2e-tester.md` | T:.claude/agents | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 87 | `internal/template/templates/.claude/agents/moai/manager-design.md` | T:.claude/agents | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 88 | `internal/template/templates/.claude/agents/moai/manager-develop.md` | T:.claude/agents | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 89 | `internal/template/templates/.claude/agents/moai/manager-git.md` | T:.claude/agents | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 90 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | T:.claude/agents | lead | 63 | 33 | 26 (agent-name 18,cli-token 4,sentinel/env 4) | 2 | 1 (agent-teams/cg 1) | 0 | 1 | 1 |  |
| 91 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | T:.claude/agents | lane | 28 | 18 | 3 (cli-token 1,notation 2) | 1 | 0 | 0 | 6 | 1 |  |
| 92 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | T:.claude/agents | worker | 36 | 21 (sub 18) | 5 (notation 5) | 6 | 0 | 2 | 2 |  |  |
| 93 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | T:.claude/agents | companion | 4 | 3 | 0 | 0 | 0 | 0 | 1 |  |  |
| 94 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | T:.claude/agents | deputy | 26 | 23 | 2 (sentinel/env 2) | 1 | 0 | 0 | 0 | 1 | Y |
| 95 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | T:.claude/agents | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 96 | `internal/template/templates/.claude/agents/moai/manager-spec.md` | T:.claude/agents | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 97 | `internal/template/templates/.claude/agents/moai/mission-governor.md` | T:.claude/agents | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 98 | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | T:.claude/agents | lead | 2 | 0 | 0 | 0 | 0 | 1 | 1 | 1 |  |
| 99 | `internal/template/templates/.claude/agents/moai/super-advisor.md` | T:.claude/agents | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 100 | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | T:.claude/agents | lead | 2 | 0 | 0 | 0 | 0 | 1 | 1 | 1 |  |
| 101 | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | T:.claude/agents | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 102 | `internal/template/templates/.claude/skills/moai-domain-html-report/SKILL.md` | T:.claude/skills | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 103 | `internal/template/templates/.claude/skills/moai-domain-humanize/modules/copy-review.md` | T:.claude/skills | lead | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 104 | `internal/template/templates/.claude/skills/moai-domain-humanize/modules/english.md` | T:.claude/skills | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 105 | `internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md` | T:.claude/skills | lane | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 106 | `internal/template/templates/.claude/skills/moai-domain-svg-infographic/references/archetypes.md` | T:.claude/skills | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 107 | `internal/template/templates/.claude/skills/moai-domain-svg-infographic/references/authoring.md` | T:.claude/skills | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 108 | `internal/template/templates/.claude/skills/moai-domain-svg-infographic/references/authoring.md` | T:.claude/skills | leader | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 109 | `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md` | T:.claude/skills | lead | 2 | 1 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 110 | `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md` | T:.claude/skills | worker | 4 | 3 (sub 2) | 0 | 1 | 0 | 0 | 0 |  |  |
| 111 | `internal/template/templates/.claude/skills/moai-foundation-cc/reference/advanced-agent-patterns.md` | T:.claude/skills | lead | 3 | 2 | 0 | 0 | 0 | 1 | 0 |  |  |
| 112 | `internal/template/templates/.claude/skills/moai-foundation-cc/reference/advanced-agent-patterns.md` | T:.claude/skills | worker | 6 | 5 (sub 2) | 0 | 1 | 0 | 0 | 0 |  |  |
| 113 | `internal/template/templates/.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` | T:.claude/skills | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 114 | `internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md` | T:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 115 | `internal/template/templates/.claude/skills/moai-foundation-thinking/SKILL.md` | T:.claude/skills | lead | 2 | 0 | 0 | 0 | 0 | 2 | 0 |  |  |
| 116 | `internal/template/templates/.claude/skills/moai-harness-learner/SKILL.md` | T:.claude/skills | coordinator | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 117 | `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` | T:.claude/skills | lane | 5 | 2 | 0 | 3 | 0 | 0 | 0 |  |  |
| 118 | `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` | T:.claude/skills | worker | 18 | 17 | 0 | 1 | 0 | 0 | 0 |  |  |
| 119 | `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` | T:.claude/skills | foreman | 11 | 10 | 0 | 1 | 0 | 0 | 0 |  |  |
| 120 | `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` | T:.claude/skills | companion | 5 | 4 | 0 | 0 | 0 | 0 | 1 |  |  |
| 121 | `internal/template/templates/.claude/skills/moai-meta-harness/references/seven-phase-workflow.md` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 122 | `internal/template/templates/.claude/skills/moai-meta-harness/references/seven-phase-workflow.md` | T:.claude/skills | worker | 3 | 3 (sub 1) | 0 | 0 | 0 | 0 | 0 |  |  |
| 123 | `internal/template/templates/.claude/skills/moai-ref-ui-polish/references/motion-principles.md` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 124 | `internal/template/templates/.claude/skills/moai-workflow-project/references/navigator.md` | T:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 125 | `internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` | T:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 126 | `internal/template/templates/.claude/skills/moai-workflow-project/templates/question-templates/spec-workflow-setup.json` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 127 | `internal/template/templates/.claude/skills/moai-workflow-testing/modules/advanced-patterns.md` | T:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 128 | `internal/template/templates/.claude/skills/moai-workflow-testing/modules/optimization.md` | T:.claude/skills | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 129 | `internal/template/templates/.claude/skills/moai-workflow-testing/modules/performance-optimization/ai-optimization.md` | T:.claude/skills | worker | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 130 | `internal/template/templates/.claude/skills/moai-workflow-testing/modules/performance-optimization/bottleneck-detection.md` | T:.claude/skills | worker | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 131 | `internal/template/templates/.claude/skills/moai-workflow-testing/modules/performance/optimization-patterns.md` | T:.claude/skills | worker | 2 | 1 | 0 | 0 | 0 | 1 | 0 |  |  |
| 132 | `internal/template/templates/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` | T:.claude/skills | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 133 | `internal/template/templates/.claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 134 | `internal/template/templates/.claude/skills/moai-workflow-worktree/modules/parallel-workflows.md` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 135 | `internal/template/templates/.claude/skills/moai-workflow-worktree/modules/tools-integration.md` | T:.claude/skills | worker | 3 | 1 | 0 | 0 | 0 | 2 | 0 |  |  |
| 136 | `internal/template/templates/.claude/skills/moai-workflow-worktree/modules/worktree-management.md` | T:.claude/skills | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 137 | `internal/template/templates/.claude/skills/moai/SKILL.md` | T:.claude/skills | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 138 | `internal/template/templates/.claude/skills/moai/workflows/design.md` | T:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 139 | `internal/template/templates/.claude/skills/moai/workflows/e2e.md` | T:.claude/skills | lane | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 140 | `internal/template/templates/.claude/skills/moai/workflows/goal.md` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 141 | `internal/template/templates/.claude/skills/moai/workflows/goal.md` | T:.claude/skills | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 142 | `internal/template/templates/.claude/skills/moai/workflows/gtd.md` | T:.claude/skills | lead | 11 | 9 | 2 (agent-name 1,sentinel/env 1) | 0 | 0 | 0 | 0 | 1 |  |
| 143 | `internal/template/templates/.claude/skills/moai/workflows/gtd.md` | T:.claude/skills | lane | 6 | 6 | 0 | 0 | 0 | 0 | 0 | 1 |  |
| 144 | `internal/template/templates/.claude/skills/moai/workflows/gtd.md` | T:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 145 | `internal/template/templates/.claude/skills/moai/workflows/gtd.md` | T:.claude/skills | foreman | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 146 | `internal/template/templates/.claude/skills/moai/workflows/harness-build-entry.md` | T:.claude/skills | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 147 | `internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | T:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 148 | `internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | T:.claude/skills | worker | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 149 | `internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | T:.claude/skills | companion | 10 | 9 | 0 | 1 | 0 | 0 | 0 |  |  |
| 150 | `internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | T:.claude/skills | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 151 | `internal/template/templates/.claude/skills/moai/workflows/harness.md` | T:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 152 | `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md` | T:.claude/skills | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 153 | `internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md` | T:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 154 | `internal/template/templates/.claude/skills/moai/workflows/run/context-loading.md` | T:.claude/skills | lead | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 155 | `internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md` | T:.claude/skills | leader | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 156 | `internal/template/templates/.claude/skills/moai/workflows/run/task-decomposition.md` | T:.claude/skills | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 157 | `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | T:.claude/skills | lane | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 158 | `internal/template/templates/.claude/output-styles/moai/moai-easy.md` | T:.claude/output-styles | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 159 | `internal/template/templates/.claude/output-styles/moai/moai.md` | T:.claude/output-styles | lead | 5 | 5 | 0 | 0 | 0 | 0 | 0 | 2 |  |
| 160 | `internal/template/templates/.claude/output-styles/moai/moai.md` | T:.claude/output-styles | lane | 22 | 20 | 0 | 2 | 0 | 0 | 0 | 4 |  |
| 161 | `internal/template/templates/.claude/loop.md` | T:.claude/loop.md | foreman | 3 | 2 | 0 | 1 | 0 | 0 | 0 |  |  |
| 162 | `internal/template/templates/.codex/agents/moai/e2e-tester.toml` | T:.codex | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 163 | `internal/template/templates/.codex/agents/moai/manager-design.toml` | T:.codex | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 164 | `internal/template/templates/.codex/agents/moai/manager-develop.toml` | T:.codex | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 165 | `internal/template/templates/.codex/agents/moai/manager-git.toml` | T:.codex | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 166 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | T:.codex | lead | 64 | 33 | 27 (agent-name 19,cli-token 4,sentinel/env 4) | 2 | 1 (agent-teams/cg 1) | 0 | 1 | 1 |  |
| 167 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | T:.codex | lane | 28 | 18 | 3 (cli-token 1,notation 2) | 1 | 0 | 0 | 6 | 1 |  |
| 168 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | T:.codex | worker | 36 | 21 (sub 18) | 5 (notation 5) | 6 | 0 | 2 | 2 |  |  |
| 169 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | T:.codex | companion | 4 | 3 | 0 | 0 | 0 | 0 | 1 |  |  |
| 170 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | T:.codex | deputy | 26 | 23 | 2 (sentinel/env 2) | 1 | 0 | 0 | 0 | 1 | Y |
| 171 | `internal/template/templates/.codex/agents/moai/manager-lead.toml` | T:.codex | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 172 | `internal/template/templates/.codex/agents/moai/manager-spec.toml` | T:.codex | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 173 | `internal/template/templates/.codex/agents/moai/mission-governor.toml` | T:.codex | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 174 | `internal/template/templates/.codex/agents/moai/plan-auditor.toml` | T:.codex | lead | 2 | 0 | 0 | 0 | 0 | 1 | 1 | 1 |  |
| 175 | `internal/template/templates/.codex/agents/moai/super-advisor.toml` | T:.codex | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 176 | `internal/template/templates/.codex/agents/moai/sync-auditor.toml` | T:.codex | lead | 2 | 0 | 0 | 0 | 0 | 1 | 1 | 1 |  |
| 177 | `internal/template/templates/.codex/agents/moai/sync-auditor.toml` | T:.codex | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 178 | `internal/template/templates/.moai/config/sections/llm.yaml` | T:.moai | lead | 4 | 0 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 179 | `internal/template/templates/.moai/config/sections/lsp.yaml.tmpl` | T:.moai | coordinator | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 180 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | T:.moai | lead | 8 | 7 | 0 | 1 | 0 | 0 | 0 |  |  |
| 181 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | T:.moai | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 182 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | T:.moai | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 183 | `internal/template/templates/AGENTS.md.tmpl` | T:root | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 184 | `internal/template/templates/CLAUDE.md` | T:root | lead | 7 | 1 | 5 (agent-name 5) | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 185 | `internal/template/templates/CLAUDE.md` | T:root | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 186 | `internal/template/templates/CLAUDE.md` | T:root | worker | 1 | 1 (sub 1) | 0 | 0 | 0 | 0 | 0 |  |  |
| 187 | `.claude/rules/local/gitflow-lane-protocol.md` | L:.claude/rules | lane | 5 | 5 | 0 | 0 | 0 | 0 | 0 |  |  |
| 188 | `.claude/rules/local/repo-local-pr-policy.md` | L:.claude/rules | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 189 | `.claude/rules/local/repo-local-pr-policy.md` | L:.claude/rules | lane | 5 | 3 | 0 | 2 | 0 | 0 | 0 | 1 |  |
| 190 | `.claude/rules/local/repo-local-pr-policy.md` | L:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 191 | `.claude/rules/moai/core/agent-common-protocol-reference.md` | L:.claude/rules | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 192 | `.claude/rules/moai/core/agent-common-protocol-reference.md` | L:.claude/rules | worker | 2 | 2 (sub 2) | 0 | 0 | 0 | 0 | 0 |  |  |
| 193 | `.claude/rules/moai/core/agent-common-protocol-reference.md` | L:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 194 | `.claude/rules/moai/core/agent-common-protocol.md` | L:.claude/rules | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 195 | `.claude/rules/moai/core/agent-common-protocol.md` | L:.claude/rules | lane | 15 | 14 | 0 | 1 | 0 | 0 | 0 | 1 |  |
| 196 | `.claude/rules/moai/core/agent-common-protocol.md` | L:.claude/rules | companion | 2 | 1 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 197 | `.claude/rules/moai/core/askuser-protocol-reference.md` | L:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 198 | `.claude/rules/moai/core/askuser-protocol.md` | L:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 199 | `.claude/rules/moai/core/glm-web-tooling.md` | L:.claude/rules | leader | 14 | 2 | 0 | 2 | 10 (agent-teams/cg 10) | 0 | 0 | 1 |  |
| 200 | `.claude/rules/moai/core/moai-constitution-detail.md` | L:.claude/rules | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 201 | `.claude/rules/moai/core/moai-constitution-detail.md` | L:.claude/rules | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 202 | `.claude/rules/moai/core/moai-constitution-detail.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 203 | `.claude/rules/moai/core/moai-constitution.md` | L:.claude/rules | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 204 | `.claude/rules/moai/core/moai-constitution.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 205 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | L:.claude/rules | lead | 10 | 6 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 206 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | L:.claude/rules | lane | 6 | 6 | 0 | 0 | 0 | 0 | 0 |  |  |
| 207 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | L:.claude/rules | worker | 5 | 5 | 0 | 0 | 0 | 0 | 0 |  |  |
| 208 | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 209 | `.claude/rules/moai/core/moai-mcp-tools.md` | L:.claude/rules | lead | 2 | 1 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 210 | `.claude/rules/moai/core/moai-mcp-tools.md` | L:.claude/rules | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 211 | `.claude/rules/moai/core/moai-mcp-tools.md` | L:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 212 | `.claude/rules/moai/core/native-idiom-and-register-detail.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 213 | `.claude/rules/moai/core/output-style-localization-catalogue.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 | 1 |  |
| 214 | `.claude/rules/moai/core/verification-claim-integrity-detail.md` | L:.claude/rules | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 215 | `.claude/rules/moai/core/verification-claim-integrity-detail.md` | L:.claude/rules | companion | 6 | 0 | 0 | 2 | 4 (doc-companion 4) | 0 | 0 | 1 |  |
| 216 | `.claude/rules/moai/core/verification-claim-integrity.md` | L:.claude/rules | companion | 3 | 1 | 0 | 0 | 2 (doc-companion 2) | 0 | 0 | 1 |  |
| 217 | `.claude/rules/moai/development/agent-authoring.md` | L:.claude/rules | lead | 8 | 2 | 3 (agent-name 3) | 1 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 218 | `.claude/rules/moai/development/agent-authoring.md` | L:.claude/rules | worker | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 219 | `.claude/rules/moai/development/agent-authoring.md` | L:.claude/rules | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 220 | `.claude/rules/moai/development/agent-patterns.md` | L:.claude/rules | lead | 7 | 0 | 1 (agent-name 1) | 0 | 6 (agent-teams/cg 6) | 0 | 0 |  |  |
| 221 | `.claude/rules/moai/development/agent-patterns.md` | L:.claude/rules | worker | 8 | 7 | 0 | 0 | 0 | 0 | 1 |  |  |
| 222 | `.claude/rules/moai/development/agent-patterns.md` | L:.claude/rules | coordinator | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 223 | `.claude/rules/moai/development/model-policy.md` | L:.claude/rules | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 224 | `.claude/rules/moai/development/model-policy.md` | L:.claude/rules | leader | 3 | 0 | 0 | 0 | 3 (agent-teams/cg 3) | 0 | 0 |  |  |
| 225 | `.claude/rules/moai/development/model-policy.md` | L:.claude/rules | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 226 | `.claude/rules/moai/development/model-policy.md` | L:.claude/rules | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 227 | `.claude/rules/moai/development/orchestrator-templates.md` | L:.claude/rules | lead | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 228 | `.claude/rules/moai/development/rule-authoring.md` | L:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 229 | `.claude/rules/moai/development/skill-authoring.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 230 | `.claude/rules/moai/development/sprint-round-naming.md` | L:.claude/rules | lane | 15 | 12 | 0 | 0 | 0 | 0 | 3 |  |  |
| 231 | `.claude/rules/moai/languages/cpp.md` | L:.claude/rules | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 232 | `.claude/rules/moai/languages/go.md` | L:.claude/rules | worker | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 233 | `.claude/rules/moai/languages/kotlin.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 234 | `.claude/rules/moai/languages/rust.md` | L:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 235 | `.claude/rules/moai/quality/boundary-verification.md` | L:.claude/rules | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 236 | `.claude/rules/moai/workflow/archived-agent-rejection.md` | L:.claude/rules | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 237 | `.claude/rules/moai/workflow/archived-agent-rejection.md` | L:.claude/rules | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 238 | `.claude/rules/moai/workflow/cache-aware-execution-reference.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 239 | `.claude/rules/moai/workflow/context-window-management-detail.md` | L:.claude/rules | companion | 5 | 1 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 240 | `.claude/rules/moai/workflow/cross-session-messaging-detail.md` | L:.claude/rules | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 241 | `.claude/rules/moai/workflow/cross-session-messaging-detail.md` | L:.claude/rules | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 242 | `.claude/rules/moai/workflow/cross-session-messaging-detail.md` | L:.claude/rules | worker | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 243 | `.claude/rules/moai/workflow/cross-session-messaging-detail.md` | L:.claude/rules | companion | 5 | 1 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 244 | `.claude/rules/moai/workflow/cross-session-messaging.md` | L:.claude/rules | lead | 4 | 2 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 245 | `.claude/rules/moai/workflow/cross-session-messaging.md` | L:.claude/rules | worker | 5 | 4 | 0 | 1 | 0 | 0 | 0 |  |  |
| 246 | `.claude/rules/moai/workflow/cross-session-messaging.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 247 | `.claude/rules/moai/workflow/cross-session-messaging.md` | L:.claude/rules | coordinator | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 248 | `.claude/rules/moai/workflow/goal-directive-detail.md` | L:.claude/rules | companion | 3 | 0 | 0 | 0 | 3 (doc-companion 3) | 0 | 0 |  |  |
| 249 | `.claude/rules/moai/workflow/goal-directive.md` | L:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 250 | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | L:.claude/rules | lead | 79 | 54 | 14 (agent-name 13,sentinel/env 1) | 2 | 2 (agent-teams/cg 2) | 4 | 3 | 2 | Y |
| 251 | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | L:.claude/rules | lane | 57 | 44 | 0 | 8 | 0 | 4 | 1 |  | Y |
| 252 | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | L:.claude/rules | worker | 9 | 5 | 4 (notation 4) | 0 | 0 | 0 | 0 |  |  |
| 253 | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | L:.claude/rules | companion | 20 | 17 | 0 | 0 | 3 (doc-companion 3) | 0 | 0 | 1 |  |
| 254 | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | L:.claude/rules | deputy | 29 | 27 | 1 (sentinel/env 1) | 1 | 0 | 0 | 0 | 1 |  |
| 255 | `.claude/rules/moai/workflow/kanban-dispatch.md` | L:.claude/rules | lead | 87 | 76 | 8 (agent-name 7,sentinel/env 1) | 0 | 0 | 3 | 0 | 20 |  |
| 256 | `.claude/rules/moai/workflow/kanban-dispatch.md` | L:.claude/rules | lane | 57 | 53 | 0 | 4 | 0 | 0 | 0 | 8 | Y |
| 257 | `.claude/rules/moai/workflow/kanban-dispatch.md` | L:.claude/rules | worker | 6 | 2 (sub 1) | 4 (cli-token 1,notation 3) | 0 | 0 | 0 | 0 |  |  |
| 258 | `.claude/rules/moai/workflow/kanban-dispatch.md` | L:.claude/rules | companion | 18 | 13 | 0 | 1 | 4 (doc-companion 4) | 0 | 0 | 3 |  |
| 259 | `.claude/rules/moai/workflow/kanban-dispatch.md` | L:.claude/rules | deputy | 10 | 10 | 0 | 0 | 0 | 0 | 0 | 5 | Y |
| 260 | `.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 261 | `.claude/rules/moai/workflow/moai-memory.md` | L:.claude/rules | lead | 2 | 0 | 0 | 0 | 0 | 2 | 0 |  |  |
| 262 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | L:.claude/rules | lead | 16 | 3 | 10 (agent-name 10) | 1 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 263 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | L:.claude/rules | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 264 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | L:.claude/rules | worker | 10 | 6 (sub 1) | 1 (cli-token 1) | 3 | 0 | 0 | 0 |  |  |
| 265 | `.claude/rules/moai/workflow/resource-slot-lease.md` | L:.claude/rules | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 266 | `.claude/rules/moai/workflow/rule-loading-budget.md` | L:.claude/rules | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 267 | `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 268 | `.claude/rules/moai/workflow/session-handoff-examples.md` | L:.claude/rules | leader | 2 | 1 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 269 | `.claude/rules/moai/workflow/session-handoff-examples.md` | L:.claude/rules | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 270 | `.claude/rules/moai/workflow/session-handoff-examples.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 271 | `.claude/rules/moai/workflow/session-handoff.md` | L:.claude/rules | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 272 | `.claude/rules/moai/workflow/skill-routing-detail.md` | L:.claude/rules | companion | 4 | 0 | 0 | 0 | 4 (doc-companion 4) | 0 | 0 |  |  |
| 273 | `.claude/rules/moai/workflow/spec-workflow.md` | L:.claude/rules | worker | 1 | 0 | 0 | 0 | 0 | 0 | 1 |  |  |
| 274 | `.claude/rules/moai/workflow/team-capability-resolver.md` | L:.claude/rules | worker | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 275 | `.claude/rules/moai/workflow/worktree-integration.md` | L:.claude/rules | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 | 1 |  |
| 276 | `.claude/rules/moai/workflow/worktree-integration.md` | L:.claude/rules | lane | 5 | 4 | 0 | 1 | 0 | 0 | 0 |  |  |
| 277 | `.claude/rules/moai/workflow/worktree-integration.md` | L:.claude/rules | worker | 2 | 2 (sub 1) | 0 | 0 | 0 | 0 | 0 | 1 |  |
| 278 | `.claude/agents/harness/hns-oss-docs-content-author-specialist.md` | L:.claude/agents | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 279 | `.claude/agents/harness/hns-oss-docs-locale-translator-specialist.md` | L:.claude/agents | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 280 | `.claude/agents/harness/hns-oss-docs-locale-translator-specialist.md` | L:.claude/agents | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 281 | `.claude/agents/harness/hns-oss-docs-structure-curator-specialist.md` | L:.claude/agents | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 282 | `.claude/agents/harness/hns-release-specialist.md` | L:.claude/agents | lane | 7 | 3 | 0 | 4 | 0 | 0 | 0 |  |  |
| 283 | `.claude/agents/harness/hns-release-update-specialist.md` | L:.claude/agents | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 284 | `.claude/agents/moai/e2e-tester.md` | L:.claude/agents | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 285 | `.claude/agents/moai/manager-design.md` | L:.claude/agents | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 286 | `.claude/agents/moai/manager-develop.md` | L:.claude/agents | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 287 | `.claude/agents/moai/manager-git.md` | L:.claude/agents | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 288 | `.claude/agents/moai/manager-lead.md` | L:.claude/agents | lead | 64 | 34 | 26 (agent-name 18,cli-token 4,sentinel/env 4) | 2 | 1 (agent-teams/cg 1) | 0 | 1 | 1 |  |
| 289 | `.claude/agents/moai/manager-lead.md` | L:.claude/agents | lane | 29 | 19 | 3 (cli-token 1,notation 2) | 1 | 0 | 0 | 6 | 1 |  |
| 290 | `.claude/agents/moai/manager-lead.md` | L:.claude/agents | worker | 33 | 19 (sub 16) | 5 (notation 5) | 6 | 0 | 1 | 2 |  |  |
| 291 | `.claude/agents/moai/manager-lead.md` | L:.claude/agents | companion | 4 | 3 | 0 | 0 | 0 | 0 | 1 |  |  |
| 292 | `.claude/agents/moai/manager-lead.md` | L:.claude/agents | deputy | 26 | 23 | 2 (sentinel/env 2) | 1 | 0 | 0 | 0 | 1 | Y |
| 293 | `.claude/agents/moai/manager-lead.md` | L:.claude/agents | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 294 | `.claude/agents/moai/manager-spec.md` | L:.claude/agents | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 295 | `.claude/agents/moai/mission-governor.md` | L:.claude/agents | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 296 | `.claude/agents/moai/plan-auditor.md` | L:.claude/agents | lead | 2 | 0 | 0 | 0 | 0 | 1 | 1 | 1 |  |
| 297 | `.claude/agents/moai/super-advisor.md` | L:.claude/agents | leader | 3 | 0 | 0 | 1 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 298 | `.claude/agents/moai/super-advisor.md` | L:.claude/agents | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 299 | `.claude/agents/moai/sync-auditor.md` | L:.claude/agents | lead | 2 | 0 | 0 | 0 | 0 | 1 | 1 | 1 |  |
| 300 | `.claude/agents/moai/sync-auditor.md` | L:.claude/agents | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 301 | `.claude/skills/hns-lsel-curator/SKILL.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 302 | `.claude/skills/hns-lsel-curator/SKILL.md` | L:.claude/skills | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 303 | `.claude/skills/hns-lsel-curator/SKILL.md` | L:.claude/skills | companion | 4 | 2 | 0 | 2 | 0 | 0 | 0 |  |  |
| 304 | `.claude/skills/hns-lsel-curator/drain.sh` | L:.claude/skills | companion | 5 | 4 | 0 | 1 | 0 | 0 | 0 |  |  |
| 305 | `.claude/skills/hns-moaiadk-patterns/SKILL.md` | L:.claude/skills | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 306 | `.claude/skills/hns-oss-docs-readme-sync/SKILL.md` | L:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 307 | `.claude/skills/hns-oss-docs-verify/SKILL.md` | L:.claude/skills | lead | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 308 | `.claude/skills/moai-domain-html-report/SKILL.md` | L:.claude/skills | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 309 | `.claude/skills/moai-domain-humanize/modules/copy-review.md` | L:.claude/skills | lead | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 310 | `.claude/skills/moai-domain-humanize/modules/english.md` | L:.claude/skills | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 311 | `.claude/skills/moai-domain-svg-infographic/SKILL.md` | L:.claude/skills | lane | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 312 | `.claude/skills/moai-domain-svg-infographic/references/archetypes.md` | L:.claude/skills | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 313 | `.claude/skills/moai-domain-svg-infographic/references/authoring.md` | L:.claude/skills | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 314 | `.claude/skills/moai-domain-svg-infographic/references/authoring.md` | L:.claude/skills | leader | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 315 | `.claude/skills/moai-foundation-cc/SKILL.md` | L:.claude/skills | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 316 | `.claude/skills/moai-foundation-cc/SKILL.md` | L:.claude/skills | worker | 5 | 4 (sub 3) | 0 | 1 | 0 | 0 | 0 |  |  |
| 317 | `.claude/skills/moai-foundation-cc/reference/advanced-agent-patterns.md` | L:.claude/skills | lead | 3 | 2 | 0 | 0 | 0 | 1 | 0 |  |  |
| 318 | `.claude/skills/moai-foundation-cc/reference/advanced-agent-patterns.md` | L:.claude/skills | worker | 6 | 5 (sub 2) | 0 | 1 | 0 | 0 | 0 |  |  |
| 319 | `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` | L:.claude/skills | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 320 | `.claude/skills/moai-foundation-core/modules/agents-reference.md` | L:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 321 | `.claude/skills/moai-foundation-thinking/SKILL.md` | L:.claude/skills | lead | 2 | 0 | 0 | 0 | 0 | 2 | 0 |  |  |
| 322 | `.claude/skills/moai-harness-learner/SKILL.md` | L:.claude/skills | coordinator | 3 | 2 | 0 | 0 | 0 | 0 | 1 |  |  |
| 323 | `.claude/skills/moai-kanban-foreman/SKILL.md` | L:.claude/skills | lane | 5 | 2 | 0 | 3 | 0 | 0 | 0 |  |  |
| 324 | `.claude/skills/moai-kanban-foreman/SKILL.md` | L:.claude/skills | worker | 18 | 17 | 0 | 1 | 0 | 0 | 0 |  |  |
| 325 | `.claude/skills/moai-kanban-foreman/SKILL.md` | L:.claude/skills | foreman | 11 | 10 | 0 | 1 | 0 | 0 | 0 |  |  |
| 326 | `.claude/skills/moai-meta-harness/SKILL.md` | L:.claude/skills | companion | 5 | 4 | 0 | 0 | 0 | 0 | 1 |  |  |
| 327 | `.claude/skills/moai-meta-harness/references/seven-phase-workflow.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 328 | `.claude/skills/moai-meta-harness/references/seven-phase-workflow.md` | L:.claude/skills | worker | 3 | 3 (sub 1) | 0 | 0 | 0 | 0 | 0 |  |  |
| 329 | `.claude/skills/moai-ref-ui-polish/references/motion-principles.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 330 | `.claude/skills/moai-workflow-project/references/navigator.md` | L:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 331 | `.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` | L:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 332 | `.claude/skills/moai-workflow-project/templates/question-templates/spec-workflow-setup.json` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 333 | `.claude/skills/moai-workflow-testing/modules/advanced-patterns.md` | L:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 334 | `.claude/skills/moai-workflow-testing/modules/optimization.md` | L:.claude/skills | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 335 | `.claude/skills/moai-workflow-testing/modules/performance-optimization/ai-optimization.md` | L:.claude/skills | worker | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 336 | `.claude/skills/moai-workflow-testing/modules/performance-optimization/bottleneck-detection.md` | L:.claude/skills | worker | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 337 | `.claude/skills/moai-workflow-testing/modules/performance/optimization-patterns.md` | L:.claude/skills | worker | 2 | 1 | 0 | 0 | 0 | 1 | 0 |  |  |
| 338 | `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` | L:.claude/skills | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 339 | `.claude/skills/moai-workflow-worktree/modules/moai-adk-integration.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 340 | `.claude/skills/moai-workflow-worktree/modules/parallel-workflows.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 341 | `.claude/skills/moai-workflow-worktree/modules/tools-integration.md` | L:.claude/skills | worker | 3 | 1 | 0 | 0 | 0 | 2 | 0 |  |  |
| 342 | `.claude/skills/moai-workflow-worktree/modules/worktree-management.md` | L:.claude/skills | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 343 | `.claude/skills/moai/SKILL.md` | L:.claude/skills | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 344 | `.claude/skills/moai/workflows/design.md` | L:.claude/skills | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 345 | `.claude/skills/moai/workflows/e2e.md` | L:.claude/skills | lane | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 346 | `.claude/skills/moai/workflows/factory.md` | L:.claude/skills | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 347 | `.claude/skills/moai/workflows/goal.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 348 | `.claude/skills/moai/workflows/goal.md` | L:.claude/skills | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 349 | `.claude/skills/moai/workflows/gtd.md` | L:.claude/skills | lead | 11 | 9 | 2 (agent-name 1,sentinel/env 1) | 0 | 0 | 0 | 0 | 1 |  |
| 350 | `.claude/skills/moai/workflows/gtd.md` | L:.claude/skills | lane | 6 | 6 | 0 | 0 | 0 | 0 | 0 | 1 |  |
| 351 | `.claude/skills/moai/workflows/gtd.md` | L:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 352 | `.claude/skills/moai/workflows/gtd.md` | L:.claude/skills | foreman | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 353 | `.claude/skills/moai/workflows/harness-build-entry.md` | L:.claude/skills | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 354 | `.claude/skills/moai/workflows/harness-builder.md` | L:.claude/skills | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 355 | `.claude/skills/moai/workflows/harness-builder.md` | L:.claude/skills | worker | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 356 | `.claude/skills/moai/workflows/harness-builder.md` | L:.claude/skills | companion | 10 | 9 | 0 | 1 | 0 | 0 | 0 |  |  |
| 357 | `.claude/skills/moai/workflows/harness-builder.md` | L:.claude/skills | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 358 | `.claude/skills/moai/workflows/harness.md` | L:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 359 | `.claude/skills/moai/workflows/project/doc-generation.md` | L:.claude/skills | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 360 | `.claude/skills/moai/workflows/project/meta-harness.md` | L:.claude/skills | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 361 | `.claude/skills/moai/workflows/run/context-loading.md` | L:.claude/skills | lead | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 362 | `.claude/skills/moai/workflows/run/phase-execution.md` | L:.claude/skills | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 363 | `.claude/skills/moai/workflows/run/task-decomposition.md` | L:.claude/skills | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 364 | `.claude/skills/moai/workflows/sync/delivery.md` | L:.claude/skills | lane | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 365 | `.claude/output-styles/moai/moai-easy.md` | L:.claude/output-styles | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 366 | `.claude/output-styles/moai/moai.md` | L:.claude/output-styles | lead | 5 | 5 | 0 | 0 | 0 | 0 | 0 | 2 |  |
| 367 | `.claude/output-styles/moai/moai.md` | L:.claude/output-styles | lane | 22 | 20 | 0 | 2 | 0 | 0 | 0 | 4 |  |
| 368 | `.claude/commands/harness/oss-docs/manifest.json` | L:.claude/commands | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 369 | `AGENTS.md` | L:root | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 370 | `AGENTS.md` | L:root | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 371 | `CLAUDE.local.md` | L:root | lead | 5 | 0 | 4 (sentinel/env 4) | 1 | 0 | 0 | 0 |  |  |
| 372 | `CLAUDE.local.md` | L:root | lane | 9 | 3 | 2 (notation 1,sentinel/env 1) | 4 | 0 | 0 | 0 | 2 |  |
| 373 | `CLAUDE.local.md` | L:root | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 374 | `CLAUDE.md` | L:root | lead | 7 | 1 | 5 (agent-name 5) | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 375 | `CLAUDE.md` | L:root | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 376 | `CLAUDE.md` | L:root | worker | 1 | 1 (sub 1) | 0 | 0 | 0 | 0 | 0 |  |  |
| 377 | `.moai/config/sections/tool-policy.yaml` | L:.moai/config | leader | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 378 | `.moai/config/sections/workflow.yaml` | L:.moai/config | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 379 | `docs-site/content/en/_index.md` | docs:en | lead | 7 | 4 | 2 (agent-name 2) | 0 | 0 | 0 | 1 |  |  |
| 380 | `docs-site/content/en/_index.md` | docs:en | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 381 | `docs-site/content/en/_index.md` | docs:en | companion | 4 | 3 | 0 | 0 | 0 | 0 | 1 |  |  |
| 382 | `docs-site/content/en/_index.md` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 383 | `docs-site/content/en/advanced/_meta.yaml` | docs:en | lead | 4 | 1 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 384 | `docs-site/content/en/advanced/_meta.yaml` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 385 | `docs-site/content/en/advanced/agent-guide.md` | docs:en | lead | 20 | 0 | 20 (agent-name 16,sentinel/env 4) | 0 | 0 | 0 | 0 |  |  |
| 386 | `docs-site/content/en/advanced/agent-guide.md` | docs:en | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 387 | `docs-site/content/en/advanced/agent-guide.md` | docs:en | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 388 | `docs-site/content/en/advanced/agent-guide.md` | docs:en | worker | 17 | 15 (sub 11) | 0 | 2 | 0 | 0 | 0 |  |  |
| 389 | `docs-site/content/en/advanced/agent-guide.md` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 390 | `docs-site/content/en/advanced/bas-navigator.md` | docs:en | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 391 | `docs-site/content/en/advanced/claude-md-guide.md` | docs:en | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 392 | `docs-site/content/en/advanced/codex-dual-harness.md` | docs:en | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 393 | `docs-site/content/en/advanced/codex-dual-harness.md` | docs:en | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 394 | `docs-site/content/en/advanced/config-sections.md` | docs:en | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 395 | `docs-site/content/en/advanced/factory-mode.md` | docs:en | lead | 28 | 23 | 2 (agent-name 2) | 0 | 0 | 0 | 3 |  |  |
| 396 | `docs-site/content/en/advanced/factory-mode.md` | docs:en | lane | 10 | 1 | 9 (cli-token 1,notation 8) | 0 | 0 | 0 | 0 |  |  |
| 397 | `docs-site/content/en/advanced/factory-mode.md` | docs:en | worker | 82 | 44 | 24 (cli-token 11,notation 13) | 1 | 0 | 0 | 13 |  |  |
| 398 | `docs-site/content/en/advanced/factory-mode.md` | docs:en | companion | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 399 | `docs-site/content/en/advanced/factory-mode.md` | docs:en | foreman | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 400 | `docs-site/content/en/advanced/factory-mode.md` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 401 | `docs-site/content/en/advanced/harness-v4-builder.md` | docs:en | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 402 | `docs-site/content/en/advanced/hooks-guide.md` | docs:en | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 403 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | lead | 31 | 25 | 2 (agent-name 2) | 0 | 0 | 2 | 2 |  |  |
| 404 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | leader | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 405 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | lane | 8 | 6 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 406 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | worker | 15 | 2 | 5 (cli-token 3,notation 2) | 2 | 0 | 0 | 6 |  |  |
| 407 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | companion | 14 | 13 | 0 | 0 | 0 | 0 | 1 |  |  |
| 408 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | foreman | 7 | 5 | 0 | 2 | 0 | 0 | 0 |  |  |
| 409 | `docs-site/content/en/advanced/kanban-mode.md` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 410 | `docs-site/content/en/advanced/manager-lead.md` | docs:en | lead | 72 | 28 | 43 (agent-name 43) | 0 | 0 | 0 | 1 |  |  |
| 411 | `docs-site/content/en/advanced/manager-lead.md` | docs:en | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 412 | `docs-site/content/en/advanced/manager-lead.md` | docs:en | worker | 11 | 9 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 413 | `docs-site/content/en/advanced/manager-lead.md` | docs:en | companion | 4 | 3 | 0 | 1 | 0 | 0 | 0 |  |  |
| 414 | `docs-site/content/en/advanced/manager-lead.md` | docs:en | coordinator | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 415 | `docs-site/content/en/advanced/moai-web-console.md` | docs:en | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 416 | `docs-site/content/en/advanced/moai-web-console.md` | docs:en | lane | 7 | 7 | 0 | 0 | 0 | 0 | 0 |  |  |
| 417 | `docs-site/content/en/advanced/moai-web-console.md` | docs:en | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 418 | `docs-site/content/en/advanced/multi-model-audit.md` | docs:en | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 419 | `docs-site/content/en/advanced/no-haiku-3tier.md` | docs:en | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 420 | `docs-site/content/en/advanced/profile-matrix.md` | docs:en | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 421 | `docs-site/content/en/advanced/skill-guide.md` | docs:en | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 422 | `docs-site/content/en/advanced/skill-guide.md` | docs:en | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 423 | `docs-site/content/en/advanced/skill-guide.md` | docs:en | foreman | 2 | 1 | 0 | 1 | 0 | 0 | 0 |  |  |
| 424 | `docs-site/content/en/advanced/statusline.md` | docs:en | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 425 | `docs-site/content/en/advanced/statusline.md` | docs:en | companion | 1 | 0 | 0 | 0 | 0 | 0 | 1 |  |  |
| 426 | `docs-site/content/en/advanced/tokenomics-overview.md` | docs:en | leader | 2 | 0 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 427 | `docs-site/content/en/claude-code/agentic/_index.md` | docs:en | worker | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 428 | `docs-site/content/en/claude-code/agentic/agent-teams.md` | docs:en | lead | 23 | 2 | 0 | 0 | 18 (agent-teams/cg 18) | 3 | 0 |  |  |
| 429 | `docs-site/content/en/claude-code/agentic/agent-teams.md` | docs:en | leader | 3 | 0 | 0 | 0 | 3 (agent-teams/cg 3) | 0 | 0 |  |  |
| 430 | `docs-site/content/en/claude-code/agentic/agent-teams.md` | docs:en | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 431 | `docs-site/content/en/claude-code/agentic/agent-view.md` | docs:en | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 432 | `docs-site/content/en/claude-code/agentic/scheduled-tasks.md` | docs:en | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 433 | `docs-site/content/en/claude-code/agentic/sub-agents.md` | docs:en | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 434 | `docs-site/content/en/claude-code/agentic/sub-agents.md` | docs:en | worker | 5 | 5 | 0 | 0 | 0 | 0 | 0 |  |  |
| 435 | `docs-site/content/en/claude-code/agentic/workflows.md` | docs:en | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 436 | `docs-site/content/en/claude-code/foundations/features-overview.md` | docs:en | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 437 | `docs-site/content/en/claude-code/foundations/tools-reference.md` | docs:en | worker | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 438 | `docs-site/content/en/cli-reference/graph.md` | docs:en | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 439 | `docs-site/content/en/cli-reference/launchers.md` | docs:en | lead | 10 | 7 | 2 (agent-name 2) | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 440 | `docs-site/content/en/cli-reference/launchers.md` | docs:en | leader | 2 | 0 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 441 | `docs-site/content/en/cli-reference/launchers.md` | docs:en | lane | 3 | 0 | 3 (notation 3) | 0 | 0 | 0 | 0 |  |  |
| 442 | `docs-site/content/en/cli-reference/launchers.md` | docs:en | worker | 27 | 12 | 12 (cli-token 6,notation 6) | 0 | 0 | 0 | 3 |  |  |
| 443 | `docs-site/content/en/cli-reference/launchers.md` | docs:en | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 444 | `docs-site/content/en/cli-reference/launchers.md` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 445 | `docs-site/content/en/cli-reference/tokens.md` | docs:en | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 446 | `docs-site/content/en/cli-reference/tokens.md` | docs:en | worker | 3 | 2 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 447 | `docs-site/content/en/core-concepts/book.md` | docs:en | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 448 | `docs-site/content/en/core-concepts/kanban-board-terms.md` | docs:en | lead | 11 | 11 | 0 | 0 | 0 | 0 | 0 |  |  |
| 449 | `docs-site/content/en/core-concepts/kanban-board-terms.md` | docs:en | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 450 | `docs-site/content/en/core-concepts/kanban-board-terms.md` | docs:en | lane | 15 | 11 | 0 | 1 | 0 | 3 | 0 |  |  |
| 451 | `docs-site/content/en/core-concepts/kanban-board-terms.md` | docs:en | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 452 | `docs-site/content/en/core-concepts/kanban-board-terms.md` | docs:en | companion | 10 | 10 | 0 | 0 | 0 | 0 | 0 |  |  |
| 453 | `docs-site/content/en/core-concepts/what-is-moai-adk.md` | docs:en | lead | 4 | 0 | 2 (agent-name 2) | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 454 | `docs-site/content/en/core-concepts/what-is-moai-adk.md` | docs:en | leader | 2 | 0 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 455 | `docs-site/content/en/core-concepts/what-is-moai-adk.md` | docs:en | worker | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 456 | `docs-site/content/en/cost-optimization/_index.md` | docs:en | lead | 1 | 0 | 0 | 0 | 0 | 1 | 0 |  |  |
| 457 | `docs-site/content/en/cost-optimization/prompt-caching.md` | docs:en | leader | 2 | 0 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 458 | `docs-site/content/en/getting-started/faq.md` | docs:en | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 459 | `docs-site/content/en/getting-started/introduction.md` | docs:en | lead | 3 | 0 | 2 (agent-name 2) | 1 | 0 | 0 | 0 |  |  |
| 460 | `docs-site/content/en/getting-started/introduction.md` | docs:en | leader | 2 | 0 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 461 | `docs-site/content/en/getting-started/windows-guide.md` | docs:en | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 462 | `docs-site/content/en/guides/mcp-server.md` | docs:en | lead | 9 | 5 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 463 | `docs-site/content/en/guides/mcp-server.md` | docs:en | lane | 7 | 7 | 0 | 0 | 0 | 0 | 0 |  |  |
| 464 | `docs-site/content/en/guides/mcp-server.md` | docs:en | worker | 5 | 5 | 0 | 0 | 0 | 0 | 0 |  |  |
| 465 | `docs-site/content/en/multi-llm/_index.md` | docs:en | leader | 3 | 2 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 466 | `docs-site/content/en/multi-llm/_index.md` | docs:en | worker | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 467 | `docs-site/content/en/multi-llm/cg-mode.md` | docs:en | leader | 2 | 0 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 468 | `docs-site/content/en/multi-llm/kanban-mode.md` | docs:en | lead | 16 | 16 | 0 | 0 | 0 | 0 | 0 |  |  |
| 469 | `docs-site/content/en/multi-llm/kanban-mode.md` | docs:en | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 470 | `docs-site/content/en/multi-llm/kanban-mode.md` | docs:en | companion | 20 | 20 | 0 | 0 | 0 | 0 | 0 |  |  |
| 471 | `docs-site/content/en/multi-llm/model-policy.md` | docs:en | lead | 5 | 0 | 4 (agent-name 4) | 0 | 0 | 1 | 0 |  |  |
| 472 | `docs-site/content/en/multi-llm/model-policy.md` | docs:en | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 473 | `docs-site/content/en/utility-commands/moai-codemaps.md` | docs:en | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 474 | `docs-site/content/en/utility-commands/moai-e2e.md` | docs:en | lane | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 475 | `docs-site/content/en/utility-commands/moai-gtd.md` | docs:en | lane | 4 | 3 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 476 | `docs-site/content/en/utility-commands/moai-todo.md` | docs:en | lead | 8 | 8 | 0 | 0 | 0 | 0 | 0 |  |  |
| 477 | `docs-site/content/en/utility-commands/moai-todo.md` | docs:en | worker | 1 | 0 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 478 | `docs-site/content/en/utility-commands/moai-todo.md` | docs:en | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 479 | `docs-site/content/en/utility-commands/moai-todo.md` | docs:en | foreman | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 480 | `docs-site/content/en/utility-commands/moai.md` | docs:en | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 481 | `docs-site/content/en/workflow-commands/moai-goal.md` | docs:en | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 482 | `docs-site/content/en/workflow-commands/moai-harness.md` | docs:en | companion | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 483 | `docs-site/content/en/workflow-commands/moai-run.md` | docs:en | lead | 2 | 1 | 0 | 0 | 0 | 1 | 0 |  |  |
| 484 | `docs-site/content/en/worktree/examples.md` | docs:en | lead | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 485 | `docs-site/content/en/worktree/faq.md` | docs:en | leader | 1 | 0 | 0 | 0 | 1 (agent-teams/cg 1) | 0 | 0 |  |  |
| 486 | `docs-site/content/ko/_index.md` | docs:ko | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 487 | `docs-site/content/ko/advanced/_meta.yaml` | docs:ko | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 488 | `docs-site/content/ko/advanced/agent-guide.md` | docs:ko | lead | 11 | 0 | 11 (agent-name 8,sentinel/env 3) | 0 | 0 | 0 | 0 |  |  |
| 489 | `docs-site/content/ko/advanced/agent-guide.md` | docs:ko | worker | 1 | 1 (sub 1) | 0 | 0 | 0 | 0 | 0 |  |  |
| 490 | `docs-site/content/ko/advanced/bas-navigator.md` | docs:ko | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 491 | `docs-site/content/ko/advanced/claude-md-guide.md` | docs:ko | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 492 | `docs-site/content/ko/advanced/codex-dual-harness.md` | docs:ko | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 493 | `docs-site/content/ko/advanced/codex-dual-harness.md` | docs:ko | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 494 | `docs-site/content/ko/advanced/factory-mode.md` | docs:ko | lead | 5 | 3 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 495 | `docs-site/content/ko/advanced/factory-mode.md` | docs:ko | lane | 9 | 0 | 9 (cli-token 1,notation 8) | 0 | 0 | 0 | 0 |  |  |
| 496 | `docs-site/content/ko/advanced/factory-mode.md` | docs:ko | worker | 29 | 2 | 24 (cli-token 11,notation 13) | 0 | 0 | 0 | 3 |  |  |
| 497 | `docs-site/content/ko/advanced/harness-v4-builder.md` | docs:ko | companion | 4 | 4 | 0 | 0 | 0 | 0 | 0 |  |  |
| 498 | `docs-site/content/ko/advanced/kanban-mode.md` | docs:ko | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 499 | `docs-site/content/ko/advanced/kanban-mode.md` | docs:ko | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 500 | `docs-site/content/ko/advanced/kanban-mode.md` | docs:ko | lane | 2 | 0 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 501 | `docs-site/content/ko/advanced/kanban-mode.md` | docs:ko | worker | 5 | 0 | 5 (cli-token 3,notation 2) | 0 | 0 | 0 | 0 |  |  |
| 502 | `docs-site/content/ko/advanced/kanban-mode.md` | docs:ko | foreman | 3 | 1 | 0 | 2 | 0 | 0 | 0 |  |  |
| 503 | `docs-site/content/ko/advanced/manager-lead.md` | docs:ko | lead | 51 | 9 | 42 (agent-name 42) | 0 | 0 | 0 | 0 |  |  |
| 504 | `docs-site/content/ko/advanced/manager-lead.md` | docs:ko | worker | 2 | 0 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 505 | `docs-site/content/ko/advanced/manager-lead.md` | docs:ko | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 506 | `docs-site/content/ko/advanced/moai-web-console.md` | docs:ko | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 507 | `docs-site/content/ko/advanced/no-haiku-3tier.md` | docs:ko | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 508 | `docs-site/content/ko/advanced/profile-matrix.md` | docs:ko | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 509 | `docs-site/content/ko/advanced/skill-guide.md` | docs:ko | foreman | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 510 | `docs-site/content/ko/claude-code/agentic/agent-teams.md` | docs:ko | lead | 6 | 3 | 0 | 0 | 3 (agent-teams/cg 3) | 0 | 0 |  |  |
| 511 | `docs-site/content/ko/claude-code/agentic/sub-agents.md` | docs:ko | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 512 | `docs-site/content/ko/cli-reference/launchers.md` | docs:ko | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 513 | `docs-site/content/ko/cli-reference/launchers.md` | docs:ko | lane | 3 | 0 | 3 (notation 3) | 0 | 0 | 0 | 0 |  |  |
| 514 | `docs-site/content/ko/cli-reference/launchers.md` | docs:ko | worker | 12 | 0 | 12 (cli-token 6,notation 6) | 0 | 0 | 0 | 0 |  |  |
| 515 | `docs-site/content/ko/cli-reference/tokens.md` | docs:ko | worker | 1 | 0 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 516 | `docs-site/content/ko/core-concepts/kanban-board-terms.md` | docs:ko | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 517 | `docs-site/content/ko/core-concepts/kanban-board-terms.md` | docs:ko | lane | 4 | 3 | 0 | 1 | 0 | 0 | 0 |  |  |
| 518 | `docs-site/content/ko/core-concepts/kanban-board-terms.md` | docs:ko | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 519 | `docs-site/content/ko/core-concepts/what-is-moai-adk.md` | docs:ko | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 520 | `docs-site/content/ko/core-concepts/what-is-moai-adk.md` | docs:ko | worker | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 521 | `docs-site/content/ko/getting-started/faq.md` | docs:ko | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 522 | `docs-site/content/ko/getting-started/introduction.md` | docs:ko | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 523 | `docs-site/content/ko/guides/mcp-server.md` | docs:ko | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 524 | `docs-site/content/ko/multi-llm/model-policy.md` | docs:ko | lead | 4 | 0 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 525 | `docs-site/content/ko/utility-commands/moai-gtd.md` | docs:ko | lane | 3 | 2 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 526 | `docs-site/content/ko/utility-commands/moai-todo.md` | docs:ko | worker | 1 | 0 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 527 | `docs-site/content/ko/utility-commands/moai.md` | docs:ko | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 528 | `docs-site/content/ja/_index.md` | docs:ja | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 529 | `docs-site/content/ja/advanced/_meta.yaml` | docs:ja | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 530 | `docs-site/content/ja/advanced/agent-guide.md` | docs:ja | lead | 20 | 0 | 20 (agent-name 16,sentinel/env 4) | 0 | 0 | 0 | 0 |  |  |
| 531 | `docs-site/content/ja/advanced/agent-guide.md` | docs:ja | worker | 2 | 1 (sub 1) | 0 | 1 | 0 | 0 | 0 |  |  |
| 532 | `docs-site/content/ja/advanced/bas-navigator.md` | docs:ja | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 533 | `docs-site/content/ja/advanced/claude-md-guide.md` | docs:ja | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 534 | `docs-site/content/ja/advanced/codex-dual-harness.md` | docs:ja | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 535 | `docs-site/content/ja/advanced/codex-dual-harness.md` | docs:ja | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 536 | `docs-site/content/ja/advanced/factory-mode.md` | docs:ja | lead | 5 | 3 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 537 | `docs-site/content/ja/advanced/factory-mode.md` | docs:ja | lane | 9 | 0 | 9 (cli-token 1,notation 8) | 0 | 0 | 0 | 0 |  |  |
| 538 | `docs-site/content/ja/advanced/factory-mode.md` | docs:ja | worker | 29 | 2 | 24 (cli-token 11,notation 13) | 0 | 0 | 0 | 3 |  |  |
| 539 | `docs-site/content/ja/advanced/hooks-guide.md` | docs:ja | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 540 | `docs-site/content/ja/advanced/kanban-mode.md` | docs:ja | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 541 | `docs-site/content/ja/advanced/kanban-mode.md` | docs:ja | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 542 | `docs-site/content/ja/advanced/kanban-mode.md` | docs:ja | lane | 2 | 0 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 543 | `docs-site/content/ja/advanced/kanban-mode.md` | docs:ja | worker | 5 | 0 | 5 (cli-token 3,notation 2) | 0 | 0 | 0 | 0 |  |  |
| 544 | `docs-site/content/ja/advanced/kanban-mode.md` | docs:ja | foreman | 3 | 1 | 0 | 2 | 0 | 0 | 0 |  |  |
| 545 | `docs-site/content/ja/advanced/manager-lead.md` | docs:ja | lead | 50 | 9 | 41 (agent-name 41) | 0 | 0 | 0 | 0 |  |  |
| 546 | `docs-site/content/ja/advanced/manager-lead.md` | docs:ja | worker | 2 | 0 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 547 | `docs-site/content/ja/advanced/manager-lead.md` | docs:ja | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 548 | `docs-site/content/ja/advanced/moai-web-console.md` | docs:ja | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 549 | `docs-site/content/ja/advanced/no-haiku-3tier.md` | docs:ja | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 550 | `docs-site/content/ja/advanced/profile-matrix.md` | docs:ja | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 551 | `docs-site/content/ja/advanced/skill-guide.md` | docs:ja | foreman | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 552 | `docs-site/content/ja/claude-code/agentic/agent-teams.md` | docs:ja | lead | 5 | 3 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 553 | `docs-site/content/ja/claude-code/agentic/sub-agents.md` | docs:ja | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 554 | `docs-site/content/ja/cli-reference/launchers.md` | docs:ja | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 555 | `docs-site/content/ja/cli-reference/launchers.md` | docs:ja | lane | 3 | 0 | 3 (notation 3) | 0 | 0 | 0 | 0 |  |  |
| 556 | `docs-site/content/ja/cli-reference/launchers.md` | docs:ja | worker | 12 | 0 | 12 (cli-token 6,notation 6) | 0 | 0 | 0 | 0 |  |  |
| 557 | `docs-site/content/ja/cli-reference/tokens.md` | docs:ja | worker | 1 | 0 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 558 | `docs-site/content/ja/core-concepts/kanban-board-terms.md` | docs:ja | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 559 | `docs-site/content/ja/core-concepts/kanban-board-terms.md` | docs:ja | lane | 4 | 3 | 0 | 1 | 0 | 0 | 0 |  |  |
| 560 | `docs-site/content/ja/core-concepts/kanban-board-terms.md` | docs:ja | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 561 | `docs-site/content/ja/core-concepts/what-is-moai-adk.md` | docs:ja | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 562 | `docs-site/content/ja/getting-started/faq.md` | docs:ja | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 563 | `docs-site/content/ja/getting-started/introduction.md` | docs:ja | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 564 | `docs-site/content/ja/guides/mcp-server.md` | docs:ja | lead | 4 | 0 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 565 | `docs-site/content/ja/multi-llm/model-policy.md` | docs:ja | lead | 4 | 0 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 566 | `docs-site/content/ja/utility-commands/moai-gtd.md` | docs:ja | lane | 3 | 2 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 567 | `docs-site/content/ja/utility-commands/moai-todo.md` | docs:ja | worker | 1 | 0 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 568 | `docs-site/content/ja/utility-commands/moai.md` | docs:ja | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 569 | `docs-site/content/zh/_index.md` | docs:zh | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 570 | `docs-site/content/zh/advanced/_meta.yaml` | docs:zh | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 571 | `docs-site/content/zh/advanced/agent-guide.md` | docs:zh | lead | 20 | 0 | 20 (agent-name 16,sentinel/env 4) | 0 | 0 | 0 | 0 |  |  |
| 572 | `docs-site/content/zh/advanced/agent-guide.md` | docs:zh | worker | 2 | 1 (sub 1) | 0 | 1 | 0 | 0 | 0 |  |  |
| 573 | `docs-site/content/zh/advanced/bas-navigator.md` | docs:zh | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 574 | `docs-site/content/zh/advanced/claude-md-guide.md` | docs:zh | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 575 | `docs-site/content/zh/advanced/codex-dual-harness.md` | docs:zh | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 576 | `docs-site/content/zh/advanced/codex-dual-harness.md` | docs:zh | companion | 1 | 0 | 0 | 0 | 1 (doc-companion 1) | 0 | 0 |  |  |
| 577 | `docs-site/content/zh/advanced/factory-mode.md` | docs:zh | lead | 5 | 3 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 578 | `docs-site/content/zh/advanced/factory-mode.md` | docs:zh | lane | 9 | 0 | 9 (cli-token 1,notation 8) | 0 | 0 | 0 | 0 |  |  |
| 579 | `docs-site/content/zh/advanced/factory-mode.md` | docs:zh | worker | 29 | 2 | 24 (cli-token 11,notation 13) | 0 | 0 | 0 | 3 |  |  |
| 580 | `docs-site/content/zh/advanced/harness-v4-builder.md` | docs:zh | companion | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 581 | `docs-site/content/zh/advanced/hooks-guide.md` | docs:zh | leader | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 582 | `docs-site/content/zh/advanced/kanban-mode.md` | docs:zh | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 583 | `docs-site/content/zh/advanced/kanban-mode.md` | docs:zh | leader | 1 | 0 | 0 | 0 | 0 | 0 | 1 |  |  |
| 584 | `docs-site/content/zh/advanced/kanban-mode.md` | docs:zh | lane | 2 | 0 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 585 | `docs-site/content/zh/advanced/kanban-mode.md` | docs:zh | worker | 5 | 0 | 5 (cli-token 3,notation 2) | 0 | 0 | 0 | 0 |  |  |
| 586 | `docs-site/content/zh/advanced/kanban-mode.md` | docs:zh | foreman | 3 | 1 | 0 | 2 | 0 | 0 | 0 |  |  |
| 587 | `docs-site/content/zh/advanced/manager-lead.md` | docs:zh | lead | 51 | 8 | 43 (agent-name 43) | 0 | 0 | 0 | 0 |  |  |
| 588 | `docs-site/content/zh/advanced/manager-lead.md` | docs:zh | worker | 2 | 0 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 589 | `docs-site/content/zh/advanced/moai-web-console.md` | docs:zh | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 590 | `docs-site/content/zh/advanced/no-haiku-3tier.md` | docs:zh | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 591 | `docs-site/content/zh/advanced/profile-matrix.md` | docs:zh | lead | 3 | 0 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 592 | `docs-site/content/zh/advanced/skill-guide.md` | docs:zh | foreman | 1 | 0 | 0 | 1 | 0 | 0 | 0 |  |  |
| 593 | `docs-site/content/zh/claude-code/agentic/agent-teams.md` | docs:zh | lead | 5 | 3 | 0 | 0 | 2 (agent-teams/cg 2) | 0 | 0 |  |  |
| 594 | `docs-site/content/zh/claude-code/agentic/sub-agents.md` | docs:zh | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 595 | `docs-site/content/zh/cli-reference/graph.md` | docs:zh | lane | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 596 | `docs-site/content/zh/cli-reference/launchers.md` | docs:zh | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 597 | `docs-site/content/zh/cli-reference/launchers.md` | docs:zh | lane | 3 | 0 | 3 (notation 3) | 0 | 0 | 0 | 0 |  |  |
| 598 | `docs-site/content/zh/cli-reference/launchers.md` | docs:zh | worker | 12 | 0 | 12 (cli-token 6,notation 6) | 0 | 0 | 0 | 0 |  |  |
| 599 | `docs-site/content/zh/cli-reference/tokens.md` | docs:zh | worker | 3 | 2 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 600 | `docs-site/content/zh/core-concepts/kanban-board-terms.md` | docs:zh | lead | 3 | 3 | 0 | 0 | 0 | 0 | 0 |  |  |
| 601 | `docs-site/content/zh/core-concepts/kanban-board-terms.md` | docs:zh | lane | 4 | 3 | 0 | 1 | 0 | 0 | 0 |  |  |
| 602 | `docs-site/content/zh/core-concepts/kanban-board-terms.md` | docs:zh | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 603 | `docs-site/content/zh/core-concepts/what-is-moai-adk.md` | docs:zh | lead | 3 | 1 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 604 | `docs-site/content/zh/getting-started/faq.md` | docs:zh | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 605 | `docs-site/content/zh/getting-started/introduction.md` | docs:zh | lead | 2 | 0 | 2 (agent-name 2) | 0 | 0 | 0 | 0 |  |  |
| 606 | `docs-site/content/zh/guides/mcp-server.md` | docs:zh | lead | 4 | 0 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 607 | `docs-site/content/zh/multi-llm/model-policy.md` | docs:zh | lead | 4 | 0 | 4 (agent-name 4) | 0 | 0 | 0 | 0 |  |  |
| 608 | `docs-site/content/zh/utility-commands/moai-codemaps.md` | docs:zh | worker | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 609 | `docs-site/content/zh/utility-commands/moai-gtd.md` | docs:zh | lane | 3 | 2 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 610 | `docs-site/content/zh/utility-commands/moai-todo.md` | docs:zh | lead | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 611 | `docs-site/content/zh/utility-commands/moai-todo.md` | docs:zh | worker | 1 | 0 | 1 (notation 1) | 0 | 0 | 0 | 0 |  |  |
| 612 | `docs-site/content/zh/utility-commands/moai.md` | docs:zh | lead | 1 | 0 | 1 (agent-name 1) | 0 | 0 | 0 | 0 |  |  |
| 613 | `README.ja.md` | README | lead | 4 | 1 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 614 | `README.ja.md` | README | lane | 4 | 2 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 615 | `README.ja.md` | README | worker | 11 | 0 | 10 (cli-token 5,notation 5) | 0 | 0 | 0 | 1 |  |  |
| 616 | `README.ja.md` | README | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 617 | `README.ko.md` | README | lead | 4 | 1 | 3 (agent-name 3) | 0 | 0 | 0 | 0 |  |  |
| 618 | `README.ko.md` | README | lane | 4 | 2 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 619 | `README.ko.md` | README | worker | 12 | 0 | 11 (cli-token 5,notation 6) | 0 | 0 | 0 | 1 |  |  |
| 620 | `README.ko.md` | README | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 621 | `README.md` | README | lead | 26 | 18 | 3 (agent-name 3) | 1 | 1 (agent-teams/cg 1) | 0 | 3 |  |  |
| 622 | `README.md` | README | leader | 6 | 2 | 0 | 0 | 4 (agent-teams/cg 4) | 0 | 0 |  |  |
| 623 | `README.md` | README | lane | 8 | 6 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 624 | `README.md` | README | worker | 31 | 8 | 10 (cli-token 5,notation 5) | 0 | 0 | 0 | 13 |  |  |
| 625 | `README.md` | README | companion | 11 | 10 | 0 | 0 | 0 | 0 | 1 |  |  |
| 626 | `README.md` | README | foreman | 2 | 2 | 0 | 0 | 0 | 0 | 0 |  |  |
| 627 | `README.md` | README | coordinator | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |
| 628 | `README.zh.md` | README | lead | 6 | 2 | 3 (agent-name 3) | 0 | 0 | 0 | 1 |  |  |
| 629 | `README.zh.md` | README | lane | 4 | 2 | 2 (notation 2) | 0 | 0 | 0 | 0 |  |  |
| 630 | `README.zh.md` | README | worker | 12 | 0 | 10 (cli-token 5,notation 5) | 0 | 0 | 0 | 2 |  |  |
| 631 | `README.zh.md` | README | companion | 1 | 1 | 0 | 0 | 0 | 0 | 0 |  |  |

---

## 2. 합계

### 2.1 토큰 × 의미 분류 (영문 토큰, 문서 계층 전체, 376 개 파일)

측정: `python3 raw/scripts/inv.py …` → `raw/inv-summary.txt` 의 `CLASS` 블록.

| 토큰 | 합계 | role | ident | compound | other-meaning | plain-en | frozen |
|---|---|---|---|---|---|---|---|
| lead | 1450 | 711 | 577 (agent-name 528 · sentinel/env 37 · cli-token 12) | 20 | 65 (agent-teams/cg) | 51 | 26 |
| leader | 87 | 22 | 0 | 7 | 57 (agent-teams/cg) | 0 | 1 |
| lane | 678 | 503 | 79 (notation 71 · cli-token 7 · sentinel 1) | 59 | 0 | 11 | 26 |
| worker | 756 | 296 + subagent-worker 89 | 254 (notation 150 · cli-token 104) | 41 | 0 | 17 | 59 |
| companion | 391 | 250 | 0 | 12 | 120 (doc-companion) | 0 | 9 |
| foreman | 53 | 38 | 0 | 15 | 0 | 0 | 0 |
| deputy | 156 | 143 | 8 | 5 | 0 | 0 | 0 |
| coordinator | 44 | 42 | 1 (agent-name) | 0 | 0 | 0 | 1 |

`role` 열이 곧 치환 대상이라는 뜻은 아니다. 분류는 매치 단위 휴리스틱이며, 알려진 오분류를 §4 에 적었다. run 단계는 줄 단위로 처분을 기록해야 한다.

### 2.2 표면 × 토큰 (영문 토큰 출현 수 — 「파일」 열은 그 표면에서 토큰이 하나라도 나온 파일 수)

측정: `raw/per-file-pivot.tsv` 를 표면별로 합산.

| 표면 | 파일 | lead | leader | lane | worker | companion | foreman | deputy | coordinator |
|---|---|---|---|---|---|---|---|---|---|
| T:.claude/rules | 44 | 230 | 9 | 162 | 66 | 103 | 0 | 39 | 8 |
| T:.claude/agents | 10 | 70 | 1 | 30 | 38 | 4 | 0 | 26 | 1 |
| T:.claude/skills | 43 | 36 | 4 | 22 | 46 | 25 | 12 | 0 | 6 |
| T:.claude/output-styles | 2 | 5 | 0 | 22 | 0 | 3 | 0 | 0 | 0 |
| T:.claude/loop.md | 1 | 0 | 0 | 0 | 0 | 0 | 3 | 0 | 0 |
| T:.codex (C3, 기계 방출) | 10 | 71 | 1 | 30 | 38 | 4 | 0 | 26 | 1 |
| T:.moai | 3 | 12 | 0 | 2 | 0 | 1 | 0 | 0 | 1 |
| T:root (`CLAUDE.md`, `AGENTS.md.tmpl`) | 2 | 7 | 1 | 1 | 1 | 0 | 0 | 0 | 0 |
| L:.claude/rules | 48 | 231 | 20 | 172 | 68 | 105 | 0 | 39 | 8 |
| L:.claude/agents | 15 | 71 | 5 | 38 | 36 | 11 | 0 | 26 | 1 |
| L:.claude/skills | 49 | 40 | 5 | 23 | 48 | 36 | 12 | 0 | 6 |
| L:.claude/output-styles | 2 | 5 | 0 | 22 | 0 | 3 | 0 | 0 | 0 |
| L:.claude/commands | 1 | 0 | 0 | 0 | 1 | 0 | 0 | 0 | 0 |
| L:root (`CLAUDE.md`, `AGENTS.md`, `CLAUDE.local.md`) | 3 | 13 | 1 | 10 | 2 | 0 | 0 | 0 | 0 |
| L:.moai/config | 2 | 3 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| docs:en | 53 | 272 | 28 | 60 | 187 | 69 | 12 | 0 | 9 |
| docs:ko | 27 | 108 | 1 | 21 | 52 | 6 | 4 | 0 | 1 |
| docs:ja | 27 | 117 | 2 | 21 | 52 | 2 | 4 | 0 | 1 |
| docs:zh | 30 | 119 | 2 | 22 | 55 | 5 | 4 | 0 | 0 |
| README (4) | 4 | 40 | 6 | 20 | 66 | 14 | 2 | 0 | 1 |

ko/ja/zh 의 영문 토큰 수는 본문에 섞인 식별자·코드 조각만 센 것이다. 해당 로케일의 역할 단어는 §3 에 따로 있다. 그래서 en 과 나머지 로케일의 영문 토큰 수가 다른 것은 결함 신호가 아니다.

### 2.3 리드 기준선 재현

리드가 준 기준선(develop `e62c3e183`, `internal/template/templates/.claude` 아래 worker 29 · lane 23 · lead 35 · leader 8)은 **파일 수**이며, 다음 명령으로 같은 값이 재현됐다.

```text
$ for t in worker lane lead leader; do grep -rlwi "$t" internal/template/templates/.claude | wc -l; done
worker files(-rlwi)=29   lane=23   lead=35   leader=8
```

대소문자를 구분하면(`-rlw`) worker 28 · lead 31 로 줄고, 줄 수(`-rhw`)는 worker 75 · lane 117 · lead 215 · leader 14 다. §1·§2 의 수치는 매치 단위라 이 값들과 직접 비교되지 않는다.

---

## 3. 비영어 로케일의 역할 단어

측정: `python3 raw/scripts/cjk2.py …` → `raw/cjk-summary.txt`, 파일별은 `raw/cjk-per-file.tsv`.

| 역할 개념 | ko (docs · README) | ja (docs · README) | zh (docs · README) |
|---|---|---|---|
| lead | 리드 132 · 20 | リード 130 · 23 | 主导 38 · 1, 主控 65 · 22, 领导 32 · 2, 负责人 11 · 0 |
| leader | 리더 57 · 12 | リーダー 51 · 8 | (전용 어휘 없음 — 领导/主控 과 겹침) |
| lane | 레인 44 · 6 | レーン 44 · 6 | 泳道 39 · 6, 通道 26 · 0 |
| worker | 워커 111 · 20 | ワーカー 131 · 19 | 工作者 125 · 17 |
| companion | 동반 세션 38 · 10 | コンパニオン 32 · 10 | 伴随 46 · 10 |
| foreman | 포어맨 9 · 0 | フォアマン 9 · 0 | 工头 9 · 2 |
| coordinator | 코디네이터 7 · 1 | コーディネーター 8 · 1 | 协调者 9 · 2 |

로컬 한국어 표면: `CLAUDE.local.md` 와 `.claude/rules/local/gitflow-lane-protocol.md` 합계 리드 42 · 레인 66. 그중 `CLAUDE.local.md` §4.1(57 행)은 리드 10 · 레인 8 · 워커 0 · `[HARD]` 6.

측정상 주의 하나. BSD `grep -rhoE '(리드|리더|…)'` 처럼 한글 대안을 한 정규식에 묶으면 리드가 **408** 로 나왔다. 단일 토큰 `grep -rho '리드' … | wc -l` 과 Python `str.count` 는 둘 다 **152**(docs 132 + README 20)였다. 다중 바이트 대안 묶음 결과는 이 문서의 근거로 쓰지 않았다.

zh 는 한 역할에 여러 어휘가 흩어져 있다(lead 에 主导·主控·领导·负责人). `leader` 를 도입하면 zh 에서 lead/leader 를 가를 어휘부터 정해야 한다(§11 Q6).

---

## 4. 동음이의 — 기계 치환이 깨뜨리는 뜻

통일 대상 단어들은 문서 계층에서 이미 다른 뜻으로 쓰이고 있다. 일괄 치환을 금지하는 근거다.

| 단어 | 역할 외 뜻 | 측정값 | 예 |
|---|---|---|---|
| leader | `moai cg` 의 Claude leader pane, Agent Teams 의 leader | other-meaning 57 / 전체 87 | `glm-web-tooling.md` §「cg-leader exception」(로컬 14회), `CLAUDE.md` §15, CC 2.1.234 「leader 에게서 /model 상속」 |
| lead | Agent Teams team lead, 동사 「lead to」 | agent-teams 65 · plain 51 | `orchestration-mode-selection.md` §C.1 「one team per session with a fixed lead」 |
| lane | 명령 병렬 묶음 「Lane A / Lane B」 | 템플릿 18 · 로컬 18 (`grep -rhoiE '\blanes? [AB]\b\|two-lane\|both lanes\|the lanes have'`) | `agent-common-protocol.md` Pre-Spawn Sync Check 「two-lane batch」 — 분류기는 이것을 `role` 로 셌다(오분류) |
| lane | Epic 하위 트랙 「Epic N Lane A」 | 템플릿 9 (`sprint-round-naming.md` 등) | `sprint-round-naming.md` 표 「Renamed (Lane retained)」 — 역시 `role` 로 오분류 |
| worker | `manager-lead` 의 leaf worker(서브에이전트) | subagent-worker 89 | `manager_lead_depth_test.go` 가 `<!-- manager-lead leaf-worker -->` 마커를 읽는다 |
| worker | service/web worker, worker pool | plain 17 | 테스트 스킬 모듈 |
| companion | 「detail companion」 문서 짝 | 120 | 거의 모든 규칙 스텁의 `> Detail companion:` 줄 |
| leader | 식별자 「leader socket」 | 6 (`grep -rhoiE 'leader socket'`) | `kanban-board-terms.md` 「It lives in the leader socket path」 |

정리하면, `leader` 는 이미 cg/Agent Teams 가 쓰는 단어라, 칸반·팩토리 lead 를 leader 로 바꾸면 문서 안에 **서로 다른 두 leader** 가 생긴다. 반대로 worker → lane 은 leaf worker(서브에이전트)와의 충돌을 없애는 방향이다.

---

## 5. 로컬 ↔ 템플릿 미러

측정: `raw/mirror-pairs.tsv` (토큰을 가진 템플릿 `.claude/**` 파일마다 로컬 짝의 존재·바이트 동일 여부).

| 상태 | 개수 |
|---|---|
| 짝 있음 · 바이트 동일 | 56 |
| 짝 있음 · 내용 다름 | 44 |
| 로컬 전용(템플릿 짝 없음) | 15 |

역할 명칭이 몰려 있는 주요 파일:

| 파일 | 짝 상태 | 비고 |
|---|---|---|
| `rules/moai/workflow/kanban-dispatch.md` | 동일 | lead 87 · lane 57 · companion 18 · deputy 10, HARD 23 줄 |
| `rules/moai/workflow/kanban-dispatch-detail.md` | 동일 | lead 79 · lane 57 · deputy 29, HARD 3 줄 |
| `agents/moai/manager-lead.md` | **다름** | 템플릿/로컬 lead 63/64 · lane 28/29 · worker 36/33 |
| `rules/moai/core/agent-common-protocol.md` | **다름** | 「Lane sessions are orchestrator-class」 문단(27행)은 양쪽 모두 있음 |
| `rules/moai/workflow/cross-session-messaging{,-detail}.md` | **다름** | |
| `rules/moai/workflow/orchestration-mode-selection.md` | **다름** | lead 16 · worker 10, §C.4 「Factory workers (default 8)」 |
| `rules/moai/core/moai-constitution.md` | **다름** | 11행 「A factory lane or kanban companion session is an orchestrator for its card」 |
| `rules/moai/core/glm-web-tooling.md` | **다름** | leader 3/14 — cg leader 뜻 |
| `skills/moai-kanban-foreman/SKILL.md` | 동일 | worker 18 · foreman 11 |
| `skills/moai/workflows/gtd.md` | 동일 | lead 11 · lane 6 |
| `output-styles/moai/moai.md` | 동일 | lane 22 (Lane Board, `[HARD]` 4 줄) |

C3(`internal/template/templates/.codex/agents/moai/*.toml`) 10 개는 C2 에서 `make agents-emit` 으로 재생성되는 사본이다. 손편집 대상이 아니며, C2 수정 뒤 재방출 → `make agents-emit-check` 로 검증한다(`CLAUDE.local.md` §2.0).

로컬 전용 15 개 중 역할 단어를 가진 것: `.claude/rules/local/gitflow-lane-protocol.md`(lane 5 + 한국어), `.claude/rules/local/repo-local-pr-policy.md`(lead 1 · lane 5 · worker 1), `.claude/agents/harness/hns-release-specialist.md`(lane 7) 등. 템플릿 미러가 없으므로 템플릿 우선 규칙의 대상이 아니고, 로컬에서만 고친다.

---

## 6. HARD 조항과 앵커

### 6.1 토큰을 담은 `[HARD]` 줄 — 87 줄

측정: `raw/hard-lines.tsv`. 파일별:

| 줄 수 | 파일 |
|---|---|
| 23 · 23 | `kanban-dispatch.md` (템플릿 · 로컬) |
| 4 · 4 | `output-styles/moai/moai.md` (Lane Board) |
| 3 · 3 | `kanban-dispatch-detail.md` |
| 2 · 1 | `verification-claim-integrity-detail.md` (템플릿 · 로컬 — 수가 다름) |
| 2 | `CLAUDE.local.md` |
| 1 씩 | `manager-lead.md`(템플릿·로컬·toml), `plan-auditor`/`sync-auditor`(md·toml), `agent-common-protocol.md`, `glm-web-tooling.md`, `worktree-integration.md`, `verification-claim-integrity.md`, `output-style-localization-catalogue.md`, `workflows/gtd.md`, `repo-local-pr-policy.md`(로컬 전용) |

이 조항들은 이름만 바꿔도 뜻이 흔들릴 수 있다. 예: 「The lead is the queue's sole producer」, 「Promotion is the operator's act, always」, 「The deputy never holds a power of consequence」, 「A lane's own claim is not an observation」. 명칭 치환은 이 불변식의 **주어만** 바꿔야 하고 조건·금지·권한 범위는 한 글자도 바뀌면 안 된다.

### 6.2 참조되는 앵커

측정 1: `raw/anchors.tsv` — 토큰을 담은 제목 133 개 중 11 개가 다른 파일에서 참조됨(참조 파일 합 45).
측정 2: `raw/section-refs.txt` — 저장소 전체의 `§ …` 참조 중 토큰을 담은 것(제목이 아닌 굵은 글씨 문단 앵커도 잡힌다).

| 앵커 | 정의 위치 | 참조 수 |
|---|---|---|
| `§ The lead works through manager-lead` | `kanban-dispatch-detail.md:123` | 9 + 3 (괄호 꼬리 형태) |
| `§ Factory in-lane 3-stage` | `kanban-dispatch-detail.md:180` | 9 |
| `§ Lane spawn authority` | `kanban-dispatch.md:268` (굵은 문단, 제목 아님) | 6 + 7 |
| `§ Deputy dispatch surface` | `kanban-dispatch.md:104`, `manager-lead.md:198/200` | 3 + 3 + 3 |
| `§ Verification load is lane-local` | `kanban-dispatch.md:211` | 3 |
| `§C.4 Factory workers (default 8)` / `factory-workers reconciliation` | `orchestration-mode-selection.md` | 3 + 3 |
| `§G.2 manager-lead non-regression note` | SPEC 기록 | 3 |

앵커를 바꾸면 정의·참조 쌍을 같은 커밋에서 바꿔야 한다. 두 번째 측정이 잡는 굵은 문단 앵커(`Lane spawn authority`)는 제목 스캔에 걸리지 않으므로, run 단계의 앵커 검사는 두 측정을 모두 다시 돌려야 한다.

---

## 7. 문서 문자열을 고정하는 테스트 (코드 계층과의 접점)

문서만 고쳐도 Go 테스트가 깨질 수 있다. 문서를 읽고 문자열을 단언하는 테스트 후보:

| 테스트 | 고정하는 것 |
|---|---|
| `internal/template/manager_lead_depth_test.go` | 파일명 `manager-lead.md`, 마커 `<!-- manager-lead leaf-worker -->`, `leaf_of: manager-lead` (문자열 31회) |
| `internal/template/evidence_citation_guard_test.go` | `agents/moai/manager-lead.md` 경로, 「detail companion」 문구 |
| `internal/template/backlog_json_disclosure_mirror_test.go` | `.claude/skills/moai-kanban-foreman/SKILL.md` 경로 |
| `internal/kanban/foreman_queue_statement_test.go`, `foreman_queue_watch_test.go` | foreman 스킬 경로(로컬·템플릿 양쪽) |
| `internal/template/cg_retirement_test.go` | 은퇴 문구 「Leader performs evaluation inline」 부재 |
| `internal/template/rule_template_mirror_test.go` | 허용 목록에 든 규칙의 로컬/템플릿 바이트 동일 |
| `internal/template/agentemit/golden_test.go` | C2 → C3 방출 결과 |

넓은 후보 목록(문서 경로를 언급하고 토큰 문자열 리터럴을 가진 `_test.go` 37 개)은 `raw/doc-pinning-tests.txt` 에 있다. 이 중 대부분은 코드 계층 어휘(`worker-N` 슬롯 등)를 단언하므로 t1256 의 범위다.

---

## 8. `manager-lead` 개명 영향 — 결정은 열어 둔다

### 8.1 측정된 영향 범위

| 영역 | 측정 | 명령 |
|---|---|---|
| Go 소스·테스트 | 22 개 파일 (`manager_lead_depth_test.go` 31 · `rosterguard/registry.go` 12 · `rosterguard_test.go` 12 · `profile_matrix.go` 8 · `profile_test.go` 7 · `codex_role_contract_test.go` 7 …) | `grep -rc 'manager-lead\|manager_lead\|managerLead' --include='*.go' internal cmd pkg` |
| 프로필·등급 표 | `internal/config/profile.go:150`, `internal/harness/v4manifest/schema.go:280`(TierRed), `internal/harness/delegationmap/types.go:121`, `internal/cli/agentlint/agent_lint.go:478` | 같은 grep |
| 모델 프로필 설정 | `internal/template/templates/.moai/config/sections/llm.yaml:120/134/148` (프로필 3 개) | `grep -rn manager-lead …/config` |
| `delegation.yaml` | 언급 없음(`lead`/`lane`/`worker` 0) | `grep -n 'lead\|lane\|worker' .moai/config/sections/delegation.yaml` |
| C3 Codex TOML | `templates/.codex/agents/moai/manager-lead.toml` (191 행) + `internal/template/agentemit/agents-codex.yaml` 의 역할 서술 | `make agents-emit` 재방출 대상 |
| 테스트 고정물 | `internal/cli/testdata/codex-rollouts-t1171/{roles,roles-other-version}/manager-lead.toml` | `grep -rn 'Lane spawn authority' internal` |
| 문서 | `manager-lead` 530 회 / 103 개 파일 (docs-site·README·루트·`.claude`·템플릿) | `grep -rho 'manager-lead' … \| wc -l` |
| docs-site 경로 | `advanced/manager-lead.md` × 4 로케일, `_meta.yaml`, `data/menu/main.yaml:736-740`(4 로케일 제목), `vercel.json:203-210`(manager-kanban → manager-lead 리다이렉트 2 건이 이미 있음) | |
| `archived-agent-rejection.md` | 현재 `manager-lead`·`manager-kanban` 어느 쪽도 행이 없음 | `grep -n 'manager-lead\|manager-kanban' …/archived-agent-rejection.md` |
| `CLAUDE.md` §4 | 선택 트리 7번, 「Retained agents (13)」 목록 | |
| 에이전트 메모리 | 로컬 `.claude/agent-memory/*lead*` 없음 | `ls -d .claude/agent-memory/*lead*` |

선례: `c55c61aa5`(2026-08-13, manager-lead → manager-kanban, 65 개 파일) 다음 `310d75dd2`(2026-08-18, manager-kanban → manager-lead, 18 개 파일). 닷새 사이에 두 번 개명했고, 그때 `archived-agent-rejection.md` 행은 추가하지 않고 docs-site 리다이렉트만 남겼다. 또 한 번 바꾸면 이 에이전트의 세 번째 이름이다.

### 8.2 선택지

| 선택지 | 내용 | 장점 | 비용·위험 |
|---|---|---|---|
| **A. 식별자 유지** | `manager-lead` 는 에이전트 이름(식별자)으로 두고, 산문의 역할 단어만 `leader` 로 바꾼다. 「`manager-lead` — the leader's coordination agent」 식 설명 한 줄을 정의 위치에 둔다. | Go 22 개 파일·llm.yaml·C3·고정물·docs 경로·리다이렉트 모두 무변경. 사용자 프로젝트의 참조가 깨지지 않는다. t1256 과 독립적으로 문서만 진행 가능. | 이름(lead)과 역할 단어(leader)가 어긋난다. 「식별자는 역할 어휘를 따르지 않는다」는 규칙을 문서에 명시해야 한다. |
| **B. 전면 개명 (`manager-leader`)** | 에이전트 파일·Go·설정·C3·docs 경로를 한 번에 바꾸고 `manager-lead` 를 `archived-agent-rejection.md` 에 행으로 추가, docs-site 에 리다이렉트 추가. | 이름과 역할 단어가 일치한다. | 코드 계층(t1256) 작업이 된다. 세 번째 개명. `moai update` 는 `.claude/agents/moai/` 를 통째로 다시 깔므로 사용자 쪽 파일은 자동 교체되지만, 사용자 문서·메모·훅 설정에 남은 `manager-lead` 참조는 거부 행에 걸린다. `manager_lead_depth_test.go` 파일명·함수명 변경. `leader` 가 cg/Agent Teams leader 와 겹치는 문제(§4)가 에이전트 이름까지 번진다. |
| **C. 단계적 개명** | 릴리스 N 에서 B 를 하되 한 릴리스 동안 `manager-lead` 를 거부가 아닌 안내 행(이름 변경 고지)으로 두고, N+1 에서 거부로 올린다. | 사용자 이행이 부드럽다. | 깊이 봉인 테스트는 Agent 도구 보유자가 하나뿐이어야 하므로 두 파일을 동시에 둘 수 없다 — 별칭은 파일이 아니라 거부 표의 행으로만 가능하다. 거부 표에 「안내」 등급이 없으면 규칙 개정이 선행돼야 한다. |
| **D. 역할 중립 이름** | `leader`/`lead` 어느 쪽도 아닌 이름(예: 보드 조정 역할을 뜻하는 이름)으로 바꾼다. | cg/Agent Teams leader 와의 충돌을 이름에서 피한다. | B 의 비용 전부 + 새 이름 합의 비용. |

권고를 내리지 않는다. 다만 문서 계층 SPEC 은 **A 를 기본 가정**으로 쓰고, B/C/D 가 선택되면 개명 작업을 t1256(코드 계층) 쪽 마일스톤으로 넘기도록 설계했다(SPEC REQ-RND-011).

---

## 9. 측정하지 못한 것 (Gaps)

- 분류는 매치 단위 휴리스틱이다. §4 의 Lane A/B(36), Epic Lane(9) 은 `role` 로 잘못 들어갔고, `compound` 159 건(lead 20 · leader 7 · lane 59 · worker 41 · companion 12 · foreman 15 · deputy 5 — `inv-summary.txt` 기준)은 식별자인지 산문인지 가리지 않았다. 줄 단위 처분은 run 단계 몫이다.
- `[HARD]` 가 제목 줄에만 있고 토큰이 본문 줄에 있는 조항은 §6.1 에서 빠졌다. 실제 HARD 영향 범위는 87 줄보다 넓다.
- 로컬 `.claude/loop.md` 는 표면 분류에서 빠졌다(템플릿 짝만 셈).
- CHANGELOG.md(토큰 줄 138)와 `.moai/specs/**`(토큰을 가진 파일 944)는 동결 기록으로 보고 표에 넣지 않았다. 치환 대상도 아니다.
- CJK 는 단어 부분 문자열 수다. 「리드」가 「리드미」 같은 다른 단어 안에 든 경우를 따로 거르지 않았다.
- t1256 의 코드 계층 SPEC 은 이 트리에서 보이지 않는다(t1256 워크트리에 `.moai/reports/t1256/raw/{agent,lane,lead,worker}.txt` 만 있고 SPEC 디렉터리는 없음, 2026-09-26 확인). 칸반 큐에도 t1256·t1257 카드 행이 보이지 않았다(`moai todo` 출력에서 `t125[0-9]` 0 행).

## 10. 잔여 위험

- `leader` 도입은 cg/Agent Teams leader 와 같은 단어를 두 뜻으로 쓰게 만든다(§4). 문맥 한정 규칙 없이 진행하면 독자가 `moai cg` leader pane 과 칸반 leader 를 혼동한다.
- 코드가 `worker-N` 을 계속 출력하는 상태에서 문서만 `lane-N` 으로 바꾸면 문서가 거짓이 된다. t1085/t1102 가 사흘 전(2026-09-22~23) lane-N → worker-N 으로 코드와 문서를 함께 바꿨고 `-f lane-<n>` 은 폐기 예정 별칭으로 남아 있다. 이번 지시는 그 방향을 되돌리는 것이므로 코드 쪽 결정이 먼저다.
- 이번 지시의 모델(「각 lane 이 카드를 스스로 배차하거나 leader 의 배차를 받아 plan > run > sync 를 끝까지 나른다」)은 현행 [HARD] 두 가지와 맞닿는다 — 「Promotion is the operator's act, always」와 칸반 모드의 열별 companion(plan/run/sync 세션) 구조. 이름만 바꾸는 작업과 모델을 바꾸는 작업을 섞으면 HARD 조항이 명칭 치환 커밋 속에서 조용히 바뀐다.

## 11. 운영자에게 물을 것

SPEC `SPEC-ROLE-NAMING-DOCS-001` 의 `research.md` §F 와 같은 목록이다. 치환 마일스톤은 Q1~Q5 가 답해지고 t1256 결론이 확정될 때까지 열리지 않는다.

1. **Q1 — worker → lane 되돌리기의 범위.** t1085/t1102 가 확정한 `worker-N` · `-f worker` 표기를 `lane-N` · `-f lane` 으로 되돌리는가, 아니면 산문의 역할 단어만 lane 으로 하고 식별자는 `worker-N` 으로 두는가? (코드 계층 t1256 이 정할 일이지만 문서 계층이 그 결과를 그대로 따라야 한다.)
2. **Q2 — 칸반 companion 의 운명.** 새 모델(각 lane 이 plan > run > sync 를 끝까지 나름)은 팩토리 모델이다. 칸반 모드의 열별 companion(plan/run/sync 세션)은 lane 으로 **이름만** 바꾸는가, 아니면 모델 자체를 팩토리 방식으로 합치는가? 후자는 명칭 작업이 아니라 기능 변경이다.
3. **Q3 — lane 자기 배차.** 「lane 이 카드를 스스로 배차한다」는 현행 [HARD] 「Promotion is the operator's act, always」·「The lead is the queue's sole producer」와 맞닿는다. 이번 카드는 명칭만 바꾸고 이 불변식은 그대로 두는가?
4. **Q4 — `manager-lead` 개명 선택지(§8.2 A/B/C/D).**
5. **Q5 — leader 동음이의.** `moai cg` 의 leader pane, Agent Teams leader 와 같은 단어를 쓰는 것을 받아들이는가? 받아들인다면 문서에 「kanban/factory leader」 식 한정어를 의무로 두는가?
6. **Q6 — 로케일 어휘.** ko 리더/레인, ja リーダー/レーン 으로 가는가? zh 는 lead 에 主导·主控·领导·负责人 네 어휘가 흩어져 있다 — leader 와 lane(泳道/通道) 에 각각 어느 하나를 고정하는가?
7. **Q7 — foreman · deputy · coordinator.** foreman(무인 루프), deputy(`manager-lead` 의 대리 역할), 「Lead Coordinator」 제목은 leader/lane 모델에서 어떤 이름을 갖는가 — 그대로 두는가, leader 의 하위 역할로 다시 이름 붙이는가?
