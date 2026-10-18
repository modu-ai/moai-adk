# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- plan 산출물 작성 완료 (2026-10-18, manager-spec, card t1865): spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (decision gate on), 이 파일. 상태: `draft` (spec.md frontmatter).
- SPEC ID 사전 점검: `SPEC-UPDATE-OWNERSHIP-REPAIR-001` 정규식 검사 PASS (이번 실행, Bash 원문 출력: `PASS`). 카탈로그 중복 없음 (`ls .moai/specs | grep -c SPEC-UPDATE-OWNERSHIP-REPAIR-001` → `0`, 작성 전).
- 기준 트리: HEAD `cdd72a814` (브랜치 `WT-update-ownership-repair`), t1668 의 `e1e163320` 병합 포함. 기존 시험 11개 GREEN (research.md B).
- 열린 결정: decision-index.md Q2 (배포 버전 기록과 이전 매니페스트 보존, 제품 수준, 운영자 확인 대기 — 자율 Kickoff 차단, 이 수정판도 답하지 않음). Q5 (프로필 쪽 대응물 읽기) 는 리더 판정 2026-10-18 로 보수적 읽기가 확정되었다 (읽지 않음, `not-checked`, 프로필 지도 사본 무조건 보존; 라벨은 증거 대기 유지). Q3, Q4 는 기본값 적용 (2026-10-18T03:07:32Z).
- 다음 단계: plan-auditor 독립 감사 (Tier L 문턱 0.85), 그 뒤 Kickoff. run 단계 필수 독립 감사는 codex 게이트 (카드 완료 조건).

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-18
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (5, Tier L) + decision-index.md (decision gate on) + progress.md

### Plan audit iteration 1 repair (2026-10-18, manager-spec, card t1865)

- 입력: `.moai/reports/t1865/plan-audit-iter1.md` — FAIL, 점수 0.68 (Tier L 문턱 0.85), MP-8 실패, 차단 결함 5건, 필수 백엔드 codex 실패, 감사 대상 SHA `a8ad48c08`.
- 처리: 결함마다 이 트리 (`15db036a3`) 에서 자기 명령으로 확인하거나 반박한 뒤 고쳤다. 측정 원문은 acceptance.md § L (L-1 ~ L-10). 판정 도구는 이 트리에서 `go build -o /private/tmp/own-fixture-bin/moai ./cmd/moai` 로 만든 빌드이다.
- 수정판: spec.md `0.2.0` — 요구사항 16 (REQ-OWR-016 추가), 인수 기준 20 (AC-OWR-20 추가). plan_artifact_hash 는 이 문서가 계산하지 않는다 (lane 이 커밋 뒤 런타임 `ComputeHash` 로 다시 계산한다).

