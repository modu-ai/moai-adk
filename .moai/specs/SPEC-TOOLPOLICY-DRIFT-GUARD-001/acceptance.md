# acceptance.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

모든 명령은 카드 워크트리 최상위(`.claude/worktrees/t619`)에서 실행한다. `$SCRATCH` 는 저장소 트리 밖의 세션 전용 스크래치 디렉터리다. 워크트리 세션 가드는 `trap`, heredoc 으로 인터프리터에 스크립트를 넘기는 형태, 변수로 계산된 스크립트 경로를 python 에 넘기는 형태를 거부하므로(2026-09-10 실측), python 호출의 스크립트·대상 경로는 리터럴 절대 경로로 적는다. 선택자가 테스트를 하나도 고르지 못해도 `go test` 는 `ok` 를 찍으므로, 테스트 통과 판정은 언제나 `-v` 출력의 `--- PASS:` 줄 개수로 한다.

기준선 표기: "현재" 값은 2026-09-10, HEAD `b2cbfd207` 워크트리에서 실측한 값이다.

## §D AC 매트릭스

### AC-TDG-001 — 커밋된 트리에서 검사 통과, 중복·겹침 없음, 정정 전 붉은색 20개 [MUST-FIX] (maps REQ-TDG-001, REQ-TDG-004)

- **Given (M1 직후, M2 전)** 검사 테스트가 추가되고 YAML 은 아직 정정되지 않은 트리
- **When** 붉은색 실행을 기록하면
  ```bash
  go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_CommittedSettingsMatchYAML$' -count=1 -v > "$SCRATCH/ac001-red.log" 2>&1; echo "rc=$?"
  grep -o -E '(allow|ask|deny) only-in-(yaml|settings): .+' "$SCRATCH/ac001-red.log" | sort -u > "$SCRATCH/ac001-red.uniq"
  wc -l < "$SCRATCH/ac001-red.uniq"
  grep -c -F 'allow only-in-settings: ' "$SCRATCH/ac001-red.uniq"
  grep -c -F 'allow only-in-yaml: ' "$SCRATCH/ac001-red.uniq"
  grep -c -F 'deny only-in-yaml: ' "$SCRATCH/ac001-red.uniq"
  grep -c -F 'deny only-in-settings: ' "$SCRATCH/ac001-red.uniq"
  grep -c -F -- '--- FAIL: TestToolPolicyDrift_CommittedSettingsMatchYAML' "$SCRATCH/ac001-red.log"
  ```
- **Then** `rc=1`, 서로 다른 차이 줄 `20`, 분할은 `7` / `1` / `12` / `0`, `--- FAIL:` 줄 `1`. 분할 기대값의 근거는 판정서 §2.2 의 `comm -3` 목록(allow 소실 7, allow 신규 `MultiEdit` 1, deny 신규 12, deny 소실 0)이다. `sort -u` 뒤 개수이므로 같은 줄 반복으로는 20이 되지 않는다.
- **Given (M2 직후)** 정정된 `tool-policy.yaml` 과 바이트 그대로인 `.claude/settings.json`
- **When** 다음을 실행하면
  ```bash
  make tool-policy-drift-check; echo "rc=$?"
  go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$' -count=1 -v > "$SCRATCH/ac001.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_CommittedSettingsMatchYAML' "$SCRATCH/ac001.log"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_NoDuplicatesOrOverlap' "$SCRATCH/ac001.log"
  grep -c -F 'only-in-' "$SCRATCH/ac001.log"
  ```
- **Then** `rc=0` 두 번, 두 `--- PASS:` 개수는 각각 `1`, `only-in-` 는 `0` 이다. 마지막 `0` 은 앞 붉은색 실행의 `20` 과 짝을 이루므로 공허하지 않다.

### AC-TDG-002 — 생성기 스크래치 실행으로 독립 확인한 집합 일치 [MUST-FIX] (maps REQ-TDG-001, REQ-TDG-002)

- **Given** 스크래치 루트에 복사한 settings.json
  ```bash
  mkdir -p "$SCRATCH/root/.claude" && cp .claude/settings.json "$SCRATCH/root/.claude/settings.json"
  ```
