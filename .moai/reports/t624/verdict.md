# t624 레인 판정 기록 — SPEC-SYNC-GATE-FAILSTATE-001

- 카드: t624 (Tier M, factory lane-10)
- 브랜치: `WT-sync-gate-failstate` (워크트리 `.claude/worktrees/t624`, 미푸시)
- 기록 시점 HEAD: `89f3d5eb7`
- 성격: 병합 전 레인 측 증거 요약이다. 최종 PASS/FAIL 판정과 부채 처분은 리드가 이 파일과 원본 증거를 읽고 내린다.

## 1. 주장 (Claim)

| 항목 | 결과 |
|---|---|
| run 단계 수락 기준 | 15/15 PASS (`acceptance.md` §D, `progress.md` §E.2·§E.3) |
| 뮤턴트 프로브 | 27/27 적색. M18은 첫 측정에서 생존했고, AC-013 S2·S3 개정(`21216ae6f`)과 테스트(`7791c0e7d`) 후 적색 |
| sync 단계 | CHANGELOG `[Unreleased]` 1건, `progress.md` §E.4, `spec.md` `in-progress → completed` (`9af4acb92`), `sync_commit_sha` 백필 (`89f3d5eb7`) |
| sync-audit (`--security`) | **PASS-WITH-DEBT 85.6** (조화평균). 차단 0건, 선택 9건 (Low 7, Info 2). Functionality 90 / Security 80 / Craft 85 / Consistency 88 |

### 커밋 목록 (카드 run·sync 구간)

| SHA | 내용 | 작성 |
|---|---|---|
| `1a3b12ce5` | plan-audit 부채 상환 (NEW-1), M1 이전 | manager-spec |
| `f611a7060` | M1 RED 테스트 | manager-develop |
| `2232e1a5f` | M1 증거 | manager-develop |
| `6f098a45e` | M2 훅 결과 기록·보조 상태 | manager-develop |
| `edecbff54` | M2 증거 | manager-develop |
| `c0e56ab09` | M3 문서·훅 주석 정정 | manager-develop |
| `989ef144b` | M3 증거 | manager-develop |
| `21216ae6f` | AC-013 S2·S3 개정 (리드 A안 승인, 재감사 없음) | manager-spec |
| `7791c0e7d` | AC-013 S2·S3 테스트 | manager-develop |
| `dc9feafb2` | moai catalog 해시 재생성 (M3 문서 수정 누락분) | manager-develop |
| `a73b3e663` | M4 뮤턴트·AC 매트릭스 증거 | manager-develop |
| `e5ff08f5d` | run 증거 SHA 백필 | manager-develop |
| `9af4acb92` | sync 단계 3-phase close | manager-docs |
| `e52f1d122` | sync-audit 보고서·재측정 출력 | sync-auditor |
| `89f3d5eb7` | `sync_commit_sha` 백필 | manager-docs |

## 2. 증거 (Evidence)

모든 원문은 `.moai/reports/t624/` 에 커밋돼 있다.

- 뮤턴트: `m4-mutant-M01.txt` … `m4-mutant-M27.txt`, `m4-mutant-M18-rerun.txt`
  - M18 재실행 적색 원문: `AC-013 S2 [regression-guard]: 2 of 2 stub invocation(s) during call 2 saw the payload file present; want every observation absent`, `AC-013 S3 [regression-guard]: call 3 stub count unchanged (delta 0); want increased — the checks must re-run`
- AC-009 패키지 실행: `m4-ac009-packages.txt` (HEAD `dc9feafb2`, `internal/hook` ok 159.262s, `internal/template` ok 51.493s)
  - 재생성 전 실패 원문: `m4-ac009-packages-before-regen.txt` (`CATALOG_HASH_UNSTABLE: moai stored hash=1d23838d…, computed hash=773956f6…`)
  - 경합 아래 타이밍 테스트 2건 실패 기록: `m4-ac009-packages-after-regen-contended.txt`
