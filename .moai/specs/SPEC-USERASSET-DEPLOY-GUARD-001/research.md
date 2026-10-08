# research.md — SPEC-USERASSET-DEPLOY-GUARD-001 (증거 문서)

> 관측 방법: HEAD `db0c514d3`(브랜치 WT-user-asset-bundle)에서 원장 좌표 12파일의 해당 영역을 Read로 관측하고, 메커니즘 부재 좌표는 grep으로 생 면을 추적했다. 이 문서의 모든 코드 인용은 plan-phase(2026-10-09) 관측치다. **런 페이즈에서 테스트 실행은 없었다** — 모든 "RED 예상"은 재현 예고이지 관측이 아니다.

## §1 좌표 재고정 전문표

판정 요약은 spec.md §4. 이 표는 결정적 라인 인용을 운반한다. 라벨: `observed-at-HEAD` / `relayed-unverified` / `repaired-at-HEAD` / `coordinate-retired` / `indeterminate-at-HEAD`.

| # | 원장 좌표 | HEAD 좌표 | 라벨 | 결정적 관측 (bounded 인용) |
|---|---|---|---|---|
| 1/10-P1 | lock_guard_windows.go:23 | 동일 | observed-at-HEAD | `fd, err = os.OpenFile(path+".guard", os.O_CREATE\|os.O_EXCL\|os.O_WRONLY, 0o644)` — 마커에 PID·소유권 기록 없음, 회수 경로 없음(마감까지 재시도 후 오류 반환). 대조: `lock.go:79`의 `.lock`은 `lockOwnerGone` 사망 확인 후 rename 회수 — 같은 패키지의 비대칭 |
| 2a | doctor_user_install.go:134 | 동일 | observed-at-HEAD | `if _, err := mgr.Load(projectRoot); err != nil { check.Status = uikit.CheckOK; check.Message = "no project manifest (nothing to compare)" }` — 임의 적재 오류를 OK로 위장. .corrupt 이동·복구 사본 덮어쓰기 메커니즘은 manifest 패키지 측(미독 — relayed) |
| 2b | doctor_user_install.go:165 | 동일 | observed-at-HEAD | `for _, root := range projectCommonAssetRels {` 게이트 + `:184 check.Message = "project tree matches the lock file"` — 루트 아래 비교 파일이 0이어도 OK |
| 3 | install.go:239 | 동일 | observed-at-HEAD | `if err := WriteJournal(JournalPath(in.Home), stage); err != nil {` — stage는 227-238에서 installable만으로 구성, 회수 항목(`journal`) 반입은 :283-299(적용 뒤). 창 [239, ~299] 중단 시 소유권 증거 소실 |
| 4 | migrate_project_assets.go:89/:119 | 동일 | observed-at-HEAD | `for _, rootRel := range projectCommonAssetRels {`(:89) + `if !hasEntry \|\| entry == nil { untouched++; migrationPreservedProjectFiles[p] = true; return nil }`(:119-122) — 미등록 파일 영구 보존. `projectCommonAssetRels`(:28-33)에 `.agents/skills/` 포함 → 구 mirrorSkills 사본이 보존 대상에 들어감 |
| 5a | deployer.go:192 | 동일 | observed-at-HEAD | `if isCommonAssetRoot(path) { return nil }` — 배포 워크 스킵(원인 축) |
| 5b | deployer.go:347 (ListTemplates) | :341-366 | repaired-at-HEAD | `// card t1547 repair round, gate r4 finding 1` … `if isCommonAssetRoot(path) { return nil }` — 제외 적용됨(코멘트에 758→414 측정 기록). 재수리 금지 |
| 5c | (구) template/bundle.go:145 | 파일 부재 | coordinate-retired | `internal/template/bundle.go` 없음. 생 근접면 `internal/cli/bundle.go`(RemoveBundle 영역) — ListTemplates 행 없음 |
| 5d | "compared 0/12" | doctor_agentemit_embed.go:150-165 | repaired-at-HEAD/회귀 가드 | `compared, differing, uncompared, err := compareEmission(committed, dir)` + `if compared < len(committed) {` — 미달 CheckFail 경로가 이미 존재(라운드 3 재독 확정). 결함이 아니라 불변 — AC-013·REQ-SRF-002 가드화 |
| 5e | skills disable "Nothing to disable" | internal/cli/skills.go:54-100 | observed-at-HEAD | `defaultSkillsProjectRoot() = os.Getwd` + Long 문서: `absent 0  the project has no skill mirror, ... Nothing to act on is not an error` — 프로젝트 미러 전용 해석, 사용자 설치 스킬 미도달 |
| 6a | install.go:297 | 동일 | observed-at-HEAD | `stage.Entries = append(stage.Entries, e)` — 반입이 구 `ExpectedSHA256` 그대로, refresh 반영 없음 |
| 6b | install.go:743 | 동일 | observed-at-HEAD | `if err := os.Chmod(tmpName, 0o644); err != nil {` — confinedWrite가 전 파일 0644, .sh 실행권 상실 |
| 7a | install.go:392 | 동일 | observed-at-HEAD | `t, err = in.fileTarget(RootCodexAgents, e.Name+".toml", tomlPath, e)` — fileTarget은 raw read(`fs.ReadFile` → sha256)·무변환. "TOML이 .claude/rules 참조" 내용 주장은 relayed-unverified |
| 7b | doctor_harness.go:58 | 동일 | observed-at-HEAD | `skillsDir = filepath.Join(home, ".claude", "skills")` — 폴백이 skillsDir 통째 교체 → L1이 사용자 디렉터리 검사(프로젝트 L1 소실), L6 참조 해석 오염 |
| 8a | agentfm.go:128 | 동일(:123-132)+:282-283 | observed-at-HEAD | `agentDirsFor`가 `homeAgentsDir` 선두 포함 + `r.PostFormValue("agentfm." + a.Name + ".model")` — 이름 키 중복 행, 첫 행 우선 |
| 8b | install.go:459 | 동일(:455-468) | observed-at-HEAD | `for _, name := range selection { pack, ok := in.Catalog.Catalog.OptionalPacks[name]; ... entries = append(entries, pack.Skills..., pack.Agents...)` — `DependsOn` 전개 없음. `DependsOn` 소비자는 catalog_loader.go:70 정의뿐(설치·prune 경로 무소비 — grep 관측). 단 remove.go에 R-f-② 유예 팔(`DeferredDeps`, :115/:305) 존재 — prune 절반 변별은 재현 소관 |
| 8c | install.go:211 | 동일 | observed-at-HEAD | `if current, readErr := os.ReadFile(abs); readErr == nil {` — Lstat 선행 없음, FIFO 대상 블록 가능 |
| 8d | deployer_mode.go:37 | :33-45(45행 전체) | coordinate-retired | 파일이 M7 리타이어로 축소 — "the plugin payload constant, the mirror policy, ... are all gone". early-return 메커니즘 소멸, 생 면은 skills.go로 이동 |
| 9a | paths.go:38 | 동일(:36-43) | observed-at-HEAD(루트 정의) / relayed(집합 불포함) | 4종 사용자 루트 정의(`{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")}` 등). 보호 집합은 `internal/hook/pre_tool.go:355` `zoneLoader func(string) config.ProtectedZoneLoad`의 config 적재 — hook은 userassets 미임포트(grep 0행). 집합에 사용자 루트 없다는 주장은 relayed-unverified |
| 9b | doctor_harness.go:57 | 동일 | observed-at-HEAD | `if _, projStat := os.Stat(filepath.Join(workflowsDir)); os.IsNotExist(projStat) {` — 디렉터리 **존재**만으로 프로젝트 유지 → 빈 프로젝트 workflows가 정상 사용자 워크플로를 음영 |
| 10 | install.go:255 | 동일 | observed-at-HEAD | `completed[tgt.manifestKey] = true` (메모리 플래그) — 영속화는 :300-304 루프 종료 후 일괄. 중단 시 기설치 파일 무플래그 |
| 10a | install.go:260 (:260 reEvaluate 루프) | 동일 | indeterminate-at-HEAD | `SchemaVersion = 1`(manifest.go:25) 뿐, `LoadJournal`(journal.go:86-99)은 버전 게이트 없이 Unmarshal — "v1 잔존이 v2 충돌로 오판" 주장의 좌표를 HEAD에서 확정 불가. REQ-JRN-004(버전 게이트)로 재정식화 |
| 11a | update.go:565 | :557-575 | repaired-at-HEAD | `// Item 5 (fix round 3), relocated by the repair round (gate r5 finding): ... AFTER the confirmation gate` + `update_template_sync.go:1185`의 `migrateProjectCommonAssets(...)` 호출이 sync flow 내부(확인창 뒤). 릴레이된 결함은 HEAD에서 재현 불가 예상 |
| 11b | init.go:947/:905 | 동일 | observed-at-HEAD | `return fmt.Errorf("user-asset install failed: %w\n  Fix the cause and re-run 'moai init' ...")`(:947) — 그러나 재실행은 `strings.Contains(err.Error(), "already initialized")` 거절(:905)에 걸림. 재개 경로 부재 |
| 12 | install.go:750 | 동일 | observed-at-HEAD | 재검증 `parentResolvedFinal, err := filepath.EvalSymlinks(parent)`(:725) → `tmp CreateTemp`(:729) → `os.Rename(tmpName, dest)`(:750) — rename 직전 재검증 없음(symlink-swap 창) |
| 13a | install.go:373 | 동일 | observed-at-HEAD(경로) / relayed-unverified(내용) | `for _, slug := range []RootSlug{RootClaudeSkills, RootAgentsSkills} {` + `dirTargets` → raw WalkDir 복사 — 변환 단계 없음. "Codex 전용 프로젝트에 파일 부재" 주장은 미검증 |
| 13b | install.go:584 | applyTarget 내부 | relayed-unverified | `shipped, err := fs.ReadFile(in.Source, tgt.sourcePath)` — 원문 판독 지점. 실 복사 결정은 :392 경로. :392와 동일 결함 추정(카드 본문 "좌표 표류·재현 시 확정") |

