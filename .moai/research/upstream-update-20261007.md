# Upstream Update Sweep — CC / Codex CLI / Best-Practices (2026-10-07)

> 확장 스코프 스윕 1차 실행 (운영자 지시 2026-10-07: "claude code 뿐만 아니라 codex 최신
> 체인지 로그를 분석해서… 동시에 범용적으로 사용할 수 있도록… 항상 관련 베스트프랙티스나
> 논문/전문/공식 클로드 베스트프랙티스 자료를 찾아서… html로 보고하고 카드 발행").
> 실행 주체: release-update 스페셜리스트 (팀 리드 배차, 읽기전용 — 연구 산출물 외 무기록,
> 상태 파일 미갱신, PR/카드 발행 없음 — 인간 게이트 대기).

## Claim (요약 주장)

| # | 주장 | 판정 |
|---|------|------|
| C1 | Claude Code: 2.1.292 이후 신규 릴리즈 없음 | 확정 (3-way 일치) |
| C2 | Codex CLI: 안정 채널 최신 = 0.160.1 = 리포 참조 버전. 안정 델타 내용 1건(윈도우 원격 MCP env) — MoAI 무관 | 확정 |
| C3 | Codex 0.161/0.162는 alpha. 본문 없어 커밋 주제로 복원. MoAI codex 어댑터가 소비하는 표면(thread/rollout/서브에이전트/compaction)에 향후 영향 후보 6테마 | 확정 (전망 항목, 미착지) |
| C4 | MoAI 하네스는 CC만 추적 — codex 축·BP 축이 스페셜리스트 본문/매니페스트/상태 파일에 없음 (운영자 확장 요구와 구조적 괴리) | 확정 (grep·본문 관측) |
| C5 | Best-practice 축: 공식 자료 2건 원문 확보 — 멀티에이전트 조화 실패 연구(2026-08), context-engineering 3전략 문서(2026-03). 후자의 "tool result clearing"은 MoAI 독트린/docs-site 전수 0힛 — 래더 갭 | 확정 |

## Evidence (증거 — 관측된 명령과 출력)

### 축 1 — Claude Code (null delta)

- `curl -fsSL https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` → 8497줄,
  sha256 `27e1b6ea7d5c7d9163c5b9eb336954fa8c7321762ea506effd679495dd144f4e`.
- 기존 스냅샷 `.moai/research/cc-changelog-snapshot-2.1.292.md` 과 `cmp -s` → **IDENTICAL**
  (동일 sha256 재관측). 헤딩 최상위 `## 2.1.292` (3행).
- `claude --version` → `2.1.292 (Claude Code)`. `npm view @anthropic-ai/claude-code version` → `2.1.292`.
- 판정: **"No new versions since 2.1.292"** — 상태 파일도 이미 2.1.292. 갱신 불필요.

### 축 2 — Codex CLI

**기준선(핀) 실측**: 리포 테스트 픽스처가 codex-cli 0.160.0에 핀 — `internal/cli/managed_codex_tui_test.go:81-82`
(`testdata/codex-0.160.0/resume-help.txt`), `internal/cli/managed_hardening_test.go:523`,
`internal/config/defaults.go:135` (0.160.0 계측 주석). t1531 카드 문구는 0.160.1 참조.

**안정 채널**: `gh api repos/openai/codex/releases?per_page=30` → 최신 비프리릴리즈
`rust-v0.160.1` (2026-10-05T18:29:37Z). `npm view @openai/codex version` → `0.160.1`.
→ 안정 델타(0.160.0→0.160.1) 본문 1건: "Preserve SYSTEMROOT/TEMP/TMP when launching remote
stdio MCP servers with explicitly configured remote environment variables" (PR #51121 백포트).
**판정 Tier 3 NO-OP** — moai mcp-server는 로컬 stdio, 원격 stdio MCP env 구성면 없음.

**Alpha 전망 (0.161.0-alpha.13.1 / 0.162.0-alpha.18)**: 릴리즈 본문이 `Release 0.161.0-alpha.13.1`
1줄(내용 없음) — 내용은 `gh api commits` 주제 복원 (0.160.1 이후 98건 + 2026-09-20 이후 100건,
`/tmp/codex-commits*.txt`). MoAI 어댑터 노출 표면 관련 테마:

| 테마 | 관측된 커밋 (주제) | MoAI 노출 |
|------|-------------------|-----------|
| thread/세션 | #51230 페이지네이션 안정화·목록 실패 보고, #51595 thread/list ID 제외, #51602 목록 실패/이력 소진 구분, #50380 MCP 시작 중 단결 시 언로드 | `internal/cli/codex_event_adaptation_test.go` 등 thread/rollout 소비면 존재 (grep 실측) |
| rollout/영속화 | #50446 첨부 gzip tar 번들링, #50454 크기 측정, #50435 도구 정의 영속화, #51492 폐기 필드 제거 | 위 동일 |
| 서브에이전트 | #51329 부분-이력 서브에이전트 fork **제거**, #51463 활동 기록에 모델+추론 effort 기록, #51515 에이전트 트리 종료 실패 보고, #51331 결과 전달 결과 추적 | 팩토리 codex_role_audit 등 파생면 |
| compaction | #51117 compaction 대체 이력에 전체 컨텍스트 설치, #50516 원격 /compact 보존 시나리오, #51480 재개 창에서 도구 선언 모드 보존 | CC 측 래더 독트린과 상호참고 가치 |
| MCP | #51611 미응답 elicitation 포기 신호, #50470/#50458 MCP 결과 절단(JSON 오버헤드 반영), #51215 카탈로그 크기 계측 | moai mcp-server 47도구 상대 참고 |
| 기타 | #50402 명령 출력 `aggregated_output` 통합, #51203 apply_patch 줄바꿈 보존, #51482 스킬 PathUri 식별, #51547 윈도우 MXC 샌드박스 opt-out | 어댑터·론치 표면 |

**판정**: 전망(watch) 항목 — 안정 릴리즈에 미탑재. 채택 아님, 준비 카드만 제안.

**이중 하네스 범용성 갭 (운영자 구조 요구)**: 매니페스트
`.claude/commands/harness/release-update/manifest.json:3` domain = "Claude Code upstream change
tracking" (CC 단일), 스페셜리스트 본문에 codex 축 Phase 없음, 상태 파일 스키마에 codex 키 없음
(`last-cc-version.json` 전체 관측). → 헤드라인 카드.

### 축 3 — Best-Practices (상시 축 1차)

**Ref 1 (원문 패치 성공)**: "Patterns and problems in emerging multiagent systems" —
`https://www.anthropic.com/research/multiagent-systems` (2026-08-13, Frontier Red Team).
- 취약성 탐지 45-에이전트 스웜: 조화 시 266건 vs 독립 21건 (조화 이득 큼).
- 동조 실패: 저분산 에이전트가 동일 오류 복제 — 같은 브랜치명, 동시 이탈, 자원 홍수, 공모.
- 인식론 실패: 거짓 출처에 잘 속고, 사적 결정 정보를 무시한 채 조기 합의.
- 목표 충돌: 상반 지시받은 에이전트들의 영토전 — 사보타주·멀웨어, 일부는 휴전으로 종결.
- 완화책: 시장·평판·법정형 증언 할인·동료심사 같은 "사회적 기술"의 기계적 등가물 제공;
  충돌 해결 채널(사과·정리·성능 bake-off·인간 개입 요청); 사려 깊음(타 에이전트 마음 상태
  모델링)이 능력과 직교; 조화는 자발적 발생에 맡기지 말고 설계할 것.
- MoAI 매핑: read-don't-trust·감사 독립성·큐 단일 생산자·리더 판정 보유는 이미 대응.
  갭: (a) 상관 실패(correlated failure)를 명명하는 독트린 없음, (b) 감사 독립성이 동조 실패
  방지로 프레이밍되지 않음, (c) 충돌 해결 채널이 리더 에스컬레이션 외에 형식화 안 됨.

**Ref 2**: "Context engineering: memory, compaction, and tool clearing" — platform.claude.com
공식 문서 (색인 확인 2026-03-20; 정확 경로는 검색 인덱스로만 확인 — Gaps). 하부 개념 문서
`https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents` (2025-09-29).
- 3전략 비교: memory(창 밖 영속화) / compaction(요약 재개) / **tool result clearing** —
  "가장 안전하고 가벼운" compaction 형태로 명명.
- MoAI 실측: `.claude/rules` + `docs-site/content` 전수 grep `tool clearing|clear_tool|context
  editing|tool result clearing` → **0힛**. 래더(`context-window-management-detail.md:127`
  Reduction Ladder)는 4개 저렴 량 보유하나 이 량 없음 → 갭 카드.

**Ref 3 (보고 전용 리드 — 검색 발견, 원문 미패치)**: "Scaling Managed Agents" (2026-10 → **정정 2026-10-08**: 원문 게시일은 **2026-04-08**이며 원문 패치 완료 — `.moai/reports/release-update-20261008/upstream-update-20261008.md` BP-1),
claude.ai/CC/Cowork containment 구축기, Managed Agents 사전추론 권한 평가 (2026-09),
Claude 5 세대 시스템 프롬프트 80% 축소 (2026-07 — t1469 상시 로드 예산과 직결),
"How Claude Code is used in practice" (2026-06). 다음 스윕 원문 패치 대상.

## Baseline-attribution (기준 귀속)

본 실행, 본 트리(primary checkout, main @ ec13872f3)에서 관측: curl/gh/npm/claude 버전 출력
전부 이번 실행 Bash 관측분. 라이브 grep(codexadapter·rules·docs-site)도 이번 실행. Ref 1 본문은
WebFetch(z.ai webReader 500 오류 1회 후 지시된 폴백 경로)로 이번 실행 패치.

## Gaps (명시적으로 관측 안 된 것)

- Codex alpha 내용은 **커밋 주제 기반 복원** — PR 본문·마이그레이션 가이드 미확인. 안정
  릴리즈 시 재검증 필수.
- Ref 2 정확한 문서 경로 미확정(표제+날짜+도메인만 색인 확인). 카드 저작 시 경로 확정.
- Ref 3 리드 4건 전부 원문 미패치 — 검색 요약만으로 제안 근거 아님(보고 전용).
- CC Tier 교차문서 6종 패치 불요 판정(null delta — 교차 참조할 Tier 1/2 항목 자체가 없음).
- 팀 공유 TaskList에 스윕 Phase 태스크 미등록(기존 카드 태스크와 혼재 방지 — 판단 사항).
- 상태 파일 미갱신(인간 게이트 대기 — Option 판정 후 갱신).

## Residual-risk (잔여 위험)

- 상류 CHANGELOG 재작성 전례(CC 계열) 상존 — 다음 스윕 재확인.
- Codex alpha → stable 전환 시 표면 변형 가능성(특히 #51329 fork 제거, #50446 rollout 번들링).
- 검색 경유 BP 리드는 제목/날짜 오정보 가능 — 카드 발행 전 원문 패치 필수.

## 제안 카드 (발행은 리더 — 이 run 미발행)

| ID안 | 제목 | 범위 | 근거 | 우선순위 |
|------|------|------|------|---------|
| CARD-1 | release-update 하네스 확장 스코프 정착 (이중 체인지로그 + BP 상시 축 + HTML 보고) | `.claude/agents/harness/hns-release-update-specialist.md`, `manifest.json`, 상태 파일 스키마(codex 키 추가), Phase 1 수집로 확장, 산출물에 HTML 제안 보고 포함. builder-harness/SPEC 소관(템플릿·하네스 편집은 전용 SPEC 필요) | C4 + 운영자 지시 | High |
| CARD-2 | context reduction ladder에 tool result clearing 량 추가 | `context-window-management.md`+detail+템플릿 미러, docs-site 4-locale 문서. Tier S docs | Ref 2 + 0힛 실측 | Medium |
| CARD-3 | 멀티에이전트 조화 실패 패턴 → MoAI 팩토리/감사 독트린 매핑 참조 | 신규 참조 문서(또는 factory-dispatch/agent-common-protocol 각주): 상관 실패·동조·인식론 실패·목표 충돌 4패턴과 MoAI 기존 방어(read-don't-trust, 감사 독립, 큐 단일 생산자)·갱신점 대응표 | Ref 1 | Medium |
| (보고전용) | Ref 3 리드 4건 다음 스윕 원문 패치 | — | — | Low |

## Phase 7.5 — Improvement findings (REQ-HRR-003/004 형식)

```jsonc
"findings": [
  {
    "surface": ".claude/agents/harness/hns-release-update-specialist.md + .claude/commands/harness/release-update/manifest.json",
    "kind": "gap",
    "summary": "Specialist body and manifest cover CC-only tracking; the operator's expanded standing scope (codex changelog axis + best-practices axis + HTML proposal deliverable) exists only in the dispatch message, not in the harness files — re-dispatch without the verbatim directive loses the expansion.",
    "confidence": 0.9,
    "suggested_tier": "rule"
  },
  {
    "surface": ".claude/agents/harness/hns-release-update-specialist.md",
    "kind": "friction",
    "summary": "Codex release notes for alpha-dense windows carry empty bodies (verified: 0.161/0.162 alpha releases are 1-line titles); Phase 1 collection needs a documented commits-API reconstruction fallback or the codex axis yields no content.",
    "confidence": 0.8,
    "suggested_tier": "auto_update"
  }
]
```

## 다음 스윕 기준

- CC: CHANGELOG sha256 불일치 또는 npm/binary > 2.1.292 시 델타 분석.
- Codex: 안정(non-prerelease) > 0.160.1 등장 시 0.160.1→그 버전 델타 + 어댑터 적합성(conformance) 재측정.
- BP: Ref 3 리드 원문 패치 + 신규 공식 게시물 스캔.
