---
id: SPEC-RELUP-STRUCT-AUDIT-001
title: "acceptance — release-update 구조 검사 설계 (마크다운 구조 파서 6면 + M1–M16 음성 대조)"
created: 2026-10-13
author: manager-spec
tier: M
---

# acceptance: SPEC-RELUP-STRUCT-AUDIT-001

## §A 범위·판정 규율

- 모든 릴리스 블로킹 AC는 verification-completeness.md §2 two-cell 규율을 따른다: **RED-now 셀**(구현 전 트리에서 적색 관측 — 단일 읽기전용 명령 + 그 명령의 축자 stdout + exit code 독립 필드 + 트리 SHA 핀)과 **green-path 셀**(어느 마일스톤이 뒤집는지 + 녹색 출력 형태)이 쌍으로 존재한다.
- **트리 핀**: 본 SPEC의 RED-now 전체 관측은 `79d0a0a47d` (카드 워크트리 `.moai/worktrees/t1627`, branch `WT-d-20261010t000832z-efde`)에서 2026-10-13 수행했다. 문서 수준 핀이 이 문서이며, 개별 AC 핀이 없는 한 이 핀이 구속한다.
- **RED의 올바른 이유 (신규 검사기 클래스)**: 본 SPEC의 블로킹 AC는 **존재하지 않는 검사기**를 착지시킨다 — RED-now의 사유는 "검사기가 오늘 부재다"이며, 그 관측 가능한 증거는 (a) 파일 부재(`test -f` exit 1)와 (b) **공집합 스윕 토큰** `[no tests to run]`(verification-completeness §1.1 — 선택 집합이 빈 녹색은 통과가 아니라 미측정이다)다. 구현(테스트 파일+헬퍼)이 양쪽을 뒤집는다. 잘못된 사유의 적색(선존재 파일이 못 만지는 wrong-reason red)은 없다.
- **클래스 데모와 본 RED의 구분**: 상위 plan-audit iter3 §9.2가 기록한 M1–M7 probe 재현(구 텍스트 패턴 검사가 변이를 통과시킨 실측)은 **상위 트리(7fd68bbe2 계열)의 이력 증거**다 — 본 SPEC의 RED-now가 아니다. 본 RED는 A11 관측(§1.1)이며 인용 관계를 혼동하지 않는다.
- **명령 형태**: 전부 단일 호출(파이프·리다이렉트·`&&`·`;`·서브셸 없음). 관측 전용 래퍼 `; echo "exit=$?"`는 상위 acceptance §A와 동일한 예외로 허용한다 — pinned 관측의 exit 코드 기록 용도에 한하며 명령 본체는 단일 호출이다. grep 일치 0개는 exit 1이다 — "빈 출력 + exit 1"은 완전한 관측이다.
- **비재현 관측 처분**: 재실행 불가능하거나 이미 녹색인 부재·유지 클레임은 회귀 가드로 분류하고 릴리스 블로킹에서 제외한다(undecidable disposition — AC-RSA-009·010·011이 해당).
- **`go test` 출력의 기계 판독**: GREEN 판정은 (i) exit 0, (ii) `ok github.com/modu-ai/moai-adk/internal/harness` 행, (iii) `-v` 모드에서 probe 단위 테스트(`M1_`…`M16_` 접두)가 `--- PASS`로 전수 출현 — 셋을 함께 본다. probe 단위 테스트는 변이에 대해 **단언이 0/실패를 확인하는 형태**여야 하고, 단언 없는 실행은 verification-completeness §1.1의 보고-비-판정 결함으로 본 AC를 만족하지 않는다.

## §D AC Matrix

