---
id: SPEC-RELUP-DUALAXIS-001
title: "release-update 하네스 CC+Codex 이중 축 정착 — codex 체인지로그 축·상태 파일 codex 키·BP 상시 절차"
version: "0.3.0"
status: draft
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/harness/hns-release-update-specialist.md, .claude/workflows/hns-release-update-run.js, .claude/commands/harness/release-update/manifest.json, .moai/state (harness-layer writer, machine-local)"
lifecycle: spec-anchored
tags: "release-update, codex, dual-axis, harness, best-practices, state-schema, dev-only"
tier: M
era: V3R6
related_specs: [SPEC-UPDATE-ADD-CODEX-001, SPEC-CC2219-UPSTREAM-ALIGN-001]
---

# SPEC: release-update 하네스 CC+Codex 이중 축 정착 — codex 체인지로그 축·상태 파일 codex 키·BP 상시 절차

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-09 | manager-spec | 최초 작성 — 카드 t1579 (High·운영자 확장 지시 2026-10-07·builder-harness/SPEC 소관). 워크트리 t1579 @ `2aab5f797` 실측 13종(앵커 grep + exit code 전수 관측)을 근거로 REQ 14건·AC 13건(릴리스 블로킹 10 + 회귀 가드 3) 확정. 근거 연구: `.moai/research/upstream-update-20261007.md` (확장 스윕 1차) + `upstream-update-20261008.md` (2차). SPEC ID 사전 검증: `SPEC-RELUP-DUALAXIS-001` 정규식 `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` 매치 **PASS** (Bash 실행 관측) |
| 0.1.1 | 2026-10-09 | manager-spec | 자체 spec-lint 피드백 수리 — `CoverageIncomplete` 3건(REQ-RDX-004/009/013 미커버)에 AC-RDX-014/015/016 신설(블로킹 10→13, 총 AC 13→16 — Tier M 천장 16에 정확히 도달). AC-RDX-016의 `source-first` 앵커로 mutant M-4의 기계 판정면 확보(LED-014, 0/exit 1 실측). lint 재실행: 0 error / 0 warning |
| 0.2.0 | 2026-10-09 | manager-spec | plan-audit iter1 수리(FAIL 0.8125 — codex required 게이트, `.moai/reports/t1579/plan-audit.md`, rcpt-3e881c87c91a47a2117220b0). **CX-4(블로커)**: plan §E3을 런타임-형태 어댑터로 교체 — require·직접 실행 모두 `SyntaxError: Illegal return statement`(run.js:153, Node v22.14.0 본 재관측 exit 1), 러너는 ESM `export const meta`+top-level return/await 하이브리드라 export 적출+AsyncFunction 래핑 어댑터가 유일한 충실 평가형 — E3-P1 본 트리 **exit 0 관측**(`adapter-ok run=fn cc=1 shape-ok`), E3-P2(M2 종료형)는 codex 단언 포함. **CX-1**: plan §C 시드 재판정을 last-analyzed 의미론으로 재작성(D1·§7 전파 — 미분석 승격은 시드 미인상). **CX-3**: LED-001/002를 `domain` 필드 스코프로 재앵커 + AC-RDX-013에 sprint_contract 판독면 신설(LED-015 기준선 출력 기록, exit 0 — AC 수 16 불변). **CX-2**: AC-RDX-003을 `selectCodexSweepTargets(args)` ≥2(정의+top-level 디스패치 호출)로 재앵커 + §E3-P2 실질 생성 면(M2에 export 목록 확장 요건 추가). **D5**: §3 한계를 AC-RDX-014/015로 확장(iter2 블록 재판독이 제2 판정면). **D6**: frontmatter version 0.2.0 — 최신 HISTORY 행 정합 유지(0.1.1 지적의 재발 방지). **D7 처분**: progress §F 선기입은 라인 지시에 의한 것 — 오케스트레이터가 Phase 4에서 확정/수정(내용 수정 불요). **D8 처분**: 운영자 ①구절의 리포 참조 버전 비교 반쪽은 CARD-4 소관 기록 유지 — run-phase 위임 프롬프트가 M1/M2에 참조-버전 문맥을 운반할 것. 미 touched AC의 RED 셀은 바이트 불변 유지 |
| 0.3.0 | 2026-10-09 | manager-spec | plan-audit iter2 수리(FAIL 0.875 — Testability 0.75, iter1 4건 수리는 전수 재검증 통과). **CX-5**: AC-RDX-003에 병합 관측면 신설 — plan §E3-P3 모의-런타임 실행 동사(mock agent/parallel로 러너 top-level 블록 실행, 시드 codexDeltas 주입)가 병합·agent 호출까지 실측; LED-016 RED 본 트리 관측(stderr `REJECTED: no-codex-dispatch:1`, exit 1 — 현재 러너는 CC 호출 1건에 codex 라벨 0건), M2 GREEN 기대 `dispatch-ok codex=1 total=2`/exit 0. §D1에 병합 변수 `allTargets`+`parallel(allTargets` 호출식+codex 라벨 접두사 `codex-release-notes:` 핀, §3 M-6 신설. **CX-6**: AC-RDX-006을 이중 사이트 기준으로 확장 — LED-006 GREEN 문턱 ≥1→≥2(Phase 0 판독·기본값 + Phase 7a 기록) + LED-017 기록 단계 리터럴 `7a-codex` 신설(RED 0/exit 1 본 트리 관측), §3 M-7 신설 — REQ-RDX-003의 기계 면 확보, AC 수 16 불변. **CN-4 처분**: 본 SPEC 산출물은 awk 계열 정렬-검증 동사를 인용하지 않는다 — 전부 단일 grep·node -e·python3 -c(가드 통과형). 미 touched AC의 RED 셀 바이트 불변 |

