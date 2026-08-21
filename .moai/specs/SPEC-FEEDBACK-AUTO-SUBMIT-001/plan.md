# Plan — SPEC-FEEDBACK-AUTO-SUBMIT-001

> 구현 계획. 경로는 워크트리 상대. 순서는 **되돌리기 어려움 순** — 바뀔 확률이 높은 결정(데이터 모델, 타입 계약, 사용자 대면 흐름)을 앞에, 기계적 편집을 뒤에 둔다.
> 설계 근거(경계·파이프라인 순서·실패 모드)는 `design.md`, 재사용/신규 판정은 `research.md`에 있다. 여기서 재론하지 않는다.

## §A Context

카드 t170의 **피드백 축**을 구현한다. todo 축은 `SPEC-TODO-ENABLE-FLAG-001`로 분리됐다.

run-phase 모듈: `internal/feedback`(신규) · `internal/sandbox`(접근자 1개) · `internal/cli` · `internal/config` · `internal/settings` · `internal/web` · 스킬 본문 2사본 · 템플릿 · docs-site 4로케일.

Route B(PR 경로) — Tier L이기도 하고, 이 저장소는 `.claude/rules/local/repo-local-pr-policy.md` 가 전 티어에 PR을 강제한다(`main` 이 `enforce_admins: true`).

**열린 결정 없음.** iter1의 결정 D5(웹 노출 경로)는 iter2에서 **선택지 A로 확정**됐다(`spec.md` §D). 착수 승인 시점에 운영자가 그 반전(SPEC-WEBCONF-SIMPLIFY-001 M3)을 재고할 수는 있으며, 그 경우 REQ-1·REQ-12와 §D를 함께 개정하는 SPEC 개정 사안이다 — 계획이 진행 중에 갈라지는 유예가 아니다.

## §B Known Issues (착수 전 알고 있어야 할 함정)

1. **카드가 지목한 배선 선례가 사장 코드다.** `internal/cli/init_autonomy_wizard.go`의 `applyAutonomyTierFromWizard`는 프로덕션 호출자가 없다. 따라 쓰면 질문이 물어지고 `WizardResult`에 저장된 뒤 버려진다. 살아 있는 경로는 M6에 명시했다. (이웃 둘 `applyWorkflowBranchGuardFlags`·`writeWorkflowAuditYAML`도 같은 상태 — `research.md` §5.)
2. **마법사 질문 추가는 번역 테스트를 깨뜨린다.** `TestWizardQuestionTranslationCompleteness`(`translations_completeness_test.go:89`)는 번역 없는 신규 질문에서 실패하도록 설계돼 있다. 고칠 테스트가 아니라 따를 관례다.
3. **`DefaultQuestions`에 넣으면 개수 고정 테스트가 깨진다.** `TestQuestionOrder`(`questions_test.go:101`) 5개, `TestReconfigureQuestions...`(`:190-210`) 12개.
4. **설정 키 1개가 파일 9곳을 건드린다.** 두 anti-rot 가드가 등록을 강제한다(`shipped_key_reader_test.go:70`, `schema_label_test.go:96`). 목록은 M8.
5. **`feedback` 섹션은 현재 `RouteExcluded`다.** 두 테스트가 고정한다(`sectionroute_test.go:27`, `scope_contract_test.go:79`). 선택지 A는 **결정을 뒤집는 편집**이지 조용히 고칠 테스트가 아니다.
6. **형제 SPEC과 파일 9종을 공유한다** — `spec.md` §E.1의 표와 [HARD] 병합 규율. 두 번째로 착지하는 쪽이 마법사 개수·번역 테스트를 다시 돌린다.
7. **분류를 마스킹 이후에 하면 조용한 미탐이 된다** — `design.md` §3. 파이프라인 순서는 설계 결정이지 구현 재량이 아니다.

## §C Pre-Flight (run-phase 진입 전 확인)

M1 착수 전 1회, 병렬 배치로:

```bash
go test ./internal/config/... ./internal/cli/wizard/...      # 초록 baseline
go build ./...                                                # 빌드 baseline
grep -rn "auto_submit\|AutoSubmit" --include='*.go' internal/ | grep -v _test   # 0건이어야 함
grep -n 'AskUserQuestion\|gh issue create' .claude/skills/moai/workflows/feedback.md  # baseline 기록(AC-F-002 반증용)
```

`spec.md` §D 결정 D5 확정 여부 확인 — 미확정이면 M1~M6은 진행 가능하나 M7은 차단이다.

