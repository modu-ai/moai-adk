---
id: SPEC-RELUP-STRUCT-AUDIT-001
title: "plan — release-update 구조 검사 설계 (마크다운 구조 파서 6면 + M1–M10 음성 대조)"
created: 2026-10-13
author: manager-spec
tier: M
---

# plan: SPEC-RELUP-STRUCT-AUDIT-001

## §A 접근 (Approach)

**한 문장**: `internal/harness/relup_structure_audit_test.go` (신설, 단일 테스트 파일, package `harness`)에 공용 마크다운 구조 헬퍼(스코프 파라미터 마스킹/제거 + 줄시작 헤딩 경계 + 이중 클래스 중복 처리 계약)와 여섯 면의 검사 함수를 두고, M1–M10 우회 probe를 인메모리 문자열 변이로 음성 대조하며, 실트리 양성 대조로 착지를 판정한다 — 게이팅은 표준 `go test ./internal/harness/` 스위트(CI 배치 릴리스 경로)다. (spec.md §1.2 결정 D1–D8)

**변경 파일 (전체 허용 경로)**:

| 경로 | 신설/수정 | 내용 |
|---|---|---|
| `internal/harness/relup_structure_audit_test.go` | 신설 | 공용 헬퍼(비공개, test 전용) + 5개 테스트 함수 + M1–M10 변이 단위 테스트 |
| `.moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/{spec,plan,acceptance}.md` | plan-phase (이미 착지) | SPEC 산출물 |
| `.moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/progress.md` | run/sync | §E 시그널 (상태 전환 소관 에이전트) |
| `.moai/reports/t1627/` | run/sync | 카드 증거 (verdict 등) |

**PRESERVE (무변경)**: `.claude/agents/harness/hns-release-update-specialist.md` · `.claude/workflows/hns-release-update-run.js` · `.claude/commands/harness/release-update/manifest.json` (판독 전용 — REQ-RSA-011) · `internal/template/templates/` (미러 금지) · 상위 SPEC `.moai/specs/SPEC-RELUP-DUALAXIS-001/` (completed — REQ-RSA-011) · 그 외 리포 전체.

**개발 방법**: TDD (tdd — 신규 테스트 파일). RED: 검사 함수+probe 단위 테스트를 먼저 쓰고 헬퍼 미구현 상태로 `go test` 실행 → 컴파일 실패(`undefined: …`)를 RED로 관측. GREEN: 헬퍼 구현 후 전체 패밀리 GREEN. REFACTOR: 없음(단일 패스; 다듬기는 GREEN 관측 뒤 같은 마일스톤 안에서).

**tier**: M — 산출물 3종, 신설 파일 1개(+SPEC 디렉터리), LOC ~500–700 (헬퍼 ~250 + 테스트 ~350). REQ 12 / AC 12 (천장 16 이내).

## §B Known Issues (카드 적용 분)

- **B5 CI 3-tier 인지**: 신설 테스트 파일은 `go test`·`golangci-lint` 양쪽에 걸린다 — 테스트 파일도 lint 대상이다(errcheck·revive 등). 착지 전 `golangci-lint run internal/harness/...` 관측을 §E에 넣는다.
- **B6 spec-lint 헤딩**: 본 SPEC 디렉터리의 spec.md는 `### Out of Scope — <topic>` h3 형식을 쓴다(MissingExclusions 회피) — 이미 적용.
- **B8 워킹 트리 위생**: `.moai/state/`, `.moai/harness/` 편집 금지. 커밋은 지정 pathspec만.
- **B10 범위 규율**: §A PRESERVE 이외 무변경. 특히 하네스 3표면은 **판독 전용** — run-phase가 실수로 고치면 scope 위반이다.
- **B11 사용자 질의 금지**: leaf worker — 열린 질문은 spec.md §1.2 결정 기록으로 봉쇄돼 있다(미결 항목: 블로킹 4 재분류는 run/sync 소관 — §D3).
- **카드 특이 제약 — `internal/` 문자열 상한**: 검사 테스트 파일은 리터럴 `last-codex-version`을 담지 않는다(상위 LED-011 `grep -rn "last-codex-version" internal/` 0힛 회귀 가드 — AC-RSA-010). S3 토큰 6종·앵커·판독 경로 어느 것도 이 문자열을 요구하지 않으므로 충돌하지 않는다. 구현 시 이 리터럴이 코멘트에라도 들어가면 가드를 깬다.
- **카드 특이 제약 — 코드펜스 안의 판정 대상**: Phase 3 URL 목록은 펜스 안에 있다(spec.md §1.1 A8) — 헬퍼의 경계 모드(마스킹)와 매칭 모드(제거)를 혼동하면 S1이 실트리에서 스스로 적색이 된다. 구현 순서에서 S1 양성 대조를 먼저 돌려 확인한다.

