# SPEC-CODEX-SKILL-LOADER-001 — 판정 기준

각 항목은 Given-When-Then 이며 이진 판정 가능하다. "관측됐다"는 명령과 그 출력이 증거 경로에 남아 있다는 뜻이다 — 요약은 증거가 아니다.

증거 경로: `.moai/reports/t452/`.

---

## §D 판정 목록

### AC-CSL-001 — 적재 뿌리 프로브가 실행되고 출력이 남는다

**Given** 격리된 `/tmp` 프로젝트와 격리 `CODEX_HOME` 이 준비되어 있고, **관측 신호가 프로브 실행 전에 선택·기록**되어 있으며(plan.md M0 의 신호 후보 중 하나), 뿌리마다 **서로 다른 표식 이름**이 심겨 있을 때,
**When** M0 의 5 개 후보 뿌리 각각에 대해 **한 번에 한 뿌리만 채운 채** 프로브를 실행하고, 표식을 하나도 심지 않은 음성 대조 실행을 1 회 추가로 실행하면,
**Then** 실행마다 다음 다섯이 한 기록에 함께 남아 있다 — (1) 실행 명령, (2) codex 출력 전문, (3) **그 명령의 exit code**, (4) **같은 실행 회차에서 관측한 `codex --version` 문자열**, (5) **이 트리의 SHA**. 그리고 뿌리마다 적재 여부 판정이 있다. 어느 한 뿌리라도 판정이 비어 있거나 다섯 중 하나라도 빠지면 FAIL.
**음성 대조가 신호 자체를 검증한다**: 표식 없는 실행에서 신호가 **나타나면** 그 신호는 적재를 판별하지 못하는 것이므로, 신호를 다시 고르고 모든 뿌리를 재실행한다. 이 대조 없이 기록된 양성은 판정이 아니다.
**판정 표면 제한**: `codex doctor` 출력은 이 AC 의 증거가 될 수 **없다** — plan-phase 프로브가 스킬 뿌리를 하나도 열거하지 않음을 확인했다(spec.md §A.5 P4). 증거는 실제 코덱스 세션의 출력이어야 한다.
**`strings` 프로브도 이 AC 의 증거가 될 수 없다** — 런타임 조립 경로를 볼 수 없으므로(§A.5), 문자열 부재는 미적재의 증거가 아니다.

### AC-CSL-002 — 분기가 기록된다

**Given** AC-CSL-001 의 관측이 끝났고 **양성 대조가 충족되었을 때 — 즉 다섯 뿌리 중 어느 하나에서든 신호가 발화했을 때**(R1 이 발화한 회차도 포함한다: R1 의 발화 자체가 계측기가 작동한다는 증거이며, 그것이 양성 대조가 하는 일의 전부다),
**When** `.agents/skills/<name>/SKILL.md` 의 적재 여부를 읽으면,
**Then** `progress.md` §E.2 에 분기 A 또는 분기 B 가 명시되어 있고, 그 판정의 근거가 된 출력 줄이 인용되어 있다.
**[HARD] 어느 뿌리에서도 신호가 발화하지 않은 회차는 `inconclusive` 로 기록하고 분기를 적지 않는다** — 계측기가 아무것도 재지 못한 상태이며, 그 침묵을 "적재되지 않았다"로 읽으면 분기 B(이 SPEC 이 더 유력하다고 본 결과)를 **깨진 계측기로 확증**하게 된다. plan.md M0 규칙 4 의 재선택(최대 3 회) 뒤에도 **어느 뿌리에서도 발화하지 않으면**(R2 만이 아니다 — 양성 대조가 어느 뿌리로도 충족되므로 소진 조건도 같은 기준이다) blocker report 로 정지하며, 이 AC 는 FAIL 이 아니라 `inconclusive` 로 기록된다.

### AC-CSL-003 — 분기 B 는 편집 없이 정지한다

