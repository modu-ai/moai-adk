# Sync-Audit Verdict — SPEC-REMOVAL-GUARD-EXTRAS-001 (card t511)

> 감사 좌표: 워크트리 `.claude/worktrees/t511` · 브랜치 `WT-danger-guard-regex` · HEAD `6028c21aa`
> 측정 트리: 본 브랜치 HEAD + 베이스 `0b1e27877` (git grep 트리 고정 재판독) · 감사 일시 2026-09-07
> 감사 모드: 단일 모델 (`.moai/config/sections/llm.yaml` 부재 — `audit_model` 설정 없음, multi 경로 미기동)

## 최종 판정: **PASS** — harmonic mean 0.92 (threshold 이상, must-pass 실패 0)

## 1. Claim (주장)

SPEC 본체 주장: 잔여 결함 정규식 `rm\s+-rf\s+/[^.]` 가 정확히 3개 security.yaml 사본(템플릿/독그푸드/testdata)에서 각 1행 삭제로 제거됐고(`numstat 0 1` × 3), 구조 체크 `dangerousRemovalTarget`(착지 전부터 존재, 본 브랜치 diff 0)이 차단을 단독으로 담당하며, `internal/hook/dangerous_removal_test.go`의 배포 policy 양방향 회귀 표면(기존 4 + 신규 6)이 두 뮤턴트 A/B에 대해 RED를 관측하고 복원됐다.

## 2. Evidence (증거 — 본 감사가 직접 실행·관측)

### 2.1 양방향 스윕 재실행 (AC-001~007, AC-011)