## 1. 문제 — 측정된 형태

release-update 하네스는 Claude Code 단일 축으로 태어났다. 운영자 확장 지시(2026-10-07 — "claude code 뿐만 아니라 codex 최신 체인지 로그를 분석해서… 동시에 범용적으로 사용할 수 있도록… 항상 관련 베스트프랙티스나 논문/전문/공식 클로드 베스트프랙티스 자료를 찾아서 html로 보고하고 카드 발행")는 1차·2차 확장 스윕을 **배차 메시지의 축어 지시문**으로 구동했을 뿐, 그 확장은 하네스 파일 어디에도 영구화돼 있지 않다. 1차 연구의 Phase 7.5 finding (confidence 0.9)이 정확히 이 결함을 명명한다: *"re-dispatch without the verbatim directive loses the expansion"*. 2차 스윕조차 재배차 메시지로 구동됐고, 그 Gaps 절은 "t1579(CARD-1) 착지 전까지 확장 스코프는 배차 메시지 의존"을 명시 기록했다.

**측정 트리**: `.moai/worktrees/t1579` (branch `WT-high-10-07`) @ `2aab5f797`. 이 SPEC의 모든 수치와 grep 관측은 이 트리의 실측이다.

```console
$ grep -c "Codex" .claude/commands/harness/release-update/manifest.json
0        # exit=1 — 매니페스트에 codex 축 표면 없음
$ grep -ci "codex" .claude/workflows/hns-release-update-run.js
0        # exit=1 — Runner fan-out에 codex 렌즈 없음
$ grep -c "last-codex-version" .claude/agents/harness/hns-release-update-specialist.md
0        # exit=1 — 본문 전체에 codex 상태 스키마 없음
$ grep -rn "last-cc-version" internal/
(출력 없음)   # exit=1 — 상태 파일의 Go 라이터 부재 — 하네스 계층 전용임을 확인
```

### 1.1 실측 요약 (트리 `2aab5f797`)

