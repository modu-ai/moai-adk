# Plan — SPEC-WORKFLOW-RULE-SPLIT-001

## §A 맥락

카드 t1586 (lane-12, run tmhxo0). 측정 트리 `.moai/worktrees/spec-workflow-split` @ `81786284e`, 브랜치 `WT-spec-workflow-split`. `.claude/rules/moai/workflow/spec-workflow.md` 가 per-file 40,000 runes 예산을 81 runes 초과(실측 40,081 runes / 40,340 bytes). 수리 방향은 카드 지시대로 **core stub + lazy companion 분할**이고, 선행은 `factory-dispatch.md` ↔ 4-companion 분할(t1483 계열)이다.

측정 근거(모두 이 트리에서 관측):

| 관측 | 명령 | 값 |
|---|---|---|
| runes | `python3 -c "len(open(...).read())"` | 40,081 |
| bytes | `wc -c` | 40,340 |
| 예산 상수 | `internal/hook/instructions_loaded.go:124` | `const charBudget = 40000` (runes, `utf8.RuneCount`) |
| CI 예산 테스트 부재 | `grep -rn TestDeployedAlwaysLoadedCharBudget internal/` | 0 적중 |
| 템플릿 쌍 현재 상태 | `cmp` (local ↔ templates) | byte-identical |

### Tier 판정 — M

| 축 | 값 | Tier M 상한 | 여유 |
|---|---|---|---|
| REQ | 8 (REQ-001..008) | 16 | 8 |
| AC | 10 (AC-01..10) | 16 | 6 |

파일 수 13개(변경 10 — core 재작성 1·cp 교체 1·등재 코드 2·인용 갱신 6 — + 신규 2 + 재생성 1; §C 행 합계와 일치), 마일스톤 3개 — S 아님(Tier S = <5 파일). L 아님(>15 파일 아님, constitutional 아님). SPEC 산출물은 3개(spec/plan/acceptance — Tier M 표준 세트).

### 편집 방향 (거울 방향) — 소스 = 배포 트리, mirror = templates 쌍

`internal/template/rule_template_mirror_test.go` 가 방향을 규정한다: source = `.claude/rules/...`, mirror = `internal/template/templates/.claude/rules/...`, 드리프트 시 오류문이 `run 'cp <src> <mirror>'` 를 지시한다(L352, L362). 즉 **콘텐츠 편집은 배포 트리에서 먼저, 템플릿 쌍은 cp 로 동기화**한다(Template-First 는 배포 원칙이고, 이 리포 개발 사이클에서 규칙 원문의 작성 지점은 배포 트리다 — mirror 테스트의 cp 지시가 그 증거). 신규 파일(companion)은 배포 트리에 작성 → `cp` 로 templates 쌍 생성 → 등재 코드 변경 → `make embed-manifest` 재생성 순서.

### go:embed 발행 경로

`internal/template/embed_manifest_gen.go` 는 파일당 `//go:embed` 라인을 두는 **생성 파일**이다. 신규 템플릿 파일은 `make embed-manifest`(`EMBED_MANIFEST_UPDATE=1 go test ./internal/template/embedemit/... -run TestGoldenCommittedArtifactsMatchEmission`)로 재생성하고, 판정은 `make embed-manifest-check`(read-only)로 한다. `make build` 가 `embed-manifest-check` 을 포함하므로 최종 게이트에서도 재확인된다.

## §B 마일스톤

### M1 — 분할 (배포 트리, 콘텐츠 이동)

**M1 종료조건**: core ≈ 28.8KB 이하, companion 존재, 이동 원문 verbatim, 등록 clause 3건·Tier 표·skip 계약이 core 에 남음.

파일별 변경:

