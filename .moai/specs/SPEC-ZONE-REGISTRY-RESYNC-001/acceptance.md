# Acceptance — SPEC-ZONE-REGISTRY-RESYNC-001

측정 기준 트리: `.claude/worktrees/t232` @ `294b4b6ab`. 아래 모든 RED 값은 이 트리에서 실제로 명령을 돌려 관측한 값이다(근거: `.moai/reports/t232/`). 각 AC는 **관측 가능한 출력**을 내는 명령으로 판정하며, RED 값이 명시돼 있으므로 "아무것도 관측하지 않는 체크"가 조용히 통과할 수 없다.

용어: **독립 체크** = 수리 대상(레지스트리 데이터)도 아니고 그 데이터를 판정하는 MoAI 코드(`internal/constitution`)도 아닌 도구로 하는 검증. 여기서는 `grep -F`(리터럴 고정 문자열, 정규화 없음)를 쓴다. 측정 결과 `grep -F` 통과 집합은 검증기 통과 집합의 **진부분집합**이다(측정: validator_ok 33 ⊃ grepF_ok 25, 역방향 0건) — 즉 독립 체크는 검증기보다 **엄격**하며, 이를 통과하면 검증기 통과는 함의된다.

---

## §D AC 매트릭스

| AC | 요구사항 | RED (현재 트리) | GREEN (목표) |
|---|---|---|---|
| AC-ZRR-001 | REQ-ZRR-015 | validate exit 1 / 67 errors | exit 0 / 0 errors |
| AC-ZRR-002 | REQ-ZRR-001, 003, 005 | 로컬 **1회 적중 24** / 0회 76 / 2회 이상 1 / 자기참조 0 | 101 / 0 / 0 / 0 |
| AC-ZRR-003 | REQ-ZRR-001, 003, 013 | 템플릿 **1회 적중 24** / 2회 이상 1 (부가 조건 동일) | 101 / 0 |
| AC-ZRR-004 | REQ-ZRR-002, 003, 009 | anchor 해석 84/101 (실패 17, §2.2 slug 규칙 기준) | 101/101 |
| AC-ZRR-005 | REQ-ZRR-004, 006 | 매처 3함수 현행 | 바이트 불변 + 기존 테스트 통과 |
| AC-ZRR-006 | REQ-ZRR-005 | 101 엔트리 / ID·zone·zone_class·canary_gate 집합 X | 동일 집합 (diff 0) |
| AC-ZRR-007 | REQ-ZRR-007, 008, 011 | 가드 부재 (`constitution validate` 배선 0건) | 변이에서 붉음(로컬 + CI job 결론) + 초록, 평가 엔트리 수 101×2 |
| AC-ZRR-008 | REQ-ZRR-011 | 유일한 constitution job 이 continue-on-error | 차단 job **그리고** 억제 없는 스텝 |
| AC-ZRR-009 | REQ-ZRR-008 | 템플릿 미러 검증 0건 | 템플릿 변이도 검출 |
| AC-ZRR-010 | REQ-ZRR-010 | SKIP=1 이면 Skipped/OK 반환 | SKIP=1 이어도 실패 |
| AC-ZRR-011 | REQ-ZRR-013 | 미러 바이트 동일 (유지 대상) | 수리 후에도 동일 + embed 클린 |
| AC-ZRR-012 | REQ-ZRR-014 | 템플릿 레지스트리 SPEC-ID 0 / 날짜 0 / SHA 0 | 0 유지 (변이 주입 시 CI 실패) |
| AC-ZRR-013 | REQ-ZRR-015 | doctor Fail 1 | Fail 0 |
| AC-ZRR-014 | REQ-ZRR-012 | heading slug 헬퍼 부재 (재사용 가능 0건) | 규칙 6단계가 코드+주석에 선언 |

---

## §D.1 AC 상세

### AC-ZRR-001 — 갓 만든 프로젝트에서 validate 가 통과한다

