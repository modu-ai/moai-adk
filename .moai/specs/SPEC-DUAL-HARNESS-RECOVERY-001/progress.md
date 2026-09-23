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