1. `.claude/rules/moai/workflow/spec-workflow.md` — **제자리 재작성(core stub)**.
   - 유지(바이트 동일): frontmatter(`paths:` 불변), Phase Overview, 2-라우트 [HARD] 정의(L19-29), step ordering [HARD] bullets(L49-61 — CONST-V3R5-027/-028 원문 포함, Late-branch bash 블록과 post-condition이 bullet L53 내부라 통째 유지), Conditional Design Route, Subcommand Classification 전체(L75-129), SPEC Complexity Tier 전체(L130-161), Plan Phase 전체(L162-187), Run Phase core(L188-199, 219-228, 241-281), Sync Phase, Context Management, Phase Transitions 전체(L311-360 — skip 계약 포함, 단 L333-334의 "see § Report Persistence" → "see `spec-workflow-detail.md` § Report Persistence" 1문장 수정), Phase 1 core(L361-371, 390-398, 414-420), Agent Teams Variant.
   - 신규(포인터): 헤더 직후 detail companion 안내 blockquote; 이동 블록 자리마다 한 줄 포인터 — Route 표 자리("Route A/B step tables + merge_method(기본 squash): `spec-workflow-detail.md` § Route tables"), anti-patterns 자리, Run 방법론 자리("DDD/TDD 모드 본문: detail § Run-phase methodology"), Depends_on Pre-flight 자리("detail § Depends_on Pre-flight Check"), Report Persistence 자리("detail § Report Persistence").
   - `#plan-phase`, `#spec-phase-discipline`, `#subcommand-classification(-pipeline-vs-multi-agent)`, `#mode-dispatch-cross-reference`, `#phase-overview`, `## Phase 1: Plan Audit Gate`, `## SPEC Complexity Tier` 헤딩 토큰 전부 보존(앵커·레지스트리·Go 파서 구속).
2. `.claude/rules/moai/workflow/spec-workflow-detail.md` — **신규**.
   - frontmatter: `description:`(소유 섹션 한 문장) + `paths: "**/spec-workflow.md,**/.claude/rules/moai/workflow/spec-workflow-detail.md"` — 선행 `cache-aware-execution-reference.md` 패턴.
   - 본문: 제목 `# SPEC Workflow — Detail Companion` + 헤더 blockquote("Detail companion of `spec-workflow.md` … The stub keeps every [HARD] rule … Load when …") + 이동 블록 원문(Route A/B 표+각주 / anti-patterns+xref / DDD·TDD 모드·self-review·drift·delegation / Depends_on Pre-flight / Report Persistence) + footer(Version/Classification).
   - 이동 블록은 헤더 조립만 하고 본문 문장은 원본 그대로(REQ-008).

### M2 — 미러 + 등재 코드

3. `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` — `cp` (M1 산출물로 교체).
4. `internal/template/templates/.claude/rules/moai/workflow/spec-workflow-detail.md` — 신규 `cp`.
5. `internal/template/rule_template_mirror_test.go` — `workflowOptMirroredPaths` 에 `".claude/rules/moai/workflow/spec-workflow-detail.md"` 추가 + 등재 사유 주석(카드 t1586 분할, 토큰 클린성 판정 결과 명시). 신규 등재는 "deliberate code change visible in PR review" 계약(L179-180)에 따른 것.
6. `internal/template/internal_content_leak_test.go` — allowance 항목 1건 추가: `{File: ".claude/rules/moai/workflow/spec-workflow-detail.md", LineStart: 0, LineEnd: 0, SpecID: "SPEC-AUDIT-SNAPSHOT-001", Rationale: "Mirror-parity-enforced provenance (…); internal SPEC provenance retained on both trees"}` — 이동되는 Report Persistence 본문이 `SPEC-AUDIT-SNAPSHOT-001 A1` 을 실어 가므로(원본 L408 실측). 그 외 이동 블록은 내부 토큰 0건(Route 표·anti-patterns·방법론·pre-flight 실측). 부모 파일과 동일한 취급(M1 유지 원문 정책)의 일관이다.
7. `make embed-manifest` 실행 — `embed_manifest_gen.go` 재생성(companion 템플릿 임베드 라인 추가).

### M3 — 크로스레퍼런스 정산 + 검증 배치

8. `.claude/skills/moai/workflows/run/phase-execution.md` L25 — "`spec-workflow.md` § Report Persistence" → "`spec-workflow-detail.md` § Report Persistence". 템플릿 쌍 동일 편집.
9. `.claude/agents/moai/plan-auditor.md` L655 — 동일 문구 교체. 템플릿 쌍 동일 편집(이 쌍은 §25 sanitized pair — byte-parity 아님, 양쪽에 같은 포인터 문장을 각각 적용).
10. `.claude/skills/moai/workflows/run/context-loading.md` L19 — "(Run Phase section)" 방법론 안내 → "`spec-workflow-detail.md` (Run-phase methodology section)". 템플릿 쌍 동일 편집.
11. 검증 배치(acceptance.md §D 전체 — 병렬 일괄).

