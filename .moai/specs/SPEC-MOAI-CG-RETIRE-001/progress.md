# SPEC-MOAI-CG-RETIRE-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

Tier L 0.1.0 초안. REQ 8 / AC 8. 코드·설치 man을 읽었으며 client/runtime는 실행하지 않았다.
iter1 SP-B1(FAIL0.81)의 명령·저장 상태·소비 경계 미결을 design의 확정 표와 AC fixture로 보강했다.
변경분 재감사는 미실행이며 실제 TEAMMATE 적용/launch 게이트는 여전히 미충족이다.

## §E.2 Run-phase Evidence

2026-09-11 Step 2 로컬 구현: raw YAML reader/guard, preview·claude-only 명시 적용, private backup·잠금·hash·atomic replace·readback, 초기 cc/glm/gpt 정책 소비를 작성했다.

| 범위 | Actual Output | Status |
|---|---|---|
| AC-CR-002/003 로컬 부분 | `go test ./internal/cli ./internal/config -run TestCG`: CLI 1.218s, config 1.308s; 신규 파일 coverage 90.0% / 97.9% | PARTIAL — 전체 AC 판정 보류 |
| AC-CR-007 로컬 mutant 부분 | guard·정책 누락·수락 누락·readback 성공오보 mutant 4개가 각각 exit 1 | PARTIAL |
| 경쟁·정적 검사 | race CLI 2.753s/config 1.516s; vet/gopls exit 0; Windows compile exit 0 | LOCAL PASS — Windows runtime 미실행 |
| 실제 TEAMMATE·전체 CG 제거·문서 | 미실행/미완료 | PENDING |

명령과 실제 출력, baseline, 남은 범위는 `.moai/reports/SPEC-MOAI-GATEWAY-001/cg-migration-verification.md`에 기록했다. hybrid capability는 사용자 값으로 우회할 수 없으며 적용·launch가 닫혀 있다. 원격 CI의 저장소 전체 verdict는 PENDING이다.

### Runtime 철거 로컬 후속

2026-09-11: cg root 등록과 apply 연결을 제거하고 내부 mode/spawn도 명시 retirement 오류로 닫았다. LegacyTeamModeCG는 데이터 식별만 유지하며 template provider 판정은 false다. hook 자동 credential 주입은 raw legacy/ambiguity guard로 막는다. 과거 kanban.Read backend=cg 문자열/원문 보존을 시험했다.

| 범위 | Actual Output | Status |
|---|---|---|
| CR-001/004 로컬 | retired root/mode/spawn, settings 원문 보존, 역사 record Read 시험 | LOCAL PASS — 전체 역사 display 및 설치 client 미검증 |
| CR-002/005 보강 | cc/glm/gpt 각각 정상 9개·legacy 9개 입력; legacy launch/spawn/worktree/credential-home=0, 정상 준비경계 27개 도달 | LOCAL PASS — 실제 tmux/provider 미검증 |
| CR-007 로컬 | 관련 CLI 5.966s; race 8.161s; hook race 1.805s; template0.539s/config0.420s | LOCAL PASS |
| runtime mutant | CG→Claude 별칭, CG→GLM backend 변형 각각 exit1 | LOCAL PASS |
| LSP | exit0, 기존 HEAD 동일 loadGLMKeyFromEnvFile 함수의 scanner.Err 경고1 | 기존 함수 경고 기록, 무수정 |
| CR-006/008 | README/template/4locale 및 실제 TEAMMATE | PENDING |

증거·명령·잔여 AC·문서 후속 후보는 `.moai/reports/SPEC-MOAI-GATEWAY-001/cg-retirement-runtime-verification.md`에 기록했다. SPEC 전체 완료/production gateway 활성화는 하지 않았다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