| # | 측정 대상 | 관측값 | 근거 |
|---|---|---|---|
| M1 | 매니페스트 domain | `"moai-adk-go dev-only maintainer tooling — Claude Code upstream change tracking"` — CC 단일 | manifest.json:3 |
| M2 | Runner codex 표면 | `grep -ci codex` = **0** (exit 1). fan-out은 CC `versionDeltas` 전용 (`selectResearchSweepTargets` 4힛 — CC 렌즈 셀렉터만 존재) | hns-release-update-run.js |
| M3 | 상태 파일 codex 키 | 스키마 키 없음. 운영 증거: 2차 스윕의 codex 결과가 CC 파일 `genuine_delta` **자유 서술**에 편입됨 — *"Codex 축(확장 스코프 2차): rust-v0.161.0 안정 승격 2026-10-07T15:58:45Z — 큐레이팅 14불릿 T1=4/T2=4/T3=6"* — 축은 존재하는데 스키마 홈이 없다 | `last-cc-version.json` (primary 체크아웃) 2026-10-08 항목 |
| M4 | Go 라이터 | `grep -rn "last-cc-version" internal/` = 0힛 — 상태 파일은 하네스 계층(specialist/runner)이 쓴다. codex 키 신설도 Go 변경을 수반하지 않는다 | 본 트리 실측 |
| M5 | BP 축 | 스페셜리스트 본문 `grep -ci best-practice` = **0**. 1·2차 스윕의 BP 축 산출(원문 패치 4건: multiagent-harmony·context-engineering·managed-agents·Opus 5.5 프롬프팅 가이드)은 절차 홈 없음 | specialist.md 본문 + 연구 2건 |
| M6 | codex 기준선 (2차 스윕 실측) | npm `@openai/codex` = 0.161.0 = 안정 (`rust-v0.161.0` 승격 2026-10-07T15:58:45Z), 설치 바이너리 0.160.1, 리포 테스트 핀 0.160.0. 0.160.1→0.161.0 델타는 **이미 큐레이팅 완료** (14불릿 T1=4/T2=4/T3=6) | upstream-update-20261008.md 축2 |
| M7 | alpha 관찰목록 | 0.161/0.162 alpha 릴리즈 본문은 1줄 제목(내용 없음) — 커밋 주제 복원으로 실측 (0.160.1 이후 98건 + 100건). 어댑터 노출 6테마(thread/rollout/서브에이전트/compaction/MCP/기타) 관측 완료, 안정 미탑재 watch 항목 | upstream-update-20261007.md 축2 표 |
| M8 | Phase 3 URL 세트 드리프트 | 스페셜리스트 Phase 3 URL 6종이 `docs.anthropic.com/en/docs/claude-code/*` — 실제 페치는 전부 `code.claude.com/docs/en/*`로 캐노니컬라이즈 (2회 연속 관측). 본문 `code.claude.com` = 0힛 | upstream-update-20261008.md finding (confidence 0.75) |
| M9 | 템픔릿 비대상 | 3개 표면 전부 사용자 소유 네임스페이스 (`.claude/agents/harness/`, `.claude/workflows/hns-*`, `.claude/commands/harness/`) — `moai update` 비대상, `internal/template/templates/` 미러 불요, `make build` 불요 | agent-authoring.md § Agent Directory Convention + dynamic-workflows.md hns-* 조항 |

### 1.2 설계 결정 기록 (재논의 금지 — plan.md §D와 쌍을 이룬다)