| AC | 분류 | 주장 (Then) | RED-now (관측) | Green path |
|----|------|-------------|----------------|------------|
| AC-RSA-001 | 블로킹 | `internal/harness/relup_structure_audit_test.go`가 존재하고(package `harness`, test 전용) `go test ./internal/harness/`가 exit 0으로 신설 패밀리를 실행한다 — `[no tests to run]` 토큰이 사라진다 | `test -f internal/harness/relup_structure_audit_test.go` → (출력 없음) · exit **1**; `go test -run '^(TestRelupStructurePhase3URLs\|TestRelupStructureCodexParagraph\|TestRelupStructureWatchNormSentence\|TestRelupStructureBPFaces\|TestRelupStructureManifestDomain)$' ./internal/harness/` → `ok github.com/modu-ai/moai-adk/internal/harness 0.584s [no tests to run]` · exit 0 (공집합 토큰 = 부재 증거; 최초 접두 선택자 관측 `0.649s`도 동일 토큰) — 핀 `79d0a0a47d` | M1 → 파일 존재 + `go test ./internal/harness/` exit 0, 신설 테스트 실행 |
| AC-RSA-002 | 블로킹 (S1 — M1·M8) | Phase 3 블록의 URL 목록 항목이 항목 단위 구조 파싱(scheme·host·path·query·fragment·종결)을 통과하고 6 facet 각각 정확히 1이다; `?next=` 랩 변이(M1)와 **주석 랩 목록 변이(M8 — 목록 전체를 HTML 주석 안에 넣음)**는 각각 0이다 — 주석 스팬은 목록 열거 전에 제거되고 펜스 내용물은 보존된다 (0.2.0 수리 P1) | AC-RSA-001 RED 셀과 동일 귀속 — 검사기 부재 (핀 `79d0a0a47d`); 클래스 데모: iter3 §9.2 M1 (구 검사 통과, 이력 증거); M8은 본 수리의 신설 probe다 | M1 → `go test -v -run '^TestRelupStructurePhase3URLs$' ./internal/harness/` exit 0 — 양성 단위 테스트(6 facet 각 1)와 `M1_`·`M8_` 음성 단위 테스트(변이 0) 전부 PASS |
| AC-RSA-003 | 블로킹 (S2 — M2) | Phase 3 블록 경계가 줄시작 `### Phase 3`·`### Phase 4`에서만 성립(펜스·주석 마스킹)하고 각각 정확히 1회다; 산문 교차 참조 변이는 경계를 옮기지 못하고, 줄시작 헤딩 2회 변이는 검사 실패(오류)다 | AC-RSA-001과 동일 (핀 `79d0a0a47d`); 실트리 전제 관측: `grep -c '^### Phase 3'` = 1 · `'^### Phase 4'` = 1 (§1.1 A10) | M1 → `M2_` 단위 테스트 2종(무이동·중복 실패) PASS + 양성 PASS |
| AC-RSA-004 | 블로킹 (S3 — M3) | 코덱스 단락이 앵커 `**Codex axis state`부터 종결 빈 줄까지 경계되고, 경계 **안**에서 기본값 문장 쌍(`rust-v0.161.0`+`emit warning`)과 6토큰이 각 ≥1이다; 문장을 경계 밖 트레일링 단락으로 옮긴 변이(M3)는 0이다 | AC-RSA-001과 동일 (핀 `79d0a0a47d`); 클래스 데모: iter3 §9.2 M3 (구 검사 `[4, 1, 3, 2, 2, 1]` 통과, 이력 증거) | M2 → `M3_` 단위 테스트 2종(후첨 무계수·이동 실패) PASS + 양성(경계 안 6토큰 ≥1) PASS — 실측 벡터를 진행 기록에 남긴다 |
| AC-RSA-005 | 블로킹 (S4 — M4) | 테마 체크리스트 블록(`// Standing 6-theme` .. `];`) 안에서 감시 규범 문장 **전체**(공백 정규화 리터럴)가 정확히 1회 일치한다; 토큰만 남긴 변이(M4)는 0이다 | AC-RSA-001과 동일 (핀 `79d0a0a47d`); 클래스 데모: iter3 §9.2 M4 (구 검사 `[1, 1]` 통과, 이력 증거) | M2 → `M4_` 음성 PASS + 양성(문장 정확히 1) PASS |
| AC-RSA-006 | 블로킹 (S5 — M5·M6) | 주석·펜스 제거 후 BP 헤딩면(대소문자 무시 줄시작 `best-practice` 헤딩) ≥1이고 BP 블록 안 deliverable `HTML proposal report` ≥1이다(비주석 행만); 주석 안 deliverable(M5)·주석 안 가짜 헤딩(M6) 변이는 0이다 — 상위 plan.md:96·acceptance.md:32의 주장(실트리 통과)은 유지된다. BP 헤딩 선택은 **class F first-match**다 (0.2.0 수리 P2 — 결정 D8: 첫 일치 헤딩이 블록을 선택하고 중복 BP 헤딩은 실패·재선택이 아니다) | AC-RSA-001과 동일 (핀 `79d0a0a47d`); 클래스 데모: iter3 §9.2 M5·M6 (이력 증거); 실트리 전제: BP 헤딩 존재(299행), HTML 주석 0건(A9 — 제거 규칙이 오늘의 통과를 바꾸지 않음) | M2 → `M5_`·`M6_` 음성 PASS + 양성(헤딩 ≥1, deliverable ≥1) PASS |
| AC-RSA-007 | 블로킹 (S6 — M7) | manifest.json을 JSON으로 디코딩해 **최상위** `domain` 문자열이 3절 리터럴(`Claude Code`·`Codex CLI upstream change tracking`·`best-practices axis`)을 보유한다; 같은 줄 뒤쪽 키/다른 키가 절을 실은 변이(M7)는 0이다 | AC-RSA-001과 동일 (핀 `79d0a0a47d`); 클래스 데모: iter3 §9.2 M7 (이력 증거); 실트리 전제: manifest.json:3 domain 값이 3절 보유 (A7) | M2 → `M7_` 음성 PASS + 양성(3절 보유) PASS |
| AC-RSA-008 | 블로킹 (probe 판독면) | `go test -v -run '^(TestRelupStructurePhase3URLs\|TestRelupStructureCodexParagraph\|TestRelupStructureWatchNormSentence\|TestRelupStructureBPFaces\|TestRelupStructureManifestDomain)$' ./internal/harness/`가 exit 0이고 `M1_`…`M16_` 접두 단위 테스트(16종)가 전부 `--- PASS`로 출현한다 — 음성 대조의 실행이 출력에서 판독된다(스윕 집합 비가시성 봉쇄, §1.1) | AC-RSA-001과 동일 (핀 `79d0a0a47d` — 검사기 부재 시 출현할 단위 테스트가 0건) | M1·M2 착지 → `-v` 출력에 16종 probe 단위 테스트 PASS 행 관측 (축자를 진행 기록에 남긴다) |
| AC-RSA-009 | 회귀 가드 | 검사 테스트 파일에 git 호출 흔적이 없다: `grep -nE 'os/exec\|exec\.Command\|"git' internal/harness/relup_structure_audit_test.go` → 0힛 · exit 1 | — (오늘은 파일 부재로 미적용 — grep이 파일을 찾지 못하는 exit 2는 0힛이 아니다; 이 가드의 첫 관측은 M3 착지 트리에서 한다, REQ-RSA-008) | M3 착지 후 유지 — 착지 트리에서 0힛·exit 1 관측 |
| AC-RSA-010 | 회귀 가드 (상위 LED-011 유지) | `grep -rn "last-codex-version" internal/` → 0힛 · exit 1 유지 — 검사 테스트 파일이 리터럴을 심지 않는다(상위 AC-RDX-011 (a) 보존) | — (오늘 녹색: 0힛 — 핀 `79d0a0a47d` 관측) | M3 착지 후 유지 — 변화 없음이 PASS |
| AC-RSA-011 | 회귀 가드 (범위) | 종결 시 4채널 변경 스코프(커밋 `git diff --name-only "$CARD_BASE"..HEAD` · 스테이지 · 비추적 · 비스테이지)가 전부 허용 경로만 반환한다 — `internal/harness/relup_structure_audit_test.go` · `.moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/` · `.moai/reports/t1627/` (§D.5 항목 4). bare git diff의 공집합은 측정이 아니라 부재다 | — (오늘 녹색: 깨끗한 트리 — 핀 `79d0a0a47d`) | M3·sync 종결 시 관측 |
| AC-RSA-012 | 형식 (인계 표) | acceptance §D.2 인계 표의 12행이 **행 단위**로 자기 상태를 읽는다 — §D.1 AC-RSA-012가 열거한 12개 행 앵커 패턴(id + 상태 토큰 + 클래스 코드) 각각의 `grep -cE` 관측이 **1**이고, 어긋난 행(관측 ≠ 1)의 개수는 **0**이다: 인계 6행(AC-RDX-009 육면·014·015·008·010·001 LED-001C)은 `인계: VERIFIED`, 보류 유지 6행(003·004·005 — CX-7, 006 — CX-8, 016 — CX-13, 017 — CX-12)은 `판정 보류 유지` + 클래스 코드를 싣는다 (REQ-RSA-012 — 0.3.0 수리 P2-b, 총계 판독 폐기) | — (plan 단계 착지 — 본 파일이 그 산물이다) | 착지됨 — §D.1 열거의 12 패턴 전수 관측 (축자를 진행 기록에 남긴다) |