- 명령: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1 -v`
- verbatim (요지): 10테스트 전부 `--- PASS` (`FlagOrderCannotBypass`, `QuotedDataIsNotACommand`, `OrdinaryCleanupIsNotProtected`, `StillBlocksAfterQuoteFolding` + `Deployed_Allows` 4종 + `Deployed_Denies` 2종), `ok github.com/modu-ai/moai-adk/internal/hook 0.679s`, exit 0
- **스윕 수 10 (기존 4 + 신규 6) — 셀렉터 0매치 아님** (AC-RGE-011 충족). 런 페이즈 증거 파일 `green-ac-matrix-sweep.txt`(0.773s)와 테스트 구성 동일.

### 2.2 AC-008 스코프 grep — 3 → 0 대조 재현

- HEAD에서 정준 형태: `grep -rn 'rm\\s+-rf' internal/ .moai/config/` → **무출력, exit 1 (0매치)**
- 베이스 통제군: `git grep -n 'rm\\s+-rf' 0b1e27877 -- internal/ .moai/config/` → **정확히 3매치**, 경로·행번호가 [LEDGER-GREP3] verbatim과 바이트 일치 (`internal/settings/testdata/sections/security.yaml:7`, `internal/template/templates/.moai/config/sections/security.yaml:12`, `.moai/config/sections/security.yaml:7`)
- docs-site: `grep -rnF 'rm\s+-rf' docs-site/content/` → 0매치 (exit 1); 베이스 `git grep`에서 4로케일 각 1행(131행) 존재 확인 → sync 커밋에서 4행 전부 제거됨
- 도구 유의: 단일 백슬래시 `'rm\s+-rf'`(BRE `\s`=공백·`+`=리터럴)은 리터럴 텍스트와 매치하지 않는 프로브 실측(`.moai/cache/t511-audit-grep-probe.txt` 5행 판별) — 원장이 기록한 이중 백슬래시 형태가 유효한 도구이며, 감사자의 초기 단일 백슬래시 실행은 오기였다(원장 결함 아님)

### 2.3 AC-010 numstat + M1 diff 내용

- `git show --numstat 85dd4a718` → yaml 3사본 각각 **`0 1`**, `internal/template/catalog.yaml` **`4 4`** (같은 커밋), `spec.md` `1 1` (frontmatter status 행만)
- M1 diff 본문: 3사본 모두 `- rm\s+-rf\s+/[^.]` 1행만 삭제, 인접 extras(curl/wget/chmod/mkfs/dd)·주석 전부 무변경, 추가 행 0 (중립성 위반 주석 부재)

### 2.4 원장 셀 대조

- `red-mutant-a.txt`: [LEDGER-MUT-A] verbatim과 일치 — Allows 3종 FAIL + `AllowsQuotedDataMention`만 PASS(quote folding 예상 동작), `0.650s`
- `red-mutant-b.txt`: [LEDGER-MUT-B] verbatim과 일치 — Denies 2종 FAIL, 11형태 전부 `decision = ""`, `0.688s` → M1 후 차단이 구조 체크에 전적으로 의존함이 입증
- `red-allow-direction-prefix.txt`: 사전 수정 RED(`5629d9448` + 미커밋 테스트) — Allows 3종 FAIL + Denies 2종 PASS = "구조 체크는 이미 차단하고 extras 정규식만 오탐"이라는 사전 상태의 정확한 스냅숏
- 감사자 자체 스윕이 `deployedPolicy` 경로(임베디드 FS → `LoadExtraSecurityConfig` → `MergeExtraPatterns`)로 통과 → AC-009(임베디드 반영) 자체 증명 + 뮤턴트 A가 그 판별 민감도를 담보

### 2.5 스코프·닫기 무결성

- `git diff 0b1e27877..HEAD --stat`: 주장한 파일만 — 테스트 +187, yaml 3사본 -1씩, docs-site 4로케일 -1씩, catalog 4/4, CHANGELOG +1, SPEC 아티팩트·증거 파일
- `git diff 0b1e27877..HEAD -- internal/hook/pre_tool.go internal/hook/dangerous_removal.go internal/hook/branch_guard.go` → **무출력 (바이트 동일 보존)**
- `git diff 1c0ddf2fb..HEAD` → `reporter-replies.md` +26, `progress.md` §E.4 backfill 2행 — **`.go` 변경 0**
- `grep -c 'SPEC-REMOVAL-GUARD-EXTRAS-001' CHANGELOG.md` → **1** (정확히 1회, `[Unreleased]` § Fixed 내, 362행)
- spec.md: M1 `draft→in-progress`, sync `in-progress→completed` — 전부 frontmatter status 1행, 본문 무변경

### 2.6 품질 게이트 (본 감사 직접 실측)

- `go test -cover ./internal/hook/ -count=1` → **`coverage: 85.3% of statements`** (42.2s) — 패키지 목표 85% 충족
- `gofmt -l internal/hook/` → 무출력 · `go vet ./internal/hook/` → clean · `golangci-lint run ./internal/hook/... ./internal/settings/...` → **`0 issues.`**
- E4 정준 형태 재판정: `grep -rn 'AskUserQuestion' internal/hook/ | grep -v _test.go | grep -v '// '` → **1매치 (`pre_tool.go:647`**, `if input.ToolName == "AskUserQuestion" {` 도구명 비교 관측 분기, 사전 존재) — 러인 diff 신규 도입 0

## 3. Baseline-attribution (귀속)

모든 수치는 본 감사 세션이 이 트리(HEAD `6028c21aa`)에서 직접 실행한 명령의 관측값이다. 베이스 트리 비교는 커밋 `0b1e27877` 고정 `git grep`으로 재판독했다(이동 브랜치명 아님). 런 페이즈의 RED 증거 3건(`5629d9448`+미커밋 테스트, `34794215f`+뮤턴트)은 본 감사에서 재실행하지 않았고 — 원장 셀과 증거 파일의 바이트 대조 + 그 행위의 산물인 현재 트리 GREEN을 독립 재측정하는 방식으로 검증했다.

## 4. 4-차원 채점

| 차원 | 점수 | 근거 요지 |
|------|------|----------|
| Functionality | 0.95 | 11 AC 전부 독립 재관측 (스윕 10/10, grep 3→0 양 트리, numstat 3×`0 1`, 커버리지 85.3%) |
| Security | 0.92 | 구조 체크가 상위 절대경로(`/usr` 등 단일 세그먼트, `dangerous_removal.go:227` + bare/home/basename 집합)를 차단 — 제거된 정규식의 잔여 보호 역할 전면 대체. flag 순서·quote 무관 차단 테스트 + `structuralDenyPrefix`("removal of protected path", `pre_tool.go:948` 서식 실측)로 판정자 고정. 뮤턴트 B가 "구조 체크가 유일한 차단자"임을, 뮤턴트 A가 "테스트가 extras를 실제 로드"함을 각각 입증. 깊은 절대경로 허용화는 #1658 제2방향의 의도된 이완이며 heredoc 본문 보호대상 과차단은 spec.md §F 문서화 한계로 특성화됨 |
| Craft | 0.90 | 양방향 테스트 + 판정자 고정 단언(공허 초록 구조적으로 차단), fixture에 패턴 리터럴 착지 0(`grep -F` 실측 clean), env-scrub 단일 호출 규율, 증거 파일-원장 바이트 일치. 감점: catalog 표기(F1), 선택 뮤턴트 미실행(F2) |
| Consistency | 0.90 | acceptance/progress/CHANGELOG/커밋 체인/reply 초안이 하나의 이야기를 하고 수치 교차 검증 일치(10, 3→0, `0 1`×3, 85.3%). 언어 규율 준수(SPEC 문서 ko, 코드 주석 en, CHANGELOG en). D3 backfill·frontmatter-only 전환·sync 후 `.go` 0. 감점: F1 |
| **Harmonic mean** | **0.92** | 4 ÷ (1/0.95 + 1/0.92 + 1/0.90 + 1/0.90) |

## 5. Findings (심각도별)

- **Major**: 0건
- **Minor 2건**:
  - **F1 (Consistency)** — AC-010·CHANGELOG의 "catalog.yaml hash mirror rides the same commit" 표현은 catalog가 security.yaml 변경을 미러링하는 것처럼 읽히나, **catalog.yaml에는 security.yaml 항목이 아예 없고**(grep 0), 4/4 변경은 무관한 agent 템플릿 4개(manager-develop·manager-lead·manager-design·e2e-tester)의 **사전 존재 해시 드리프트 수복**이다. 실측: 신규 해시 4개 = 현재 템플릿 파일 sha256 정확 일치, 베이스 트리 파일 해시는 구 catalog 기록과 어긋남(예: manager-design 베이스 실측 `3863fa6a…` vs 구 기록 `fb7af236…`), 본 범위에서 템플릿 파일 내용 변경 0. 실질 무해(make build 재생성 부산물, 커밋 메시지에 부산물로 기재) — 표기만 미도달.
  - **F2 (Craft, 공지된 Gap)** — AC-010의 판별 뮤턴트(타 extras 라인 변경 시 numstat 이탈) 미실행. acceptance.md가 스스로 '선택'으로 표기. numstat `0 1` 측정 자체는 3사본 모두 정확.
- **Info**: E3 사전 커버리지 기준선 부재(absolute 85.3%로 대체 — progress.md §E.2가 스스로 Gap 기재), `internal/hook/mx/complexity` 83.6% 사전 결함 미접촉

## 6. Gaps (미관측)

- 런 페이즈 뮤턴트 A/B 실행 행위와 라이브 세션 RED(R1-R4)는 본 감사가 재실행하지 않았다 — 트리 고정 원장 셀 + 증거 파일 바이트 대조 + 현재 트리 GREEN 독립 재측정으로 간접 검증했다. (직접 재현을 위해 변이 커밋을 만드는 것은 감사 권한 밖이다.)
- catalog.yaml 전체(모든 항목)의 `make build` 멱등성은 검증하지 않았다 — 본 SPEC 범위는 4개 변경 항목의 정합성이다.
- `go build ./...` 및 `GOOS=windows` 크로스빌드는 재실행하지 않았다(러인 E2 기록에 귀속; 본 감사 범위는 changed-package 테스트·린트·커버리지).
- docs-site hugo 재빌드 결과는 검증 대상 아님(본 변경은 콘텐츠 1행/로케일 삭제뿐).

## 7. Residual-risk (잔여 위험)

**가장 무거운 잔여 위험 — 배포 지연 구간**: 수리는 이 브랜치의 임베드 템플릿과 테스트에 착지했지만, **리드의 develop 일괄 push + 바이너리 재설치 전까지 라이브 세션의 PreToolUse 가드는 여전히 구 정규식을 로드해 #1658/#1686 오탐을 재생산한다.** 이는 본 트리의 결함이 아니라 배포 순서의 구조적 간극이며, 감사 중 라이브 거부를 만나면 그것은 수리 전 배포 policy의 증거다. 부차 위험: 깊은 절대경로(예: `/usr/share/…`, 2세그먼트 이상)에 대한 실제 `rm`은 이제 허용된다 — 구 정규식이 거부하던 행동의 의도된 이완이고 표적은 scratch 경로이지만, 최상위 1세그먼트 아래 시스템 하위 경로에 대한 실행형 삭제는 텍스트 가드의 어떤 층도 받지 않는다(구조 체크의 보호 집합은 root/home/top-level/.git/node_modules로 한정). 이는 spec.md §F가 문서화 한계로 기재한 설계 판정이다.

---

판정자: sync-auditor (독립 감사 — 코드·문서·템플릿 편집 0, 커밋 0)
