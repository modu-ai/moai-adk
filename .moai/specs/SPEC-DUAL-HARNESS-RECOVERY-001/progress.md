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

- SPEC status: `draft`, version `0.3.0` (Tier L, 산출물 spec.md, plan.md, acceptance.md, design.md, research.md, progress.md).
- 카드·워크트리·브랜치: `t1100` / `.claude/worktrees/t1100-recovery` / `WT-dual-harness-recovery`. 초안 기준 HEAD `d87e9af2e`, 초안 SPEC 커밋 `09e24fd08`.
- REQ 25개(REQ-DHR-001 ~ 025), AC 23개(AC-DHR-001 ~ 023, 그중 LIVE 3개: 012, 018, 023).
- 설계 기준 매핑: AC-MIG-01 → AC-DHR-001~005, 021, 022 / AC-WT-01 → 006~009 / AC-AGENT-01 → 010~013, 023 / AC-MSG-01 → 014~016, 020 / AC-FACT-01 → 017~019.
- 미결 표식: 0건. 결정 1~4와 D6은 칸반 리드가 2026-09-23 운영자 상설 위임 아래에서 Jev 답을 근거로 내린 결정이며, 건별 운영자 확인은 아니다(`.moai/reports/t1100/operator-decisions.md`, `plan.md` §B). `CLAUDE.local.md` §29와의 긴장은 리드가 운영자에게 올린 미해결 항목이다.
- 독립 plan 감사: iter-1 FAIL 0.76(`review-1.md`), iter-2 FAIL 0.81(`review-2.md`). 이번 개정은 iter-3(마지막 회차) 재감사 대상이다.
- REQ·AC 연속성: `moai spec lint`는 REQ 번호 빈칸과 존재하지 않는 REQ를 가리키는 AC·표 행을 보고하지 않는다(plan-audit iter-2 변이 m1·m2). 그래서 연속성의 근거는 `spec.md` §D 표와 `acceptance.md` §A 매핑 표의 수동 대조다. 이번 개정 뒤 두 표의 (REQ, AC) 쌍을 다시 대조했다(아래 Revision iter-3).
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

- plan_complete_at: 2026-09-23T01:51:02Z
- plan_status: audit-ready
- plan-audit iter-4(delta) 결과 PASS-WITH-DEBT 0.93(`.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-4-delta.md`). 선택 채무: D1(plan.md:132 "도달"→"초과" 문구), D2(operator-decisions.md가 gitignore 대상), 이월 N2·N5·N6.

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

## Revision iter-3

plan-audit iter-2(`.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-2.md`, FAIL 0.81)의 ND1~ND14를 다음과 같이 처리했다. 새 범위는 넣지 않았다. REQ 25개·AC 23개는 그대로다.

| 결함 | 처리 | 바뀐 곳 |
|---|---|---|
| ND1 `done`의 L1 거부 | 코드 재확인: `done.go:76-81`·`:277` L1 거부(`--force` 무관), `codex_launcher.go:411`·`session_worktree.go:47` Codex 트리는 L1, `:356-378` 절대 경로는 기존 L2 진입만. REQ-DHR-010을 "`done`은 L1 거부 유지(SPEC-WORKTREE-DONE-TIER-001, 뒤집지 않음)" + "지울 수 있는 경로(`clean --stale`, `remove`, PR-merge 정리)가 lock-aware anchor 판정을 쓴다"로 고침. 세션 종료 정리는 Codex launch가 부르지 않으므로(`init.go:504`, `web.go:115`만) 대상에서 뺌. PR-merge 정리는 `WT-*` 트리 전부를 지울 수 있어(`session_worktree_prmerge.go:170-173`, `:217`) 넣음. AC-DHR-008 재작성(`done` 두 사례 포함, 옛 `--apply`는 실제 플래그 `--yes`로 정정) | spec.md REQ-DHR-010·frontmatter related_specs, design.md B.3, acceptance.md AC-DHR-008, plan.md M5·§G |
| ND2 재부여 뒤 같은 키 재전송 | 리드 조정 결정(2026-09-23)을 §E에 기록: t1082가 계약을 좁힘(재전송은 새 키, 같은 키 `duplicate`는 같은 generation 안). 현행 `UNIQUE(sender_session, idem_key)`와 충돌 없음. 이 경우 lane slot 범위는 후속 후보(§F). 운영자 결정 3의 조건부 분기와 별개이며 그대로 유효. REQ-DHR-017의 "different recipient"를 수신 세션 UUID 또는 수신 generation으로 정의(현행 `store.go:624-626`). AC-DHR-015에 재부여 뒤 같은 키 재전송 거부·새 키 저장 확인 추가 | spec.md REQ-DHR-017·§D·§E·§F, design.md D.1·§E, acceptance.md AC-DHR-015·매핑 표 |
| ND3 test2json 1 KiB 분할 | 증거 채널을 파일로 바꿈: 테스트가 `MOAI_T1100_EVIDENCE_DIR`(리터럴 `../../.moai/reports/t1100`)에 증거 JSON을 쓰고 짧은 `<TAG>_SHA256` 줄만 찍음. 판정은 `shasum`으로 파일 해시를 재어 태그와 같을 때만 내용 판정. AC-DHR-012, 018, 020, 023 전부 적용. AC-DHR-019에 채널 자체의 양·음 사례 명령 추가 | acceptance.md §A, AC-DHR-012·018·019·020·023 |
| ND4 AC-012 변이 통과 | 역할 이름 집합 = 생성된 TOML 이름 집합(판정 명령이 `find`로 만듦), nonce 비지 않음·일치·12개 서로 다름, 양성 대조 sha256 64자, 감사 역할 두 개가 각각 한 번, 호출 수 정확히 14. 변이 목록을 AC에 적음 | acceptance.md AC-DHR-012 |
| ND5 AC-020 공허 측정 | 경로를 지났다는 조건(세션 UUID 둘이 다름, generation 증가, 첫 전송 ID, 대조 메시지 claim 1, 행·claim 수 ≥ 1)을 먼저 요구. 결과 판정: `reproduced` ⇔ 2/2/`new-id`, `not-reproduced` ⇔ 1/1/`same-id`. 경로를 지나지 않은 실행은 `NOT_RUN`이며 분기를 고르지 않음. REQ-DHR-025에 같은 조건과 기록 항목을 적음. 분기 로직(재현 → lane slot 이관, 아니면 유지)은 바꾸지 않음. AC-DHR-014는 AC-DHR-020이 `true`가 아니면 `UNPROVEN` | spec.md REQ-DHR-025, acceptance.md AC-DHR-014·020 |
| ND6 프로필 전환 전제 | 코드 재확인(`update_template_sync.go:405-424`, `deploy.go:56-85`, `harness_fs.go:112`). design A.6 전제 정정(코드 판독·미측정으로 표시). REQ-DHR-007·AC-DHR-005를 `.codex/` 쪽과 배선 파일로 한정하고 `.claude/` 관리 뿌리를 단언하지 않음. Claude 쪽 삭제는 범위 밖 발견 + 후속 카드 후보 | spec.md REQ-DHR-007·§F, design.md A.6, acceptance.md AC-DHR-005, plan.md M4·§G, research.md |
| ND7 변이 설명 | "두 번 적용"을 "(a)는 (7) `invalid-state`, (c)는 (6) `stale`이 되어 FAIL"로 고침 | acceptance.md AC-DHR-014 |
| ND8 EOF 줄바꿈 | 배경 문장과 A.1 `region`에 원본이 줄바꿈으로 끝나지 않을 때 더해지는 한 바이트가 영역 밖이며 unwire가 되돌리지 않음을 적음(`configtoml.go:225-227`, `:130-141`) | spec.md §A, design.md A.1 |
| ND9 `mcp-server` 이중 매핑 | REQ-DHR-013에 "제한 종류가 둘 이상인 축은 종류별로 하나씩 매핑"을 적음. C.1 매핑 칸에 같은 문구 | spec.md REQ-DHR-013, design.md C.1 |
| ND10 순서 범위 | 순서 보존을 배열 원소(이벤트별 entry, entry별 handler)로 한정하고 객체 키 순서는 비교하지 않는다고 적음 | spec.md REQ-DHR-005, design.md A.5 |
| ND11 `diverged` 임시 파일 | `diverged` 항목이 참조하는 임시 파일을 지운다고 REQ-DHR-004와 A.4 표에 적음. AC-DHR-002에 해당 음성 사례 추가 | spec.md REQ-DHR-004, design.md A.4, acceptance.md AC-DHR-002 |
| ND12 `sandbox` 근거 | "필드 수용: measured / 쓰기 강제: AC-DHR-012 전까지 미측정"으로 나눔. REQ-DHR-013에 `enforced`의 근거는 필드에 대한 관측이고 런타임 차단 주장은 REQ-DHR-014 LIVE 증거에만 기댄다고 적음 | spec.md REQ-DHR-013, design.md C.1 |
| ND13 C1/C2 표기 | C2(`internal/template/templates/.claude/agents/moai/`)와 C1(`.claude/agents/moai/`) 경로를 모두 적음 | acceptance.md AC-DHR-013 |
| ND14 `collision` 도달 경로 | (f)의 구성을 적음: 메시지 층 거부를 먼저 확인한 뒤 결과 적용 연산을 직접 호출 | acceptance.md AC-DHR-014 |

이번 실행에서 확인한 것:

- ND3 분할 재현: scratch 모듈의 테스트가 4552바이트 증거 파일을 쓰고 같은 내용을 한 줄로도 찍음. `go test -json` 출력에서 긴 줄은 `len=1024` 네 개와 `len=472` 하나로 쪼개졌고, 옛 판정식은 `jq` exit 5. 새 판정식(acceptance.md의 판정 명령 그대로)은 같은 실행에서 `true`.
- 새 판정 명령을 acceptance.md에서 그대로 뽑아 합성 입력에 실행한 결과: AC-DHR-012 실제 실행 `true`, 변이 4종 `false`, 증거 파일 부재 시 판정식 미도달. AC-DHR-018 정상 `true`, 실행 뒤 수정된 증거 `false`. AC-DHR-020 `reproduced`·`not-reproduced` 정상 `true`, 행 0·claim 0 `false`, 두 번째 전송 오류 `false`. AC-DHR-023 정상 `true`, 해시 불일치 `false`. AC-DHR-019 두 명령 모두 `true`.
- REQ·AC 연속성: `grep -c '^### REQ-'` 25, `grep -c '^### AC-'` 23. spec.md §D와 acceptance.md 매핑 표의 (REQ, AC) 쌍을 뽑아 비교해 일치함을 확인했다(REQ-DHR-017 ↔ AC-DHR-015 쌍을 양쪽에 추가).

## §E.2 Run-phase Evidence

### M1 — AC-DHR-020 멱등 범위 재현 측정 (REQ-DHR-025)

증거 디렉터리 `.moai/reports/t1100/`는 gitignore 대상이므로 아래 판정 근거를 값으로 적는다(`acceptance.md` §D 기록 의무).

- 상태 전이: `spec.md` `status: draft → in-progress` (이 M1 커밋). 나머지 산출물은 frontmatter에 `status` 필드가 없다.
- 측정 트리: `ac020-head.txt` = `9ba43c05f`. 그 트리의 factorymsg 스키마는 바뀌지 않았고, 추가된 것은 새 테스트 파일 `internal/factorymsg/idem_scope_repro_test.go` 하나뿐이다(프로덕션 코드·스키마 변경 없음). 테스트 파일은 측정 시점에 미커밋이었고 이 M1 커밋으로 들어간다.
- 실행: acceptance.md AC-DHR-020의 실행·판정 명령을 그대로 실행했다. 실행 exit 0, 판정 출력 `true`, 판정 exit 0.
- 결과(`outcome`): **`reproduced`**
- 증거 파일 sha256: `9ff051e039834a34833a22838dbec45c28a8b0c3abeebcceecc954b36b7685f8` (`ac020-evidence.sha`와 jsonl 태그 줄 `IDEM_SCOPE_REPRO_SHA256 9ff051e0…685f8`이 일치)
- 저장소가 연 unique 제약(`sqlite_master`에서 읽음): `UNIQUE(sender_session,idem_key)`
- 수치: `sender_sessions` = [`t1100-sender-session-1`, `t1100-sender-session-2`], `sender_generations` = [1, 2], `first_send_id` = `27ab1b87110c5e1e25b21d654093b653`, `second_send_result` = `new-id`, `rows` = 2, `claimed_ids` = 2, `control_claimed` = 1
- 판정식 음성 변이(저장소 밖 합성 입력, AC 판정식 그대로): 정상 `reproduced`(2/2/`new-id`) `true`, 정상 `not-reproduced`(1/1/`same-id`) `true`. `not-reproduced`에 행 0·claim 0, 두 번째 전송 `error`, 같은 세션 두 번, 대조 claim 0, `reproduced`인데 행 1, generation 미증가(2→2), 출력에 `NOT_RUN` — 7건 모두 `false`.
- 선택된 REQ-DHR-017 분기: **A** (송신 lane slot 기준으로 멱등 범위를 옮기고 기존 행을 손실 없이 이관). M2가 이 분기로 스키마를 정한다. AC-DHR-014는 분기 A 명령만 판정하고 분기 B는 `N/A (branch)`로 적는다.
- 리드 조정 필요: 결과가 `reproduced`이므로 t1082 조정이 필요하다(`plan.md` §B-3, `spec.md` §E 항목 2). t1082 REQ-FLH-009의 "current t1074 schema's idempotency key unchanged" 문구가 분기 A와 맞지 않는다.
- 측정 방식 메모: 재시작은 같은 slot `lane-1`에 새 세션 UUID와 새 process_start로 재등록하고 이전 소유 프로세스를 죽은 것으로 두는 방식으로 모사했다(`ownerCurrent` 테스트 대체). auto-assign 센티넬이 아니라 명시 slot을 써서 로컬 develop의 센티넬 개명(t1085)과 무관하게 한다.
- `MOAI_T1100_EVIDENCE_DIR`가 비어 있을 때: 측정과 경로 통과 검사는 그대로 돌고(통과 못 하면 `NOT_RUN`을 찍고 실패), 증거 파일과 태그 줄만 만들지 않는다. 그래서 일반 `go test ./internal/factorymsg`는 초록으로 남는다. 이는 `acceptance.md` §A의 "값이 비어 있으면 테스트는 `NOT_RUN`을 찍고 실패한다"와 다르며, 리드 지시(CI 초록 유지)를 따른 것이다 — 리드 판정 대상.

품질 게이트(측정 트리 + 새 테스트 파일):