**Given** AC-CSL-002 가 분기 B 를 기록했을 때,
**When** run-phase 를 종료하면,
**Then** `git diff "$BASELINE_SHA" -- internal/template/agentemit/agents-codex.yaml` 이 빈 출력이고(**기준 없는 `git diff` 는 여기서도 쓸 수 없다** — N1 과 같은 이유로, 커밋된 변경을 못 보고 미편집을 확증한다), blocker report 가 (a) 실제 적재 뿌리 (b) `.agents/skills` 미적재의 관측 근거 (c) 미러 무력화 결론과 그것이 만드는 의존 관계 — §A.4 방출 행만으로는 카드가 닫히지 않는다 (d) 대체 기구가 별개 카드인 이유 (e) **관측에 쓰인 `codex --version` 문자열** 다섯을 모두 담는다.
**(e) 가 분기 B 의 유일한 버전 고정점이다** — 이 경로에서는 AC-CSL-009 가 해당 없음이므로, 여기서 버전을 못박지 않으면 "물려받은 전제가 **이 버전에서** 거짓"이라는 이 분기의 결론 전체가 버전 귀속을 잃는다.
**적용 조건**: 분기 A 이면, **그리고 `inconclusive` 회차에서도**, 이 판정은 해당 없음으로 기록하고 그 사실 자체를 §E.3(분기 A) 또는 §E.2(`inconclusive`)에 남긴다(조용한 생략 금지). 두 경우 모두 이 AC 의 `Given`(AC-CSL-002 가 분기 B 를 기록했을 때)이 성립하지 않는다.

### AC-CSL-004 — 에이전트 `skills` 키의 세 성질이 각각 관측된다

**Given** 분기 A 가 확정되었을 때,
**When** 격리 프로젝트에서 존재·값 형태·오값 거동 프로브를 실행하면,
**Then** 세 관측 각각에 대해 명령과 출력이 남아 있고, 각 관측이 "확인됨 / 확인되지 않음" 중 하나로 판정되어 있다. 셋 중 하나라도 판정 없이 넘어가면 FAIL.

### AC-CSL-005 — 미확인 값 집합은 방출되지 않는다

**Given** AC-CSL-004 의 관측 결과가 있을 때,
**When** **이 SPEC 의 diff 가 새로 도입한 키 집합**을 `$BASELINE_SHA`(plan.md §C.0 에서 동결, `progress.md` §E.2 에 기록) 기준으로 구하면 — `git diff "$BASELINE_SHA" -- internal/template/templates/.codex/agents/moai/ | grep '^+[a-z0-9_]* *=' | sed 's/^+//;s/ *=.*//' | sort -u` —
**[HARD] 기준 없는 `git diff` 는 이 판정에 쓸 수 없다** — 커밋 뒤에는 빈 출력을 내고, 그 빈 출력이 아래 "빈 집합 → 해당 없음" 경로로 흘러 필드를 도입한 실행을 "도입 없음"으로 기록하게 만든다. 키 문자 집합도 `[a-z0-9_]` 여야 한다(숫자를 담은 키를 `[a-z_]` 는 놓친다).
**Then** 그 집합의 모든 원소가 AC-CSL-004 에서 "확인됨"으로 판정되어 있다.
**빈 집합은 통과가 아니다**: 집합이 비면(이 SPEC 이 새 필드를 하나도 도입하지 않았다는 뜻) 그 사실을 기록하고 **해당 없음**으로 판정한다 — 스윕한 것이 없는 초록은 아무것도 주장하지 않는다.
**[HARD] 스윕 범위는 이 SPEC 이 도입한 필드에 한정된다.** 방출된 TOML 이 이미 담고 있는 7 개 키(`args`, `command`, `description`, `developer_instructions`, `model_reasoning_effort`, `name`, `sandbox_mode`)는 **선행 SPEC 의 측정으로 확인된 필드**이며 이 AC 의 대상이 아니다 — 그 측정 기록은 `internal/template/agentemit/agents-codex.yaml` 의 `fields:` 절과 `classes:` 절 rationale(각 필드의 P-0N MEASURED 근거)에 있다. 전체 키 집합을 쓸면 AC-CSL-004 가 판정한 적 없는 선행 필드들 때문에 **도착 시점부터 붉고 어떤 올바른 작업으로도 초록이 되지 않는다**(REQ-CSL-005 는 이 SPEC 이 새로 방출하는 필드만 구속한다).

### AC-CSL-006 — 방출 규칙 전환 (분기 A + 키 확인 경로)

