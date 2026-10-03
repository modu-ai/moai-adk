# t1452 겹침 판정서 — 병합 창 대기·보유 시간 단축

- 카드: t1452 (Class C·Tier M) · 판정 단계: 겹침 판정 (plan 이전)
- 판정자: lane-20 (워크트리 `.claude/worktrees/t1452`, 브랜치 `WT-merge-window-hold-time`, 기점 develop `2b9e4a4d0`)
- 대조 대상(전부 읽기 전용으로 대조, 각 브랜치의 현재 HEAD 기준):
  - t1479 `SPEC-MERGE-WINDOW-QUEUE-001` v0.7.0 (status: draft) — `WT-merge-window-queue` @ `98e0296cb`
  - t1478 `SPEC-CANDIDATE-CI-001` — `WT-candidate-ci` @ `838f73226`
  - t1453 `SPEC-GITHUB-FLOW-DEFAULT-001` — `WT-github-flow-default` @ `53e3f7ebb`

## 결론

세 갈래 중 **(a)와 (b)는 t1479(+t1478)가 SPEC 본문으로 이미 덮고 있어 남는 범위가 없다.** **(c) 동일 테스트 명령 재실행 억제만 어느 SPEC 도 덮지 않는다.** 고유 범위 = (c) 하나. (a)·(b)는 리더가 t1479 로 흡수 처리하면 된다.

## 범위별 대조

| t1452 범위 | 덮는 조항 | 판정 |
|---|---|---|
| (a) 지명 없이 merge-ready 레인이 acquire | t1479 REQ-MWQ-002(`--wait` FIFO 대기)·006(open 정책 자동 승격)·012(policy open/hold, 리더는 정책만)·013(지명 조항 삭제, 문장 고정) | 덮임 |
| (a) 조건 3종(sync-audit PASS·충돌 없음·트리 항등)과 `moai factory merge` 게이트 정합 | t1479 REQ-MWQ-017(카드 게이트·트리 항등·ancestry)·019(complete 게이트 순서)·021(merge-ready 점검에 기록 유효성 4번째 조건); t1478 REQ-CCI-002(충돌 시 중단)·011(착지 전 공용 landing check) | 덮임 |
| (a) 리더 검토를 병합 후 push 전으로 이동 | t1478 REQ-CCI-012/013(`moai integration push` — 임계·적색 보류·pinned SHA push), t1479 §B(리더는 open/hold 정책과 push 전 최종 읽기만) | 덮임 |
| (b) 흡수 트리 재측정을 창 획득 전에 수행 (merge-tree 사전 점검) | t1479 REQ-MWQ-014(재측정은 대기열 진입 전, 후보 트리 기준, 트리 SHA 키 기록)·016; t1478 REQ-CCI-001(워킹트리 변경 없이 would-be merge 계산) | 덮임 |
| (b) 창은 병합 커밋 동안만 보유 | t1479 REQ-MWQ-017(창 안은 트리 항등 확인 + `git merge --no-ff <pinned SHA>` 뿐, 테스트 스위트 실행 금지) | 덮임 |
| (c) 동일 테스트 명령 재실행 억제(결과 캐시 또는 재사용 규율) | t1479 REQ-MWQ-014/015 는 **창 게이트용 기록**(트리 SHA 키·명령·exit·테스트 수)만 규정하고, 같은 (트리, 명령) 에 유효 기록이 있을 때 재실행하지 않는다는 조항이 없다. t1478 REQ-CCI-006 의 재사용은 후보 push 에 한정. t1453 은 해당 없음. | **미덮임 — 고유 범위** |

t1453 과의 관계: REQ-GFD-004~006 은 github-flow 전환 후 카드 전달을 PR 로 돌리고 로컬 병합 창을 전달의 선행 조건에서 뺀다. 따라서 t1452 의 (a)(b)는 전환 전에만 가치가 있는 범위이며(전환은 배치 경계에서만 이뤄지므로 그 전까지는 유효), t1453 이 (a)(b) 의 고유 범위를 새로 만들지는 않는다. (c) 는 PR·CI 경로로 가도 레인 로컬 반복 실행은 남으므로 t1453 에도 걸리지 않는다.

## (c) 고유 범위 — 계획으로 올릴 후보