**집계**: AC 12 = 블로킹 8 (001·002·003·004·005·006·007·008) + 회귀 가드 3 (009·010·011) + 형식 1 (012). RED-now 귀속 = `79d0a0a47d` 단일 핀(문서 수준 — §A), 블로킹 8종이 AC-RSA-001의 관측을 공유 귀속으로 인용한다.

## §D.1 시나리오 (Given-When-Then — 블로킹 8종 + 가드 3종 + 형식 1종 = 12)

- **AC-RSA-001** — **Given** 검사기가 존재하지 않는 트리(`79d0a0a47d`)로, **When** `test -f internal/harness/relup_structure_audit_test.go`와 `go test -run '^(TestRelupStructurePhase3URLs|TestRelupStructureCodexParagraph|TestRelupStructureWatchNormSentence|TestRelupStructureBPFaces|TestRelupStructureManifestDomain)$' ./internal/harness/`를 실행하면, **Then** 전자는 출력 없음·exit 1이고 후자는 `[no tests to run]` 토큰을 출력한다(공집합 스윕). **Green**: M1 착지 뒤 같은 두 관측이 파일 존재·테스트 실행으로 뒤집힌다.
- **AC-RSA-002** — **Given** Phase 3 블록에 6종 캐노니컬 URL 목록이 펜스 안 목록 항목으로 존재하는 상태로, **When** S1 검사가 블록 안 각 목록 항목을 URL로 파싱하면, **Then** 6 facet(`hooks`·`sub-agents`·`skills`·`plugins`·`mcp`·`settings`)이 각각 정확히 1이고(scheme=`https`·host=`code.claude.com`·path=`/docs/en/<facet>`·query 없음·fragment 없음·항목 끝 종결), `?next=https://evil.example.com` 랩 변이(M1)에서 해당 facet은 0이다. 목록 전체를 HTML 주석 안에 넣은 변이(M8)도 0이다 — 주석 스팬은 목록 열거 전에 제거되고 펜스 내용물은 보존된다(0.2.0 수리 P1).
- **AC-RSA-003** — **Given** 같은 문서로, **When** S2 검사가 블록 경계를 계산하면, **Then** 경계는 줄시작 `### Phase 3`·`### Phase 4`에서만 성립하고 각각 정확히 1회다. 산문 교차 참조 변이(펜스 안·들여쓰기·역따옴표 안 `### Phase 3` 텍스트)는 경계를 옮기지 못하고, 줄시작 헤딩이 2회 나타나는 변이는 첫 출현 침묵 해석 대신 **오류로 실패**한다.
- **AC-RSA-004** — **Given** specialist Phase 0의 코덱스 단락(앵커 `**Codex axis state`)으로, **When** S3 검사가 앵커부터 종결 빈 줄(또는 다음 줄시작 헤딩)까지 경계하면, **Then** 경계 안에서 `rust-v0.161.0`·`emit warning`·`last_analyzed_version`·`last_analyzed_date`·`last_master_research`·`analysis_history[]`가 각 ≥1이다. 블록 끝 뒤에 덧붙인 문장은 계수되지 않고, 기본값 문장을 경계 밖 트레일링 단락으로 옮긴 변이(M3)는 0이다 — 경계 축소 전 구 검사의 `[4, 1, 3, 2, 2, 1]` 벡터는 이력이다.
- **AC-RSA-005** — **Given** run.js의 테마 체크리스트 블록으로, **When** S4 검사가 주석 접두를 제거·행을 결합·공백을 정규화한 뒤 감시 규범 문장 전체를 세면, **Then** 일치가 정확히 1이다. `WATCH-LIST`·`stable release` 토큰만 남기고 문장을 해체한 변이(M4)는 0이다.
- **AC-RSA-006** — **Given** specialist의 BP 섹션으로, **When** S5 검사가 주석·펜스를 제거한 텍스트에서 헤딩면과 deliverable면을 재면, **Then** 줄시작 `best-practice` 헤딩(대소문자 무시)이 ≥1이고 BP 블록 안 `HTML proposal report`가 ≥1이다. deliverable을 주석 안에만 두는 변이(M5)와 가짜 헤딩을 주석 안에 두는 변이(M6)는 0이다 — 상위 SPEC의 plan.md:96(AC-RDX-010 앵커 행)·acceptance.md:32(AC-RDX-008 행) 주장은 실트리에서 유지된다. BP 헤딩 선택은 **class F first-match**다(결정 D8, 0.2.0 수리 P2) — 첫 일치 헤딩이 블록을 선택하고 중복 BP 헤딩은 실패·재선택이 아니다(REQ-RSA-001(d)의 이중 클래스 계약; S2 경계 마커만 class E exactly-once다).
- **AC-RSA-007** — **Given** manifest.json으로, **When** S6 검사가 JSON 디코딩 후 최상위 `domain` 값을 재면, **Then** 그 값은 문자열이고 `Claude Code`·`Codex CLI upstream change tracking`·`best-practices axis` 세 절을 보유한다. `"note": "Claude Code"` 같은 같은 줄 뒤쪽 키 변이(M7)는 0이다 — 줄 패턴이 아니라 키의 값만 판정한다.
- **AC-RSA-008** — **Given** 착지된 검사 파일로, **When** `go test -v -run '^(TestRelupStructurePhase3URLs|TestRelupStructureCodexParagraph|TestRelupStructureWatchNormSentence|TestRelupStructureBPFaces|TestRelupStructureManifestDomain)$' ./internal/harness/`를 실행하면, **Then** exit 0이고 `M1_`…`M16_` 접두 단위 테스트(16종)가 전부 `--- PASS`로 출력된다 — 음성 대조가 실행됐다는 사실이 출력에서 판독된다.
- **AC-RSA-009** — **Given** 착지된 검사 파일로, **When** `grep -nE 'os/exec|exec\.Command|"git' internal/harness/relup_structure_audit_test.go`를 실행하면, **Then** 0힛·exit 1이다 — 검사기는 git 채널이 없다(REQ-RSA-008).
- **AC-RSA-010** — **Given** `internal/` 전체로, **When** `grep -rn "last-codex-version" internal/`을 착지 트리에서 실행하면, **Then** 0힛·exit 1이 유지된다 — 상위 LED-011 회귀 가드 보존.
- **AC-RSA-011** — **Given** 종결 시점 트리로, **When** 4채널 변경 스코프를 개별 관측하면, **Then** 네 채널 전부가 허용 경로만 반환한다. bare `git diff --name-only`의 공집합은 측정이 아니라 부재다.
- **AC-RSA-012** — **Given** 본 acceptance 문서로, **When** §D.2 인계 표의 행별 앵커 패턴 12개를 각각 `grep -cE '<패턴>' .moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/acceptance.md`로 실행하면, **Then** 12개 관측이 전부 1이고 **어긋난 행(관측 ≠ 1)은 0**이다 — 총계 판독(`grep -c 'VERIFIED'` ≥6)은 보류 6행 전부가 삭제돼도 통과하므로 0.3.0 수리(P2-b)에서 판정면에서 폐기됐다. 열거(행 단위 — 인계 6행은 상태 토큰 `인계: VERIFIED`, 보류 유지 6행은 `판정 보류 유지` + 클래스 코드를 앵커로 쓴다):
  - 인계 VERIFIED 6행 — `^\| AC-RDX-009 .*인계: VERIFIED` · `^\| AC-RDX-014 .*인계: VERIFIED` · `^\| AC-RDX-015 .*인계: VERIFIED` · `^\| AC-RDX-008 .*인계: VERIFIED` · `^\| AC-RDX-010 .*인계: VERIFIED` · `^\| AC-RDX-001 .*인계: VERIFIED`
  - 판정 보류 유지 6행(클래스 코드 포함) — `^\| AC-RDX-003 .*판정 보류 유지 — CX-7` · `^\| AC-RDX-004 .*판정 보류 유지 — CX-7` · `^\| AC-RDX-005 .*판정 보류 유지 — CX-7` · `^\| AC-RDX-006 .*판정 보류 유지 — CX-8` · `^\| AC-RDX-016 .*판정 보류 유지 — CX-13` · `^\| AC-RDX-017 .*판정 보류 유지 — CX-12`
  - 집계 연출(보조 — 판정면 아님): `grep -cE '^\| AC-RDX-.*인계: VERIFIED'` = 6 · `grep -cE '^\| AC-RDX-.*판정 보류 유지'` = 6. 집계는 표 행만 앵커(`^\| AC-RDX-`)로 제한해 본문 서술의 VERIFIED 언급을 세지 않는다.

