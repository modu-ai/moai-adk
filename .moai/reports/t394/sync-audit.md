# Sync-Phase Audit — SPEC-TODO-ARCHIVE-QUERY-001

- **Auditor**: sync-auditor (독립 판정, 카드 t394)
- **측정 트리**: `.claude/worktrees/t394` @ `ffab66fba` (브랜치 `WT-todo-done-history`, origin/develop 기준 11커밋, 미푸시 — 통합은 별도 창)
- **측정 시점**: 2026-09-01, 이 실행(audit run)에서 직접 관측. 별도 표기 없는 한 모든 수치는 이 트리·이 실행의 것이다.

## Overall Verdict: **PASS** — integration-ready

## Dimension Scores

| Dimension | Score | Verdict | Evidence (발췌 — 전문은 §증거) |
|-----------|-------|---------|----------|
| Functionality (40%) | 97/100 | PASS | AC 스위트 15/15 `--- PASS`(swept count 확인, 0 FAIL/SKIP); clause 1 출처 블록 7/7 관측 PASS; 뮤테이션 4종 전부 가드 발동 확인 |
| Security (25%) | 98/100 | PASS | 저장소 바이트동일성(SHA-256 전수) PASS; 사용자 데이터의 SQL 결합 없음(vouch 프로브는 `sqlite_master` 카운트만); `--limit` 음수 거부; 프롬프트 없음(뮤테이션으로 재확인) |
| Craft (20%) | 88/100 | PASS | 신규 코드 커버리지: `newTodoHistoryCmd` 100% / `runTodoHistory` 75% / lookup 92.3% / listing 86.7% / vouch 100% / `archiveTablesPresent` 86.7% — §E.3 기록과 정확히 일치(독립 재측정). 패키지 baseline cli 80.0%·kanban 86.5% |
| Consistency (15%) | 93/100 | PASS | 오류 경로가 `todo_why.go:29`와 동일 패턴; `normalizeTodoRef` 공유(REQ-TAQ-005 기계적 동등); lint `0 issues.`, vet clean; verb-surface 가드 선언(`permittedVerbAdditions`) 존재 |

- **Weighted**: 0.4×97 + 0.25×98 + 0.2×88 + 0.15×93 = **0.95**
- **Must-pass firewall**: Functionality + Security 모두 통과 → 강제 FAIL 없음.

## 검증 근거 (Claim / Evidence / Baseline-attribution)

### 1. AC-TAQ-011 clause 1 — 출처·무결성 (내 실행, 7/7 PASS)

```
git log --diff-filter=A -- ...live-readers/  → e7eec612213b1a63...  (1건)
git rev-parse HEAD                           → ffab66fba1295e38...  (C ≠ HEAD)
git merge-base --is-ancestor e7eec6122 HEAD  → PASS (exit 0)
git cat-file -e e7eec6122:internal/cli/todo_history.go → exit 128 (C 트리에 동사 부재)
git grep -q newTodoHistoryCmd e7eec6122 -- internal/cli/ → exit 1 (기호 부재)
git diff --exit-code e7eec6122 -- goldens    → PASS (바이트 무이동)
```

스코프 준수: clause 1은 카드 브랜치(사전 통합) 소관이며 내 검증도 그 스코프에서 실행. plan.md §D의 병합-방식 의존성 기록과 일치.

### 2. AC 스위트 — swept count 확인 (빈 스윕 아님)

```
go test ./internal/cli/ -run 'TestTodoHistory|TestLiveReadersUnchanged|TestTodoSkillDocuments' -count=1 -v
  → --- PASS 15건, --- FAIL/SKIP 0건 (ok 19.665s)
go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1 -v → PASS
```

15개 AC 결정 테스트 전부 존재·실행·통과. acceptance.md §D 매트릭스와 1:1.