```text
$ go test ./internal/factorymsg -count=1
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	3.297s
$ go vet ./internal/factorymsg
(출력 없음, exit 0)
$ golangci-lint run ./internal/factorymsg/...
10 issues: errcheck 10 — store.go 8, launch_pending_rollback_test.go 2 (새 파일 0건, 기존 파일의 기존 지적)
```

### M2 — dispatch record, 결과 적용 판정 순서, 분기 A 범위 이관 (REQ-DHR-016 ~ 021)

증거 디렉터리는 gitignore 대상이므로 판정 근거를 값으로 적는다.

커밋(순서가 곧 작업 순서다. 특성 테스트 커밋이 이관 커밋보다 앞선다 — `verification-claim-integrity.md` §2.3):

| 커밋 | 내용 |
|---|---|
| `2e6389bf7` | `dispatches` 테이블, 판정 순서 표(design §D.3), 재할당, 같은 attempt 재부여, superseded status, `ErrStalePeer` |
| `d070e8ea5` | 이관 전 특성 테스트 `TestLegacyBrokerRowsSurviveOpen` (구 스키마 DB의 행·ID·열 값 보존) |
| `60de6bd44` | 분기 A: 멱등 범위를 `(project_key, run_id, sender_slot, idem_key)`로 이관, 기존 DB 마이그레이션 |
| `158c91289` | `TestIdemScopeRestartReproduction`의 빈 `MOAI_T1100_EVIDENCE_DIR` 처리를 `NOT_RUN` skip으로 변경(리드 결정) |

AC 판정(명령은 `acceptance.md`의 것을 그대로 실행, HEAD `158c91289`):

| AC | 명령 | 판정 출력 | 증거 파일 | pass 이벤트 / fail·skip |
|---|---|---|---|---|
| AC-DHR-014 공통 | `^TestDispatchResultExactlyOnce$` | `true` | `.moai/reports/t1100/ac014.jsonl` | 8 / 0 |
| AC-DHR-014 분기 A | `^(TestDispatchResultExactlyOnce\|TestIdemScopeLaneMigration)$` | `true` | `.moai/reports/t1100/ac014-branch.jsonl` | 11 / 0 |
| AC-DHR-014 분기 B | — | `N/A (branch)` | — | AC-DHR-020 = `reproduced` |
| AC-DHR-015 | `^TestDispatchStaleFencingAndSuperseded$` | `true` | `.moai/reports/t1100/ac015.jsonl` | 1 / 0 |
| AC-DHR-016 | `^TestDispatchReassignmentFencesLateResult$` | `true` | `.moai/reports/t1100/ac016.jsonl` | 1 / 0 |

AC-DHR-014 변이(작업 트리에서 임시 수정 후 원복, 원복은 파일 sha1 일치로 확인):

- 판정 순서 5 삭제: `duplicate_delivery` → `invalid-state`, `sender_restart_after_result` → `stale`(AC가 예측한 값), 6개 공통 하위 테스트 모두 FAIL.
- 4·5 순서 교환: `old_generation_redelivery` → `duplicate`(`want "stale"`)로 FAIL.
- 분기 A 범위를 송신 세션으로 되돌림(스키마 UNIQUE와 충돌 재조회 둘 다): `sender_restart_lane_scope` → `rows=2`로 FAIL.

마이그레이션 메모:

- 스키마 차이: 구 `UNIQUE(sender_session,idem_key)` → 신 `sender_slot TEXT NOT NULL DEFAULT ''` 열 추가 + `UNIQUE(project_key,run_id,sender_slot,idem_key)`. `SchemaVersion` 1 → 2.
- 이관은 `Open`과 `OpenExistingWithDeadline` 둘 다에서 돈다(훅 경로가 기존 run을 여는 경우 대비). 이미 이관된 DB에서는 카탈로그 조회 한 번만 하고 쓰지 않는다.
- 이관은 immediate 트랜잭션 하나에서 새 테이블 생성 → 복사 → 행 수 대조 → 구 테이블 삭제 → 이름 변경 → 수신 인덱스 재생성 순이다. 트랜잭션 안에서 열 존재를 다시 확인하므로 동시에 여는 두 프로세스 중 하나만 이관한다.
- `sender_slot` 매핑: 송신 세션이 현재 `peers`에 있으면 그 slot, 없으면 `legacy:<session>`. 한 실제 slot에 매핑되는 행은 모두 그 slot의 단일 현재 세션에서 왔으므로 구 per-session 유일성이 충돌 없음을 보장한다.
- 특성 테스트의 구 스키마 상수 `legacySchemaV1`이 이관 직전 트리(`2e6389bf7`)의 `schema` 상수와 바이트 동일함을 일회성 테스트로 확인했다(`legacySchemaV1 == schema (1192 bytes)`, 커밋하지 않음).
- `TestIdemScopeLaneMigration`은 `open`·`open_existing` 두 진입점에서 행 보존·ID 보존·slot 매핑·같은 범위 재시도 → 원래 ID·수신자/수신 generation/kind/task_ref/correlation/payload 차이 거부(행 불변)·재오픈 멱등을 확인한다. `TestMessagesLaneScopeDDLMatchesSchema`가 이관 대상 DDL과 새 DB의 DDL이 갈라지지 않게 잡는다.

SPEC 재량 안의 결정:

- 결과 메시지 종류는 기존 `status_report`를 쓴다. 새 kind를 만들지 않았다.
- `ApplyResult` 판정 4에서 lane의 `peers` 행이 없으면(현재 generation을 알 수 없음) `stale`로 닫는다.
- 재할당의 "이전 소유자 비생존 확인"은 이전 assignee lane의 현재 `peers` 행 소유 프로세스가 살아 있지 않은 것으로 판정한다(같은 프로세스가 새 generation을 가진 경우도 살아 있는 것으로 본다 — 보수적).
- 재부여(`RegrantDispatch`)는 같은 lane의 더 높은 generation으로만 허용하고, 호출자 상태 변경을 콜백으로 받아 같은 트랜잭션에서 실행한다(콜백 오류 시 재부여도 롤백 — AC-DHR-015 테스트가 확인).
- 메시지 receipt는 dispatch를 전혀 움직이지 않는다. `delivered`·`started`는 assignee의 명시 호출로만 오른다.
- superseded 판정: pending·claimed 메시지의 수신 (session, generation)이 현재 `peers` 어느 행과도 맞지 않으면 superseded. `Send`는 현재 endpoint에만 보내므로 불일치는 곧 lane 재등록을 뜻한다. 기존 `TestFactorySessionGenerationOwnership`의 기대값(`Pending == 1`)을 REQ-DHR-020에 맞춰 `Superseded == 2, Pending == 0, Claimed == 0`으로 바꿨다(재등록 전 pending 1건 + 재등록 전 claim된 1건).
- 새 `sender_slot` 열에 `DEFAULT ''`를 뒀다. `Send`는 항상 검증된 slot을 넣고, 기본값은 열을 모르는 쓰기(구 바이너리, `internal/hook` 테스트의 raw INSERT 픽스처)를 깨지 않기 위한 것이다.
- `Envelope`/`Claim`에 `SenderSlot`을 더했다(수신자가 재시작한 송신자의 결과 보고를 만들 때 필요). MCP 도구 JSON에는 필드가 하나 늘 뿐이다.
- `internal/cli/mcp_factory_msg.go`, `internal/mcp/catalog.go`는 고치지 않았다. AC-DHR-014 ~ 016 중 도구 표면을 요구하는 것이 없고, `factory_msg_status`는 `Status`를 그대로 직렬화하므로 `Superseded`가 자동으로 실린다.

`TestIdemScopeRestartReproduction`(M1 측정 테스트) 처리 — 리드 결정(skip 형태):

- 빈 `MOAI_T1100_EVIDENCE_DIR`: `t.Skipf("NOT_RUN MOAI_T1100_EVIDENCE_DIR unset: …")`로 끝난다. fail 형태와의 차이는 그 한 줄(`Skipf` ↔ `Fatalf`)뿐이다. `acceptance.md` §A:22 문구 수정은 manager-spec 몫으로 남긴다(이 레인은 고치지 않았다).
- 변이 확인: AC-DHR-020 실행 명령을 환경 변수 없이 저장소 밖 scratch jsonl로 실행하고, 보존된 증거 파일의 사본과 함께 AC-DHR-020 판정식을 그대로 돌렸다. 출력 `false`(exit 1). jsonl의 해당 줄: `idem_scope_repro_test.go:79: NOT_RUN MOAI_T1100_EVIDENCE_DIR unset: the measurement writes its evidence only through the acceptance command`, `--- SKIP: TestIdemScopeRestartReproduction (0.00s)`.
- 분기 A 이후 측정(환경 변수를 scratch로 지정): `outcome` = `not-reproduced`, `unique_constraint` = `UNIQUE(project_key,run_id,sender_slot,idem_key)`, `second_send_result` = `same-id`, `rows` = 1, `claimed_ids` = 1, `control_claimed` = 1, 테스트 PASS. 측정 로직은 바꾸지 않았다. 이 결과는 AC-DHR-020의 M1 측정을 대신하지 않는다.
- M1 보존 증거(`.moai/reports/t1100/ac020*`)는 건드리지 않았다. `ac020-evidence.json` sha256 = `9ff051e0…685f8`(M1 기록과 같음).

품질 게이트(HEAD `158c91289`):

```text
$ go test ./internal/factorymsg -count=1
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	4.939s
$ go vet ./internal/factorymsg
(출력 없음, exit 0)
$ golangci-lint run ./internal/factorymsg/...
10 issues: errcheck 10 — store.go 8, launch_pending_rollback_test.go 2 (M1과 같은 수·같은 파일 분포, 새 파일 0건)
$ unset MOAI_KANBAN … && go test ./internal/cli -run 'FactoryMsg|McpFactory|FactoryMixed|FactoryOperational|LaunchPending|CodexLauncher' -count=1
HEAD 158c91289에서 5회: ok 4회(13.9s ~ 15.4s), FAIL 1회(첫 실행 — tail -1만 남겨 실패 테스트 이름을 잡지 못함, 미귀속)
$ unset MOAI_KANBAN … && go test ./internal/hook -run 'Factory' -count=1
FAIL — TestFactoryHookBenchmarkBudget (MOAI_FACTORY_BENCH=1 필요, 기준 트리 e503d07a4에서도 같은 FAIL),
       TestFactoryUserPromptSubmitRebindsLaunchPendingPeer (간헐 — 아래)
```

`TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`는 기준 트리(`git archive e503d07a4`를 scratch에 풀어 실행)에서도 간헐 실패한다. 실패 원인은 200ms 훅 바인드 기한 안의 `context deadline exceeded`다. 같은 부하에서 기준·변경 테스트 바이너리를 번갈아 12회씩 돌린 결과 기준 FAIL 4/12, 변경 FAIL 3/12(load average 약 26~34). 이 변경이 만든 실패로 보지 않는다.

### M3 — 배선 소유 기록, 저널, 배선 잠금 (REQ-DHR-001 ~ 004, 006)

증거 디렉터리는 gitignore 대상이므로 판정 근거를 값으로 적는다.

커밋:

| 커밋 | 내용 |
|---|---|
| `b9b425db0` | 리드 게이트 ① 특성 테스트 `TestIdemScopeMigrationPreservesForeignV1Tables` (t1082 v1 DDL이 얹힌 DB의 이관) |
| `206a0b701` | manifest: `generated_managed` provenance, `FileEntry.Parts`(kind/key/origin/hash/region) |
| `78ab5b83d` | codexwiring: 배선 잠금, 저널 쓰기 경로(P1~P6, D1/D2), 복구, Lstat 경계, 부분 origin 규칙 |
| `3f74a5f9e` | cli: update 경로의 lock-held 보고, enable의 lock-held·conflict 비영 종료, doctor 저널 보고(읽기 전용) |

AC 판정(명령은 `acceptance.md`의 것을 그대로 실행, HEAD `3f74a5f9e`, 작업 트리 깨끗함):

| AC | 판정 출력 | 증거 파일 | pass / fail / skip 이벤트 |
|---|---|---|---|
| AC-DHR-001 | `true` | `.moai/reports/t1100/ac001.jsonl` | 11 / 0 / 0 |
| AC-DHR-002 | `true` | `.moai/reports/t1100/ac002.jsonl` | 11 / 0 / 0 (하위 10건: P0~P6, P3_user_modified, D1, D2) |
| AC-DHR-021 | `false` | `.moai/reports/t1100/ac021.jsonl` | 4 / 0 / 1 — `disable` 하위 테스트가 M4 차단으로 skip |
| AC-DHR-022 | `true` | `.moai/reports/t1100/ac022.jsonl` | 5 / 0 / 0 |

AC-DHR-021은 PASS가 아니다. `moai tool disable codex`(REQ-DHR-005)가 M4 범위라 이 마일스톤에 없다. `disable` 하위 테스트는 `t.Skip("BLOCKED on M4: …")`로 남겼고 판정식은 skip을 `false`로 센다. M4가 disable 명령을 만들면서 이 하위 테스트를 채운다. 복구 진입점 자체(`codexwiring.Recover`)는 이미 있고 enable·update 하위 테스트가 그 경로를 지난다.

변이(작업 트리에서 임시 수정 후 원복, 원복은 파일 비교로 확인):

- AC-DHR-001: 테스트 안의 변이 표(`writeGuards`)로 rename 직전 재확인 제거, 기록 후 대조 제거, Lstat 검사 제거 — 셋 다 각 테스트의 `mutant_*` 하위 테스트에서 "탐지 안 됨"으로 관측되고, 정상 경로는 탐지됨.
- AC-DHR-022: `WiringLockRelPath`를 `.moai/.update.lock`으로 바꾼 변이(배선 잠금 대신 update 잠금을 다시 잡음) → `update_holds_update_lock` FAIL: `refresh under the update lock did not refresh (warn=warning: Codex wiring not refreshed: lock held by another process (.moai/.update.lock); …)`. 나머지 셋은 PASS.
- origin 규칙(`TestCodexWiringPartOrigins`): 증거 무시 변이 → `unknown_with_sidecar_evidence` FAIL(`recorded "preexisting", want unknown`), 기록 유지를 created에만 한정한 변이 → `preexisting_without_evidence` FAIL(`second pass promoted preexisting to "unknown"`).
- 리드 게이트 ① 특성 테스트: 이관 중 `DELETE FROM lane_endpoint_tombstones` 변이 → FAIL(`foreign v1 tables changed`), 이관 건너뛰기 변이 → FAIL(`sender_slot columns=0`).

RED 증거(E8):