- catalog 재생성: `m4-catalog-regen.txt`, `m4-catalog-tests.txt`
- sync-audit: `sync-audit.md`, `sync-audit-selector.txt`, `sync-audit-template.txt`, `sync-audit-vet.txt`, `sync-audit-lint.txt`
- 레인 오케스트레이터 독립 재측정:
  - HEAD `7791c0e7d`: `go test ./internal/template/ -run 'TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles' -count=1 -v` → 두 테스트 FAIL (catalog 누락 재현)
  - HEAD `a73b3e663`: `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai --dry-run` → `773956f641a95d67bb5d49ffb5f77c4736425861b3ce16bfa5e9badf3daf92b7`; `go test ./internal/template/ -count=1` → `ok … 28.645s`
  - `git show 7791c0e7d:internal/template/catalog.yaml | grep -c '#'` → `0` (`--all` 재생성으로 잃은 주석 없음)
  - HEAD `9af4acb92`: `git merge-tree --write-tree --name-only develop HEAD` (로컬 develop `81c1d58f9`) → 충돌 파일 `internal/template/catalog.yaml` 1개

## 3. sync-audit 선택 발견 (부채 후보)

| ID | 심각도 | 위치 | 내용 |
|---|---|---|---|
| F1 | Low | `sync-phase-quality-gate.sh:347` | payload 는 헤더만 검증하고 본문을 그대로 재방출한다. 헤더를 맞춘 조작 JSON 이 exit 0 으로 출력됨(감사자 실측) |
| F2 | Low | `:214-215` | 상태 파일 자리에 디렉터리나 디렉터리 링크가 있으면 `mv -f` 가 그 안으로 옮긴다. 매 턴 재게이트되며 조용한 통과는 아니다 |
| F3 | Low | `:214` | 임시 파일 이름 `.<name>.tmp.$$` 가 예측 가능해, 미리 심은 링크를 `cat >` 가 따라가 덮어쓴다 |
| F4 | Low | `:253-261`, `:355` | 기록 파일 mtime 이 미래이면 나이가 음수라 60초 창을 통과한다. 매 턴 notice 만 나오고 검사는 돌지 않는다 |
| F5 | Low (카드 이전 결함) | `:574-576` | `.moai/logs` 가 디렉터리가 아니면 block JSON 출력 뒤 `set -e` 로 exit 1. 그 턴의 차단이 버려진다. base `3f5dc3f8c` 에도 `set -e`(44행)와 같은 `mkdir -p`(391행)가 있음을 확인했다. spec §4 "always exit 0" 과 새 211행 주석이 이 경로에선 거짓이다 |
| F6 | Info | `:225` | 로그 쓰기 실패 메시지 stderr 누수 |
| F7 | Low | `CHANGELOG.md` 해당 항목 | 두 거짓 문구를 훅 주석에도 귀속했다. 실제 훅 주석에는 `dependency manifest audit` 만 있었다 |
| F8 | Info | `:246` | 중첩 객체 안의 `stop_hook_active` 도 매치한다 (§D.17 선언 갭) |
| F9 | Info (범위 밖) | `agent-common-protocol.md` § Hook Invocation Surface | "advisory unless `MOAI_SYNC_GATE_BLOCKING=1`" 서술이 낡았다(훅 기본값은 차단). t644 합류 권고 |

## 4. 범위 밖 거짓 서술 — t644 확장 목록 (리드 결정 그대로)

이 카드에서 고치지 않는다. 리드는 큐 저장소 이전 뒤 t644 범위를 이 목록으로 넓히는 개정으로 처리한다.

"이 훅이 lint + test + coverage delta 를 돌린다"는 서술 (훅 자체 주석 19·272행 "tests/coverage are NOT run", 409-410행 c1 vet/lint + c2 build):

