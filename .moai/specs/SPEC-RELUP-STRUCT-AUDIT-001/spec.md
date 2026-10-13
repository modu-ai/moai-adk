---
id: SPEC-RELUP-STRUCT-AUDIT-001
title: "release-update 구조 검사 설계 — 마크다운 구조 파서 6면 검사와 M1–M8 음성 대조 (SPEC-RELUP-DUALAXIS-001 분할 잔여분)"
version: "0.3.0"
status: draft
created: 2026-10-13
updated: 2026-10-13
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/harness/relup_structure_audit_test.go (신설), .claude/agents/harness/hns-release-update-specialist.md (판독 전용), .claude/workflows/hns-release-update-run.js (판독 전용), .claude/commands/harness/release-update/manifest.json (판독 전용)"
lifecycle: spec-anchored
tags: "release-update, structure-audit, markdown-parser, mutant-probe, dev-only, dualaxis-sibling"
tier: M
---

# SPEC: release-update 구조 검사 설계 — 마크다운 구조 파서 6면 검사와 M1–M8 음성 대조

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-13 | manager-spec | 최초 작성 — 카드 t1627 (상위 카드 t1579 / SPEC-RELUP-DUALAXIS-001의 plan-audit 3회 FAIL 잔여분; 운영자 결정 `d-20261010T000832Z-efde` 분할 + 부채 원장 `.moai/worktrees/t1579/.moai/reports/t1579/debt-for-t1627.md`). 상위 SPEC의 6개 구조 검사 면(S1–S6)을 텍스트 패턴 검사에서 파서 기반 구조 검사로 설계한다. 근거 실측: 카드 워크트리 t1627 @ `79d0a0a47d`에서 하네스 3표면 앵커 전수 재검증 — 49·71·96행 일치, `### Phase 3` 165→167행·Named deliverable 314→316행 **드리프트 관측** (내용 앵커로 재고정 — 결정 D6), run.js 79/85/92행 일치·규범 문장 81–84행 걸침, manifest.json:3 일치, Phase 3 URL 목록이 코드펜스 안(184–192행)임을 확인. 상위 감사 근거 인용: iter3 §7 F1–F10·§8.2 정체 분류·§9.2 probe 재현(M1–M7), iter7 FAIL 0.69 (`rcpt-6a9b6d0789c19d1647b123d8`, audited_sha `80837ff8f`). RED 선관측(본 트리, 2026-10-13): `test -f internal/harness/relup_structure_audit_test.go` → 출력 없음·exit 1, 고정 앵커 5함수 선택자(spec §1.1 A11 명령)로 `go test` 실행 → `ok … [no tests to run]` 공집합 토큰. REQ 12건·AC 12건 (Tier M 천장 16 이내). SPEC ID `SPEC-RELUP-STRUCT-AUDIT-001` — 정규식 `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` 매치, 모든 중간 세그먼트 글자 선행 (RELUP·STRUCT·AUDIT) |
| 0.2.0 | 2026-10-13 | manager-spec | 수리 라운드 1 — codex 백엔드 수렴 감사의 유효 발견 2건 (Claude 측 PASS와 불일치, 둘 다 실재 확인). **P1**: S1 절차에 주석 정리 단계 신설 — Phase 3 블록 내용물에서 HTML 주석 스팬을 **먼저 제거**(펜스 밖 `<!--…-->`만, 펜스 내용물은 보존)한 뒤 목록 항목을 읽는다(REQ-RSA-002·plan §D1·AC-RSA-002) + 음성 대조 **M8**(주석 랩 URL 목록 → 0) 신설(spec §3·AC-RSA-002·008 — probe 보캬블러리 M1–M7 → M1–M8). **P2**: 중복 처리 계약을 두 클래스로 통일 — class E **exactly-once**(이름 붙은 경계 마커: S2의 Phase 3/Phase 4, 0회·2회 = 실패) / class F **first-match 선택**(선택 앵커: S5의 BP 헤딩, 첫 출현이 블록 선택·중복은 실패·재선택 아님) — REQ-RSA-001(d) 재작성 + 결정 D8 신설 + plan §D1·acceptance §D.2 반영. 상위 LED-008의 ≥1 의미론은 class F로 보존. lint 0 findings 재확인 |
| 0.3.0 | 2026-10-13 | manager-spec | 수리 라운드 2 — codex 재감사(v0.2.0 FAIL)의 유효 발견 2건. **P2-a**: plan §D2 D7 전파의 「상위 부채 원장(t1579 트리)에 기록」 지시 제거 — 그 트리는 본 카드의 편집 범위 밖이다. 재분류 결정과 근거는 본 카드의 자기 표면(`.moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/progress.md`·`.moai/reports/t1627/`·리더 완료 보고)에 기록하고, 원장 갱신이 필요하면 레인이 쓰지 않고 **리더가 기록**한다(완료 보고에 갱신 요청 정보를 실어 전달). **P2-b**: AC-RSA-012 판정면을 총계(`grep -c 'VERIFIED'` ≥6 — 보류 6행 삭제에도 통과하는 약점)에서 **행 단위**로 강화 — acceptance §D.1에 12개 행 앵커 패턴(id + 상태 토큰 + 클래스 코드: 인계 6행 `인계: VERIFIED`, 보류 유지 6행 `판정 보류 유지 — CX-<n>`)을 열거하고 각 관측 1·**어긋난 행 0**을 요구한다. 인계 표 상태 토큰의 볼드 제거(리터럴 정합), 총계 grep은 집계 연출로 강등. lint 0 findings 재확인 |