- **When** 생성기를 스크래치 루트 대상으로만 실행하고 목록을 정렬해 비교하면
  ```bash
  moai tool-policy build --repo-root "$SCRATCH/root" --policy .moai/config/sections/tool-policy.yaml --local-only
  jq -r '.permissions.allow[]' .claude/settings.json | LC_ALL=C sort > "$SCRATCH/committed.allow"
  jq -r '.permissions.allow[]' "$SCRATCH/root/.claude/settings.json" | LC_ALL=C sort > "$SCRATCH/generated.allow"
  diff "$SCRATCH/committed.allow" "$SCRATCH/generated.allow"; echo "allow diff_rc=$?"
  jq -r '.permissions.deny[]' .claude/settings.json | LC_ALL=C sort > "$SCRATCH/committed.deny"
  jq -r '.permissions.deny[]' "$SCRATCH/root/.claude/settings.json" | LC_ALL=C sort > "$SCRATCH/generated.deny"
  diff "$SCRATCH/committed.deny" "$SCRATCH/generated.deny"; echo "deny diff_rc=$?"
  jq -r '.permissions.ask // [] | length' .claude/settings.json "$SCRATCH/root/.claude/settings.json"
  ```
- **Then** build 출력에 `allow=114 ask=0 deny=48 env_gated_skipped=5` 가 있고, `allow diff_rc=0`, `deny diff_rc=0` 이며, 마지막 명령은 `0` 을 두 줄 출력한다.
- 기준선(현재): 같은 build 가 `allow=108 ask=0 deny=60 env_gated_skipped=5` 를 출력한다(plan 단계와 1회차 감사 양쪽에서 측정). `env_gated_skipped=5` 는 M2 가 env_gate 항목을 건드리지 않으므로 그대로여야 한다.

### AC-TDG-003 — 적용 권한 불변, 검사는 읽기 전용 [MUST-FIX] (maps REQ-TDG-002, REQ-TDG-003)

- **Given** run 단계 커밋이 끝난 카드 브랜치(develop 흡수 전이든 후든)
- **When** 기준을 흡수 대상과의 merge-base 로 잡고 비교하면
  ```bash
  git merge-base HEAD develop
  ```
  출력된 SHA 를 `<BASE>` 에 그대로 적어
  ```bash
  git diff --exit-code <BASE> HEAD -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl; echo "rc=$?"
  shasum -a 256 .claude/settings.json .moai/config/sections/tool-policy.yaml > "$SCRATCH/before.sha"
  make tool-policy-drift-check
  shasum -a 256 .claude/settings.json .moai/config/sections/tool-policy.yaml > "$SCRATCH/after.sha"
  cmp "$SCRATCH/before.sha" "$SCRATCH/after.sha"; echo "cmp_rc=$?"
  ```
- **Then** `git diff` 는 출력 없이 `rc=0`, `cmp_rc=0` 이다. develop 을 흡수한 뒤에는 merge-base 가 흡수 지점으로 옮겨가므로, develop 쪽 변경은 비교에 섞이지 않고 이 카드의 변경만 본다.
- 기준선(현재): `git merge-base HEAD develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`, 그 SHA 로 같은 `git diff --exit-code` 를 실행하면 출력이 없다.

### AC-TDG-004 — 수동 뮤테이션 대조 (YAML 쪽만) [MUST-FIX] (maps REQ-TDG-003, REQ-TDG-004)

단계마다 별도의 평범한 명령으로 실행한다. `trap` 과 heredoc 인터프리터 호출은 워크트리 가드가 거부하므로 쓰지 않는다. 중단 시 복원 규칙은 plan.md §G 를 따른다.

- **Given** M2 가 **커밋된** 트리(AC-TDG-001 통과), 그리고 스크래치에 둔 변이 스크립트 `mutate.py`. 스크립트는 `cat > <스크래치 절대 경로>/mutate.py` 로 만들고 내용은 아래와 같다.
  ```python
  import sys

  path, tool = sys.argv[1], sys.argv[2]
  lines = open(path, encoding="utf-8").read().split("\n")
  marker = '  - tool: "%s"' % tool
  starts = [i for i, l in enumerate(lines) if l == marker]
  if len(starts) != 1:
      sys.exit("expected exactly one entry for %s, found %d" % (tool, len(starts)))
  start = starts[0]
  end = next((i for i in range(start + 1, len(lines)) if lines[i].startswith("  - tool: ")), len(lines))
  del lines[start:end]
  open(path, "w", encoding="utf-8").write("\n".join(lines))
  ```