- `.claude/skills/moai/workflows/sync.md:43` (L) / `internal/template/templates/.claude/skills/moai/workflows/sync.md:43` (T)
- `.claude/rules/moai/core/agent-common-protocol-reference.md:222` (L) / 템플릿 사본 (T)
- `.claude/rules/moai/workflow/archived-agent-rejection.md:81·88·99` (L) / 템플릿 사본 (T)
- `docs-site/content/en/advanced/hooks-reference.md:171`, `docs-site/content/ko/advanced/hooks-reference.md:216` — ja·zh 에는 해당 행 자체가 없음
- 감사 권고 추가분: `agent-common-protocol.md` § Hook Invocation Surface 의 "advisory unless `MOAI_SYNC_GATE_BLOCKING=1`" (F9)

위 행 번호는 HEAD `9af4acb92` 기준이다. develop 흡수 뒤에는 달라질 수 있다(develop 이 `sync.md` 를 2줄 바꿨다).

## 5. 기준 귀속 (Baseline-attribution)

- run 단계 측정: 각 증거 파일에 적힌 HEAD (`989ef144b`, `7791c0e7d`, `dc9feafb2`)
- sync-audit 측정: HEAD `9af4acb92`, clean
- 이후 커밋 `e52f1d122`·`89f3d5eb7` 은 `.moai/reports/t624/` 보고서와 `progress.md` 한 줄만 바꿨다. 코드·훅·테스트·catalog 는 `9af4acb92` 이후 변경 없음

## 6. 미검증 (Gaps)

- develop 흡수 후 트리는 아직 측정하지 않았다. 병합 전 검증을 병합 후 근거로 재사용하지 않는다 — 창에서 재측정한다.
- Windows 실제 실행 없음. S3 는 Windows 에서 skip 하므로 그 플랫폼에서 M18 은 S2 하나만 막는다.
- `shellcheck` 미설치로 셸 정적 분석 없음.
- 뮤턴트 증거 27건 중 감사자가 다시 읽은 것은 M06·M18·M20·M24 뿐이다.
- 경합 아래 실패한 SessionStart 타이밍 테스트 2건의 원인은 부하 경합으로 추정할 뿐 증명하지 않았다.
- 디스크 가득 참·SIGKILL 중간 종료는 코드 판독만 했다.

## 7. 잔여 위험 (Residual-risk)

- F1: 로컬 쓰기 권한이 있는 프로세스가 모델에 넘어가는 `reason` 텍스트를 바꿀 수 있다. 새 권한을 주지는 않지만 재방출 설계를 유지하는 한 남는다.
- F5: 첫 실패 턴의 차단이 한 번 사라질 수 있다.
- 같은 루트의 `.moai/state` 를 두 세션이 공유하면 기록과 payload 가 서로 다른 실행에서 올 수 있다 (spec §5 가 잠금을 범위 밖으로 둠).
- S3 는 훅이 `mv` 를 이름으로 호출한다는 사실에 기댄다. 절대경로로 바뀌면 준비 단계 단언이 실패로 드러난다.

## 8. 병합 창 계획

1. `moai integration acquire --name lane-10`
2. 카드 워크트리에서 그 시점 로컬 develop 최신을 흡수 (`git merge develop`)
3. `internal/template/catalog.yaml` 충돌은 한쪽을 고르지 않고 `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai` 로 재생성 (`--all` 금지). develop 쪽 `moai-workflow-worktree`·`manager-develop`·`plan-auditor` 해시는 자동 병합분을 유지
4. 병합 트리에서 재측정: catalog 해시 테스트(4항목), AC-009 두 패키지(`internal/hook`, `internal/template`), `^TestSyncGateFailState` 셀렉터(스윕 수 확인)
5. `EnterWorktree(.claude/worktrees/develop)` → `git merge --no-ff WT-sync-gate-failstate` → 병합 커밋 트리와 재측정 트리의 `^{tree}` 동일성 확인
6. `moai integration release` → `ExitWorktree` → 리드에게 병합 SHA 보고 (push 는 리드 일괄)