## §2 릴레이 출처·라운드 (progress.md §Gate-Relay와 동일 원전)

- 게이트 릴레이 라운드 1-13이 t1547기 턴종료 게이트에서 누적. 오버레이 재현만 수행, 전체 CI/windows 미검증은 게이트 자체 선언.
- r8(8좌표 수렴 선언) 이후 r9-13이 잔여를 추가. r4 검증 라인(역사적 verbatim): `go test -p 2 -overlay /tmp/t1547-review-overlay.json ./internal/userassets ./internal/cli -run '^TestReview(FIFOInstall|RemovePendingBundle|InitRetryAfterInstallFailure|CodexInstalledSchemaReference)$' -count=1 -timeout 120s -v` → `--- FAIL` 4케이스 @ 고정 HEAD 40e6c310.
- 이전 세대 /tmp 오버레이 스위트는 스테일 — 본 SPEC의 재현은 본 트리에서 새로 작성한다(카드 지시).
- 16커밋 FF 흡수(81786284e→db0c514d3)가 t1547 수리 라운드 일부를 포함해, 릴레이 일부(11a, 5-ListTemplates)가 이미 수리 상태로 재고정됐다 — §1 표 참조.

## §3 구조 관측 (설계 입력)

- `harnessFS`(internal/template/harness_fs.go): 초기 관측 "배포 필터이지 변환기가 아니다"는 라운드 3 부록에서 **반박**됐다 — 같은 파일의 `NormalizeCodexRoleForDeploy`(:180)·`normalizedOpen`(:209, :271-277 배선)이 참조 변환을 수행하며 doctor도 변환 바이트와 비교한다(:297). 유지되는 관측: userassets 설치 경로가 이 변환면을 쓰지 않고 원문 복사한다는 것. 헤더 주석만으로 파일 성격을 결론 내린 초기 관측의 불완전을 정정 기록한다.
- `internal/hook` → `internal/userassets` 임포트 없음 — 동결 가드 집합 반영은 config 목록+패리티 테스트 방향(design.md §5).
- `agentDirsFor`(web/agentfm.go:123-132): 홈 디렉터리 선두 + 프로젝트 2 디렉터리 — 동명 중복 행의 구조적 원인.
- `collectEntries`와 `catalog_loader.go:70` `DependsOn []string \`yaml:"depends_on"\``: 필드는 적재되나 설치·prune 경로 소비자 없음(graph/navigator/gtd의 동명 필드는 무관).
- `remove.go`: R-f-② 의존 유예 팔 존재(RF2 주석, `DeferredDeps`) — 8b prune 절반의 현지 판정 필요.

