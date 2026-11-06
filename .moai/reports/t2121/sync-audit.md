auditor-model: claude-opus-5-5

# t2121 sync 단계 감사 판정

verdict: FAIL
audited_sha: 7723481e7eef685cf26d3a431fa8c0cbb56d0a85

- 카드: t2121 (P0, Class B, SPEC 없음), 브랜치 `WT-complete-merge-readiness`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2121`
- 판정: **FAIL**, 종합 점수 **0.70**
- 감사 범위: `git diff main..HEAD` (local main `04d8c27a7e`가 HEAD의 조상임을 `git merge-base --is-ancestor main HEAD`로 확인). 감사 시작과 판정 작성 직전에 `git rev-parse HEAD` = `7723481e7eef685cf26d3a431fa8c0cbb56d0a85`, `git status --short` = 빈 출력.
- 판정 근거 한 줄: 카드 목표가 명시한 "흡수 merge에 충돌 해결·evil 편집이 끼면 거부"가 두 경로에서 뚫린다 — (1) 바이너리(또는 `-diff` 속성) 파일의 내용 교체, (2) 카드가 바꾼 줄의 위치 이동. 두 경로 모두 재현 테스트로 확인했고, Codex 독립 감사도 같은 두 건을 P1로 짚었다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 55/100 | FAIL | 요구 (2) 증거 파일 이름 선택은 충족. 요구 (1)의 거부 조건 중 "evil 편집 거부"가 F1·F2로 깨짐. 레인 테스트 자체는 통과: `ok  github.com/modu-ai/moai-adk/internal/factorylane	77.072s` |
| Security (25%) | 60/100 | FAIL | 머지 게이트의 무결성 우회: 감사되지 않은 내용이 PASS 판정과 함께 통과. 재현: `RESULT P1_binary_card_file_content_swapped_in_merge passed=true`, `RESULT P2b passed=true`, `RESULT P7_gitattributes_marks_card_text_binary_then_edit passed=true` |
| Craft (20%) | 78/100 | PASS | `go vet` exit 0, `gofmt -l` 빈 출력, `golangci-lint run ./internal/factorylane/` → `0 issues.`; 실제 git 저장소 픽스처 테스트 15개; 변이 10개 중 6개 사망(F4) |
| Consistency (15%) | 88/100 | PASS | 기존 `refuse` 형식·`GitRunner`·실패 시 닫힘 관례를 그대로 따름. 레인 진행 기록에 RED→GREEN 커밋 분리(`e6f5ad9c51`→`be6e18347b`) |

must-pass 방화벽: Functionality·Security 둘 다 기준 미달 → 종합 FAIL. 점수는 가중 평균 64.3, 차원 조화평균 약 0.68 — 보고 점수 0.70.

## 결함 목록

### F1 [High] [blocking] [in-diff] `internal/factorylane/merge_specless.go` `cardPatch` (diff 인자·`index` 줄 제거) — 바이너리 내용 교체가 비교에 안 잡힘

- 신뢰도: 높음(재현함, Codex도 독립 지적).
- 기본 `git diff`는 바이너리 파일에 대해 `Binary files a/x and b/x differ`만 찍는다. 내용이 다르다는 유일한 표지는 `index <old>..<new>` 줄인데, `cardPatch`가 바로 그 줄을 지운다. 그래서 카드가 추가·수정한 바이너리 파일을 흡수 merge 안에서 다른 내용으로 바꿔도 감사 시점과 HEAD의 정규화 패치가 같아진다.
- 텍스트 파일도 같은 구멍에 빠진다: 작업 트리 `.gitattributes`의 `-diff`(또는 `binary`)가 걸린 경로는 바이너리로 취급된다(P7). textconv 드라이버가 설정된 경로도 마찬가지로 내용이 가려질 수 있다(재현은 안 함).
- 코드 주석 "(b) is the load-bearing proof: an evil merge … changes the card diff … so it cannot pass"는 이 경우 거짓이다.
- Required fix: `cardPatch`의 diff에 `--binary --no-textconv`를 추가해 바이너리 내용을 비교에 넣는다(`--binary`는 blob 내용을 패치로 내보내므로 `index` 줄을 지워도 내용 차이가 남는다). P1·P7을 재현 테스트로 커밋에 추가한다.

### F2 [High] [blocking] [in-diff] `internal/factorylane/merge_specless.go` `cardPatch` (`@@` 위치 지우기) — 카드가 바꾼 줄을 다른 자리로 옮겨도 통과

- 신뢰도: 높음(재현함, Codex도 독립 지적).
- 재현 P2b: main에 있던 `code.txt`(a..e)에 카드가 `guard` 한 줄을 a 다음에 넣고 감사받음 → 흡수 merge에서 `guard`를 d 다음으로 옮김 → `passed=true`. Codex 예시: 감사받은 함수의 `return false`→`return true` 교체를 되돌리고 같은 교체를 다른 함수에 적용해도 통과.
- 진행 기록의 Residual-risk에 "순서 유지 재배치는 못 잡는다"고 적혀 있지만, 카드 목표의 "evil 편집 거부"를 정면으로 깨는 경로라 잔여 위험으로 넘길 수 없다. 위치 이동은 의미가 크게 바뀌는 편집이다(가드 줄을 위험 동작 뒤로 옮기는 식).
- Required fix: 위치를 지운 패치 비교만으로 merge를 믿지 말고, first-parent 경로의 각 흡수 merge `M`(부모 `p1 p2`)을 `git merge-tree --write-tree p1 p2`로 다시 merge해 결과 트리가 `M^{tree}`와 같은지 확인한다. 충돌(종료 코드 1)이나 트리 불일치는 거부, octopus merge도 거부. 기존 (b) 비교는 그대로 둔다. 감사관이 overlay로 이 수정안을 시험했다: 프로브 11개 전부 기대대로(P10 정상 흡수는 통과 유지), 레인 테스트도 통과 — `ok  github.com/modu-ai/moai-adk/internal/factorylane	78.321s` (`-run 'TestAuditProbe|T2121|SpecLess'`). `merge-tree --write-tree`는 git 2.38 이상 필요(로컬 `git version 2.54.0`), CI 러너 버전은 확인 안 함.

### F3 [Low] [optional] [in-diff] 테스트 공백 — 정규화가 존재하는 이유를 고정하는 테스트가 없음

- 변이 M4(`@@` 위치 보존)·M5(`index` 줄 보존)가 레인 테스트에서 살아남았다. 레인 픽스처에서는 main이 카드 파일을 건드리지 않아 위치·blob id가 안 바뀌기 때문이다. 즉 "main이 같은 파일 앞쪽에 줄을 넣어도 판정 유지"라는 설계 이유가 테스트로 잠겨 있지 않다. 감사관 프로브 P10(main이 카드 파일 맨 앞에 줄 추가 후 흡수 → 통과)은 현 구현에서 통과한다.
- Required fix(선택): P10 형태의 테스트를 커밋에 추가.

### F4 [Info] [optional] [in-diff] (a) 단계의 merge 부모 검사·비-merge 경로 검사는 (b)와 중복

- 변이 M1(부모 조상 검사 끔)·M2(비-merge 커밋 경로 검사 끔)가 살아남았다. 다른 브랜치 merge나 늦은 코드 커밋은 어차피 (b) 패치 비교가 잡기 때문이다. 이중 방어로서 해롭지 않으니 유지해도 된다. 기록만 한다.

### F5 [Low] [optional] [in-diff] 증거 파일 이름 패턴이 관례 이름 하나를 놓침

- `specLessSyncAuditPattern = ^sync-audit(?:-iter(\d+))?\.md$`는 `sync-audit-verdict*.md`(sync-auditor 지침이 허용하는 이름)를 받지 않는다. 그런 이름만 있으면 실패 시 닫힘으로 거부되므로 위험은 거짓 거부 쪽이다. `strconv.Atoi` 넘침은 n=0으로 떨어져 iter 없는 이름에 진다(무해).
- `verdict.md`가 있으면 더 새로운 `sync-audit-iterN.md`가 있어도 `verdict.md`를 읽는다 — 카드 목표 문구대로의 설계이므로 결함 아님.

### F6 [Medium] [optional] [outside] 증거 파일을 커밋이 아니라 작업 트리에서 읽음 (기존 동작)

- `checkSyncAuditVerdictFile`은 `os.ReadFile(repoDir/…)`로 읽어, 커밋되지 않은 파일도 증거로 받는다. 이번 diff는 그 표면을 `sync-audit*.md`로 넓혔을 뿐 원인은 기존 코드다. 위치가 diff 밖이므로 판정에 반영하지 않고 리더에게 보여줄 항목으로만 남긴다.

### F7 [Low] [optional] [outside] 통합 ref 이름 모호성

- `develop+"^{commit}"`은 `refs/tags/main`이 있으면 브랜치보다 태그를 먼저 고른다. 다른 merge 조건도 같은 이름 해석을 쓰는 기존 동작이다. 기록만 한다.

## 실행한 명령과 관측 결과 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
레인 테스트·정적 검사는 통과하지만, 카드가 명시한 evil-merge 거부가 두 경로(F1, F2)에서 실패한다.

### Evidence
- `timeout 580 go test ./internal/factorylane/ -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/factorylane	77.072s`
- `timeout 580 go test ./internal/cli/ -run MergeReady\|SpecLess\|Specless\|FactoryMerge -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	87.768s`
- `go vet ./internal/factorylane/` → `vet exit 0`; `gofmt -l internal/factorylane` → 빈 출력; `golangci-lint run ./internal/factorylane/` → `0 issues.`
- 감사관 프로브(overlay, 커밋 안 함, `-run TestAuditProbe -v`), 현 구현:
  ```
  RESULT P1_binary_card_file_content_swapped_in_merge passed=true      ← 기대 false (F1)
  RESULT P2_card_added_line_moved_in_merge passed=false                 (새 파일이라 전 줄이 + — 위치 무관하게 잡힘)
  RESULT P3_mode_change_on_card_file_in_merge passed=false
  RESULT P4_delete_main_file_in_merge passed=false
  RESULT P5_path_with_spaces_edit_in_merge passed=false
  RESULT P6_symlink_retarget_in_merge passed=false
  RESULT P7_gitattributes_marks_card_text_binary_then_edit passed=true  ← 기대 false (F1)
  RESULT P8_CRLF_flip_on_card_file_in_merge passed=false
  RESULT P2b passed=true                                                ← 기대 false (F2)
  RESULT P9 passed=false                                                (octopus: main + 무관 브랜치)
  RESULT P10 passed=true                                                (main이 카드 파일 앞에 줄 추가 후 정상 흡수 — 기대대로 통과)
  ```
  P1 거부 실패 상세: `.moai/reports/t9/verdict.md: verdict: PASS, audited_sha 036f698796b4… binds HEAD a80ddb8fdcbd… — the SPEC-less card's audit record reads closed`