- 목적: 같은 트리·같은 명령·같은 환경의 반복 테스트 실행 제거. 카드 본문 실측: go test 711회 중 동일 명령 반복 약 286회, t1414 의 ./internal/cli/... 30분짜리 2회.
- 후보 재료: 저장소에 이미 `internal/verify`(`Key()` = HEAD SHA + status digest + diff hash + untracked 해시 → 트리 상태 키, `receipt.go`·`store.go`)와 MCP `verify_snapshot` 이 있다. 재사용 가능성은 있으나 **이 판정에서 읽은 것은 `key.go` 머리뿐**이며, 임의 테스트 명령 결과를 키 아래 저장·재사용하는 경로가 이미 있는지는 plan 단계 조사 항목이다.
- 접점(충돌 주의): 재측정 verb 는 t1479 REQ-MWQ-015/016 이 새로 만드는 표면이다. (c)를 그 verb 에 얹으려면 t1479 착지 후이거나 t1479 의 한 조항(REQ-MWQ-014 보강)으로 흡수하는 쪽이 파일 충돌이 적다. 레인 규율 쪽(`gitflow-lane-protocol.md`·`kanban-dispatch-mechanics.md` 검증 절) 문장은 t1479 REQ-MWQ-013·022 와 같은 파일을 건드린다.

## 권고 (리더 판단 요청)

1. (a)(b) → t1479 로 흡수(리더 처리). t1452 본문에서 (a)(b) 삭제.
2. (c) → 선택지: (i) t1479 에 REQ 하나 추가해 흡수, (ii) t1452 를 (c)만으로 축소해 Tier S 로 plan → run. (ii)면 t1479 착지 순서와 파일 겹침(재측정 verb·doctrine 문서)이 직렬 조건이 된다.
3. 이 레인은 별도 회신이 없으면 (ii) 가정으로 (c) 한정 plan 조사(기존 `internal/verify` 재사용 가능성)부터 진행한다. push 는 하지 않는다.

## Claim / Evidence / Baseline / Gaps / Residual-risk

- **Claim**: 위 대조표의 "덮임" 판정은 SPEC 본문 조항 문구 기준이다.
- **Evidence**: t1479 spec.md 전문(306행)을 읽음; t1478 `REQ-CCI-001~025` 조항 첫 줄을 grep 으로 열거, 재실행·캐시 관련 키워드 grep 은 REQ-CCI-006·022 만 적중; t1453 spec.md 118~147행(REQ-GFD-002~009)을 읽음; develop `internal/verify/key.go` 머리 40행을 읽음.
- **Baseline-attribution**: 각 SPEC 의 읽기 시점 브랜치 HEAD 는 위 SHA. develop 기점 `2b9e4a4d0` 은 `git rev-parse --short develop` 실측. 카드 본문 수치(21.5h·711회·286회 등)는 카드 본문 인용이며 이 판정에서 재측정하지 않았다.
- **Gaps**: ① t1478 spec.md 는 전문이 아니라 조항 첫 줄 + 키워드 grep 만 읽었다 — (c) 에 해당하는 조항이 본문 깊은 곳에 있을 가능성은 낮지만 0 이라 단정하지 않는다. ② t1453 spec.md 는 118~147행 외 미독. ③ t1479·t1478 의 "덮임"은 **SPEC(draft) 기준**이며 구현·착지 아님(t1479 는 plan-audit 반복 중, status draft). ④ `internal/verify` 의 임의 명령 결과 재사용 능력은 미검증.
- **Residual-risk**: t1479 가 추가 감사 반복으로 범위가 또 줄면 (a)(b) 일부가 다시 t1452 로 돌아올 수 있다(예: REQ 삭제). 판정은 v0.7.0 본문 기준이다.

## 리더 판정 접수 및 (c) 조사 결과 (추가)

- 리더 판정: (ii) — (a)(b)는 t1479·t1478이 흡수한 것으로 기록, t1452는 (c) (트리, 명령) 재측정 억제만 Tier S로 plan→run→sync. t1479 착지 전에는 그쪽 표면 파일(internal/kanban, internal/cli integration·factory complete/merge, internal/homestate, internal/factorylane, AGENTS.local.md, gitflow-lane-protocol.md, kanban-dispatch-mechanics.md)을 건드리지 않는다. 판정서는 `git add -f` 로 커밋에 포함한다.
- 흡수 기록: t1452 (a) → SPEC-MERGE-WINDOW-QUEUE-001 / SPEC-CANDIDATE-CI-001 이 흡수. (b) → 동일. 이 카드의 범위에서 제외.
- (c) 조사 (develop `2b9e4a4d0` 트리에서 읽음):
  - `internal/verify/key.go` `Key()` = HEAD SHA + `status --porcelain=v2`/`diff HEAD`/untracked 내용 해시 → 트리 상태 키.
  - `internal/verify/receipt.go` 는 (Head, TreeDigest, ConfigDigest, Command, ToolVersion) 5필드 일치 + TTL 로 재사용 판정(`CheckReceipt`), `schema.go` `FindCommand` 는 바이트 단위 명령 일치.
  - CLI: `moai verify record` (명령·exit 기록, 같은 명령은 덮어씀) / `moai verify check --key-current` (키 기준 신선도, exit 0/1). 두 단계 수동 조합이라 "실행하거나 재사용" 단일 동사가 없고, `check` 는 `--command` 로 명령을 지정하지 않는다(키·check-id 기준).
  - 소비처 문서화 범위: 기록된 소비자는 sync Stop 훅·sync-audit-4dim·sync-auditor 뿐(`.claude/rules/moai/workflow/snapshot-consumer-contract.md`). 레인의 카드 내 반복 `go test` 는 이 경로를 쓰지 않는다.
  - 결론: 저장·키·신선도 부품은 이미 있다 → (c) 의 최소 구현은 **run-or-reuse 단일 verb** + **레인 규율 문장** 이다. 새 저장소는 만들지 않는다.