- codexwiring 새 테스트는 구현 전 컴파일 RED: `undefined: PartKeyMCPTable`, `undefined: passOptions`, `undefined: wireWith`, `undefined: ErrWiringConflict` 등, `FAIL … [build failed]`.
- doctor 읽기 전용 보고는 구현 전 단언 RED: `doctor report lacks ".codex/hooks.json"`, `… ".codexwiring-orphanentry"`, `… "moai tool enable codex"`, `… "incomplete"`, `--- FAIL: TestCodexWiringRecoveryEntryPoints/doctor_readonly`.
- 편차(정직하게 적는다): `TestCodexWiringLockUnderUpdateLock`, `TestCodexWiringRecoveryEntryPoints`의 enable/update, `TestWiringLockOwnership`, `TestInspectJournalReadOnly`, `TestCodexWiringPartOrigins`, `TestGeneratedManagedPartsRoundTrip`은 해당 코드를 먼저 쓴 뒤 작성했다(test-after). 대신 위 변이로 각 테스트가 공허하지 않음을 확인했다.

품질 게이트(HEAD `3f74a5f9e` 트리):

```text
$ go test ./internal/codexwiring ./internal/manifest -count=1
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.565s
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.284s
$ go test -cover ./internal/codexwiring ./internal/manifest -count=1
codexwiring coverage: 85.9% (기준 트리 b9b425db0: 89.6%)  manifest coverage: 88.3%
$ go vet ./internal/codexwiring ./internal/manifest ./internal/cli
(출력 없음, exit 0)
$ GOOS=windows GOARCH=amd64 go build ./internal/codexwiring/ ./internal/manifest/ ./internal/cli/
(출력 없음, exit 0)
$ golangci-lint run ./internal/codexwiring/... ./internal/manifest/... ./internal/cli/...
25 issues: errcheck 25 — 기준 트리 b9b425db0(git archive → scratch)와 같은 수·같은 파일 분포, 새 지적 0
$ unset MOAI_KANBAN … && go test ./internal/cli -run 'Codex|Doctor|ToolEnable|UpdateCodex|Wiring|RunInit|InitCodex|Harness' -count=1
pass 1542, skip 10, FAIL 1 — TestCodexSpawn_RealAssemblyThroughStubTmux: 세션 환경의 MOAI_KANBAN_BACKEND/MOAI_FACTORY_WORKER가 unset 목록 밖이라 새어 들어간 것. 그 셋까지 지우고 단독 재실행하면 ok. 이 변경과 무관.
$ unset MOAI_KANBAN … MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test ./internal/cli -run 'FactoryOperational|FactoryMixed|LaunchPending|CodexLauncher|McpFactory|FactoryMsg' -count=1
exit 0
```

보존: `wire_test.go`, `sidecar_test.go`, `hooks_test.go`, `configtoml_test.go`, `statusline_test.go`는 수정하지 않았고 전부 통과한다(REQ-CW-005/006).

SPEC 재량 안의 결정:

- 저널은 `.moai/state/codex-wiring-journal.json` 한 파일(JSON 문서, 잠금 아래 원자적 교체). 새 패스는 복구 뒤 해결된 항목을 비우고 그 패스의 항목만 남긴다. 잠금은 `.moai/state/codex-wiring.lock`(pid + 프로세스 시작 지문). 죽은 pid·재사용 pid는 정리, 살아 있거나 판별 불가면 held, 내용을 못 읽는 잠금 파일은 1분 동안 held.
- 저널 항목 상태: `staged`, `complete`, `conflict`, `diverged`, `discarded`. 복구는 `staged`만 분류한다.
- 파일을 다시 쓸 필요가 없는 패스(바이트 동일)는 저널 없이 manifest 부분 기록만 갱신한다(파일이 바뀌지 않으므로 중단 지점이 없음).
- `whole-file` created 기록은 파일이 MoAI의 마지막 쓰기와 같을 때만 해시를 따라간다. 다른 무엇이 파일을 바꿨으면 그 주장을 버리고 부분 단위로 내려간다(사용자 handler가 섞인 파일을 unwire가 통째로 지우지 않게 하려는 것).
- 복구 진입점: `Wire`·`RefreshWiring`(enable과 update 경로)이 쓰기 전에 잠금 아래서 복구를 돈다. `RefreshWiring`은 배선 파일이 없어도 저널이 있으면 복구만 돈다. disable은 M4가 `codexwiring.Recover`를 부른다.
- doctor 보고는 `codexwiring.InspectJournal`(잠금·쓰기 없음)을 쓴다. Codex가 없는 머신의 비배선 프로젝트에서도 미완료 저널은 보고한다(그 경우 다른 Codex 소견은 내지 않음).
- `MOAI_T1100_EVIDENCE_DIR` skip 규약은 이번 AC 테스트에 적용하지 않았다. `acceptance.md` §A가 그 증거 채널을 AC-DHR-012·018·020·023에만 두고, AC-DHR-001·002·021·022 명령은 그 변수를 넘기지 않는다. 적용하면 네 판정식이 모두 skip으로 `false`가 된다 — 리드 판정 대상.

리드 게이트 ② — M2 검증에서 이름을 잡지 못한 `internal/cli` FAIL:

- 테스트 이름: `TestFactoryOperationalFixtureUsesProductionInit`
- 실패 출력: `factory_operational_fixture_test.go:151: Codex UserPromptSubmit did not run built binding path: {"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"factory messaging degraded: context deadline exceeded"}}`
- 재현(HEAD `3e7115d3d`, M2 부분집합 명령 그대로 `-json`): 1회 실행 FAIL, 실패 테스트는 이것 하나(통과 75).
- 반복: 단독 실행을 HEAD와 기준 트리(`git archive e503d07a4`를 scratch에 풀어 실행 — 이 트리의 체크아웃이 아님)로 번갈아 6회씩. HEAD FAIL 3/6, 기준 FAIL 1/6, 실패 메시지는 양쪽 같음. load average가 69→23으로 내려가는 동안 뒤쪽 실행은 양쪽 모두 통과.
- 코드 경로: 훅 바인드 전체가 200ms 기한(`internal/hook/user_prompt_submit.go:111`, `factoryHookInspectionDeadline`) 안에서 `ProbeProcessIdentity`, `ValidateActiveRun`, `factorymsg.Open`, `Peer`, `RegisterPeer`를 돈다. `git log e503d07a4..HEAD` 중 이 경로에 닿는 것은 `internal/factorymsg/store.go`·`schema_migrate.go`로, `Open`마다 카탈로그 조회 한 번(`ensureSchema`)이 더해졌다.
- 귀속: 기존 부하 민감 간헐 실패다(기준 트리에서도 같은 메시지로 재현). M2는 같은 예산 안에 조회 한 번을 더해 노출을 조금 늘렸을 수 있으나, 3/6 대 1/6은 표본이 작아 가를 수 없다. 결정적 실패가 아니므로 이 카드에서 고치지 않았다.

리드 게이트 ③ — `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` intermittent failure — t1109 소관, 기존 트리에서도 재현 (base 4/12, changed 3/12); not fixed in this card.

리드 게이트 ① — t1082 스키마 호환(병합하지 않고 확인만):

- 1차(t1082 tip `b08569bac`, merge-base `7755dce38`): `git merge-tree --write-tree HEAD WT-factory-lane-worktree-handoff` → 충돌 없음(트리 `c30db17fd`). 그 트리를 scratch에 풀어 `go build`(factorymsg·hook·cli)와 `go vet ./internal/factorymsg` 통과, `go test ./internal/factorymsg` ok. 의미 술어("t1082 DDL이 이미 적용된 v1 DB에서 t1100 이관이 행 수 대조를 통과하고 새 테이블을 보존한다")는 `TestIdemScopeMigrationPreservesForeignV1Tables`로 이 브랜치와 병합 트리 양쪽에서 PASS(`open`, `open_existing` 두 진입점). 역순(t1082가 먼저 develop에 착지) 경우도 같은 모양이다 — t1082만 가진 바이너리가 만든 DB는 v1 messages + 핸드오프 테이블이고, 병합 스키마의 `CREATE TABLE IF NOT EXISTS`는 기존 테이블을 건드리지 않으며 이관은 `sender_slot` 열 유무로만 판단한다(스키마 버전 상수에 의존하지 않음).
- 2차(이 카드 작업 중 t1082 tip이 `2c22fb10a`로 이동): 같은 명령이 `CONFLICT (content): Merge conflict in internal/factorymsg/store.go`를 낸다. 충돌 두 곳 — `verifyPeer`(이쪽 `ErrStalePeer` ↔ 저쪽 `s.staleOrUnregistered(ctx, p)`), `Send`의 멱등 충돌 재조회(이쪽 lane slot 범위 ↔ 저쪽 `lane_message_releases` 원 수신자 조회). `handoffSchema` DDL은 두 tip 사이에 바뀌지 않았고, 새로 `handoffBindSchema` 테이블 셋이 더해졌다(특성 테스트 픽스처에는 없음). 리드 지시에 따라 고치지 않았다 — 결함과 선택지는 완료 보고의 blocker 절에 있다.

### Absorb develop (post-t1112)

흡수: 로컬 `develop` `987fa3e73`을 `git merge --no-ff develop`로 흡수, 병합 커밋 `6a4d75f32`(부모 `b21036c1f`, `987fa3e73`). 충돌은 `internal/factorymsg/store.go` 두 곳. 해소는 t1112 이음매를 그대로 채택했다 — `ErrStalePeer`(Is 훅 godoc 포함)·`queryer`·`verifyPeerOn`/`verifyPeer`는 develop 판. 이 브랜치의 중복 `ErrStalePeer`는 버리고, `schema_migrate.go`에 있던 중복 `queryer` 인터페이스를 지웠다. dispatch의 트랜잭션 안 stale 검사 네 곳은 이미 `s.verifyPeerOn(ctx, tx, …)`를 거친다. `Send` 멱등 재조회의 lane slot 범위(분기 A)는 그대로다. 이음매 동일성: `git diff develop HEAD -- internal/factorymsg/stale_peer_seam_test.go internal/factorymsg/verify_peer_test.go` 출력 0바이트, `store.go`의 이음매 두 구간(`// ErrStalePeer reports`…`const launchPendingSessionPrefix`, `// verifyPeer checks p`…`func (s *Store) Send(`)은 develop 판과 바이트 동일(2473 바이트, 양쪽 같음).

병합 트리(`6a4d75f32`) 재측정 — acceptance.md 명령 그대로, AC-DHR-020은 재실행하지 않음(M1 증거 `ac020*` 보존, `outcome: reproduced` → 분기 A):

| AC | jq 판정 |
|---|---|
| AC-DHR-014 공통 | `true` |
| AC-DHR-014 분기 A | `true` |
| AC-DHR-015 | `true` |
| AC-DHR-016 | `true` |
| AC-DHR-001 | `true` |
| AC-DHR-002 | `true` |
| AC-DHR-022 | `true` |
| AC-DHR-021 | `false` — 유일한 원인은 `TestCodexWiringRecoveryEntryPoints/disable` skip(`BLOCKED on M4: moai tool disable codex (REQ-DHR-005) does not exist yet`). enable·update·doctor_readonly와 상위 테스트는 pass, fail 0 |

- `go test ./internal/factorymsg/... -count=1 -v` → `ok … 5.988s`. develop 이음매 테스트 `TestStalePeerOutcomeMatchesErrStalePeer`·`TestVerifyPeerOnRunsInsideOpenTx`와 이 브랜치의 `TestIdemScopeMigrationPreservesForeignV1Tables` 모두 PASS. SKIP 하나는 `TestIdemScopeRestartReproduction`(`NOT_RUN MOAI_T1100_EVIDENCE_DIR unset`) — AC-DHR-020 증거 작성 테스트라 의도된 skip.
- `go test ./internal/codexwiring/... ./internal/manifest/... -count=1` → 둘 다 `ok`.
- `go vet ./internal/factorymsg ./internal/codexwiring ./internal/manifest ./internal/cli` → exit 0.

(a) 증거 디렉터리 skip의 경계(리드 확인): `MOAI_T1100_EVIDENCE_DIR` NOT_RUN skip은 증거 파일을 쓰는 테스트에만 적용한다. 판정식이 pass 이벤트만 세는 AC(AC-DHR-001·002·021·022)의 테스트는 skip하지 않는다 — 거기서 skip하면 판정식이 공허한 `false`가 된다.

(b) M2가 `factorymsg` 열기에 더한 카탈로그 조회(`ensureSchema`의 `dispatches` 테이블·`sender_slot` 열 조회) 비용. 이미 이관된 DB에서 잰 벤치마크(`internal/factorymsg/ensure_schema_bench_test.go`, 운영 코드 무변경). 명령: `go test ./internal/factorymsg -run '^$' -bench 'BenchmarkEnsureSchema|BenchmarkPeersCount|BenchmarkOpenExisting' -benchtime=2000x -count=5`, 이어서 조회 유무 비교는 `-bench 'BenchmarkOpenExisting' -benchtime=200x -count=5`. 머신: Apple M4 Max, 측정 중 load average 약 15~16.

| 벤치마크 | ns/op (5회) |
|---|---|
| `BenchmarkEnsureSchemaMigrated`(조회 단독) | 33888 / 27310 / 30384 / 20805 / 18715 |
| `BenchmarkPeersCountReference`(같은 핸들의 기존 `peers` 조회, 척도용) | 3098 / 3175 / 3002 / 3019 / 3028 |
| `BenchmarkOpenExistingWithoutProbeReplica`(열기 − `ensureSchema`, 테스트 안 복제) | 39011324 / 39970116 / 39320995 / 38514948 / 39494672 |
| `BenchmarkOpenExistingMigrated`(`OpenExistingWithDeadline` + `Close` 전체) | 43113090 / 38977421 / 39117844 / 39078249 / 39358968 |

조회 하나의 비용은 약 19~34µs로, 훅 바인드 기한 200ms(`internal/hook/factory_messages.go:20` `factoryHookInspectionDeadline`, 사용처 `internal/hook/user_prompt_submit.go:111`)의 약 0.01~0.02%다. 열기 전체는 조회 유무와 관계없이 약 39ms/op로 같아서 두 열기 벤치마크의 차이는 잡음 안에 있다. 즉 M2가 더한 조회는 리드 게이트 ②의 간헐 실패를 설명할 크기가 아니다. 한편 조회와 무관한 기존 열기 비용 약 39ms 자체는 기한의 약 20%를 차지하며, 이것은 이 카드 이전부터 있던 비용이다(원인은 재지 않았다).