- **Given** 이 변경으로 빌드한 `bin/moai` 로 `moai init --non-interactive` 한 빈 프로젝트가 있고,
- **When** 그 프로젝트 루트에서 `moai constitution validate` 를 실행하면,
- **Then** 종료 코드가 `0` 이고 출력에 `[DRIFT]` 가 0회 나타난다.

판정 명령: `moai constitution validate; echo exit=$?` 및 `moai constitution validate 2>&1 | grep -c DRIFT`
**RED**: `exit=1`, DRIFT `67`. (전문: `.moai/reports/t232/validate-repro.txt`)

### AC-ZRR-002 — 로컬 트리에서 101/101 clause 가 리터럴로 발견된다 (독립 체크)

- **Given** 수리된 로컬 레지스트리와 로컬 규칙 트리가 있고,
- **When** 각 엔트리의 `clause:` 값을 그 엔트리의 `file:` 에 대고 `grep -F -q --` 로 찾으면,
- **Then** 101건 전부가 적중한다(적중 실패 0건).
- **And** 어떤 엔트리의 `clause:` 도 빈 문자열이 아니며, **각 clause 는 자기 `file:` 안에서 정확히 1회 적중한다**(측정: 현재 적중 25건 중 24건이 1회 적중 — 진짜 verbatim 인용은 유일하다). 0회는 미적중이고, 2회 이상은 clause 가 지나치게 짧다는 신호이므로 둘 다 실패로 센다.
- **And** 어떤 엔트리의 `file:` 도 **레지스트리 파일 자신**(`.claude/rules/moai/core/zone-registry.md`)이 아니다 — 자기참조는 clause 를 정의상 적중시키고(측정: `grep -F -c -- '16-language neutrality' <registry>` → `1`) 레지스트리의 heading 50개가 anchor 까지 해석시켜, 수리를 한 줄도 하지 않고 이 AC 를 통과시킨다.
- **And** `file:` 값이 바뀐 엔트리의 목록(구 → 신)을 `progress.md` §E.2 에 인용하고, sync-phase 리뷰가 각 이동의 타당성을 판정한다.

판정: 레지스트리를 파싱해 엔트리별 `grep -F -c` 를 돌리고 **적중 횟수 자체**(boolean 아님)를 집계하는 체크 — 1회 적중 수 / 0회 수 / 2회 이상 수 / 자기참조 `file:` 수를 출력한다. 구현은 run-phase 소관이며 `internal/constitution` 의 매칭 코드를 호출하지 않는다.

**RED (`294b4b6ab`)**: 1회 적중 `24`, 0회 `76`, 2회 이상 `1`(`CONST-V3R2-002` = `TRUST 5`, 3회), 자기참조 `file:` `0`. **GREEN 은 1회 적중 101 / 0회 0 / 2회 이상 0 / 자기참조 0.**

boolean 이 아니라 횟수를 세는 이유: "빈 문자열 아님"만 요구하면 공백 한 칸이나 아주 짧은 흔한 토큰이 한 글자 차이로 금지를 비껴가면서 같은 거짓 GREEN 에 도달한다. 유일 적중은 그 우회로를 통째로 닫는다 — 진짜 verbatim 인용은 자기 파일 안에서 유일하기 때문이다(clause 길이 중앙값 93자; 20자 미만은 3건뿐이고 그중 하나가 위 3회 적중 엔트리다).

> 이 AC는 검증기보다 엄격하다: 정규화가 없으므로 clause 는 **한 줄 안에 연속으로 존재하는 구간**이어야 한다. 현재 검증기는 통과하지만 이 체크는 실패하는 엔트리가 8건 있다(`CONST-V3R5-004/005/006/007/008/010/011/013`) — 이들도 GREEN 대상이다.

### AC-ZRR-003 — 템플릿 트리에서도 101/101 이 리터럴로 발견된다