## 1. 문제 — 측정된 형태

상위 SPEC-RELUP-DUALAXIS-001의 여섯 구조 검사 면 S1–S6이 전부 **텍스트 패턴 검사**로 작성됐다 — 줄 단위 grep, `b.index("### Phase 3")` 부분 문자열 슬라이스, 고정 블록 계수. 감사자가 면마다 우회 probe를 하나씩 재현했고(plan-audit iter3 §9.2), 패턴을 고칠 때마다 새 우회가 생겼다. 상위 감사 iter3 §8.2의 정체 분류가 이 결함 클래스를 명명한다: *"F1, F2, F3, F7 and F8 are one instrument class — a text pattern standing in for a structural fact. The class has changed in each round but has not closed."* 상위 카드는 plan-audit 3회 FAIL(d36e97571 0.625 · 0cccc0d31 0.75 · 7fd68bbe2 0.75)과 iter7 FAIL 0.69(차단 5, `rcpt-6a9b6d0789c19d1647b123d8`, audited_sha `80837ff8f180ab9902d5bdad914415bd8ec2e43c`) 끝에 여섯 면을 판정 보류로 강등하고 구조 검사 설계를 본 카드로 분할했다(운영자 결정 `d-20261010T000832Z-efde`; 후속 초안 `.moai/worktrees/t1579/.moai/reports/t1579/followup-card-draft.md`).

**측정 트리**: `.moai/worktrees/t1627` (branch `WT-d-20261010t000832z-efde`) @ `79d0a0a47d`. 이 SPEC의 모든 앵커 관측과 RED 선관측은 이 트리의 2026-10-13 실측이다.

### 1.1 실측 요약 (트리 `79d0a0a47d`)