## §C 파일 변경 목록 요약

| # | 파일 | 동작 | 소관 M |
|---|---|---|---|
| 1 | `.claude/rules/moai/workflow/spec-workflow.md` | 재작성(core stub) | M1 |
| 2 | `.claude/rules/moai/workflow/spec-workflow-detail.md` | 신규 | M1 |
| 3 | `internal/template/templates/…/spec-workflow.md` | cp 교체 | M2 |
| 4 | `internal/template/templates/…/spec-workflow-detail.md` | cp 신규 | M2 |
| 5 | `internal/template/rule_template_mirror_test.go` | allowlist +1 | M2 |
| 6 | `internal/template/internal_content_leak_test.go` | allowance +1 | M2 |
| 7 | `internal/template/embed_manifest_gen.go` | 재생성 | M2 |
| 8-9 | `run/phase-execution.md` ×2트리 | 인용 경로 교체 | M3 |
| 10-11 | `agents/moai/plan-auditor.md` ×2트리 | 인용 경로 교체 | M3 |
| 12-13 | `run/context-loading.md` ×2트리 | 방법론 안내 교체 | M3 |

## §D 크로스레퍼런스 인벤토리 (repo 전수 실측)

`spec-workflow.md` 를 인용하는 파일 36개(§-이름/앵커/일반 참조 합산). 해소 판정:

| 인용 형태 | 건수 | 이동? | 조치 |
|---|---|---|---|
| `spec-workflow.md#subcommand-classification(-pipeline-vs-multi-agent)` | 49 외부 + 자기인용 1 (.claude/ 12 · internal/ 12 · .moai/ 25) | 아님 | 없음 (core 유지) |
| `spec-workflow.md#mode-dispatch-cross-reference` | 6 | 아님 | 없음 |
| `spec-workflow.md#phase-overview` | 1 | 아님 | 없음 |
| `§ SPEC Complexity Tier` (`.claude/rules/**.md` 슬라이스 한정 — .go 주석 슬라이스는 별도 9건) | 7 | 아님 | 없음 (Go 파서 대상 표 core 유지) |
| `§ Subcommand Classification` / `§ Mode Dispatch` / `§ Phase 1 Plan Audit Gate` / `§ Plan Audit Gate skip policy` | 6 | 아님 | 없음 |
| `§ SPEC Phase Discipline` / 앵커 `#plan-phase`, `#spec-phase-discipline` (zone-registry 3엔트리 포함) | — | clause 원문 core 유지 | 없음 |
| **`§ Report Persistence`** (plan-auditor.md:655, run/phase-execution.md:25 — 각 ×2트리) | 4 | **이동** | **companion 경로로 교체** |
| **`(Run Phase section)` 방법론 안내** (run/context-loading.md:19 ×2트리) | 2 | 방법론 본문 이동 | **companion 경로로 교체** |
| `@spec-workflow.md` 일반 참조 (DDD/TDD "for details" 등) | 다수 | 파일 존속 | 해소됨(필요 시 정밀화, 의무 아님) |
| Go 코드 고정 경로 `tierArtifactRuleRelPath` (`internal/spec/lint_tier_artifacts.go:36`) | 1 | 아님 | 없음 (Tier 표 core 유지로 파서 불변) |
| mirror allowlist / leak allowance / embed manifest (내부 등재) | — | companion 추가 필요 | M2 코드 변경 |
| `paths:` 로 spec-workflow.md 를 로딩 스코프에 두는 규칙 | 1 (`cache-aware-execution-reference.md` — `**/.claude/rules/moai/workflow/*.md` 와일드카드) | — | 없음 (companion 이 같은 디렉터리라 같은 스코프에 들어옴 — 오히려 의도된 lazy 로딩) |

앵커 슬러그 주의: `## Subcommand Classification (Pipeline vs Multi-Agent)` 의 슬러그가 `#subcommand-classification-pipeline-vs-multi-agent` — core 재작성 시 헤딩 텍스트를 **한 글자도** 바꾸지 않는다.

## §E 예산 재측정 명령 (카드 질의 3의 답)

