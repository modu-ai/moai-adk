# SPEC-PRECOMMIT-PRESERVE-001 — sync 감사 보고서 (card t230)

- **감사자**: sync-auditor (Claude-only — 프로젝트 config에 `audit_model` 미설정, 그레이프 0히트 → 다중 백엔드 트리거 조건 미충족)
- **기준선**: worktree `.claude/worktrees/t230`, branch `WT-precommit-preserve`, HEAD `12d837b9b`, working tree clean. 통합 기준 `origin/main` (merge `6d9f584c7`). run `d5c706e25`, sync `cd90ec40b` + backfill `12d837b9b`.
- **최종 판정**: **PASS** — **95.0 / 100** (조화 평균; 산술 평균 아님)

## 차원별 점수

| 차원 | 가중치 | 점수 | 판정 | 근거 (실측) |
|---|---|---|---|---|
| Functionality | 40% | 97 | PASS | 12/12 AC를 감사자 재실행으로 확인: `go test ./internal/cli/ -run 'TestPreCommit' -count=1 -v` → exit 0, `ok github.com/modu-ai/moai-adk/internal/cli 15.354s`, 최상위 `--- PASS` **35건**, FAIL 0, **SKIP 0**. AC-PCP-013 `--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)` 실행됨. AC-PCP-005(c) `--- PASS: TestPreCommitLegacyNoRecord/c_no_record_pinned_released_body (0.19s)` 실행됨. 뮤테이트 증거 5건 표본 검사 전부 진짜 기준 RED(아래 표). |
| Security | 25% | 95 | PASS | Go 변경 전량(구현 1파일 + 호출부 2행) 열독: 신규 외부 입력 표면 없음(신규 입력은 sidecar digest뿐 — 길이 64 + hex 검증 후 사용, `readPreCommitProvenance`), 비밀·하드코딩 토큰 없음, `go.mod` 무변경(의존성 변동 0), 백업은 `O_EXCL`로 점유 경로 clobber/경쟁 차단, 경로는 `filepath.Join` + 고정 상수. OWASP Critical/High 해당 발견 0. |
| Craft | 20% | 92 | PASS | 신규 실측: `golangci-lint run --timeout=3m ./internal/cli/...` → `0 issues.` exit 0. `go tool cover -func` (신규 프로파일): `InstallPreCommitHook` 89.7% / `backupPreCommitHook` 82.4% / `installPreCommitHookOptional` 92.9% / `classifyPreCommitHook` 100% / `writePreCommitProvenance` 100% / `digestOfBytes` 100% / `readPreCommitProvenance` 77.8%. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. `moai spec lint spec.md` → `✓ No findings` rc 0. 뮤테이트 15건 기록 + 환원 규율(`git restore` 후 HEAD `d5c706e25` 유지 확인 기록). 감점 요인: F2 (malformed-record 분기 미실행). |
| Consistency | 15% | 96 | PASS | 카드 델타가 선언된 표면과 정확히 일치: `git diff --name-status origin/main HEAD` → 구현 1 + 테스트 3 + 호출부 2(각 1행) + SPEC 산출물 + CHANGELOG 1행. `git diff --numstat` → `init.go` `1 1`, `update_template_sync.go` `1 1`. 커밋 주제 Conventional Commits + card id `(t230)` 준수. sync 커밋 `cd90ec40b` markdown-only(progress.md/spec.md/CHANGELOG.md 3파일), backfill `12d837b9b`는 progress.md 1행(placeholder-exemption D3 부합). CHANGELOG 단일 항목, `[Unreleased] → Added` 최상단, 수치(12 AC, 35 테스트) 실측과 정확히 일치. |

**조화 평균** = 4 / (1/97 + 1/95 + 1/92 + 1/96) = **94.96 ≈ 95.0**

Must-pass 방화벽: Functionality(97)·Security(95) 모두 임계 통과 — 방화벽 발동 없음.

## AC별 판정 (12/12 PASS — 전 항목 감사자 재실행 + 코드 대조)

