# t319 verdict — tab_schema.json 을 읽으라고 알려주는 곳이 없다

카드: 인터뷰 스키마가 고아다 — 소비자 0건. 트리: `WT-tab-schema-orphan` @ develop `2660bcd09`, darwin.

## 판정 (카드의 두 질문에 대한 답)

- **(a) 실제 쓰이는 경로가 있는가** — 없다. 독립 재판독으로 카드 실측 일치: Go 참조는
  `internal_content_leak_test.go` 1건(중립성 검사, 스키마를 읽는 소비가 아님), 템플릿
  `.claude/` 전역 grep 0건, 스킬 SKILL.md 미참조, 그 외는 CHANGELOG·manifest 해시 목록뿐.
- **(b) 포인터 vs 은퇴** — **포인터 추가**. 근거: (1) t316(`914c4edf5` — tab_schema 의
  죽은 경로 auto_branch 질문 제거)이 이미 이 파일을 **정리** 쪽으로 정비 — 은퇴 논거는 t316
  착지로 소진; (2) 50KB 의 구조화된 인터뷰 정의(6 탭 · 46 field 질문 · 18 배치 — python
  실측)가 있고 /moai project 의 인터뷰 품질이 이를 따르게 하는 것이 자원 활용; (3) 카드 유의의
  "다음에 읽히기 시작할 때 맞아야 한다" — 포인터가 그 "시작"을 만들고, 같은 문구에 정합 책임을
  명시(config templates 가 권위, 어긋나면 스키마를 고친다).

## Claim

1. tab_schema.json 의 소비 경로는 수리 전 0건이었다 (전수 재판독).
2. 수리 후 소유 스킬의 SKILL.md 와 references/workflows.md — 인터뷰가 실제로 수행되는 소비
   시점 — 에서 스키마를 가리키며, 정합 책임(스키마 어긋나면 config templates 권위로 수리)을
   같이 선언한다.
3. 포인터 문구는 스테일 스키마가 죽은 키를 물어보게 하는 위험을 차단하는 지시를 포함한다.

## Evidence

| # | 명령 | 관측 출력 |
|---|------|----------|
| E1 | `grep -rln "tab_schema" internal/ --include="*.go"` + `.claude/` 전역 + repo 전체 | Go: `internal_content_leak_test.go` 1건(중립성 검사) · 템플릿 .claude/ 0건 · 나머지 CHANGELOG.md, .moai/manifest.json — 카드 실측과 일치 |
| E2 | `git log --oneline -3 -- …/tab_schema.json` | `914c4edf5 fix(schema): drop dead-path auto_branch questions from tab_schema 3.3/3.6 (t316)` — t316 정합 작업 착지 확인; 잔여 auto_branch 3건은 올바른 경로(`{mode}.automation.auto_branch`) |
| E3 | `python3` field 카운트 | tab_schema fields 46개, 6 탭 — 스키마 실체 실측 |
| E4 | 편집 후 `grep -rln "tab_schema" …/moai-workflow-project/` | `SKILL.md`, `references/workflows.md` 2건 — 포인터 확보 |
| E5 | `go run ./internal/template/scripts/gen-catalog-hashes.go --all` 후 `git diff --stat catalog.yaml` | `1 insertion(+), 1 deletion(-)` — moai-workproject SKILL.md 해시 1개만 (workflows.md 는 v1 스키마의 sub-file — t323 착지 후 무결성에 편승) |
| E6 (회귀) | `go test ./internal/template/ -count=1` | `ok 24.761s` |

## Baseline-attribution

모든 측정은 본 커밋 직전 이 워크트리(브랜치 `WT-tab-schema-orphan`, HEAD = `2660bcd09` = 로컬
develop)에서 이 실행으로 수행.

## Gaps

- **스키마 전체의 키 정합 전수 검증은 하지 않았다** — 46 개 field 를 config-schema.json 과
  yaml.tmpl 에 대조한 것은 t316 의 소관이었고 그 일부(auto_branch)만 확인. 포인터가 생긴 뒤
  스키마를 실제로 읽게 되는 첫 실행에서 어긋남이 나오면 config-templates-authoritative 문구가
  그 지시를 담는다.
- 포인터가 인터뷰 품질을 실제로 바꾸는지의 런타임 관측은 없다 — /moai project 실행 기반 확인은
  이 카드 범위 밖 (배차 메시지 규율상 통합 테스트는 /tmp 프로젝트 소관).
- overview.md 의 schemas 나열이 다른 구조(`__init__.py` 등)를 서술하는 이질 문서임을 발견했으나
  범위 밖으로 손대지 않음.

## Residual-risk

- **병합 순서 의존**: t323 이 catalog.yaml 을 전체 트리 해시로 재생성했고 본 카드는 SKILL.md 를
  고쳤다 — t323 이 develop 에 먼저 병합된 뒤 본 카드가 병합되면, 본 카드의 SKILL.md 내용 변경이
  트리 해시를 다시 바꾸므로 **본 카드 병합 전에 develop 을 흡수하고 gen-catalog-hashes --all 을
  다시 돌려야** CATALOG_HASH 무결성 테스트가 통과한다. 리드 창 지명 시 이 순서 필요.
- 스키마의 `config_version: "0.29.0"`·`schema_updated: 2025-12-22` 는 현재 버전 대비 오래된
  스탬프 — 스키마 내용이 현재 config 와 전수 정합한다는 보증은 없다 (위 Gaps 1행과 같은 축).