- 미검증(이 조사에서): 레인이 실제로 `moai verify record/check` 를 쓰는지의 사용 이력(소비처는 문서·코드 grep 으로만 확인), 환경변수(GOFLAGS 등)가 키에 반영되지 않는 위험(t1413 계열) 의 정확한 범위.

## plan-audit 결과와 리더 처분 (추가)

- 1회차: FAIL 0.76 (`plan-audit-iter1.md`, D1~D4 차단). 수리 후 2회차: **FAIL 0.88** (`plan-audit-iter2.md`, 감사관 claude-sonnet-5-5, 차단 결함 N1 1건 — D1~D4 전부 해소 확인). 감사 상한 2회 도달.
- 리더 결정 (미션 계약 11c79e1a, 수렴 규칙): **N1 해소 실측으로 PASS-WITH-DEBT 처분, 선택 2건 반영.** 감사는 다시 열지 않는다.
- N1 해소 실측 (트리 HEAD `2b9e4a4d0`, SPEC 0.2.1):
  - 수정 전(대조): `grep -n -F "--env" AGENTS.md; echo "exit=$?"` → `ugrep: invalid option --env ...` / `exit=2`
  - 수정 후: `grep -n -F -e "--env" AGENTS.md; echo "exit=$?"` → stdout 없음 / `exit=1` (정당한 RED-now: 구현 전 매치 없음)
  - 수정 후 템플릿: `grep -n -F -e "--env" internal/template/templates/AGENTS.md.tmpl; echo "exit=$?"` → stdout 없음 / `exit=1`
  - 양성 대조: `grep -n -F -e "--env" .moai/specs/SPEC-VERIFY-RUN-REUSE-001/spec.md | head -2` → 2줄(spec.md:29, :50) 출력, 파이프 앞 exit 0 확인을 위해 별도 실행 시 exit=0 (spec.md 에 `--env` 존재)
  - `moai spec lint spec.md` → `✓ No findings — all SPEC documents are valid`
- 반영한 선택 2건: AC-005 서브테스트 분리(PASS 줄 4), REQ-VRR-008 "재사용 결과는 Gap, Claim 아님" 문구.
- Gaps: go test RED-now 의 exit 코드는 `ok`/`PASS` 출력 모양으로 읽음(직접 exit 미관측), 2회차 이후 수정본은 감사관이 재독하지 않음(PASS-WITH-DEBT 의 채무).

## sync-audit 경과와 후속 카드 문안 (추가)

- sync-audit 1차: **FAIL 62** (`sync-audit.md`) — F0(P1) `moai verify run` 이 실제 CLI 에 미등록(verify.go `init()` 이 verify_run.go `init()` 보다 먼저 돌아 `verifyExtraCommands` 가 빈 채 순회), 테스트 16개가 루트 명령을 거치지 않아 놓침. 수리 `a05ec62de` (직접 등록 + 루트 명령 경유 테스트 3개 + F6 프로세스 그룹 kill), 증거 `28843a6d1`. 레인 직접 측정(HEAD 28843a6d1): 빌드한 바이너리 `verify run -- echo hi-t1452` 1회차 miss 후 실행 exit 0, 2회차 `reuse key=28843a6d1…:6e340b9cffb37a98` exit 0·미실행; 신규 3테스트 PASS 3줄 exit 0.
- 리더 결정: 수리분(F0·F6)과 증거에 한해 sync-audit 델타 1회 허용(상한 연장 승인). 치명적이지 않은 지적만 남으면 PASS-WITH-DEBT 처분. 결과는 `sync-audit-delta.md`.
- 채무: F1 — Ctrl-C 시 자식 명령이 계속 실행됨(Setpgid 로 별도 그룹, 신호 전달 없음). Windows 실행 미관측. TDD 순서는 커밋 그래프로 증명되지 않음(M1·M2 테스트와 구현 동일 커밋).

