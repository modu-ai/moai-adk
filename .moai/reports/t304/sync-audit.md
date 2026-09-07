# sync-audit — SPEC-CODEMAPS-ACCURACY-001 (card t304)

- **판정**: **PASS-WITH-DEBT**
- **종합 점수**: **0.91** (4차원 조화평균; 0–1 척도)
- **감사 트리**: `.claude/worktrees/t304` @ `51319c589` (branch `WT-codemaps-accuracy`, sync 커밋 = HEAD)
- **감사 커밋 범위**: `061985ec8..51319c589` (흡수 `d79bdd2b3` + M1 `e29cde398` + M1 후속 `125f77825` + M2 `732995c0a` + M3 `97bff985c` + M4 `fa7b06bf0` + sync `51319c589`)
- **감사 시점**: 2026-09-02 · 독립 재측정(레인 보고 미신뢰 — 전 항목 자체 관측)

---

## §1 Claim — AC별 독립 판정 (9/9 달성)

| AC | 판정 | 근거 요지 (§2 Evidence 대응) |
|----|------|------------------------------|
| AC-CMA-001 | **PASS** | 자체 전수: 양성(비-blockquote) 인용 95 토큰 중 부재 **0**; 부재 집합 = {design, migrate, state, research, evaluator, bodp} **6토큰**, 전원 blockquote 전용 — 열거된 부정 인용 집합과 정확히 일치 |
| AC-CMA-002 | **PASS** | `### internal/factory` 0행; `### internal/kanban` 1행(modules.md:83); 병합 절이 인용하는 경로 전부 실존; 양성 bodp 0건 + blockquote 노트 1건(dependencies.md:185, `> **`internal/bodp`** —` 형식) |
| AC-CMA-003 | **PASS** | `ListActive` 0건(3지점 전부 수정 확인); 인터페이스 블록 = registry.go 실측 서명과 식별자·파라미터·반환 전부 일치(`[]Session` 0건) |
| AC-CMA-004 | **PASS** | known-5 토큰 각 1회, 전원 blockquote 경고 노트; 양성 출현 0건 |
| AC-CMA-005 | **PASS** | 테스트 3방향 초록 + 실문서 gate `citations value=0 verdict=fresh` + **감사자 자체 뮤턴트 왕복**: 주입→`value=1 stale, driving_paths=[internal/zzz-audit-phantom]`→원복→`fresh` — 공허 초록 불가 입증 |
| AC-CMA-006 | **PASS** | 양본 바이트 동일(diff rc=0); Phase 2 "NEVER as an authority for package existence" + Phase 4 실행 명령·blockquote 규약 문구 확인; 템플릿 중립성(SPEC-ID/카드-ID 0건) |
| AC-CMA-007 (SHOULD) | **PASS** | progress.md §E.2 M0 블록: 게이트 명령+출력(MERGED)·흡수 SHA `d79bdd2b3`·재측정 좌표 기록 |
| AC-CMA-008 | **PASS** | f1-record.md 존재(26 vs 27 정정 + sed 증거); t432 트리 ref `WT-codemaps-refresh` 직독 일치, 범위 내 커밋에 t432 경로 0건 |
| AC-CMA-009 | **PASS** | §E.2 AC 행렬 = 명령+관측 출력 기반; 신선도 green을 정확성 근거로 인용하지 않음(§E.4·CHANGELOG 명시) |

DoD(§D.3): 범위 한정 테스트 초록 ✔ · lint 0 ✔ · 템플릿 미러 동일 ✔ · 커밋 규약(Conventional + 카드 id + 🗿 MoAI) ✔

## §2 Evidence — 관측된 명령과 출력 (요지; 전출력은 감사 세션 기록)

**AC-CMA-001 자체 전수** (Go 구현과 독립적인 재구현 — 토큰 정규식·정리 규칙을 acceptance.md 본문에서 직접 재구현):

```
양성 라인:  grep -hvE '^[[:space:]]*>' codemaps/*.md | extract | normalize | sort -u → 95 토큰
존재검사    → POSITIVE-ABSENT 출력 0건
blockquote: grep -hE  '^[[:space:]]*>' ... → 10 토큰, 부재 6건:
  internal/bodp · internal/design · internal/evaluator · internal/migrate · internal/research · internal/state
  (전원 bq 전용 — 양성 라인에 동일 토큰 0건)
```