- **Given** 수리된 템플릿 레지스트리(`internal/template/templates/.claude/rules/moai/core/zone-registry.md`)가 있고,
- **When** 각 엔트리의 `clause:` 를 `internal/template/templates/<file>` 에 대고 `grep -F -q --` 로 찾으면,
- **Then** 101건 전부가 적중한다.
- **And** 각 clause 가 템플릿 트리의 자기 `file:` 안에서 **정확히 1회 적중한다**(0회 = 미적중, 2회 이상 = 지나치게 짧음, 둘 다 실패).
- **And** AC-ZRR-002 의 나머지 부가 조건(자기참조 `file:` 금지 / `file:` 변경 목록 인용)이 템플릿 미러에도 동일하게 적용된다 — 두 미러는 바이트 동일하므로 이 mutant 들은 양쪽을 한 번에 통과시킨다.

**RED**: 1회 적중 `24 of 101`(2회 이상 1건 = `CONST-V3R2-002`). AC-ZRR-002 와 별개 판정이다 — 인용 대상 17개 파일 중 **2개 파일**이 두 트리에서 다르고 그 2개가 3개 엔트리를 물고 있으므로(`spec.md` §5), 로컬만 맞춘 수리는 여기서 걸린다.

### AC-ZRR-004 — anchor 101/101 이 heading 으로 해석된다

**측정 규칙 (이 AC의 수치는 이 규칙 아래에서만 의미가 있다)** — `spec.md` §2.2 의 6단계 slug 규칙: ① 코드 펜스 안 행 제외 → ② `#` 접두 제거 + trim → ③ 백틱 제거 → ④ 소문자화 → ⑤ `[a-z0-9]`·공백·`-` 이외 제거 → ⑥ 연속 공백 → `-`, 앞에 `#`.

- **Given** 수리된 레지스트리와 위 slug 규칙이 있고,
- **When** 각 엔트리의 `anchor:` 를 그 엔트리의 `file:` 안의 heading slug 집합에 대고 해석하면,
- **Then** 101건 전부가 해석된다(미해석 0건).
- **And** clause 가 이미 통과하던 9건(`CONST-V3R5-004/005/006/007/008/010/011/012/013`)도 해석된다.

**RED (`294b4b6ab`, 위 slug 규칙 기준)**: 해석 실패 `17` / 101, 그중 clause 통과 `9`건. 다른 slug 규칙을 쓰면 이 수치는 달라진다 — 그래서 규칙을 AC 본문에 적는다(AC-ZRR-014 가 그 규칙의 코드 고정을 요구).

이 9건을 이름으로 못박는 이유: clause 만 고치는 수리에서 이들은 `validate` 초록·doctor 초록·사람 눈에 "고쳐진 것"으로 보이면서 그대로 남는다.

### AC-ZRR-005 — 매처가 불변이다

- **Given** base 커밋 `294b4b6ab` 의 `internal/constitution/validator.go` 가 있고,
- **When** 변경 후 `git diff 294b4b6ab..HEAD -- internal/constitution/validator.go` 를 보면,
- **Then** `normalizeWhitespace`, `stripCodeFences`, 그리고 `Validate` 의 DRIFT 판정 블록(clause 부분문자열 검사)에 **변경 라인이 0** 이다.
- **And** `go test ./internal/constitution/...` 이 통과한다.
- **And** 변경 diff 전체에 clause 길이 임계, 파일 제외 목록, 토큰 중첩/fuzzy 매칭, 새 환경변수 우회가 **도입되지 않는다**.

판정: 위 diff 라인 수 + `go test ./internal/constitution/... 2>&1 | tail -1`
**RED**: diff 0 라인(변경 전이므로 자명), `ok github.com/modu-ai/moai-adk/internal/constitution` — 기존 테스트는 **지금 통과한다**. 즉 이 AC의 GREEN 조건은 "새로 통과시키기"가 아니라 "깨뜨리지 않기"다.

### AC-ZRR-006 — 엔트리 집합이 보존된다

- **Given** base 트리에서 뽑은 `moai constitution list --format json` 결과가 있고,
- **When** 수리 후 같은 명령의 결과와 비교하면,
- **Then** 엔트리 수가 `101` → `101` 이고, `id` 집합이 완전히 동일하며, 엔트리별 `zone` / `zone_class` / `canary_gate` 값이 전부 동일하다.
- **And** 달라진 필드는 `clause` / `anchor` / `file` 뿐이다.