## §D Constraints (Hard)

- 템플릿 주석에 SPEC ID·REQ 토큰 금지(CLAUDE.local.md §2.1). CI 가드가 잡는다.
- 값 마스킹 출력 형태를 새로 만들지 않는다.
- 마스킹 로그·`findings`에 원문 값을 절대 넣지 않는다.
- 제출 경로 fail-closed, 로깅·큐잉 fail-open. 두 축을 섞지 않는다.
- 형제 SPEC 공유 파일은 **다른 항목만 추가**한다(`spec.md` §E.1).
- 로컬 검증은 패키지 스코프로만. `go test ./...`는 로컬에서 돌리지 않는다(CLAUDE.local.md §4).
- 건드린 모든 패키지에 `GOOS=windows go vet`.
- 신규 테스트는 `t.TempDir()`, 병렬 테스트에 `t.Setenv("HOME", ...)` 금지.

## §E Self-Verification (설계 결정 요약)

전체 설계 판단은 `design.md`에 있다. 여기서는 run-phase가 코드를 쓰기 직전에 다시 볼 4가지만 재확인한다.

- **P1 패키지 경계**: 변환 로직은 `internal/feedback`, CLI는 얇은 배선(`design.md` §2).
- **P2 패턴 재사용**: 정책 객체를 통째로 받는다. 패턴 문자열 복사 금지(`design.md` §2, `research.md` §3).
- **P3 판정/종료코드 축 분리**: `verdict`는 stdout JSON, 종료 코드는 도구 실패만(`design.md` §4).
- **P4 파이프라인 순서**: 분류는 원문, 변환은 env → 시크릿 → 홈경로, 결과는 멱등(`design.md` §3).

## §F Milestones (되돌리기 어려움 순)

### M1 — 설정 데이터 모델 (형태가 바뀔 확률 최대)

**파일**:
- `internal/config/types.go:1310-1314` — `FeedbackConfig`에 `AutoSubmit bool \`yaml:"auto_submit"\``.
- `internal/config/defaults.go` — `DefaultFeedbackAutoSubmit = false` 상수(`:212` 부근) + `NewDefaultFeedbackConfig()`(`:451-455`)에서 세팅.
- `internal/config/feedback_accessors.go` — `func (c *Config) FeedbackAutoSubmit() bool`(`:20` 이후).

**Exit**: `go test ./internal/config/...` 초록. 부재→false, 명시 true→true 를 단언하는 테스트 존재(AC-F-001).

### M2 — 스크러버 타입 계약 + 마스킹 변환 (신규 타입 계약)

**파일**:
- `internal/feedback/scrub.go` — `Result` / `Finding`(`Where` 포함) 타입, `Scrub(in Input, opt Options) (Result, error)` — `Input`은 **제목과 본문 둘 다** 담는다(REQ-3 D1).
- `internal/feedback/patterns.go` — 정책 재사용 + `AIza` 합집합 + **치환 span 규칙**(REQ-4 하위 조항): 마커 앵커 패턴은 블록 종료자까지, 대소문자 민감 패턴은 `(?i)` 없이 재컴파일.
- `internal/feedback/paths.go` — `paths.Home()` 기반 홈 축약 + `.moai/` 마커 상향 탐색으로 프로젝트 루트 해석(REQ-3 D5).
- `internal/feedback/env.go` — 이름 어휘 기반 값 마스킹.
- **`internal/sandbox/env.go` — `func DefaultEnvDenyList() []string` 신설**(REQ-6 D3). `defaultDenyList`의 **사본**을 반환한다(호출자가 원본을 변경하지 못하게). 이 SPEC이 `internal/feedback` 밖 패키지를 편집하는 유일한 지점이다.
- `internal/feedback/scrub_test.go` — 항목별 양극성(마스킹돼야 함 / 되면 안 됨) + 멱등성 + **개인키 블록 전체 마스킹**.

**Exit**: `go test ./internal/feedback/... ./internal/sandbox/...` 초록. 대응 AC: **F-005 ~ F-011, F-014, F-024**.

### M3 — 취약점 분류기 (선례 없는 유일 부분)

**파일**:
- `internal/feedback/classify.go` — 신호 3종(`design.md` §6), 어휘는 상수, 거부 메시지는 `SECURITY.md` 인용.
- `internal/feedback/classify_test.go` — 차단 케이스 + **오탐 대조** 케이스(축퇴 구현 배제).

**Exit**: `go test ./internal/feedback/...` 초록. 대응 AC: **F-008, F-012, F-013**.

