# Research — SPEC-FEEDBACK-AUTO-SUBMIT-001

> 계획 단계 조사 요약. 원본은 `.moai/reports/t170/lens-{feedback,masking,init,web-todo}.md`(총 83KB, 모든 주장이 `file:line` 동반). 여기서는 **설계 판단에 쓰인 결론만** 옮기고, 재측정 없이 뒤집지 않는다.

## §1 조사 방법

읽기 전용 렌즈 4종을 병렬로 돌렸다. 실행한 명령은 `grep` / `diff` / `git log` 뿐이며 **어떤 테스트도 실행하지 않았다** — 각 렌즈가 자기 보고서에 그 사실을 gap으로 명시했다. 따라서 아래의 "테스트가 실패한다" 류 서술은 전부 **테스트 본문을 읽고 내린 예측**이지 관측이 아니다. run-phase에서 실제로 돌려 확인해야 한다.

## §2 재사용 / 신규 판정표 — 이 SPEC의 핵심 산출

| 요구 조각 | 판정 | 대상 | 근거 |
|---|---|---|---|
| 시크릿 패턴 집합 | **그대로 재사용** | `hook.DefaultSecurityPolicy().SensitiveContentPatterns` (`internal/hook/pre_tool.go:262-273`) | 이미 export, `(?i)` 컴파일, `security.extra_sensitive_content_patterns`로 확장 가능 |
| 패턴 완전성 | **확장 1건** | `AIza[0-9A-Za-z_-]{35}` (`.moai/astgrep-rules/security/credentials.yml`) | `hook`의 목록에 없음. 두 집합은 서로 포함관계 아님 |
| 세 번째 Go 패턴 목록 | **조사됨 — 미채택** | `internal/github/workflow/validator.go:155` (`AIza[0-9A-Za-z\\-_]{35}` 포함) | D9: 워크플로 **파일 검증**용 리터럴 목록이며 재사용 가능한 정책 객체가 아니다. `AIza` 를 이미 담고 있다는 사실은 위 확장 판단을 뒤집지 않는다(합집합 방향 동일). iter1의 "Go 목록에는 없다"는 서술이 리포 전체로는 거짓이었고 `hook` 한정으로 정정됨 |
| 마스킹 변환 자체 | **신규** | — | 텍스트를 재작성하는 함수가 리포에 0건. 기존 3종은 전부 값 단위 표시용 |
| 마스킹 출력 형태 | **재사용(택1)** | `MaskSecret` / `maskAPIKey` / `maskPartial` | 형태가 이미 3종 — 네 번째 금지 |
| 환경변수 이름 어휘 | **재사용 — 단 접근자 신설 필요** | `internal/sandbox/env.go:32` `defaultDenyList` + `env_scrub_extra` | D3: 이 변수는 **unexported**라 `internal/feedback`이 import할 수 없고, 그 표면의 유일한 export는 이미 배제한 `ScrubEnv`(`:51`)다. `DefaultEnvDenyList()` 접근자를 신설해 재사용한다(복사 금지). iter1은 export 여부를 확인하지 않고 "재사용"으로 적었다 — REQ-4는 확인했는데 REQ-6은 하지 않은 비대칭 |
| 홈 경로 축약 | **신규** | — | `~` 축약 헬퍼 0건. grep **3건**은 모두 반대 방향(`core/git/branch.go:179` refname 검사, `shell/detect.go:135` 폴백 리터럴, `shell/config.go:222` `HasPrefix(path, "~/")` 확장). D10: iter1은 2건으로 셌으나 결론(축약 헬퍼 부재)은 불변 |
| 홈 해석 | **재사용** | `paths.Home()` (`internal/paths/paths.go:49`) | `paths.go:8`의 HOME-first 계약 |
| 취약점 분류 | **신규** | — | 분류기·CVE/CWE 탐지기·어휘 전부 0건. **이 SPEC 최대 신규 작업** |
| 취약점 정책 문구 | **재사용** | `SECURITY.md` | 인간용 규칙이 이미 존재 — 재발명 금지 |
| 분류의 구조적 선례 | **참고** | 템플릿 중립성 가드(`.github/workflows/template-neutrality-check.yaml` + `internal_content_leak_test.go`) | "내용을 분류해 배포를 거부"라는 같은 형태, 다른 분류기 |
| 마스킹 로그 형태 | **재사용** | `internal/config/log.go:21-65` | mkdir → `O_APPEND\|O_CREATE\|O_WRONLY` → RFC3339 항목 → best-effort |
| 마스킹 로그 권한 | **재사용(다른 쪽)** | `internal/hook/failure_observer.go:156`의 `0o600` | 주제가 시크릿 인접이므로 config.log의 `0o644` 대신 |
| 재시도 큐 | **형태 재사용, 코드 신규** | `internal/kanban/backlog_store.go` | 단일 JSON + 형제 lock + `Mutate()` + `atomicfile.Replace`. 실패한 외부 전송 스풀은 리포에 0건 |
| 전송-1회 대안 | **미채택(기록)** | handoff claim-rename (`handoff_inject.go:116,252`) | 재시도가 아니라 1회 소비 의미론 |
| 웹 토글 배선 | **재사용** | `internal/settings/schema_sections.go:334` (`workflow.branch_guard.enabled`) | 라우팅된 섹션이면 필드 1줄, 나머지는 제네릭 |
| 마법사 질문 형태 | **재사용** | `worktree_auto_create` (`wizard/questions.go:365-374`) + `worktree_test.go`의 3종 테스트 세트 | |
| 설정 파일 기록 | **재사용** | `yamlpatch.PatchFile` via `WritePhase1Configs` (`initializer_expansion.go:30`) | 주석 보존 + 두 배포 경로 모두에서 실행 |

