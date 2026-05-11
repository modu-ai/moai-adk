# IDEA-003 Ideation — 3-Tier Catalog + Safe Update Sync

> Generated: 2026-05-12 · Phase 4 (Converge) + Phase 5 (Critical Evaluation) of `/moai brain`

## Converged Concept

**"Lean Core + Opt-in Packs + Harness Auto-Gen + cruft-style Safe Sync"**

moai-adk의 skill/agent 카탈로그를 3-tier 구조로 재편한다:

```
┌─────────────────────────────────────────────────────────────┐
│  Tier 1 — Core (always deployed)                            │
│  • Workflow-* 13 + Foundation-* 4 + moai 1 = 18 skills      │
│  • manager-* 8 + evaluator-active + plan-auditor +           │
│    핵심 expert-* (backend/frontend/security/refactoring/    │
│    performance) + builder-harness = ~15 agents              │
│  • moai init / moai update 시 항상 동기화                    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Tier 2 — Optional Packs (opt-in install)                   │
│  • moai pack add backend / frontend / mobile / chrome-ext / │
│    auth / deployment / design (brand+copywriting+handoff)   │
│  • Anthropic marketplace 표준 활용 (/.claude/plugins/)       │
│  • moai update 시 install된 pack만 동기화                    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Tier 3 — Harness-Generated (project-local)                 │
│  • my-harness-* skills, .claude/agents/my-harness/*         │
│  • /moai project 인터뷰에서 자동 제안 (opt-out)              │
│  • moai update 시 절대 보존 (frozen zone)                    │
└─────────────────────────────────────────────────────────────┘

                  ╔═══════════════════════════════╗
                  ║  Safety: moai update --sync   ║
                  ║  (cruft-style drift detection) ║
                  ╚═══════════════════════════════╝
```

## Lean Canvas (9 blocks)

### 1. Problem

신규 moai-adk 설치 시 모든 36 skills + 29 agents가 deploy되어:

- **Context budget 압박**: Anthropic의 skill description budget = context window 1%. Opus 4.7 1M의 1% = 10K tokens. 사용자가 marketplace plugin 추가 시 즉시 overflow → MoAI 자체 skill의 trigger 정확도 저하
- **사용 안 하는 자산 노이즈**: Chrome-extension, Electron, mobile 등 도메인 미사용 프로젝트도 모두 deploy됨
- **harness 인프라 미활용**: 동적 skill/agent 생성 능력이 이미 있는데도 정적 deploy
- **moai update 위험**: 카탈로그가 비대해질수록 update 시 사용자 변경 손실 위험 증가

### 2. Customer Segments

- **(주) 신규 moai-adk 설치자**: `moai init`으로 새 프로젝트 시작하는 모든 개발자
- **(주) 기존 프로젝트 maintainer**: 이미 moai-adk를 도입한 프로젝트의 owner — `moai update` 안전성이 최우선
- **(부) MoAI core contributor**: 카탈로그를 유지보수하는 사람 — 코어/팩 경계가 명확해야 추가/제거 결정 가능

### 3. Unique Value Proposition

**"Lean by Default, Grow by Choice — moai update에서 0 손실 보장"**

- 기본 설치 = 코어 18 skills + 15 agents (현재 대비 -50%)
- 도메인 팩은 `moai pack add` 1줄 명령으로 추가
- harness가 프로젝트 인터뷰로 맞춤 자산 자동 생성
- `moai update`는 cruft-style drift 감지 + 사용자 확인 + 자동 백업 → 데이터 손실 0

### 4. Solution

#### 4.1 카탈로그 재배치 (코드 변경)

`internal/template/templates/.claude/skills/` 구조:

```
.claude/skills/
├── moai/                          # 코어 (always)
├── moai-foundation-{cc,core,quality,thinking}/  # 코어 4개
├── moai-workflow-*/               # 코어 13개 (워크플로우 진입)
├── moai-meta-harness/             # 코어 (harness 엔진 자체)
├── moai-harness-learner/          # 코어 (Tier 4 진화)
└── # 옵션 팩은 별도 디렉토리로 이동:
    # internal/template/packs/
    #   ├── backend/.claude/skills/moai-domain-backend/
    #   ├── frontend/.claude/skills/moai-domain-frontend/
    #   │                            moai-domain-database/
    #   │                            moai-ref-react-patterns/
    #   ├── mobile/
    #   ├── chrome-extension/
    #   ├── auth/
    #   ├── deployment/
    #   └── design/.claude/skills/moai-domain-{brand-design,copywriting,design-handoff}/
    #                              moai-design-system/
    #                              moai-workflow-{design-context,design-import,gan-loop}/
```

