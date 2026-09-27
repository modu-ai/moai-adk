# t1175 회귀 수리 — SPEC-ALWAYS-LOADED-DIET-002 병합 후 가드 복구

- 카드: t1175 (회귀 수리, 리드 배정)
- 트리: `.claude/worktrees/t1175-fix`, 브랜치 `WT-diet-regress-fix`
- 기준: HEAD `7fe658815`(t1175 병합, CI 실패), 비교 기준선 `f01c7889a`(병합 제1부모, t1175 이전 develop)
- 방법: cycle_type=ddd — 동작 보존 가드를 되살리되 다이어트의 목적(상시 로드 표면 축소)은 유지
- 수리 커밋: `1b68c2633`(소스·테스트), 이 보고서와 증거는 그 뒤 커밋

## Claim

병합 `7fe658815` 에서 깨진 네 갈래(헌법 검증 9건 DRIFT, `internal/constitution` 동기 가드, `internal/cli` MCP 카탈로그 테스트 2건, `internal/harness/rosterguard` 3건)를 모두 닫았다. 판단 원칙은 다음과 같다.

- 문구가 **바뀌기만 한** 조항(대소문자·구두점·어순)은 레지스트리가 지목하는 파일에 **원문을 복원**했다. 상시 로드 파일에 남아 있던 조항이므로 위치를 옮길 이유가 없다.
- 절 전체가 **설계 문서의 판정에 따라 컴패니언으로 옮겨진** 경우는, 옮겨간 자리에 조항이 축자로 존재함을 grep 으로 확인한 뒤 **레지스트리·테스트의 대상을 새 위치로 옮겼다(re-point)**.
- 구속 조항 동결 해시(AC-ALD2-002)는 수리 전후 **동일**하다. 복원한 줄 가운데 `[HARD]`·`MUST`·`shall ` 을 담은 줄이 없기 때문이다.

### 항목별 판정