## §3 기각한 대안과 그 이유

| 대안 | 기각 사유 |
|---|---|
| 패턴 문자열을 스크러버에 복사 | `extra_sensitive_content_patterns` 확장이 전달되지 않아 두 목록이 갈라진다. `MergeExtraPatterns`(`pre_tool.go:300-318`)가 "확장은 기본값을 대체하지 않는다"를 이미 보장하므로 정책 객체를 통째로 받는 쪽이 그 계약을 상속한다 |
| 스크러버를 `internal/hook`에 배치 | 훅 패키지가 CLI용 코드를 떠안고, 무관한 테스트가 큰 스위트에 얹힌다 |
| 스크러버를 `internal/cli`에 직접 배치 | 변환 테스트가 cobra 조립에 결합된다 |
| `blocked`를 종료 코드로 인코딩 | 도구 실패와 정책 차단이 같은 채널에서 섞인다 |
| 마스킹 본문 stdout + 판정 stderr | 판정이 사람용 로그와 섞인다 |
| 재시도 큐를 append-only JSONL로 | 성공 시 삭제를 표현할 수 없다 |
| 로그와 큐를 한 형태로 통일 | 요구가 반대다 — 로그는 잠금 없는 fail-open, 큐는 잠금 있는 read-modify-write |
| 마스킹 후 본문을 분류 | 마스킹이 분류 신호를 지운다. 조용한 미탐(design.md §3) |
| `hook` 원본 패턴에 `AIza` 추가 | Write/Edit **deny** 판정이 함께 넓어진다 — 기존 동작 변경이라 별도 판단 |

## §4 카드 전제 반증 — 측정 근거

| 카드 전제 | 반증 근거 | 이 SPEC의 대응 |
|---|---|---|
| `auto_submit=true`가 확인 질의를 생략 | `AskUserQuestion`이 `:52`/`:156`/`:178` 세 곳뿐이고 그 사이 없음. 영문 문서가 "automatically" 명시 | 게이트를 **신설**(D1) |
| 마스킹을 기계화 | 경로 전체가 산문. `gh issue create` 플래그 집합이 문자열로도 없음 | Go 스크러버 신설 + 한계를 §E.3에 명시(D2) |
| init 문구는 영어(16언어 중립성) | 프롬프트가 Go 소스이고 4로케일이 관례. 영어 단독은 완전성 테스트 실패 | 4로케일 동반(D4) |

## §5 이 SPEC이 건드리지 않기로 한 발견들

조사가 부수적으로 찾아낸 결함들. 전부 **범위 밖**이며 후속 카드 후보다.

1. **사장 writer 3건** — `applyAutonomyTierFromWizard`(`internal/cli/init_autonomy_wizard.go:34`), `applyWorkflowBranchGuardFlags`(`init_workflow_flags.go:36`), `writeWorkflowAuditYAML`(`initializer_audit.go:37`). 전부 프로덕션 호출자 0건(테스트 제외). 두 번째의 결과로 `--branch-guard` / `--worktree-auto-*` 플래그 4개가 등록만 되고 적용되지 않는다. **카드가 배선 선례로 지목한 파일이 바로 첫 번째다** — 형제 SPEC(`SPEC-TODO-ENABLE-FLAG-001`) §B에도 같은 경고가 있다.
2. **`feedback` 섹션의 잠재 불일치** — `parseSchemaForm`이 라우트와 무관하게 `PersistSeam` 필드를 받아들이므로, 위조 POST가 `feedback.repository`를 밀면 `WriteSectionViaSeam`에서 하드 에러가 나고 "프로필은 저장됐으나 섹션 쓰기 실패"라는 부분 성공 상태가 사용자에게 노출된다. 결정 D5 선택지 A를 택하면 부수적으로 해소된다.
3. **스킬 본문 미러 드리프트 1줄** — 소스 `feedback.md:184`에 `Last Updated: 2026-02-07`가 있고 템플릿에는 없다. 이 파일 쌍은 byte-parity 허용목록 밖이라 강제되지 않는다. 편집하면서 드리프트를 확대하지 않는다.
4. **`internal/telemetry` 무마스킹** — 훅 트레이스 표면에 redaction이 0건. 스크러버의 잠재 소비자이지만 이 SPEC 범위 밖.

## §6 조사가 확인하지 못한 것 (gap)

- **테스트를 하나도 실행하지 않았다.** "실패한다/통과한다"는 전부 소스 읽기 기반 예측이다.
- 오케스트레이터가 실제 세션에서 일반 askuser 규약에 따라 임의의 확인 질의를 넣는지는 소스로 알 수 없다 — 런타임 행동 문제다. 그래서 REQ-2는 "있는 것을 고친다"가 아니라 "없는 것을 만든다"로 썼다.
- `internal/web`의 렌더·영속 경로를 스키마·i18n 등록 지점 너머까지 훑지 않았다. 새 bool 필드가 건드리는 웹 표면이 더 있을 수 있다.
- `internal/cli/init_test.go`(23KB), `init_coverage_test.go`, `wizard/wizard_test.go`(42KB)를 전량 읽지 않았다. 추가 개수/순서 단언이 있을 수 있다.
- 사장 코드 3건 판정은 `grep -rn ... | grep -v _test` 기반이다. 인터페이스 값·생성 파일·빌드 태그를 통한 호출은 잡히지 않는다(셋 다 unexported 평범 함수라 가능성은 낮지만 증명되지 않았다).
