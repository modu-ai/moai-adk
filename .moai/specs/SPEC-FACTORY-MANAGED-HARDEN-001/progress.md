# progress.md — SPEC-FACTORY-MANAGED-HARDEN-001

> 단계별 진행 기록. `§E.1` 만 plan 단계(manager-spec)가 채우고, `§E.2`·`§E.3` 은 run 단계(manager-develop), `§E.4` 는 sync 단계(manager-docs)가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-after-audit-iter1 (재감사 대기)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 16) · plan.md · acceptance.md (AC 16) · design.md (D-1..D-3) · progress.md
tier: M
measured_tree: 7109e0900
open_clarifications: 0

### plan-audit iteration 1 — FAIL 0.75 (audited_sha 95dfd85c8)

판정서: `.moai/reports/t1409/plan-audit-iter1.md`(로컬 사본, gitignore 대상). 차단 D1–D5, 주요 D6–D7, 경미 D8–D10. 이 개정(0.2.0)이 그 처분이며 재감사가 필요하다.

| 결함 | 처분 | 개정 내용 |
|---|---|---|
| D1 RED 산출물 경로 불착지 | fixed | 산출물을 추적되는 `.moai/specs/SPEC-FACTORY-MANAGED-HARDEN-001/red-baseline.md` 로 이동(측정: `git check-ignore -v` 대 `.moai/reports/…` 는 `.gitignore:235` exit 0, SPEC 경로는 출력 없음 exit 1; 감사 캐시 해시 목록 밖). M1·AC-MH-013·DoD 2·REQ-MH-015 일관 수정. `.moai/reports/t1409/` 는 로컬 사본뿐 |
| D2 `Close`/`Start` 경합 | fixed | design D-3 에 수명주기 규칙(뮤텍스+closed, 기록은 열려 있을 때만, `Start` 자기 정리, `Close` 스냅숏 후 한 번 정리, `Close` 이후 `Start` 거부, 준비 대기를 세션 취소 함수에 연결). REQ-MH-012 개정. AC-MH-016 신설(두 소유자 6개 하위 케이스, `-race`, 자식·토큰 디렉터리 부재 단언). spec §F 에 t1410(F8)과의 겹침 경계 재설정 |
| D3 표 이스케이프 파이프로 시험 0개 선택 | fixed | 파이프 문자를 담은 명령은 표 밖 fenced 블록(원문)으로, AC id로 참조. 선택 수 재측정: 이스케이프 없는 형태 66, 표 이스케이프(백슬래시 파이프) 형태 0; AC-MH-004 패턴은 시험 부재라 `-list` 가 이름 0개. AC-MH-004·015 기대 관측에 선택 수와 `[no tests to run]`=실패 명시. plan.md §C·§E 도 같은 블록으로 맞춤 |
| D4 시험 기반 AC의 RED-now 부재 | fixed (옵션 b) | 시험 기반 릴리스 차단급 AC 를 "plan 시점 미채택, M1 이 채택"으로 명시하고 선결(D1)을 적음. M1 재현 시험을 8개로 확장(AC-MH-001 11개 하위·AC-MH-004 Close 부분 포함). RED-first 면제(쓰기 경합, `codex_connection_published`, AC-MH-009, AC-MH-008, AC-MH-007)는 이유와 대체 채택(변이 확인)을 명시 |
| D5 MP-6 D8 문면 | fixed | spec.md C.3 과 §D 에 "cross-platform exemption (EXCL-syscall)" 명시 선언: 신호 상수만, 시스템 호출 없음, AC-MH-011 이 Windows 빌드를 게이트 |
| D6 부모 REQ-MS-012 문면과 운영 문서 | fixed | spec.md C.3 에 부모 REQ-MS-012 문면에 대한 명시적 예외 선언(이유, 선례 `mcp_server.go:123`) 추가. AC-MH-012 가 운영 문서의 "관리 계층은 `syscall` 을 쓰지 않는다" 문장(기준 트리 62행) 정정과 `launch_signals.go` 언급을 grep으로 확인 |
| D7 elicitation 전제 오귀속·조용한 루프 | fixed | design D-1 의 "부모 쪽 관측"을 미관측 전제(부모 research.md `:87`, 부모 라이브 테스트 SKIP)로 정정. 답한 서버 요청마다 stderr 한 줄(elicitation 은 `serverName` 포함)을 REQ-MH-002 와 AC-MH-001 에. 거부된 `moai` broker elicitation 은 그 턴의 턴 단위 실패(REQ-MH-006, design D-1·D-2, AC-MH-006 10행). REQ-MH-014 한계 목록에 "거부된 elicitation 이 수신 확인을 막으면 TTL까지 재배달" 추가. 라이브 관측은 run 진입 조건 아님 — 이름 붙은 Gap(`MOAI_FACTORY_LIVE_ROOT`/`MOAI_FACTORY_LIVE_RUN`, 볼 것 3가지) |
| D8 AC-MH-012 양성 절반 | fixed | 남은 한계 고정 앵커 7개와 t1409 단락 부재를 `grep -c` 로(acceptance.md §1.2) |
| D9 REQ-MH-006/007 우선 턴 겹침 | fixed | REQ-MH-006 에 "after the priming turn" |
| D10(a) 호출부 수 | fixed | 정의 1 + 호출 9(출현 10곳, 5개 파일) 로 정정 |
| D10(b) 재현 시험 수 | fixed | 8개(F3 4·F4 1·F5 3)로 정정 |
| D10(c) 쓰기 데드라인·`id: null` | disclosed | design D-3 "공시한 한계": 데드라인 상수는 필요하다 판단하지 않아 더하지 않음(근거와 시그널 경로 영향 없음 명시); REQ-MH-014 한계 목록에 포함 |
| D10(d) 정리 중 두 번째 시그널 | disclosed | design D-3, acceptance.md §4(자동 시험 없음 — Gap), REQ-MH-014 |
| D10(e) 후손 프로세스 | disclosed | design D-3, acceptance.md §4, REQ-MH-014 |

개정 뒤 REQ 16(추가 0, 개정: REQ-MH-002·006·012·014·015), AC 16(신설 AC-MH-016, 개정: AC-MH-001·004·006·012·013·015).

미측정(개정 시점): 설치된 `moai` 바이너리가 이 트리보다 뒤처져 있으므로 `moai spec lint` 출력은 지연 빌드의 증거일 뿐 이 트리 빌드의 판정이 아니다. `Close`∥`Start` 경합과 elicitation 조용한 루프는 재현·라이브 관측하지 않았다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
