# MoAI 자율 에이전틱 워크플로우 통합 전략

> **상태**: 전략 제안 (synthesizer 산출, 후속 SPEC 분해 대상)
> **대상**: moai-adk-go 메인테이너
> **최적화 목표**: BALANCED — quality · autonomy · efficiency 3축 균형
> **근거 primitive 버전**: Dynamic Workflows = research preview / Claude Code v2.1.154+ · `/goal` = v2.1.139+

---

## 1. 요약 (Executive Summary)

본 전략은 Claude Code의 두 신규 오케스트레이션 primitive — **Dynamic Workflows**(스크립트가 수십~수백 에이전트를 fan-out)와 **`/goal`**(세션 범위 종료 조건으로 turn 자동 연속) — 을 기존 **`/moai loop`**(Ralph Engine, 결정론적 진단 루프)와 결합해 MoAI 전 서브커맨드를 "목표를 향해 자율 수렴"하도록 만드는 3-엔진 모델을 제안한다. 핵심은 **phase 내부에는 자율을 부여하되 phase 경계에는 human gate를 보존**하는 것이다: 구현 착수 승인(plan→run), sync→PR, destructive op은 절대 자율 bypass 대상이 아니며 모두 `AskUserQuestion`으로 남는다. 두 primitive 모두 doctrine 레이어(`dynamic-workflows.md` / `goal-directive.md`)에는 정착했으나 **어떤 서브커맨드 skill body에도 wiring되지 않은** 상태가 본 전략이 해소할 핵심 gap이다. 자율의 4대 불변식은 (1) 구현 착수 승인 AskUserQuestion 게이트, (2) AskUserQuestion 채널 독점, (3) subagent-cannot-spawn-subagent flat 계층, (4) background-write 금지 — 이며 세 엔진 모두 이 불변식을 상속한다.

---

## 2. 세 가지 자율 엔진의 본질

| 엔진 | 누가 다음 단계를 결정하는가 | 중간 결과 위치 | 규모 | 종료 조건 | MoAI 적합 영역 |
|------|------------------------------|----------------|------|-----------|----------------|
| **Dynamic Workflows** | 스크립트(코드가 loop·branching 보유) | 스크립트 변수 (대화 컨텍스트 오염 없음) | 동시 16 / 회당 총 1000 에이전트 | 스크립트 종료 (각 stage = 별도 run) | codebase-wide sweep, 대규모 migration, cross-check 연구, 다차원 adversarial 검토 |
| **`/goal`** | 직전 turn이 끝나면 자동 진입 | Claude 대화 컨텍스트(transcript) | 단일 세션 turn 연속 | fresh 모델(Haiku)이 transcript에서 조건 충족 판정 | "이 end-state가 transcript에 입증될 때까지 계속" — run/coverage/fix 수렴 |
| **`/moai loop`** (기존) | 진단 사이클(LSP/AST-grep/test/coverage)이 잔여 발견 | `.moai/cache/loop-snapshots/` 스냅샷 | 반복 iteration (default max 100) | 모든 진단 clean OR `--max N` | "도구가 flag한 모든 것을 고칠 때까지" — 결정론적 품질 수렴 |

**Dynamic Workflows** 는 *plan을 코드로 이전*하는 primitive다. 스크립트가 loop·branching·중간 결과를 보유하므로 오케스트레이터 대화 컨텍스트에는 최종 답만 돌아온다. 이것이 두 효과를 낳는다: (1) 1M 윈도우를 오염시키지 않고 대규모 read-only fan-out 가능, (2) 재사용 가능한 품질 패턴(에이전트가 서로의 발견을 adversarial 교차 검증, 여러 각도로 plan을 초안하고 가중) 코드화. 단, **스크립트 자체는 파일시스템·셸 접근 불가** — 실제 IO는 spawn된 에이전트가 수행하며, 이들은 항상 `acceptEdits` 모드로 돌고 세션 allowlist를 상속한다. "phase"는 진행 뷰의 관찰 가능성 grouping일 뿐 문서화된 스크립트 API가 아니다(이 점을 MoAI 문서가 typed API로 단정하면 안 됨).

**`/goal`** 은 *세션 범위 종료 조건* primitive다. 매 turn 후 별도의 빠른 모델(기본 Haiku)이 조건을 transcript에 대해 판정해 yes/no + 사유를 반환한다. "no"의 사유는 다음 turn의 가이드가 된다. 결정적 제약은 **평가자가 도구를 호출하거나 파일을 읽지 않는다는 것** — Claude가 이미 transcript에 *surface한 내용*만 판정한다. 따라서 조건은 "transcript에 입증 가능한" 형태여야 한다(예: "go test ./... exits 0"을 오케스트레이터가 출력에 노출). 기계적으로는 세션 범위 prompt-기반 Stop hook의 thin wrapper다.

**`/moai loop`** (Ralph Engine)은 MoAI 고유의 *결정론적 진단 루프*다. LSP·AST-grep·test·coverage를 직접 실행하고 잔여 발견에서 다음 iteration을 시작한다. `/goal`과는 **상보적**이다: `/goal`은 "직전 turn이 끝나면" 다음 turn을 시작하고 모델이 transcript로 완료를 판정, `/moai loop`은 "진단 사이클이 잔여를 발견하면" 다음 iteration을 시작하고 도구가 ground truth로 완료를 판정한다.

---

## 3. MoAI 자율성 아키텍처

### 3-Layer 자율 모델

```
┌─────────────────────────────────────────────────────────────────────┐
│  Layer 1 — /goal  ("언제 멈추나" = 종료 조건)                          │
│    세션 turn을 transcript-측정 가능한 end-state까지 연속.              │
│    예: "all blocking AC PASS가 transcript에 입증, max N turns".        │
│    fresh Haiku 평가자 = work 모델과 독립 판정.                         │
└───────────────────────────────┬─────────────────────────────────────┘
                                 │ (자율 turn 연속을 감싼다)
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Layer 2 — Dynamic Workflows  ("어떻게 분산하나" = fan-out)            │
│    한 turn 내에서 수십~수백 read-only/mechanical 에이전트 병렬.        │
│    중간 결과 = 스크립트 변수 (컨텍스트 오염 0).                        │
│    예: 4-locale doc-parity 검사, 다차원 adversarial review.           │
└───────────────────────────────┬─────────────────────────────────────┘
                                 │ (각 fan-out 결과를 합성)
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Layer 3 — /moai loop (Ralph Engine)  ("무엇이 남았나" = 결정론 진단) │
│    LSP/AST-grep/test/coverage 실행 → 잔여 발견 → 수정 → 반복.          │
│    도구 ground truth가 완료를 판정 (모델 판정 아님).                   │
└─────────────────────────────────────────────────────────────────────┘
```