## §D.2 추적성 (AC ↔ REQ) + 인계(hanback) 표

| AC | REQ | mutant 봉쇄 |
|----|-----|-------------|
| AC-RSA-001 | REQ-RSA-001, REQ-RSA-010 | 검사기 부재·`[no tests to run]` 공집합 — 파일 신설로 뒤집힘 |
| AC-RSA-002 | REQ-RSA-002, REQ-RSA-009 | M1 (`?next=` 랩)·M8 (주석 랩 목록) — 항목 단위 구조 파싱 + 주석 스팬 정리 후 열거 (0.2.0 수리 P1) |
| AC-RSA-003 | REQ-RSA-003 | M2 (교차 참조 슬라이스 이동·중복 헤딩) — 줄시작+마스킹+**class E exactly-once** (결정 D8) |
| AC-RSA-004 | REQ-RSA-004 | M3 (후첨 계수·기본값 문장 이동) — 종결 빈 줄 경계+안-필수 |
| AC-RSA-005 | REQ-RSA-005 | M4 (토큰 잔존·문장 해체) — 문장 전체 정확히 1회 |
| AC-RSA-006 | REQ-RSA-006 | M5·M6 (주석 안 deliverable·가짜 헤딩) — 제거 후 매칭 + **class F first-match** 선택 (결정 D8) |
| AC-RSA-007 | REQ-RSA-007 | M7 (같은 줄 뒤쪽 키) — JSON 최상위 값 단언 |
| AC-RSA-008 | REQ-RSA-009, REQ-RSA-010 | 스윕 비가시성 — probe 단위 테스트 이름 판독면 |
| AC-RSA-009 | REQ-RSA-008 | git 호출 침입 — exec 흔적 0힛 |
| AC-RSA-010 | REQ-RSA-011 (상위 LED-011 보존) | `internal/` 리터럴 침입 |
| AC-RSA-011 | REQ-RSA-011 | 범위 초과 변경 — 4채널 스코프 |
| AC-RSA-012 | REQ-RSA-012 | 인계 누락·무단 인계·행 삭제 은폐 — 행 단위 앵커 판독면(12 패턴, 어긋난 행 0 — 0.3.0 수리 P2-b) |