존재하는 집행기는 런타임 hook 이다(`§1.1` 표). CI 테스트는 없다. run-phase 가 실행할 판정 명령:

```bash
# 양쪽 트리 core 모두 — runes 기준 (hook 과 동일 단위)
python3 -c "print(len(open('.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"
python3 -c "print(len(open('internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"
# 교차 확인 (locale UTF-8 가정 하 wc -m)
wc -m .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
# 회귀 확인 (hook 셀프테스트 — 픽스처 기반, 실측 아님을 명시)
go test ./internal/hook/ -run TestInstructionsLoaded -count=1
```

RED 값(현 트리): 40,081 / 40,081 (40,340 bytes). GREEN: 양쪽 모두 **39,999 이하** — 목표 28.8K 부근(≈28% 여유).

## §F 등재·검증 명령 (M2/M3)

```bash
cp .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
cp .claude/rules/moai/workflow/spec-workflow-detail.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow-detail.md
make embed-manifest
make embed-manifest-check
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak|TestDeclaredRuleMirrorForks' -count=1
go test ./internal/constitution/ -count=1
go test ./internal/spec/ -count=1        # TierArtifactMissing/Unreadable 포함
make constitution-check                   # 빌드 후 bin/moai 로 레지스트리 검증
make build                                # embed-manifest-check 포함 최종 게이트
```

## §G mutant 대응

| mutant | 이를 잡는 AC |
|---|---|
| 등록 clause 를 companion 으로 옮기고 registry `file:` 만 수정 | AC-04 (registry 3 clause × core grep -F), REQ-001 |
| Tier 표를 companion 으로 이동 | AC-05 (`internal/spec` Tier 파서 테스트 + "Artifact set" 헤더 core grep) |
| 포인터 없이 섹션 삭제 (내용 유실) | AC-07 (이동 블록 원문 verbatim 존재 — 블록별 시그니처 라인 grep) + AC-06 |
| 한쪽 트리만 수정 | AC-02 (양쪽 `cmp` byte-identical), AC-03 |
| skip 계약을 companion 으로 이동 | AC-04 (skip 계약 시그니처 core grep) |
| 큰 통로(빈 블록)를 남기고 예산만 회피 | AC-01(예산) + AC-07(원문 보존) 동시 충족이 불가능한 구조 |
| 인용 경로 교체 누락 | AC-06 (`spec-workflow.md § Report Persistence` 잔존 0건 grep) |

## §H 리스크·비고

- **zone-registry 앵커**: 헤딩 보존의 판정 주체는 AC-04의 `grep -c "^## Plan Phase$"` / `grep -c "^## SPEC Phase Discipline$"` 다. `constitution-check` 은 **clause 텍스트만** 검증한다 — validator 의 `SentinelAnchorNotFound` 는 정의만 있고 발화 지점이 없는 죽은 코드다(`internal/constitution/validator.go:27-28` 정의, 발화 0건 실측). `make constitution-check` 이 clause 축의 최종 판정.
- **Run Phase 중간 발췌**: L200-218(DDD·TDD 모드 본문)·L229-239(self-review + drift + delegation)를 이동하고 L188-199·L219-228(Methodology Auto-Detection — core 잔류)·L241-281 은 남는다 — 같은 `## Run Phase` 섹션 안의 부분 이동. core 의 남는 블록 순서는 원본 순서를 유지하고 사이에 포인터를 넣는다.
- **leak allowance 최소성**: companion 이 실어 가는 내부 토큰은 `SPEC-AUDIT-SNAPSHOT-001` 1종(실측) — allowance 1건으로 닫힌다. M2에서 move 후 전수 재확인(`grep -oE "SPEC-[A-Z0-9-]+" spec-workflow-detail.md | sort -u`).
- **contract-mode allowlist**: `TestContractModeChangeSetAllowlist` 가 spec-workflow.md 를 contract-mode 변경 금지 경로로 둔다 — 본 SPEC 은 contract-mode 세션이 아니므로 무관하나, 편집 경로가 contract 모드로 돌면 거부된다(§A 트리 환경 실측 기록).
- 예산 단위 주의: hook 은 runes(`utf8.RuneCount`), `wc -c` 는 bytes — 두 수를 혼용하지 않는다. 판정은 runes.