| # | 항목 | 선택 | 근거(조항이 지금 있는 곳) | 이유 |
|---|---|---|---|---|
| 1 | CONST-V3R2-013~017 (`CLAUDE.md` §7 Rule 1~5) | 복원 | `CLAUDE.md`·템플릿 사본 §7, 각 Rule 본문 첫 글자를 대문자로 되돌림 (`Before`/`When`/`After`/`Write`/`When`) | 다이어트가 조항을 옮기지 않고 첫 글자만 소문자로 바꿨다. 레지스트리가 요구하는 문구가 같은 파일·같은 절에 거의 그대로 있으므로 대소문자 5글자 복원이 최소 수리다 |
| 2 | CONST-V3R2-033 (`moai-constitution.md` #agent-core-behaviors) | 복원 | `### 4. Enforce Simplicity` 첫 문단 — `Actively resist overcomplexity. The natural tendency of code generation is toward over-engineering. Resist it.` | 다이어트가 두 문장을 세미콜론으로 합치면서 `Resist it.` 를 뺐다. 행동 조항(Agent Core Behaviors)이므로 상시 로드 파일에 원문을 되돌렸다. 나머지 압축 문장은 그대로 둔다 |
| 3 | CONST-V3R2-049 (`agent-common-protocol.md` #skeptical-evaluation-stance) | 대상 이동 | `agent-common-protocol-reference.md` 316행 `### Skeptical Evaluation Stance`, 320행 `The reviewer mode operates as a fresh-judgment auditor:` (두 사본 모두 1회) | 설계 문서 `design.md` R-01 이 이 절을 「구속 없음(판단)」으로 분류해 통째로 옮겼다. 원래도 `[HARD]` 표지가 없던 서술형 절이다. 레지스트리의 의미는 「규칙 트리 안의 조항 열거」(`zone-registry.md` 머리말)이며 상시 로드를 요구하는 필드(`zone_class: evolvable-tuning`, `canary_gate: false`)가 아니다. 따라서 `file:` 만 컴패니언으로 옮겨도 의무가 약해지지 않는다. 튜플 다이제스트는 `file`/`anchor` 를 포함하지 않아 변하지 않는다 |
| 4 | CONST-V3R2-152 (`session-handoff.md` #auto-memory-integration-mandatory) | 복원 | 81행 `1. Save the message to a memory project entry. Filename pattern: \`project_<epic>_<spec>_<status>.md\` (…` | 「entry named …」로 바뀐 어구를 원래 「entry. Filename pattern: …」로 되돌렸다. 조항은 제자리에 있었고 어구만 달라졌다 |
| 5 | CONST-V3R2-153 (`session-handoff.md` #canonical-format-verbatim-spec) | 복원(한 줄) | `## Canonical Format (Verbatim Spec)` 코드 블록 바로 뒤 한 줄 — `` The `✂` symbol (U+2702 BLACK SCISSORS) is **preserved verbatim across all locales** — never translate or substitute (full marker spec: …) `` | 마커 명세 전체는 `session-handoff-format.md` 로 옮겨졌지만, 그 컴패니언 머리말이 스스로 「`session-handoff.md` 가 SSOT 이며 모든 구속 조항을 보존한다」고 선언한다. 「번역·치환 금지」는 구속 조항이므로 SSOT 에 한 줄로 되살리고 세부는 컴패니언을 가리키게 했다 |
| 6 | `TestRegistrySyncGuard/local`·`/template` | 1~5의 결과로 해소 | 두 레지스트리 사본 바이트 동일(`cmp` 무출력), 101 엔트리·다이제스트 불변 | 가드 자체는 수정하지 않았다 |
| 7 | `TestCodexAuditMCPTool` | 대상 이동(테스트) | 도구 이름은 `moai-mcp-tools-catalogue.md` 의 family 표(213행)와 도구 카탈로그(103~105행)에 백틱으로 존재 — 두 사본 동일 | 설계 문서 R-37 이 family 표를 컴패니언으로 옮겼고, stub 에는 총수 문장과 포인터만 남았다(stub 의 `codex_role_audit` 출현 0회). 총수 검사는 stub 에 그대로 두고, 이름 검사만 컴패니언 두 사본으로 옮겼다 |
| 8 | `TestMCPToolCatalogueFiguresMatchRegistry` | 대상 이동(테스트) | 컴패니언 204행 `## Tool families (35 of the 39 tools; …)` | 위와 같은 이동이다. family 헤더 검사를 stub 에서 컴패니언으로 옮겼고, 검사 강도(35=전체−session_msg, 39=전체)는 그대로다 |
| 9 | rosterguard `TestRegisteredSitesMatchTheirDeclaredAxis` | 대상 이동(가드 등록부) | `CLAUDE.md`·템플릿 §4 `**Retained agents (13)**` — 두 사본 각 1회 | 다이어트가 §4 의 중복 문장(「consists of exactly **13 retained agents**」)을 지웠고, 같은 수치를 담은 제목형 문구가 남았다. `CountPattern` 을 그 제목형으로 옮겨 수치 검사를 유지했다 |
| 10 | rosterguard `TestNumeralBreadthSetEqualsTheDeclaredUnion`·`TestNumeralResidualArithmeticCloses` (CLAUDE.md 두 행) | 도달 불가 선언 | `rosterNounRe` 는 대소문자를 구분하며(`numeral.go` 97행), 제목형의 `Retained agents` 는 대문자라 명사 부류 밖이다 | 기존 선례(`model-policy-profile-matrix-size`)와 같은 방식으로 `NumeralUnreachable` 에 이유를 적었다. 수치 주장 자체는 9번의 `CountPattern` 이 계속 검사한다 |
| 11 | rosterguard 면제 행 `skill-routing-section-marker`(+미러) | 면제 삭제 | 다이어트 후 `skill-routing.md` 두 사본에 `§4` 인용 0회 | 오탐 억제용 면제였고, 억제 대상 문구가 사라졌다. 면제를 지우는 것은 검사를 약화시키지 않는다 — 문구가 다시 나타나면 가드가 보고한다 |

## Evidence

모든 명령은 이 트리(`.claude/worktrees/t1175-fix`)에서 이 실행에 돌렸다. 긴 출력은 `.moai/reports/t1175/fix/` 에 있다.

**헌법 검증 (이 트리에서 빌드한 바이너리)**

```
$ go build -o <scratchpad>/moai-fix ./cmd/moai && <scratchpad>/moai-fix constitution validate
constitution validate: OK — no drift or violations detected (97 of 101 entries checked)

  4 retired entry/entries skipped ([SUPERSEDED …] marker); re-check them with --strict
exit=0
```

수리 전: `.moai/reports/t1175/fix/validate-before.txt` — `constitution validate: FAILED — 9 error(s) found`.

**패키지 테스트** (각각 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && …` 한 호출)

```
$ go test -count=1 ./internal/constitution/...
ok  	github.com/modu-ai/moai-adk/internal/constitution	1.236s        (exit=0, FAIL 0건)

$ go test -count=1 ./internal/harness/...
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	26.980s   (외 15개 패키지 ok, exit=0, FAIL 0건)

$ go test -count=1 ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	100.179s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.397s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.256s  (exit=0, FAIL 0건)

$ go test -count=1 -run TestAlwaysLoadedTokenBudget -v ./internal/config/
token_budget_guard_test.go:70: always-loaded surface = 65648 tokens (budget 77600, headroom 11952, 16 entries)
--- PASS: TestAlwaysLoadedTokenBudget                                (exit=0)
```

**cli 전체 실행** (`go-test-cli-hook` 슬롯 보유 중 실행, 해제 확인)

1차 — 리드가 준 명령 그대로:

```
$ go test -count=1 ./internal/cli/...
panic: test timed out after 10m0s
	running tests:
		TestConfigOrphanedWorktree_PrimaryGateEnforced (0s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	601.297s        (exit=1, `--- FAIL` 0건)
```

단언 실패가 아니라 Go 기본 패키지 시한(10분) 초과다. 이 저장소는 SPEC-CLI-TEST-TIMEOUT-001 에 따라 전 패키지 실행에 `-timeout 60m` 을 쓴다(`Makefile` 104행). 원문: `fix/test-cli-timeout10m.txt`.

2차 — 같은 명령에 저장소 관례 시한만 더함 (시작 11:34:55Z, 종료 11:56:30Z, 부하 평균 약 14~20):

```
$ go test -count=1 -timeout 60m ./internal/cli/...
ok  	github.com/modu-ai/moai-adk/internal/cli	1291.049s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	1.082s
… (하위 패키지 16개 ok)
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	27.604s
exit=0, `--- FAIL` 0건, `ok` 18줄
```

원문: `fix/test-cli.txt`.

수리 전 rosterguard: `.moai/reports/t1175/fix/rosterguard-before.txt` — 3건 FAIL(`34 reachable Count paths but 32 discharged hits`, CLAUDE.md·skill-routing 4경로 breadth 누락, §4 `CountPattern` 0회 매칭).

**구속 조항 동결 해시 (AC-ALD2-002 파이프라인 그대로)**

| 시점 | sha256 | 줄 수 |
|---|---|---|
| 수리 전 (HEAD `7fe658815`) | `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` | 170 |
| 수리 후 | `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` | 170 |

수리 후 다중집합 원본: `.moai/reports/t1175/fix/frozen-after.txt`.

**토큰 예산**: 65,591 → 65,648 토큰(+57), 예산 77,600, 여유 12,009 → 11,952. 수리 전 출력 `.moai/reports/t1175/fix/token-budget-before.txt`, 후 `token-budget-after.txt`.

**lint**: `~/go/bin/golangci-lint run ./...` (v2.1.6) → `0 issues.`, exit=0 (`fix/lint.txt`).

**카탈로그 해시**: `go run ./internal/template/scripts/gen-catalog-hashes.go --all` → `catalog.yaml` 무변경(수정한 파일이 해시 대상이 아니다). 에이전트 `.md` 는 건드리지 않아 `make agents-emit` 은 필요 없었다.

## Baseline-attribution

- 비교 기준선은 병합 제1부모 `f01c7889a` 의 파일 내용(`git show f01c7889a:<path>`)과 수리 착수 시점 HEAD `7fe658815` 에서 이 실행에 잰 값이다.
- 동결 해시의 수리 전 값은 이 트리에서 이 실행에 잰 것이며, `acceptance.md` AC-ALD2-002 가 적은 기준선과 같은 값이다.
- 토큰 예산 수리 전 값(65,591)은 수정 전 이 트리에서 같은 명령으로 잰 것이다.
- cli 두 테스트의 수리 전 실패는 리드가 보고한 것이다. 이 실행에서는 그 원인(stub 의 `codex_role_audit` 0회, stub 에 `Tool families (N of the M tools` 헤더 부재)를 grep 으로 확인했고, 실패 자체를 재현하지는 않았다(아래 Gaps).

## Gaps

- `internal/cli` 두 테스트의 **수리 전 실패 출력**은 이 실행에서 재현하지 않았다. 근거는 리드의 CI 보고와 이 실행의 grep 결과다.
- 리드가 지정한 `go test -count=1 ./internal/cli/...` 는 기본 시한 10분에서 끝나지 못했다. 통과 판정은 `-timeout 60m` 을 더한 2차 실행에 근거한다. 10분 안에 끝나는지는 이 머신 부하(평균 14~20)에서 관측되지 않았다.
- `go test ./...` 전체는 규정대로 돌리지 않았다. 전 패키지 판정은 develop push 후 CI 몫이다.
- darwin 이외 플랫폼(linux·windows) 실행은 관측하지 않았다.
- 재지정한 테스트(7·8번)에 대해 변이 실험(컴패니언에서 이름을 지웠을 때 실패하는지)은 하지 않았다. 다만 두 검사 모두 매칭 실패 시 `t.Fatalf`/`t.Errorf` 로 끝나는 구조이며, 통과했다는 것은 정규식이 실제로 매칭됐다는 뜻이다.

## Residual-risk

- 3번(049)과 7·8번은 조항·표를 **지연 로드 컴패니언** 기준으로 검사한다. 스켑티컬 자세 절을 상시 로드에서 뺀 판단은 t1175 설계(R-01)의 것이며, 이 수리는 그 판단을 뒤집지 않았다. 감사 에이전트가 이 절을 로드하지 않은 채 동작할 가능성은 설계 단계의 잔여 위험으로 남는다.
- 9·10번으로 CLAUDE.md §4 는 수치 층(numeral layer)의 탐색 범위에서 빠졌다. §4 에 소문자 「N retained agents」 식의 새 문장이 들어오면 수치 층은 여전히 잡지만, 제목형 수치는 `CountPattern` 하나로만 검사된다.
- 5번은 `✂` 규칙을 SSOT 와 컴패니언 두 곳에 둔다. 한쪽만 고치는 편집이 생기면 두 문구가 갈라질 수 있다 — 헌법 검증은 SSOT 쪽만 본다.
- 이 수리 뒤에도 t1175 가 병합 과정에서 바꾼 다른 문구가 레지스트리 밖의 테스트에 걸릴 가능성은 CI 전체 실행 전까지 배제할 수 없다.