## §4 미관측 (Gaps — 정직 목록)

- 라운드 3 재독 추가 관측: doctor_user_install.go:24-38의 checkUserInstallIntegrity는 CorruptError를 CheckWarn·기타 적재 오류를 CheckFail로 정직 보고한다 — 원장 2a의 도달 가능한 RED 앵커는 프로젝트 측 checkProjectVsLock(:134-137)이다 (AC-019a/b 분할 근거).
- 라운드 9 게이트 실측 + 소스 재독: unix `acquireGuard`는 flock(2) 기반(`lock_guard_unix.go:18-43` — `LOCK_EX|LOCK_NB` 재시도)이라 죽은 소유자의 pid-less 잔존 파일이 재획득을 막지 않는다(재획득 성공) — 원장 1의 차단 결함은 windows 마커 특유 (AC-001 플랫폼 분할 근거).
- 라운드 8 게이트 실측(릴레이 — 본 트리 미재현, 설계 반영): (i) 일시중단 생존 프로세스의 빈 마커가 연령 조건 통과 → 제2 소유자 인수 → 제1 소유자 release가 제2 소유자 마커 삭제("A release deleted B marker: true") — 연령 기반 무소유 마커 회수의 UNSAFE 증명, design.md §3 재정의 근거. (ii) 보존 가드 테스트 TestRF2F3b_MigrationPreservesUntracked·TestRF5_IdenticalUntrackedNotJournaled 존재·통과(cli/review_fix2_test.go·userassets/review_fix_test.go 관측) — 해시 일치의 소유권 증명 부적격 근거. (iii) 비windows 실행에서 `go list` IgnoredGoFiles = `[lock_guard_windows.go lock_owner_windows.go]` 본 머신 재현 — AC-023 플랫폼별 판정 분리 근거.