`internal/template/templates/.claude/agents/moai/` 구조:

```
agents/moai/
├── builder-harness.md             # 코어 (모든 builder 통합)
├── manager-*.md                   # 코어 8개
├── evaluator-active.md            # 코어
├── plan-auditor.md                # 코어
├── expert-{backend,frontend,security,refactoring,performance}.md  # 코어 5개
└── # 옵션:
    # internal/template/packs/<pack>/.claude/agents/moai/
    #   expert-{mobile,devops,testing,debug}.md
    #   manager-brain.md (brain workflow 사용자만)
    #   researcher.md (research-heavy 프로젝트)
    #   claude-code-guide.md
```

#### 4.2 `moai pack` 명령어 (신규)

```
moai pack add <name>      # 팩 install → .claude/skills/+agents/에 카피
moai pack add backend frontend design   # 다중 install
moai pack remove <name>   # 팩 uninstall (사용 중이면 경고)
moai pack list            # 설치된 팩 목록
moai pack available       # 설치 가능한 팩 목록 (built-in + marketplace)
```

#### 4.3 `moai update --catalog-sync` 안전 동기화

cruft 모델 기반:

1. **Snapshot**: 현재 `.claude/` 와 `.moai/config/`를 `.moai/cache/catalog-snapshot-{timestamp}/`에 백업
2. **Compute drift**:
   - 코어 자산의 사용자 직접 수정 감지 (hash 비교)
   - 사용자 추가 자산 (`my-harness-*`, 사용자 custom skill/agent) 식별
   - 코어에서 제거된 자산이 어디서 사용 중인지 grep 분석
3. **Plan**: drift 보고서 + 동기화 plan을 AskUserQuestion으로 표시
   - "코어 자산 N개 업데이트 (사용자 수정 M개 발견 — 어떻게 처리?)"
   - "옵션 팩 P개 자동 동기화"
   - "사용자 추가 자산 Q개 절대 보존"
4. **Apply**: 사용자 승인 후 적용. 실패 시 snapshot으로 rollback
5. **Audit**: `.moai/cache/catalog-sync-{timestamp}.log` 에 모든 변경 기록

#### 4.4 `/moai project` harness opt-out 자동 제안

