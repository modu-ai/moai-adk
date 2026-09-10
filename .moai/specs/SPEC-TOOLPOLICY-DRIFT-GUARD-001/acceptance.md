# acceptance.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

모든 명령은 카드 워크트리 최상위(`.claude/worktrees/t619`)에서 실행한다. `$SCRATCH` 는 저장소 트리 밖의 세션 전용 스크래치 디렉터리다. 선택자가 테스트를 하나도 고르지 못해도 `go test` 는 `ok` 를 찍으므로, 테스트 통과 판정은 언제나 `-v` 출력의 `--- PASS:` 줄 개수로 한다.

## §D AC 매트릭스

### AC-TDG-001 — 커밋된 트리에서 검사 통과, 중복·겹침 없음 [MUST-FIX] (maps REQ-TDG-001, REQ-TDG-004)

- **Given** M2 로 정정된 `tool-policy.yaml` 과 바이트 그대로인 `.claude/settings.json`
- **When** 다음을 실행하면
  ```bash
  make tool-policy-drift-check; echo "rc=$?"
  go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$' -count=1 -v > "$SCRATCH/ac001.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_CommittedSettingsMatchYAML' "$SCRATCH/ac001.log"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_NoDuplicatesOrOverlap' "$SCRATCH/ac001.log"
  ```
- **Then** 첫 명령은 `rc=0`, 두 번째도 `rc=0`, 두 `grep -c` 는 각각 `1` 을 출력한다.
- 재현 증거(M1): 정정 **전** 트리에서 같은 `go test` 는 `rc=1` 이고, 로그에 판정서의 20개 명세자가 모두 나타난다(`grep -c -F 'only-in-' "$SCRATCH/ac001-red.log"` → `20`).

### AC-TDG-002 — 생성기 스크래치 실행으로 독립 확인한 집합 일치 [MUST-FIX] (maps REQ-TDG-001, REQ-TDG-002)

- **Given** 스크래치 루트에 복사한 settings.json
  ```bash
  mkdir -p "$SCRATCH/root/.claude" && cp .claude/settings.json "$SCRATCH/root/.claude/settings.json"
  ```
- **When** 생성기를 스크래치 루트 대상으로만 실행하고 목록을 정렬해 비교하면
  ```bash
  moai tool-policy build --repo-root "$SCRATCH/root" --policy .moai/config/sections/tool-policy.yaml --local-only
  for k in allow deny; do
    jq -r ".permissions.$k[]" .claude/settings.json | LC_ALL=C sort > "$SCRATCH/committed.$k"
    jq -r ".permissions.$k[]" "$SCRATCH/root/.claude/settings.json" | LC_ALL=C sort > "$SCRATCH/generated.$k"
    diff "$SCRATCH/committed.$k" "$SCRATCH/generated.$k"; echo "$k diff_rc=$?"
  done
  jq -r '.permissions.ask // [] | length' .claude/settings.json "$SCRATCH/root/.claude/settings.json"
  ```
- **Then** build 출력에 `allow=114 ask=0 deny=48 env_gated_skipped=5` 가 있고, `allow diff_rc=0`, `deny diff_rc=0` 이며, 마지막 명령은 `0` 을 두 줄 출력한다.

### AC-TDG-003 — 적용 권한 불변, 검사는 읽기 전용 [MUST-FIX] (maps REQ-TDG-002, REQ-TDG-003)

- **Given** run 단계 커밋이 끝난 브랜치
- **When** 다음을 실행하면
  ```bash
  git diff --exit-code d1b61005d HEAD -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl; echo "rc=$?"
  shasum -a 256 .claude/settings.json .moai/config/sections/tool-policy.yaml > "$SCRATCH/before.sha"
  make tool-policy-drift-check
  shasum -a 256 .claude/settings.json .moai/config/sections/tool-policy.yaml > "$SCRATCH/after.sha"
  cmp "$SCRATCH/before.sha" "$SCRATCH/after.sha"; echo "cmp_rc=$?"
  ```
- **Then** `git diff` 는 출력 없이 `rc=0`, `cmp_rc=0` 이다.

### AC-TDG-004 — 수동 뮤테이션 대조 (YAML 쪽만) [MUST-FIX] (maps REQ-TDG-003, REQ-TDG-004)

- **Given** AC-TDG-001 이 통과한 트리와, 복원용으로 저장한 YAML 사본과 그 sha256
  ```bash
  cp .moai/config/sections/tool-policy.yaml "$SCRATCH/tool-policy.yaml.bak"
  shasum -a 256 .moai/config/sections/tool-policy.yaml | cut -d' ' -f1 > "$SCRATCH/yaml.sha"
  ```