### 3. 뮤테이션 — 가드가 실제로 무는지 (throwaway 사본 /tmp/t394-audit, 이후 사본은 /tmp 잔존 — 샌드박스가 rm -rf 차단, 트리 영향 없음)

| 뮤테이션 | 관측 |
|---|---|
| M1: golden 1바이트 변조 (`t1`→`tX`) | `TestLiveReadersUnchangedByHistoryVerb` RED — got/want 전문 diff로 물림 |
| M2: absent 한정자를 아카이브 공백 키로 (`len(rec.Archived)==0 &&`) | `TestTodoHistoryDisclosesPreArchiveQueue` RED — 정확히 post-done 절: "the qualifier went silent once the archive was non-empty" |
| M3: vouch 프로브를 LoadPure 뒤로 이동 | `TestTodoHistoryDegradesWithoutArchiveTables` RED — dropped-tables 양쪽 단언(stderr 공백). 프로브 선행 주장이 실제 하중 지는 것 확인 |
| M4: `bufio.NewReader(os.Stdin)` 심기 | `TestTodoHistoryNeverPrompts` RED — `todo_history.go:25 carries a prompting call` 라인 지목. 되돌린 뒤 0매치 |

모든 뮤테이션은 사본에서만, 실행 뒤 되돌림. 추적 트리 무변경( `git status --porcelain` 비-tracked 0행).

### 4. 커버리지 — §E.3 기록 대조 (독립 재측정)

```
go test -coverprofile ... ./internal/cli/    → 80.0% (total 80.2%)
go test -coverprofile ... ./internal/kanban/ → 86.5%
newTodoHistoryCmd 100% / runTodoHistory 75% / renderTodoHistoryLookup 92.3% /
renderTodoHistoryListing 86.7% / InspectBacklogArchiveVouch 100% / archiveTablesPresent 86.7%
```

§E.3 `coverage_note`의 수치와 전 항목 일치. 미커버 라인은 io-쓰기-오류 반환뿐이라는 서술도 코드 대조상 정확.

### 5. 빌드·정적·문서

- `go build ./...` OK, `GOOS=windows GOARCH=amd64 go build ./...` OK
- `go vet ./internal/cli/... ./internal/kanban/...` clean
- `golangci-lint run ./internal/cli/... ./internal/kanban/...` → `0 issues.`
- 양면 문서: live + 미러 모두 51행에 `moai todo history` 행; 미러 중립성 grep( SPEC-ID/REQ/날짜/SHA ) 0히트
- `catalog.yaml` 이 브랜치에서 미변경(§E.2 M5 "by design" 서술과 일치 — 스킬 해시는 SKILL.md만 커버)

### 6. sync 산출물·전이 정당성

- sync 커밋 `973832f94`: 3파일 — progress.md(+21, §E.4), spec.md(1행: `status: in-progress`→`completed`), CHANGELOG(+2). manager-docs 소관 경계(frontmatter status만) 준수.
- 백필 커밋 `ffab66fba`: progress.md §E.4의 `sync_commit_sha` pending-backfill → 실측 SHA (D3 패턴, 자기참조 물리 한계).
- CHANGELOG 전문 대조: 16→17 동사, 기본 20 바운드, stderr 유실 카운트, absent 한정자의 last_seq 키잉, 프로브 선행(DDL 마스킹 방지), 명시적 빈 줄, LoadPure 무잠금·무마이그레이션, 6 golden 재생 — 전 항목이 측정된 행동과 일치. 사용자 오해 소지 서술 없음.
- 15/15 전이 근거: §E.3 `ac_pass_count: 15` + 내 스위트 실행이 상호 확인.
- verb-surface zero-delta 가드: `permittedVerbAdditions`에 `history` SPEC 인용과 함께 선언(커밋 `05cddca2b`) — 가드의 선언-계약 관습 준수.

### 7. 운영자 큐 (읽기 전용, plain sqlite3)