세 레이어는 직교한다. `/goal`은 "전체 SPEC이 끝날 때까지" 같은 거시 종료 조건을 잡고, 그 안에서 한 turn이 Workflow fan-out을 띄울 수 있으며, 그 turn이 다시 `/moai loop`을 호출해 결정론적 진단을 돌릴 수 있다. **충돌하지 않는 이유**: `/goal`은 turn-경계 판정, Workflow는 turn-내부 분산, `/moai loop`은 iteration-경계 진단으로 각자 다른 시간 척도에서 동작한다.

### Autonomy Boundary 다이어그램

```
   사용자 ──AskUserQuestion──▶ [phase 경계 게이트]
                                     │
          ┌──────────────────────────┼──────────────────────────┐
          │ 구현 착수 승인 (plan→run)        │ sync→PR 생성/머지 승인     │
          │ destructive op 확인       │ ambiguity Socratic 인터뷰  │
          │ hook-block 에스컬레이션   │ CI autofix iter>3 에스컬   │
          └──────────────────────────┼──────────────────────────┘
                                     │ (게이트 통과 = 모든 preference 수집 완료)
                                     ▼
   ┌───────────────────────────────────────────────────────────────┐
   │  phase 내부 자율 영역 (게이트 통과 후에만 진입)                  │
   │    · /goal 종료 조건 loop                                       │
   │    · Workflow fan-out (read-only / mechanical)                  │
   │    · /moai loop 결정론 진단                                     │
   │  ⚠ 이 영역의 어떤 에이전트도 사용자에게 prompt 불가             │
   │  ⚠ launch 전에 모든 preference가 수집되어 있어야 함            │
   └───────────────────────────────────────────────────────────────┘
```

### 4대 불변식

1. **구현 착수 승인 (plan→run human gate)** — plan-auditor PASS·score ≥0.90이어도 run-phase 진입 직전 `AskUserQuestion`(진입/추가검토/중단, 첫 옵션 권장)으로 명시 승인. skip-eligible ≥0.90 자율 bypass는 **Phase 0.5 verdict 재실행에만** 적용되며 구현 착수 승인에는 적용 안 됨 (CLAUDE.local.md §19.1 REQ-ATR-015). `/goal`이 turn STOP을 제거해도 구현 착수 승인 의무는 불변.
2. **AskUserQuestion 채널 독점** — 모든 user-facing 질문은 `AskUserQuestion` 경유. 자율 모드가 prose-prompting으로 대체 불가. deferred tool이므로 매 호출 직전 `ToolSearch(query: "select:AskUserQuestion")` preload 필수.
3. **subagent-cannot-spawn-subagent (Finding A1)** — 오케스트레이터(main 세션)만 spawn 주체. Workflow는 *scaling 메커니즘*이지 subagent 중첩 spawn 수단이 아니다. Workflow agent도, /goal turn 내 agent도 사용자 prompt 불가(비대칭 boundary 상속).
4. **background-write 금지** — `run_in_background: true` agent는 Write/Edit 금지. read-only(research/analysis/review)만 background 허용. 파일 수정 agent는 `run_in_background: false` foreground 순차. (단, Workflow agent는 별개 primitive로 `acceptEdits`로 파일 쓰기 가능 — background-write 금지는 `Agent(run_in_background:true)`에만 적용.)

---

## 4. 서브커맨드별 통합 전략

### 4.1 요약 표

| 서브커맨드 | 현재 흐름(요약) | 적용 엔진 | /goal 조건 | Workflow fan-out | 보존 게이트 |
|-----------|----------------|-----------|:----------:|:----------------:|-------------|
| `project` | 인터뷰 → product/structure/tech.md + codemaps 생성 | Workflow(문서 fan-out) + /goal(전체 완성) | O | O (문서 섹션 병렬) | 인터뷰 Socratic, harness 생성 승인 |
| `plan` | manager-spec 인터뷰·author + plan-auditor 게이트 | Workflow(Phase 0.5 연구 + 2.3 audit) + /goal(2.3 수렴) | O (Phase 2.3 한정) | O (연구·audit) | DP1, Tier, BODP, 품질게이트, 실행모드 |
| `run` | Phase 0.95 모드선택 → manager-develop TDD/DDD 구현 | /goal(AC 수렴) + /moai loop(진단) + Workflow(다파일 sweep만) | O | 조건부(코딩=No, 기계적 migration=Yes) | **구현 착수 승인**, 모드선택은 자율 |
| `sync` | manager-docs 문서·frontmatter + sync-auditor + PR | Workflow(4-locale parity) + /goal(parity 0) | O | O (locale·doc 병렬) | **sync→PR 생성/머지** |
| `fix` | LSP/lint/type 오류 자동 감지·수정 | /moai loop(주) + /goal(수렴 래핑) | O | X (순차 수정) | Level 3+ 수정 승인 |
| `loop` | scan→fix→verify 반복 (Ralph Engine) | /moai loop(기존) + /goal(상위 종료) | O | X | Level 3 AskUserQuestion, 메모리 압박 체크포인트 |
| `review` | 보안·@MX 태그·코드 리뷰 | Workflow(다차원 adversarial) + /goal(미해결 0) | O | O (차원별 병렬) | finding 적용 승인(쓰기 시) |
| `coverage` | 갭 분석 → 누락 테스트 생성 | /goal(임계 도달) + /moai loop(생성·검증) | O | 조건부(대규모 갭=Yes) | 신규 테스트 커밋 승인 |
| `e2e` | Chrome/Playwright E2E 작성·실행 | /goal(시나리오 green) | O | X (브라우저 순차) | 시나리오 정의, destructive nav 확인 |
| `gate` | lint+format+type+test 병렬 (pre-commit) | Workflow(검사 fan-out) | X (단발) | O (검사 병렬) | 없음(read-only 검사) → hook-block 에스컬 |
| `mx` | @MX 주석 스캔·추가 | Workflow(파일 sweep) + /goal(coverage 도달) | O | O (파일 병렬 스캔) | 태그 자율(승인 불요), 쓰기는 foreground |
| `codemaps` | 코드베이스 스캔 → 아키텍처 문서 | Workflow(모듈 sweep) | X (단발 fan-out) | O (모듈 병렬) | 문서 overwrite 확인 |
| `clean` | dead code 식별·삭제 (test 검증) | /moai loop(검증) + /goal(clean 도달) | O | 조건부 | **삭제 = write, foreground 강제** + 삭제 승인 |
| `brain` | 7-phase 아이디어→검증 제안 | Workflow(Phase 3 연구) + /goal(Phase 수렴) | O (연구 한정) | O (연구 fan-out) | 각 Phase 전환, 제안 선택 |
| `design` | Claude Design import(A) / brand 코드(B) | /moai gan-loop(기존) + /goal(점수 임계) | O | X | path 선택, GAN 수렴 승인 |
| `db` | 스키마 설계·동기화 | /goal(스키마 일치) | O | X | **migration = destructive 확인** |
| `harness` | 학습 subsystem status/apply/rollback | (자율 없음 — 학습 제안은 AskUserQuestion) | X | X | apply/rollback 승인 |
| `feedback` | 피드백 수집 → GitHub 이슈 | (자율 없음 — 단발) | X | X | 이슈 생성 승인 |