(c) errcheck: 흡수 전 `b21036c1f`에서 `golangci-lint run ./internal/factorymsg/... ./internal/codexwiring/... ./internal/manifest/... ./internal/cli/...` → `35 issues: * errcheck: 35`. 35건 전부 이 브랜치가 바꾸지 않은 줄이다 — `internal/factorymsg/store.go` 8건은 `git blame` 커밋 `45285bf1b`·`cb099897a`·`6bde8412c`·`8c5d9be99`로 모두 `7755dce38..b21036c1f` 밖이고, 나머지 27건이 있는 `internal/cli`·`internal/factorymsg` 파일은 이 브랜치의 변경 파일 목록에 없다. 흡수 후 `6a4d75f32`에서 같은 명령 → `0 issues.`(exit 0, golangci-lint 2.10.1). 이 브랜치가 따로 고친 줄은 없다 — 35건은 t1097이 develop에서 이미 고친 것이다. 벤치마크 파일 추가 뒤 `golangci-lint run ./internal/factorymsg/...` → `0 issues.`.

### M4 — `moai tool disable codex`, 고아 배선·미배포 템플릿 보고 (REQ-DHR-005, 007)

커밋:

| 커밋 | 내용 |
|---|---|
| `f8c08006f` | codexwiring: `Unwire`(created 부분만 제거, 저널 쓰기 경로·잠금·복구 재사용), 사이드카 disable 표식, `OwnedWiringFiles` |
| `2368cbd1c` | cli: `moai tool disable codex`, update 경로의 고아 배선 보고(재작성 없음), 더 이상 배포되지 않는 `.codex/` 템플릿 보고, AC-DHR-021 `disable` 하위 테스트 활성화 |

AC 판정(명령은 `acceptance.md`의 것을 그대로 실행, HEAD `2368cbd1c`, 작업 트리 깨끗함):

| AC | 판정 출력 | 증거 파일 | pass / fail / skip 이벤트 |
|---|---|---|---|
| AC-DHR-003 | `true` | `.moai/reports/t1100/ac003.jsonl` | 6 / 0 / 0 |
| AC-DHR-004 | `true` | `.moai/reports/t1100/ac004.jsonl` | 5 / 0 / 0 |
| AC-DHR-005 | `false` | `.moai/reports/t1100/ac005.jsonl` | 6 / 0 / 1 — `claude_to_both` 행 skip(아래 blocker) |
| AC-DHR-021 | `true` | `.moai/reports/t1100/ac021.jsonl` | 5 / 0 / 0 (`disable` 하위 테스트가 이제 실행·통과) |
| AC-DHR-001 (회귀) | `true` | `.moai/reports/t1100/ac001.jsonl` | 11 / 0 / 0 |
| AC-DHR-002 (회귀) | `true` | `.moai/reports/t1100/ac002.jsonl` | 11 / 0 / 0 |
| AC-DHR-022 (회귀) | `true` | `.moai/reports/t1100/ac022.jsonl` | 5 / 0 / 0 |

AC-DHR-005는 PASS가 아니다. 여섯 행 중 다섯이 통과하고 `claude_to_both` 한 행을 `t.Skip("BLOCKED on the template deployer: …")`로 남겼다. 그 행은 `.codex/` 단계에 닿기 전에 update 자체가 실패한다 — `template sync: Deploy Templates: deploy templates: template deploy mkdir ".agents/skills/moai": mkdir .agents/skills/moai: file exists`. claude 배포 뒤 `.agents/skills/moai`는 심볼릭 링크다(관측: `Lstat` mode `Lrwxr-xr-x`, both 배포 뒤는 `drwxr-xr-x`). 링크 대상이 update의 관리 경로 정리 단계에서 지워져 링크가 끊긴 채 both 배포자의 `MkdirAll`을 막는다는 것은 코드 판독 가설이며 재지 않았다. 수리 위치는 `internal/template` 배포자로, M4 파일 범위(`plan.md` §F) 밖이라 고치지 않았다. 이 M4 변경은 배포 단계보다 뒤(`renderUpdateOutcome` 다음)에만 한 줄을 더하므로 원인이 아니다(코드 판독 — 기준 트리에서 같은 행을 재지는 않았다).

RED 증거(E8):

- codexwiring 새 테스트, 구현 전 컴파일 RED: `undefined: UnwireResult`, `undefined: Unwire`, `undefined: ReasonUserOwned`, `undefined: ReasonNoProvenance`, `undefined: ReasonUnknownOrigin`, `undefined: ReasonModified`, `undefined: OwnedWiringFiles`, `FAIL … [build failed]`.
- cli 새 테스트, 구현 전 컴파일 RED: `undefined: runToolDisableCodexAt`(`codex_wiring_recovery_test.go:124`, `tool_disable_codex_test.go:51`, `:72`).
- AC-DHR-005 단언 RED(보고 구현 전): `undeployed .codex/agents/moai/builder-harness.toml not reported`(열두 역할 모두), `orphaned wiring not reported with "moai tool disable codex"`(`both_to_claude`, `gpt_to_claude`), `undeployed .codex/agents/moai/retired-role.toml not reported`(`older_binary_to_newer`).
- 첫 GREEN 시도의 실패 두 건(구현 수정으로 해소): 사용자가 키 순서를 바꿔 재직렬화한 MoAI handler가 `modified`로 남음 → handler 해시를 생성기 필드 순서로 정규화. 소유 기록이 없는 프로젝트에도 disable 표식 사이드카를 만들어 트리가 바뀜 → 표식은 이전 배선 증거가 있거나 무언가를 제거했을 때만.
- 편차: `TestOwnedWiringFiles`, `TestCodexUnwireHooksModifiedAndUnparseable`은 커버리지를 85% 위로 올리려고 구현 뒤에 썼다(test-after). AC-DHR-004의 두 번째 재배선 라운드는 변이 `unknown_promoted`가 살아남은 것을 보고 더했다(아래 표).

변이(작업 트리 임시 수정 → 실행 → 원복, 원복은 `shasum` 대조로 확인 — 네 파일 모두 수정 전과 같은 해시):

| 변이 | 대상 테스트 | 결과 |
|---|---|---|
| (i) unwire가 `unknown` 부분을 지움 | AC-DHR-004 | FAIL `old_install_unknown_origin`: `re-wired old install config changed` |
| (i) 같은 변이 | AC-DHR-003 | PASS — AC-DHR-003 픽스처에는 `unknown` 부분이 없다(사전 존재는 `preexisting`). 대신 `TestToolDisableCodexCommand/report`가 FAIL(`report lacks "kept … [mcp_servers.moai] (unknown-origin)"`) |
| 사전 존재 테이블을 `created`로 기록 | AC-DHR-003 | FAIL `preexisting_table_kept`: `config bytes` |
| 구분 빈 줄을 영역에서 뺌 | AC-DHR-003 | FAIL `config_appended_tables_bytes`, `config_status_line_key_bytes`, `preexisting_table_kept` |
| provenance 확인 없이 해시만 비교해 삭제 | AC-DHR-004 | FAIL `no_provenance_canonical_content`: `tree changed` |
| 재배선 첫 회에 `unknown` 대신 `created` 기록 | AC-DHR-004 | FAIL round 1 |
| 기존 `unknown` 기록을 다음 배선에서 `created`로 승격 | AC-DHR-004 | 처음엔 PASS(생존) → 두 번째 재배선 라운드를 더한 뒤 FAIL round 2 |
| (ii) update가 고아 배선을 보고 대신 제거(`Unwire` 호출) | AC-DHR-005 | FAIL `both_to_claude`, `gpt_to_claude`: `orphaned wiring rewritten` |
| (ii-b) update가 미배포 `.codex/` 템플릿을 지움 | AC-DHR-005 | FAIL `both_to_claude`, `gpt_to_claude`, `older_binary_to_newer`: `not preserved byte-identical` 25건 |

품질 게이트(HEAD `2368cbd1c` 트리):

```text
$ go test ./internal/codexwiring/... ./internal/manifest/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.665s	coverage: 86.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.312s	coverage: 88.3% of statements
$ golangci-lint run ./internal/codexwiring/... ./internal/manifest/... ./internal/cli/...
0 issues.
$ go vet ./internal/codexwiring ./internal/manifest ./internal/cli         (커밋 직전 같은 내용의 트리)
(출력 없음, exit 0)
$ GOOS=windows GOARCH=amd64 go build ./internal/codexwiring/ ./internal/manifest/ ./internal/cli/   (커밋 직전 같은 내용의 트리)
(출력 없음, exit 0)
$ unset MOAI_KANBAN … MOAI_FACTORY_WORKER && go test ./internal/cli -run 'Codex|Doctor|Tool|Wiring|Harness|TemplateSync|UpdateLLM|Mirror|RunUpdate|RunInit|InitCodex' -count=1 -json
pass 1750, skip 12, fail 1 — TestCodexSpawn_RealAssemblyThroughStubTmux: M3 때와 같은 환경 누출(MOAI_FACTORY_WORKERS가 unset 목록 밖). 그 변수까지 지우고 단독 재실행하면 ok.
```

`go vet`과 Windows 빌드는 두 커밋 직전, 테스트 파일 한 곳(`unwire_test.go` 재배선 두 번째 라운드)만 다른 트리에서 실행했다. 이후 커밋 HEAD에서 다시 돈 것은 lint와 두 패키지 테스트다.

범위: `internal/factorymsg`는 건드리지 않았다(`git diff --stat e8d42f959 HEAD -- internal/factorymsg` 출력 없음). `Store.Send`의 lane slot 멱등 조회 모양도 그대로다(리드 지시, t1082가 그 위에 원 수신자 조회를 얹는다). `CleanMoaiManagedPaths` 동작도 바꾸지 않았고 AC-DHR-005 테스트는 `.claude/` 아래를 단언하지 않는다.

SPEC 재량 안의 결정:

- **disable 표식.** `moai tool disable codex` 뒤에도 사용자 소유 `config.toml`이 남으면, 파일 존재를 opt-in으로 읽는 update 갱신(REQ-CW-009)이 다음 update에서 MoAI 테이블을 다시 붙인다. 그러면 결정 1("제거는 disable로만")이 update 한 번에 무력해진다. 그래서 사이드카에 `disabled: true`를 기록하고, 존재 게이트가 있는 update 갱신은 이 표식을 opt-out으로 읽어 배선하지 않는다(복구는 그대로 돈다). `moai tool enable codex`가 표식을 지운다. 표식은 이전 배선 증거가 있거나 무언가를 제거했을 때만 쓴다 — MoAI가 배선한 적 없는 프로젝트의 트리는 disable로 바뀌지 않는다(AC-DHR-004 첫 대상).
- **고아 판정.** "구성된 프로필이 Codex를 쓰지 않는다(`claude`)" 그리고 "manifest가 `.codex/` 템플릿 배포를 기록한다(생성 배선 파일 제외)"일 때만 고아다. claude 프로필에서 `moai tool enable codex`로 배선한 프로젝트는 `.codex/` 템플릿이 배포된 적이 없으므로 고아가 아니고 예전처럼 갱신된다. 고아로 보고하는 파일은 `OwnedWiringFiles` — created 부분을 가진 파일, 그리고 부분 기록이 전혀 없는데 사이드카가 있는 구버전 설치의 배선 파일.
- **미배포 템플릿 보고.** update가 로드한 manifest의 `.codex/` 항목(생성 배선 제외) 중 이번 배포 목록(`restoredSet`)에 없고 디스크에 있는 경로를 보고한다. 프로필 전환(both → claude)과 구버전 → 신버전의 폐기 템플릿을 같은 규칙이 덮는다. 지우지 않는다.
- **update는 manifest를 저장하지 않는다(관측).** update 경로는 배포 추적을 메모리에만 하고 `manifest.json`에 쓰지 않는다(`grep '\.Save()'`로 저장처는 `core/project/initializer.go`와 `cli/profile_setup.go`뿐; update만 거친 픽스처의 manifest에는 `.codex/` 템플릿 키가 없었다). 그래서 AC-DHR-005 테스트는 각 행의 출발 배포를 `moai init --llm <from>`으로 만든다. update만으로 만들어진 프로젝트에서는 미배포 템플릿 보고와 고아 판정이 발화하지 않는다 — 잔여 위험으로 적는다.
- **hooks.json handler 비교.** 기록 해시는 생성기가 쓴 필드 순서(`type`, `command`, `timeout`)의 compact JSON 해시다. 세 키만 가진 handler는 그 순서로 다시 직렬화해 비교하고, 그 밖은 compact 바이트로 비교한다. 객체 키 순서를 비교하지 않는다는 REQ-DHR-005와 맞춘 것이다.
- **unwire 결과 파일은 화이트리스트 게이트를 거치지 않는다.** 부분 제거는 새 키를 만들지 않으므로 생성기의 REQ-CW-003 게이트를 적용하지 않았다. 사용자 파일에 이미 있던 비화이트리스트 키(테스트의 `x_user_note`)는 그대로 둔다.
- **메시지 언어.** 기존 `tool enable codex`와 update 경고처럼 CLI 문구는 영어(현지화 계층 없음), 경고는 `warning:` 접두사로 stderr.
- `MOAI_T1100_EVIDENCE_DIR` skip 규약은 적용하지 않았다 — 이번 AC 테스트는 모두 통과 이벤트만 세고 증거 파일을 쓰지 않는다(리드 확인 경계).

후속 카드 후보(리드 결정, 착지 후): factorymsg `Open` 자체가 부하 걸린 머신에서 약 39–43 ms/op(훅 바인드 예산 200 ms의 약 20%)이며 원인은 재지 않았다(측정은 "Absorb develop" 절 (b)).

### M5 — Codex worktree anchor, 폐기 동등성, `moai codex -k` (REQ-DHR-008 ~ 012)

커밋:

| 커밋 | 내용 |
|---|---|
| `990ab3416` | session: lock 보유자 접근자(`CodexAnchorLockReason`, `LockReasonPID`, `LockHolderConfirmedDead`) — 새 파일 `anchor_lock_holder.go`. `anchor_lock.go`는 바꾸지 않았다 |
| `669dda570` | worktree: `remove`가 lock-aware 판정으로 anchor를 가린다. `TestWorktreeDisposalRefusesUnintegratedCodexTree` |
| `26227a7b4` | cli: `moai codex -w` anchor lock·교체 가드·생성 base 대조·동시 writer 거부, `moai cc -w` 읽기 전용 사전 판정, `moai codex -k` |

AC 판정(명령은 `acceptance.md`의 것을 그대로 실행. 측정 트리는 커밋 `26227a7b4`와 같은 내용 — AC 실행 뒤 코드 변경 없음, 이 문서 커밋만 뒤따른다):