| 결함 | 처분 | 근거와 수정 위치 |
|------|------|-------------------|
| D1 AC-OWR-02 픽스처 경로가 템플릿에 없음 | 고침 (확인됨) | L-2: 종전 두 경로 모두 `git ls-files --error-unmatch` 종료 1. L-1: 실재 경로 `.claude/skills/moai-ref-owasp-checklist/SKILL.md`, `.agents/skills/moai-plan/SKILL.md`. AC-OWR-02 를 실재 경로 + 템플릿 조회 전제 단계 + 승격 변이 대조군(잘못된 `provenanceSource` 주입, design.md D-2) 으로 다시 썼다. AC-OWR-14 의 313개도 내장 원본에서 읽도록 명시했다 (L-5) |
| D2 AC-OWR-08 이 삼켜진 읽기·범위 밖 도우미를 못 잡음 | 고침 (확인됨, 한계 명시) | 0o000 방식 철회. 추가 발견: `claude-skills@` 는 Go 소스에 0건 (L-7) 이라 종전 문자열 스캔은 공허했다. 기준선 대비 패키지 전체 소스 스캔 (L-10 의 19개 파일) + 차분 카나리 + 카드 커밋 이력 기반 파일 불변 (L-6) 으로 교체. 접근 기록 시임은 채택하지 않음 (수용 부채, design.md D-8) |
| D3 RED-now 셀 없음 + 「게이트」 모순 (MP-8) | 고침 (확인됨, 선택지 a) | 판정 시험이 트리에 없어 RED 를 재실행할 수 없음. 20개 모두 regression-guard 로 분류하고 기준마다 승격 지점·승격 규칙 (시험 커밋 A → 셀 기록 커밋 B → 구현 커밋 C) 을 적었다. 선례 SPEC-AC-COLLECTOR-ANCHOR-001. 실측 가능한 전제 셀과 기준선 GREEN 셀 (L-1 ~ L-10) 만 원장에 넣고 RED 셀은 한 칸도 미리 채우지 않았다 |
| D4 AC-OWR-10 이 commands 에 `name` 요구 | 고침 (확인됨) | L-3a: commands 17개 모두 `name:` 없음, `description:` 은 모두 있음 (L-3a 둘째 명령). L-3b: agents 12개 모두 `name:` 있음. 기존 `commands_audit_test.go` 는 `description`·`allowed-tools` 를 요구한다. 조건을 commands/agents 로 분리하고 렌더 기준 비교를 명시했다 |
| D5 `manifest.prior.json` 쓰기 실패 뒤에도 제거 계속 | 고침 (확인됨) | design.md D-3 항목 3 을 fail-closed 로 바꾸고 REQ-OWR-016, AC-OWR-20 (root 에서도 성립하는 디렉터리 충돌 주입) 을 추가했다 |
| D6 Q2 「나」 낙하 목록 불완전 | 고침 (확인됨) | decision-index.md Q2 와 plan.md § G.1 에 접힘·축소·남음 전체 목록. 증거 해석기와 인터페이스를 M1b 에서 M1a 로 옮겨 AC-OWR-05 가 남는다. REQ-OWR-008, AC-OWR-13 (A·C), AC-OWR-18 (c), AC-OWR-17 변이, M5, M8 을 목록에 넣었고 운영자 가시 결과 (현장 증상 (5)(6) 일부 미이행) 를 적었다. Q2 는 열어 둔 채 |
| D7 AC-OWR-16 이 함수만 검증 | 고침 (확인됨) | `update.go:361` 의 dry-run 분기 전체를 기존 헬퍼 `runDryRunUpdateInProject` 로 구동. 디렉터리 mtime 을 트리 해시에 넣어 임시 쓰기를 잡는다 (design.md D-9). REQ-OWR-011 을 명령에 걸었다 |
| D8 `removal-eligible` 정의 불일치 | 고침 (확인됨) | spec.md § B 에 정의 한 곳, design.md D-1 과 AC-OWR-01 이 인용 |
| D9 M6 픽스처가 진단 원인을 재현 못 할 수 있음 | 고침 (확인됨, 한계 명시) | M6 과 AC-OWR-09 에 V1 같은 버전 / V2 낡은 기록 (rc.35 기록 + rc.36 바이너리 모양) 두 변형. 두 변형을 모두 돌린 뒤에만 「재현 안 됨」. V2 를 기존 시임으로 구동 못 하면 셀에 적는다 (갭 G-A) |
| D10 TMPDIR 디렉터리 생성 단계 없음 | 고침 (확인됨) | plan.md § C 의 P0 와 M0 첫 단계 `mkdir -p /private/tmp/own-fixture-d1`. acceptance.md 규칙 2번째 항이 P0 를 가리킨다 |
| D11 AC-OWR-06(b) 에 요구사항 문구 없음 | 고침 (확인됨) | REQ-OWR-015 에 매니페스트 수준 `Version` 갱신과 `DeployedAt` 변경 규칙을 넣었다 |
| D12 `initializer.go:607-608` 한 줄 어긋남 | 반박 (증거) | `sed -n 605,609p internal/core/project/initializer.go` 의 번호 매김으로 607 = `mf.Version = version.GetVersion()`, 608 = `mf.DeployedAt = …`. 이 트리에서 인용이 정확하다. 변경 없음 |
| D13 요구사항 문구의 구현 세부 | 일부 고침 + 수용 부채 | REQ-OWR-008 (코드 이름 제거), REQ-OWR-015 (필드 수준 서술 제거), AC-OWR-01 (상수 개수 단언을 모든 조합 매트릭스로) 고침. REQ-OWR-005 의 파일·디렉터리 이름은 리더 구속 제약 문구라 유지 (decision-index.md 수용 부채) |
| D14 AC-OWR-09 「또는」 결과 | 고침 (확인됨) | 상태마다 관측 결과를 하나로 고정 (S1 ~ S4) |
| D15 병합 충돌 노출 | 수용 부채 + 완화 | 새 판정 코드는 새 파일로, 두 파일의 변경은 호출 줄로 좁히고 M8 의 운영 코드 변경을 없앴다. 잔여 위험은 decision-index.md 수용 부채와 plan.md § D5 |

추가로 발견해 고친 것: (X1) plan.md M4 가 `probeCaseSensitiveFS` 재사용을 지시했으나 design.md D-4 는 모드 프로브 없는 별칭 확정을 정했고 그 함수는 비시험 호출자가 없다 (L-9) — M4 를 design.md 와 맞췄다. (X2) 위 D2 의 문자열 스캔 공허. (X3) 기존 런치 원장 판정이 프로필 설정 디렉터리 항목을 stat 한다 (L-8) — 이 SPEC 이 더하지 않는 기존 접근이며 plan.md OQ-5 로 리더에게 올렸다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