### 4.2 그룹별 상세

#### 그룹 A — plan (SPEC 기획)

**현재 흐름**: manager-spec + Explore + manager-git 소유의 단일 phase 체인. Step 0 brain-proposal 스캔 → Phase 1A Explore 코드베이스 분석 → Phase 0.3 명료도 점수 → Phase 0.3.1 deep 인터뷰(1-5 라운드) → Phase 0.4 UltraThink → Phase 0.5 Explore research.md → Phase 1B GEARS planning → **DP1 HUMAN GATE**(Proceed/Annotate/Draft/Cancel) → Phase 1.6 Tier 판정 → Phase 2 spec/plan/acceptance 작성 → **Phase 2.3 plan-auditor adversarial audit + FAIL→revise 재시도(max 3)** → Phase 3 BODP 게이트 → Phase 3.6 SPEC 품질게이트 → DP2/3/3.5 실행모드 선택. **plan은 audit-ready 신호에서 종료하며 구현 착수 승인(run 진입)를 건너지 않는다.**

**적용 엔진**: hybrid — Workflow(Phase 0.5 multi-angle 연구 + Phase 2.3 adversarial audit) + /goal(Phase 2.3 FAIL→revise 수렴 루프). `/moai loop`은 **미사용** — plan-phase는 명세 markdown만 생산하며 컴파일러/test 표면이 없다(Ralph Engine의 진단 대상이 run-phase에 있음).

**/goal 조건** (plan group 검증 verdict의 4개 NEEDS-FIX 교정 모두 적용):
- 측정 술어는 **파일 경로가 아니라 오케스트레이터가 transcript에 surface한 verdict 라인** (교정 1).
- STOP-on-score-regression 조기 이탈을 명시적 두 번째 bound로 추가 — STOP 또는 3×FAIL 발화 시 `/goal` clear하고 기존 Step 2.3.5 에스컬레이션 AskUserQuestion으로 제어 반환(루프가 직접 해결 시도 금지) (교정 2).
- `/goal` 및 Phase 2.3 audit Workflow launch를 `plan_audit.enabled == true` (harness ∈ {standard, thorough})에 게이트. `minimal` harness에서는 Phase 2.3 전체 skip — /goal·audit Workflow 모두 없음 (교정 3).
- frontmatter 12필드·GEARS↔GWT 하위절은 verdict가 이미 포함하므로 조건에서 제거(transcript-측정 가능성 강화) (교정 1·4).
- 정확한 조건 문자열은 §5.1 참조.

**Workflow shape**: (1) **Phase 0.5 multi-angle 연구** (`/deep-research` 형): fan-out 단위 = 연구 각도당 1 에이전트(코드베이스-아키텍처는 Explore, 요구·엣지케이스, 선행기술은 WebSearch+Context7, 보안-범위, 의존성-영향). parallel-barrier(5-8 read-only 에이전트, 동일 파일 미접촉) → adversarial 교차 검토(claim voting·모순 표면화) → 단일 research.md 합성. (2) **Phase 2.3 adversarial audit ensemble** (선택, Tier L / thorough만): fan-out 단위 = audit 차원당 1 에이전트(frontmatter-schema, GEARS↔AC coverage, 보안-범위, simplicity/scope-creep, GEARS-pattern). parallel-barrier → "judge" 에이전트가 차원 발견을 단일 PASS/FAIL verdict로 reconcile. **두 스크립트 모두 read-only** — verdict는 run 반환 *후* 오케스트레이터/manager-spec이 review 리포트에 기록(workflow agent가 surface 단계 수행 안 함) (교정 4).

**보존 게이트**: Phase 0.3.1 명료도 인터뷰, DP1 plan-review + annotation 사이클, Phase 1.6 Tier 판정, Phase 2.3 Step 2.3.5 3×FAIL 에스컬레이션, Phase 3.0 BODP, Phase 3.6 SPEC 품질게이트, DP2/3/3.5 실행모드 — 모두 `AskUserQuestion`이며 **모든 preference를 Workflow/goal launch 전에 수집**. 구현 착수 승인는 plan 범위 밖. brain-proposal 선택(Step 0)은 명시 AskUserQuestion(SPEC 자동 생성 금지).

#### 그룹 B — run (구현)

> 주의: run group의 원본 design은 rate-limit으로 미전달되어 adversarial verdict가 "missing-input UNSOUND"(내용 거부 아님)였다. 아래는 primitive 프로파일 + orchestration-mode-selection 그라운드트루스 + HARD 제약에서 재설계한 안이며, verdict가 요구한 6개 안전 조항을 모두 명시 충족한다.

**현재 흐름**: Phase 0.5 plan-auditor verdict → **구현 착수 승인 HUMAN GATE** → Phase 0.95 5-mode 자율 선택(trivial/background/agent-team/parallel/sub-agent) → manager-develop 구현(cycle_type ∈ {ddd, tdd, autofix}). 코딩-heavy 기본은 Mode 5 sequential sub-agent (Finding A4 caveat).

**적용 엔진**: /goal(AC 수렴 래핑) + /moai loop(진단) 주축, Workflow는 **대규모 기계적 변환에 한정**. 코딩-heavy 본체는 sequential sub-agent 유지 — Anthropic "coding has fewer parallelizable subtasks than research" caveat가 MoAI Mode 5 default와 일치하므로 Workflow를 코딩 구현 자체에 쓰지 않는다.

**Phase 0.95 모드 카탈로그 확장 제안**: 현재 5-mode는 dynamic-workflows를 누락한다. **Mode 6: workflow**를 추가하되 진입 조건을 엄격히: scope ≥ ~30 파일 AND 변환이 기계적(call-site rename, import 경로 일괄 변경 등) AND 진정 병렬(파일 간 의존 없음). 코딩-heavy + multi-domain은 여전히 Mode 5 우선(Finding A4). Mode 6 선택 시에도 구현 착수 승인는 이미 통과한 상태이며, Workflow launch 전 모든 preference 수집 완료를 progress.md에 로깅.

**/goal 조건**: run-phase 전체를 "모든 blocking AC가 transcript에 PASS 입증 + 회귀 없음 + max N turns"로 래핑(§5.2). 단, **구현 착수 승인 통과 후에만** /goal 설정 — /goal이 구현 착수 승인를 대체하거나 우회하지 않는다. semantic 실패(data race, deadlock, panic, test assertion)는 /goal이 자율 수정 금지, AskUserQuestion 에스컬레이션(CONST-V3R5-010).

**verdict 요구 6조항 충족**: (1) 구현 착수 승인는 plan-auditor 점수와 무관하게 mandatory AskUserQuestion human gate로 유지; (2) 모든 user preference를 Workflow/subagent launch 전 오케스트레이터가 수집; (3) subagent-spawns-subagent 중첩 없음(flat, Mode 6 Workflow도 scaling이지 중첩 아님); (4) background agent는 read-only만(구현=foreground); (5) 각 /goal은 transcript-측정 가능 + turn bound; (6) Workflow는 진정 병렬 고볼륨에만, 결정론 진단은 /moai loop.