| AC | 판정 출력 | 증거 파일 | pass / fail / skip 이벤트 |
|---|---|---|---|
| AC-DHR-006 | `true` | `.moai/reports/t1100/ac006.jsonl` | 7 / 0 / 0 |
| AC-DHR-007 | `true` | `.moai/reports/t1100/ac007.jsonl` | 10 / 0 / 0 |
| AC-DHR-008 | `true` | `.moai/reports/t1100/ac008.jsonl` | 6 / 0 / 0 |
| AC-DHR-009 | `true` | `.moai/reports/t1100/ac009.jsonl` | 11 / 0 / 0 |
| AC-DHR-001 (회귀) | `true` | `.moai/reports/t1100/ac001.jsonl` | 11 / 0 / 0 |
| AC-DHR-002 (회귀) | `true` | `.moai/reports/t1100/ac002.jsonl` | 11 / 0 / 0 |
| AC-DHR-003 (회귀) | `true` | `.moai/reports/t1100/ac003.jsonl` | 6 / 0 / 0 |
| AC-DHR-004 (회귀) | `true` | `.moai/reports/t1100/ac004.jsonl` | 5 / 0 / 0 |
| AC-DHR-021 (회귀) | `true` | `.moai/reports/t1100/ac021.jsonl` | 5 / 0 / 0 |
| AC-DHR-022 (회귀) | `true` | `.moai/reports/t1100/ac022.jsonl` | 5 / 0 / 0 |
| AC-DHR-005 | `false` — **PARTIAL / UNPROVEN** (아래) | `.moai/reports/t1100/ac005.jsonl` | 6 / 0 / 1 |

AC-DHR-006의 Windows 분기는 `NOT_RUN`이다. `codex_direct_windows.go`(`//go:build windows`)의 `codexDirectAnchorPID`는 darwin 로컬 실행에서 컴파일되지 않고, CI Windows 잡에서 이 테스트의 pass를 관측하기 전에는 인용하지 않는다. 로컬에서 잰 것은 `GOOS=windows GOARCH=amd64` 빌드와 vet이 통과한다는 것뿐이다(아래). 위 `true`는 darwin 분기다.

**AC-DHR-005 — PARTIAL / UNPROVEN.** PASS로 세지 않는다. M5는 이 AC의 행을 바꾸지 않았다(M4와 같은 6 / 0 / 1).

- 통과한 행(5): `gpt_to_both`, `both_to_claude`, `both_to_gpt`, `gpt_to_claude`, `older_binary_to_newer`.
- 건너뛴 행(1): `claude_to_both` — `t.Skip`. 건너뛴 이유(테스트 출력 그대로): `BLOCKED on the template deployer: update fails at mkdir .agents/skills/moai (claude skill-mirror link) before any .codex/ step`. 그 밑의 update 오류(M4 관측): `moai update`가 `mkdir .agents/skills/moai: file exists`로 실패한다.
- 추정 원인: 관리 경로 정리 단계가 symlink의 대상을 지워, 끊긴 링크가 배포자의 `MkdirAll`을 막는다. **재현·측정하지 않았다** — 코드 판독 가설이다.
- 건너뛴 행은 PASS로 세지 않는다. 판정식은 skip 0을 요구하므로 `false`가 맞는 출력이다.
- 후속: 배포자 결함은 별도 카드로 간다(리드 결정 (a), 2026-09-23). t1100은 고치지 않는다. 그 카드가 develop에 착지하면 t1100은 develop을 흡수하고 sync 전에 AC-DHR-005를 다시 돌린다.

RED 증거(E8):

- 컴파일 RED(구현 전, `go vet`): `vet: internal/cli/worktree/disposal_codex_tree_test.go:63:18: undefined: session.CodexAnchorLockReason`, `internal/cli/codex_worktree_anchor_test.go:149:35: undefined: codexWorktreeWriterCheck`, `…:149:61: undefined: codexWorktreeAnchorLock`, `…:149:86: undefined: codexResolveBaseCommit`.
- 동작 RED(동작 없는 스켈레톤을 넣고 실행, 스켈레톤은 GREEN 전에 지움):
  - `codex_worktree_anchor_test.go:194: …/new-tree: no git worktree lock after launch`
  - `…:221: launch accepted a tree whose HEAD is not the resolved base`
  - `…:257: launch replaced or ignored a live lock`
  - `…:300: round 0: both launchers proceeded`
  - `…:418: codex -w live_lock: launch into an anchored tree was allowed` (unreadable_lock·registry_only, cc 쪽 세 칸도 같은 모양)
  - `…:584: anchor got (0, ""), cleanups 0; want the pane identity and no cleanup`
  - `codex_kanban_test.go:151: codex [-k SPEC-PARITY-001]:  (stderr "unknown verb - usage: moai codex [cli] [-w [worktree]] [-- codex-args...] | moai codex status | moai codex app\n")`
  - `disposal_codex_tree_test.go:259: …/codex-locked: refusal must name the anchor source (lock), got: remove worktree: remove worktree at "…` (claude-locked도 같음)
- 스켈레톤에서 이미 통과한 것: `TestPRMergeCleanupRefusesAnchoredCodexTree`, `TestWorktreeDisposalRefusesUnintegratedCodexTree/{done,done_force,clean}`, `TestCodexKanbanEntryParity/unsupported`. 설계 §B.3이 "시험만"이라 한 경로이고, 빈틈은 드러나지 않았다. 그래서 `clean.go`와 `session_worktree_prmerge.go`는 바꾸지 않았다.
- 편차: 교체 경합 테스트의 두 번째 루프(한 launcher가 끝난 뒤 낡은 읽기를 가진 다른 launcher가 진행)는 GREEN 뒤, 아래 변이 (v)가 첫 루프만으로는 약하게 잡힌 것을 보고 더했다.

변이(작업 트리 임시 수정 → 실행 → 원복. 원복은 `shasum` 대조로 확인 — 세 파일 모두 수정 전 해시와 같음):