| # | 결정 | 근거 |
|---|---|---|
| D1 | codex 축 상태 파일은 **별도** `.moai/state/last-codex-version.json`으로 신설하고 CC 파일과 동일 키 계열(`last_analyzed_version` / `last_analyzed_date` / `last_master_research` / `analysis_history[]`)을 미러링한다. 시드 `last_analyzed_version` = **`rust-v0.161.0`** — 원본 릴리즈 태그 형태 그대로 기록한다(CC 파일의 plain semver "2.1.294"와 표기 형식이 달라도 태그 형태가 스윕 비교 기준이므로). 카드 후보안의 0.160.1 시드는 **폐기** | 카드 지시 "별도" 그대로. 2차 스윕이 0.160.1→0.161.0 안정 델타를 이미 분석·큐레이팅했으므로(§1.1 M6) 기준선은 0.161.0이다 — 0.160.1 시드는 1차 스윕 시점(2026-10-07) 기준이며 2차 실측(2026-10-08)이 우선한다. 실제 파일 생성은 다음 스윕 실행 시점의 하네스 절차가 수행한다(기계 로컬 — 본 SPEC은 본문의 스키마 문서화만 소유). **시드 의미론은 last-analyzed다** — 분석되지 않은 신규 안정 승격(예: 0.162 선행 승격)이 관측돼도 시드는 고정되고 그 델타는 다음 스윕의 분석 대상으로 기록된다(plan §C 재판정 규칙 — plan-audit iter1 CX-1) |
| D2 | Runner는 CC 렌즈와 병렬로 **codex 렌즈**(fan-out)를 얻는다. codex 렌즈는 릴리즈 본문이 비어 있을 때(밀도 높은 alpha 기간) 커밋 API 복원 폴백을 **요구 절차**로 문서화하고, 복원 항목은 전부 커밋-주제-유래로 라벨링한다. 관찰 목록 형식: 테마 행 = 테마 키 + 관측 PR 번호 목록 + MoAI 노출면 | 1차 연구 Phase 7.5 finding (b, confidence 0.8): *"Phase 1 needs a documented commits-API reconstruction fallback or the codex axis yields no content"*. 2차 #49713이 보인 정합 절차 — 커밋 제목만으로 판정하지 않고 PR 본문 확인으로 격상 — 를 렌즈 절차에 흡수 |
| D3 | 6테마 어댑터-노출 관찰목록은 **Runner의 codex 렌즈 프롬프트**에 상주한다(영어 키: `thread` / `rollout` / `subagent` / `compaction` / `MCP` / `other`). 스페셜리스트는 러너 산출을 받아 큐레이션·티어 분류·안정 승격 판정을 수행한다 — alpha 테마는 watch 관찰목록으로만 기록되고, 안정 릴리즈 탑재 시에만 채택 판정한다 | 기존 CC 축의 러너(비대화형 스윕)/스페셜리스트(인간 게이트) 분업 계승(§1.1 M2, specialist "Runner integration" 절). 영어 키는 coding-standards.md 에이전트 정의 영어 규정. 1차 스윕의 watch 판정("채택 아님, 준비 카드만 제안")이 절차 규범이 된다 |
| D4 | 매니페스트 `sprint_contract.dimensions`는 **Functionality/Consistency 2개 유지, thresholds(0.85/0.80) 불변**. `domain` 문자열만 이중 축 + BP 축을 명명한다 | 차원 추가는 이 하네스 향후 모든 run의 sync 채점 의미론을 바꾸는 정책 변경이다 — 운영자 지시(도메인 확장)는 그것을 요구하지 않는다. Enforce Simplicity 사다리 1단(YAGNI) |
| D5 | BP 축은 스페셜리스트 본문에 **상시 절차 섹션**으로 영구화한다: 스윕마다 공식(Anthropic/OpenAI) 게시 면을 스캔하고, BP 항목이 제안(카드 발행·문서 싱크 권고)의 근거가 되려면 **원문 패치(verbatim fetch) 선행**을 요구한다 — 검색 요약·2차 자료는 보고 전용 리드. HTML 제안 보고를 명명 산출물로 기록한다. 실행은 본 SPEC 소관이 아니다(절차 영구화만) | 운영자 "항상" 지시. 1차 Residual-risk "검색 경유 BP 리드는 제목/날짜 오정보 가능"이 2차 BP-1에서 실현됐다(게시일 2026-10 → 원문 패치로 2026-04-08 정정) — 원문 선행 강제가 그 재발 방지다 |
| D6 | 스페셜리스트 Phase 3 문서 URL 세트를 `code.claude.com/docs/en/*` 캐노니컬 형태로 갱신한다 | 2차 스윕 finding (confidence 0.75, *"t1579 owns the body — recorded as input to that card"*): 2회 연속 페치 전부 캐노니컬라이즈 관측. 본문을 고치는 SPEC이 본 SPEC이므로 함께 정착한다 |
| D7 | 본 SPEC은 절차의 **영구화**만 소유한다 — 스윕 실행 자체는 소관 밖(§6) | 카드 본문 "하네스 절차화". 1·2차 스윕은 이미 배차 메시지로 실행 완료 |

## 2. 원인 — 확장이 파일이 아니라 배차문에 살았다

하네스는 SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001에서 CC 전용 능력으로 포팅됐고(매니페스트·러너·스페셜리스트·상태 파일 4면 모두 CC 전용으로 설계), 이후 운영자가 스코프를 확장할 때마다 새 카드가 아니라 **배차 메시지에 확장 지시문을 얹는** 방식으로 운영됐다. 배차문은 세션 생존 자료다 — 세션이 끝나면 소실되고, 재배차 시 지시문을 다시 붙이지 않으면 확장은 잃어버린다. 1차 스윕(2026-10-07)이 발견한 것은 기능 결함이 아니라 **영속성 결함**이다: 축은 운영 중인데(C3 판정 "확정"), 그 축을 담는 스키마·렌즈·절차가 하네스 파일에 없다(C4 판정 "구조적 괴리"). 2차 스윕이 같은 결함을 재확인했고, 상태 파일에서는 codex 결과가 CC 모양의 자유 서술 필드에 짓눌려 저장되고 있다(§1.1 M3) — 스키마 홈 부재가 이미 관측 비용을 내고 있다.