**F-013을 빠뜨리지 않는다**: 분류가 **마스킹 이전 원문**을 본다는 순서 가드이며, 역순은 `design.md` §3이 "조용한 미탐이며 테스트로 잡기 어렵다"고 부른 바로 그것이다. 분류기를 만드는 이 마일스톤이 그 순서를 고정하는 유일한 자리다.

### M4 — 온디스크 산출물 2종 (형식 계약)

**파일**:
- `internal/feedback/masklog.go` — `.moai/logs/feedback-mask.log`, `0o600`, fail-open.
- `internal/feedback/queue.go` — `.moai/state/feedback/queue.json`, `BacklogStore` 형태.
- 대응 테스트 2종.

**Exit**: `go test ./internal/feedback/...` 초록. 대응 AC: **F-015 ~ F-018**(로그 내용·권한 / 로그 fail-open / 큐 적재 / 큐 제거). 권한 단언은 Windows skip.

**주의(D4)**: 큐는 `gh issue create` **실패** 전용이다. `gh auth status` 실패·레이트리밋은 기존 초안 경로(`feedback.md:36-44`, 스크럽 이전 원문)가 담당한다. 재전송 코드가 `.moai/state/feedback-draft-*.md` 를 큐 항목으로 읽어서는 안 된다 — 읽으면 스크럽 이전 원문이 공개 이슈로 나간다.

### M5 — CLI 배선 (`moai feedback scrub` + 큐 동사)

**파일**:
- `internal/cli/feedback.go` — `feedback` 부모 명령 + `scrub` 서브커맨드. 인자: `--title <제목>`(REQ-3 D1), `--root <path>`(REQ-3 D5, 미지정 시 `.moai/` 마커 상향 탐색). 본문은 stdin, 결과는 stdout JSON 한 덩어리. 큐 조작 동사(`queue enqueue|list|resolve`)는 스킬 본문이 호출할 최소 집합만.
- 등록 지점(`rootCmd.AddCommand`).
- `internal/cli/feedback_test.go` — 계약(AC-F-003, `title` 필드 포함) + 도구 실패 fail-closed(AC-F-004).

**Exit**: `go test ./internal/cli/...` 초록. 대응 AC: **F-003, F-004**. 빌드된 바이너리로 스모크 1회(`--title` 포함).

### M6 — 스킬 본문 + 마법사 질문 (사용자 대면 흐름)

**파일 — 스킬**:
- `.claude/skills/moai/workflows/feedback.md` + `internal/template/templates/.claude/skills/moai/workflows/feedback.md` — (a) 스크러버 경유 [HARD] 조항(verbatim 규칙 `:104`의 명시적 예외로 기술), **제목과 본문을 함께 통과시킬 것**(`--title`), (b) `gh issue create`(`:118`) 앞에 확인 게이트 3옵션(`design.md` §7) — 라벨과 findings 요약은 `conversation_language`, 템플릿 미러 예시는 영어(D11), (c) **3문장 [HARD] 조항**(`design.md` §9): 종료코드 ≠ 0 → 제출 금지 / `verdict != ok` → 제출 금지(필드 부재·파싱 불가 포함) / 60초 무응답 → 중단, (d) `gh issue create` 실패 시 큐잉 — 단 `gh auth status` 실패·레이트리밋은 기존 초안 경로(`:36-44`)가 담당한다는 분기를 명시(D4), (e) 게이트가 보여주는 것은 **마스킹된 제목 + 마스킹된 본문 전문 + findings 요약(위치 포함)**.

**파일 — 마법사(살아 있는 경로만)**:
- `internal/cli/wizard/questions.go` — `Page3Questions`의 "Quality & Workflow" 그룹에 `feedback_auto_submit`(`Default: "false"`).
- `internal/cli/wizard/types.go` / `wizard.go:459` / `translations.go`(ko/ja/zh) / `internal/cli/init.go:185` / `internal/core/project/initializer_expansion.go:30`(yamlpatch writer).
- 테스트: `wizard/worktree_test.go`(`:8`, `:29`, `:47`) 3종 세트 복제.

**Exit**: `go test ./internal/cli/wizard/... ./internal/cli/... ./internal/core/project/...` 초록. 대응 AC: **F-002, F-019 ~ F-022**(게이트 존재 / 스킬 [HARD] 조항 / 마법사 질문 / 번역 / 파일 기록).

**주의**: 스킬 소스/템플릿 쌍의 기존 1줄 드리프트를 확대하지 않는다(`research.md` §5-3).