| 변이 | 대상 테스트 | 결과 |
|---|---|---|
| (i) 동시 writer 판정이 lock을 무시(`AnchorDecision(tree, session.LockInfo{}, now)`) | AC-DHR-007 | FAIL `live_lock/codex`, `live_lock/cc`, `unreadable_lock/codex`, `unreadable_lock/cc` (registry_only는 통과 — lock 무관 경로) |
| (ii) `remove`가 lock을 무시 | AC-DHR-008 | FAIL `remove`: `codex-locked`, `claude-locked` 둘 다 `refusal must name the anchor source (lock)` |
| (iii) `moai cc -w` 사전 판정이 lock을 씀 | AC-DHR-007 | FAIL `dead_lock_allowed/cc` (`cc pre-check rewrote the lock: "claude session dead (pid 94231)" -> "moai codex session dead (pid 90202)"`), `cc_precheck_never_locks` (`cc pre-check wrote a lock`) |
| (iv) 생성 base 대조를 끔 | AC-DHR-006 | FAIL `base_mismatch_refused_tree_kept`: `launch accepted a tree whose HEAD is not the resolved base` |
| (v) 교체 가드의 O_EXCL과 가드 안 재읽기를 함께 뺌 | AC-DHR-006 | FAIL(첫 루프만 있을 때) — 다만 "패자 거부문이 worktree lock을 말하지 않음"으로만 잡혔다(git 자신의 `is not locked`로 패자가 떨어짐) |
| (v') 가드 안 재읽기만 뺌 | AC-DHR-006(두 번째 루프 추가 후) | FAIL `stale round 0: the launcher holding a stale read also proceeded` — 설계 §B.1이 적은 위험(늦은 쪽 unlock이 먼저 쪽 새 lock을 지움)을 그대로 재현 |

품질 게이트(HEAD `26227a7b4`와 같은 내용의 트리):

```text
$ go vet ./internal/cli/ ./internal/cli/worktree/ ./internal/session/
(출력 없음, exit 0)
$ golangci-lint run ./internal/cli/... ./internal/session/...
0 issues.
$ GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/cli/worktree/ ./internal/session/
(출력 없음, exit 0)
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/ ./internal/session/ ./internal/cli/worktree/
(출력 없음, exit 0)
$ go test ./internal/cli/worktree/... ./internal/session/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	40.450s	coverage: 87.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/session	18.633s	coverage: 88.0% of statements
$ unset MOAI_KANBAN … MOAI_FACTORY_WORKERS && go test ./internal/cli -run 'Codex|CC|Cc|Claude|Worktree|Kanban|Kanb|Launch|Launcher|Spawn|PRMerge|SessionWorktree|Factory|Lead|Companion|Glm|GLM' -count=1 -json
pass 2177, fail 0, skip 18 (skip는 LIVE 테스트와 기존 skip), 패키지 pass 318.1s
```

`internal/cli` 패키지 전체 커버리지는 재지 않았다(전체 스위트 금지). M5 새 함수의 함수별 커버리지(M5 테스트만 실행): `stripCodexKanbanFlag` 96.6%, `applyCodexKanbanEntry` 100%, `placeCodexAnchorLock` 82.1%, `worktreeWriterRefusal` 83.3%, `readWorktreeLock` 87.5%, `verifyCodexWorktreeBase` 83.3%, `ccWorktreeWriterPrecheck` 76.9%, `lockCodexTree` 62.5% — 덮이지 않은 쪽은 git 명령 실패 분기다.

범위: `git diff --stat 47a886ded HEAD -- internal/factorymsg internal/cli/worktree/done.go internal/session/anchor_lock.go` 출력 없음. `Store.Send`의 lane slot 멱등 조회 모양 그대로(t1082가 그 위에 얹는다). `done`의 L1 거부 그대로이며 `done`/`done --force`가 여덟 트리 모두 `L1_SESSION_WORKTREE`로 거부하는 것을 AC-DHR-008이 잰다.

SPEC 재량 안의 결정:

- **새 파일 셋.** `internal/session/anchor_lock_holder.go`(접근자), `internal/cli/codex_kanban.go`(`-k` 파서·적용), 테스트 파일들. 계획 §F 파일 목록 밖이지만 모두 목록에 있는 기능의 몸체다. `anchor_lock.go`를 고치지 않고 재사용하려면 lock 보유자 pid와 "확정 사망"을 읽을 수단이 필요했고, 판정 로직을 복제하지 않으려고 같은 패키지의 비공개 함수에 위임하는 접근자만 더했다.
- **`cc.go` 한 곳 호출.** `moai cc -w` 사전 판정의 몸체는 `session_worktree.go`(`ccWorktreeWriterPrecheck`)에 있고, 호출은 `runClaudeEntry`의 worktree 처리 지점(`resolveWorktreeExistingBranch` 다음, `normalizeWorktreeFlag` 전) 한 줄이다. 이미 있는 트리에만 적용하고, lock은 읽기만 한다.
- **lock pid.** 직접 launch는 POSIX·Windows 모두 `os.Getpid()`다. 값은 같지만 이유가 달라(POSIX는 exec로 pid가 이어짐, Windows는 기다리는 부모) build-tag 파일마다 `codexDirectAnchorPID`로 따로 두고 이유를 적었다. lock은 `codexDirectLaunch` 호출 직전에 건다. `--spawn`은 pane이 생긴 뒤에야 pid가 있으므로 pane 안에서 lock을 걸고, 실패하면 pane을 닫는다(`TestCodexSpawnAnchorsToPanePID`, AC 밖).
- **lock 사유.** `moai codex session <트리 이름> (pid <n> start <process-start>)`. start를 모르면 생략.
- **교체 가드.** `<git-dir>/moai-anchor-replace`를 O_EXCL로 만들고 끝나면 지운다. 가드 생성 실패, 가드 안 재읽기의 사유 변화, `git worktree lock` 실패, 쓴 뒤 읽은 사유가 자기 pid가 아닌 경우 모두 거부. 거부문에 가드 경로와 "launcher가 없으면 지우고 다시 시도"를 적었다.
- **생성 base 대조.** base를 커밋으로 먼저 해석하고(`rev-parse --verify <base>^{commit}`), 생성 뒤 HEAD와 비교한다. 다르면 거부하고 트리는 남긴다. 해석 정책(`codexWorktreeBase`)은 바꾸지 않았다.
- **동시 writer 판정.** 기존 lock-and-registry 합집합(`AnchorDecision`)을 그대로 쓰고, lock 사유의 pid가 이 프로세스이면 자기 것으로 본다. 진단은 `WORKTREE_WRITER_ANCHORED:` 접두사 + 출처(`source: lock|registry`) + 보유자(lock 사유 원문, registry면 `session <id> pid <n>`). lock 목록을 읽지 못하면 미확정으로 거부한다. 새로 만든 트리에는 판정을 걸지 않는다(작성자가 있을 수 없음).
- **`remove`.** 레지스트리 경로는 기존 문구 그대로 먼저 본다. 레지스트리가 비었을 때 lock을 읽어 살아 있거나 미확정인 보유자면 `ANCHORED_SESSIONS_PRESENT: live session anchored in <path> - <detail> (source: lock, holder: <사유>)`로 거부. lock 목록 판독 실패도 미확정으로 거부(기존에는 레지스트리만 봤으므로 동작 변화). 죽은 lock은 moai가 막지 않고 git 판정에 맡긴다(기존 명시 제거 의미 유지).
- **`moai codex -k`.** 토큰은 동사 조회 전에 떼어 내고, 모양 판정은 cc의 `parseKanbanFlag`·`parseCompanionLabel`·`parseLeadLabel`에 넘긴다. 이름은 cc와 같은 레지스트리(`resolveCompanionName`, `appendLeadName`)로 claim하고, Codex에는 `--name`이 없으므로 환경변수(`MOAI_KANBAN_LABEL`, `MOAI_KANBAN_LEAD_NAME`)로 전한다. 받지 않는 모양: `-k <수>`, `-k --name worker-<n>`, `--kanban=<값>`, `-k`와 `-f` 동시, 읽기 전용 `status`와 `-k`, `-k` 없는 `--name`, 값 없는 `--name`. `--spawn` 경로에도 kanban 사실이 가도록 tmux 명령 앞 환경 목록에 `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`을 더했다(값이 있을 때만 붙음).
- **캡처 하네스.** 기존 launch 메커니즘 테스트는 git 저장소가 아닌 평범한 디렉터리를 쓰므로 `withCodexLaunchCapture`가 anchor seam 넷(writer 판정, lock, base 해석, base 대조)을 열어 둔다. anchor를 재는 테스트는 `withRealCodexWorktreeAnchor`로 실제 몸체를 되돌린다. 위 변이 (i)~(v')는 모두 실제 몸체에서 잡혔다.
- **메시지 언어.** 기존 CLI 문구처럼 영어.

잔여 위험(기록만, 착지 후 리드 분류):

1. update 경로는 manifest를 저장하지 않는다. update만 거친 프로젝트에는 고아 배선·미배포 템플릿 보고가 나오지 않는다(M4 관측).
2. claude 프로필로 바꾼 뒤 운영자가 의도적으로 `moai tool enable codex`를 한 경우에도, manifest에 `.codex/` 템플릿 배포 기록이 남아 있으면 고아로 보고된다.
3. 미배포 템플릿 경고가 update마다 12줄씩 반복된다.
4. (M5) 교체 가드 파일은 가드를 쥔 launcher가 교체 도중 죽으면 남는다. 그 뒤 그 트리의 죽은 lock 교체는 가드가 지워질 때까지 거부된다(거부문이 가드 경로와 조치를 적는다). 자동 회수는 없다.
5. (M5) Windows에서 기다리는 부모만 강제 종료되면 자식 Codex가 살아 있어도 lock이 죽은 것으로 읽힌다(설계 §B.1, plan §G). Windows 분기 자체가 로컬에서 `NOT_RUN`이다.
6. (M5) POSIX 직접 launch에서 exec가 실패하면 이미 건 lock이 남는다. 그 pid는 곧 죽으므로 anchor가 아니게 되고 다음 launch가 교체하지만, 그 사이 짧게 anchored로 읽힐 수 있다.
7. (M5) `moai cc -w`의 사전 판정은 이름 값을 `<project root>/.claude/worktrees/<name>`으로 해석한다. Claude Code가 이름을 다른 뿌리(예: git 최상위가 project root와 다른 경우)로 해석하면 판정이 다른 트리를 볼 수 있다. 이 저장소 구성에서는 둘이 같다(코드 판독, 미측정).

### Absorb develop 5f264c381 (t1125) + AC-005 re-measure

위 M5 절의 AC-DHR-005 PARTIAL / UNPROVEN 기록은 그대로 둔다. 이 절은 그 뒤에 이어진 측정이다.

흡수: 리드가 고정한 로컬 develop 커밋 `5f264c38141139cbad47bd4b0e0240d841d6acbb`(card t1125, `moai update`의 `.agents/skills/moai` 끊긴 링크 수리)을 `git merge --no-ff 5f264c381…`로 흡수했다. 움직이는 브랜치 이름이 아니라 고정 SHA를 병합했다. 병합 커밋 `ab4a2662d`(부모 `cff38a43e`, `5f264c381`). 충돌 없음. 병합 직후 기본 메시지 커밋(`67ffc8065`)을 카드 id가 든 메시지로 고쳐 쓴 것이 `ab4a2662d`이며, 트리와 부모는 같다.

unskip: `internal/cli/harness_profile_transition_test.go`의 `claude_to_both` 행에서 `BLOCKED` skip 사유와 그 설명 주석을 지웠다(커밋 `d46141178`). `internal/template`은 고치지 않았다.

AC 판정(명령은 `acceptance.md`의 것을 그대로 실행. `internal/cli` 세 건은 같은 compound 호출 앞에 `unset MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS &&`를 붙였다. 측정 트리 = HEAD `d46141178`, `git rev-parse HEAD^{tree}` → `0eb3352c4a895b54699fb474dec16eeefbccf72a`):

| AC | 판정 출력 | 증거 파일 | pass / fail / skip 이벤트 |
|---|---|---|---|
| AC-DHR-005 | `true` | `.moai/reports/t1100/ac005.jsonl` | 7 / 0 / 0 — 여섯 행 전부 pass(`claude_to_both` 0.95s 포함) |
| AC-DHR-001 (회귀) | `true` | `.moai/reports/t1100/ac001.jsonl` | 11 / 0 / 0 |
| AC-DHR-002 (회귀) | `true` | `.moai/reports/t1100/ac002.jsonl` | 11 / 0 / 0 |
| AC-DHR-003 (회귀) | `true` | `.moai/reports/t1100/ac003.jsonl` | 6 / 0 / 0 |
| AC-DHR-004 (회귀) | `true` | `.moai/reports/t1100/ac004.jsonl` | 5 / 0 / 0 |
| AC-DHR-021 (회귀) | `true` | `.moai/reports/t1100/ac021.jsonl` | 5 / 0 / 0 |
| AC-DHR-022 (회귀) | `true` | `.moai/reports/t1100/ac022.jsonl` | 5 / 0 / 0 |

pass/fail/skip 수는 각 파일에 `jq -sr '"pass=\([.[]|select(.Action=="pass" and .Test!=null)]|length) fail=\([.[]|select(.Action=="fail")]|length) skip=\([.[]|select(.Action=="skip")]|length)"'`로 셌다.

`claude_to_both`가 t1125 때문에 통과하는지 확인한 변이: `internal/template/skill_mirror.go`의 `releaseOwnMirrorLink` 첫 줄에 `return nil`을 넣고 `go test -run 'TestHarnessProfileTransitionPreservesUserData/claude_to_both' ./internal/cli/`를 돌리면 `deploy templates: template deploy mkdir ".agents/skills/moai": mkdir .agents/skills/moai: file exists`로 FAIL한다 — M4·M5가 기록한 오류와 같은 문구다. 변이를 지운 뒤 `git status --short`에 남은 변경은 테스트 파일 하나뿐이었다(템플릿 파일은 커밋 상태로 복귀).

M5 가설과의 대조: M5는 "관리 경로 정리 단계가 symlink의 대상을 지워, 끊긴 링크가 배포자의 `MkdirAll`을 막는다"를 미측정 가설로 적었다. t1125는 `DeployWithResult`에서 부모 디렉터리 `MkdirAll` 바로 앞에 `releaseOwnMirrorLink`를 넣어, 배포가 `.agents/skills/<skill>/` 아래에 실제 파일을 쓸 때 본문이 정확히 `MirrorLinkTarget(skill)`인 자기 링크만 지운다. 고친 자리(`MkdirAll` 직전)와 막힌 경로(`.agents/skills/moai`)는 가설과 일치하고, 위 변이로 이 수리가 `claude_to_both`를 통과시키는 원인임을 확인했다. 다만 "정리 단계가 링크 대상을 지운다"는 전반부는 t1125의 godoc이 같은 내용을 적고 있을 뿐, 이 흡수에서 링크 대상의 부재를 `Lstat`로 따로 재지는 않았다.

품질 게이트(HEAD `d46141178` 트리):

```text
$ go vet ./internal/cli ./internal/codexwiring ./internal/manifest
(출력 없음) vet_exit=0
$ golangci-lint run ./internal/cli/... ./internal/codexwiring/... ./internal/manifest/...
0 issues.
```

범위: `internal/factorymsg`의 `Store.Send` lane slot 조회는 건드리지 않았다 — `git diff --stat cff38a43e d46141178 -- internal/factorymsg` 출력 없음. unskip 커밋은 `git diff --stat ab4a2662d d46141178` 기준 테스트 파일 하나(1 insertion, 7 deletions)다. M6는 시작하지 않았다.

미측정: AC-DHR-006~009는 이번에 다시 돌리지 않았다(요청 범위 밖). `internal/cli` 전체 스위트와 Windows 빌드도 돌리지 않았다. golangci-lint의 exit code는 파이프 뒤 `tail`의 것이라 판정 근거는 `0 issues.` 출력이다.

### M6 — 역할 권한 계약과 감사 역할 Codex 예외 (REQ-DHR-013 ~ 015)

커밋: RED `4c2c316fc`(테스트 두 파일만) → GREEN `8d1480148`(구현·manifest·생성물·AGENTS 행). 시작 HEAD `0a635a852`. 측정 트리 = HEAD `8d1480148`, `git rev-parse HEAD^{tree}` → `b7e4b2c3c3af7a24d940c48c15b225b4d415cd7d`. RED 커밋이 GREEN 커밋보다 앞서므로 테스트가 먼저라는 순서는 커밋 그래프가 증언한다.

AC 판정(명령은 `acceptance.md`의 것을 그대로 실행, HEAD `8d1480148`에서 재실행):

| AC | 판정 출력 | 증거 파일 |
|---|---|---|
| AC-DHR-010 | `true` | `.moai/reports/t1100/ac010.jsonl` |
| AC-DHR-011 | `true` | `.moai/reports/t1100/ac011.jsonl` |
| AC-DHR-013 | `true` | `.moai/reports/t1100/ac013.jsonl` |

AC-DHR-012·023은 LIVE(M8)라 이번에 돌리지 않았다 — `NOT_RUN`.

RED(구현 전, 원문 발췌 — 전체는 `.moai/reports/t1100/m6/red-ac013.txt`, `red-ac010-011.txt`):

```text
--- FAIL: TestCodexAuditRolesReadOnlyScopedException (0.02s)
    audit_role_exception_test.go:55: sync-auditor: emitted sandbox_mode = "workspace-write", want read-only
    audit_role_exception_test.go:65: .codex/agents/moai/plan-auditor.toml: committed artifact does not carry the return-text instruction
    audit_role_exception_test.go:110: AGENTS.md template carries no "| audit-verdict-file |" capability row
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.388s
---
internal/template/agentemit/permission_contract_test.go:29:85: undefined: agentemit.AxisMapping
internal/template/agentemit/permission_contract_test.go:99:27: undefined: agentemit.BuildPermissionReport
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit [build failed]
```

설계 요점:

- **계약의 위치.** `agents-codex.yaml`에 `permission_contract`(축 표 8행 — `mcp-server`는 grant/deny 두 제한 종류로 나뉜다 — 와 역할별 계약 sandbox)와 `codex_role_addenda`를 둔다. 역할이 요구하는 제한은 `permission.go` `RoleRequirements`가 Claude 도구 목록과 계약 sandbox에서 도출한다: sandbox/mode 항상, 쓰는 역할은 write-path-scope, Claude 쪽에 없는 도구 계열은 shell/subagent/web deny, moai MCP 보유 역할은 mcp-server grant + mcp-tool subset, 비보유 역할은 mcp-server deny. UNSUPPORTED 보고 집합은 이 도출에서 계산되며 고정 목록이 아니다.
- **축 매핑(설계 §C.1과 같음).** enforced 2행: sandbox/mode(`sandbox_mode`, measured — Codex가 값을 받아들인다는 관측이며 런타임 쓰기 차단은 주장하지 않음), mcp-server/grant(`mcp_servers`, measured). UNSUPPORTED 6행: write-path-scope(measured), shell(unmeasured), mcp-server/deny(unmeasured), mcp-tool(documented), subagent(unmeasured), web(documented).
- **거부 조건.** manifest 검증: 7축 전부 존재, 행마다 두 값 중 하나, unmeasured enforced 거부, enforced 행이 생성기가 실제로 쓰지 않는 필드를 가리키면 거부, 증거 문구 없는 행 거부. 생성 시점: 요구 제한에 맞는 행이 없으면 실패, 방출 `sandbox_mode`가 계약 sandbox와 다르면 실패(계약 sandbox는 `role_values`와 일부러 따로 둔다 — 한쪽만 넓히면 걸린다).
- **보고.** `BuildPermissionReport`는 생성기가 방출하면서 검사한 판정을 그대로 돌려준다. `Pass`에는 enforced만 들어가고 UNSUPPORTED는 `Unsupported`에만 들어간다.
- **감사 역할 예외.** `plan-auditor`·`sync-auditor`의 `role_values`와 계약 sandbox를 `read-only`로 두고, 두 역할에만 Codex 전용 addendum("판정·보고 텍스트를 최종 응답으로 반환하고 파일을 직접 쓰지 않는다")을 본문 뒤에 붙인다. `plan-auditor`는 Claude 쪽 도구에 Write·Edit가 있어 기존 "read-only 역할의 쓰기 도구 거부" 검사에 걸리므로, addendum이 있는 역할에만 그 검사를 면제한다(쓰기가 부모로 옮겨졌기 때문). `read-vs-write-distinction` 근거와 `shell` 계열 근거 문구를 새 계약에 맞게 고쳤다.
- **부모 지시 표면 = 배포 `AGENTS.md`(`templates/AGENTS.md.tmpl`)의 capability 표 새 행 `audit-verdict-file`.** 고른 이유: Codex는 `AGENTS.md`를 항상 읽고, 이 표는 "어느 하네스가 그 능력을 갖지 못한 곳에만 행을 둔다"는 규칙으로 정확히 이 경우를 담는다. Claude 구현 열은 "감사자가 자기 판정 파일을 쓴다"로 두어 Claude 경로가 그대로임을 같은 행이 말한다. Codex 전용 발행물(`.agents/skills/moai-*`)은 부모 세션이 반드시 읽는다는 보장이 없어 택하지 않았다. 저장소 루트 `AGENTS.md`는 고치지 않았다 — 이 저장소는 `.codex/agents/`를 배포받아 쓰지 않으므로(`git ls-files .codex` → `.codex/config.toml`만) 그 행이 가리킬 Codex 감사 역할이 없다.
- **Claude 경로 불변.** `git diff --stat 0a635a852 HEAD -- .claude/agents/moai/plan-auditor.md .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/factorymsg` 출력 0줄.

생성물 diff(`make agents-emit`, 손편집 없음): `plan-auditor.toml`·`sync-auditor.toml` 각 +6/−1 — 머리 주석 한 줄(`# Codex-only addendum appended after the verbatim body (mapping manifest).`), 본문 뒤 `## Codex Runtime Addendum` 단락, `sandbox_mode = "workspace-write"` → `"read-only"`. 나머지 10개 TOML은 바이트 그대로다.

기존 테스트 수정(REQ-DHR-013/015가 요구): `golden_test.go` 감사 역할 sandbox 기대값을 read-only로, 본문 비교를 "원문 본문 + manifest addendum"으로 바꿨다(원문 본문이 앞에 그대로 오는지도 따로 단언). `agentemit_test.go`의 ship-omitted 경로는 계약이 있는 상태에서 `sandbox_mode`를 빼면 실패해야 함을 먼저 단언하고, 계약까지 뺀 변형에서 기존 동작을 본다.

변이(각각 적용 → 세 AC 테스트 실행 → scratch 백업으로 복원, 복원 후 `git diff --stat` 확인):

| 변이 | 결과 |
|---|---|
| (i) `role_values` plan-auditor → workspace-write | 010·011·013 FAIL — `role "plan-auditor" emits sandbox_mode "workspace-write" but its permission contract states "read-only"` |
| (i-b) `role_values`와 계약 둘 다 workspace-write(일관된 확장) | 013 FAIL(`emitted sandbox_mode = "workspace-write", want read-only`), 011 FAIL(`plan-auditor/write-path-scope/path-scope: must not be in the UNSUPPORTED set`). 010은 PASS — 계약과 방출이 서로 일치하므로 010이 잡을 대상이 아니다 |
| (ii-a) shell(unmeasured)을 enforced로 | 010·011·013 FAIL — `permission contract shell/deny is enforced on an unmeasured basis` |
| (ii-b) web(documented)을 생성기가 안 쓰는 필드로 enforced | 010·011·013 FAIL — `claims enforcement through field "web_enabled", which this emitter does not write` |
| (iii) `AGENTS.md.tmpl`의 `audit-verdict-file` 행 삭제 | 013 FAIL — `AGENTS.md template carries no "| audit-verdict-file |" capability row` |
| (iv) C2 `plan-auditor.md`에 반환 지시 문장 삽입 | 013 FAIL — `Claude definition carries the Codex-only return-text instruction` |
| (v) 보고가 UNSUPPORTED를 Pass로 셈 | 011 FAIL — `counted as PASS with mapping "UNSUPPORTED"` + 기대 집합 40항목 전부 missing |
| (vi) web 축 행 삭제 | 010·011·013 FAIL — `permission contract does not cover axis web` |

테스트 안의 변이(항상 실행): 축 매핑 한 개 삭제(메모리) → EmitAll이 `web`을 이름으로 거부, `mission-governor` read-only 계약 역할을 workspace-write로 방출 → 거부, unmeasured→enforced와 축 이름 변경(YAML) → ParseManifest 거부, 안 쓰는 필드 enforced(메모리) → 거부. AC-013 테스트 안: `role_values`에서 plan-auditor 삭제·sync-auditor workspace-write → 거부.

품질 게이트(HEAD `8d1480148`와 같은 내용의 작업 트리에서 측정, 증거 `.moai/reports/t1100/m6/`):

```text
$ make build                     → exit=0 (agents-emit-check·commands-emit-check·tool-policy-drift-check 선행 통과, catalog.yaml 변경 없음)
$ make agents-emit-check         → exit=0  ok  .../internal/template/agentemit
$ make commands-emit-check       → exit=0  ok  .../internal/template/commandemit
$ go test ./internal/template/... -count=1 -timeout 20m
ok  .../internal/template 187.680s / ok .../agentemit 0.996s / ok .../commandemit 0.238s
$ go test ./internal/template -run 'Neutral|Leak|AgentsDisclosure' -count=1 -v
17 PASS, 0 SKIP (TestTemplateNoInternalContentLeak, TestTemplateNeutralityAudit, TestAgentsDisclosureCompleteness 등)
$ go test ./internal/config -run '^TestCodexContractByteCeiling$' -v
internal/template/templates/AGENTS.md.tmpl = 18723 bytes (ceiling 24576, headroom 5853) — PASS
$ go vet ./internal/template/agentemit/ ./internal/template/   → exit=0, 출력 없음
$ golangci-lint run ./internal/template/agentemit/... ./internal/template/   → exit=0, 0 issues.
$ GOOS=windows go build ./internal/template/...   → exit=0
$ go test ./internal/codexwiring -count=1   → ok 2.299s
$ unset MOAI_KANBAN … MOAI_FACTORY_WORKERS && go test ./internal/cli -run '<AC-001~009 비LIVE 이름 7개>|InitCodex|CodexInit|CodexOnly|UpdateHarness|CodexContract' -count=1 -v
ok 59.694s, --- PASS 23건, FAIL/SKIP 0
```

미측정·주의:

- `make embed-check`는 exit 2 — `compared 0/12 artifacts — moai carries no embedded counterpart for` **12개 전부**(바뀌지 않은 10개 포함). 검사기가 `moai init --non-interactive`로 추출하는데(`internal/cli/doctor_agentemit_embed.go:346`) 기본 프로필이 `.codex`를 배포하지 않는 것으로 보이며, M6 내용과 무관한 추출 경로 문제로 판단한다(이 판단은 코드 판독이고 HEAD `0a635a852` 바이너리로 대조하지 않았다). 대신 바이트 탐침: `LC_ALL=C grep -a -c 'Codex-only addendum appended after the verbatim body' bin/moai` → `2`, 바뀌지 않은 `manager-git` 머리 주석 → `1`.
- 첫 lint 실행은 다른 프로세스의 golangci-lint 잠금(`parallel golangci-lint is running`, exit 3)에 걸렸고, 단독 재실행이 위 `0 issues.`다.
- `internal/cli` 전체 스위트, `.github/workflows/template-neutrality-check.yaml` CI 스크립트 자체, Codex 실제 바이너리에서의 read-only 강제(AC-DHR-012)와 부모의 원문 기록(AC-DHR-023)은 돌리지 않았다.
- REQ-DHR-013은 enforced 근거로 documented를 허용한다. 리드 지시문의 변이 (ii)는 "documented/unmeasured 근거의 enforced"를 거부 대상으로 적었지만, 이 구현은 SPEC을 따라 unmeasured만 거부하고 documented는 생성기가 실제로 쓰는 필드일 때만 허용한다. 현재 계약에 documented enforced 행은 없다.
- AC 판정 증거 파일 세 개는 HEAD `8d1480148`에서 다시 만든 것이다. 변이·게이트 증거는 같은 내용의 커밋 전 작업 트리에서 쟀다(커밋 직후 `git status --short` 0줄).

리드 요청 기록(M6 이후 추가):

- **(a) embed-check 재현.** M6 이전 커밋 `0a635a852`를 `git archive`로 scratch에 풀어 `go build`한 바이너리(87,396,386 bytes)로 HEAD `9555d3fad`에서 `make embed-check BIN=<scratch>/moai-pre-m6`를 실행 → exit 2, `compared 0/12 artifacts — moai-pre-m6 carries no embedded counterpart for: <12개 TOML 전부>`. M6 실행과 같은 모양이므로 M6가 만든 결함이 아니라 이전부터 있던 결함이다. 가설(미측정): `doctor_agentemit_embed.go:346`이 `moai init --non-interactive`로 추출하는데 그 기본 프로필이 `.codex`를 배포하지 않는다. Repair moved to card t1134 (lead-issued; cause hypothesis recorded as unmeasured).
  - **정정(M8 close-out):** develop에 원래 있던 결함이 아니다. 이 브랜치의 커밋 `8925682d2`가 원인이며, 이 카드에서 수리했다(아래 M8 Part 3 참조). 증거: t1134 워크트리의 lead/agent-36 `.moai/reports/t1134/verdict.md`. (위 (a)의 "이전부터 있던 결함" 판단은 M6 이전 커밋 `0a635a852`가 이미 `8925682d2` 뒤였기 때문에 나온 것이다 — `0a635a852`는 develop이 아니라 이 브랜치 위의 커밋이다.)
- **(b) enforced 근거 해석.** 레인 배차문은 documented 또는 unmeasured 근거의 enforced 매핑을 거부하라고 적었지만, REQ-DHR-013이 documented를 허용하므로 구현은 unmeasured만 거부한다(리드가 SPEC 우선을 확인). 현재 계약의 documented-enforced 행은 0개다.

### M7 — 결정적 4조합 카드 흐름과 판정식 (REQ-DHR-022, 024)

커밋: `00ba7cb90` — 새 테스트 파일 `internal/factorymsg/card_flow_test.go`(218줄) 하나. 제품 코드 변경 없음: `git diff --stat 9555d3fad -- internal/factorymsg/store.go internal/factorymsg/dispatch.go` 출력 0줄(`Store.Send` lane slot 멱등 조회 모양 그대로 — t1082가 그 위에 얹는다). 시작 HEAD `9555d3fad`(기록 커밋 `02a8cc438` 이후).

테스트 설계: `TestFactoryCardFlowFourCombinations`의 하위 테스트 `claude-claude`·`codex-codex`·`claude-codex`·`codex-claude`(이름은 `<lead>-<worker>`). 각 조합은 lead와 worker 두 lane을 그 backend 표지로 등록하고 `ResolveLane`으로 브로커가 표지를 기록했는지 확인한 뒤, M2 API로 `assigned → delivered → started`를 진행한다(할당 메시지 receipt 뒤 `assigned`, 추가 메시지 receipt 뒤에도 `delivered` 유지). 첫 worker가 결과를 보낸 뒤 중단되고(프로세스 비생존), `ReassignDispatch`(revoke 없이)로 attempt 2가 다른 lane에 간다. 중단된 attempt의 `StartDispatch`는 `ErrDispatchFenced`, 그 lane이 새 generation으로 돌아온 뒤 옛 endpoint의 전이는 t1112 seam(`verifyPeerOn`, 트랜잭션 안)에서 `ErrStalePeer`다. 늦은 attempt 1 결과는 attempt 2가 started이고 아직 보고하기 전에 한 번, 보고 뒤에 한 번, `integrated` 뒤에 한 번 적용해 모두 `stale`, attempt 2 결과 재전달은 `duplicate`. attempt 2 결과 메시지의 도착·claim만으로는 `started`에 머문다. `accepted` 판정 수가 정확히 1이고 기록된 결과가 attempt 2의 것인지 끝에서 확인한다. 이 테스트는 증거 파일을 쓰지 않으므로 `MOAI_T1100_EVIDENCE_DIR`로 skip하지 않는다(리드 확인).

AC 판정(명령은 `acceptance.md` 그대로, 테스트 커밋과 같은 내용의 작업 트리에서 실행):

| AC | 판정 출력 | 증거 파일 |
|---|---|---|
| AC-DHR-017 | `true` | `.moai/reports/t1100/ac017.jsonl` |
| AC-DHR-019 (명령 1, 판정식 6종) | `true` | 명령 자체(합성 입력) |
| AC-DHR-019 (명령 2, 증거 채널) | `true` | 명령 자체(`mktemp -d`, 실행 뒤 삭제) |
| AC-DHR-014 (회귀) | `true` | `.moai/reports/t1100/ac014.jsonl` |
| AC-DHR-015 (회귀) | `true` | `.moai/reports/t1100/ac015.jsonl` |
| AC-DHR-016 (회귀) | `true` | `.moai/reports/t1100/ac016.jsonl` |

AC-DHR-018은 LIVE(M8)라 돌리지 않았다 — `NOT_RUN`.

RED: M2 코드에서 이 흐름은 처음부터 동작한다 — 첫 실행의 실패(`card_flow_test.go:175: claim identity mismatch` ×4)는 테스트 쪽 결함이었다(receipt 뒤 `ReadBody`로 재적용하려 함). 재전달을 "claim 시점에 읽은 `ResultReport` 재사용"으로 고친 뒤 4/4 PASS. 따라서 RED는 변이로 보인다(각 변이 적용 → AC-017 명령 → scratch 백업으로 `dispatch.go` 복원, 복원 후 sha256 `87028a35…8196105` 원본과 같음):

| 변이 | AC-017 테스트 출력(원문) | AC-017 판정 |
|---|---|---|
| (i) `ApplyResult`에서 (5) duplicate/collision 검사와 (7) state 검사 제거 — 한 dispatch에 결과가 두 번 적용됨 | `card_flow_test.go:177: attempt 2 redelivery: outcome "accepted", want "duplicate"` ×4 조합 | `false` (`m7/mutant-i-applied-twice.jsonl`) |
| (i-b) (2)(3) attempt·lane 검사 제거 — 늦은 attempt 1 결과가 먼저 적용됨 | `card_flow_test.go:155: late attempt 1 before attempt 2 reports: outcome "accepted", want "stale"` ×4 | `false` (`m7/mutant-ib-late-accepted.jsonl`) |
| (ii) 판정식이 SKIP을 pass로 셈 — AC-019 명령 1의 `P`에서 `.Action=="pass"`를 `(.Action=="pass" or .Action=="skip")`로, fail 절에서 skip 제거 | 전부 skip 흐름 `b=true` | AC-019 명령 1 `false` |
| (ii-b) AC-017 판정식에 합성 입력: 하위 테스트 하나 skip + 부모·패키지 pass | — | `false` (`m7/synthetic-ac017-one-skip.jsonl`) |
| (ii-c) AC-017 판정식에 패키지 `ok` 줄과 패키지 pass만 | — | `false` (`m7/synthetic-ac017-package-ok-only.jsonl`) |

품질 게이트(같은 작업 트리):

```text
$ go test ./internal/factorymsg/... -count=1   → exit=0  ok .../internal/factorymsg 15.699s (m7/factorymsg-full.txt)
$ go vet ./internal/factorymsg/                 → exit=0, 출력 없음
$ golangci-lint run ./internal/factorymsg/...   → exit=0, 0 issues.
```

미측정·주의:

- 증거 파일(`.moai/reports/`)은 `.gitignore:235`로 미추적이다. 판정을 가른 명령과 출력은 위 표에 옮겨 적었다.
- 조합의 backend 표지는 브로커 `peers.backend` 열의 값일 뿐이다. 결정적 흐름은 backend에 따라 분기하는 코드를 지나지 않으므로, 네 조합이 서로 다른 경로를 검사한다고 주장하지 않는다 — 실제 CLI 차이는 AC-DHR-018(LIVE)의 몫이다.
- `internal/cli`, 다른 패키지는 돌리지 않았다(변경 없음).

### M8 — LIVE 판정(AC-DHR-012, 018, 023)과 run-phase 마감

측정 도구: codex-cli 0.156.1. 증거는 `.moai/reports/t1100/`(`.gitignore`로 미추적) — 판정을 가른 값은 아래에 옮겨 적었다.

**인증 확인.** 실행 전 인증 확인(codex `Logged in using ChatGPT`, claude `loggedIn: true`)은 리드가 실행 전 수행한 것으로 전달받았고, 그 출력 파일은 증거 디렉터리에 없다. 마감 시점에 다시 쟀다: `codex login status` → `Logged in using ChatGPT`, `claude auth status` → `"loggedIn": true`. 판별 탐침은 실행 전후 `auth.json` sha256이 같다(`m8-sbx/summary.json` `auth_sha256_before` = run1·run2 `auth_sha256_after` = `eb7c45bd…6630`). 잔존 프로세스: 마감 시점 `pgrep -fl "codex exec"`, `pgrep -fl "TestFactoryLive|TestCodexRoleLive"` 모두 출력 없음.

**예산 원장** (호출 수·경과 초는 증거 JSON에서 읽음; 어느 실행도 `aborted: true`가 아니다):

| 실행 | 호출/예산 | 경과 | 결과 | 증거 파일 (sha256은 아래 전체 목록) |
|---|---|---|---|---|
| AC-012 + AC-023 (같은 실행) | 14/14 | 289.25 s | 역할 로드 12 + 감사 쓰기 시도 2 — ABORTED 아님 | `ac012-evidence.json`, `ac023-evidence.json`, `ac012-live.jsonl` |
| AC-018 claude-claude **INVALID** | 6/8 | 128.54 s | 19:58:49 실행 — 테스트 수정 커밋 `755ae6576`(19:59:40) **이전**. `error: result … body lacks the run nonce`. 판정에 쓰지 않음, 보존만 | `ac018-evidence-claude-claude.invalid-pre-755ae6576.json`, `…invalid-pre-755ae6576.jsonl` |
| AC-018 claude-claude 재실행 | 6/8 | 144.59 s | PASS (`late_result_outcome: stale`, `applied_results: 1`, 정리 pid 7) | `ac018-evidence-claude-claude.json`, `.jsonl` |
| AC-018 codex-codex | 6/8 | 196.45 s | PASS (정리 pid 10) | `ac018-evidence-codex-codex.json`, `.jsonl` |
| AC-018 claude-codex | 6/8 | 155.80 s | PASS (정리 pid 9) | `ac018-evidence-claude-codex.json`, `.jsonl` |
| AC-018 codex-claude | 6/8 | 147.41 s | PASS (정리 pid 8) | `ac018-evidence-codex-claude.json`, `.jsonl` |
| 판별 탐침 (sandbox 상속) | 2/4 | 25.8 s + 23.6 s | 아래 AC-012 참조 | `m8-sbx/summary.json`, `m8-sbx/invocations.log` |

전체 sha256(64자)은 `shasum -a 256`으로 이 마감 시점에 쟀다: `ac012-evidence.json` c7970997a4d4ed0679d166faf7565f74d342d16fa0ae71f1cf2ffcd0cb7da614 · `ac023-evidence.json` 5b8c22fbb58dce93b248d5a968a86aae69ba455ba70664324bb7559e0f5936c8 · `ac012-live.jsonl` 2297362e5cc8e57de78ae8bb5155ab73d617a1be24af332fde7aafad98f93cf1 · `ac018-evidence-claude-claude.invalid-pre-755ae6576.json` 119017da7b00d8089aa3f5a849d1646f37b4fe857d0c9a71668b24803afee95f · `ac018-live-claude-claude.invalid-pre-755ae6576.jsonl` 78638bca060fe4befafcb779d2ccb9277989e3e1581421d0b86abd5125ef629d · `ac018-evidence-claude-claude.json` 9c3255f79bdf8b6cebf357c276db0a64d413a3791d9fb1359b28f2782246e875 · `ac018-evidence-codex-codex.json` 2eab83babe1ce24f3bb939ebaac2c3c59798ddf382828eb99438af01d96dbfc4 · `ac018-evidence-claude-codex.json` 8aabecd1f9b0ffc7421db045bf78283d3b13a635f2e13be22695b3bc29a0218d · `ac018-evidence-codex-claude.json` 2e4cd907e5dbefe8ead53b687d562876c4dccffc4b49a24b622703915190ac48 · `ac018-live-claude-claude.jsonl` 941407ae4affb735ce33e6aef0922a9cbc54d80aad5531052281235d38a98913 · `ac018-live-codex-codex.jsonl` a9a2864f721fe86ee6bd33febd997062d6565a6e19035968617a4ee7ea861f0d · `ac018-live-claude-codex.jsonl` 8af3423049bc411bde6fd906774374b13068998c0eb474c63dec873d25b0cf8d · `ac018-live-codex-claude.jsonl` 46c9ad8d9ee08c6f9a2b7d6e0cb2aecb5891def371c5267f4ffb9585d4cf1b32 · `ac018-evidence.sha` 0c98285775bcf02df7d7c3b7d90888346ae3a8be0b73ec7986e328772d9eb156 · `m8-sbx/summary.json` ee41c9e8701f265296d1bd686c636863b87ffc7a6fba892b7914e7a929b47cc9 · `m8-sbx/invocations.log` 366dc091707b39ab268a9183a405399fb659e3ac6dae91c703c5040b1c5de623.

**AC-DHR-012 — FAIL (알려진 FAIL, t1143으로 이월).** `ac012-evidence.json`: `invocations: 14`, `aborted: false`, 두 쓰기 시도 모두 `denied: false`, `probe_exists: true` → 판정식의 마지막 절(`denied==true and probe_exists==false` 2건)이 성립하지 않는다. 원인은 측정됨 — spawn된 하위 에이전트는 부모 세션의 sandbox를 물려받고 역할 TOML의 `sandbox_mode`는 적용되지 않는다(좁히지도 넓히지도 않음). 판별 탐침(`m8-sbx/summary.json`):
- 탐침 1(run1): 부모는 config로 `sandbox_mode = "workspace-write"`(`-s` 없음), `plan-auditor`(TOML read-only) spawn → 하위 세션 `sandbox_policy.type = workspace-write`, final `VERDICT plan-auditor write=allowed`, 루트에 `probe-plan-auditor.txt` 생성.
- 탐침 2(run2): 부모 `-s read-only`, `manager-docs`(TOML workspace-write) spawn → 하위 세션 `sandbox_policy.type = read-only`, final `VERDICT manager-docs write=denied`, 루트에 탐침 파일 없음(`.codex`만).
보안 관련: "Codex 감사 역할은 read-only"라는 전제가 런타임에서 성립하지 않는다. 이 카드는 Part 1 커밋 `17bfceaa8`로 그 약속을 코드·템플릿에서 거뒀고(계약 sandbox 행 `UNSUPPORTED`/`measured`), 최상위 read-only 실행 경로는 t1143의 몫이다.

**AC-DHR-023 — 충족되지 않음(t1143으로 이월).** `ac023-evidence.json`의 해시는 일치한다: plan-auditor `returned_sha256` = `verdict_file_sha256` = `efd4ddda07b4758e77a76915f4261ff9fd4f4ea61e92fb07c5162baadd815ca3`, sync-auditor 둘 다 `6101dc9c740d0b00e15e360cb70ed1b83351360eda88ef62eb83ab7d14a5a2fc`, 둘 다 `returned_contains_nonce: true`. 그러나 두 항목 모두 `write_denied: false`이고 반환문이 `write=allowed`다 — 감사자가 판정 파일을 직접 썼으므로, 이 AC가 전제하는 "read-only 감사자가 반환하고 부모가 쓴다" 경로는 실행되지 않았다. 해시 일치는 부모 기록 경로의 증거가 아니다.

**AC-DHR-018 — PASS.** `acceptance.md`의 네 조합 합산 판정식을 그대로 실행(`ac018-evidence.sha`와 `ac018-evidence-all.json`을 다시 씀):

```text
$ shasum -a 256 …ac018-evidence-{claude-claude,codex-codex,claude-codex,codex-claude}.json > .moai/reports/t1100/ac018-evidence.sha && jq -s . … > …ac018-evidence-all.json && jq -se … (acceptance.md 원문) …
true
exit=0
```

INVALID 실행 증거 파일(`…invalid-pre-755ae6576.*`)은 판정식 입력에 들어가지 않는다(판정식이 조합 이름 4개 파일만 연다).

**Part 1 — 코드·템플릿 주장 정정 (커밋 `17bfceaa8`).** 권한 계약의 `sandbox/mode` 행을 `enforced`에서 `UNSUPPORTED`(basis `measured`, 사유: 하위 에이전트의 sandbox 상속)로 바꿨다. 생성된 `sandbox_mode`가 계약 값과 같아야 한다는 검사는 행의 매핑과 무관하게 돌도록 `enforced` 분기 밖으로 옮겼다(그래야 AC-DHR-010 변이 "read-only 계약 역할을 workspace-write로 내보냄"이 계속 거부된다). `sandbox_mode`는 계속 방출한다 — 필드는 수용되고 무해하며, AC-DHR-013이 `read-only` 방출을 요구한다. 감사 역할 addendum과 AGENTS.md `audit-verdict-file` 행은 read-only 약속을 빼고 상속 동작을 적었다(부모가 반환문 그대로 판정 파일을 쓰는 지시는 유지). 재측정(HEAD `17bfceaa8`와 같은 내용의 작업 트리, 명령은 `acceptance.md` 원문): AC-DHR-010 `true`, AC-DHR-011 `true`, AC-DHR-013 `true`. `make agents-emit-check` ok, `make commands-emit-check` ok, `make build` exit 0, `go test ./internal/template/agentemit/...` ok, `go test ./internal/template/` ok(91.1 s — `TestTemplateNoInternalContentLeak`, `TestTemplateNeutralityAudit` 포함), `go vet ./internal/template/...` 출력 없음, `golangci-lint run ./internal/template/...` 0 issues. 루트 `AGENTS.md`에는 `audit-verdict-file` 행이 없다(M6에서 의도적으로 뺐다 — 이 저장소는 `.codex/agents`를 추적하지 않는다).

**Part 3 — embed-check 수리 (커밋 `d819443c2`).** 원인은 `8925682d2`의 두 변경이다: 기본 `moai init --non-interactive`가 `.codex/agents/moai`를 배포하지 않게 됐고, 배포기가 Codex 역할 파일 안의 Claude 트리 참조를 다시 쓴다(`.claude/skills/` → `.agents/skills/`, `.claude/rules/moai/` → `.moai/policies/`). 추출 명령에 `--llm both`를 붙이자 12/12 비교까지는 갔지만 12개 모두 "stale"로 나왔고(정규화 차이), 커밋본을 배포기와 같은 정규화(`template.NormalizeCodexRoleForDeploy`, 신규 공개 래퍼)를 거친 뒤 비교하도록 고쳤다. RED: `TestExtractEmissionViaInit_RequestsCodexDeployment` → `extraction did not request the Codex deployment (--llm both): stat …/extract/.codex/agents/moai/manager-git.toml: no such file or directory`; `TestAgentEmitEmbed_DeployNormalizationIsNotDrift` → `message = "moai embeds stale agent-emit artifacts (2/2 compared): builder-harness.toml, manager-docs.toml", a normalization-only difference must not read as drift`. GREEN 후 `make build` → `make embed-check`: `ok Agent Emit Embed 12/12 embedded agent-emit artifacts match the committed set (moai)`, `Pass 1 Warn 0 Fail 0`. 변이 확인: 커밋본 `sync-auditor.toml`에 한 줄을 덧붙이고 `make embed-check` → exit 2, `moai embeds stale agent-emit artifacts (12/12 compared): sync-auditor.toml`; 파일은 scratch 백업으로 복원(`git status` 변경 없음).

미측정·주의:
- 정규화 뒤 비교이므로, 커밋본과 임베드본이 **정규화 대상 토큰에서만** 다른 경우(예: 한쪽 `.claude/skills/`, 다른 쪽 `.agents/skills/`)는 embed-check가 잡지 못한다.
- LIVE 재실행은 하지 않았다(이 마감의 지시: 모델 호출 없음). AC-012/023의 판정은 위 증거 파일의 기존 값이다.
- `internal/cli` 전체 스위트는 돌리지 않았다 — embed 관련 테스트만(`AgentEmitEmbed|ExtractEmission|BoundedTail|FindEmbedCheckRoot|NearestProjectRoot`, ok).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## Revision iter-4 (delta)

plan-audit iter-3(FAIL 0.89, `.moai/reports/plan-audit/SPEC-DUAL-HARNESS-RECOVERY-001-review-3.md`)의 blocking 결함 N1과 선택 결함 N3·N4·N7, §E 리드 조정 항목 ①만 고친 좁은 델타다. 아래 줄 외에는 바꾸지 않았고, REQ·AC 수(25·23)와 모든 jq 판정식은 그대로다.

| 항목 | 바뀐 줄 | 내용 |
|---|---|---|
| N1 | spec.md:164 (REQ-DHR-023) | "reaches its invocation or time budget" → 예산을 **넘게 될 때**(9번째 호출이 필요하거나 900초 초과) 그 호출 전에 중단·`ABORTED`, 정확히 8회로 끝난 조합은 `ABORTED` 아님 |
| N1 | plan.md:116 | "예산 도달 시 중단" → "예산 초과 시 … 그 호출 전에 중단", 예산과 같은 호출 수로 끝난 실행은 `ABORTED` 아님 |
| N1 | acceptance.md:213 (AC-DHR-012) | "14에 닿으면" → "15번째 호출이 필요해지면", 정확히 14회로 끝난 실행은 `ABORTED` 아님 (판정식 `$e.invocations==14` 그대로) |
| N1 | acceptance.md:312 (AC-DHR-018) | "예산에 닿은 조합" → "예산을 넘게 된 조합(9번째 호출 또는 900초 초과)", 8회 이하·900초 이하는 `ABORTED` 아님 (판정식 `<=8`·`<=900` 그대로) |
| §E ① | spec.md:237 | 리드 조정 항목 1을 "① 해소 — t1082 `745ae0e6d` (v0.5.2): REQ-FLH-009(t1082 spec.md:103), AC-FLH-008(t1082 acceptance.md:152-158)이 같은 recipient generation 안의 계약을 담음, 칸반 리드가 커밋에서 확인"으로 바꿈. 항목 2(분기 A 조건부)는 그대로 |
| N4 | spec.md:20, :208, :237 | t1082 SPEC 상태 서술 "unmerged draft"/"미병합 draft" → 미병합, t1082 브랜치에서 status in-progress (v0.5.2, `745ae0e6d`) |
| N3 | progress.md:98 (§E.2) | AC-DHR-020의 `outcome`, 측정 커밋 SHA, 증거 sha256을 값으로 §E.2에 적는 기록 의무 한 줄 추가(새 AC 없음). §E.2 자리표시는 그대로 |
| N7 | acceptance.md:412 (AC-DHR-023) | 반환문 부재 시 출력에 `NOT_RUN`을 찍지 않고 `ac023-evidence.json`의 `"not_run": true`로만 기록 → 같은 jsonl을 읽는 AC-DHR-012 판정식과 분리 |

손대지 않은 것: N2·N5·N6(범위 밖), plan.md:132 위험 표의 "예산 도달 시 `ABORTED`"(N1과 같은 표현이지만 이번 델타의 지정 줄이 아니어서 보고만 함).

N3 위치 이동(후속 커밋): 위 N3 행이 progress.md:98(§E.2)에 넣은 기록 의무 문장을 acceptance.md §D 완료 정의 목록(436행 바로 다음 줄)으로 옮겼다. 위치만 바꿨고 문구는 그대로다. 제자리에서 읽히도록 끝의 참조 "이 절에"만 "`progress.md` §E.2에"로 고쳤다. §E.2는 run 단계 증거 절(manager-develop 소관)이라 plan 단계 의무를 두지 않는다. §E.2 자리표시 `_<pending run-phase>_`는 그대로다.