- **When** 백업하고 원본 sha 를 기록한 뒤(이 두 값은 변이 전에 progress.md §E.2 에 적는다) `ExitWorktree` 항목 하나를 지우고 검사를 실행하면
  ```bash
  cp .moai/config/sections/tool-policy.yaml "$SCRATCH/tool-policy.yaml.bak"
  shasum -a 256 .moai/config/sections/tool-policy.yaml | cut -d' ' -f1 > "$SCRATCH/yaml.orig.sha"
  python3 <스크래치 절대 경로>/mutate.py <워크트리 절대 경로>/.moai/config/sections/tool-policy.yaml ExitWorktree; echo "mutate_rc=$?"
  shasum -a 256 .moai/config/sections/tool-policy.yaml | cut -d' ' -f1 > "$SCRATCH/yaml.mut.sha"
  cmp -s "$SCRATCH/yaml.orig.sha" "$SCRATCH/yaml.mut.sha"; echo "mutation_applied_cmp_rc=$?"
  make tool-policy-drift-check > "$SCRATCH/mut.log" 2>&1; echo "rc=$?"
  grep -c -F 'allow only-in-settings: ExitWorktree' "$SCRATCH/mut.log"
  grep -o -E '(allow|ask|deny) only-in-(yaml|settings): .+' "$SCRATCH/mut.log" | sort -u | wc -l
  ```
- **Then** `mutate_rc=0`, `mutation_applied_cmp_rc=1`(변이가 실제로 파일을 바꿈), make `rc` 는 0 이 아니다(GNU make 는 `2`), `allow only-in-settings: ExitWorktree` 는 `1` 이상, 서로 다른 차이 줄은 정확히 `1` 이다. 뮤테이션은 커밋되지 않은 워킹 트리 수정이므로, 붉은색은 검사가 워킹 트리를 읽는다는 증거이기도 하다.
- **When** 복원 전 변이 상태를 확인하고, 저장한 사본으로 복원한 뒤 다시 실행하면
  ```bash
  shasum -a 256 .moai/config/sections/tool-policy.yaml | cut -d' ' -f1 | cmp -s - "$SCRATCH/yaml.mut.sha"; echo "still_mutated_cmp_rc=$?"
  cp "$SCRATCH/tool-policy.yaml.bak" .moai/config/sections/tool-policy.yaml
  shasum -a 256 .moai/config/sections/tool-policy.yaml | cut -d' ' -f1 | cmp -s - "$SCRATCH/yaml.orig.sha"; echo "sha_rc=$?"
  make tool-policy-drift-check > "$SCRATCH/restored.log" 2>&1; echo "rc=$?"
  grep -c -F 'only-in-' "$SCRATCH/restored.log"
  ```
  그리고 별도 명령으로
  ```bash
  git diff --exit-code -- .moai/config/sections/tool-policy.yaml; echo "diff_rc=$?"
  ```
- **Then** `still_mutated_cmp_rc=0`(복원 직전 파일이 자신이 만든 변이 상태), `sha_rc=0`, `rc=0`, `only-in-` 는 `0`, `diff_rc=0` 이다. 복원 뒤 `0` 은 변이 때 `1` 과 짝을 이루므로 공허하지 않다.
- 변이 스크립트 실측(현재, 스크래치 사본 대상, `TodoWrite` 항목): 항목 줄 `173` → `172`, 삭제 뒤 `tool: "TodoWrite"` `0`, `moai tool-policy list` 로딩 `172`, 백업 복원 뒤 sha `cmp` 종료 `0`, 없는 도구(`NoSuchTool`) 지정 시 종료 `1`.
- 금지: `.claude/settings.json` 뮤테이션(실행 중 세션이 읽는다), `git restore` / `git checkout --` 로 복원(미커밋 작업을 버린다).

### AC-TDG-005 — 자동 뮤테이션 대조: 양방향 + 중복·겹침 [MUST-FIX] (maps REQ-TDG-004)

