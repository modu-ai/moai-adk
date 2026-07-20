# SPEC-CLI-TUX-V3-003 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. **Given** 사용자 소유 하네스 자산(`hns-*` skill, `.claude/agents/harness/`)이 존재하는 프로젝트에서, **When** TTY에서 `--yes` 없이 `moai update`를 실행하면, **Then** 프리뷰 테이블이 해당 자산을 `preserved (user-owned)`로 분류 표시하고, 승인 후 배포가 끝나도 그 파일들은 byte-identical하게 남아 있다.
2. **Given** CI 환경(non-TTY, `NO_COLOR=1`)에서, **When** `moai update --yes`를 실행하면, **Then** TUI 없이 동일 분류 모델 기반 텍스트 요약이 출력되고 ANSI 이스케이프가 한 바이트도 없으며 exit code는 분해 이전과 동일하다.
3. **Given** 템플릿 버전이 이미 최신인 프로젝트에서, **When** `moai update`를 실행하면, **Then** 통일 outcome 카드가 "Already up-to-date"를 표시하고 고속 스킵 경로(version-match)가 characterization과 동일하게 동작한다.
4. **Given** update.go가 5개 서브패키지로 분해된 후, **When** 전체 namespace 가드 테스트를 실행하면, **Then** 5계열 전부가 assertion 무수정 상태로 green이다.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-TUX3-001 | REQ-TUX3-001 | `go test ./internal/cli/update/... -run 'Classification\|ChangeClass' -count=1 -v` | PASS — 4-class enum 단일 타입 + 전 파일이 정확히 1개 class로 분류됨 단언 (신규 테스트 실존 grep 병행) |
| AC-TUX3-002 | REQ-TUX3-002 | `go test ./internal/cli/update/... -run 'PreserveClassSource\|PredicateShared' -count=1 -v` + 코드 리뷰: 프리뷰 분류부가 보호 predicate 함수를 직접 호출 | PASS — 분류가 `isUserOwnedNamespace` 계열 predicate에서 파생됨 (병렬 휴리스틱 부재; 호출 관계 grep으로 도달 가능성 증명) |
| AC-TUX3-003 | REQ-TUX3-003 | `go test ./internal/cli/... -run 'UpdateCharacterization' -count=1 -v` (분해 전 커밋과 분해 후 커밋에서 각 1회) | 양 시점 모두 PASS — 테스트 본문 무수정 (`git diff <M3b-commit>..HEAD -- <characterization 파일>` 공백) |
| AC-TUX3-004 | REQ-TUX3-004 | `ls internal/cli/update/plan internal/cli/update/backup internal/cli/update/deploy internal/cli/update/merge internal/cli/update/report && wc -l internal/cli/update.go` | 5개 서브패키지 실존 + root update.go가 cobra 배선 수준으로 축소(pre-flight 3,276 LOC 대비 대폭 감소 — 실측치 progress.md §E.2 기록) |
| AC-TUX3-005 | REQ-TUX3-005 | `go test ./internal/template/ -run 'TestSplitHarnessNamespaceNoLeak' -count=1 -v && go test ./internal/cli/... -run 'Namespace\|SecurityM2' -count=1` + `git diff <preflight-HEAD>..HEAD -- internal/template/split_namespace_test.go internal/cli/update_namespace_hns_test.go internal/cli/update_namespace_harness_v2_test.go internal/cli/update_namespace_harness_v4_test.go internal/cli/update_security_m2_test.go` | 가드 5계열 전부 PASS + **full diff** 내용이 import 경로 조정 라인에 한정(파일별 사유 progress.md §E.2 기재; assertion 라인 무변경 — §E E4와 동일한 full-diff 증거 형식; `--stat` 파일 단위 요약으로는 라인 무변경을 증명할 수 없음) |
| AC-TUX3-006 | REQ-TUX3-006 | `go test ./internal/merge/... -count=1` + `git diff <preflight-HEAD>..HEAD --stat -- internal/merge/` 검토 | merge 스위트 PASS (confirm.go 승격분 제외 3-way 로직 테스트 무수정) |
| AC-TUX3-007 | REQ-TUX3-007 | `go run ./cmd/moai update --help \| grep -E 'yes\|templates-only'` + `go test ./internal/cli/... -run 'UpdateCharacterization.*ExitCode\|UpdateFlagMatrix' -count=1 -v` | 플래그 표면 보존 + exit code characterization PASS |
| AC-TUX3-008 | REQ-TUX3-008 | `go test ./internal/cli/update/... -run 'PreviewTable' -count=1 -v` + `grep -rn 'charm.land/bubbletea/v2' internal/cli/update --include='*.go' \| grep -v '_test.go'` | PASS — 분류별 카운트 + 행 렌더 + 키보드 내비 계약; grep ≥ 1 (v2 런타임 실사용 도달 증명) |
| AC-TUX3-009 | REQ-TUX3-009 | `go test ./internal/cli/update/... -run 'DiffPreview\|Viewport' -count=1 -v` | PASS — 테이블 행 선택 → 파일 diff viewport 도달 단언 |
| AC-TUX3-010 | REQ-TUX3-010 | `go test ./internal/cli/update/... -run 'PreviewFallback\|YesNonTTY' -count=1 -v` | PASS — `--yes`/non-TTY에서 텍스트 요약(동일 분류 카운트) + `\x1b[` 부재 단언 |
| AC-TUX3-011 | REQ-TUX3-011 | `grep -n '@MX:DEBT\|@MX:CEILING\|@MX:UPGRADE' internal/merge/confirm.go \| wc -l` + `grep -rn 'charm.land/bubbles/v2' internal/merge --include='*.go' \| grep -v '_test.go'` + `go test ./internal/merge/ -count=1` | grep 태그 0 + bubbles v2 list import·사용 ≥ 1 + confirm 스위트 PASS (M1 이연 부채 해소의 도달 가능 증명) |
| AC-TUX3-012 | REQ-TUX3-012 | `go test ./internal/cli/update/... -run 'OutcomeCard' -count=1 -v` | PASS — Already up-to-date / Updated N / Dry-run 3-outcome이 단일 렌더러 경유(파라미터화) 단언 |
| AC-TUX3-013 | REQ-TUX3-013 | `go test ./internal/cli/update/... -run 'BackupPathAlways\|RecoveryCommand' -count=1 -v` | PASS — 백업 생성된 모든 outcome에서 백업 경로 + 복구 커맨드 문자열 존재 단언 |
| AC-TUX3-014 | REQ-TUX3-014 | `go test ./internal/cli/update/... -run 'PreservedLabel' -count=1 -v` | PASS — TUI 테이블·텍스트 폴백 양쪽 출력에 `preserved (user-owned)` 라벨 존재 |
| AC-TUX3-015 | REQ-TUX3-015 | AC-TUX3-005의 가드 5계열 green + `go test ./internal/cli/... -run 'UserAssetPreservation' -count=1 -v` | 전부 PASS — 종전 보존 플래그 조합에서 user-owned 파일 삭제/덮어쓰기 0건 |
| AC-TUX3-016 | REQ-TUX3-016 | `go test ./internal/cli/update/... -run 'PreviewEnforcementCoherence' -count=1 -v` (신규 통합 테스트 실존 grep 병행) | PASS — `preserve (user-owned)` 분류 파일 전부가 배포 후 byte-identical (분류⇔집행 정합) |
| AC-TUX3-017 | REQ-TUX3-017 | `go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2` + `grep 'github.com/charmbracelet/bubbletea v1' go.mod \| grep -vc '// indirect'` + `go mod why github.com/charmbracelet/bubbletea` | v2 2모듈 resolve; go.mod의 bubbletea v1 **비-indirect(direct) 행 0**(`grep -vc` = 0) — huh v1(§C 범위 외 잔존)의 transitive로 인한 `// indirect` 잔존은 허용(완전 소거는 불가, plan §B #3); `go mod why` 출력에 first-party import 체인 부재. first-party 소비자 잔존 시 progress.md §E.2에 소비자·사유 문서화(문서화 존재가 대체 PASS 조건) |
| AC-TUX3-018 | REQ-TUX3-018 | `go test -cover ./internal/cli/update/...` + `go test -cover ./internal/cli/ ./internal/template/ ./internal/hook/...` | 서브패키지 각 ≥ 85.0% + cli/template/hook가 pre-flight #9 baseline 이상 |
| AC-TUX3-019 | REQ-TUX3-019 | `go test ./... -count=1 && golangci-lint run --timeout=5m && GOOS=windows GOARCH=amd64 go build ./... && GOOS=linux GOARCH=amd64 go build ./...` | 전체 green + lint NEW 0 + 3-타깃 빌드 exit 0 |
| AC-TUX3-020 | REQ-TUX3-020 | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | pre-flight baseline **미만** — 실측치 progress.md §E.2 기록 |

## §C Edge Cases

- **분류 배타성**: 한 파일이 user-owned이면서 conflict일 수 없다 — 보호 predicate가 우선하며 분류는 상호 배타(REQ-TUX3-001 "exactly one" 단언 테스트).
- **빈 변경 집합**: 변경 0건이면 프리뷰 테이블 자체를 생략하고 Already up-to-date 카드로 직행.
- **거대 diff**: 파일 diff가 수천 라인이어도 viewport는 지연 렌더 — 전체 diff 문자열 사전 구축으로 인한 메모리 폭증 금지.
- **Windows non-TTY stdin**: confirm.go의 기존 Windows 폴백(confirm.go:858 주석 경로)이 bubbles v2 list 승격 후에도 보존 — Windows characterization 선행.
- **백업 실패 시 outcome 카드**: 백업이 생성되지 않은 실패 경로에서는 복구 커맨드 대신 백업 부재를 명시(허위 복구 안내 금지).
- **`--yes` + TTY 조합**: TTY라도 `--yes`면 프리뷰 TUI 생략 — CI가 아닌 로컬 자동화 시나리오.
- **dry-run + user-owned**: dry-run 프리뷰에도 `preserved (user-owned)` 분류가 동일하게 표기(집행 없이도 분류 모델 일관).

## §D.5 Quality Gate / Definition of Done

- 19 AC PASS (AC-TUX3-018 deploy/backup coverage ≥85% 완료 via amendment re-close) + AC-TUX3-020 (Printer migration)은 별도 SPEC으로 split (debt) — amendment re-close 기준. verbatim 명령 출력은 progress.md §E.2에 인용 (vci §3 5-section 형식).
- M3a~M3f 마일스톤별 커밋 분리; 분해(M3d)는 모듈별 이동 커밋으로 세분 + 각 커밋 후 가드 테스트 green 기록.
- namespace 가드 5계열 assertion 무수정 증명(diff 인용) — 본 SPEC의 HARD 게이트.
- 신규 `internal/cli/update/*` 커버리지 각 ≥ 85%, cli/template/hook 비회귀, ratchet baseline 미만.
- confirm.go @MX:DEBT/@MX:CEILING/@MX:UPGRADE 태그 제거 + bubbles v2 list 실사용 증명.
- frontmatter `status` 전이는 소유 매트릭스 준수 (draft→in-progress는 manager-develop).
