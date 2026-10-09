---
id: SPEC-RELUP-DUALAXIS-001
title: "plan — release-update 하네스 CC+Codex 이중 축 정착"
created: 2026-10-09
author: manager-spec
tier: M
---

# plan: SPEC-RELUP-DUALAXIS-001

## §A Context

- **측정 트리**: `.moai/worktrees/t1579` (branch `WT-high-10-07`) @ `2aab5f797`. 모든 RED-now 관측은 이 SHA에서 수행했고 acceptance.md §D.3 증거 원장에 전수 기록돼 있다.
- **카드**: t1579 (High·운영자 확장 지시 2026-10-07·builder-harness/SPEC 소관).
- **SPEC artifacts**: `.moai/specs/SPEC-RELUP-DUALAXIS-001/{spec,plan,acceptance,progress}.md` — Tier M 3-artifact 세트 + progress.md.
- **변경 표면 (3개 — 전부 사용자 소유 dev-only 네임스페이스)**:
  1. `.claude/agents/harness/hns-release-update-specialist.md` (280행 — codex 축 Phase + 상태 스키마 문서화 + BP 상시 섹션 + Phase 3 URL 세트)
  2. `.claude/workflows/hns-release-update-run.js` (175행 — codex 렌즈 fan-out + 커밋 복원 폴백 + 6테마 체크리스트)
  3. `.claude/commands/harness/release-update/manifest.json` (29행 — domain 문자열만)
- **Go 코드 변경: 0.** 템플릿 미러: 없음(불요 — §1.1 M9). `make build` 불요.
- **근거 연구**: `.moai/research/upstream-update-20261007.md` (1차 — CARD-1 정의 + 6테마 표 + Phase 7.5 findings 2건) / `upstream-update-20261008.md` (2차 — codex 0.161.0 승격 큐레이팅 + URL 세트 finding).
- **선례 SPEC**: SPEC-UPDATE-ADD-CODEX-001 (completed) — codex 배선 선례로 인용, 스코프 중복 없음(그쪽은 `moai update --add-codex` CLI 동사).

### §A.5 PRESERVE 목록 (건드리지 않는다)

| 표면 | 보존 이유 |
|---|---|
| `.moai/state/last-cc-version.json` (primary 체크아웃, 기계 로컬) | CC 축 상태 파일 — 스키마·내용 불변. run-phase가 편집 금지 |
| manifest.json `source_request` 필드 | 3-harness 분리의 역사 서술(프로비넌스) — 갱신 대상 아님 |
| manifest.json `sprint_contract.dimensions` + thresholds | 결정 D4 — Functionality/Consistency 0.85/0.80 불변 |
| manifest.json `entry_command` / `runner_workflow` / `specialists[]` (primitive·isolation) | 하네스 골격 불변 — `hns-release-update-run.js` 참조 1힛 유지 (AC-RDX-013) |
| 스페셜리스트 Phase 5 (인간 게이트) · Phase 6 (docs 4-locale) · Phase 7b (manager-git PR) · Phase 7.5 (REQ-HRR-006 findings) · Phase 8 | 기존 절차 전부 유지 — codex/BP 확장은 추가다 |
| Runner HARD 제약 (AskUserQuestion·gh pr 금지, Date.now()/Math.random() 금지, top-level 실행 + CommonJS export 가드) | AC-DHC-007a + 결정성 계약 |
| `.moai/research/upstream-update-*.md` 2건 | 읽기전용 연구 산출물 — 편집 금지 |
| 타 SPEC 디렉터리 · `.moai/reports/` | 스코프 밖 (B10) |

## §B Known Issues (관련 카테고리만)

- **B4 Frontmatter 스키마**: spec.md 12 필드 canonical (`created:`/`updated:`/`tags:` — snake_case 금지). 본 plan/acceptance는 status 축 stateless(§ Artifact Statelessness).
- **B6 Out of Scope 헤딩**: `## Out of Scope` h2 단독은 `MissingExclusions` lint ERROR — `### Out of Scope — <topic>` h3 + `-` 불릿으로 작성했다(spec.md §6).
- **B8 워킹 트리 위생**: 런타임 관리 파일(`.moai/harness/`, `.moai/state/`) 편집 금지. 커밋은 지정 pathspec만.
- **B10 범위 규율**: §A.5 PRESERVE 이외 무변경. 특히 Go 트리(`internal/`) 0변경 — AC-RDX-011이 지키는 회귀 가드다.
- **B11 사용자 질의 금지**: leaf worker — 열린 질문은 전부 spec.md §1.2 결정 기록으로 봉쇄했다. `[NEEDS CLARIFICATION]` 마커 0개.
- **상태 파일 특이사항**: `last-codex-version.json`은 gitignored 기계 로컬이라 CI가 판정할 수 없다 — AC는 본문 쓰기 지점을 측정면으로 삼는다(§5.2).
- **node --check 한계**: 러너 JS 파스 검증에 `node --check`는 무음 통과 한계가 있다(운영 교훈) — §E에서 CommonJS require() 스모크로 보강한다.

