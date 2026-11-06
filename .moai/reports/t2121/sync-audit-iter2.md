auditor-model: claude-opus-5-5

# t2121 sync 단계 재감사 판정 (iter2, 수리 델타)

verdict: FAIL
audited_sha: 0a1192eeacf2a412f371a62f7625361a37c7e762

- 카드: t2121 (P0, Class B, SPEC 없음), 브랜치 `WT-complete-merge-readiness`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2121`
- 판정: **FAIL**, 종합 점수 **0.77**
- 감사 범위: 수리 델타 `git diff 7723481e7eef685cf26d3a431fa8c0cbb56d0a85..HEAD -- internal/ .moai/reports/t2121/progress.md` (커밋 fa64aa0e35 RED, 58dd8cb496 GREEN, 0a1192eeac 기록) + 지시에 따른 흡수 허용 경로 전체에 대한 새 공격 프로브. 감사 시작과 판정 작성 직전 모두 `git rev-parse HEAD` = `0a1192eeacf2a412f371a62f7625361a37c7e762`, `git status --short` = 빈 출력.
- 한 줄 근거: 지난 감사의 F1·F2는 닫혔다(재현 프로브 전부 거부, Codex 델타 감사 pass). 하지만 새 프로브가 **충돌을 카드 쪽으로 해결한 흡수 merge(-X ours)와 main 새 파일을 지운 흡수 merge가 머지 준비 판정 전체(all_passed=true)를 통과**하는 경로를 재현했다. 카드가 약속한 "충돌 해결·evil 편집이 낀 흡수 merge는 거부"를 정면으로 깨고, 착지하면 main의 변경이 조용히 되돌려진다. 수리 유래는 아니며(main 기준 코드에서도 재현) 고치는 코드는 5줄 남짓이고 감사관이 overlay로 검증했다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 70/100 | FAIL | F1·F2 닫힘: `RESULT P1_binary_swap passed=false`, `RESULT P7_minus_diff_text_edit passed=false`, `RESULT P2b_moved_hunk passed=false`. 그러나 G1: `RESULT N8_X_ours_absorb_drops_main_change passed=true want=false UNEXPECTED` + `TRIPLE all_passed=true failed=""` |
| Security (25%) | 72/100 | FAIL | 머지 게이트 무결성: 감사 뒤 main 변경을 지운 merge가 PASS로 통과(G1). 수리 자체의 우회 시도 14건 중 G1 두 건과 G2(태그 그림자) 외에는 모두 거부 |
| Craft (20%) | 80/100 | PASS | `go vet` exit 0, `gofmt -l` 빈 출력, 레인 테스트 통과(`ok ... factorylane 73.447s`). 변이 7개 중 5개 사망, M1(`--binary` 제거)·M5(`--no-textconv` 제거) 생존(G3) |
| Consistency (15%) | 90/100 | PASS | 기존 `refuse`·`GitExitError`·실패 시 닫힘 관례 유지, RED→GREEN 커밋 분리(fa64aa0e35→58dd8cb496), 진행 기록에 수리 라운드 기록 |

must-pass 방화벽: Functionality·Security 기준 미달 → 종합 FAIL. 가중 평균 75.5, 차원 조화평균 0.772 → 보고 점수 0.77.

## 지난 결함 처리 상태

| 결함 | 상태 | 근거 |
|---|---|---|
| F1 바이너리/`-diff` 내용 교체 통과 | **닫힘** | HEAD에서 `RESULT P1_binary_swap passed=false want=false OK`, `RESULT P7_minus_diff_text_edit passed=false want=false OK`, 추가로 textconv 드라이버로 내용을 가린 경우 `RESULT P1t_textconv_hides_edit passed=false want=false OK`. 커밋된 테스트 `TestT2121AbsorbRepair/P1…`, `/P7…`도 존재 |
| F2 이동된 hunk 통과 | **닫힘** | HEAD에서 `RESULT P2b_moved_hunk passed=false want=false OK`; 재merge 검사를 끄는 변이 M6이 `P2b_card_hunk_moved_in_merge_refuses`에서 사망 |
| F3 위치 지우기 근거 테스트 | 닫힘(선택 항목) | 커밋된 `P10_clean_absorb_with_shifted_line_numbers_passes` |
| F5 `sync-audit-verdict*.md` 이름 | 미처리(선택, 진행 기록에 사유) | 거짓 거부 방향 — 판정 영향 없음 |

## 결함 목록

### G1 [High] [blocking] [touched-file, 재현 증거 있음] `internal/factorylane/merge_specless.go` `checkSyncAuditVerdictFile`의 경로 비교 루프 — 트리가 감사 시점과 같아지는 흡수 merge는 흡수 검사를 아예 거치지 않음

- 신뢰도: 높음(재현함, main 기준 코드에서도 재현).
- 흡수 검사(`absorbOnlyChange`)는 `git diff --name-only audited HEAD`에 증거 디렉터리 밖 경로가 있을 때만 호출된다. 흡수 merge가 감사 뒤 main에 들어온 변경을 **전부 버리면** HEAD 트리가 감사 커밋 트리와 같아져 경로 목록이 비고, merge 재도출·카드 diff 비교가 한 번도 돌지 않은 채 PASS가 나온다.
- 재현(HEAD, 머지 준비 판정 전체):
  ```
  TRIPLE all_passed=true failed=""
  RESULT N4c_delete_main_file_in_absorb passed=true want=false UNEXPECTED | .moai/reports/t9/verdict.md: verdict: PASS, audited_sha 3574667e… binds HEAD 91532a86… — the SPEC-less card's audit record reads closed
  TRIPLE all_passed=true failed=""
  RESULT N8_X_ours_absorb_drops_main_change passed=true want=false UNEXPECTED | .moai/reports/t9/verdict.md: verdict: PASS …
  ```
  N8: 감사 뒤 main이 카드와 같은 줄을 바꿈 → 레인이 `git merge -X ours main`으로 카드 쪽을 택함. N4c: 흡수 merge 안에서 main이 새로 추가한 파일을 삭제.
- 결과: 카드 브랜치를 main에 착지시키면 main의 해당 커밋이 조용히 되돌려진다. 감사는 이 되돌림을 본 적이 없다. 카드의 커밋된 테스트 이름 `conflict_resolved_merge_refuses`와 코드 주석 "(b) is the load-bearing proof: an evil merge … cannot pass"가 이 경우 거짓이 된다.
- 귀속: main 기준 코드(`git show main:internal/factorylane/merge_specless.go`, 시그니처만 맞춤)에 같은 프로브를 돌려도 `N4c … passed=true`, `N8 … passed=true` — **기존 결함이며 이번 수리가 만든 것이 아니다**(수리 유래 아님, 반복 정지 규칙 해당 없음). 그래도 카드가 흡수를 정식 흐름으로 만들고 "충돌 해결 흡수는 거부"를 약속했으므로 이 카드의 요구 위반으로 본다.
- Required fix: `audited != HEAD`일 때 `git rev-list --min-parents=2 audited..HEAD`가 비어 있지 않으면(또는 실패하면) 경로 목록과 무관하게 `absorbOnlyChange`를 호출한다. 감사관 overlay 검증(파일 무변경):
  ```go
  if merges, mErr := git.Git("rev-list", "--min-parents=2", auditedFull+".."+head); mErr != nil || strings.TrimSpace(merges) != "" {
      if absorbed, why := absorbOnlyChange(git, develop, auditedFull, head, card); !absorbed {
          return refuse("audited_sha %s is stale: a merge after it is not a plain absorb of %s: %s", auditedFull, develop, why)
      }
  }
  ```
  결과: `RESULT N4c_delete_main_file_in_absorb passed=false want=false OK`, `RESULT N8_X_ours_absorb_drops_main_change passed=false want=false OK`, 정상 흡수 `N2a`·`N3a`는 `passed=true` 유지, 패키지 전체 `ok  github.com/modu-ai/moai-adk/internal/factorylane	125.578s`. N4c·N8 형태의 재현 테스트를 커밋에 추가할 것.

### G2 [Medium] [optional] [in-diff] `absorbOnlyChange`의 `develop+"^{commit}"` — 같은 이름의 태그가 브랜치를 가림

- 신뢰도: 높음(재현함).
- `refs/tags/main`을 main이 아닌 커밋(미감사 파일 포함)에 만들고 그 커밋을 흡수하면 통과: `RESULT N5_tag_main_shadows_branch passed=true want=false UNEXPECTED`, `TRIPLE all_passed=true`. main 기준 코드에서는 `N5 … passed=false`였으므로 이 카드가 연 경로다.
- 선택 처리 이유: 레인이 의도적으로 `main` 태그를 만들어야 하며, 같은 ref 저장소에 쓸 수 있는 레인은 `refs/heads/main`을 직접 옮길 수도 있다. 머지 준비 판정의 다른 조건(`checkConflictFree`, `checkTreeIdentityPreMerge`)도 같은 맨 이름을 쓴다. 지난 감사 F7과 같은 뿌리.
- 권장 수정: 통합 ref를 `refs/heads/<develop>`(또는 `--symbolic-full-name`으로 모호성 거부)로 해석. 후속 카드 후보.

### G3 [Low] [optional] [in-diff] 테스트 공백 — `--binary`·`--no-textconv`를 지켜 주는 테스트가 없음

- 변이 M1(`--binary` 제거)·M5(`--no-textconv` 제거)가 커밋된 테스트에서 살아남았다. P1·P7은 이제 재merge 검사가 먼저 잡으므로 `--binary`가 유일한 방어인 경우를 테스트가 덮지 않는다.
- 감사관 프로브 N7b(바이너리 파일 + 저장소 설정의 사용자 merge 드라이버가 내용을 주입, 재merge도 같은 내용을 재현)가 그 경우다: HEAD에서 `RESULT N7b_binary_merge_driver_injects_content passed=false want=false OK`, M1에서 `RESULT N7b_binary_merge_driver_injects_content passed=true want=false UNEXPECTED`. 즉 `--binary`는 쓸모가 있지만 잠겨 있지 않다.
- 권장: N7b 형태와 P1t(textconv) 형태 테스트 추가.

### G4 [Info] [optional] [in-diff] 옥토퍼스 거부는 메시지로만 고정됨

- 변이 M3(옥토퍼스 허용)은 `detail lacks "octopus"`로 죽었고, 동작은 재merge 트리 불일치로 여전히 거부된다(이중 방어). 기록만.

## 새 공격 프로브 결과 (HEAD 0a1192eeac, overlay, 트리 무변경)

| 프로브 | 내용 | 결과 | 기대 | 맞음? |
|---|---|---|---|---|
| N1 | 두 번째 부모가 main 자손 + main에 없는 커밋 | 거부 | 거부 | 예 |
| N2a | 오래된 main 커밋(분리 HEAD)을 깨끗이 흡수 | 통과 | 통과 | 예 |
| N2b | 오래된 main 커밋 태그를 흡수하며 비카드 경로 편집 | 거부 | 거부 | 예 |
| N3a | 흡수 두 번, 둘 다 깨끗 | 통과 (`TRIPLE all_passed=true`) | 통과 | 예 |
| N3b | 두 번째 흡수에서 main 파일 편집 | 거부 | 거부 | 예 |
| N4a | 흡수 안에서 카드 파일 이름 변경 | 거부 | 거부 | 예 |
| N4b | 흡수 안에서 실행 권한 변경 | 거부 | 거부 | 예 |
| N4c | 흡수 안에서 main 새 파일 삭제 | **통과** | 거부 | **아니오 (G1)** |
| N5 | `main` 태그가 브랜치를 가림 | **통과** | 거부 | **아니오 (G2)** |
| N6 | `commit-tree`로 만든 무변경 merge에 편집 실음 | 거부 | 거부 | 예 |
| N7 | 사용자 merge 드라이버가 텍스트 내용 주입 | 거부 | 거부 | 예 |
| N7b | 사용자 merge 드라이버가 바이너리 내용 주입 | 거부 | 거부 | 예 (`--binary`가 잡음, G3) |
| N8 | `-X ours`로 충돌 해결, main 변경 버림 | **통과** | 거부 | **아니오 (G1)** |
| N9 | main이 카드 파일 이름 변경 후 깨끗한 흡수 | 거부 | 거부(거짓 거부 허용) | 예 (실패 시 닫힘 방향) |

## 실행한 명령과 관측 결과

### Claim
수리는 F1·F2를 닫았고 정적 검사·지정 테스트는 모두 통과한다. 그러나 흡수 허용 경로에 트리 동일 우회(G1)가 남아 있어 카드 요구를 만족하지 못한다.

### Evidence
- `timeout 580 go test ./internal/factorylane/ -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/factorylane	73.447s`
- `timeout 580 go test ./internal/cli/ -run MergeReady -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	48.296s`
- `timeout 580 go test ./internal/cli/ -run SpecLess -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	22.082s`
- `timeout 580 go test ./internal/cli/ -run FactoryMerge -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	27.811s`
- `go vet ./internal/factorylane/` → `vet_exit=0`; `gofmt -l internal/factorylane` → 빈 출력, `gofmt_exit=0`
- 프로브: `go test -overlay <scratch>/probe-overlay.json ./internal/factorylane/ -run TestAuditProbeIter2 -count=1 -v` → exit 0, 위 표의 RESULT/TRIPLE 줄 (`ok … 55.403s`)
- 기준선 프로브(main의 `merge_specless.go`, 시그니처만 맞춤): `N4c … passed=true`, `N8 … passed=true`, `N5 … passed=false`, `N2a`·`N3a` `passed=false`(기준선은 흡수를 허용하지 않음)
- 변이(`go test -overlay <mutant>.json ./internal/factorylane/ -run T2121 -count=1`):
  ```
  M1_drop_binary              rc=0 SURVIVED  ok … 48.091s
  M2_drop_tree_equality       rc=1 KILLED    --- FAIL: TestT2121AbsorbRepair (28.09s)
  M3_allow_octopus            rc=1 KILLED    --- FAIL: TestT2121AbsorbRepair/octopus_merge_refuses (detail lacks "octopus"; 동작은 트리 불일치로 여전히 거부)
  M4_invert_conflict_exit     rc=1 KILLED    --- FAIL: TestT2121AbsorbRepair/conflict_resolved_merge_refuses
  M5_drop_no_textconv         rc=0 SURVIVED  ok … 50.828s
  M6_skip_remerge             rc=1 KILLED    --- FAIL: …/P2b_card_hunk_moved_in_merge_refuses, …/conflict_resolved_merge_refuses, TestT2121AbsorbRepairUnsupportedGitFailsClosed
  M7_oldtool_passes           rc=1 KILLED    --- FAIL: TestT2121AbsorbRepairUnsupportedGitFailsClosed (2.31s)
  ```
- G1 수정안 overlay: 프로브 `N4c`·`N8` → `passed=false … OK`, `N2a`·`N3a` → `passed=true … OK`; `go test -overlay FIX.json ./internal/factorylane/ -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/factorylane	125.578s`
- `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2121` → `config_status: ok`, `model: codex`, codex `required`, claude·glm `off`, `enforced_required: ["codex"]`, `cross_model_active: true`
- `mcp__moai__codex_audit` (native, target baseBranch, delta_base `7723481e7eef685cf26d3a431fa8c0cbb56d0a85`, project_root = 이 트리, 1회) → `verdict: pass`, findings 없음, `review_base: delta_base (merge base 7723481e7eef685cf26d3a431fa8c0cbb56d0a85)`, `resolved_model: gpt-6.1-sol`, **audit_receipt `rcpt-035ee381a55be2dfc36e2f8d`**. Codex는 정적 검토만 했다고 밝혔고, G1은 델타 밖(기존 코드)이라 그 범위에 들지 않는다.

### Baseline-attribution
모든 명령은 이 감사 실행에서 트리 HEAD `0a1192eeacf2a412f371a62f7625361a37c7e762`(작업 트리 깨끗)에 대해 실행했다. Go 테스트는 이 트리 소스로 `go test`가 빌드했다(설치된 바이너리 미사용). `moai verify audit-plan`과 MCP 서버는 설치된 빌드 `04d8c27a7e`(HEAD의 조상, Codex 결과의 `build_lag` 알림과 일치) — 이번 diff는 감사 계획 로직을 건드리지 않으므로 판정에 영향 없음. 기준선 비교는 `git show main:internal/factorylane/merge_specless.go`(local main `04d8c27a7e`)를 overlay로 대체해 측정했다.

### Gaps
- 작업 트리 격리 가드가 네 번 거부했다: (1) 셸 변수로 경로를 만든 `go test -overlay` 명령, (2) 인라인 heredoc python 변이 생성기, (3) 파일 이름에 `git`이 든 변이 overlay 실행, (4) 여러 git 명령을 묶은 확인 명령. 각각 경로를 풀어 쓴 단일 명령 / `python3 -I` 스크립트 파일 / 파일 이름 변경(`M7_oldtool_passes`) / 분리된 명령으로 같은 동작을 다시 실행했다 — 측정 대상은 바뀌지 않았다.
- cli 테스트는 지시대로 패턴별 개별 명령으로 돌렸다. `./internal/cli/` 전체·저장소 전체 테스트는 실행 안 함(CI 몫).
- CI 러너의 git 버전 미확인. 다만 `merge-tree --write-tree`는 이 카드 이전부터 `merge.go:334`·`:357`(`checkConflictFree`, `checkTreeIdentityPreMerge`)이 이미 쓰고 있어 git ≥ 2.38 요구는 새로 생긴 것이 아니다.
- N7b의 거부 사유 문자열은 260자 절단으로 직접 보지 못했다. `--binary`가 잡는다는 판단은 M1 변이에서 같은 프로브가 통과로 바뀐 차이로 추론했다.
- 커버리지 수치는 측정하지 않았다.
- `audit_multi`는 리더 지시로 부르지 않았다(필수 backend는 codex 하나, 그 감사는 수행·영수증 발급됨).

### Residual-risk
- G1 수정안을 적용해도 main 변경과 카드 변경이 같은 파일에서 충돌 없이 자동 merge되며 의미가 상호작용하는 경우는 텍스트 수준 검사로 못 잡는다(지난 감사와 같은 잔여 위험).
- G2(태그 그림자)는 ref 저장소 신뢰를 전제로 한 다른 조건들과 같은 가정에 기대며, 선택 항목으로 남는다.
- 증거 파일을 커밋이 아니라 작업 트리에서 읽는 기존 동작(지난 F6)은 그대로다.
- 프로브·변이 원본 출력은 머신 로컬 스크래치 `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/1df9d274-f8c9-4593-8790-efd5997fbc70/scratchpad/t2121-iter2/`에만 있다. 판정을 가른 줄은 위에 옮겨 적었고, 나머지는 사라질 수 있다.

## 반복 이력
- iter1: FAIL 0.70 @7723481e7e — F1(바이너리 내용), F2(hunk 이동) 차단.
- iter2(이번): FAIL 0.77 @0a1192eeac — F1·F2 닫힘, 새 차단 G1(기존 결함, 수리 유래 아님 → 반복 정지 규칙 해당 없음).
- Binding run conditions: `progress.md`에 해당 제목 없음(`grep -c "Binding run conditions"` → `0`).

## 재감사 범위
G1 수정 커밋과 그 재현 테스트(N4c·N8 형태)만 델타로 본다. G2~G4는 선택 사항이다.