| # | 측정 대상 | 관측값 | 근거 |
|---|---|---|---|
| A1 | specialist Phase 0 헤딩 | 49행 `### Phase 0 — Load State` — 카드 초안 49행과 일치 | `sed -n '49p'` 본 트리 실측 |
| A2 | specialist Codex 축 상태 앵커 | 71행 `**Codex axis state — separate file (REQ-RDX-001)**` — 일치 | 동일 |
| A3 | specialist `### Phase 1` 헤딩 | 96행 — 일치 | 동일 |
| A4 | specialist `### Phase 3` 헤딩 | **167행** — 카드 초안의 165행에서 **+2 드리프트**. 줄 번호는 앵커가 아니므로 내용 앵커로 재고정한다(결정 D6) | `grep -n '### Phase'` 실측 — 헤딩 10건: 49·96·130·167·195·202·214·225·243·288행 |
| A5 | specialist Named deliverable | **316행** `**Named deliverable (REQ-RDX-014)**` — 초안 314행에서 +2 드리프트. 내용 앵커로 재고정 | `grep -n 'Named deliverable'` 실측 |
| A6 | run.js 주석·상수·닫는 괄호 | 79행 `// Standing 6-theme …` · 85행 `const CODEX_THEME_CHECKLIST = [` · 92행 `];` — 초안과 일치. **감시 규범 문장은 81–84행에 걸쳐 있다** — 초안의 "82–83행 문장"은 감사 당시 트리 기준이며 내용 앵커로 재고정한다(결정 D6) | run.js 79–92행 본 트리 실측 |
| A7 | manifest `domain` | 3행 — 일치. 값: `moai-adk-go dev-only maintainer tooling — Claude Code + Codex CLI upstream change tracking and the best-practices axis` (두 축 + BP 축을 명명) | manifest.json 본 트리 실측 |
| A8 | Phase 3 URL 목록의 위치 | **코드펜스 안**이다 — specialist 184행 개펜스 · 192행 닫는 펜스, 6종 URL 목록 항목은 186–191행. 블록 경계 검출용 마스킹과 내용 매칭용 제거를 구분하지 않으면 URL 목록이 사라진다(결정 D5) | 펜스 경계 grep 실측 — 펜스 4쌍: 60–69 · 80–89 · 184–192 · 253–263행 |
| A9 | specialist HTML 주석 | 현재 **0건** — M5·M6 probe 변이는 "검사가 제거해야 할" 합성 입력이다. 실측은 방어 대상이 없음이 아니라, 주석 제거 규칙이 오늘의 통과를 바꾸지 않음을 보장한다 | `grep '^<!--\|-->'` 0힛 |
| A10 | `### Phase 3`·`### Phase 4` 줄시작 출현 | **각각 정확히 1회**. 산문 언급(42행 `Phase 4`·108행 `Phase 3:`·302행 `Phase 3,`)은 줄시작 헤딩이 아니다 — exactly-once 규칙의 전제가 실트리에서 성립한다 | `grep -c '^### Phase 3'` = 1 · `grep -c '^### Phase 4'` = 1 |
| A11 | 검사기 부재 (RED 선관측) | `test -f internal/harness/relup_structure_audit_test.go` → (출력 없음) · **exit 1**. `go test -run '^(TestRelupStructurePhase3URLs\|TestRelupStructureCodexParagraph\|TestRelupStructureWatchNormSentence\|TestRelupStructureBPFaces\|TestRelupStructureManifestDomain)$' ./internal/harness/` → `ok github.com/modu-ai/moai-adk/internal/harness 0.584s [no tests to run]` · exit 0 — **공집합 스윕 토큰** (verification-completeness §1.1). 최초 관측은 접두 형태 선택자(`0.649s`)였고 동일 토큰을 출력했다 — 선택자 고정형으로 재관측한 값이 본 셀의 귀속값이다 | 본 트리 실측 2026-10-13 |
| A12 | LED-021 전제 (상위 제거면) | `grep -c "docs.anthropic.com" .claude/agents/harness/hns-release-update-specialist.md` → `0` · exit 1 — 구형 문장은 이미 제거됐다(상위 M3 항목 5, 착지 `cb54103ea`). LED-021은 본 카드 비소관이다(결정 D7) | 본 트리 실측 2026-10-13 |
| A13 | 패키지 존재 | `internal/harness` 패키지 존재 (`Package harness — frontmatter modification applier`) — 검사 테스트 파일의 착지 지점 | `head -3 internal/harness/applier.go` 실측 |

### 1.2 설계 결정 기록 (재논의 금지 — plan.md §D와 쌍을 이룬다)