## 3. 이 카드의 대표 mutant

아래 mutant들이 이 SPEC의 AC를 통과하려면 AC가 너무 얕은 것이다(verification-completeness §2 mutant probe).

- **M-1 "출처 필드 기만"**: `source_request`(역사 서술 필드)에만 codex를 언급하고 `domain` 문자열은 CC-only로 남기는 mutant. AC-RDX-001이 `domain` 키 행 스코프 패턴(`'"domain".*Codex CLI upstream change tracking'`)을 grep하므로 잡힌다 — source_request에 동일 문구를 넣어도 매치되지 않는다(plan-audit iter1 CX-3 재앵커).
- **M-2 "주석 코덱스"**: JS 주석에만 `// TODO codex lens`를 추가하는 mutant. AC-RDX-003이 `selectCodexSweepTargets(args)` 출현 ≥2(정의+top-level 디스패치 병합 호출)와 plan §E3-P2 어댑터의 실측 target 생성을 요구하므로 잡힌다 — 주석·미연결 정의는 두 면 모두에서 좌초한다(plan-audit iter1 CX-2 재앵커).
- **M-3 "seed 누락"**: `last-codex-version.json` 스키마는 문서화하되 시드값을 빼는 mutant. AC-RDX-006은 통과하고 AC-RDX-007(`rust-v0.161.0`)에서 잡힌다 — 둘이 쌍인 이유다.
- **M-4 "BP 껍데기 섹션"**: `Best-Practices` 헤딩만 넣고 원문-패치 강제를 빼는 mutant. AC-RDX-008(섹션 존재)은 통과할 수 있다 — REQ-RDX-012(shall not)와 plan §D 앵커(`source-first` 리터럴)가 잡는다. §3에 한계를 명시한다: 이 mutant는 grep 단일 판정면 밖이며 plan-auditor 서술 검증이 보완 판정면이다. 동일 한계는 AC-RDX-014/015에도 적용된다 — 기계 면은 LED-006/005 공유 grep이고, 서술 면(부재-기본값 경고·watch 규범의 실제 기재)은 plan-audit iter2의 해당 블록 재판독이 제2 판정면이다(plan-audit iter1 D5).
- **M-5 "Go 침입"**: 상태 파일 쓰기를 Go 런타임(`internal/`)으로 옮기는 mutant. AC-RDX-011(회귀 가드 — `internal/` grep 0힛 유지)에서 잡힌다. 상태 파일은 하네스 계층 소유가 측정으로 확인된 구조적 사실이다(§1.1 M4).
- **M-6 "병합 제외"**: `selectCodexSweepTargets`를 정의·export·직접 호출하되 그 target을 `parallel(...)` 병합에서 제외하는 mutant. LED-003(≥2)은 통과할 수 있으나 LED-016(plan §E3-P3 모의-런타임 관측 — codex 라벨 agent 호출 0건)에서 좌초한다(plan-audit iter2 CX-5).
- **M-7 "단일 사이트 기록"**: codex 상태 문서를 Phase 0에만 두고 Phase 7a 기록 단계를 CC-only로 남기는 mutant. LED-006 단독(≥1)은 통과하나 LED-006 ≥2 + LED-017(`7a-codex`)에서 좌초한다(plan-audit iter2 CX-6).

## 4. 요구사항 (GEARS)

### codex 축 — 상태 스키마 (M1)