## §C Pre-flight (run-phase 진입 시 재실행)

```bash
git branch --show-current ; git rev-parse --short HEAD     # WT-high-10-07 이후 재확인
# RED-now 앵커 11종 재측정 (acceptance.md §D.3 원장의 명령 그대로 — 전부 단일 호출):
grep -c "Codex CLI upstream change tracking" .claude/commands/harness/release-update/manifest.json   # 기대 0 (M4 전)
grep -ci "best-practice" .claude/commands/harness/release-update/manifest.json                        # 기대 0 (M4 전)
grep -c "selectCodexSweepTargets" .claude/workflows/hns-release-update-run.js                          # 기대 0 (M2 전)
grep -c "CODEX_COMMITS_FALLBACK" .claude/workflows/hns-release-update-run.js                           # 기대 0 (M2 전)
grep -c "CODEX_THEME_CHECKLIST" .claude/workflows/hns-release-update-run.js                            # 기대 0 (M2 전)
grep -c "last-codex-version.json" .claude/agents/harness/hns-release-update-specialist.md              # 기대 0 (M1 전)
grep -c "rust-v0.161.0" .claude/agents/harness/hns-release-update-specialist.md                        # 기대 0 (M1 전)
grep -ci "best-practice" .claude/agents/harness/hns-release-update-specialist.md                       # 기대 0 (M3 전)
grep -c "code.claude.com" .claude/agents/harness/hns-release-update-specialist.md                      # 기대 0 (M3 전)
grep -c "HTML proposal report" .claude/agents/harness/hns-release-update-specialist.md                 # 기대 0 (M3 전)
grep -c "source-first" .claude/agents/harness/hns-release-update-specialist.md                          # 기대 0 (M3 전)
# PRESERVE 앵커 3종:
grep -rn "last-codex-version" internal/   # 0힛 유지 (AC-RDX-011)
grep -c "last-cc-version.json" .claude/agents/harness/hns-release-update-specialist.md                 # ≥3 유지 (AC-RDX-012)
grep -c "hns-release-update-run.js" .claude/commands/harness/release-update/manifest.json              # 1 유지 (AC-RDX-013)
```

**시드값 재판정 (M1 착지 직전)**: `npm view @openai/codex version` + `gh api repos/openai/codex/releases?per_page=5` 재실행 — 최신 비프리릴리즈가 `rust-v0.161.0`이 아닌 값이면 시드를 그 값으로 갱신하고 spec.md D1·REQ-RDX-002·AC-RDX-007을 동시 갱신한다(§3 계층 수정 규율). 승격이 없으면 시드 고정.

## §D Constraints (앵커 고정 — run-phase 재량 금지)

### §D1 고정 앵커 (grep 판정면 — spec.md §1.2·§4와 쌍)