- **Given** 커밋된 YAML 과 settings.json 을 `t.TempDir()` 로 복사해 한 곳씩 바꾸는 하위 테스트
- **When** 다음을 실행하면
  ```bash
  go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_Mutation$' -count=1 -v > "$SCRATCH/ac005.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_Mutation/' "$SCRATCH/ac005.log"
  grep -c -F -- '--- FAIL:' "$SCRATCH/ac005.log"
  ```
- **Then** `rc=0`, 첫 `grep -c` 는 `5`(`settings_side_missing`, `yaml_side_flip`, `duplicate_in_settings`, `allow_deny_overlap`, `unmutated`), 두 번째는 `0` 이다.
- 각 하위 테스트가 단정하는 내용:
  - `settings_side_missing`: settings 사본에서 allow 한 개를 지우면 차이가 정확히 1개, 그 명세자 이름 포함
  - `yaml_side_flip`: YAML 사본에서 한 항목의 결정을 바꾸면 그 명세자가 차이에 포함
  - `duplicate_in_settings`: settings 사본의 한 목록에 명세자를 하나 더 넣으면 실패 보고
  - `allow_deny_overlap`: 같은 명세자를 allow 와 deny 에 함께 두면 실패 보고
  - `unmutated`: 바꾸지 않은 사본은 차이 0개
- 뮤턴트 점검(DoD): 비교 함수가 항상 빈 차이를 돌려주도록 바꾸면 이 테스트가 `rc≠0` 이 된다. 결과는 progress.md §E.2 에 기록한다.

### AC-TDG-006 — 누락·해석 실패·권한 블록 부재·네 집합 비어 있음에서 실패, 건너뛰기 없음 [MUST-FIX] (maps REQ-TDG-004)

- **Given** 새 테스트 파일
- **When** 다음을 실행하면
  ```bash
  go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_FailClosed$' -count=1 -v > "$SCRATCH/ac006.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_FailClosed/' "$SCRATCH/ac006.log"
  grep -c -F -- '--- SKIP:' "$SCRATCH/ac006.log"
  grep -c -F 't.Skip' internal/config/toolpolicy/drift_check_test.go
  grep -c -F 'TestToolPolicyDrift' internal/config/toolpolicy/drift_check_test.go
  ```
- **Then** `rc=0`, 첫 `grep -c` 는 `9`, `--- SKIP:` 는 `0`, `t.Skip` 는 `0`, 양성 대조인 `TestToolPolicyDrift` 는 `1` 이상이다.
- 하위 테스트 9개와 각 입력(모두 비교 함수가 오류를 돌려줌을 단정):
  - `missing_yaml`: YAML 경로에 파일 없음
  - `missing_settings`: settings 경로에 파일 없음
  - `malformed_yaml`: 문법이 깨진 YAML (예: 닫히지 않은 따옴표)
  - `malformed_settings_json`: `"permissions"` 키는 있으나 그 객체의 JSON 이 깨진 settings
  - `no_permissions_region`: 올바른 JSON 이지만 `"permissions"` 키가 없는 settings
  - `empty_yaml_allow`: deny 항목만 있는 YAML
  - `empty_yaml_deny`: allow 항목만 있는 YAML
  - `empty_settings_allow`: `allow` 가 빈 배열인 settings
  - `empty_settings_deny`: `deny` 가 빈 배열인 settings
- 뮤턴트 점검(DoD): 해석 오류를 빈 정책으로 삼키도록 바꾸면 `malformed_yaml`·`malformed_settings_json` 이 실패해야 한다.

### AC-TDG-007 — 주장 정정 다섯 곳 전수 [MUST-FIX] (maps REQ-TDG-005)

- **Given** M5 가 끝난 트리
- **When** 다음을 실행하면
  ```bash
  grep -c -F 'structurally preventing' .moai/config/sections/tool-policy.yaml internal/config/toolpolicy/types.go
  grep -c -F 'are both generated' .moai/config/sections/tool-policy.yaml
  grep -c -F 'ANALOGOUS drift class' .moai/config/sections/tool-policy.yaml
  grep -c -F 'comment (audit surface) are generated' internal/config/toolpolicy/types.go
  grep -c -F 'settings.json.tmpl#permissions' .moai/config/sections/tool-policy.yaml
  grep -c -F 'tool-policy-drift-check' .moai/config/sections/tool-policy.yaml internal/config/toolpolicy/types.go
  grep -c -F '.claude/settings.json#permissions' .moai/config/sections/tool-policy.yaml
  grep -c -F 'Package toolpolicy' internal/config/toolpolicy/types.go
  ```