판정: 두 JSON 을 엔트리 단위로 비교(`clause`/`anchor`/`file` 제외 필드에 대해 diff 0).
**RED**: base 측 `entries: 101`. 원시 `grep -c canary_gate` 는 `104` 를 반환한다(3건은 산문/정책 절에 있는 언급) — **따라서 이 AC는 grep 카운트가 아니라 파싱된 엔트리 단위 비교로 판정한다.** grep 카운트로 세는 구현은 이 AC를 만족하지 못한다.

### AC-ZRR-007 — 가드가 **붉어지는 것이 관측됐다** (변이 주입)

이 AC는 "가드가 통과한다"가 아니다. **알려진 불량 입력에서 붉어지는 것을 보았고, 수리된 트리에서 초록인 것도 보았다** — 두 관측이 함께여야 충족된다. 통과만 본 가드는 자기가 무엇을 막는지 증명하지 못한다.

- **Given** 수리와 가드가 착지한 트리에서 가드가 통과하는 것을 확인했고,
- **And** 그 통과 출력이 **가드가 평가한 엔트리 수를 보고하며 그 값이 `101` 이다 — 두 미러 각각에 대해**(이 단언이 없으면 가드 쪽 부분 순회·조기 반환·제외 목록이 전부 살아남는다),
- **When** **명시된 엔트리 `CONST-V3R2-004` 1건과 무작위로 고른 1건**의 `clause:` 값에 문자 하나를 넣어 일부러 깨뜨린 뒤 가드 명령을 실행하면(대상을 고정하는 이유: "임의의 한 엔트리"는 실행마다 다른 것을 골라도 둘 다 충족으로 읽혀 재현 가능한 판정이 아니다),
- **Then** 가드가 **0이 아닌 종료 코드**로 실패하고, 실패 메시지가 그 엔트리의 ID를 지목한다.
- **And** 그 실패 출력(종료 코드 + 실패 엔트리 ID 를 포함한 실제 출력)이 `progress.md` §E.2 에 **그대로 인용**된다.
- **And** anchor 값으로도 같은 변이를 1회 수행해 같은 결과를 관측한다.
- **And** 변이를 담은 커밋을 푸시해 **해당 CI job 이 실제로 `fail` 로 결론나는 것**을 `gh pr checks` 출력으로 관측하고 인용한다 — 로컬 종료 코드만으로는 `|| true` 로 감싼 스텝을 구분하지 못한다.
- **And** 변이를 전부 되돌린 뒤 가드가 다시 통과하는 것을 확인한다.

**RED**: 현재 트리에는 가드가 없다 — `grep -c "constitution validate" Makefile .github/workflows/*.yml` 이 `0` 을 반환하고(전체 트리 grep 기준, 템플릿 포함), 어떤 Go 테스트도 `Validate` 를 프로젝트 레지스트리에 대고 호출하지 않는다. 지금 clause 를 깨뜨려도 CI 는 초록이다.

**미충족 판정**: 실패 출력 인용이 없으면 이 AC는 미충족이다. "확인했다"는 서술은 관측의 대체물이 아니다.

### AC-ZRR-008 — 가드가 차단 경로에 있다 (권고 아님)

- **Given** 가드가 배선된 CI 설정이 있고,
- **When** 가드를 실행하는 job **과 스텝**의 정의를 읽으면,
- **Then** 그 job 이 `continue-on-error: true` 가 **아니고**, 가드가 차단 경로에서 실행된다(기존 테스트 job `go test ./...` 이 그 경로다).
- **And** 가드를 실행하는 **스텝**이 `|| true`, 스텝 단위 `continue-on-error`, 그 밖의 어떤 종료 코드 억제로도 감싸이지 않았다 — job 단위 조건만 걸면 `go test ./... || true` 를 차단 job 안에 넣은 구현이 이 AC 를 문자 그대로 만족한다.