- **REQ-RDX-001** (Ubiquitous) — The release-update harness shall maintain a dedicated codex-axis state file at `.moai/state/last-codex-version.json`, separate from `.moai/state/last-cc-version.json`, mirroring the CC file's key family (`last_analyzed_version` / `last_analyzed_date` / `last_master_research` / `analysis_history[]`) (결정 D1).
- **REQ-RDX-002** (Ubiquitous) — The specialist body shall document the `.moai/state/last-codex-version.json` schema in Phase 0 (read) and Phase 7a (write) with `last_analyzed_version` seeded at `rust-v0.161.0` — the stable promotion already analyzed by the 2026-10-08 sweep (결정 D1).
- **REQ-RDX-003** (Event-driven) — **When** Phase 7a persists state after a sweep, the specialist shall write BOTH `last-cc-version.json` (CC axis) and `last-codex-version.json` (codex axis) — a CC-only write leaving the codex baseline stale is prohibited.
- **REQ-RDX-004** (Event-driven) — **When** the codex state file is missing at Phase 0, the specialist shall default the codex since-baseline to `rust-v0.161.0` and emit a warning, mirroring the CC file's missing-file behavior.
- **REQ-RDX-005** (Unwanted) — The codex-axis state write shall not introduce a Go-side writer: `internal/` shall keep zero references to `last-codex-version` (the state file is harness-layer-owned, §1.1 M4 — mutant M-5 봉쇄).

### codex 축 — Runner 렌즈 (M2)

- **REQ-RDX-006** (Ubiquitous) — The Runner shall fan out a codex lens alongside the CC lens, implemented as a parallel selector function `selectCodexSweepTargets` emitting one read-only analysis target per codex release-window delta (stable channel, GitHub releases API snapshot) (결정 D2, plan §D 앵커).
- **REQ-RDX-007** (Event-driven) — **When** a codex release body is empty (alpha-dense windows carry 1-line titles), the codex lens shall reconstruct content from the commits API per the documented fallback procedure (`CODEX_COMMITS_FALLBACK`), and shall label every reconstructed item as commit-topic-derived — never as release-note text.
- **REQ-RDX-008** (Ubiquitous) — The codex lens shall classify observations against the standing 6-theme adapter-exposure checklist (`CODEX_THEME_CHECKLIST`: `thread` / `rollout` / `subagent` / `compaction` / `MCP` / `other`), each theme row carrying observed PR numbers and the MoAI exposure surface (결정 D3).
- **REQ-RDX-009** (Unwanted) — An alpha-window theme entry shall not be reported as adopted drift: alpha themes remain watch-list observations, and adoption judgment happens only when the theme lands in a stable release (1차 스윕 watch 판정의 절차화).

### 매니페스트 (M4)

- **REQ-RDX-010** (Ubiquitous) — The manifest `domain` string shall name both axes (Claude Code + Codex CLI upstream change tracking) and the best-practices axis (plan §D 고정 리터럴).
- **REQ-RDX-011** (Ubiquitous) — The manifest `sprint_contract.dimensions` shall remain `Functionality` + `Consistency` with unchanged thresholds — the dual-axis expansion is scored within the existing dimensions (결정 D4).

### BP 축 + 본문 정착 (M3)

- **REQ-RDX-012** (Ubiquitous) — The specialist body shall carry a standing best-practices procedure section: per-sweep scan of official Anthropic/OpenAI publishing surfaces, with the BP item inventory recorded per sweep (결정 D5).
- **REQ-RDX-013** (Unwanted) — A BP item shall not back a proposal (card issuance, docs-sync recommendation) unless its source article has been fetched verbatim (`source-first` 원문 패치 선행) — search-result summaries and secondary sources are report-only leads (결정 D5, mutant M-4 봉쇄).
- **REQ-RDX-014** (Ubiquitous) — The BP axis shall name the HTML proposal report as a deliverable, and the specialist Phase 3 doc-fetch URL set shall list the `code.claude.com/docs/en/*` canonical URLs (결정 D5 + D6).

## 5. 알려진 구속 조건

1. **3개 표면 전부 dev-only 사용자 소유 네임스페이스다.** `.claude/agents/harness/`·`.claude/workflows/hns-*`·`.claude/commands/harness/`는 `moai update` 비대상이고 `internal/template/templates/` 미러가 없다(§1.1 M9) — 템픔릿 미러 작업과 `make build`는 이 카드에 존재하지 않는다. run-phase가 실수로 템플릿을 건드리면 그것은 scope 위반이다.
2. **상태 파일은 기계 로컬 gitignored다.** `last-codex-version.json`의 실제 생성은 CI/테스트가 판정할 수 없으므로, 본 SPEC의 AC는 스페셜리스트 본문의 쓰기 지점(스키마 문서화)을 측정면으로 삼는다. 2차 스윕이 이미 보여줬듯 스키마 홈 부재는 관측 비용을 내고 있다 — 본문 문서화가 스키마의 규범면이다.
3. **Runner 불변식은 유지된다.** AskUserQuestion·gh pr 호출 금지(HARD, AC-DHC-007a), `Date.now()`/`Math.random()` 금지(결정성), top-level 실행 + CommonJS export 가드 패턴 유지. codex 렌즈는 이 불변식 위에 병렬 구조로 얹힌다.
4. **에이전트·러너 본문은 영어다**(coding-standards.md Language Policy). 6테마 키와 절차 서술은 영어로 기록하고, 한국어 테마명(서브에이전트/기타)은 대응표로만 남긴다.
5. **Tier M 예산** — REQ 14건 / AC 16건 (천장 16/16 — AC가 천장에 정확히 도달한다. 추가 AC는 티어 상향 또는 SPEC 분할 신호다).