**Given** 분기 A 이고 AC-CSL-004 가 `skills` 키를 확인했을 때,
**When** `grep -A2 'class: skill-loader' internal/template/agentemit/agents-codex.yaml` 을 실행하면,
**Then** `disposition` 이 `deferred-m1` 이 아니며, **방출 규칙이 선택한 에이전트 각각**의 TOML 에 관측된 값 형태를 따르는 스킬 필드가 존재한다.
**개수는 리터럴이 아니라 도출한다**: 기대 개수는 방출 규칙이 선택하는 에이전트 수이고, 전체 개수는 `ls internal/template/templates/.codex/agents/moai/*.toml | wc -l` 로 그 회차에 센다. 두 수가 같아야 하는지(모든 에이전트에 균일 적용)는 **M0/M2 측정이 아직 정하지 않은 설계 사항**이므로 이 AC 가 미리 정하지 않는다 — 규칙이 선택한 집합과 실제 필드를 담은 집합이 **일치**하면 PASS 다.

### AC-CSL-007 — documented-drop 전환 (키 미확인 경로)

**Given** 분기 A 이지만 AC-CSL-004 가 사용 가능한 키를 확인하지 못했을 때,
**When** 같은 행을 읽으면,
**Then** `disposition: documented-drop` 이고 `rationale` 이 관측 명령·출력 요지·codex 버전 셋을 담는다.
**AC-CSL-006 과 AC-CSL-007 은 분기 A 안에서 상호 배타적이다** — **분기 A 회차에서** 정확히 하나가 PASS 로 판정되고 다른 하나는 해당 없음으로 기록되며, 분기 A 에서 둘 다 해당 없음이면 REQ-CSL-006·007 위반이므로 FAIL.
**[HARD] 분기 B 와 `inconclusive` 회차에서는 둘 다 해당 없음이며 위반이 아니다** — REQ-CSL-006 은 `Where 분기 A 가 확정되고`, REQ-CSL-007 은 `분기 A 이지만…` 을 전제로 하므로, 두 요구사항의 가드가 애초에 충족되지 않는다. 이 조항에서 범위를 떼면 분기 B(이 SPEC 이 더 유력하다고 본 결과)가 **판정할 것이 없던 한 쌍 때문에 강제 FAIL 로 끝난다.**

### AC-CSL-008 — 소스 축과 임베드 축이 갈리지 않는다

**Given** `agents-codex.yaml` 이 변경되었을 때,
**When** `make agents-emit` → `make build` → `make agents-emit-check` → `go test ./internal/template/agentemit/... -count=1` 을 순서대로 실행하면,
**Then** 네 명령 모두 exit 0 이고, 마지막 명령의 출력에 `ok` 가 나타난다. 각 명령의 출력이 인용된다.

### AC-CSL-009 — 측정 버전이 박힌다

**Given** 프로브가 사용한 codex-cli 버전이 기록되어 있을 때,
**When** `grep -A4 'class: skill-loader' internal/template/agentemit/agents-codex.yaml` 과 `grep 'codex_measured_version' internal/template/agentemit/agents-codex.yaml` 을 함께 실행하면,
**Then** `skill-loader` 행의 rationale 이 `codex --version` 이 이 회차에 출력한 버전을 문자열로 담는다.
**[HARD] 최상단 `codex_measured_version` 은 이 회차에 재측정한 필드만 덮을 때 갱신한다** — 그 필드는 "아래 필드 의미론을 잰 버전"을 뜻하므로(`agents-codex.yaml:13`), 이 SPEC 이 재지 않은 선행 7 개 필드(AC-CSL-005 가 재판정 대상에서 제외한 그 집합)까지 새 버전에서 확인된 것처럼 넓히면 매니페스트가 갖지 않은 커버리지를 주장하게 된다. 값을 유지했다면 그 사실과 이유를 증거 경로에 기록한다 — 조용한 미갱신과 구분되어야 한다.
**적용 조건**: 분기 B 와 `inconclusive` 회차에서는 해당 없음으로 기록한다 — REQ-CSL-003 이 매니페스트 편집을 금지하므로 이 판정이 요구하는 rationale 갱신 자체가 도달 불가능하고, 그 경로의 버전 고정점은 AC-CSL-003(e) 의 blocker report 다. `Given`/`When` 이 성립하는데도 해당 없음인 이유가 이것이며, 처분표만으로 이 자리를 덮지 않는다.