| AC | 판정 기준 명령 → 실측 | 코드 대조 | 뮤테이트 표본 | 판정 |
|---|---|---|---|---|
| AC-PCP-001 (기록 갱신) | `TestPreCommitProvenanceRecorded` PASS | 성공적 hook write **마다** `writePreCommitProvenance` 호출(`:292`), run 2에서 digest 갱신 | `mutant-b-write-once.txt`(갱신 안 함), `mutant-b2-empty-stub.txt` 기록됨 | PASS |
| AC-PCP-002 (버전범프 무음) | `TestPreCommitVersionBumpIsSilent` PASS — 파일·출력 단언 + 직접 installer 판정 감시 2번째 팔 포함 | `installed digest == recorded` → `unmodified`(incoming 무관) → 백업·통지 없음 | `mutant-a-twoway-ac002.txt` 기록됨 | PASS |
| AC-PCP-003 (백업 후 교체) | `TestPreCommitModifiedHookIsBackedUp` PASS — pre-run 바이트 단언 | 백업이 `installed` 바이트로 **write 이전**에 수행(`:269`) | 사전편집 트리 RED `found 0: []` (`m2-preedit-red-ac003-ac004ii.txt`) | PASS |
| AC-PCP-004(i) (통지 2요소) | `TestPreCommitBackupNoticeContent` PASS — 경고 라이터 **전체 출력 정확 일치** 단언 | 통지 = 백업 경로 + 교체 사실, `pre-commit.local` 미포함(`:409`) | headline silent mutant → 경고 라이터 `""` (아래 표) | PASS |
| AC-PCP-004(ii) (stderr 배선) | `TestPreCommitWarningWriterWiring` 3 서브테스트 전부 PASS (`no_other_call_sites` 포함) — go/ast 기반 | 감사자 독립 grep: `errOut := cmd.ErrOrStderr()` (`update_template_sync.go:72`), `out := cmd.OutOrStdout()` (`:69`), init `cmd.ErrOrStderr()` (`:898`), 비테스트 호출부 정확히 2곳 | 사전편집 RED: 양 호출부 `expected 4 arguments … got 3` 기록됨 | PASS |
| AC-PCP-005 (무기록 legacy a/b/c) | `TestPreCommitLegacyNoRecord` a/b/c 전부 PASS, **0 SKIP** — (c)는 `git show v3.1.2:…` fixture, 태그 실패 시 `t.Fatalf`(코드 확인) | 무기록 → installed vs incoming, 차이 있으면 user-modified(소음 방향) | `mutant-c-silent-unknown.txt`(a만 붉음), `mutant-d-body-edit.txt`(c만 붉음), skip 변형 재현(아래 표) | PASS |
| AC-PCP-006 (무음 교체 금지) | `TestPreCommitNoSilentReplacement` PASS — 한 케이스에서 두 산물 동시 단언 | 백업 + 통지가 같은 run에서 생성 | silent mutant → 백업 있음·통지 없음 RED | PASS |
| AC-PCP-007 (베어 성공줄 금지) | `TestPreCommitBackupOutputNotBareSuccess` PASS — 조항 (i)+(ii) | 경고 라이터 비었거나 경로 못 쓰면 붉음 | silent mutant → 두 조항 모두 RED | PASS |
| AC-PCP-008 (마커 없음 불변) | `TestPreCommitInstall_PreservesForeignHook` PASS (`assertNoBackup` 확장) | `!hasMarker` → `ErrUserHookExists` 조기 반환, 백업 경로 미진입(`:252`) | `mutant-m2-ac008-markerless-backup.txt` 기록됨 | PASS |
| AC-PCP-009 (백업 비덮어쓰기) | `TestPreCommitBackupNoClobber` PASS — `now` seam으로 동일 초 충돌 | `O_EXCL` + `.N` 접미 루프(`:360-383`) | `mutant-m2-ac009-fixed-name.txt` → `expected two distinct backups, found 0` | PASS |
| AC-PCP-010(a) (백업 실패 → 미교체) | `TestPreCommitSupportWriteFailureNonFatal/a_…` PASS (darwin) | 백업 실패 → `errPreCommitBackupFailed` 반환, hook write **미실행**, wrapper가 경고 | overwrite-anyway mutant → **POST-STATE 조항에서만** RED(아래 표) | PASS |
| AC-PCP-010(b) (기록 실패 → 교체 유지) | `TestPreCommitSupportWriteFailureNonFatal/b_…` PASS — run 2 자가치유 단언 포함, 크로스플랫폼 fixture(기록 경로에 디렉터리) | `lastProvenanceErr` 기록 후 정상 반환, 교체 유지 | precedence-reordering mutant → (b)만 RED, (a) PASS | PASS |
| AC-PCP-013 (상수/템플릿 정합) | `TestPreCommitTemplateMatchesConstant` **PASS, SKIP 아님** (`--- PASS … (0.00s)` 실행됨) | 본 카드는 constant·template 불변 | `mutant-m2-ac013-constant-edit.txt` → `template len: 3245, constant len: 3292` | PASS |
| AC-PCP-014 (3-way 판정) | `TestPreCommitThreeWayAttribution` 양 케이스 PASS | record 기반 판정이 incoming과 무관(`:315-318`) | `mutant-a-ttoway.txt` — case one만 RED, case two는 요행 통과(아래 표) | PASS |