판정: 가드를 실행하는 job 이름 + 그 job 의 `continue-on-error` 값 + 그 스텝의 `run:` 본문과 스텝 단위 `continue-on-error` 를 함께 출력.
**RED**: 현존하는 유일한 constitution job(`.github/workflows/ci.yml:445-451`)은 `continue-on-error: true` 다. **이 job 에 `validate` 를 한 줄 추가하는 구현은 이 AC를 만족하지 못한다.** 억제 래핑 여부는 AC-ZRR-007 의 CI job 결론 관측이 함께 판정한다.

### AC-ZRR-009 — 가드가 템플릿 미러도 본다

- **Given** 가드가 착지한 트리가 있고,
- **When** **템플릿 쪽** 레지스트리(`internal/template/templates/...`)만 깨뜨리고 로컬은 그대로 둔 채 가드를 실행하면,
- **Then** 가드가 실패하고 실패가 템플릿 미러를 지목한다.

**RED**: 현재 템플릿 미러를 검증하는 코드는 0건이다. 로컬만 보는 가드는 이 AC에서 걸린다.

### AC-ZRR-010 — 가드가 우회로 조용히 통과하지 않는다

- **Given** 깨진 엔트리가 하나 있는 트리와 착지한 가드가 있고,
- **When** `MOAI_CONSTITUTION_SKIP_VALIDATE=1` 을 환경에 둔 채 가드를 실행하면,
- **Then** 가드가 여전히 실패한다(검증이 건너뛰어졌다는 사실 자체를 실패로 취급한다).

**RED**: `Validate` 는 이 환경변수가 `1` 이면 `ValidationResult{Status: Skipped, Skipped: true}` 를 반환하고 에러를 내지 않는다(`internal/constitution/validator.go` `Validate` 도입부). 그 반환을 그대로 성공으로 읽는 가드는 CI 환경변수 한 줄로 무력화된다.

### AC-ZRR-011 — 미러 동일성과 임베드가 유지된다

- **Given** 수리가 끝난 트리에서,
- **When** `diff -q` 로 로컬/템플릿 레지스트리를 비교하고 `make build` 후 `git status --porcelain` 을 보면,
- **Then** 두 파일이 바이트 동일하고, `make build` 가 성공하며, 빌드가 만들어낸 미커밋 변경이 없다.

**RED**: 현재도 바이트 동일하다(측정 확인). 이는 **유지 AC**이며, 판정력은 "수리 후에도 동일한가"에서 나온다 — 한쪽만 고치는 구현이 여기서 걸린다.

### AC-ZRR-012 — 템플릿 중립성이 유지된다

- **Given** 수리된 템플릿 레지스트리가 있고,
- **When** SPEC-ID 패턴 / ISO 날짜 / 40자리 커밋 SHA 를 grep 하면,
- **Then** 각각 0건이다.

판정 명령:
`grep -cE 'SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}' <template-registry>` / `grep -cE '20[0-9]{2}-[0-9]{2}-[0-9]{2}' …` / `grep -cE '\b[0-9a-f]{40}\b' …`
**RED**: `0` / `0` / `0`. 셋 다 이미 0이므로 이 AC는 **비회귀 AC**다 — 채택 근거는 변이로 준다: 템플릿 레지스트리에 SPEC ID 를 한 개 넣으면 CI 중립성 가드(`.github/workflows/template-neutrality-check.yaml`)가 실패해야 한다.
주의: 느슨한 `grep -c 'SPEC-'` 는 현재 `1` 을 반환한다(clause 안의 일반 토큰 `SPEC-ID`). 그 형태로 쓴 체크는 오탐이므로 위 엄격 패턴을 쓴다.

### AC-ZRR-013 — doctor 가 깨끗하다

- **Given** 이 변경으로 빌드한 바이너리로 초기화한 새 프로젝트가 있고,
- **When** `moai doctor` 를 실행하면,
- **Then** `Constitution Registry` 항목이 `fail` 이 아니고 요약의 `Fail` 카운트가 `0` 이다.