`/moai project` Socratic 인터뷰 끝부분에 추가:

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  question: "프로젝트 도메인에 맞는 맞춤 스킬/에이전트를 harness로 생성하시겠습니까?",
  options: [
    { label: "예 (권장)", description: "moai-meta-harness가 인터뷰 결과 기반으로 my-harness-* skills + .claude/agents/my-harness/* 생성. 도메인 특화 안내가 즉시 동작." },
    { label: "아니오, 기본 코어만 사용", description: "코어 18 skills만 사용. 필요 시 나중에 `moai pack add` 또는 `/moai harness` 수동 호출 가능." },
    { label: "팩만 install, harness는 생략", description: "프로젝트 유형 기반 도메인 팩만 install (예: backend). 동적 생성 없음." }
  ]
})
```

기본값 = "예". 거부 시 코어만, 또는 팩 선택만.

### 5. Channels

- `moai init <project>` — 신규 프로젝트는 코어만 deploy + 인터뷰 시작
- `moai update [--catalog-sync]` — 기존 프로젝트 안전 동기화
- `moai pack {add|remove|list|available}` — 신규 명령어
- `/moai project` — harness opt-out 자동 제안 (기존 명령 확장)
- `/moai harness` — 수동 harness 실행 (기존 entry point 유지)
- `moai doctor catalog` — 카탈로그 상태 + budget 진단 (기존 doctor 확장)

### 6. Revenue Streams

N/A (오픈소스 MIT license).

### 7. Cost Structure

| 항목 | 영향 |
|------|------|
| 코드 변경 (`internal/template/`) | 큼 — pack 디렉토리 신설 + deploy 로직 분기 |
| `moai pack` 명령 추가 | 중 — 신규 CLI subtree (≈200-300 LOC) |
| `moai update --catalog-sync` 재설계 | **매우 큼** — drift 감지, 백업, rollback, AskUserQuestion 흐름 모두 신규. 가장 위험 영역. |
| `/moai project` 인터뷰 확장 | 작 — AskUserQuestion 한 라운드 추가 |
| 마이그레이션 가이드 + CHANGELOG | 중 — 4개국어 docs-site sync (§17 규칙) |
| 테스트 (특히 update 안전성) | 매우 큼 — drift 모든 시나리오 (수정/추가/삭제 교차) |

### 8. Key Metrics

1. **코어 크기**: 36 → 18 skills (-50%), 29 → 15 agents (-48%)
2. **신규 설치 후 첫 /moai brain 호출 컨텍스트**: 사용 안 하는 skill description으로 인한 노이즈 감소 (정량 측정 — A/B testing pattern from `.claude/rules/moai/development/skill-ab-testing.md`)
3. **`moai update` 사용자 보고 손실 사례**: 0 (절대 목표)
4. **harness 활성화율**: opt-out 비율 < 20% (사용자가 harness 가치 인식)
5. **pack 평균 설치 수**: 신규 프로젝트당 2-3개 (over-installation 방지)

### 9. Unfair Advantage

- **Anthropic 공식 plugin marketplace 표준 위에 얹음** — 자체 메커니즘 발명 비용 0, cross-tool 호환 (Cursor, Gemini CLI 등)도 보너스
- **harness가 이미 동작** — `moai-meta-harness` + `builder-harness` + `moai-harness-learner` 인프라 완성. 정리만 하면 즉시 이점
- **cruft / Backstage PR-based migration 패턴이 검증된 표준** — 안전성 설계의 invariant가 명확

---

## Phase 5: Critical Evaluation

### 5.1 Adversarial Evaluation

#### 비판 1: "코어 18개도 너무 크지 않은가?"

대안 — 코어를 더 줄여 foundation 4개 + moai 1개 = 5개만.

**반론**:
- Workflow-* 13개는 `/moai plan`, `/moai run`, `/moai sync` 등 모든 핵심 명령의 진입점. 이걸 코어에서 제외하면 `/moai plan` 호출 시 즉시 fail 또는 추가 install 마찰 발생.
- Anthropic 공식 marketplace의 13개 plugin도 같은 패턴 (workflow 진입은 코어).
- `moai-meta-harness` 와 `moai-harness-learner` 가 코어에 있어야 사용자가 "harness 사용하시겠습니까?" 인터뷰를 처음부터 받음.

→ 18개가 **함수적 최소(functional minimum)**. 5개로 줄이면 첫 사용 마찰이 큼.

#### 비판 2: "사용자가 코어 자산을 직접 수정했다면 `moai update`가 어떻게 처리?"

이게 **가장 위험한 시나리오**. 사용자가 `.claude/skills/moai-workflow-plan/SKILL.md`를 직접 편집하여 자기 사용 패턴에 맞춤. moai update가 이를 overwrite하면 손실.

**대응 (cruft 모델 적용)**:
1. moai update 시 코어 자산마다 hash 계산 → 이전 배포 시 hash와 비교
2. hash가 다르면 = 사용자 수정 → **3-way merge 시도** (이전 / 현재 사용자 / 신규 upstream)
3. merge 충돌 시 사용자에게 AskUserQuestion으로 옵션 제시:
   - "사용자 버전 유지 (upstream 무시)" — 안전, 그러나 신규 기능 누락
   - "Upstream 적용 (사용자 변경 백업 후 덮어쓰기)" — 신규 기능 흭득, 백업으로 복구 가능
   - "수동 merge 도우미 표시" — diff 출력 후 사용자가 직접 결정
4. 모든 사용자 수정은 `.moai/cache/user-overrides/{path}` 에 영구 백업 (deletion 없음)

#### 비판 3: "harness가 잘못된 skill을 생성하면?"

`moai-meta-harness`의 출력 품질은 인터뷰 답변에 의존. 사용자가 모호하게 답하면 부정확한 skill 생성 가능.

**대응**:
- harness 생성 자산은 `.claude/skills/my-harness-*/` 로 명확히 격리 → 삭제 쉬움
- `moai harness rollback` 명령으로 즉시 복구 가능 (기존 `moai harness rollback <date>` 인프라 활용)
- 생성 후 사용자에게 즉시 review AskUserQuestion (현재 `moai-harness-learner`의 Tier 4 패턴과 동일)

#### 비판 4: "옵션 팩이 진정 도움이 되는가, 아니면 노이즈?"

사용자가 backend 프로젝트를 시작했다고 가정. 자동으로 `moai pack add backend` 제안 → 5개 skill 추가. 그러나 진짜 필요한 건 1-2개 일 수 있음.

**대응**:
- 팩 단위로 묶는 대신 **개별 skill install 도 허용**: `moai skill add moai-domain-backend`
- 팩은 "추천 묶음", 개별 skill은 "정밀 install"
- /doctor catalog 명령이 6개월간 한 번도 trigger되지 않은 skill을 "유휴(idle)" 표시 → uninstall 제안

#### 비판 5: "왜 Anthropic 공식 marketplace에 그냥 publish하지 않는가?"

대안 — MoAI 자체 carteralog 없이 Anthropic marketplace에 plugin으로만 publish.

**반론**:
- MoAI는 **단일 통합 워크플로우 (plan-run-sync, brain, design)** 를 제공하는 메타 시스템. 코어 skills이 서로를 참조하고 의존성이 강함. 개별 plugin으로 분해하면 의존성 그래프가 복잡해짐.
- 그러나 옵션 팩은 marketplace에 publish 가능: `moai-pack-backend` 등을 marketplace plugin으로 만들면 cross-tool 호환 + 자동 update 이점 흭득.

→ **하이브리드**: 코어는 moai-adk가 직접 배포 (현재 방식 유지), 옵션 팩은 marketplace publish 검토 (Phase 후속 작업).

### 5.2 First Principles Decomposition

**질문: skill이 무엇인가, 본질적으로?**

→ Skill = Claude의 context에 로드되는 instruction set. 그 본질은 "trigger 조건이 충족되면 Claude의 행동을 특정 방향으로 유도하는 텍스트".

**원자 단위로 분해**:
1. **Trigger** (keywords/agents/phases/paths) → Claude가 언제 이걸 사용해야 하는지
2. **Description** (frontmatter) → context budget을 차지, trigger 매칭에 사용됨
3. **Body** (SKILL.md 본문) → trigger 활성화 시 실제로 로드되는 지침
4. **References** (bundled files) → on-demand로 Claude가 fetch

**도출되는 invariant**:
- Skill의 **존재 자체가 비용** (description은 항상 budget 차지)
- Skill의 **활용은 trigger 매칭률에 의존**
- 결론: **사용 빈도가 낮은 skill일수록 존재 비용 > 활용 가치** → 제거하거나 lazy install

이 invariant가 코어 분류의 객관적 기준:
- 코어 = 모든 사용자가 거의 항상 사용 (workflow 진입)
- 옵션 = 도메인별 일부 사용자만 사용 (backend skill은 backend 프로젝트만)
- harness-generated = 특정 프로젝트에서만 사용

**숨겨진 가정 검증**:
- 가정 A: "사용자는 자기 도메인을 알고 적절히 install할 능력이 있다" — 위험. 일부 사용자는 모호함. → harness 자동 제안이 안전망.
- 가정 B: "moai update는 자주 실행된다" — 불확실. → update가 드물게 실행될수록 catalog drift가 누적되어 안전 동기화의 가치 큼.
- 가정 C: "Anthropic marketplace는 안정적" — 검증 필요. 베타 기능일 가능성. → MoAI는 marketplace 미사용 fallback도 유지해야 함.

### 5.3 Evaluation Summary

| 측면 | 평가 | 보완 |
|------|------|------|
| 컨셉 일관성 | 강함 (3-tier가 Anthropic plugin scope와 매칭) | — |
| 사용자 마찰 | 보통 (harness 자동 제안으로 완화) | — |
| 구현 난이도 | 높음 (특히 moai update 재설계) | Phase 6의 SPEC 분해로 위험 격리 |
| 실패 시나리오 | 식별됨 (5가지 비판 + 4가지 대응) | — |
| 가역성 | 강함 (rollback, snapshot, opt-in 구조) | — |
| 기존 사용자 영향 | 0 (moai update의 hash diff + 사용자 확인 + 백업) | 대규모 통합 테스트 필요 |

**최종 판단**: 컨셉은 견고. 위험은 `moai update --catalog-sync` 한 곳에 집중되어 있어 SPEC 분해 시 이 영역에 evaluator-active와 통합 테스트를 집중 배치하면 관리 가능.