## §C Pre-flight (착지 전 재측정 — 귀속: 카드 워크트리 t1627 @ `79d0a0a47d`, branch `WT-d-20261010t000832z-efde`, 2026-10-13 plan 단계 실측)

```bash
git branch --show-current; git rev-parse --short HEAD     # 기대: WT-d-20261010t000832z-efde / 79d0a0a47d (이동 시 STOP — 분기 기준 재확인)
go build ./...                                            # 기저 녹색
test -f internal/harness/relup_structure_audit_test.go    # RED-now 귀속 관측: 출력 없음 · exit 1 (acceptance AC-RSA-001 RED 셀)
go test -run '^(TestRelupStructurePhase3URLs|TestRelupStructureCodexParagraph|TestRelupStructureWatchNormSentence|TestRelupStructureBPFaces|TestRelupStructureManifestDomain)$' ./internal/harness/   # RED-now: `[no tests to run]` 공집합 토큰 관측 (A11 — 2026-10-13 재관측 0.584s)
grep -c "last-codex-version" internal/harness/applier_test.go   # 0 · exit 1 — 신설 파일이 들어갈 패키지의 현재 상태 (AC-RSA-010 전제)
```

RED-now 전체 귀속은 acceptance.md §A 트리 핀(`79d0a0a47d`)이다. 상태 파일 특이사항 없음 — 본 카드는 `.moai/state/`를 만지지 않는다.

## §D Constraints (앵커 고정 — run-phase 재량 금지)

### §D1 고정 앵커 (내용 앵커 — 줄 번호는 `79d0a0a47d` 참고값; 검사 구현이 줄 번호 상수를 갖는 것 금지 — spec 결정 D6)

| 면 | 고정 앵커 (리터럴) | 판정 리터럴 | AC |
|---|---|---|---|
| S1 | specialist Phase 3 블록 — 줄시작 `### Phase 3` 헤딩부터 `### Phase 4` 헤딩 직전 (참고: 167행–194행). 블록 **내용물은 펜스 포함 원문**(URL 목록은 펜스 안 — 186–191행)에서 **HTML 주석 스팬을 먼저 제거**한 뒤 목록 항목을 읽는다 — 주석은 펜스 밖에서만 인정하고 펜스 내용물은 보존한다 (0.2.0 수리 P1; 주석 랩 목록 mutant = M8) | 목록 항목( `- ` 접두) 6종 전문: `https://code.claude.com/docs/en/{hooks, sub-agents, skills, plugins, mcp, settings}` — 각각 항목 단위 파싱: scheme=`https` · host=`code.claude.com` · path=`/docs/en/<facet>` · query 없음 · fragment 없음 · 항목 끝 = 종결 · 각 facet 정확히 1 (변이: M1 `?next=` 랩 · M8 주석 랩 목록 — 모두 0) | AC-RSA-002 |
| S2 | 동일 블록 경계 — 줄시작 `### Phase 3` / `### Phase 4`, 펜스·주석 마스킹, **각각 정확히 1회** (실측 전제: 각 1회 — spec §1.1 A10) | 0회·2회 = 검사 실패(오류) — 첫 출현 침묵 해석 금지 | AC-RSA-003 |
| S3 | specialist 코덱스 단락 — 줄시작 앵커 `**Codex axis state`부터 **종결 빈 줄**(참고: 77행) 또는 다음 줄시작 헤딩 중 먼저 오는 것까지 | 경계 안 6토큰 각 ≥1: `rust-v0.161.0` · `emit warning` · `last_analyzed_version` · `last_analyzed_date` · `last_master_research` · `analysis_history[]`. 기본값 문장 쌍(`rust-v0.161.0`+`emit warning`)이 경계 안 필수 | AC-RSA-004 |
| S4 | run.js 테마 체크리스트 블록 — 줄시작 `// Standing 6-theme`(참고: 79행)부터 `const CODEX_THEME_CHECKLIST` 이후 첫 `];` 행(참고: 92행)까지 | 감시 규범 문장 전체(공백 정규화 리터럴, 주석 접두 제거·행 결합 재구성): `Alpha-window themes remain WATCH-LIST observations — adoption is judged only when the theme lands in a stable release; an alpha-window theme is never reported as adopted drift (REQ-RDX-009).` — 블록 안 정확히 1회 | AC-RSA-005 |
| S5 | specialist BP 섹션 — 매칭 전 **주석·펜스 제거**(매칭 모드). 헤딩면: 대소문자 무시 줄시작 헤딩 `best-practice` ≥1 (참고: 299행 `## Best-Practices Axis (standing procedure — every sweep)`). deliverable면: 첫 best-practice 헤딩부터 다음 동급·상위 헤딩 직전까지의 블록 안 `HTML proposal report` ≥1 (참고: 316행) — **비주석 행만 계수**. BP 헤딩 선택은 **class F first-match**(첫 일치 헤딩이 블록을 선택하고 중복 헤딩은 실패·재선택이 아니다 — 결정 D8, 0.2.0 수리 P2) | AC-RSA-006 |
| S6 | manifest.json **JSON 디코딩** — 최상위 `domain` 키 존재·문자열형·3절 리터럴 보유: `Claude Code` · `Codex CLI upstream change tracking` · `best-practices axis` (참고값: 3행). 다른 키는 무게 0 | AC-RSA-007 |
| 헬퍼 | 비공개 함수群(같은 파일) — **스코프 파라미터 마스크/제거 원시**(주석·펜스 독립 선택, 행 위치 보존: S1은 주석만 제거·펜스 보존(0.2.0 수리 P1), S5는 둘 다 제거) · `lineStartHeadingBlocks`(**class E exactly-once** 오류 반환 — S2 경계 마커) · **class F first-match 선택자**(S5 BP 헤딩 — 결정 D8) · `countInBlock` 계열. **git 호출 금지**: 파일에 `os/exec`·`exec.Command`·`"git ` 문자열 0건 | AC-RSA-001·009 |
| 테스트 이름 | 5개 테스트 함수: `TestRelupStructurePhase3URLs`(S1·S2) · `TestRelupStructureCodexParagraph`(S3) · `TestRelupStructureWatchNormSentence`(S4) · `TestRelupStructureBPFaces`(S5) · `TestRelupStructureManifestDomain`(S6) — probe 단위 테스트 이름에 `M1_`…`M10_` 접두 각 1건 이상, 10종 (`go test -v` 판독면) | AC-RSA-001·008 |