- **Then**

  | # | 명령 | 정정 뒤 기대값 | 현재 실측값 | 역할 |
  |---|---|---|---|---|
  | 1 | `structurally preventing` | 두 파일 모두 `0` | `tool-policy.yaml:1`, `types.go:1` | 대상 1·4 부재 |
  | 2 | `are both generated` | `0` | `1` | 대상 1 부재 |
  | 3 | `ANALOGOUS drift class` | `0` | `1` | 대상 2 부재 |
  | 4 | `comment (audit surface) are generated` | `0` | `1` | 대상 3 부재 |
  | 5 | `settings.json.tmpl#permissions` | `0` | `1` | 대상 5 부재 |
  | 6 | `tool-policy-drift-check` | 두 파일 모두 `1` 이상 | 두 파일 모두 `0` | 정정 문구가 검사를 가리킴(양성) |
  | 7 | `.claude/settings.json#permissions` | `1` 이상 | `171` | 같은 YAML 에서 grep 이 동작함을 보이는 대조 |
  | 8 | `Package toolpolicy` | `1` | `1` | 같은 types.go 에서 grep 이 동작함을 보이는 대조 |

- 토큰은 모두 한 줄 안에 있다(현재 판독: `tool-policy.yaml:7`, `:8`, `:14`, `:54`, `types.go:5`, `:97`). 정정 문구가 줄바꿈 위치를 바꿔도 부재 grep 이 우연히 0 이 되지 않도록, 1-5 번이 `0` 이 된 이유가 해당 문장 삭제·교체임을 정정 diff 로 확인해 §E.2 에 적는다.

### AC-TDG-008 — `make build` 선행 연결과 CI 경로 필터 블록 [MUST-FIX] (maps REQ-TDG-003)

- **Given** M4 가 끝난 트리
- **When** 다음을 실행하면
  ```bash
  make -n build 2>/dev/null | grep -c -F 'TestToolPolicyDrift_'
  make -n build 2>/dev/null | grep -c -F 'TestGoldenCommittedArtifactsMatchEmission'
  awk '/^            go_code:$/{f=1;next} f&&!/^              - /{exit} f' .github/workflows/ci.yml > "$SCRATCH/gocode.txt"
  wc -l < "$SCRATCH/gocode.txt"
  grep -c -F -- "- '.claude/settings.json'" "$SCRATCH/gocode.txt"
  grep -c -F -- "- '.moai/**'" "$SCRATCH/gocode.txt"
  ```
- **Then**

  | # | 명령 | M4 뒤 기대값 | 현재 실측값 | 역할 |
  |---|---|---|---|---|
  | 1 | `make -n build` 의 `TestToolPolicyDrift_` | `1` 이상 | `0` | 검사 본체가 build 선행에 연결됨 |
  | 2 | `make -n build` 의 `TestGoldenCommittedArtifactsMatchEmission` | `2` | `2` | 기존 선행 검사가 같은 방식으로 보이는지 대조 |
  | 3 | `go_code:` 블록 줄 수 | `15` | `14` | 블록 추출 자체가 동작함을 보이는 대조 |
  | 4 | 블록 안 `.claude/settings.json` | `1` | `0` | 필터 블록 안에 추가됨 |
  | 5 | 블록 안 `.moai/**` | `1` | `1` | 블록 안 grep 대조 |

### AC-TDG-009 — 영향받는 기존 테스트 무회귀 [MUST-FIX] (maps REQ-TDG-001, REQ-TDG-005)

- **Given** M2·M5 가 끝난 트리
- **When** 다음을 실행하면
  ```bash
  go test -count=1 -v ./internal/config/toolpolicy/... > "$SCRATCH/ac009-pkg.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestRoundTripEquivalence_JSON' "$SCRATCH/ac009-pkg.log"
  grep -c -F -- '--- FAIL:' "$SCRATCH/ac009-pkg.log"
  go test -count=1 -timeout 600s ./internal/cli/ -run 'TestToolPolicy' -v > "$SCRATCH/ac009-cli.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyList_QueryFilters' "$SCRATCH/ac009-cli.log"
  go test -count=1 ./internal/config/ -run 'TestAuditLoaderCompleteness' -v > "$SCRATCH/ac009-cfg.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestAuditLoaderCompleteness' "$SCRATCH/ac009-cfg.log"
  ```