**AC-CMA-002/003/004 그렙**:

```
grep -n '### internal/factory' modules.md        → 0행 (exit 1)
grep -c '### internal/kanban' modules.md         → 1  (modules.md:83)
grep -c 'ListActive' data-flow.md                → 0  (exit 1)
known-5: modules.md 105/165/244/268/315 — 각 1회, 전원 `>` 접두 행
bodp  : dependencies.md:185 `> **`internal/bodp`** —` (1회, 유일 출현)
```

**registry.go 대조** (직독, `internal/session/registry.go`): receiver `Register(sessionID, specID, phase string) error` / `Heartbeat(sessionID string) error` / `Deregister(sessionID string) error` / `Query(optSpecID string) ([]Entry, error)` + 패키지 함수 `QueryActiveWork(optSpecID string) ([]Entry, error)` — data-flow.md 인터페이스 블록(352–362행)과 전부 일치.

**실문서 gate + 뮤턴트 왕복** (감사자 실행, 트리 `51319c589`):

```
go run ./cmd/moai graph check --json → exit 1 (mx-index/edges absent — 별개 신선도 축)
  citations: value=0 threshold=0 verdict=fresh
뮤턴트: modules.md 말미에 `### internal/zzz-audit-phantom` 주입(비-blockquote 양성)
  → citations: value=1 verdict=stale, reason "first cited in modules.md",
    driving_paths=["internal/zzz-audit-phantom"], stderr가 citations 레이어 지목