| 표면 | 고정 앵커 (리터럴) | AC |
|---|---|---|
| manifest.json `domain` | 부분 문자열 `Claude Code + Codex CLI upstream change tracking` | AC-RDX-001 |
| manifest.json `domain` | 부분 문자열 `best-practices axis` | AC-RDX-002 |
| runner | 식별자 `selectCodexSweepTargets` — `selectResearchSweepTargets`와 병렬 함수 (동일 반환 형태: purpose/agentType/isolation/label/prompt) | AC-RDX-003 |
| runner | 상수 `CODEX_COMMITS_FALLBACK` — 본문 비어 있을 때의 커밋 API 복원 절차 문서 블록 앵커. 절차 내용: (1) 릴리즈 본문 1줄 제목만 관측되면 `gh api repos/openai/codex/commits`/`pulls` 주제 복원, (2) 복원 항목 전부 "commit-topic-derived" 라벨, (3) 잠재 티어1 후보는 PR 본문 확인으로 격상(#49713 정합 절차) | AC-RDX-004 |
| runner | 상수 `CODEX_THEME_CHECKLIST` — 6테마 리터럴 `thread` / `rollout` / `subagent` / `compaction` / `MCP` / `other`. 행 형식: 테마 키 + 관측 PR 번호 목록 + MoAI 노출면 | AC-RDX-005 |
| specialist | `last-codex-version.json` — Phase 0(판독·부재 기본값) + Phase 7a(쓰기)에 등장 | AC-RDX-006 |
| specialist | `rust-v0.161.0` — 시드 기술 | AC-RDX-007 |
| specialist | `Best-Practices` 상시 절차 섹션 (헤딩 문자열 대소문자 무관 `best-practice` 매치) | AC-RDX-008 |
| specialist | `code.claude.com/docs/en/` — Phase 3 URL 세트 갱신 (기존 `docs.anthropic.com/en/docs/claude-code/*` 6종을 캐노니컬 형태로) | AC-RDX-009 |
| specialist | `HTML proposal report` — BP 축 명명 산출물 | AC-RDX-010 |
| specialist | `rust-v0.161.0` 부재 기본값 — AC-RDX-006의 스키마 블록이 Phase 0 부재-기본값(`rust-v0.161.0` + 경고)을 포함해야 한다 (앵커 LED-006 공유) | AC-RDX-014 |
| runner | alpha watch 규범 — AC-RDX-005의 체크리스트 블록이 "alpha 테마는 watch 관찰목록, 안정 탑재 시에만 채택 판정"을 포함해야 한다 (앵커 LED-005 공유) | AC-RDX-015 |
| specialist | `source-first` — 원문 패치 선행 강제 리터럴 (REQ-RDX-013, mutant M-4의 기계 판정면) | AC-RDX-016 |

### §D2 결정 전파 (재논의 금지 — spec.md §1.2)

- D1 시드 `rust-v0.161.0` — 변경 시 spec.md D1 + REQ-RDX-002 + plan §C + AC-RDX-007 4곳 동시 갱신.
- D4 — manifest `sprint_contract` 무변경. 변경 제안은 별도 카드로.
- D7 — 스윕 실행·상태 파일 실생성 금지 (run-phase에서도).

### §D3 금지 목록

- `internal/`·`internal/template/templates/` 편집 금지 (Go 0변경 — AC-RDX-011).
- `--no-verify`·force-push 금지. 커밋은 Conventional + 카드 id + `🗿 MoAI` 트레일러.
- 상태 파일(`.moai/state/*.json`) 편집 금지.
- `[NEEDS CLARIFICATION]` 마커 신설 금지 — 판단 필요 시 결정 기록으로 자결.

## §E Self-Verification (run-phase 납품)

- **E1 AC 매트릭스** — acceptance.md §D 13종 PASS/FAIL + 검증 명령 + 실측 출력 (§E 삼중 귀속: 명령·출력·HEAD SHA).
- **E2 RED→GREEN 전수 재측정** — §C의 11종 RED 앵커가 대응 마일스톤 착지 후 뒤집혔는지 exit code 포함 재실행.
- **E3 러너 파스 스모크** — `node --check`의 무음 통과 한계(§B)를 보강하는 CommonJS 경로 스모크:
  `node -e "const m = require('./.claude/workflows/hns-release-update-run.js'); console.log(typeof m.run, typeof m.selectResearchSweepTargets, typeof m.selectCodexSweepTargets)"` → 기대 `function function function`, exit 0.
- **E4 JSON 파스** — `python3 -c "import json;json.load(open('.claude/commands/harness/release-update/manifest.json'))"` exit 0 (domain 문자열 편집 후).
- **E5 회귀 가드** — §C PRESERVE 앵커 3종 (internal/ 0힛 · last-cc-version.json ≥3 · runner_workflow 참조 1).
- **E6 spec-lint** — `go run ./cmd/moai spec lint SPEC-RELUP-DUALAXIS-001` (또는 프로젝트 규약 형태) exit 0 — MissingExclusions·FrontmatterInvalid 0건 확인.

## §F Milestones (결정 가역성 순 — 변동 가능성 높은 결정부터)

### M1 — 스페셜리스트 codex 축 절차 + 상태 스키마 (데이터 모델 결정 — 최상위)

파일: `.claude/agents/harness/hns-release-update-specialist.md`

1. Phase 0: codex 상태 파일 판독 절차 + 부재 시 기본값 `rust-v0.161.0` + 경고 (REQ-RDX-004). 스키마 블록 문서화 — CC 파일 키 계열 미러 + `rust-v0.161.0` 시드 명기 (REQ-RDX-001/002).
2. Phase 1: codex 수집로 신설 — `gh api repos/openai/codex/releases` (비프리릴리즈 판정) + `npm view @openai/codex version` + 본문 비었을 때 커밋 API 폴백 지시 (러너 §D1 앵커와 정합).
3. Phase 2: codex 티어 분류 — 러너 산출(테마별 관찰목록)을 받아 T1/T2/T3 큐레이션. alpha 테마 watch 규범 (REQ-RDX-009).
4. Phase 7a: Step 7a-codex 신설 — `last-codex-version.json` 병행 기록 (REQ-RDX-003). CC 단독 기록 금지.
5. **§C 시드값 재판정 실행 지점** — M1 착지 직전 npm/gh 재측정.

### M2 — Runner codex 렌즈 (신규 분석 면)

파일: `.claude/workflows/hns-release-update-run.js`

1. `selectCodexSweepTargets(args)` 신설 — CC 셀렉터와 병렬 형태 (REQ-RDX-006). codex 스윕 버전 창은 `args.codexDeltas` 주입 + 스크립트 본문 시드 상수(CC 셀렉터의 `CURRENT_SWEEP_VERSIONS` 패턴 계승 — args 불신뢰 교훈).
2. `CODEX_COMMITS_FALLBACK` 절차 블록 — §D1 (1)-(3) 내용 (REQ-RDX-007).
3. `CODEX_THEME_CHECKLIST` — 6테마 리터럴 + 행 형식 (REQ-RDX-008). 프롬프트 문자열에 체크리스트 주입.
4. top-level 실행부에 codex 렌즈 병렬 fan-out 편입 + 반환 형태에 codex 영향 표 추가. 불변식(§A.5) 유지 확인.

### M3 — 스페셜리스트 BP 상시 섹션 + Phase 3 URL 세트 (절차 영구화)

파일: `.claude/agents/harness/hns-release-update-specialist.md`

1. `Best-Practices` 상시 절차 섹션 신설: 스윕마다 공식 게시 면 스캔(Anthropic engineering/research·docs, OpenAI 블로그), 인벤토리 기록 (REQ-RDX-012).
2. `source-first` 원문 패치 선행 강제 — 검색 요약·2차 자료는 보고 전용 리드로만 (REQ-RDX-013). 2차 BP-1 게시일 오정보 정정 사례를 절차 근거로 인용.
3. `HTML proposal report` 명명 산출물 기록 (REQ-RDX-014 전반).
4. Phase 3 URL 세트 6종을 `code.claude.com/docs/en/*` 캐노니컬 형태로 갱신 (REQ-RDX-014 후반, 결정 D6).

### M4 — 매니페스트 domain 문자열 (기계적 — 최하위)

파일: `.claude/commands/harness/release-update/manifest.json`

1. `domain`만 갱신: `"moai-adk-go dev-only maintainer tooling — Claude Code + Codex CLI upstream change tracking and the best-practices axis"` (§D1 앵커 2종 내포 — AC-RDX-001/002).
2. `source_request`·`sprint_contract`·골격 불변 확인 (AC-RDX-013).

### 마일스톤 순서 근거

M1(상태 스키마·시드 — 데이터 모델, 0.162 승격 시 변동 가능성 최고) → M2(러너 렌즈 — 신규 면 설계) → M3(BP·URL — 절차 문안) → M4(매니페스트 문자열 — 기계적 1줄) 순으로, 인간 검토 집중도가 높은 결정을 앞세운다.

## §G Anti-Patterns

| 금지 | 이유 |
|---|---|
| 러너에 `gh`/`npm` 네트워크 호출 직접 삽입 | 러너는 읽기전용 fan-out 조정자다 — 네트워크 수집은 렌즈 에이전트(Explore)의 Read/WebFetch 몫 |
| codex 렌즈를 스페셜리스트 본문에만 두고 러너를 안 거치게 함 | 기존 CC 축 분업(러너=비대화형 스윕, 스페셜리스트=인간 게이트) 파괴 (결정 D3) |
| `grep -ci codex`를 AC로 쓰는 것 | 주석 mutant(M-2) 통과 — 식별자 앵커가 판정면이다 |
| sprint_contract 차원 추가 | 결정 D4 — 채점 의미론 변경은 요구 밖 |
| Go 라이터 신설 | 결정/REQ-RDX-005 — 하네스 계층 소유가 측정으로 확인된 구조 |
| 템플릿 미러 편집·make build | 미러 자체가 존재하지 않는다 (§1.1 M9) — 존재하지 않는 것을 만들면 스코프 위반 |

## §H Cross-References

- spec.md §1.2 설계 결정 기록 (D1-D7) / §4 REQ-RDX-001..014
- acceptance.md §D AC-RDX-001..013 + §D.3 증거 원장 (2aab5f797)
- `.moai/research/upstream-update-20261007.md` (1차: C1-C5·6테마 표·Phase 7.5 findings) / `upstream-update-20261008.md` (2차: codex 0.161.0 큐레이팅·URL 세트 finding)
- SPEC-UPDATE-ADD-CODEX-001 (codex 배선 선례) · SPEC-CC2219-UPSTREAM-ALIGN-001 (upstream 정렬 선례)
- `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell 규율) · `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 (관측 클레임)
- 카드 t1605 (GD-1 Haiku 5.5 docs 싱크) · CARD-4 (Codex 0.161.0 conformance) — 스코프 인접, 중복 없음