| # | 결정 | 근거 |
|---|---|---|
| D1 | **구현 언어는 Go 테스트**다. 근거 3가지: (i) 이 리포는 Go 리포고 `go test`가 표준 기계 판정면이다 — 기존 `.claude/` 감사 테스트 전수가 Go 테스트다(`internal/template/commands_audit_test.go` 등 24개 `*_audit_test.go`). (ii) 판정 면이 CI 게이트에 얹힌다 — 배치 릴리스 경로(`origin/main` CI)가 `go test`를 실행하므로 별도 게이트 신설 없이 구조 검사가 기계 게이트가 된다. 파이썬 원샷·node 스크립트는 감사자 재실행은 가능해도 버전화된 게이트가 없다. (iii) 감사 대상이 JS/JSON/MD라도 검사기 언어가 대상 언어와 같아야 한다는 제약은 없다 — 검사기가 읽는 것은 텍스트와 JSON이다 | 리포 사실(§1.1 A11·A13) + 기존 감사 테스트 관례 + CI 게이트 무료 편승. 파이썬/node 후보는 (ii)에서 탈락 |
| D2 | **위치는 `internal/harness/relup_structure_audit_test.go` 단일 테스트 파일**, 패키지 `harness`, 공용 헬퍼는 같은 파일의 비공개(test 전용) 함수다. `internal/template`이 아닌 이유: 감사 대상 3표면은 dev-only 사용자 소유 하네스 자산이고 템플릿 미러가 없다(상위 §1.1 M9) — `internal/template`의 헌장은 배포 템플릿 트리다. `internal/harness`는 하네스 도메인 패키지로 주제상 정확한 집이다. 단일 파일인 이유: 6면이 하나의 헬퍼를 소비하므로 한 리뷰 단위가 맞고, 파일 쪼개기는 고립 이득 없이 파일 수만 늘린다(Enforce Simplicity). 헬퍼를 비테스트 파일로 두면 프로덕션 빌드에 죽은 코드가 실린다 — 테스트 파일 안에 둔다 | `internal/harness` 실측(A13) + template 헌장 + 단순성 사다리 |
| D3 | **probe 인코딩은 순수 인메모리 문자열 변이다.** 헬퍼 함수는 문자열 내용을 받는다(파일 경로가 아니라) — probe는 실트리 내용을 메모리에서 변형한 변이 문자열이고, 임시 파일도 서브프로세스도 없다. 각 probe 단위 테스트는 변이에 대해 0(또는 실패)을, 실제 파일 바이트에 대해 기대값을 단언한다. probe id(M1–M8)는 단위 테스트 이름에 싣는다 — `go test -v` 출력에서 음성 대조의 실행이 기계 판독된다(verification-completeness §1.1 — 스윕 집합 비가시성 봉쇄) | 워크트리 가드의 git 이름 복합 명령 거부(상위 acceptance §D.3-d Gaps)를 구조적으로 회피하는 가장 단순한 형태 — exec 자체가 없다 |
| D4 | **게이팅은 CI(go test)이고, 인계(hanback)는 면 단위로 결정한다.** 여섯 면(AC-RDX-009 육면 · AC-RDX-014 · AC-RDX-015 · AC-RDX-008 · AC-RDX-010 · AC-RDX-001 LED-001C 보존면)은 본 카드가 VERIFIED(CI 게이트)로 인계받는다. 강등 10종 목록(list E) 가운데 이 카드의 검사가 실제로 닮지 못하는 6종(AC-RDX-003·004·005·006 — CX-7/CX-8 내용 계측, AC-RDX-016 — CX-13 리터럴 반전, AC-RDX-017 — CX-12 paraphrase)은 **판정 보류를 유지**한다 — 강등 사유가 구조가 아니라 계측 깊이·의미론이므로 블록 파서가 그 클래스를 닫지 못한다. 인계 표는 acceptance §D.2가 SSOT다 | 상위 §D.2 게이팅 처분 표 + iter3 §8.2 권고 2(계측 깊이 부족 면은 이관) — 본 카드가 못 닫는 것을 닫는다고 쓰는 것을 금지한다(verification-claim-integrity §1) |
| D5 | **헬퍼는 이중 모드다**: (경계 모드) 코드펜스·HTML 주석 내용을 **마스킹**(행 위치 보존, 내용은 잃지 않음)해 줄시작 헤딩 검출이 펜스·주석 안의 가짜 헤딩을 세지 않게 한다 — 이 모드에서 S1은 Phase 3 블록 안 펜스 **내용물**(URL 목록 186–191행)을 그대로 판독한다. (매칭 모드) 토큰·문장 매칭 전에 주석·펜스를 **제거**한다(S5). 일괄 제거로 통일하면 A8에서 실측한 대로 S1의 판정 대상 URL 목록 자체가 사라진다 — 마스킹과 제거는 같은 함수의 두 출력이며 **스코프 파라미터**(주석/펜스 독립 선택)를 갖는다. 0.2.0 수리(P1): S1은 블록 내용물에서 **주석만 제거**(펜스 내용 보존)한 뒤 목록 항목을 읽는다 — 주석 랩 목록 mutant는 M8로 봉쇄된다 | §1.1 A8 실측 (URL 목록이 펜스 안) + M5/M6의 제거 요구 (iter3 F7) + P1 수리 |
| D6 | **앵커는 줄 번호가 아니라 내용이다.** 카드 초안의 165행·314행은 본 트리에서 167행·316행으로 드리프트했고(A4·A5), run.js 규범 문장도 82–83행에서 81–84행으로 걸침이 넓어졌다(A6). 검사 구현은 줄 번호를 상수로 갖지 않는다 — 헤딩 텍스트·앵커 접두사·문장 리터럴로 경계를 찾는다. 본 문서의 행 번호는 `79d0a0a47d` 시점 참고값으로만 기록한다 | A4–A6 드리프트 실측 — 검증-청구 이동 좌표 위험(verification-claim-integrity §2.1)의 예방 적용 |
| D7 | **LED-021(제거면)과 블로킹 4 재분류는 본 카드 소관이 아니다.** LED-021의 전제는 이미 소멸했다(A12 — `docs.anthropic.com` 0힛, 상위 M3 항목 5 착지 `cb54103ea`) — 본 카드는 LED-021을 자기 AC로 만들지 않고 상위 게이트에 남긴다. 상위 블로킹 4(001·002·007·009)의 유지·복귀 재분류는 초안 §4 4번 항목대로 **run/sync 단계의 미결 항목**이며 plan에서 결정하지 않는다(plan §D3 기록) | A12 실측 + 카드 지시(재분류는 이 카드의 감사 결과로 run/sync에서 재결정) |
| D8 | **중복 처리 계약은 두 클래스다** (0.2.0 수리, P2): **class E exactly-once** — 이름 붙은 경계 마커(S2의 `### Phase 3`·`### Phase 4`)는 0회·2회 출현 시 검사 **실패**(첫 출현 침묵 해석 금지). **class F first-match** — 선택 앵커(S5의 BP 헤딩)는 첫 출현으로 블록을 선택하고 끝은 다음 동급·상위 헤딩이며, 중복 출현은 실패도 재선택도 아니다. 모든 면은 자기 클래스를 REQ-RSA-001(d)에 명명한다 — 상위 LED-008의 ≥1 의미론(2+ BP 헤딩 문서에서도 성립)이 class F 선택으로 보존된다 | codex 수렴 감사 P2 — S1/S2 exactly-once와 S5 first-heading 선택의 규칙 충돌 해소; 상위 acceptance §D.3-c BP 블록 정의(첫 헤딩 ~ 다음 동급·상위 헤딩)와 정합 |

