# SPEC-MODEL-CATALOG-SSOT-002 plan.md — 구현 계획 (Tier M)

작성: 2026-09-20 (manager-spec, 카드 t1030). 우선순위 라벨 사용, 시간 추정 없음 (AGENTS.md §6).

## §A Context

- 카드 t1030. 워크트리 `.claude/worktrees/t1030`, 브랜치 `WT-model-matrix-rewrite`, 기준 HEAD `9bd329c17`(= 로컬 `develop`).
- 전임: `SPEC-MODEL-CATALOG-SSOT-001`(카드 t842, `status: draft`, 미푸시 브랜치 `WT-model-catalog-ssot`). 본 SPEC이 축소 재작성한다 — 근거는 `spec.md` §A.2.
- 선행 감사 둘: 카드 t842 판정서, 카드 t1024 판정서(`.claude/worktrees/t1024/.moai/reports/t1024/verdict.md`). 둘 다 「현행 develop 기준 축소 재작성」을 관측이 지지하는 처분으로 지목했고, 운영자가 2026-09-20 그 갈래를 선택했다.
- 카드 완료 후 develop으로 병합 (`CLAUDE.local.md` §4.1).

## §B Known Issues

- **전임 SPEC의 AC 9건이 공허하다** (완전 5 + 부분 4, 카드 t1024 측정). 인계 수치 「6/15」는 과소계상이었다. 본 SPEC은 그 AC를 하나도 상속하지 않고 새로 썼다 — 승계 경로 자체를 끊는 것이 의도다.
- **단가는 전부 신규 데이터다.** 트리 안의 유일한 선행 근거는 주석 한 줄이다 — `internal/template/model_policy.go:47-48` "Opus 5 is priced identically to its predecessor Opus 4.8 ($5/$25 per MTok)". 나머지 단가는 차트 유래 근사값이며 단위가 미확인이다. **처리 결정(확정)**: REQ-MCS2-009가 `provenance: approximate` + `unit: unconfirmed` 표기를 의무화하고, 이 카드에는 단가를 소비하는 리더가 없으므로 값의 정확성이 어떤 판정에도 영향을 주지 않는다. 이 결정은 run-phase 착수를 막지 않는다.
- **Tier 근거**: 측정된 범위는 6파일 / 비주석 30행이지만 작업량은 그보다 크다 — 새 설정 섹션(타입·로더·Defaults·Validate·감사 레지스트리 2곳 등록) + 4개 패키지에 걸친 리더 2개 수렴 + 행위 보존 characterization + 전수 스캔 가드. 단일 표면을 넘으므로 Tier S가 아니고, 스키마가 전임 설계의 축소판이라 별도 설계 탐색(`design.md`)이 없어도 결정이 서므로 Tier L도 아니다. **Tier M** — `spec.md` / `plan.md` / `acceptance.md` / `progress.md` 4종.
- **transitive 소비자 미조사**: `ModelAliasTable`의 fan_in ≥ 3 소비자 3곳(`launcher.go` `expandModelString`, `profile_setup.go` `normalizeModel`, `settings/schema.go` `modelOptions`)에 각각 기존 테스트가 있는지는 재지 않았다 — M2 Pre-flight의 첫 항목이다.

## §C Pre-flight

M1 착수 전에 실행하고 출력을 인용한다.

1. 베이스라인 녹색 확인: `go test ./internal/config/... ./internal/template/... ./internal/statusline/... ./internal/profile/...`
2. `ModelAliasTable` 소비자 3곳의 기존 테스트 존재 확인 — 없으면 M2 RED에서 characterization으로 신규 작성
3. `glmContextWindows` 기존 테스트 확인: `internal/statusline/memory_test.go`(측정 시점 7적중)
4. 감사 테스트 목록 확인: `go test ./internal/config/ -run 'TestAudit' -v` — 새 섹션이 어느 감사를 건드리는지 미리 본다
5. 층 A census 베이스라인 재측정 후 `.moai/reports/t1030/`에 보존 (두 층 모두 — §E)

## §D Constraints

`spec.md` §C 전부. 특히: 행위 보존(REQ-MCS2-003), Template-First(`make build` 임베드), 템플릿 중립성(§25 C-class), 섹션 로더 관례 + 감사 레지스트리 등록(REQ-MCS2-002), 시드 단일성(REQ-MCS2-010).