### 후속 카드 문안 (리더가 큐에 올림)

- **제목 후보**: `moai verify sync-gate` 미등록 선행 결함 — verify 하위 명령 등록 순서 의존 제거
- **본문 초안**: develop 기점 `2b9e4a4d0` 에서 빌드한 바이너리의 `moai verify sync-gate --help` 가 sync-gate 도움말이 아니라 부모 `verify` 도움말을 출력한다. 원인: `internal/cli/verify.go` 의 `init()` 이 `newVerifyCmd()` 를 호출해 `verifyExtraCommands` 를 순회하는데, `verify_receipts.go`(sync-gate)·codex-review 등록용 `init()` 은 파일명 순서상 그 뒤에 실행돼 슬라이스가 빈 채로 순회된다(Go 는 같은 패키지의 `init()` 을 파일명 순으로 실행). t1452 가 같은 결함으로 `verify run` 을 직접 등록으로 고쳤다(`a05ec62de`). 범위: (1) sync-gate·codex-review 를 순서 비의존 방식으로 등록 (2) 루트 명령을 경유해 하위 명령이 해석되는지 검사하는 가드 테스트 — 모든 `verifyExtraCommands` 항목이 `rootCmd.Find` 로 해석되고 부모 자신이 아님을 단언 (3) Codex Stop 체인이 sync-gate 를 호출하는 경로가 실제로 동작하는지 실측(그동안 조용히 도움말만 받았을 가능성 — 호출부 확인). 완료 기준: 빌드한 바이너리의 `moai verify sync-gate --help` 가 sync-gate 전용 도움말, 가드 테스트 적색→초록. 근거: t1452 `.moai/reports/t1452/sync-audit.md` F0 및 progress.md §E.2. Class B(원인 확정, 수리 범위 소형).
- 미검증: sync-gate 가 다른 경로(훅 직접 호출 등)로 이미 호출되고 있는지, 언제부터 미등록인지(이력)는 조사하지 않았다.

### sync-audit 델타 결과 (추가)

- `sync-audit-delta.md`: **PASS-WITH-DEBT 90/100**, 차단 지적 없음 (감사관 claude-sonnet-5-5, audited_sha 28843a6d1868616e67605086540c012b37b10146). F0·F6 해소를 실제 바이너리(miss→reuse, exit 3 통과)·신규 테스트 3개 PASS·변이 2건 적색 입증으로 확인. 채무: F1(Ctrl-C), F7(P3: 같은 init 순서 결함이 `verify sync-gate`뿐 아니라 `codex-review`·`audit-plan` 등록에도 해당할 수 있음 — 개별 미조사), F8(P3: F6 테스트의 고정 3초 sleep).
- 영수증: 델타 감사는 교차 모델 감사 도구를 호출하지 않았다(1차 델타 시도가 55분 무응답으로 중단돼, 재시도 지시에서 호출 금지를 명시). Stop 훅이 `AUDIT_RECEIPT_VIOLATION` 을 냈고 verdict 줄은 `receipts=none`. **우회하지 않고 리더에 보고**한다.
- 후속 카드 문안 보강(F7): 범위 (1)에 `codex-review`·`audit-plan` 등록도 포함해 각각 빌드 바이너리로 `--help` 가 전용 도움말인지 실측한다.

### 영수증 처분 (리더 결정)

- 영수증 없음 — 리더 수용, 미션 계약 11c79e1a, 영수증 저장소 결함. 델타 감사 `sync-audit-delta.md` 의 직접 실측 기반 PASS-WITH-DEBT 90 을 리더 결정으로 수용한다. 이 처분은 감사관 verdict 를 바꾸지 않으며(verdict 줄은 `receipts=none` 그대로), 병합 근거는 감사관 실측 + 리더 수용이다.
- F7 후속 카드 문안(위 "후속 카드 문안" 절에 병합해 읽을 것): `verifyExtraCommands` init 순서 결함이 `moai verify sync-gate` 뿐 아니라 `codex-review`·`audit-plan` 등록에도 해당할 수 있다(`verify_receipts.go:16`, `audit_plan_cmd.go:46`, `codex_review_receipt.go:177` 가 같은 init-append 방식). 각 verb 를 빌드한 바이너리로 `--help` 실측해 전용 도움말인지 확인하고, 순서 비의존 등록 + 루트 명령 경유 해석 가드 테스트로 고친다.