## 2. 원인 — 패턴이 구조를 대신했다

여섯 면의 공통 원인은 하나다: 마크다운 문서의 **구조적 사실**(이 텍스트는 어느 헤딩 블록에 속하는가, 이 행은 주석·펜스 안인가, 이 JSON은 어느 키의 값인가, 이 URL은 어디서 끝나는가)을 **텍스트 패턴**(줄 grep·부분 문자열·개수)으로 대신 판정한 것. 패턴은 구조의 한 투영일 뿐이라 모든 변이를 열거할 수 없고, 감사자가 변이를 하나씩 추가할 때마다 수리는 다른 투영을 고쳤다 — 클래스는 살아 남았다(iter3 §8.2). 구조적 사실은 구조적 판독기(헤딩 경계 파서·주석/펜스 마스커·JSON 디코더·URL 파서)로만 닫는다. 본 카드는 그 판독기를 공용 헬퍼 하나로 만들고 여섯 면을 그 위에 얹는다.

## 3. 이 카드가 봉쇄하는 mutant (M1–M8)

아래 probe들이 본 SPEC의 검사를 통과하면 검사가 너무 얕은 것이다(verification-completeness §2 mutant probe). 각 probe의 원 재현은 상위 plan-audit iter3 §9.2에 기록돼 있다(구 검사 통과 실측).

| probe | 면 | 재현 형태 (변이) | 구 검사 통과 실측 (iter3 §9.2) | 본 검사 요구 |
|---|---|---|---|---|
| M1 | S1 | 실제 URL을 외부 호스트의 `?next=` 값으로 감싼 목록 항목 | 육면 게이트 전부 rc 0 · out 1 (통과 — 결함) | 항목 단위 구조 파싱으로 scheme·host·path·종결 경계를 정확히 단언 — query·fragment·접미 존재 시 그 항목은 0 |
| M2 | S2 | Phase 2의 교차 참조가 `b.index("### Phase 3")` 슬라이스 시작점을 앞당김(+실 URL 제거) | 육면 게이트 전부 rc 0 · out 1 (통과 — 결함) | 줄시작 `### Phase 3`·`### Phase 4`만(펜스·주석 마스킹), 각각 정확히 1회 — 0회·2회는 오류. 실 URL이 블록 밖으로 빠지면 그 항목은 0 |
| M3 | S3 | 기본값 문장을 블록 끝 뒤 트레일링 단락으로 이동 / 블록 끝 뒤에 문장 덧붙임 | BLK-014G rc 0 · `[4, 1, 3, 2, 2, 1]` (통과 — 결함) | 코덱스 단락을 종결 빈 줄(또는 다음 헤딩)까지로 경계 — 기본값 문장(`rust-v0.161.0`+`emit warning`)이 경계 안에 있어야 하고 6토큰 각 ≥1은 경계 안에서만 계수 |
| M4 | S4 | 규범 문장을 빼고 `WATCH-LIST`·`stable release` 토큰만 주석으로 유지 | BLK-015G rc 0 · `[1, 1]` (통과 — 결함) | 블록 안에서 감시 규범 문장 **전체**(공백 정규화 리터럴)가 정확히 1회 일치해야 GREEN — 토큰만으로는 0 |
| M5 | S5 | deliverable(`HTML proposal report`)을 HTML 주석 안에만 배치 | BP-010G rc 0 · out 1 (통과 — 결함) | 주석·펜스 제거 후 매칭 — 주석 안 deliverable은 0 |
| M6 | S5 | 실제 BP 헤딩을 주석으로 감싸고 가짜 헤딩을 주석 안에 배치 | BP-008G rc 0 · out 1 (통과 — 결함) | 주석 안 헤딩은 세지 않는다 — 실헤딩 제거 변이에서 0 |
| M7 | S6 | 같은 줄 뒤쪽 키(`"note": "Claude Code"`)가 `"domain".*Claude Code` 줄 패턴을 통과시킴 | LED-001C rc 0 · out 1 (통과 — 결함) | manifest.json을 JSON으로 디코딩해 **최상위 `domain` 문자열**만 단언 — 다른 키가 절을 실어도 0 |
| M8 | S1 | Phase 3 블록 안 URL 목록 전체(또는 동일 목록 추가분)를 HTML 주석 안에 넣은 변이 — 주석 랩 목록 (0.2.0 수리 신설, P1) | (신설 probe — 구 텍스트 패턴 검사도 주석 안 목록을 세는 동일 클래스이며 iter3 §9.2 M5(주석 안 deliverable)와 같은 모드다; 본 검사 착지 전 실트리 재현은 run-phase의 `M8_` 단위 테스트가 한다) | 주석 스팬을 목록 열거 전에 제거한다(펜스 내용 보존) — 주석 안 항목은 0이고 실 펜스 목록만 계수된다; 실 목록째 주석으로 감긴 변이는 전 facet 0 |