### 인계 표 — 상위 SPEC-RELUP-DUALAXIS-001 판정 보류 면의 처분 (SSOT — REQ-RSA-012)

상위 부채 원장(항목 C·D·E — 소유자 t1627)의 10종(list E: 003·004·005·006·008·010·014·015·016·017)과 비-list-E 강등 면 2종(AC-RDX-009 육면·AC-RDX-001 LED-001C)에 대한 본 카드의 처분이다.

| 상위 면 (강등 사유) | 본 카드 표면 | 처분 |
|---|---|---|
| AC-RDX-009 육면 LED-022..027 (F1·F2 — 텍스트 패턴 우회) | S1·S2 | 인계: VERIFIED — CI 게이트 (`go test ./internal/harness/`) |
| AC-RDX-014 블록면 BLK-014G (F3 — 블록 끝 넘침) | S3 | 인계: VERIFIED — CI 게이트 |
| AC-RDX-015 블록면 BLK-015G (F4 — 토큰≠문장) | S4 | 인계: VERIFIED — CI 게이트 |
| AC-RDX-008 BP-008G (F7 — 주석 헤딩) | S5 | 인계: VERIFIED — CI 게이트 |
| AC-RDX-010 BP-010G (F7 — 주석 deliverable) | S5 | 인계: VERIFIED — CI 게이트 |
| AC-RDX-001 LED-001C 보존면 (F8 — 줄 스코프 누수) | S6 | 인계: VERIFIED — CI 게이트 |
| AC-RDX-003 (REQ-RDX-006 — 원장 C) | — | 판정 보류 유지 — CX-7 내용 계측 클래스(빈 상수·"Return ok." 프롬프트 mutant)로 구조 파서가 닫지 못한다 |
| AC-RDX-004 (REQ-RDX-007 — 원장 C) | — | 판정 보류 유지 — CX-7 동일 클래스 |
| AC-RDX-005 (REQ-RDX-008 — 원장 C) | — | 판정 보류 유지 — CX-7 동일 클래스 |
| AC-RDX-006 (REQ-RDX-003 — 원장 C) | — | 판정 보류 유지 — CX-8 산개 언급 클래스(전-file 계수) |
| AC-RDX-016 (REQ-RDX-013 — 원장 E) | — | 판정 보류 유지 — CX-13 리터럴 유지·규칙 역전 클래스(의미론) |
| AC-RDX-017 (REQ-RDX-015 — 원장 C) | — | 판정 보류 유지 — CX-12 paraphrase 클래스(의미론) |

