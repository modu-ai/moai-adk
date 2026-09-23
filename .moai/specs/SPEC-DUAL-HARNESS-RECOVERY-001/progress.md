---
id: SPEC-DUAL-HARNESS-RECOVERY-001
document: progress
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
card: t1100
---

# Progress — SPEC-DUAL-HARNESS-RECOVERY-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC status: `draft`, version `0.2.0` (Tier L, 산출물 spec.md, plan.md, acceptance.md, design.md, research.md, progress.md).
- 카드·워크트리·브랜치: `t1100` / `.claude/worktrees/t1100-recovery` / `WT-dual-harness-recovery`. 초안 기준 HEAD `d87e9af2e`, 초안 SPEC 커밋 `09e24fd08`.
- REQ 25개(REQ-DHR-001 ~ 025), AC 23개(AC-DHR-001 ~ 023, 그중 LIVE 3개: 012, 018, 023).
- 설계 기준 매핑: AC-MIG-01 → AC-DHR-001~005, 021, 022 / AC-WT-01 → 006~009 / AC-AGENT-01 → 010~013, 023 / AC-MSG-01 → 014~016, 020 / AC-FACT-01 → 017~019.
- 미결 표식: 0건. 결정 1~4와 D6은 칸반 리드가 2026-09-23 운영자 상설 위임 아래에서 Jev 답을 근거로 내린 결정이며, 건별 운영자 확인은 아니다(`.moai/reports/t1100/operator-decisions.md`, `plan.md` §B). `CLAUDE.local.md` §29와의 긴장은 리드가 운영자에게 올린 미해결 항목이다.
- 독립 plan 감사: iter-1 FAIL 0.76(`.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-1.md`). 이번 개정은 iter-2 재감사 대상이다.
- 구현: 시작하지 않음. Implementation Kickoff Approval: 요청하지 않음(운영자 게이트).
- plan 단계 기준 측정(초안 시점, 이 트리):

```text
$ go test ./internal/codexwiring ./internal/factorymsg ./internal/template/agentemit -count=1
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.608s
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	3.044s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.384s
$ go test ./internal/cli -run '^TestFactoryLive(CodexCodex|CodexClaude|ClaudeCodex|ClaudeClaudeCompletionSeparation)$' -count=1 -v
--- SKIP x4, 패키지 결과 ok (SKIP은 PASS 아님)
```

## Revision iter-2

plan-audit iter-1의 결함 D1~D21을 이번 개정에서 어떻게 처리했는지 적는다. 각 결함의 근거 명령은 이번 실행에서 이 트리에 다시 돌렸다.