#### 그룹 C — sync (문서·PR)

> 주의: sync group design도 rate-limit 미전달. 아래는 재설계안이며 sync→PR 게이트를 핵심 보존 대상으로 명시한다.

**현재 흐름**: manager-docs 문서·CHANGELOG·README + frontmatter status 전이(in-progress→implemented) + sync-auditor 4차원 점수 + **PR 생성(AskUserQuestion)**. docs-site 4-locale 동기화 의무(§17).

**적용 엔진**: Workflow(4-locale doc-parity fan-out) + /goal(parity 0 수렴). 4개 locale(en/ko/ja/zh) × N 문서의 parity 검사·동기화는 진정 병렬 read-heavy 작업이라 Workflow에 최적.

**/goal 조건**: "4-locale parity 위반 0이 transcript에 입증 + frontmatter status 전이 완료 + max N turns"(§5.3). **PR 생성/머지는 /goal 조건에 포함 금지** — 이는 명시 AskUserQuestion 게이트("PR 즉시 생성/검토 후 PR/중단")로 남으며 /goal 수렴 후 별도 단계로 surface.

**보존 게이트**: sync→PR 생성은 AskUserQuestion. PR 머지 승인은 yes/no AskUserQuestion 확인. `git push`/force-push 등 destructive는 별도 확인. sync-auditor의 skeptical 4차원 점수는 자율 산출하되 must-pass 실패 시 AskUserQuestion 에스컬레이션.

#### 그룹 D — fix + loop (자동 수정 루프)

> 주의: fixloop group design도 rate-limit 미전달. 아래는 재설계안.

**현재 흐름**: `/moai loop` = scan(LSP/AST-grep/test/coverage 병렬)→fix(에이전트 위임)→verify→repeat (default max 100). `/moai fix` = LSP/lint/type 오류 단발 자동 수정. Level 1(자동)/2(로그)/3(AskUserQuestion)/4(수동) fix 레벨.

**적용 엔진**: `/moai loop`(Ralph Engine) **주축 유지** + `/goal`(상위 종료 조건 래핑). 둘은 상보적이므로 `/goal`이 `/moai loop`을 대체하지 않는다 — `/goal`은 "전체 품질 수렴"이라는 거시 종료 조건을, `/moai loop`은 각 iteration의 결정론 진단을 담당.

**핸드오프 규칙** (현재 gap — 둘 사이 결정점 부재): `/moai loop`은 도구 ground truth가 완료 판정 가능할 때(LSP/test 표면 존재) 우선. `/goal`은 도구로 직접 검증 불가하고 transcript에 입증해야 하는 상위 조건(예: "모든 SPEC AC + 모든 진단 clean")에 사용. 동시 사용 시 `/goal`이 바깥, `/moai loop`이 안쪽.

**/goal 조건**: "zero errors + all tests pass + coverage ≥ 임계가 transcript에 입증, max N iterations"(§5.4). Level 3+ 수정은 /goal이 자율 적용 금지 — AskUserQuestion. 메모리 압박(세션 25분 초과 / iteration 시간 doubling) 시 체크포인트 저장 + paste-ready resume + `/clear` 권고(context-window-management.md 임계).

#### 그룹 E — quality (review · coverage · e2e · gate)

> 주의: quality group design도 rate-limit 미전달. 아래는 재설계안.

- **review**: Workflow 다차원 adversarial fan-out(보안/성능/@MX-compliance/simplicity/correctness 차원별 1 에이전트 → judge reconcile). prompting-best-practices.md "finding과 filtering 분리" 적용 — 각 차원은 confidence+severity 포함 모든 issue 보고, judge가 랭킹. /goal = "미해결 high-severity 0이 transcript에 입증". finding **적용(쓰기)** 시 AskUserQuestion.
- **coverage**: /goal(임계 도달) + /moai loop(갭 분석·테스트 생성). 대규모 갭(파일 ≥30)은 Workflow로 파일별 갭 분석 병렬화 가능. 신규 테스트 커밋은 AskUserQuestion.
- **e2e**: /goal("시나리오 N개 green이 transcript에 입증"). 브라우저 자동화는 순차 의존이 강해 Workflow fan-out 부적합. destructive navigation(데이터 변경 폼 제출 등)은 확인.
- **gate**: lint+format+type+test 병렬 단발 — Workflow 검사 fan-out에 적합하나 단발이므로 /goal 불요. hook 차단(exit 2) 시 AskUserQuestion 에스컬레이션(accept-fix/override-logged/abort).

#### 그룹 F — codebase sweep (mx · codemaps · clean)

> 주의: sweep group design도 rate-limit 미전달(verdict는 missing-input UNSOUND). 아래 재설계안은 verdict가 요구한 6개 안전 조항을 명시 충족한다.

- **mx**: @MX 주석 스캔·추가. 파일 sweep은 Workflow 병렬 스캔에 최적. /goal = "@MX coverage 임계 도달이 transcript에 입증". 태그는 자율(승인 불요, constitution) 이나 **쓰기는 foreground** (background-write 금지). Workflow agent는 acceptEdits로 쓰기 가능하나 mx 쓰기는 mechanical하므로 적합.
- **codemaps**: 모듈 스캔 → 아키텍처 문서. Workflow 모듈 병렬 sweep(단발 fan-out, /goal 불요). 기존 문서 overwrite 확인.
- **clean**: dead code 식별·삭제. **삭제 = write path → foreground 강제** (verdict 조항 d). /moai loop(test 검증)으로 삭제 후 회귀 검증. /goal = "dead code 0 + 모든 test pass가 transcript에 입증". **삭제 승인은 AskUserQuestion** — clean은 destructive 성격이므로 자율 bypass 금지.

**verdict 요구 6조항 충족**: (a) mx/codemaps/clean의 run-phase 진입도 구현 착수 승인 human 승인 보존(plan에서 파생된 SPEC일 경우); (b) 모든 preference를 launch 전 AskUserQuestion 수집, mid-run prompt 없음; (c) subagent 중첩 spawn 없음; (d) clean 삭제는 foreground; (e) 각 /goal은 transcript-측정 가능 + turn bound; (f) Workflow는 진정 병렬 sweep, /moai loop은 결정론 진단.

#### 그룹 G — meta (brain · design · db · harness · feedback)

> 주의: meta group design도 rate-limit 미전달. 아래는 재설계안.