### M7 — 웹 콘솔 노출 (결정 D5 확정 — 선택지 A)

**파일**: `internal/settings/schema_sections.go`(필드) + `internal/settings/sectionroute.go`(`RouteSeam` + `ExcludedSections()`에서 제거) + `internal/web/schemaform.go`(탭·패널) + `internal/web/assets/i18n.js`(4로케일) + 고정 테스트 2건 갱신(`internal/settings/sectionroute_test.go:27`, `internal/web/scope_contract_test.go:79`).

[HARD] 두 테스트 갱신은 **결정 반전**이다 — 커밋 본문에 SPEC-WEBCONF-SIMPLIFY-001 M3 반전임을 명시한다(AP-9).

**Exit**: `go test ./internal/settings/... ./internal/web/...` 초록. 대응 AC: **F-023의 웹 절반**(스키마·라우트·i18n).

### M8 — Template-First 미러 + 키 인벤토리 (기계적, 누락 시 CI 실패)

**파일**:
- `internal/template/templates/.moai/config/sections/feedback.yaml` — `auto_submit: false` + 중립 주석.
- `.moai/config/sections/feedback.yaml`, `internal/settings/testdata/sections/feedback.yaml` — 템플릿 밖 사본 2개.
- `internal/config/testdata/shipped_key_inventory.yaml` — `feedback.auto_submit` 항목(`evidence`에 "스킬 본문이 소비, Go 호출자 없음"을 정직하게 기록 — `design.md` §8).
- `internal/settings/schema_sections_test.go:452-468` — per-key 맵에 기대값 추가.
- `make build`.
- `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` — "Feedback Settings" 절(en:91 / ko:91 / ja:95 / zh:95)에 4로케일 동시 반영(CLAUDE.local.md §17).

**Exit**: `go test ./internal/config/... ./internal/template/... ./internal/settings/...` 초록. `make build` 성공. 대응 AC: **F-023의 템플릿 절반**(미러·인벤토리·빌드·중립성).

### M9 — 검증 스윕 + PR

```bash
go test ./internal/feedback/... ./internal/sandbox/... ./internal/config/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/... ./internal/web/... ./internal/template/...
go test -race ./internal/feedback/...
GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/... ./internal/config/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/...
golangci-lint run --timeout=2m
make build
```

커밋: 마일스톤 단위 Conventional Commits, footer에 SPEC ID. Tier L이므로 PR 경로.

## §G Anti-Patterns (피할 것)

- **AP-1**: `init_autonomy_wizard.go` / `applyWorkflowBranchGuardFlags` / `writeWorkflowAuditYAML`를 배선 선례로 따르기 — 셋 다 사장 코드.
- **AP-2**: 신규 질문을 `DefaultQuestions`에 넣기.
- **AP-3**: 마법사 문자열을 영어 단독으로 싣고 번역 테스트 실패를 "낡은 테스트"로 읽기.
- **AP-4**: 시크릿 패턴 문자열을 스크러버에 복사.
- **AP-5**: 값 마스킹 출력 형태를 네 번째로 신설.
- **AP-6**: `findings`·마스킹 로그에 원문 값 남기기 — 로그가 유출 경로가 된다.
- **AP-7**: 재시도 큐를 append-only JSONL로.
- **AP-8**: 분류를 마스킹 이후에 수행 — 조용한 미탐.
- **AP-9**: `feedback` 라우트 고정 테스트 2건을 조용히 고치기.
- **AP-10**: 형제 SPEC 공유 파일에서 기존 항목 재배치·재서식.
- **AP-11**: 로컬에서 `go test ./...` 실행.
- **AP-12**: 스크러버 도입을 "마스킹이 이제 강제된다"로 보고하기 — 규약 강제이지 샌드박스가 아니다(`spec.md` §E.3).
- **AP-13**: `hook` 원본 패턴에 `AIza`를 추가해 Write/Edit deny 판정을 함께 넓히기 — 범위 밖 동작 변경.

## §H Cross-References

- SPEC: `spec.md` · 설계: `design.md` · 조사: `research.md` · 수용: `acceptance.md`
- 형제 SPEC: `.moai/specs/SPEC-TODO-ENABLE-FLAG-001/`
- 근거 렌즈: `.moai/reports/t170/lens-{feedback,masking,init,web-todo}.md`
- CLAUDE.local.md §2(Template-First), §2.1(중립성), §4(검증 규율), §6(테스트 격리), §14(하드코딩), §17(4로케일 문서)
