# t431 판정서 — CI 실패 군집 2건 해소 확인 (doctor 군집 + CodexInit 군집)

작성: t431 레인 (`WT-doctor-codex-check`, base `ad272be20` = `origin/develop`)
일시: 2026-09-02 (KST)

## 판정

**두 군집 모두 해소 확인. 수리 불필요 — 확인 기록만으로 종결 후보.**
두 군집은 **같은 원인 계통(클래스)의 다른 인스턴스**이며, 각각 별도 수리 커밋이 이미 착지해 있다.
카드 분할 불필요 — 확인 임무가 양쪽 모두 완료됐고 남는 수리 범위가 0이다.

## 1. 원인과 수리

### Doctor 군집 (TestRunDoctor_* / TestDoctorCmd_* 9종)

- 실패 본문 (run 33123997760, 2026-08-27 22:49 UTC):
  `✗ Error: no readable binary to judge at /home/runner/work/moai-adk/moai-adk/bin/moai (11 committed artifacts to compare)`
  → `coverage_improvement_test.go: unexpected error: doctor: 1 check(s) failed`
- 원인: `internal/cli/doctor_agentemit_embed.go` 의 agent-emit embed check 이 "판정 대상(binary) 부재"를 실패 판정으로 취급. CI 러너는 `go test`만 돌려 바이너리를 빌드하지 않으므로 러너에서는 항상 하드에러 → doctor 종료코드 1 → 실행 테스트 9종 동반 실패.
- 수리: `f6c027fa0` fix(SPEC-CI-DOCTOR-BIN-001) M1+M2 — "bin-absent embed check becomes an informational skip" (t346, 2026-08-28 12:14 UTC). bin 부재 시 `CheckFail` → `CheckOK`(정보성 skip, 메시지에 미판정 사실 기록). REQ-CDB-003(읽을 수 있는 binary가 있을 때의 fail 경로)은 유지.

### CodexInit 군집 (TestCodexInit* 3종 + 하위테스트, 전부 `/spawn=true` 셀만)

- 실패 본문 (run 33123997760 · 33171483207):
  `codex_init_test.go:471: unexpected non-exit-code error: --spawn needs the moai binary in PATH: exec: "moai": executable file not found in $PATH`
- 원인: `checkSpawnPrereqs()` (`internal/cli/spawn.go`)가 `exec.LookPath("moai")` 로 전제 검사. 러너 PATH 에 moai 가 없으므로 spawn=true 셀만 비-종료코드 에러로 실패. spawn=false 셀은 전부 통과 — 지문이 원인을 직접 가리킨다.
- 수리: `e76ce9520` fix(cli): route --spawn prereq lookups through a PATH-resolution seam (t349, 2026-08-28 12:41 UTC). `spawnLookPath` seam 변수로 `exec.LookPath` 를 감싸고 테스트 하네스가 스텁(`/stub/bin/<file>` 반환) — 셀이 호스트가 설치한 바이너리가 아니라 spawn seam 에 의해 한정됨.

## 2. 판별 실험 — 어떤 수리가 어떤 군집을 끝냈는가

두 실패 run 의 헤드 SHA 에 대한 조상 관계 측정 (`git merge-base --is-ancestor`):

| run (UTC) | 헤드 | t346 (doctor 수리) | t349 (codex 수리) | doctor 실패 | CodexInit 실패 |
|---|---|---|---|---|---|
| 33123997760 (08-27 22:49) | `44095ddc` | 없음 | 없음 | 9종 | 3종 |
| 33171483207 (08-28 12:32) | `e08d5e55` | **있음** | 없음 | **0건** | 3종 (여전) |

t346 만 있는 트리에서 doctor 군집은 소멸하고 CodexInit 은 남았다 — 두 수리의 효과가 자연 실험으로 분리된다.
`ad272be20`(develop 팁)은 t346·t349 양쪽을 모두 조상으로 포함한다(측정 완료).

## 3. 긍정 증거 — baseline 러너 run 에서 실제 실행·통과 직독

`gh run list` 부재는 증거가 아니므로(리드 [HARD] 규율), baseline run **33568757908** (`ad272be20`, 2026-09-01 22:57 UTC)의
`test-stream.json.gz` artifact (`go test -json` 이벤트 스트림)를 내려받아 개별 pass 이벤트를 직독했다:

- 군집 부모 테스트 **39종 전부 `Action=pass`** (하위테스트 포함 pass 이벤트 186건):
  - `TestCodexInit*` 6종 — AcceptDelegation / Decline / FailurePaths / GateInjectedState / GateStateMatrix / PromptIssuance — **`/spawn=true` 하위테스트 전부 포함**
  - `TestRunDoctor_*` 15종 + `TestDoctorCmd_*` 13종 (+ `internal/cli/harness` 2종)
  - `TestBinaryLag_*` 3종 — `OneSeamServesBothSurfaces` 포함