- **brain**: 7-phase 아이디어→제안. Phase 3(시장·생태계 연구)만 Workflow fan-out(WebSearch+Context7 병렬 + cross-check). /goal은 "각 Phase 산출물 완성" 한정. **각 Phase 전환은 AskUserQuestion** (Phase 자율 연속 금지 — brain은 인간 의사결정 heavy).
- **design**: Claude Design import(A) / brand 코드(B). 기존 `/moai gan-loop`(Builder-Evaluator) 유지 + /goal("4차원 점수 ≥ Sprint Contract 임계가 transcript에 입증"). path A/B 선택은 AskUserQuestion. GAN 수렴 승인은 design.yaml `require_approval`.
- **db**: 스키마 설계·동기화. /goal("스키마-코드 일치가 transcript에 입증"). **migration 실행은 destructive(drop/alter) → 확인 필수**, 자율 bypass 금지.
- **harness**: 학습 subsystem. **자율 없음** — Tier 4 자동 업데이트 제안은 오케스트레이터가 AskUserQuestion으로 surface(moai-harness-learner 계약). apply/rollback 승인.
- **feedback**: 피드백→GitHub 이슈. **자율 없음** — 단발 작업, 이슈 생성 승인.

#### 그룹 H — project (초기화·문서 생성)

> 주의: project group design도 rate-limit 미전달. 아래는 재설계안.

**현재 흐름**: Socratic 인터뷰 → product.md/structure.md/tech.md + codemaps 생성. harness 생성 시 moai-meta-harness 7-phase.

**적용 엔진**: Workflow(문서 섹션 병렬 생성 — 각 문서가 독립 fan-out 단위) + /goal("3개 핵심 문서 + codemaps 완성이 transcript에 입증"). **인터뷰 Socratic는 AskUserQuestion** (자율 launch 전 완료). harness 생성(`harness-*` skill + `.claude/agents/harness/`)은 사용자 승인 — user-owned namespace이므로 자동 생성 전 명시 확인.

---

## 5. /goal 조건 템플릿 카탈로그

> 모든 조건은 (1) transcript-측정 가능한 단일 end-state, (2) 명시된 검증 방법(오케스트레이터가 surface), (3) turn/time bound를 포함한다. 평가자(Haiku)는 도구를 호출하지 않으므로 조건은 Claude가 출력에 *입증*해야 하는 형태여야 한다.

### 5.1 plan (Phase 2.3 author↔audit 수렴) — verdict 4개 교정 적용

```text
The most recent plan-auditor verdict line surfaced in the conversation
reads "Verdict: PASS" for SPEC-{ID} at the tier threshold
(S=0.75, M=0.80, L=0.85); OR an orchestrator-surfaced line states the
3-iteration cap was reached; OR an orchestrator-surfaced line states a
STOP-on-score-regression signal was emitted (iter N+1 aggregate < iter N).
Stop on any of these. Max 3 plan-auditor iterations.
On 3rd FAIL or on STOP-on-regression, clear this goal and return control
to the Step 2.3.5 escalation AskUserQuestion — do NOT attempt to resolve
the escalation autonomously.
[LAUNCH PRECONDITION: only when plan_audit.enabled == true
 (harness in {standard, thorough}); under minimal harness skip Phase 2.3
 entirely — no goal, no audit workflow.]
```

### 5.2 run (AC 수렴 — 구현 착수 승인 통과 후에만 설정)

```text
Every blocking acceptance criterion in
.moai/specs/SPEC-{ID}/acceptance.md has its PASS evidence surfaced in
the conversation (test output, build exit 0, or explicit AC-id: PASS
line); AND `go test ./...` exit 0 is surfaced; AND no test file outside
the SPEC scope was modified (surfaced via git status). Stop when all
hold. Max 20 turns.
On any semantic failure (data race, deadlock, panic, test assertion
failure), clear this goal and escalate via AskUserQuestion — do NOT
auto-fix semantic failures.
[PRECONDITION: 구현 착수 승인 user approval already obtained; this goal does
 NOT substitute for or bypass 구현 착수 승인.]
```

### 5.3 sync (4-locale parity 수렴 — PR 게이트 제외)

```text
A surfaced parity report shows 0 violations across all 4 docs-site
locales (en/ko/ja/zh) for the touched documents; AND the SPEC-{ID}
frontmatter status transition (in-progress -> implemented) is surfaced
as complete; AND a surfaced sync-auditor line shows all must-pass
dimensions PASS. Stop when all hold. Max 15 turns.
[EXCLUSION: PR creation and merge approval are NOT part of this goal —
 they remain explicit AskUserQuestion gates surfaced after convergence.]
```

### 5.4 loop / fix (품질 수렴)

```text
A surfaced diagnostic summary shows zero LSP errors, zero failing tests,
and coverage >= the project threshold (default 85%) for the detected
language. Stop when all hold. Max 50 iterations (memory-safe bound).
On any Level 3+ fix (logic change, API modify) or semantic failure,
clear this goal and request approval via AskUserQuestion — do NOT
auto-apply Level 3+ fixes.
```

### 5.5 coverage (임계 도달)

```text
A surfaced coverage report shows package-level coverage >= 85% (critical
packages cli/template/hook >= 90%) with zero failing tests for the
detected language. Stop when the threshold holds for every package in
scope. Max 30 iterations.
[Newly generated test commits require AskUserQuestion approval before
 commit — surface the proposed test list first.]
```

### 5.6 review (high-severity 0)

```text
A surfaced review summary lists every issue with confidence + severity
(finding and filtering separated), AND a surfaced judge line states
zero unresolved high-severity findings remain. Stop when zero
high-severity unresolved. Max 10 turns.
[Applying any finding that modifies files requires AskUserQuestion
 approval — review itself is read-only until approval.]
```

### 5.7 mx / codemaps / clean (sweep)

```text
# mx
A surfaced @MX coverage report shows annotation coverage >= the target
for all in-scope files (high fan_in functions have @MX:ANCHOR,
dangerous patterns have @MX:WARN). Stop when target met. Max 20 turns.

# clean (foreground-only; deletion requires approval)
A surfaced report shows zero dead-code findings AND all tests pass after
removal. Stop when both hold. Max 15 turns.
[Each deletion batch requires AskUserQuestion approval before write —
 clean is destructive; no autonomous deletion bypass.]
```

---

## 6. Workflow 스크립트 패턴 카탈로그

> 의사코드는 개념 모델이다. 공식 docs는 named 스크립트 API(agent/parallel/pipeline/phase 함수)를 문서화하지 않으므로 아래는 *coordinate 에이전트 → 결과를 스크립트 변수에 보관 → 최종 합성* 의 개념 흐름만 표현한다. 실제 스크립트는 Claude가 작업 설명에서 생성한다. **모든 스크립트는 launch 전 모든 preference 수집 완료를 전제**하며, 어떤 workflow agent도 사용자에게 prompt하지 않는다.

### 6.1 plan-phase multi-angle 연구 (parallel-barrier)

**언제**: Phase 0.5에서 진정 병렬·read-only·cross-check 이득이 있는 연구. `--team` researcher+analyst+architect 경로보다 컨텍스트 윈도우 보존에 유리.