### §D2 결정 전파 (재논의 금지 — spec.md §1.2와 쌍)

- D6 전파 — 앵커가 내용 기준이므로, S4 문장 리터럴을 바꾸려면 spec §4 REQ-RSA-005와 본 §D1 S4 행을 동시 갱신한다. S6 절 리터럴 변경은 상위 REQ-RDX-010(결정 D9·LED-001/002)과 동시 갱신을 요구한다(spec §7).
- D4 전파 — 인계 표의 VERIFIED→보류 재분류는 acceptance §D.2만 고치는 것이 아니라 REQ-RSA-012의 열거와 함께 갱신한다.
- D7 전파 — 블로킹 4 재분류는 run/sync에서 결정되며, 그 결정과 근거는 **본 카드의 자기 표면**에 기록된다: `.moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/progress.md` + `.moai/reports/t1627/` 증거 + 리더 완료 보고. 상위 부채 원장(`.moai/worktrees/t1579/…`)은 **본 카드의 편집 범위 밖**이라 레인이 쓰지 않는다 — 원장 갱신이 필요하면 완료 보고에 갱신 요청 정보를 실어 **리더가 기록**한다 (0.3.0 수리 P2-a — 범위 밖 트리 기입 지시 제거).

### §D3 상위 부채 B1–B5 배제 (감사자 필독)

상위 카드 t1579 iter7의 차단 발견 **B1–B5는 본 카드의 결함 후보가 아니다**. 부채 원장(`.moai/worktrees/t1579/.moai/reports/t1579/debt-for-t1627.md`)이 승인 id `d-20261010T064334Z-8df9`(재-FAIL 분기 — 남은 지적을 t1627 부채로 이전)와 `d-20261010T074921Z-bb5d`(debt-admit)로 처분했고, 원장 항목 A가 "재감사에서 차단 제외 대상"이라고 명시한다. B1–B5(runner.js 폴백 창 상한 없음 · 고정 창 목록 전진 없음 · plan.md 인용 따옴표 · HEAD 개수 불일치 · D9 인용 누락)는 **상위 SPEC의 plan.md 내용에 관한 항목**으로 본 카드의 검사 대상(하네스 3표면 구조 면)과 무관하다. 본 카드의 plan-audit에서 B1–B5를 차단으로 계수하는 것은 원장과 승인 id의 재해석이므로 금지다. 본 카드가 이월받은 것은 원장 항목 C·D·E(검증 보류 10종의 소유권)다 — 그 소유권의 처분이 REQ-RSA-012·AC-RSA-012의 인계 표다.