### AC-CSL-010 — 개발 저장소의 사용자 계층이 불변이다

**Given** M0 진입 전에 `$CODEX_HOME/config.toml` 의 해시와 `$CODEX_HOME/skills/` 목록을 떠 두었을 때,
**When** 모든 프로브가 끝난 뒤 같은 방식으로 다시 뜨면,
**Then** 두 측정이 동일하다. 프로브 실행에 쓰인 `CODEX_HOME` 은 이 값과 다른 격리 경로임이 프로브 명령 문면에서 확인된다.

### AC-CSL-011 — 사용자 계층 쓰기가 도입되지 않는다

**Given** 이 SPEC 의 변경 집합이 `$BASELINE_SHA`(plan.md §C.0) 기준으로 정해져 있을 때 — `git diff --name-only "$BASELINE_SHA" -- '*.go'` —
**When** 그 목록의 Go 소스에 대해 쓰기 계열 호출(`WriteFile`, `os.Create`, `MkdirAll`, `Remove`)과 `skills.config` 문자열을 함께 grep 하면,
**Then** 사용자 계층 경로를 대상으로 하는 매치가 0 이다. 셀렉터가 살아 있음을 보이는 대조군(실제로 매치되는 표현) 결과가 함께 기록된다 — 0 매치의 침묵만으로는 판정하지 않는다.

### AC-CSL-012 — 템플릿 중립성

**Given** `internal/template/templates/` 아래에 변경이 있을 때,
**When** 변경 파일 목록을 `$BASELINE_SHA`(plan.md §C.0) 기준으로 먼저 세고(`git diff --name-only "$BASELINE_SHA" -- internal/template/templates/ | wc -l`) **그 diff 가 더한 줄**(`git diff "$BASELINE_SHA" -- internal/template/templates/ | grep '^+' | grep -v '^+++'`)에 대해 `SPEC-`, 4 자리 연도 형태의 날짜, 40/7 자리 커밋 해시, `/Users/` 를 grep 하면,
**Then** 스윕한 파일 수가 **1 이상**이고 매치가 0 이다. 셀렉터가 살아 있음을 보이는 대조군(금지 토큰 하나를 실제로 담은 표현으로 같은 grep 을 돌린 결과)이 함께 기록된다.
**[HARD] 스윕 대상이 0 이면 해당 없음으로 판정한다** — 분기 B 는 이 트리 아래를 하나도 바꾸지 않으므로 여기서 PASS 를 기록하면 빈 집합의 초록을 판정으로 올리는 것이 된다(AC-CSL-011 의 대조군 조항이 같은 형태를 이미 막고 있다).

**[정정 기록 — grep 대상 축소]** 이 판정은 원래 **변경 파일 전체**를 grep 했다. 그 형태는 틀렸다. 이 회차가 `internal/template/templates/` 아래에서 건드린 파일은 `.codex/agents/moai/sync-auditor.toml` 하나인데, 그 파일 60 번째 줄의 `SPEC: {SPEC-ID}` 는 보고서 서식의 **자리표시자**이지 SPEC ID 가 아니며 이 작업 이전부터 있던 내용이다. 판정의 의도는 "이 작업이 템플릿 트리에 SPEC-ID 토큰을 **새로 들여놓지 않았다**" 인데, 파일 전체를 훑으면 건드리지도 않은 기존 내용 때문에 아무리 옳게 작업해도 빨간불이 켜진다 — 지배 규칙보다 엄격해서 생긴 틀린 이유의 적색이다. 그래서 grep **대상**만 diff 가 더한 줄로 좁혔다.

측정(baseline `c529b2e4aaf5148aee7e6c67649bf392837bbb06`):

- `git show c529b2e4a:internal/template/templates/.codex/agents/moai/sync-auditor.toml | grep -c 'SPEC: {SPEC-ID}'` → `1` (baseline 에 이미 있음)
- `git diff c529b2e4a -- internal/template/templates/.codex/agents/moai/sync-auditor.toml | grep '^+' | grep -c 'SPEC-'` → `0` (이 diff 는 그런 줄을 더하지 않았음)