- **When** YAML 에서 `ExitWorktree` allow 항목 한 개를 지우고 검사를 실행하면
  ```bash
  python3 - <<'PY'
  p = ".moai/config/sections/tool-policy.yaml"
  lines = open(p, encoding="utf-8").read().split("\n")
  start = next(i for i, l in enumerate(lines) if l == '  - tool: "ExitWorktree"')
  end = next((i for i in range(start + 1, len(lines)) if lines[i].startswith("  - tool: ")), len(lines))
  del lines[start:end]
  open(p, "w", encoding="utf-8").write("\n".join(lines))
  PY
  make tool-policy-drift-check > "$SCRATCH/mut.log" 2>&1; echo "rc=$?"
  grep -c -F 'tool-policy drift' "$SCRATCH/mut.log"
  grep -c -F 'ExitWorktree' "$SCRATCH/mut.log"
  ```
- **Then** `rc` 는 0 이 아니다(GNU make 는 `2`). 두 `grep -c` 는 각각 `1` 이상이다. 뮤테이션은 커밋되지 않은 워킹 트리 수정이므로, 붉은색은 검사가 워킹 트리를 읽는다는 증거이기도 하다.
- **When** 저장한 사본으로 복원하고 다시 실행하면
  ```bash
  cp "$SCRATCH/tool-policy.yaml.bak" .moai/config/sections/tool-policy.yaml
  shasum -a 256 .moai/config/sections/tool-policy.yaml | cut -d' ' -f1 | cmp - "$SCRATCH/yaml.sha"; echo "sha_rc=$?"
  make tool-policy-drift-check > "$SCRATCH/restored.log" 2>&1; echo "rc=$?"
  grep -c -F 'tool-policy drift' "$SCRATCH/restored.log"
  git diff --exit-code -- .moai/config/sections/tool-policy.yaml; echo "diff_rc=$?"
  ```
- **Then** `sha_rc=0`, `rc=0`, `grep -c` 는 `0`, `diff_rc=0` 이다. 복원 뒤 `grep -c` 의 `0` 은 뮤테이션 때 같은 문구를 잡은 `1` 이상과 짝을 이루므로 공허하지 않다.
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

### AC-TDG-006 — 누락·빈 입력에서 실패, 건너뛰기 없음 [MUST-FIX] (maps REQ-TDG-004)

- **Given** 새 테스트 파일
- **When** 다음을 실행하면
  ```bash
  go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift_FailClosed$' -count=1 -v > "$SCRATCH/ac006.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyDrift_FailClosed/' "$SCRATCH/ac006.log"
  grep -c -F -- '--- SKIP:' "$SCRATCH/ac006.log"
  grep -c -F 't.Skip' internal/config/toolpolicy/drift_check_test.go
  grep -c -F 'TestToolPolicyDrift' internal/config/toolpolicy/drift_check_test.go
  ```
- **Then** `rc=0`, 첫 `grep -c` 는 `5`(`missing_yaml`, `missing_settings`, `no_permissions_region`, `empty_allow`, `empty_deny` — 각각 비교 함수가 오류를 돌려줌을 단정), `--- SKIP:` 는 `0`, `t.Skip` 는 `0`, 양성 대조인 `TestToolPolicyDrift` 는 `1` 이상이다.

### AC-TDG-007 — 주장 정정 전수 [MUST-FIX] (maps REQ-TDG-005)

- **Given** M5 가 끝난 트리
- **When** 다음을 실행하면
  ```bash
  grep -c -F 'structurally preventing' .moai/config/sections/tool-policy.yaml internal/config/toolpolicy/types.go
  grep -c -F 'are both generated' .moai/config/sections/tool-policy.yaml
  grep -c -F 'settings.json.tmpl#permissions' .moai/config/sections/tool-policy.yaml
  grep -c -F 'tool-policy-drift-check' .moai/config/sections/tool-policy.yaml internal/config/toolpolicy/types.go
  grep -c -F '.claude/settings.json#permissions' .moai/config/sections/tool-policy.yaml
  ```
- **Then**
  - 첫 명령: `.moai/config/sections/tool-policy.yaml:0`, `internal/config/toolpolicy/types.go:0`
  - 둘째, 셋째: `0`
  - 넷째(양성 대조): 두 파일 모두 `1` 이상
  - 다섯째(같은 파일에서 grep 이 동작함을 보이는 대조): `1` 이상
- 기준선: 정정 전 트리에서 첫 명령은 두 파일 모두 `1`, 셋째 명령은 `1` 이다(plan 단계 판독: `tool-policy.yaml:7-8`, `types.go:97`, `tool-policy.yaml:54`). run 단계에서 정정 전에 한 번 재서 §E.2 에 기록한다.