## §E 검증 동사 (단일 호출 — 관측 전용 `; echo "exit=$?"` 래퍼 허용은 상위 acceptance §A와 동일한 예외다)

```bash
# GREEN 1 — 전체 신설 패밀리 (AC-RSA-001): 기대 exit 0 + `ok github.com/modu-ai/moai-adk/internal/harness`
go test ./internal/harness/
# GREEN 2 — probe 판독면 (AC-RSA-008): 기대 exit 0 + M1_…M10_ 단위 테스트 전수 PASS 행
go test -v -run '^(TestRelupStructurePhase3URLs|TestRelupStructureCodexParagraph|TestRelupStructureWatchNormSentence|TestRelupStructureBPFaces|TestRelupStructureManifestDomain)$' ./internal/harness/
# RED 재관측 (구현 전 — §C와 동일): compile error = RED 증거 (E8 축자)
go test -run '^(TestRelupStructurePhase3URLs|TestRelupStructureCodexParagraph|TestRelupStructureWatchNormSentence|TestRelupStructureBPFaces|TestRelupStructureManifestDomain)$' ./internal/harness/
# guard 1 — no-git (AC-RSA-009): 기대 0힛 · exit 1
grep -nE 'os/exec|exec\.Command|"git' internal/harness/relup_structure_audit_test.go
# guard 2 — 상위 LED-011 유지 (AC-RSA-010): 기대 0힛 · exit 1 (착지 후에도)
grep -rn "last-codex-version" internal/
# guard 3 — 인계 표 행 단위 판정 (AC-RSA-012): §D.1 열거의 12개 행 앵커 패턴 전수 — 각 관측 1, 어긋난 행 0.
#   열거 SSOT는 acceptance §D.1 AC-RSA-012다(인계 6행 `인계: VERIFIED` + 보류 유지 6행 `판정 보류 유지 — CX-<n>`). 형태 예 2건:
grep -cE '^\| AC-RDX-014 .*인계: VERIFIED' .moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/acceptance.md   # 기대 1
grep -cE '^\| AC-RDX-006 .*판정 보류 유지 — CX-8' .moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/acceptance.md   # 기대 1
grep -cE '^\| AC-RDX-.*인계: VERIFIED' .moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/acceptance.md   # 집계 연출: 기대 6 (판정면 아님 — 0.3.0 P2-b)
grep -cE '^\| AC-RDX-.*판정 보류 유지' .moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/acceptance.md   # 집계 연출: 기대 6 (판정면 아님 — 0.3.0 P2-b)
# lint — golangci-lint (신설 파일 포함 범위): 기저 대비 신규 이슈 0건
golangci-lint run internal/harness/... --timeout=2m
# 포맷 — gofmt: 기대 출력 없음
gofmt -l internal/harness/
```

AC-RSA-002..007의 기계 판정은 GREEN 2의 `-v` 출력이다 — 각 면의 양성 단위 테스트(실트리 기대값)와 음성 단위 테스트(probe 변이 0)가 전부 PASS여야 하고, probe 단위 테스트는 변이에 대해 **단언이 0을 확인하는 형태**다(단언 없는 실행은 §1.1 보고-비-판정 결함이다 — 구현 시 주의).

## §F 마일스톤

- **M1 — 헬퍼 + S1·S2 (RED→GREEN)**: 테스트 파일 골격 + 5테스트 중 `TestRelupStructurePhase3URLs`(M1·M2·M8 변이 + 양성)를 먼저 쓰고 RED 관측(헬퍼 미구현 컴파일 실패 축자) → 헬퍼(스코프 마스킹·경계·이중 클래스 중복 처리·URL 파싱) 구현 → GREEN. §C 전 배시 재실행.
- **M2 — S3·S4·S5·S6 (RED→GREEN)**: 나머지 4개 테스트 함수 + M3–M7 변이 → RED(신규 헬퍼 기능 미구현 단언 실패 축자) → 구현 → §E GREEN 2(5함수 고정 앵커 선택자 `-v`) 전면 GREEN. BLK-014G 경계 축소 토큰 벡터 실측값을 진행 기록에 남긴다(기대: 전 토큰 ≥1 — 정확 벡터는 관측 기록).
- **M3 — 가드·문서·착지**: AC-RSA-009(no-git grep)·AC-RSA-010(LED-011 유지)·AC-RSA-012(인계 표) 관측, `gofmt -l`·`golangci-lint` 0 신규, progress 기록, §E 동사 전수 출력을 증거로 남긴다. sync는 표준 3-phase close(manager-docs).