비고: (i) 카드 본문이 AC-RDX-010의 REQ를 REQ-RDX-009로 적었으나 상위 acceptance §D.2의 매핑(AC-RDX-010 ↔ REQ-RDX-014)이 정본이다 — 본 표는 상위 매핑을 따른다. (ii) VERIFIED 인계의 게이트는 `go test` 스위트다 — 상위의 `grep`/python 단일 호출 형태를 대체하는 것이고, 상위 AC 문서 자체는 수정하지 않는다(completed — spec §6). (iii) 상위 블로킹 4(001·002·007·009)의 유지·복귀 재분류는 run/sync 미결 항목이다(spec 결정 D7·plan §D3). (iv) 상위 부채 B1–B5는 본 카드의 결함 후보가 아니다(plan §D3 — 감사자 필독). (v) **중복 처리 계약은 이중 클래스다** (결정 D8, 0.2.0 수리 P2): 이름 붙은 경계 마커(S2의 `### Phase 3`·`### Phase 4`)는 **class E exactly-once** — 0회·2회 출현은 검사 실패다; 선택 앵커(S5의 BP 헤딩)는 **class F first-match** — 첫 출현이 블록을 선택하고 중복은 실패·재선택이 아니다. 모든 면의 클래스는 spec REQ-RSA-001(d)에 명명돼 있고 이 표의 mutant 봉쇄 열과 §D.1 시나리오가 같은 클래스 이름을 쓴다.