(표의 AC-PCP-004/010은 조항을 함께 적어 12행 구성 유지.)

### 뮤테이트 증거 표본 검사 (요구 3건 + 2건 추가, 전부 진짜 기준 RED — 컴파일 오류·무관 실패 아님)

| 파일 | 기록된 내용 | 검증 소견 |
|---|---|---|
| `mutant-m2-silent.txt` | `TestPreCommitBackupNoticeContent`/`NoSilentReplacement`/`BackupOutputNotBareSuccess` FAIL, 경고 라이터 `""` | 카드의 headline mutant(정확히 복구 가능하게 백업하되 침묵)가 통지 단언 3곳에서 붉음 — 정확히 SPEC 설계 지점 |
| `mutant-m2-ac010a-overwrite-anyway.txt` | `POST-STATE violated: hook was replaced despite the failed backup (3245 bytes, want the unchanged 78-byte pre-run body)` — (a)만 FAIL, (b) PASS | "정상 반환 + 경고"는 유지한 채 파괴하는 변이를 오직 post-state 조항이 잡음 — AC-010(a)의 존재 이유 그대로. fixture의 chmod 분리(디렉터리 0500 + 파일 0755)로 "디렉터리 못 써서 우연히 통과"도 아님 |
| `mutant-e2-skip-reads-green.txt` | 가짜 태그에서 `--- SKIP: …/c_…` 인데 부모 `--- PASS`, `ok` (exit 0) | t.Skip 변형이 초록으로 위장하는 것 재현 — skip 검사 조항이 왜 장착됐는지 증명 |
| `mutant-a-ttoway.txt` | case one `{class:user-modified basis:record}` vs want `unmodified` FAIL, case two PASS | 2-way 설계가 단일 케이스 기준으로는 통과함을 실증 — AC-014 양-케이스 강제의 근거 |
| `mutant-m2-ac010b-skip-hook-write.txt` | `the hook must BE replaced when only the post-write provenance write fails; got 78 bytes` — (b)만 FAIL | 선행 규칙을 잘못된 쪽에 과적용한 변이를 (b)가 잡음 — (a)/(b) 비대칭의 상호 판별력 확인 |

## 스코프 불변식 (전부 실측)

- hook body: `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit | cmp - <worktree 동일 경로>` → **rc 0** (바이트 동일)
- 카드 자체 델타: `git diff origin/main HEAD -- internal/template/templates/.git_hooks/pre-commit` → empty
- origin/main 자체도 v3.1.2 대비 무드리프트 (`git diff v3.1.2 origin/main -- <경로>` empty) — §D.3 릴리스 구성 항목, 오늘 기준 통과
- 호출부: `git diff --numstat` → `1 1` / `1 1` (정확히 한 줄씩); 비테스트 호출부 전수 2곳; `errOut`/`out` 유도 grep 확인
- 카드 전체 diff: 선언된 표면(구현 1 + 테스트 3 + 호출부 2 + SPEC 4 + 보고서 3 + CHANGELOG) 외 없음

## 결함 목록 (심각도순)