```
# 입력: SPEC 주제, 연구 각도 목록 (사전 수집)
angles = ["codebase-architecture", "requirements-edge-cases",
          "prior-art", "security-scope", "dependency-impact"]

# parallel-barrier: 5-8 read-only 에이전트 동시, 동일 파일 미접촉
findings = parallel_for(angle in angles):
    spawn read-only agent(angle)   # Explore / WebSearch / Context7
    return agent.findings           # 스크립트 변수에 보관

# adversarial 교차 검토: 각 발견을 다른 발견에 대조
crosschecked = adversarial_review(findings)   # claim voting, 모순 표면화

# 합성: 단일 research.md 후보 (오케스트레이터가 run 반환 후 Write)
return synthesize(crosschecked)
```

**parallel-barrier 선택 이유**: 각 각도는 독립이며 서로의 출력을 입력으로 받지 않는다(barrier에서 모두 모인 뒤 교차 검토). pipeline이 아니다.

### 6.2 run-phase TDD fan-out (조건부 — 기계적 변환만)

**언제**: 대규모 기계적 변환(call-site rename, import 경로 일괄 변경, ≥30 파일, 파일 간 의존 없음). **코딩-heavy 신규 구현에는 사용 금지**(Mode 5 sequential 우선, Finding A4).

```
# 입력: 변환 대상 파일 목록, 변환 규칙 (사전 수집)
files = glob(target_pattern)   # ≥30, 상호 독립

# parallel-barrier: 동시 16 cap, 파일당 1 에이전트, acceptEdits 쓰기
results = parallel_for(file in files):
    spawn agent(file, transform_rule)   # Read → Edit → 로컬 test
    return {file, applied, test_result}

# 합성: 변환 요약 + 실패 파일 목록 (오케스트레이터 컨텍스트로 반환)
return summarize(results)
```

**주의**: allowlist에 `go test`, `golangci-lint`, `gh` 등 사전 등록(mid-run permission stall 방지). 변환 규칙은 구현 착수 승인 통과 후 확정된 것만.

### 6.3 sync-phase 4-locale doc-parity (parallel-barrier)

**언제**: docs-site 4개 locale × N 문서의 parity 검사·동기화. read-heavy, 진정 병렬.

```
# 입력: 변경 문서 목록, locale 목록 (사전 수집)
locales = ["en", "ko", "ja", "zh"]
docs = changed_docs()

# parallel-barrier: locale×doc 조합당 1 에이전트
violations = parallel_for((loc, doc) in product(locales, docs)):
    spawn read-only agent(loc, doc)   # frontmatter/H1/glossary parity 검사
    return agent.violations

# 합성: parity 리포트 (위반 0 = /goal 충족)
return parity_report(violations)
```

**parallel-barrier 선택 이유**: 각 locale×doc 검사는 독립. 위반 발견 후 *수정*은 별도 foreground 단계(쓰기는 workflow가 아니라 manager-docs가 순차).

### 6.4 review 다차원 adversarial fan-out (parallel-barrier + judge)

**언제**: 코드 리뷰를 차원별로 분산하고 judge가 reconcile. 단일 pass보다 precision 향상.

```
# 입력: 리뷰 대상 diff, 차원 목록 (사전 수집)
dims = ["security", "performance", "mx-compliance",
        "simplicity", "correctness"]

# parallel-barrier: 차원당 1 read-only 에이전트
# finding과 filtering 분리: 각 차원은 confidence+severity 포함 모든 issue 보고
per_dim = parallel_for(dim in dims):
    spawn read-only agent(dim, diff)
    return agent.issues   # [{issue, confidence, severity}]

# judge: 차원 발견을 단일 랭킹 리포트로 reconcile
return judge_reconcile(per_dim)   # high-severity 0 = /goal 충족
```

**parallel-barrier + judge 선택 이유**: 차원은 독립 병렬(barrier), 이후 단일 judge가 합성(반-pipeline 마무리). adversarial cross-review는 judge 단계에서.

### 6.5 codebase sweep — mx / codemaps (parallel-barrier, 단발)

**언제**: 파일/모듈 단위 독립 sweep. /goal 불요(단발 fan-out)이나 mx는 coverage 임계가 있으면 /goal 래핑.

```
# 입력: 스캔 대상 (사전 수집)
targets = glob(scan_pattern)   # 파일(mx) 또는 모듈(codemaps)

# parallel-barrier: target당 1 에이전트
# mx: acceptEdits 쓰기(mechanical 태그) | codemaps: read-only 분석
outputs = parallel_for(t in targets):
    spawn agent(t)
    return agent.output

# 합성: mx coverage 리포트 또는 아키텍처 문서 후보
return aggregate(outputs)
```

**주의**: clean(dead-code 삭제)은 이 패턴에 **미포함** — 삭제는 destructive + 승인 필요 + foreground 강제이므로 Workflow fan-out이 아니라 /moai loop + AskUserQuestion 경로.

### 6.6 패턴 선택 가이드 (pipeline vs parallel-barrier)

| 작업 특성 | 패턴 | 근거 |
|-----------|------|------|
| 단위가 상호 독립, 동시 실행 가능 | **parallel-barrier** | fan-out 후 barrier에서 합성 (6.1·6.3·6.4·6.5) |
| 단위 A 출력이 단위 B 입력 | pipeline (단계별 별도 workflow) | mid-run 사용자 sign-off 불가 → 각 stage = 별도 run |
| 합성에 judge/reconcile 필요 | parallel-barrier + judge | 병렬 발견 → 단일 판정 (6.4) |
| 쓰기 + 승인 필요 | Workflow 미사용 | foreground + AskUserQuestion (clean) |

---

## 7. 안전·경계 보존

### 7.1 4대 불변식의 자율 모드 적용

**구현 착수 승인 (§19.1 REQ-ATR-015)**: plan→run 경계는 항상 명시 AskUserQuestion. plan-auditor PASS·score ≥0.90이어도 자율 bypass 금지. skip-eligible ≥0.90은 Phase 0.5 verdict 재실행에만 적용, 구현 착수 승인에 적용 안 됨. `/goal`은 turn STOP을 제거하지만 구현 착수 승인 의무를 제거하지 않는다. **Dynamic Workflows는 mid-run 사용자 입력 불가이므로 구현 착수 승인를 run 내부에 삽입 불가** — 따라서 구현 착수 승인는 반드시 Workflow launch *전*에 통과해야 하며, gated phase 간 sign-off는 각각 별도 workflow로 분리.

**AskUserQuestion 채널 독점**: 자율 모드가 prose-prompting으로 질문 대체 불가. 매 호출 직전 `ToolSearch(query: "select:AskUserQuestion")` preload. 자율 흐름이 게이트로 재진입할 때마다 재-preload. AskUserQuestion이 자동 부착하는 "Other" 옵션으로 free-form 답변 지원(prose 질문 불요).

**subagent-cannot-spawn-subagent (Finding A1)**: 오케스트레이터(main 세션)만 spawn. Workflow는 scaling이지 중첩 spawn이 아니다. Workflow agent도 사용자 prompt 불가(비대칭 boundary 상속) — 입력 부재 시 structured blocker report 반환, 오케스트레이터가 AskUserQuestion 라운드 후 fresh prompt로 재위임.