## 4. 요구사항 (GEARS)

> 게이팅 상태와 인계(hanback) 처분의 SSOT는 acceptance.md §D.2다.

### 공용 헬퍼

- **REQ-RSA-001** (Ubiquitous) — The shared markdown structure helper shall provide four capabilities consumed by S1–S5: (a) HTML-comment and code-fence **masking/stripping primitives parameterized by scope** (comments and fences independently selectable) that preserve line positions for boundary detection — S1 masks comments while preserving fence content (P1), S5 strips both; (b) block boundaries derived from **line-start headings only** on unmasked lines; (c) whole-token and whole-sentence matching inside a block; (d) a **duplicate-handling contract with two explicit classes** — **class E exactly-once**: a NAMED boundary marker occurring zero times or more than once fails the check rather than resolving to a first occurrence (S2의 `### Phase 3`·`### Phase 4`); **class F first-match selection**: a selection anchor resolves to its first occurrence and duplicates neither fail nor re-select (S5의 BP 헤딩 — 첫 `best-practice` 헤딩이 블록을 선택, 끝은 다음 동급·상위 헤딩) — and every surface names its class. 결정 D5·D2·D8. **게이팅**: AC-RSA-001 (검사기 존재·전체 GREEN).
- **REQ-RSA-002** (Ubiquitous) — S1: the Phase 3 URL check shall FIRST clear the block content of HTML comment spans (주석 스팬 제거 — 펜스 **밖에서만** 주석으로 인정하고 펜스 내용물은 보존 — P1) and only then parse **each list item** of the Phase 3 block (fence content included — 경계 모드) as a structural URL and assert scheme `https`, host `code.claude.com`, path exactly `/docs/en/<facet>` (facets: `hooks`·`sub-agents`·`skills`·`plugins`·`mcp`·`settings`), empty query, empty fragment, and termination at end-of-item — each facet exactly once, counted from comment-cleared items only (M1·M8 봉쇄 — LED-022..027 면). 결정 D5. **게이팅**: AC-RSA-002.
- **REQ-RSA-003** (Ubiquitous) — S2: the Phase 3 block boundaries shall be the line-start headings `### Phase 3` and `### Phase 4`, each required to match **exactly once** outside fences and comments; a document carrying zero or two line-start occurrences shall fail the check (M2 봉쇄). **게이팅**: AC-RSA-003.
- **REQ-RSA-004** (Ubiquitous) — S3: the codex paragraph block shall be bounded by the line-start anchor `**Codex axis state` and its **terminating blank line** (or the next line-start heading, whichever comes first); inside that boundary the default sentence literals (`rust-v0.161.0` AND `emit warning`) and the four key-family tokens (`last_analyzed_version`·`last_analyzed_date`·`last_master_research`·`analysis_history[]`) shall each count ≥1, counted **inside the bounded block only** (M3 봉쇄 — BLK-014G 면). **게이팅**: AC-RSA-004.
- **REQ-RSA-005** (Ubiquitous) — S4: the watch-norm check shall require the **entire normative sentence** (pinned as a whitespace-normalized literal reconstructed from the run.js comment block: "Alpha-window themes remain WATCH-LIST observations — adoption is judged only when the theme lands in a stable release; an alpha-window theme is never reported as adopted drift (REQ-RDX-009).") to match **exactly once** inside the theme-checklist block bounded by the `// Standing 6-theme` line and its closing `];` line — token-only presence (`WATCH-LIST`·`stable release`) counts nothing (M4 봉쇄 — BLK-015G 면). 결정 D6. **게이팅**: AC-RSA-005.
- **REQ-RSA-006** (Ubiquitous) — S5: the BP faces shall be asserted on **comment- and fence-stripped** text (매칭 모드): (a) a case-insensitive line-start heading matching `best-practice` counts ≥1 (BP-008G 면), and (b) the deliverable literal `HTML proposal report` counts ≥1 inside the BP block (first best-practice heading .. next same-or-higher heading) on non-comment rows only (BP-010G 면) — the BP heading is **class F first-match selection** per REQ-RSA-001(d): the first matching heading selects the block, and duplicate BP headings neither fail nor re-select (상위 LED-008의 ≥1 의미론 보존 — P2). Comment- and fence-masked content counts zero in both faces (M5·M6 봉쇄). **게이팅**: AC-RSA-006.
- **REQ-RSA-007** (Ubiquitous) — S6: the manifest check shall **decode manifest.json as JSON** and assert the **top-level** `domain` value is a string carrying all three clause literals — `Claude Code`, `Codex CLI upstream change tracking`, `best-practices axis` — with no weight given to other keys (LED-001C 보존면; M7 봉쇄). **게이팅**: AC-RSA-007.