### AC-TDG-008 — `make build` 선행 연결과 CI 경로 필터 [MUST-FIX] (maps REQ-TDG-003)

- **Given** M4 가 끝난 트리
- **When** 다음을 실행하면
  ```bash
  make -n build | grep -c -F 'tool-policy drift'
  make -n build | grep -c -F 'agent-emit drift'
  grep -c -F -- "- '.claude/settings.json'" .github/workflows/ci.yml
  grep -c -F -- "- '.moai/**'" .github/workflows/ci.yml
  ```
- **Then** 첫째는 `1` 이상, 둘째(기존 선행 검사가 같은 방식으로 보이는지 확인하는 대조)는 `1` 이상, 셋째는 `1`, 넷째(필터 블록이 판독되는지 확인하는 대조)는 `1` 이다.

### AC-TDG-009 — 영향받는 기존 테스트 무회귀 [MUST-FIX] (maps REQ-TDG-001, REQ-TDG-005)

- **Given** M2·M5 가 끝난 트리
- **When** 다음을 실행하면
  ```bash
  go test -count=1 ./internal/config/toolpolicy/...
  go test -count=1 -timeout 600s ./internal/cli/ -run 'TestToolPolicy' -v > "$SCRATCH/ac009-cli.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestToolPolicyList_QueryFilters' "$SCRATCH/ac009-cli.log"
  go test -count=1 ./internal/config/ -run 'TestAuditLoaderCompleteness' -v > "$SCRATCH/ac009-cfg.log" 2>&1; echo "rc=$?"
  grep -c -F -- '--- PASS: TestAuditLoaderCompleteness' "$SCRATCH/ac009-cfg.log"
  ```
- **Then** 첫째는 `ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy` 줄, 나머지 둘은 `rc=0`, 두 `grep -c` 는 각각 `1` 이상이다.

### AC-TDG-010 — YAML 항목 단위 결과 [MUST-FIX] (maps REQ-TDG-001)

- **Given** M2 가 끝난 YAML
- **When** 다음을 실행하면
  ```bash
  moai tool-policy list --policy .moai/config/sections/tool-policy.yaml --format json > "$SCRATCH/list.json"; echo "rc=$?"
  jq '[.[] | select(.tool == "MultiEdit")] | length' "$SCRATCH/list.json"
  jq '[.[] | select(.decision == "deny" and (.tool == "Glob" or .tool == "Grep" or .tool == "Write"))] | length' "$SCRATCH/list.json"
  for t in CronCreate CronDelete CronList EnterPlanMode ExitPlanMode EnterWorktree ExitWorktree; do
    printf '%s ' "$t"; jq --arg t "$t" '[.[] | select(.tool == $t and .decision == "allow")] | length' "$SCRATCH/list.json"
  done
  jq '[.[] | select(.tool == "Read" and .decision == "deny")] | length' "$SCRATCH/list.json"
  ```
- **Then** `rc=0`(스키마 로딩 성공), `MultiEdit` 는 `0`, Glob/Grep/Write deny 는 `0`, 7개 도구는 각각 `1`, 마지막 대조(Read deny 가 남아 있음)는 `5` 다. 이 값은 plan 단계에 정정 전 YAML 로 실측했다(`moai tool-policy list ... | jq` → `5`). M2 는 Read 항목을 건드리지 않으므로 같은 값이 유지돼야 한다.

## §D.1 경계 사례

- `ask` 키가 없는 settings.json 과 ask 0개를 유도하는 YAML 은 같다(REQ-TDG-001).
- env_gate 항목은 settings.json 에 없어도 차이가 아니다(생성기와 같은 규칙).
- `defaultMode` 가 달라도 차이가 아니다(YAML 에서 유도되지 않음).
- 목록 순서·들여쓰기 차이는 차이가 아니다(집합 비교).

## §D.2 품질 게이트

- `go vet ./internal/config/toolpolicy/...` 오류 0.
- `golangci-lint run ./internal/config/toolpolicy/...` 신규 지적 0.
- 로컬 검증은 건드린 패키지로 한정하고, 전체 스위트 판정은 CI 에 맡긴다.

## §D.3 Definition of Done

- [ ] AC-TDG-001..010 모두 증거와 함께 통과, progress.md §E.2 에 명령·출력 기록
- [ ] M1 붉은색 재현(20개 명세자) 기록
- [ ] 뮤턴트 점검(항상 빈 차이) 결과 기록
- [ ] `.claude/settings.json` 과 템플릿 바이트 불변
- [ ] 산출물·코드에 역슬래시-u 이스케이프 표기와 Cf 범주 문자 0개