## §D.5 종결 게이트 (Definition of Done)

1. 블로킹 AC 8종 전부 GREEN — RED→GREEN 관측 출력(exit code 포함)을 진행 기록이 축자 보유. AC-RSA-008의 `-v` 출력(16종 probe PASS 행)이 포함된다.
2. 회귀 가드 3종(AC-RSA-009·010·011) 기준선 관측 — 착지 트리에서 각 명령의 축자 출력.
3. 상위 면 양성 값 기록 — BLK-014G 경계 축소 토큰 벡터(기대: 전 토큰 ≥1, 실측 벡터 기록), S4 문장 일치 1, S5 헤딩/deliverable ≥1, S6 3절 보유 — 전부 `go test -v` 출력으로 귀속.
4. 변경 스코프 단언 — `CARD_BASE=$(git merge-base main HEAD)` 고정 후 4채널 개별 관측: (a) 커밋분 `git diff --name-only "$CARD_BASE"..HEAD`, (b) 스테이지분 `git diff --cached --name-only`, (c) 비추적분 `git status --porcelain`의 `??` 행, (d) 비스테이지 추적편집 `git diff --name-only`(인자 없음). 네 채널 전부가 §D AC Matrix AC-RSA-011 행의 허용 경로만 반환. bare `git diff`의 공집합은 부재다.
5. 하네스 3표면 무변경 — `git diff --name-only "$CARD_BASE"..HEAD -- .claude/` 공집합(빈 출력이 이 항목의 PASS다 — 대상 경로 자체가 무변경임을 보는 관측).
6. 인계 표(§D.2)가 그대로 판독된다 — AC-RSA-012.
7. 미결 판단 표지 0 — 블로킹 4 재분류는 run/sync의 명시적 이관 항목이지 열린 판단 표지가 아니다(spec §7·plan §D3에 근거 기록).

## Gaps (plan 단계가 관측하지 않은 것)

- **실트리 probe 재생과 양성 벡터의 실측** — 본 plan 단계에서 관측한 것은 검사기 부재(A11)와 실트리 전제(A7·A9·A10·A12)뿐이다. M1–M16 변이에 대한 0 판정·실트리 양성 값·BLK-014G 축소 경계의 실측 벡터는 run-phase의 테스트 실행 출력으로 관측된다.
- **iter3 §9.2 probe 재현의 재실행 없음** — 클래스 데모는 상위 감사의 이력 증거(7fd68bbe2 계열 트리)로 인용했고 본 plan 단계에서 재실행하지 않았다. 본 카드의 RED는 A11이다.
- **golangci-lint 기저** — plan 단계에서 `internal/harness` 패키지의 lint 기저를 재측정하지 않았다. run M3가 착지 전후 기저 대비 신규 0건을 관측한다(§E).
- **S4 문장 리터럴의 행걸침 변화** — A6 실측(81–84행)은 `79d0a0a47d` 기준이다. run-phase 착지 시점에 주석 행이 다시 걸침되면 리터럴 자체는 불변(공백 정규화)이므로 판정은 유효하다 — 행 번호는 참고값이다(결정 D6).