**background-write 금지**: `Agent(run_in_background:true)`는 Write/Edit 금지. read-only(research/analysis/review)만 background. 파일 수정은 `run_in_background:false` foreground 순차. (Workflow agent는 별개 primitive로 acceptEdits 쓰기 가능 — 이 금지는 background Agent에만.)

### 7.2 "preferences를 launch 전에 수집" 패턴

```
[게이트 통과 시퀀스]
Step 1: ToolSearch(query: "select:AskUserQuestion")
Step 2: AskUserQuestion — 모든 preference/decision 수집
        (Tier, base branch, 실행 모드, PR 전략 등)
Step 3: 수집 완료 확인 (100% intent clarity)
Step 4: /goal 설정 OR Workflow launch OR /moai loop 시작
        (이 시점부터 mid-run 사용자 입력 없음)
Step 5: 자율 수렴 → 종료 조건 충족 → 다음 게이트로 surface
```

핵심: launch 후에는 mid-run prompt가 불가능하므로(Workflow), 또는 turn STOP이 제거되므로(/goal), **모든 사용자 결정은 launch 전에 소진**되어야 한다. 누락된 결정이 mid-run에 필요해지면 자율을 중단하고 게이트로 복귀.

### 7.3 위반 anti-pattern 목록

- **AP-1**: plan-auditor 점수가 높다는 이유로 구현 착수 승인 없이 `/moai run` 자율 시작 (CLAUDE.local.md §19.1 명시 anti-pattern).
- **AP-2**: `/goal` 평가자가 파일을 읽는다고 가정한 조건(파일 경로 술어). 평가자는 transcript만 판정 — 측정 대상을 surface된 라인으로 작성해야 함 (plan group 교정 1).
- **AP-3**: STOP-on-regression / 3×FAIL 에스컬레이션을 `/goal` 루프가 직접 해결 시도 (AskUserQuestion swallow) (교정 2).
- **AP-4**: `minimal` harness에서 `plan_audit.enabled:false`인데 audit /goal launch → 영원히 미충족, 3-cap까지 무의미 spin (교정 3).
- **AP-5**: Workflow agent / `/goal` turn 내 agent가 사용자에게 prose 질문 출력 (비대칭 boundary 위반).
- **AP-6**: destructive op(force-push, rm -rf, drop table, push, clean 삭제)을 active `/goal`이 사전 승인 — 평가자는 continue 여부만 판정, destructive 승인 안 함.
- **AP-7**: background Agent가 Write/Edit 수행 (clean 삭제를 background로 — foreground 강제 위반).
- **AP-8**: turn/time bound 없는 `/goal` 조건 → unbounded 토큰 spend (항상 "max N turns" 임베드).
- **AP-9**: 코딩-heavy 신규 구현을 Workflow fan-out으로 분산 (Finding A4 — Mode 5 sequential 우선).
- **AP-10**: subagent가 다시 subagent를 spawn (Finding A1 — flat 계층 위반).

---

## 8. 구성(config) 제안

### 8.1 `.moai/config/sections/workflow.yaml` autonomy profile 스키마

서브커맨드별 "autonomy profile"을 선언적으로 정의해 오케스트레이터가 런타임에 엔진·조건·보존 게이트를 결정한다.

```yaml
workflow:
  # 기존 team 설정과 공존 (별도 키)
  team:
    enabled: false
    role_profiles: { ... }   # 기존 유지

  # 신규: 서브커맨드별 autonomy profile
  autonomy:
    enabled: false           # [HARD] 기본 off — 명시 opt-in (research preview)
    default_max_turns: 20    # /goal bound 기본값 (unbounded 방지)

    profiles:
      plan:
        engine: hybrid                    # workflow + goal
        goal_condition_template: "plan_audit_pass"   # §5.1 참조
        goal_launch_precondition: "plan_audit.enabled == true"
        workflow_pattern: "multi_angle_research"     # §6.1
        preserved_gates: [dp1, tier, bodp, quality_gate, exec_mode]
        max_turns: 3                       # Phase 2.3 hard cap
      run:
        engine: goal_loop                 # goal + /moai loop
        goal_condition_template: "ac_converge"       # §5.2
        workflow_pattern: null            # 코딩-heavy = sequential (Finding A4)
        workflow_pattern_conditional: "mechanical_migration_ge_30_files"  # §6.2
        preserved_gates: [gate2]          # [HARD] 절대 자율 bypass 금지
        max_turns: 20
      sync:
        engine: hybrid
        goal_condition_template: "locale_parity_zero"  # §5.3
        workflow_pattern: "doc_parity_fanout"          # §6.3
        preserved_gates: [pr_creation, merge_approval, destructive_push]
        max_turns: 15
      loop:
        engine: ralph_goal                # /moai loop + goal 래핑
        goal_condition_template: "quality_converge"    # §5.4
        workflow_pattern: null            # 순차 수정
        preserved_gates: [level3_fix, memory_checkpoint]
        max_turns: 50
      review:
        engine: hybrid
        goal_condition_template: "high_severity_zero"  # §5.6
        workflow_pattern: "multidim_adversarial"       # §6.4
        preserved_gates: [finding_apply_write]
        max_turns: 10
      clean:
        engine: ralph_goal
        goal_condition_template: "deadcode_zero"
        workflow_pattern: null            # [HARD] 삭제 = foreground, fan-out 금지
        preserved_gates: [deletion_approval, destructive]   # [HARD]
        max_turns: 15
      # ... project, coverage, e2e, gate, mx, codemaps, brain, design, db ...

      # 자율 미적용 (단발/인간결정 heavy)
      harness: { engine: none }
      feedback: { engine: none }
```

### 8.2 기존 team/role_profiles 와의 관계