- **F1** [minor] [optional] `internal/cli/hook_install_precommit_disclosure_test.go:406-411` — AC-PCP-010(a) 서브케이스가 `runtime.GOOS == "windows"` 및 root 러너에서 `t.Skip`된다. REQ-PCP-010(a)의 post-state 불변식은 POSIX CI 레그에서만 실행되며, Windows 레그에서는 이 불변식이 기계적으로 관측되지 않는다. AC-010의 Decides에는(AC-005(c)·AC-013과 달리) skip 거부 조항이 없어 기준 위반은 아니다. 완화: (b)는 크로스플랫폼 fixture, 선행 규칙 로직은 플랫폼 무관, 변이 RED는 darwin에서 관측됨. — 권장 수정(선택): Windows에서도 실패를 유도할 수 있는 fixture(예: 기록 경로를 디렉터리로 만들어 백업 디렉터리 생성 실패 유도)로 대체하거나, 해당 skip을 SPEC 잔여 위험으로 명시.
- **F2** [minor] [optional] `internal/cli/hook_install_precommit.go:331-336` — `readPreCommitProvenance`의 malformed-record 분기(길이 ≠ 64, 비-hex)가 이 테스트 패밀리에서 실행되지 않는다(함수 77.8%). §D.1 엣지 표의 "Record present but malformed → treated as absent → legacy 경로" 행이 코드로는 구현됐으나(열독 확인) 직접 테스트가 없다. — 권장 수정(선택): malformed record 2종(짧은 문자열, 64자 비-hex)을 심는 서브케이스 추가.

차단(blocking) 결함 없음 — F1·F2 모두 optional이므로 판정을 뒤집지 않는다.

## 미검증 (Gaps — 명시적)

- `go test ./internal/cli/ -count=1` 풀패키지(약 550초)는 감사자가 재실행하지 않았다. 구현 시점 기록 `ok … 552.193s`(`m2-cli-full.txt`, `@d5c706e25` 귀속)에 의존. 감사자 재실행은 필터 패밀리(35/35 PASS) + lint(0 issues) + windows 빌드(exit 0). 최종 전수 판정은 main PR의 CI 몫.
- 네이티브 `go build ./...` 재실행 안 함 — 테스트 실행이 이미 네이티브 컴파일을 증명, windows 교차 빌드만 별도 실측.
- MCP 교차 모델 백엔드(codex/glm) 미호출 — `.moai/config/` 전역 그레이프에서 `audit_model` 0히트, `multi` 트리거 조건 미충족(Claude-only 진행).

## 잔여 위험

- (c)의 v3.1.2 핀은 태그 있는 클론에서만 실행된다. fork/tarball/shallow clone에서는 `t.Fatalf`로 붉게 되는데, 이는 설계된 동작(측정 못 한 인구를 통과로 위장하지 않기 위함). CI는 `fetch-depth: 0`.
- 형제 카드 t237(#1641, OPEN — progress.md 기록)이 hook body를 바꾸면 출시 시점 §D.3 구성 점검이 붉어질 수 있다. 이는 SPEC 외부의 릴리스 체크리스트 항목으로 명시돼 있으며(acceptance §D.4), 본 감사 시점 integration ref는 무드리프트로 확인 완료.
- 백업 파일이 사용자 `.git/hooks/`에 축적될 수 있다(재발 시 매번 새 이름). SPEC은 축적 관리를 요구하지 않았고 noisy-but-safe 방향은 설계 의도이나, 장기 운영 시 사용자 동의 없는 디스크 사용 증가 가능 — 후속 카드 감성의 항목이지 본 SPEC 결함 아님.

## 근거 파일 (감사자 신규 생성)

- `.moai/state/verify/t230/audit-rerun-testprecommit.txt` — 35 PASS / 0 FAIL / 0 SKIP 원문
- `.moai/state/verify/t230/audit-rerun-lint.txt` — `0 issues.`
- `.moai/state/verify/t230/audit-cover.out` + `audit-cover-run.txt` — 신규 커버리지 프로파일(`ok … 13.798s coverage: 6.7% of statements` — 패키지 전체 수치는 필터된 패밀리 기준으로, 변경 파일의 함수별 수치가 유효 렌즈)

최종 판정: **PASS — 95.0/100** (12/12 AC, 차단 결함 0, optional 2건 기록)