- 변이 시험(`go test -overlay`, `-run 'T2121|SpecLess'`, 트리 무변경):
  ```
  M1_no_extra_parent_check SURVIVED
  M2_no_nonmerge_path_check SURVIVED
  M3_skip_patch_compare KILLED rc=1 --- FAIL: TestT2121AbsorbTolerance (30.36s)
  M4_keep_hunk_positions SURVIVED
  M5_keep_index_lines SURVIVED
  M6_lexical_iter_pick KILLED rc=1 --- FAIL: TestT2121EvidenceFileNames (4.12s)
  M7_sync_audit_wins_over_verdict KILLED rc=1 --- FAIL: TestT2121EvidenceFileNames (5.23s)
  M8_unresolvable_dev_accepts KILLED rc=1 --- FAIL: TestT2121AbsorbTolerance (40.48s)
  M9_absorb_always_true KILLED rc=1 --- FAIL: TestT2121AbsorbTolerance (27.08s)
  M10_no_suffix_counts_as_99 KILLED rc=1 --- FAIL: TestT2121EvidenceFileNames (4.07s)
  ```
  (M6 첫 실행은 `declared and not used: bestN` 빌드 실패라 사망으로 치지 않고, 변이를 고쳐 재실행한 결과를 적었다.)
- 수정안 overlay(F2 merge-tree 재검증 + F1 `--binary --no-textconv`) 시험: 위 프로브 전부 기대대로(P1·P7·P2b → `passed=false`, P10 → `passed=true`), `ok  	github.com/modu-ai/moai-adk/internal/factorylane	78.321s`.
- `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, codex gate `required`, claude·glm `off`, `enforced_required: ["codex"]`.
- `mcp__moai__codex_audit` (native, target baseBranch, project_root = 이 트리, 1회) → `verdict: fail`, 결과 P1 2건(바이너리 blob ID 소실, hunk 위치 소실), `review_base: main (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)`, `resolved_model: gpt-6.1-sol`, **audit_receipt `rcpt-e11ac44f0a5ac8f7d8cabdf7`**.

### Baseline-attribution
모든 명령은 이 감사 실행에서, 트리 HEAD `7723481e7eef685cf26d3a431fa8c0cbb56d0a85`(작업 트리 깨끗)에 대해 실행했다. Go 테스트는 이 트리에서 `go test`로 빌드해 돌렸다(설치된 바이너리 미사용). `moai verify audit-plan`과 MCP 서버는 설치된 빌드 `04d8c27a7e`(HEAD의 조상)이다 — 이번 diff는 감사 계획 로직을 건드리지 않으므로 판정에 영향 없음.

### Gaps
- cli 테스트 첫 실행은 작업 트리 격리 가드가 거부했다(패턴의 `|`가 든 따옴표 인자). 같은 명령을 `\|` 이스케이프로 다시 실행해 결과를 얻었다 — 대체가 아니라 같은 명령의 재실행. 변이 생성 스크립트도 같은 가드에 한 번 거부돼, 스크립트 파일로 옮겨 `python3 -I`로 실행했다.
- 감사 계획은 `cross_model_active: true`라 지침상 `audit_multi` 경로이지만, 리더 지시("codex_audit 정확히 1회")를 따랐다. 필수 backend는 codex 하나뿐이고 그 감사는 수행·영수증 발급됨. `audit_multi`는 부르지 않았다.
- `./internal/cli/` 전체, 저장소 전체 테스트는 실행 안 함(CI 몫).
- textconv 우회 경로는 추론이고 재현하지 않았다.
- CI 러너의 git 버전(`merge-tree --write-tree`의 2.38 요구)은 확인 안 함.
- 커버리지 수치는 측정하지 않았다.

### Residual-risk
- F2 수정안(merge-tree 재검증)을 적용해도 "main 변경과 카드 변경이 같은 파일에서 충돌 없이 자동 merge되며 의미가 상호작용"하는 경우는 걸러지지 않는다. 이는 어떤 텍스트 수준 검사로도 못 잡는 영역이라 카드 범위 밖이다.
- 증거 파일을 작업 트리에서 읽는 기존 동작(F6)은 남는다.
- 프로브·변이 출력 원본은 머신 로컬 스크래치(`/private/tmp/claude-501/.../scratchpad/`)에 있으며, 판정을 가른 줄은 위에 옮겨 적었다. 나머지 원본은 이 파일 밖에서 사라질 수 있다.

## 반복 이력
- iter1(이번): FAIL 0.70 @7723481e7e. 이전 감사 없음 — 수리 유래(repair-born) 판정 해당 없음.
- Binding run conditions: `progress.md`에 해당 제목 없음(`grep -c` → 0).

## 재감사 범위
F1·F2 수리 커밋과 그 재현 테스트(P1·P7·P2b 형태)만 델타로 본다. F3~F7은 선택 사항이다.