- 어떤 Go 테스트도 실행하지 않았다(런 M0 소관). 본 문서의 RED 예상은 전부 재현 예고다.
- `internal/cli/codex_skills_disable.go` 본문 미독 — skills.go Long 문서·해석 구조로 갈음 관측.
- manifest 패키지의 .corrupt 이동·복구 사본 경로(원장 2a의 세부) 미독 — doctor 측 위장만 관측.
- `internal/hook` 보호 집합의 config 실제 내용(config.LoadProtectedZone 목록) 미독 — 9a 집합 불포함은 relayed 상태 유지.
- FIFO 무한 블록, .sh 0644 재현, 중단 창 재현 등 런타임 행위는 전부 미관측(코드 독해 기반 정합 판정).

## §5 M0 재현 인벤토리 (런 페이즈 최초 과업)

acceptance.md §C의 25 AC가 지시하는 테스트 이름군을 본 트리에 작성하고, 각각 HEAD에서 1회 실행해 verbatim 관측한다. 판정 규율: `EXPECTED_RED`(의도 단정 실패)만 RED로 인정 — `TOOL_FAILURE`(컴파일 실패 등)와 무관 회귀는 RED가 아니다(tdd-result-contract). 이미 수리 판정 좌표(11a, 5-ListTemplates)는 회귀 가드로 전환 기록.