### 검사기 운반과 경계

- **REQ-RSA-008** (Unwanted) — The helper and every surface check shall make **no git invocation of any form** — no `os/exec`, no subprocess, no git-named command string; they read working-tree files directly (`../../.claude/…` 상대경로, 패키지 디렉터리 기준). 근거: 워크트리 세션 가드의 git-named 복합 명령 거부 (상위 acceptance §D.3-d Gaps). **게이팅**: AC-RSA-009.
- **REQ-RSA-009** (Event-driven) — **When** any probe variant (M1–M8, §3 table) is fed to its surface check, the check shall report zero matches / failure; **when** the real working-tree files are fed, the positive expectations hold — six URL facets each exactly 1, codex block six tokens each ≥1, watch-norm sentence exactly 1, BP heading ≥1 and deliverable ≥1, top-level `domain` three clauses present. **게이팅**: AC-RSA-002..007 (양성 대조 절).
- **REQ-RSA-010** (Ubiquitous) — The checks shall live in `internal/harness/relup_structure_audit_test.go` (package `harness`, test-only file — helper functions carry no production surface) and shall gate through the standard `go test ./internal/harness/` suite on the batch-release CI path. 결정 D1·D2·D4. **게이팅**: AC-RSA-001·008.
- **REQ-RSA-011** (Unwanted) — The checks shall not modify the three harness target files (판독 전용), shall not create a template mirror, and shall change no path outside: `internal/harness/relup_structure_audit_test.go`, `.moai/specs/SPEC-RELUP-STRUCT-AUDIT-001/`, SPEC `progress.md`, `.moai/reports/t1627/`. **게이팅**: AC-RSA-011.
- **REQ-RSA-012** (Event-driven) — **When** the acceptance artifact is authored, it shall carry the per-item handback table for the parent's deferral list E (003·004·005·006·008·010·014·015·016·017) plus the two non-list-E deferred faces (AC-RDX-009 육면, AC-RDX-001 LED-001C): six faces handed back VERIFIED (CI-gated), six items remaining verdict-deferred with the stated class reasons (CX-7/CX-8 내용 계측, CX-12 paraphrase, CX-13 리터럴 반전). 결정 D4. **게이팅**: AC-RSA-012.

## 5. 알려진 구속 조건