프로젝트 자체 중립성 가드도 초록이었다 — `go test ./internal/template/ -count=1` 종료코드 0 (`TestTemplateNeutralityAudit`, `TestTemplateNoInternalContentLeak` 포함).

**바뀌지 않은 것**: 금지 토큰 집합(`SPEC-`, 4 자리 연도 형태의 날짜, 40/7 자리 커밋 해시, `/Users/`), "스윕한 파일 수가 **1 이상**" 이라는 문턱, 셀렉터가 살아 있음을 보이는 대조군 기록 의무, 위 [HARD] 해당 없음 조항. 좁힌 것은 grep 의 **대상**뿐이며, 그 밖을 좁히면 완화가 된다.

### AC-CSL-013 — `CODEX_HOME` 존중과 오류 접힘 금지

**Given** 이 SPEC 이 **커밋한** 프로브 스크립트·Go 코드가 있을 때,
**When** 그 대상 집합을 먼저 세고(커밋된 스크립트 파일 수 + 새로 추가된 `os.Stat` 사용처 수), 홈 경로 하드코딩(`/Users/`, `$HOME/.codex` 리터럴)을 grep 하고 각 `os.Stat` 사용처를 읽으면,
**Then** 두 대상 집합 중 **적어도 하나가 비어 있지 않고**, 하드코딩 매치가 0 이며, 각 `os.Stat` 사용처가 `IsNotExist` 와 그 밖의 오류를 구분해 다룬다. 하드코딩 grep 에는 대조군(실제로 매치되는 표현)이 함께 기록된다.
**[HARD] 두 대상 집합이 모두 비면 해당 없음으로 판정한다** — 분기 B 는 Go 코드를 더하지 않고, 프로브 스크립트의 커밋을 요구하는 REQ 도 없다. 두 절 모두 주어가 없는 상태에서의 PASS 는 빈 스윕의 초록이다.

---

## 완료 정의

- 위 13 개 판정이 각각 PASS / FAIL / 해당 없음 중 하나로 기록되어 있다(**AC-CSL-002 에 한해 `inconclusive` 를 추가로 쓸 수 있다**). 미기록은 FAIL 과 같다.
  - **네 번째 상태는 AC-CSL-002 에만 열려 있다** — 상태 집합을 먼저 넓게 열고 뒤에서 좁히면, 앞의 허용문이 뒤의 제한을 근거 없이 이긴다. 다른 12 개 판정에 `inconclusive` 를 기록하는 것은 이 목록의 위반이다.
  - `inconclusive` 는 계측기가 아무것도 재지 못한 회차(어느 뿌리에서도 신호 미발화)를 가리키며, 분기가 기록되지 않았으므로 아래 두 경로 중 어느 것도 적용되지 않는다. 이 상태로 끝난 회차는 blocker report 로 닫힌다(plan.md M0 규칙 4).
- 분기 A 경로: AC-CSL-003 = 해당 없음, AC-CSL-006 과 AC-CSL-007 중 하나가 PASS.
- 분기 B 경로: AC-CSL-003 = PASS, AC-CSL-004~009 = 해당 없음, **AC-CSL-010·011 = PASS**(프로브 실행만으로도 주어가 있다), **AC-CSL-012·013 = 해당 없음**(분기 B 는 템플릿 트리를 바꾸지 않고 Go 코드를 더하지 않으므로 스윕 대상이 비어 있다).
  - **012·013 을 이 경로에서 PASS 로 적는 것은 금지한다** — 빈 스윕의 초록을 판정으로 승격시키는 형태이며, 각 AC 본문의 [HARD] 조항이 같은 결론을 독립적으로 강제한다. 두 자리가 어긋나면 AC 본문이 정본이다.