## 6. 범위 밖 (Non-goals)

### Out of Scope — Go 측 codex fixture 재핀

- `internal/cli/managed_codex_tui_test.go`의 `testdata/codex-0.160.0` 핀 갱신, `internal/config/defaults.go:135` 계측 주석, auth.json/키링 실측, `resume --help` 대조 — 어댑터 conformance 재측정은 별도 카드(CARD-4, 2026-10-08 리더 발행)의 소관이다. 본 SPEC은 본문 절차만 소유하고 Go 코드를 1줄도 만지지 않는다.

### Out of Scope — 스윕 실행 자체

- 1차(2026-10-07)·2차(2026-10-08) 스윕은 이미 배차 메시지로 실행 완료됐다. 본 SPEC 착지 후의 실제 스윕 실행(codex 상태 파일 최초 생성 포함)은 다음 `/harness:release-update` 배차의 일이다(결정 D7).

### Out of Scope — CARD-2·CARD-3 (연구 제안 표의 형제 카드)

- CARD-2: context reduction ladder에 tool result clearing 량 추가(`context-window-management.md`+detail+미러+docs-site 4-locale). CARD-3: 멀티에이전트 조화 실패 패턴의 MoAI 독트린 매핑 참조 문서(+BP-1·BP-2 매핑 합류). 둘 다 1차 연구 제안 표의 독립 카드다.

### Out of Scope — docs-site 4-locale 동기화

- release-update 하네스는 dev-only 메인테이너 인프라로 사용자 프로젝트에 배포되지 않는다. 스페셜리스트가 수행하는 문서 싱크 절차(Phase 6)는 변경하지 않는다.

### Out of Scope — 템픔릿 미러링

- `internal/template/templates/` 측 미러 생성·동기화. 3개 표면이 전부 사용자 소유 네임스페이스라 미러가 존재하지 않는다(§1.1 M9, §5.1).

### Out of Scope — BP 매핑 참조 문서 작성

- BP 항목의 MoAI 독트린 매핑(부패성 휴리스틱 클래스 명명, elapsed-time budget 실험 등)은 CARD-3의 확장 범위다. 본 SPEC은 "원문 패치 → 제안 근거"의 절차 강제만 만든다.

## 7. 미검증 항목 (Gaps)

- **시드값의 유효기간** — `rust-v0.161.0`은 2026-10-08 기준 안정 최신이며 시드 의미론은 **last-analyzed**다. run-phase 착지 전에 0.162 승격이 관측돼도 시드는 올리지 않는다 — 미분석 델타는 다음 스윕의 분석 대상으로 기록된다(plan §C 재판정 규칙 — plan-audit iter1 CX-1).
- **6테마 관찰목록의 후속 변동** — 0.162 승격 시 테마 추가/삭제가 예상된다. 본 SPEC은 6테마를 시드로 고정하고(plan §D), 확장/축소는 스윕 재량으로 기록된다.
- **BP 공식 자료 URL 목록의 완결성** — BP-3(context engineering 문서)의 정확 경로는 2회 스윕에서도 미확정(검색 색인만 확인). BP 절차는 "원문 패치 시점에 경로 확정"을 요구하고 본 SPEC은 URL 인벤토리를 완결하지 않는다.
- **러너 codex 렌즈의 실제 fan-out 동작** — 본 SPEC 착지는 절차 편집이지 실행이 아니다. 첫 codex 렌즈 실행은 다음 스윕에서 관측되며, 그때까지 렌즈 프롬프트의 커밋 복원 실효성은 미검증 상태로 남는다(§6 스윕 실행 제외와 동일 뿌리).