1. **헬퍼·검사는 git을 호출하지 않는다** (REQ-RSA-008). 상위 SPEC이 워크트리 가드의 거부를 Gaps로 기록한 바 있다(상위 acceptance §D.3-d) — 본 검사기는 exec 채널 자체가 없어야 하고, 파일 판독은 작업 트리 직접 읽기다.
2. **하네스 3표면은 판독 전용이다.** dev-only 사용자 소유 네임스페이스(상위 §1.1 M9) — 본 카드는 표면 본문을 고치지 않고, 템플릿 미러도 `make build`도 만들지 않는다.
3. **`internal/`에 심는 문자열 상한이 있다.** 상위 LED-011 회귀 가드가 `grep -rn "last-codex-version" internal/` 0힛을 요구한다 — 검사 테스트 파일은 리터럴 `last-codex-version`을 담지 않는다(S3의 6토큰·앵커·경로 어느 것도 이 문자열을 요구하지 않는다 — plan §D 제약). AC-RSA-010이 유지를 판정한다.
4. **앵커는 내용이고 줄 번호는 아니다** (결정 D6) — 검사 구현이 줄 번호 상수를 갖는 것은 금지다. 본 문서의 행 번호는 `79d0a0a47d` 참고값이다.
5. **Tier M 예산** — REQ 12건 / AC 12건 (천장 16 이내). 판정면은 `go test` 단일 스위트다.
6. **RED 귀속** — 본 SPEC의 RED-now 관측은 `79d0a0a47d`(카드 워크트리 t1627, branch `WT-d-20261010t000832z-efde`)에서 2026-10-13 수행했다. 문서 수준 핀이 이 절이며, 개별 AC 핀이 없는 한 이 핀이 구속한다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — 상위 카드 부채 B1–B5

- 상위 iter7의 차단 발견 B1–B5(`.moai/worktrees/t1579/.moai/reports/t1579/debt-for-t1627.md` — runner.js 폴백 창 상한·고정 창 목록 전진 없음·plan.md 인용 따옴표·개수 불일치·D9 인용 누락)는 부채 원장의 이월 항목이며 승인 id `d-20261010T064334Z-8df9`·`d-20261010T074921Z-bb5d`로 **재감사 차단 제외 대상**으로 이미 처분돼 있다. 본 카드의 plan-audit은 B1–B5를 본 카드의 결함으로 계수하지 않는다 — plan §D3에 감사자가 읽도록 명기한다.

### Out of Scope — LED-021 제거면과 블로킹 4 재분류

- LED-021(구형 `docs.anthropic.com` 제거면)은 상위 AC-RDX-009의 게이트며 전제가 이미 소멸했다(§1.1 A12). 본 카드는 LED-021을 자기 AC로 만들지 않는다. 상위 블로킹 4(001·002·007·009)의 유지·복귀 재분류는 run/sync 단계 미결 항목이다(결정 D7).

### Out of Scope — 내용 계측과 의미론 판별기

- 상위 강등 10종(list E) 가운데 본 카드가 닫지 않는 6종: AC-RDX-003·004·005·006(CX-7/CX-8 — 빈 상수·"Return ok." 프롬프트가 기계 면을 통과하는 내용 계측 클래스), AC-RDX-016(CX-13 — 리터럴 유지·규칙 역전 클래스), AC-RDX-017(CX-12 — 동의어 바꿔쓰기 클래스). 이들은 판정 보류를 유지하며 인계 표(acceptance §D.2)에 사유와 함께 기록한다(REQ-RSA-012).

### Out of Scope — 상위 SPEC 산물과 하네스 본문 수정

- 상위 SPEC-RELUP-DUALAXIS-001은 completed(debt-admit close, 0.11.1)다 — 상위 spec/plan/acceptance를 고치지 않는다. 하네스 3표면 본문도 고치지 않는다(구속 조건 2).

## 7. 미검증 항목 (Gaps)

- **실트리 probe 재생은 run-phase의 첫 테스트 실행이 한다** — 본 plan 단계에서 관측한 것은 검사기 부재(A11)뿐이다. M1–M8 변이에 대한 0 판정과 실트리 양성 값(BLK-014G 경계 축소 토큰 벡터의 실측값 포함)은 구현 테스트의 실행 출력으로 run-phase에서 관측·기록된다.
- **상위 블로킹 4 재분류 시점** — run/sync 단계 미결(결정 D7). 본 plan은 그 결정을 하지 않는다.
- **grep 수준의 산문 언급 관측(A10)은 텍스트 관측이다** — "산문 언급은 헤딩이 아니다"의 최종 판정은 구현된 헤딩 파서의 실제 동작으로 run-phase에서 관측된다.
- **S6의 세 절 리터럴은 오늘의 domain 값을 핀한다** — domain 문자열이 리포 정책 변화로 바뀌면 검사는 RED가 된다. 이는 회귀 가드로서의 의도된 동작이며(§5 조건 3과 같은 클래스), 절 리터럴의 변경은 상위 REQ-RDX-010의 decision D9·LED-001/002 게이트와 동시 갱신을 요구한다(plan §D2 전파).