- **`inconclusive` 경로: AC-CSL-002 = `inconclusive`, AC-CSL-001 = PASS(관측 자체는 다섯 요소를 갖춰 수행됐다), **AC-CSL-010 = PASS**(분기 B 행과 같은 근거 — 프로브 실행만으로도 주어가 있다), 나머지 10 개(AC-CSL-003~009 · 011 · 012 · 013) = 해당 없음.** 그 사실과 blocker report 경로를 `progress.md` §E.2 에 기록한다.
  - **[HARD] AC-CSL-010 을 이 경로에서 해당 없음으로 적는 것은 금지한다.** 그 판정의 `Given`(M0 진입 전 해시)과 `When`(모든 프로브가 끝난 뒤 재측정)은 `inconclusive` 회차에서 **성립한다** — 프로브가 실제로 돌았기 때문이며, 같은 이유로 이 행이 AC-CSL-001 을 PASS 로 적는다. 판정의 `Given`/`When` 이 성립하는데 처분표가 해당 없음을 지시하면 두 자리가 같은 조건에서 반대를 말한다.
  - **이 경로가 프로브를 가장 많이 돌리는 경로다** — 재선택 상한 3 회 아래에서 다섯 뿌리 스윕이 최대 네 벌, 각각에 음성 대조가 붙는다. AC-CSL-010 은 **리드 HARD 제약 (a)**(이 저장소의 `~/.codex` 를 쓰지 않는다)를 판정하는 자리이므로, 여기서 해당 없음으로 접으면 **격리가 가장 많이 시험된 경로에서 그 제약을 아무도 확인하지 않는다.**
  - **AC-CSL-011 은 해당 없음이 맞다** — 그 판정의 스윕 대상은 `$BASELINE_SHA` 이후 변경된 Go 소스인데, 이 경로는 Go 코드를 더하지 않아 대상 집합이 빈다. 010 과 011 이 갈리는 이유는 하나가 **실행**(프로브가 돌았다)을 보고 다른 하나가 **변경 집합**(코드가 늘었다)을 보기 때문이다.
  - **이 줄이 없으면 `inconclusive` 회차의 10 개 판정 중 6 개에 아무 처분도 허가되지 않는다** — AC-CSL-004 · 005 · 008 · 011 · 012 · 013 은 자기 본문이 `inconclusive` 를 지목하지 않아 분기 B 행에만 기대는데, 분기가 기록되지 않았으므로 그 행이 발효하지 않는다. 그 상태에서 "미기록 = FAIL" 만 남으면 아무것도 실패하지 않은 회차가 FAIL 로 기록된다. 나머지 4 개는 이 줄이 없어도 허가된다 — AC-CSL-003 과 AC-CSL-009 는 각자의 적용 조건 절이, AC-CSL-006 과 AC-CSL-007 은 AC-CSL-007 본문의 [HARD] 조항이 `inconclusive` 회차를 이름으로 지목한다. 네 자리 모두 이 행과 같은 처분(해당 없음)을 말하므로 충돌은 없고, 이 줄이 유일한 근거인 범위만 6 개다.
    - **`Given`/`When` 의 성립 여부로 이 10 개를 묶지 않는다** — 그 서술은 **AC-CSL-009 에 대해 거짓**이다(그 `Given` 은 버전이 기록되어 있으면 성립하고, 이 경로에서도 성립한다). 009 가 해당 없음인 이유는 `Given` 이 안 서기 때문이 아니라 **REQ-CSL-003 이 rationale 갱신을 도달 불가능하게 만들고 버전 고정점이 AC-CSL-003(e) 로 옮겨져 있기 때문**이며, 그 근거는 009 본문의 적용 조건이 담는다.
    - **001 과 010 은 이 행의 허가가 아니라 자기 판정으로 PASS 다** — 두 판정은 `Given`/`When` 이 성립하고 `Then` 도 만족되므로, 처분표가 없어도 PASS 다. 표는 그 사실을 기록할 뿐 만들지 않는다.
  - **이 경로의 산출물은 blocker report 이며, 그 내용은 plan.md M0 규칙 4 가 열거한다**(시도한 신호 3 개와 각각의 관측, R2 미발화, 이 코덱스 버전에서 세션 출력으로 적재를 관측할 방법이 확인되지 않았다는 결론). AC-CSL-003 의 (a)-(e) 는 **분기 B** 의 blocker report 를 판정하므로 이 경로에는 적용되지 않는다 — 두 blocker report 는 다른 문서다.
- 모든 판정의 근거 명령과 출력이 `.moai/reports/t452/` 에 남아 있고, 그 경로가 완료 보고에서 인용된다.