**RED**: `fail Constitution Registry  registry loads (101 entries) but validate found 67 error(s)` / `Pass 22  Warn 2  Fail 1`.

### AC-ZRR-014 — slug 규칙이 코드에 선언돼 있다

- **Given** 착지한 가드 코드가 있고,
- **When** anchor 해석부를 읽으면,
- **Then** `spec.md` §2.2 의 6단계 규칙이 코드로 그대로 구현돼 있고, "이 규칙 아래에서 착지 시점 17건이 실패했다"는 취지의 주석이 그 옆에 있다.
- **And** 그 해석기로 측정한 결과가 AC-ZRR-004 와 같은 수치를 낸다.

**RED**: 리포에 재사용 가능한 heading slug 헬퍼가 **없다**. `grep -rn "func.*[Ss]lug" --include="*.go" internal/` 이 반환하는 5개 — `i18nSlug`(`internal/web/render_helpers.go:19`), `memoryProjectSlug`(`internal/cli/memory.go:44`), `memorySlug`(`internal/cli/preference/cmd.go:159`), `slugify`(`internal/cli/preference/filestore.go:505`), `projectSlug`(`internal/hook/session_end.go:234`) — **전부 경로 또는 i18n 키를 slug 화하며, heading 을 다루는 것은 0개**다. 즉 규칙은 새로 선언될 수밖에 없고, 선언 없이 구현하면 그 수치가 무엇을 뜻하는지 아무도 재현할 수 없다.

---

## §D.2 심각도

| 등급 | AC |
|---|---|
| BLOCKING (하나라도 실패 시 close 불가) | 001, 002, 003, 004, 005, 006, 007, 010, 014 |
| SHOULD-FIX | 008, 009, 011, 013 |
| MINOR (비회귀) | 012 |

## §D.3 추적성

| REQ | AC |
|---|---|
| REQ-ZRR-001 | 002, 003 |
| REQ-ZRR-002 | 004 |
| REQ-ZRR-003 | 002, 003, 004 |
| REQ-ZRR-004 | 005 |
| REQ-ZRR-005 | 002, 006 |
| REQ-ZRR-006 | 005, 010 |
| REQ-ZRR-007 | 007 |
| REQ-ZRR-008 | 007, 009 |
| REQ-ZRR-009 | 004, 007 |
| REQ-ZRR-010 | 010 |
| REQ-ZRR-011 | 007, 008 |
| REQ-ZRR-013 | 011 |
| REQ-ZRR-014 | 012 |
| REQ-ZRR-015 | 001, 013 |
| REQ-ZRR-012 | 014 |

## §D.4 간접 검증 항목

- **"교리가 이사한 곳을 제대로 짚었는가"** 는 기계적으로 판정되지 않는다. AC-ZRR-002/003/004 는 인용이 verbatim 이고 anchor 가 해석된다는 것만 증명하지, **의미상 맞는 문장**을 골랐다는 것은 증명하지 않는다. 이 축은 sync-phase 리뷰(사람 판정) 소관이며, 이 SPEC 은 그 사실을 덮지 않고 남긴다.
- **재발 억제 효과**(가드가 미래의 깨짐을 실제로 잡는가)는 AC-ZRR-007/009 의 변이 주입으로 대리 검증한다. 실제 재발률은 관측 불가.

## §D.5 Definition of Done

- BLOCKING AC 9건 전부 GREEN, 각 판정 명령의 **실제 출력**이 `progress.md` §E.2 에 인용됨
- 가드의 **실패 출력**(변이 주입 시)이 §E.2 에 인용됨 — 통과 출력만으로는 AC-ZRR-007 미충족
- SHOULD-FIX 4건 GREEN 또는 명시적 미해소 사유 기록
- 로컬/템플릿 미러 바이트 동일 + `make build` 클린
- `go test ./internal/constitution/...` 및 `go test ./internal/template/...` 통과
- `spec.md` §5 의 3건 이중 트리 제약이 해소됐거나, 해소 불가 사유가 blocker 로 기록됨