## §E Self-Verification

각 마일스톤 종료 시 대상 패키지 `go test` + `golangci-lint run` 실측 출력을 보고한다.

최종 census 재측정은 **두 층을 분리해 보고한다** — 한 층만 인용하면 판정이 서지 않는다:

```bash
# raw 층 (주석 포함) — 변하지 않아도 정상. 감소를 성공 지표로 읽지 않는다.
grep -rn -E "$(cat .moai/reports/t1030/census.re)" internal/ --include='*.go' \
  | grep -v '_test.go' | tee census-raw-after.txt | wc -l
cut -d: -f1 census-raw-after.txt | sort -u | wc -l

# 코드 층 (비주석) — 30 → 13, 6파일 → 4파일이 판정 대상이다.
grep -rn -E "$(cat .moai/reports/t1030/census.re)" internal/ --include='*.go' \
  | grep -v '_test.go' | grep -v -E ':[0-9]+:[[:space:]]*(//|#)' \
  | tee census-code-after.txt | wc -l
cut -d: -f1 census-code-after.txt | sort -u | tee census-files-code-after.txt
```

**양성 대조 의무**: 수렴 후 0행을 주장하는 명령마다 같은 명령 형태의 비어 있지 않은 대조를 짝짓는다(예: `memory.go` 0행 ↔ `defaults.go` 9행). 양성 대조 없는 무출력은 부재가 아니라 미측정이다.

## §F Milestones (결정 가역성 순)

데이터 모델(바뀔 가능성 최대) → 와이어 동작 → 기계적 정리 순. 각 단계는 독립 커밋 가능하고, 어느 단계에서 멈춰도 트렁크는 녹색이다.

### M1 (High) — 카탈로그 스키마·로더·기본값 (순수 추가, 동작 변화 0)

가장 되돌리기 비싼 결정이 여기 있다 — 스키마 필드 집합과 시드의 소유권이다.

1. `ModelCatalogConfig` / `ModelCatalogEntry` 타입 (`internal/config`), top-level 키 `model_catalog:`
2. `NewDefaultModelCatalogConfig()` — **`internal/config/defaults.go`의 GLM `Default*` 상수를 참조**해 구성한다. 모델 id 문자열을 다시 적지 않는다 (REQ-MCS2-010). Claude 계열 id는 `internal/template/model_policy.go`의 `ModelIDOpus5` / `ModelIDOpus48` 상수를 참조하거나, 패키지 의존 방향이 막으면 `internal/config` 쪽에 시드 상수를 신설하고 `model_policy.go`가 그것을 참조하도록 방향을 **한 쪽으로만** 정한다 — 양방향 참조는 시드를 둘로 만든다
3. `Validate()` — REQ-MCS2-006 전항목. 오류 메시지에 entry id + 필드명 포함
4. 섹션 매니저 배선 + **감사 레지스트리 2곳 등록**: `internal/config/audit_registry.go` `yamlToStructRegistry`, `internal/config/types.go:1125` `sectionNames`
5. `.moai/config/sections/model_catalog.yaml` + 템플릿 미러 작성, `make build`로 임베드
6. 로더 단위 테스트 (85%+): 정상 로드, 부재→기본값, 타입 불일치, REQ-MCS2-006 결함 fixture 7종
7. 판정: `go test ./internal/config/...` 녹색 + 기존 감사 테스트(`TestAuditParity` 계열) 녹색 + 기존 스위트 무변화

### M2 (High) — model policy alias 수렴

사용자 대면 동작(wizard picker, prefs 정규화)을 바꿀 수 있는 축이라 윈도우보다 앞에 둔다.

1. **RED (characterization 선행 커밋)**: `ModelAliasCanonicalID("opus"/"sonnet"/"fable"/"haiku"/"opusplan")`, `ModelAliasFromCanonicalID("claude-opus-5")`, 폐기 id 정규화 4행(`claude-opus-4-6`/`4-7`/`4-8`→`opus`, `claude-sonnet-4-6`→`sonnet`), 그리고 표에 없는 입력의 무변경 반환
2. **GREEN**: `ModelAliasTable` / `ModelDeprecatedCanonicalIDs`를 카탈로그 alias 데이터에서 빌드. `opusplan` 자기매핑은 카탈로그의 `aliases` 데이터로 표현하되 "전체성(totality)" 성질을 잃지 않게 한다
3. `@MX:ANCHOR` 이동 갱신 + 태그 변경 보고 (fan_in ≥ 3)
4. 소비자 3곳(`launcher.go` / `profile_setup.go` / `settings/schema.go`) 테스트 녹색 확인
5. 판정: RED 테스트가 **무편집으로** 통과 + `internal/template` / `internal/cli` / `internal/settings` 패키지 녹색