원복(cp 백업) → citations value=0 fresh · git status 깨끗(?? .moai/reports/t304/ 만) · diff --stat 0행
```

**테스트·커버리지·린트·빌드**:

```
go test -count=1 ./internal/graph/... → ok (graph 17.6s, symbol 0.6s)
go test -count=1 ./internal/cli/...   → ok (전 패키지)
go test -count=1 -cover ./internal/graph/ → ok, coverage: 88.9% of statements
go test -count=1 -cover ./internal/cli/   → ok, coverage: 80.3% of statements
golangci-lint run ./internal/graph/... ./internal/cli/... → 0 issues
GOOS=windows GOARCH=amd64 go build (변경 pkg + ./cmd/moai) → exit 0
```

**클로즈 계약·문서**:

```
51319c589 spec.md hunk: `status: in-progress → status: completed` (frontmatter 전용, 2줄)
fa7b06bf0(M4)의 spec.md diff: 빈(diff 없음 — progress.md만) — 소유 규칙 위반 아님
스킬 양본 diff → MIRRORS-IDENTICAL; 템플릿에서 't304|t432|SPEC-CODEMAPS|SPEC-CMR' grep → 0건(exit 1)
CHANGELOG 'SPEC-CODEMAPS-ACCURACY-001' 출현 = 1 ([Unreleased] → Added 선두)
§E.3 run_commit_sha: "fa7b06bf0" (백필 확인) · §E.4 sync_commit_sha: placeholder (D3 참조)
```

**보안 검토** (`internal/graph/check_citations.go` 전량 독후):
- 명령 주입면 없음 — 신규 코드는 stdlib만 사용(ReadDir/ReadFile/Lstat/regexp), exec 0건. check.go의 `gitOutput`은 기존 신선도 축(본 카드 미변경 부분).
- 경로 탐색: 토큰 charset에 `.`·`/` 포함 → 이론상 `internal/../../…` 형태가 projectRoot 밖을 Lstat 가능. **읽기 전용 stat 존재 오라클**일 뿐(내용 무유출·무기록), 악용에는 문서(repo) 쓰기 권한 필요 = 코드 자체에 대한 권한과 동등. MINOR(D6).
- blockquote 면제 남용(D2 위음성): spec.md D2·코드 주석·CLI help에 위험 형태가 명시되어 있고, 스킬 규약(부정 인용 blockquote 의무화)이 완화층. 현 코퍼스 관측: blockquote 부재 6토큰 전부 진짜 부정 문맥("존재하지 않음"/"제거됨"), blockquote의 실존 경로 4토큰은 무해한 주석 — 현재 남용 0건. **완화는 정직하다**: 기계적 판별자의 한계(형식 판별, 의도 추론 없음)를 숨기지 않고 선언.
- 비밀·외부 입력 검증면: 해당 없음(문서 경로 읽기 전용).

## §3 Baseline-attribution

전 항목 이번 감사 런·이 트리(`.claude/worktrees/t304` @ `51319c589`)에서 직접 실행·관측. 예외를 못박는다:
- t432 무결: `.git/worktrees/t432/HEAD` 직독 = `ref: refs/heads/WT-codemaps-refresh`(레인 기록과 동일 관측). 크로스트리 `git -C` status는 worktree-session 가드가 거부(레인이 문서화한 동일 제약) — ref 직독으로 대체, tip 불변 단언은 불가(Gap).
- M1 RED(`m1-red-evidence.md`): 트리 `061985ec8` 기준 — 단일 M1 커밋이 테스트+구현을 함께 실으므로 이력으로 재현 불가. 증상(컴파일 실패 `undefined:` — 실제 구현된 심볼명과 정확히 일치)과 트리 고정으로 신뢰성 인정(Residual-risk).
- internal/cli 커버리지 사전 기준: 본 트리 미측정. 동일일 t284 sync CHANGELOG의 "package root 80.1% is a pre-existing property" 기록을 방향성 근거로 인용(이번 실측 80.3%).

## §4 Gaps (명시적으로 관측하지 않은 것)

- `make build` 미재실행 — 임베드 재생성은 §E 기록에 의존(양본 바이트 동일성은 자체 검증). 템플릿 내용이 스킬 파일 외 변경 없음.
- 교차 모델 2차 의견(audit_multi) 미실행 — 프로젝트 설정에 `audit_model` 없음(요청 없는 Claude-only 감사 경로).
- plan.md 본문 미재감사(plan-phase 산출물 — sync 감사 범위 밖).
- §E.3에 커버리지 수치 자체가 기록돼 있지 않음(레인이 E3 항목을 생략 — D2).
- t432 tip의 시계적 불변 — 감사자에게 이전 측정이 없어 단언 불가.

## §5 Residual-risk

- **D2 위음성 잔여**(SPEC이 스스로 선언): blockquote 행에 양성 주장을 쓰면 기계 검사를 우회. 스킬 규약 + 감리 읽기가 완화층이나 기계적 폐쇄 아님 — 재생성 카드(§D.4 forward-looking)에서 관측 대상.
- `.go` 복원 휴리스틱의 오폭 여지(실제 `go`로 끝나는 경로가 우연히 `x.go` 실재와 매칭) — 오탐 방향(부재를 숨김), 현 코퍼스 무관측.
- 이 트리에서 gate exit 1은 mx-index/edges absent에서 옴 — exit code만 읽는 소비자는 citations-red와 구분 불가(stderr이 레이어를 지목하므로 운영상 지장 없음).
- 신선도 축(codemaps value=18<40 fresh)은 본 감사 범위 밖이며, 정확성 판정은 오직 citations 축·전수로 했음(REQ-CMA-010 준수 확인).

---

## §6 차원 점수

| 차원 | 점수 | 판정 | 근거 |
|------|------|------|------|
| Functionality (40%) | 96 | **PASS** (MUST) | AC 9/9 독립 재현 — 전수·실API 대조·뮤턴트 왕복 전부 자체 관측. 감점: §E.2 기록 수치 부정확(§D1) |
| Security (25%) | 92 | **PASS** (MUST) | Critical/High 0. 신규 코드 무-exec·stdlib 전용; blockquote 잔여는 정직하게 문서화; MINOR 경로 탐행면(D6) |
| Craft (20%) | 87 | PASS | graph 88.9% 커버·린트 0·테스트 품질 우수(왕복·면제·정규화표·레이어 순서·exit-1 stderr 계약); 125f77825는 완화가 아니라 픽스처 수리(주장 3→4 레이어로 강화). 감점: §E.3 E3 커버리지 기록 누락 + cli 80.3%(사전 기준치 미달, 기존 부채) |
| Consistency (15%) | 88 | PASS | CHANGELOG↔§E.2 수치 불일치(패러프레이즈 드리프트) 없음; 클로즈 계약·미러·트리 청결 전부 성립. 감점: "7토큰" 오기가 M2→§E.2→CHANGELOG로 전파, 플레이스홀더 미치환, 파일 수 11(실제 12) |

**조화평균**: 4 / (1/0.96 + 1/0.92 + 1/0.87 + 1/0.88) = **0.906 ≈ 0.91**

MUST-PASS 방화벽: Functionality·Security 모두 임계 통과 → FAIL 사유 없음.

## §7 결함 목록

| ID | 심각도 | 분류 | 위치 | 내용 | 요구 조치 |
|----|--------|------|------|------|-----------|
| D1 | SHOULD-FIX | blocking=false | progress.md §E.2(M2)·CHANGELOG.md:12·commit 732995c0a 메시지 | 부재 집합을 "7 blockquote 토큰"으로 기록 — **독립 전수 실측 6토큰**({P1–P5, P7}; 열거 자체가 6원소). AC 이항 판정에는 무영향(집합·전원-blockquote는 정확)이나 반출된 측정 수치가 거짓 | sync_commit_sha 백필 커밋에서 7→6 정정(또는 docs 후속) |
| D2 | SHOULD-FIX | blocking=false | progress.md §E.3 | E3(커버리지) 측정이 §E 자기검증에서 누락 — 명령+출력 없음. 본 감사 실측: graph 88.9% / cli 80.3% | 백필 시 수치 기록 |
| D3 | MINOR | blocking=false | progress.md §E.4 | `sync_commit_sha: "pending-backfill-<parent-of-sync-commit>"` — 템플릿 문법이 미치환(형제 카드 관례는 `pending-backfill-fa7b06bf0`형). 리드 백필 닻이 모호 | 백필 시 실측 SHA `51319c589` 기입(부모 = fa7b06bf0) |
| D4 | MINOR | blocking=false | progress.md §E.3 | `total_run_phase_files: 11` — 실제 12(Go 7 + codemaps 3 + 스킬 2; progress.md·spec.md 별도) | 기록 정정(선택) |
| D5 | MINOR | blocking=false | internal/cli (패키지) | 커버리지 80.3% < 85% 패키지 목표 — 동일일 타 카드 기록 80.1%로 보아 기존 부채(본 카드는 커버리지를 추가하는 방향). 카드 범위 밖 정당 | 별도 부채 항목으로 인계 |
| D6 | MINOR | blocking=false | internal/graph/check_citations.go:137-155 | `normalizeCitedPath`이 `..` 세그먼트를 거부하지 않음 — 문서 토큰이 projectRoot 밖 Lstat 가능(읽기 전용 존재 오라클). `validateDescribedRoot`류 어휘적 봉쇄와 불일치 | 선택 강화: `..` 세그먼트 거부 또는 봉쇄 검사 재사용 |
| D7 | MINOR(관찰) | blocking=false | progress.md §E.3 | `run_commit_sha` 백필을 manager-docs가 sync 커밋에서 수행 — D3 예외 조항 괄호는 §E.3의 소유자를 manager-develop로 적음. 값은 정확(fa7b06bf0), 기계적 SHA 완성이라 실질 위반 없음 | 기록상 관찰로 남김(조치 불요) |

## §8 권고

1. **리드 백필 창구 통합**: sync_commit_sha 백필(fa7b06bf0의 자식인 51319c589 기입) 시 D1(7→6)·D2(커버리지 수치)·D3(플레이스홀더)를 같은 커밋에서 정정 — 1회 왕복으로 폐쇄 가능.
2. D6은 후속 카드 후보(다음 citations 축 접촉 시 1줄 방어). 긴급 아님 — 악용 전제가 repo 쓰기 권한이라 실위험 낮음.
3. 재생성 replay 관측(§D.4)은 본 카드 범위 밖임이 명시돼 있으므로 후속 카드에서 스킬 규약 준수 여부를 관측할 것 — D2 위음성 잔여의 실효 검증은 그 카드의 몫.

---

감사자: sync-auditor (독립 세션, 트리 무수정 — 유일 예외: 뮤턴트 왕복의 주입/원복, 원복 후 `git diff --stat` 0행·`git status` 청결로 검증)
🗿 MoAI