`last_seq=422 / items=109 / archived=24` — 스냅샷 C(408/108/11) 대비 이동한 live counter로, SPEC §A.4가 예고한 대로(수치 비부담). 내 테스트 실행은 전부 `todoFixture`(CLAUDE_PROJECT_DIR 고정)를 경유해 오염 없음. t414-t420 수리 기록은 판독했으나 행 단위 재검증은 하지 않음(Gaps).

## Gaps (명시적으로 관측하지 않은 것)

1. **AC-TAQ-011 clause 2 (C에서의 재생성)** — §E.2에 절차·결과(IDENTICAL cmp ×5)가 구체적으로 기록됐으나 재실행하지 않음(detached worktree + C 빌드 필요). clause 1+3이 공동으로 구속하고 M1 뮤테이션이 clause 3의 민감도를 증명했으므로 위험 낮음.
2. **CI 판정** — 브랜치 미푸시가 설계대로라 전 스위트 판정은 통합 후 origin/develop CI의 몫.
3. **t414-t420 수리의 행 단위 확인** — 기록 판독만, 아카이브 행 직접 조회는 생략.

## Residual-risk

- **clause 3은 develop 드리프트에 RED** (acceptance.md §D.3이 예고): 통합 후 develop의 기본 읽기 변경이 무관한 변경이어도 RED. 재생성 전 golden↔develop diff 필수 — acceptance.md가 절차를 못박아 둠.
- **F1 (아래)**: REQ-TAQ-010의 문자적 보편성은 degraded DB에서는 성립하지 않음 — 기록된 의도적 단순화.

## Findings (structured defect-list)

- F1 [MINOR] [optional] `internal/cli/todo_history.go:86` — REQ-TAQ-010의 "모든 바이트 불변"은 dropped-tables DB에서는 성립하지 않는다: 첫 읽기의 엔진 오픈이 DDL을 돌려 아카이브 테이블을 재생성한다(저장소 변이). `list`/`next`/`why`가 공유하는 LoadPure의 보편 오픈 동작으로, 본 변경이 도입한 것이 아니며 progress.md §E.2 M2가 문서화했다. AC-TAQ-010의 픽스처는 정상 큐라 AC는 통과. - Required fix: 없음 (기록된 한계로 수용). 경질 시 읽기 전용 오픈 모드가 필요 — SPEC 범위 밖.
- F2 [MINOR] [optional] `internal/cli/todo_history.go:112-117` — 조회 id와 함께 준 `--limit`은 조용히 무시되고, 음수 `--limit` 거부는 listing 경로에서만 (`history t1 --limit -5` 성공). AC 밖 인체공학 가장자리. - Required fix: 없음, 또는 flag+id 조합 거부.
- F3 [MINOR] [optional] `internal/cli/todo_history.go:88-91` — 로드 오류 경로가 `Error: ...`를 stderr에 인쇄한 뒤 err를 반환해 cobra가 한 번 더 인쇄(이중 출력). `todo_why.go:29`와 정확히 같은 기존 패턴 — 파일 관습 부합으로 일관성 차원에서는 옳음. - Required fix: 없음 (패키지 차원의 기존 관습; 본 SPEC이 고칠 것이 아님).

**전 결함 optional-class — must-pass firewall 발동 없음. All-optional findings 목록은 PASS를 FAIL로 바꾸지 않는다.**

## Recommendations

1. 통합 창에서 병합 후 origin/develop CI를 최종 판정면으로 (로컬 스코프 검증은 조기 신호).
2. develop 재생성 시 acceptance.md §D.3의 golden↔develop diff 우선 절차 준수 (무관한 기본 읽기 성장을 조용히 수용하는 역방향 실패 방지).
3. F1은 `@MX:DEBT` 후보 성격이나 이미 progress.md에 기록돼 있어 추가 조치 불요. F2/F3는 후속 카드 재료로 남겨 둠.

## Disposition

**integration-ready** — 수리 선결 사항 없음. 통합은 리드의 창에서.