### M3 (Medium) — statusline 윈도우 수렴

1. **RED**: 우선순위 2 fixture — (a) `llm.yaml`의 `glm.context_windows` 부재/비어 있음 → 빌트인 값(flash/5.3/5.2 = 1,000,000 · 5.1 = 200,000 · glm-5/4.7/4.6/4.5/4.5-air = 128,000), (b) 1건 이상 존재 → **전면 대체**(병합 아님). 그리고 미등록 `glm-5.3-*` 변종의 1M 부분문자열 상속, `glm-4.5` vs `glm-4.5-air` longest-match 비마스킹, 카탈로그 손상 시 fail-open
2. **GREEN**: `glmContextWindows` 빌트인 map을 카탈로그 데이터 판독으로 교체. 우선순위·매칭·fail-open 구조는 손대지 않는다 — 데이터 홈만 옮긴다
3. 핫 패스 성질 보존: 현행과 동일 등급의 1회 파일 읽기. 캐시 도입 여부는 현행 대비 지연 측정으로 판단하며, 현행도 매 렌더 파일을 읽으므로 필수가 아니다
4. 판정: RED 무편집 통과 + `internal/statusline` 녹색 + `memory_test.go` 기존 7적중 테스트 무변화

### M4 (Low) — 가드·최종 검증

1. **REQ-MCS2-007 가드 테스트 작성** — 설계가 곧 판정이므로 두 성질을 테스트가 **직접 단정**한다:
   - 축: 비주석 필터를 적용한다. 필터를 끄면 17파일이 보인다는 사실을 가드 자신의 fixture로 고정해, 다음 사람이 필터를 「정리」하지 못하게 한다
   - 열거: `internal/` 전수 스캔으로 얻은 비주석 리터럴 보유 파일 집합이 선언된 예외 집합 4파일과 **정확히 같은지** 단정한다(부분집합 아님 — 7번째 파일이 나타나면 실패해야 한다)
2. census 두 층 재측정 보고 (§E 명령 verbatim, 양성 대조 포함)
3. `go vet ./internal/...` · `golangci-lint run` 실측
4. 잔여 위험 기록: `glm.go:310`의 1M 집합 재진술 (`spec.md` §E)

## §G Anti-Patterns

- **「같이 고치면 깔끔하다」식 현행값 교정 금지** — 벤더 문서와 어긋나는 현행 값은 그대로 인코딩하고 노트만 단다 (REQ-MCS2-003 위반)
- **`Defaults()`에 모델 id 리터럴을 다시 적기 금지** — 시드가 둘이 되면 이 카드가 없애려는 결함을 이 카드의 산출물이 싣는다 (REQ-MCS2-010)
- **가드를 6개 파일명 고정 목록으로 쓰기 금지** — 7번째 파일의 재유입을 원리상 못 본다 (REQ-MCS2-007)
- **가드에서 비주석 필터 생략 금지** — 주석만 담은 11파일을 전부 위반으로 신고하는 과잉 가드가 된다 (같은 요구사항)
- **소멸한 축에 대한 코드 작성 금지** — `internal/gateway`, `gptEffortAllowlist`, `agentFMModelValues()` 수렴은 전부 Out of Scope다
- **`internal/web/agentfm.go` 손대기 금지** — `SPEC-MODEL-PROFILE-MATRIX-002` M3·M5가 미착수 상태로 그 파일을 겨누고 있다
- **한 층만 인용한 census 보고 금지** — raw만 보면 성과가 없어 보이고, 코드층만 보면 가드 설계 근거가 사라진다

## §H Cross-References

- `spec.md` / `acceptance.md` / `progress.md`
- `internal/config/CLAUDE.md` — 섹션 로더 관례
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — 템플릿 중립성
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — M2의 `@MX:ANCHOR` 이동 보고
- `.moai/reports/t1030/` — census 두 층, MATRIX-002 겹침, REQ 생존표