- 군집 테스트의 `Action=skip` 이벤트 **0건** (skip 이 아니라 실제 실행 후 통과)
- census totals: `packages=137 passed=20129 skipped=104 failed=2` — 실패 2건은 `internal/web` 의 `TestDataI18nKeysSubsetOfDictionary` / `TestI18nKeySetParity` (군집과 무관 — §6 부수 관측)

## 4. 동일 계통 판정 (리드 질의 3)

**같은 클래스, 다른 인스턴스, 다른 수리.** 공통 형태: "러너 환경에서 moai 바이너리를 해석할 수 없다는 사실을 전제로 삼은 검사(doctor 판정 check, spawn 전제검사)가 CI 에서 하드페일을 낸다". 서브시스템과 수리는 별개다 (t346 = doctor check 의 bin-absent 분기 변경 / t349 = spawn 전제검사의 seam 화). 분할 제안은 하지 않는다 — 둘 다 이미 수리돼 남는 범위가 없어 분할할 실체가 없다.

## 5. 군집 실패 관측 전수 (2026-08-26T00Z ~ 09-02, ci.yml)

- failure run 13건 (develop, 08-28 13:06 ~ 08-30 15:42): 군집 0히트
- failure run 7건 (08-31, develop 6 + release/v3.1.4 PR 1): 군집 0히트
- failure run 3건 (PR 이벤트: fix/heavy-gate-nested-toolchain, WT-lead-debottleneck, release/v3.1.4): 군집 0히트
- cancel run 63건 (08-26 ~ 09-02): **군집 실패 2건** — §1·§2 의 33123997760(양쪽), 33171483207(CodexInit만)
- 재시도(run_attempt>1) run: 창 내 1건 (32993568893, WT-t250-followup) — attempt-1 군집 0히트

## 6. 부수 관측 (카드 밖 — 리드 전달용)

1. **baseline `ad272be20` 가 red** — run 33568757908 (09-01 22:57 UTC) 실패 원인은 `internal/web` i18n 2건 (`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`, Test job + Race job 양쪽). 리드 관측("최신 실패는 전부 TestConcurrencyStress")과 다른 신규 군집. t413 병합 push 직후의 run 이다.
2. `TestBinaryLag_OneSeamServesBothSurfaces` 의 08-28 23:17 UTC 단발 실패 (run 33219939019)는 binary-lag 로직 결함이 아니라 **TempDir RemoveAll cleanup race** (`unlinkat ... directory not empty`) — SPEC-TEMPDIR-CLEANUP-RACE-001 (t352, `410f6241d`) 이 수리한 계열. 91ec14d 에는 t352 가 없었고(조상 측정), 08-29 이후 0건, baseline 에서 pass.

## 7. 카운트 정산 (lane-1 보고 8 run / 5 run 대비)

레인의 전수 방법(§5)으로 재현된 군집 실패는 **2 run** 이다 (doctor: 1 run, CodexInit: 2 run — 33123997760 에 양쪽 동반).
lane-1 의 8/5 는 이 방법으로는 재현되지 않았다 — 스윕 창·grep 면이 다를 가능성. 판정(해소)은 카운트와 무관하게 성립한다: 원인·수리·판별·긍정 증거가 모두 독립적으로 성립하기 때문.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: doctor 군집·CodexInit 군집 모두 baseline `ad272be20` 에서 해소. 각각 t346 / t349 가 수리. 두 군집은 같은 클래스의 다른 인스턴스.
- **Evidence**: §1 실패 본문(run 로그 인용), §2 조상 관계 측정 6건, §3 artifact pass 이벤트 직독(39종·186건·skip 0), §5 전수 목록. run 로그·artifact 는 run ID 로 재인출 가능 (`gh run view <id> --log-failed`, `gh run download <id> -n test-stream-ci-test-ubuntu-latest`).
- **Baseline-attribution**: 모든 통과 측정은 run 33568757908 (헤드 `ad272be20` = `origin/develop` = 레인 워크트리 base, 커밋 `git rev-parse HEAD` = `ad272be20`)의 러너 스트림. 로컬 재현 아님.
- **Gaps**: (1) lane-1 의 8/5 카운트 미재현(§7). (2) 08-25 이전 창은 미조사 — 다만 CodexInit 테스트는 `576dbad05`(08-27 19:09 UTC) 에 탄생해 이전 실패가 구조적으로 불가능하고, doctor 군집의 하드페일 분기도 `f3e5006ce`(08-27 14:27 UTC) 이후의 형태다. (3) baseline run 이후 develop 에 새 push 가 생기면 이 판정서의 baseline 은 과거가 된다(현재 관측 시점엔 33568757908 이 최신 develop run).
- **Residual-risk**: (1) 두 수리는 CI 환경 결함을 정보성 skip / seam 스텁으로 처리한다 — bin 부재 상태에서의 embed 판정은 CI 에서 계속 불가(의도된 절충, SPEC-CI-DOCTOR-BIN-001 소관)이고, spawn=true 의 실바이너리 경로는 CI 가 아니라 로컬·spawn 환경의 몫이다. (2) §6 의 i18n 군집은 이 카드에서 판정하지 않았다.