- `workflow.team` (기존)과 `workflow.autonomy` (신규)는 **별도 키, 공존**. team은 Agent Teams(3-5 teammate) 모드, autonomy는 엔진 선택(goal/workflow/loop). Phase 0.95 모드 선택(`orchestration-mode-selection.md`)이 두 키를 모두 읽어 trivial/background/agent-team/parallel/sub-agent/**workflow**(신규 Mode 6) 중 결정.
- `role_profiles` (researcher/analyst/architect/implementer/tester/designer/reviewer)는 Workflow fan-out 에이전트의 prompt 합성에도 재사용 가능 — Workflow의 read-only fan-out 단위가 researcher/analyst profile을 상속, 단 Workflow agent는 사용자 prompt 불가 제약 추가.
- **[HARD] autonomy.enabled 기본 off**: research preview이고 org/user가 `disableWorkflows`/`CLAUDE_CODE_DISABLE_WORKFLOWS`로 비활성 가능하므로 항상 가용하다고 가정 금지. preflight에서 Claude Code 버전(workflows ≥2.1.154, /goal ≥2.1.139) + hook 활성(/goal은 hook 의존, `disableAllHooks`/`allowManagedHooksOnly` 시 불가) 검증.

---

## 9. 단계별 도입 로드맵

### Priority High

- **SPEC-AUTONOMY-RUN-GOAL** (Tier M): `/moai run` Phase 0.95 모드 카탈로그에 Mode 6(workflow) 추가 + run-phase `/goal` 래핑(§5.2) wiring. **구현 착수 승인 보존 회귀 테스트 필수** — plan-auditor 점수 무관 AskUserQuestion 게이트 유지 검증. run group design이 rate-limit으로 미전달되었으므로 이 SPEC의 plan-phase에서 design을 정식 author + adversarial 재검증(verdict 6조항 충족 확인) 필요.
- **SPEC-AUTONOMY-GOAL-CONDITIONS** (Tier S): §5 /goal 조건 템플릿 카탈로그를 `.claude/rules/moai/workflow/goal-directive.md`에 정식 등재 + 각 서브커맨드 skill body에 조건 emit 로직 wiring. plan group의 4개 교정(파일경로→transcript술어, STOP-regression bound, harness 게이트, 중복절 제거) 반영.

### Priority Med

- **SPEC-AUTONOMY-WORKFLOW-PATTERNS** (Tier M): §6 Workflow 스크립트 패턴 5종을 `dynamic-workflows.md`에 패턴 카탈로그로 등재 + plan/sync/review/sweep skill body에 launch 트리거 wiring. **named 스크립트 API 단정 금지**(docs 미문서화) — coordinate-agents 개념 모델만 기술. AskUserQuestion boundary 보존 + launch-전-preference-수집 패턴 명시.
- **SPEC-AUTONOMY-CONFIG** (Tier M): §8 `workflow.yaml` autonomy profile 스키마를 Go nested struct(`internal/config`)로 구현 + accessor + template SSOT + preflight(버전/hook 검증). 기존 team 키와 공존 보장 테스트.
- **SPEC-AUTONOMY-PLAN-AUDIT-GOAL** (Tier S): plan group 검증 완료 design을 §5.1 교정 적용해 wiring. harness `minimal` 게이트 + STOP-on-regression 이탈 테스트.

### Priority Low

- **SPEC-AUTONOMY-SYNC-PARITY** (Tier M): sync-phase 4-locale doc-parity Workflow(§6.3) + /goal(§5.3). sync→PR 게이트 보존 회귀 테스트. sync group design 정식 author + adversarial 재검증 필요(rate-limit 미전달분).
- **SPEC-AUTONOMY-SWEEP** (Tier M): mx/codemaps/clean Workflow + /goal. **clean foreground-강제 + 삭제 승인 게이트** 회귀 테스트. sweep group design 정식 author 필요.
- **SPEC-AUTONOMY-META** (Tier S): brain/design/db/harness/feedback autonomy profile. brain Phase 전환 게이트 + db migration destructive 확인 보존.

> 로드맵 주의: run/sync/fixloop/quality/sweep/meta/project group의 원본 design이 rate-limit으로 미전달되어 adversarial verdict가 "missing-input UNSOUND"였다. 이는 **내용 거부가 아니라** 검증 입력 부재이므로, 각 후속 SPEC의 plan-phase에서 design을 정식 author하고 adversarial 재검증(verdict 6조항: 구현 착수 승인 mandatory / preference-launch전수집 / no-nested-spawn / background-read-only / transcript-측정가능 goal / workflow-병렬·loop-진단)을 통과시킨 뒤 run-phase 진입할 것.

---

## 10. 리스크 및 완화

| 리스크 | 설명 | 완화책 |
|--------|------|--------|
| **dark-flow** | 자율 모드가 사용자 인지 없이 진행 | autonomy.enabled 기본 off + launch 전 AskUserQuestion으로 엔진/조건 명시 + progress.md 로깅 + /goal `◎` 인디케이터 + bare `/goal` 상태 조회 |
| **비용 폭증** | Workflow 1 run이 대화 대비 토큰 多, ultracode는 모든 task를 workflow화 | launch 전 AskUserQuestion으로 비용 trade-off surface + 작은 slice(단일 디렉터리/좁은 질문)로 spend 측정 후 확장 + context-window 임계(1M=50%) 카운트 + ultracode는 명시 opt-in만 |
| **goal false-completion** | 평가자가 transcript만 보고 잘못 PASS 판정 | 조건을 surface된 verifiable 라인으로 작성(파일경로 금지) + ground-truth가 필요하면 /moai loop(도구 검증) 사용 + must-pass 차원은 judge/auditor 별도 검증 |
| **goal unbounded spin** | bound 없는 조건이 영원히 미충족 | [HARD] 모든 조건에 "max N turns/iterations" 임베드 + harness 게이트(plan_audit.enabled) + STOP-on-regression 조기 이탈 |
| **workflow 토큰 비용 + 컨텍스트** | 대규모 fan-out이 plan usage·rate limit 소비 | 동시 16 / 총 1000 cap 인지 + 중간 결과는 스크립트 변수(컨텍스트 오염 0)이나 최종 합성은 컨텍스트 카운트 + allowlist 사전 등록(mid-run stall 방지) |
| **multi-session race** | 병렬 세션이 동일 SPEC 작업 | pre-spawn `git fetch origin main` + `git rev-list --count` + `moai session list --json --filter-spec` 3-command 배치 (agent-common-protocol.md) — 자율 launch 전에도 실행 |
| **destructive 자율 실행** | clean 삭제, db migration, force-push를 /goal/workflow가 실행 | [HARD] destructive op은 active /goal/workflow가 사전 승인 안 함 — 별도 AskUserQuestion 확인 + clean/db는 foreground 강제 |
| **버전/hook 부재 환경** | workflows<2.1.154 또는 hook 비활성 환경 | preflight 버전 검증 + /goal은 hook 의존(disableAllHooks/allowManagedHooksOnly 시 불가) 감지 → graceful degradation(자율 off, 수동 흐름) |
| **cross-session 미재개** | Workflow는 세션 내에서만 resume, exit 후 fresh | 장기 작업은 paste-ready resume(session-handoff.md) + `ultrathink.` opener + 재설정 /goal 페어링으로 /clear 후 자율 루프 복원 |
| **named script API 환각** | 문서 미존재 스크립트 함수를 단정 | [HARD] MoAI 문서는 typed workflow script API 단정 금지 — "phase"는 관찰 가능성 grouping일 뿐. coordinate-agents 개념 모델만 기술 |

---

> **종합**: 본 전략은 doctrine 레이어에만 머물던 두 primitive(Dynamic Workflows / `/goal`)를 기존 `/moai loop`과 3-엔진 모델로 통합해 모든 서브커맨드의 phase *내부* 자율을 실현하되, 구현 착수 승인·AskUserQuestion 독점·flat 계층·background-write 금지의 4대 불변식과 모든 phase *경계* 게이트(구현 착수 승인, sync→PR, destructive, ambiguity, hook-block, CI autofix iter>3)를 보존한다. plan group의 검증된 4개 교정은 §4.2-A·§5.1에 반영했고, rate-limit으로 design이 미전달된 그룹은 후속 SPEC의 plan-phase에서 정식 author + adversarial 재검증을 거치도록 로드맵에 명시했다.