- **Then** 세 `rc=0`, `TestRoundTripEquivalence_JSON`·`TestToolPolicyList_QueryFilters`·`TestAuditLoaderCompleteness` 의 `--- PASS:` 는 각각 `1` 이상, 패키지 로그의 `--- FAIL:` 은 `0` 이다.
- 대조(현재): 세 테스트 함수 선언은 각각 `internal/config/toolpolicy/codegen_test.go`, `internal/cli/tool_policy_test.go`, `internal/config/audit_loader_completeness_test.go` 에 `^func …(` 형태로 `1` 개씩 있다.

### AC-TDG-010 — YAML 항목 단위 결과, env_gate 항목 보존 [MUST-FIX] (maps REQ-TDG-001)

- **Given** M2 가 끝난 YAML
- **When** 다음을 실행하면
  ```bash
  moai tool-policy list --policy .moai/config/sections/tool-policy.yaml --format json > "$SCRATCH/list.json"; echo "rc=$?"
  jq '[.[] | select(.tool == "MultiEdit")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.decision == "deny" and (.tool == "Glob" or .tool == "Grep" or .tool == "Write") and .env_gate == null and (.args_pattern == "./secrets/**" or .args_pattern == "~/.ssh/**" or .args_pattern == "~/.aws/**" or .args_pattern == "~/.config/gcloud/**"))] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "Write" and .decision == "deny" and .env_gate != null)] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.env_gate != null)] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "CronCreate" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "CronDelete" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "CronList" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "EnterPlanMode" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "ExitPlanMode" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "EnterWorktree" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "ExitWorktree" and .decision == "allow")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.tool == "Read" and .decision == "deny")] | length' "$SCRATCH/list.json"
  ```
- **Then**

  | # | 질의 | M2 뒤 기대값 | 현재 실측값 | 역할 |
  |---|---|---|---|---|
  | 0 | `list` 종료 코드 | `rc=0` | `rc=0` | 스키마 로딩 성공 |
  | 1 | `MultiEdit` | `0` | `1` | 삭제 확인 |
  | 2 | env_gate 없음 + 네 경로의 Glob/Grep/Write deny | `0` | `12` | 삭제 12개 확인 |
  | 3 | env_gate 있는 Write deny | `1` | `1` | env_gate 항목 보존(양성 대조) |
  | 4 | env_gate 항목 전체 | `5` | `5` | env_gate 항목 전부 보존 |
  | 5-11 | 7개 도구 allow 각각 | `1` | `0` | 추가 확인 |
  | 12 | Read deny | `5` | `5` | M2 가 건드리지 않는 항목 대조 |

## §D.1 경계 사례

- `ask` 키가 없는 settings.json 과 ask 0개를 유도하는 YAML 은 같다(REQ-TDG-001). ask 가 빈 것은 실패 사유가 아니다(REQ-TDG-004).
- env_gate 항목은 settings.json 에 없어도 차이가 아니다(생성기와 같은 규칙).
- `defaultMode` 가 달라도 차이가 아니다(YAML 에서 유도되지 않음).
- 목록 순서·들여쓰기 차이는 차이가 아니다(집합 비교).

## §D.2 품질 게이트

- `go vet ./internal/config/toolpolicy/...` 오류 0.
- `golangci-lint run ./internal/config/toolpolicy/...` 신규 지적 0.
- 로컬 검증은 건드린 패키지로 한정하고, 전체 스위트 판정은 CI 에 맡긴다.

## §D.3 Definition of Done

- [ ] AC-TDG-001..010 모두 증거와 함께 통과, progress.md §E.2 에 명령·출력 기록
- [ ] M1 붉은색 재현(서로 다른 20줄, 7/1/12/0 분할) 기록
- [ ] 뮤턴트 점검 두 가지(항상 빈 차이, 해석 오류 삼키기) 결과 기록
- [ ] `.claude/settings.json` 과 템플릿 바이트 불변(merge-base 기준)
- [ ] env_gate 항목 5개 보존
- [ ] 산출물·코드에 역슬래시-u 이스케이프 표기와 Cf 범주 문자 0개