| 결함 | 처리 | 바뀐 곳 |
|---|---|---|
| D1 명확화 게이트 | 네 결정과 D6 결정을 반영하고 미결 표식을 모두 없앴다. 결정 출처(리드의 Jev 기반 결정, 건별 운영자 확인 아님)와 §29 긴장을 기록했다. 감사가 요구한 "AskUserQuestion 재확인"은 리드 지시에 따라 하지 않았다(기록 파일 참조) | plan.md §B, 전 파일 표식 제거 |
| D2 update 잠금 재사용 불가 | 재확인: `acquireUpdateLock`은 cli 비공개·O_EXCL(`update_cleanup.go:56-88`), cli가 codexwiring을 import(`update_codex_wiring.go:14`), `runUpdate`가 잠금을 쥔 채 갱신 호출(`update.go:275-279`, `:506`). codexwiring 소유 배선 잠금으로 교체, update 잠금 → 배선 잠금 순서만 허용, 저널 추가를 잠금 안으로 옮김, lock-held 결과를 보고. update 경로·단독 실행 AC 추가 | REQ-DHR-002, design A.2·A.3, AC-DHR-022 |
| D3 hooks.json 바이트 보존 불가 | 재확인: `hooks.go:122` `MarshalIndent`, `:198-212` `marshalEntry`. 형식별 보존 기준(config.toml 바이트, hooks.json 파싱 구조)과 기준점(unwire 직전 파일)을 정의. 부분 종류 `json-key`(description, `hooks.go:115-117`)와 MoAI가 만든 `[tui]` 테이블(`configtoml.go:145`)을 추가. 구분 빈 줄(`appendSection` `:221-229`)을 영역에 포함 | REQ-DHR-001, 005, design A.1·A.5, AC-DHR-003 |
| D4 소유 분류와 롤백 주장 | MoAI 식별 handler(접두사·네임스페이스)는 기존 교체 동작대로 `created`. 이전 배선 증거가 있는데 부분 기록이 없으면 `unknown`(승격 금지). "한 번 재배선하면 제거 가능" 주장을 handler에만 한정하고 구버전 config 부분은 제거 불가로 명시. manifest 반영을 저널 완료보다 앞에 두고 저널에 provenance를 실어 복구가 재적용 | REQ-DHR-001, 002, design A.1·A.3, plan §G |
| D5 결과 적용 판정 순서 | 8단계 판정 순서 표를 REQ에 적음(존재 → attempt → lane → lane 현재 generation → 기존 결과 → assignee generation → 상태 → 적용). 멱등 키를 (dispatch, attempt) 단위로 정의해 다른 lane 재할당이 키 충돌에 걸리지 않게 함. 이전 generation 재시도는 stale(t1082 AC-FLH-008과 같은 방향), 현재 generation 재시도는 duplicate. AC-DHR-014 (c)·(e), 015, 016을 표에 맞춤 | REQ-DHR-016, 018, design D.1·D.3, AC-DHR-014~016 |
| D6 감사 역할 sandbox | 리드 결정 (a) 반영: Codex에서 두 감사 역할 read-only, 부모가 반환문 그대로 판정 파일 기록, Codex 경로만의 예외. 사후 탐지 판정 삭제. 설계 §10을 차단 지시까지 정확히 인용(원문 177행). 변경을 REQ-DHR-015 + AC-DHR-013 + AC-DHR-023으로 한정 | REQ-DHR-015, design C.2·C.3, AC-DHR-013, 023 |
| D7 AC-DHR-018 명령 거부 | 조합별 리터럴 명령 네 개와 판정 jq 한 개로 바꿈. 새 판정식은 합성 입력으로 양·음 사례를 확인함 | AC-DHR-018 |
| D8 LIVE 양성 증거 | 증거 줄(`AC012_EVIDENCE`, `AC018_EVIDENCE`, `AC023_EVIDENCE`)에 nonce 반환·쓰기 시도 출력·양성 대조·호출 수 ≤ 예산·정리 pid를 요구. 합성 입력으로 예산 초과·양성 대조 없음·시도 증거 없음·증거 줄 없음이 모두 `false`임을 확인 | AC-DHR-012, 018, 023 |
| D9 설치 경로 symlink | 쓰기·unwire·복구 모두 Lstat 기반 거부. 대상 자체 symlink, 프로젝트 밖으로 해석되는 상위 디렉터리 symlink 사례를 AC에 추가 | REQ-DHR-006, design A.3, AC-DHR-001 |
| D10 중단 지점 | 저널 추가를 임시 파일보다 앞에 두고, P0~P6·D1·D2 아홉 지점과 사용자 수정 한 건을 표와 AC로 정의. 참조 없는 임시 파일 정리 규칙 추가 | REQ-DHR-004, design A.4, AC-DHR-002 |
| D11 REQ-DHR-020 대 t1082 방출 | 새 attempt 외에 같은 attempt 재부여 연산을 제공하고, 부를 조건은 호출자 SPEC(t1082)에 둠. 대화형 `/cd`에서 비생존 확인이 성립하지 않는 점을 경계 규칙에 적음 | REQ-DHR-020, spec §E, design D.2·D.5, AC-DHR-015 |
| D12 doctor가 파일을 바꿈 | doctor는 읽기 전용 보고만. 복구 진입점은 enable·disable·update 경로 | REQ-DHR-004, AC-DHR-021 |
| D13 템플릿 파일 삭제 범위 초과 | 프로필 전환 때 배포되지 않게 된 템플릿 파일은 보고만 하고 지우지 않음. 범위 밖 절에 명시 | REQ-DHR-007, spec §F, AC-DHR-005 |
| D14 AC-DHR-008 판정식·remove 경로 | 패키지별 테스트 이름과 `Package` 필드로 각각 `==1` 판정. `moai worktree remove` 포함. 추가 발견: `done`·`remove`는 레지스트리만 봐서 lock만 가진 Codex 트리를 anchor로 보지 않음(`done.go:86, 284`, `remove.go:51`) → REQ에 lock 판정 요구 | REQ-DHR-010, design B.3, AC-DHR-008 |
| D15 추적 빈틈 | REQ-DHR-004 진입점 AC(AC-DHR-021) 추가. REQ-DHR-015의 워크플로 의무는 D6 결정으로 "부모가 반환문 그대로 기록"으로 바뀌었고 AC-DHR-013(지시 표면 존재)과 AC-DHR-023(LIVE 원문 일치)이 추적 | AC-DHR-013, 021, 023 |
| D16 호스트 표현력 | 재확인: `agents-codex.yaml:196-207`(역할별 `[mcp_servers.moai]` 테이블 실측), 생성 TOML 7개에 테이블 존재. 축을 7개로 통일하고 서버 부여(`enforced`, measured)와 서버 안 도구 필터(`UNSUPPORTED`, documented)를 나눔. 각 축에 근거(measured/documented/unmeasured)를 붙이고 unmeasured는 enforced 금지. UNSUPPORTED 기대 집합을 고정 목록 대신 계약에서 계산 | REQ-DHR-013, design C.1, AC-DHR-010, 011 |
| D17 REQ-DHR-019 주체 | 주체를 "결과 메시지의 수신자"로 고침 | REQ-DHR-019 |
| D18 네임스페이스 handler | `.codex/hooks/moai/` 네임스페이스 handler를 MoAI 소유로 명시 | REQ-DHR-001, design A.1 |
| D19 트리 밖 SPEC 참조 | `related_specs`에 "미병합 draft, t1082 브랜치에만 있음" 주석, §E 첫 문단에 위치 명시 | spec.md frontmatter, §E |
| D20 t1082 문구 결정 | 경계 규칙을 제안으로 바꾸고 t1082의 현재 문구(REQ-FLH-009, design.md 204행)를 인용. 문구 수정 제안은 리드에게 전달할 사항으로 표시 | spec §E |
| D21 Windows·교체 경합 | Windows는 대기하는 부모 pid로 lock(잔여 위험 명시), 로컬 실측 범위 밖, CI 관측 전까지 Windows 분기 `NOT_RUN`. 교체는 트리별 O_EXCL 가드 안에서 사유 재확인 후에만, 실패는 모두 launch 거부 | REQ-DHR-008, design B.1, AC-DHR-006 |

반박하거나 고쳐 적은 점:

- D6 인용 위치: 설계 §10의 차단 문장은 `reports/moai-dual-harness-full-design-20260922.md` 177행이다(`grep -n` 결과 177: "역할별 권한을 호스트가 표현하지 못하면…"). 감사 보고는 176행으로 적었다. 내용 판정에는 영향이 없다.
- D16의 "역할별 부여는 있지만 전역 config에서 상속됨": 부여하지 않은 역할이 전역 서버를 물려받는다는 측정 기록을 이 트리에서 찾지 못했다(`agents-codex.yaml`에는 부여 쪽 측정만 있다). 사실로 받지 않고 `unmeasured`로 기록했다.
- D1의 "AskUserQuestion으로 확정" 요구: 리드가 재확인하지 말라고 지시했으므로 따르지 않았고, 그 사실과 한계를 plan.md §B에 남겼다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
