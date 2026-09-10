# t392 sync-audit 판정문 — SPEC-VERSION-STAMP-PREDICATE-001

- **Auditor**: sync-auditor (카드 t392, 리드 위임 — 독립 재측정)
- **측정 트리**: `.claude/worktrees/t392` 브랜치 `WT-version-stamp-predicate`, HEAD **`e9704b73c`** (이 판정문의 모든 측정은 별도 핀 없으면 이 SHA에 귀속된다)
- **측정일**: 2026-09-01
- **감사 대상**: run-phase M1-M5 (`3c7b55adf..009de6c09`) + sync 커밋 `9c0679098`, `e9704b73c`
- **변형 없음 증명**: 뮤턴트 배터리는 과도 편집 후 `git show HEAD:<file>` 복원으로 수행했고 매 뮤턴트 뒤 `git status --short <file>` 공복 + 최종 `git diff --stat` 공복으로 바이트 동일 복원을 확인했다. 감사가 남긴 것은 이 파일 하나다.

---

## 판정 — **PASS-WITH-DEBT** (조화평균 **0.91**)

| 차원 | 점수 | 한 줄 근거 |
|---|---|---|
| Functionality | 0.93 | 재실행한 판정 명령 전부가 AC 기대와 일치 (010 배터리 4/4, 015 배터리 3/3, §D.2 문자열 verbatim, 녹색 경로 전부) |
| Security | 0.95 | 판독 전용 검사 — argv 배열 exec(셸 없음), 이력·타 ref·네트워크 없음, 순수 코어/얇은 드라이버 분리, 주입면 없음 |
| Craft | 0.90 | 모듈 문서·상수 사유·단일 축 RED 설계가 모범적; modernize 힌트 2건은 CI 불발 확인 |
| Consistency | 0.86 | 선례·메시지 계약·B12 자기검증 준수 우수하나, 진행 기록 내부에 측정 귀속 결함 1건 (D1) |

- **D1** (SHOULD-FIX, 기록 수준): `progress.md` 최종 검증 블록의 "34(불변)" 재확인 주장이 최종 트리에서 거짓 — 실측 27 (아래 §1.6).
- **D2** (MINOR): modernize 힌트 2건 — CI lint 불발 실측 (아래 §1.7). 조치 불요.
- **D3** (MINOR, 문구): `spec.md` R-9 괄호 안 부수 판단이 미측정 낙관 — 본문 잔여 서술 자체는 정직 (§5).

차단 결함 0. 릴리스 차단 AC 여섯의 재측정 샘플(003 녹색·004 RED 재관측·005 녹색·008 녹색·012 판정 명령·015 배터리) 전부 적합.

---

## 1. 재측정 배치 (감사자 독립 실행, 트리 `e9704b73c`)

### 1.1 녹색 경로

```
$ go test ./internal/cli/ -run 'TestVersionStamp' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.103s

$ go test ./internal/cli/ -run 'TestStatus_Current|TestDoctorGolden_Light|TestDoctorGolden_Dark|TestDoctorGolden_NoColor' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.755s

$ gofmt -l internal/cli/version_stamp_registry_test.go internal/cli/status_golden_test.go internal/cli/doctor_golden_test.go
(공복 — 포맷 위반 0)

$ go vet ./internal/cli/
VET-OK (exit 0)
```

### 1.2 AC-VSP-010 출처 배터리 — **4/4 적발·통과**

파일 수준 판정(무수정, 통과 방향 = 뮤턴트 4):

```
A1 grep -c '"git"'                                  → 1   (허용 목록: git 리터럴 정확히 하나)
A2 grep -cF '[]string{"git", "ls-files"}'           → 1   (그 하나가 정확히 이 argv)
A3 awk 버전StampSweep 블록 | grep -cE '\bexec\.|os/exec' → 0   (드라이버 코어 exec 없음)
A4 awk judgeVersionStampRegistry 블록 | 같은 grep    → 0   (판정 코어 exec 없음)
```

뮤턴트 1 — argv를 `[]string{"git", "ls-tree", "-r", "HEAD"}` 로 교체:

```
A2 허용 목록 grep → 0  (rc=1 — (b) 발화, 뮤턴트 적발)
A1 git 개수 → 1       (개수만 보면 통과하는 것을 argv 정확-일치가 잡는다 — 판별식 확인)
```

뮤턴트 2 — 판정 코어에 `_ = exec.Command("git", "show", "HEAD")` 삽입:

```
A4 코어 블록 grep → 1  ((a) 발화)
A1 파일 전체 "git" → 2 ((b) 동시 발화 — run-phase 기록과 동일)
```

뮤턴트 3 — `ls-files` 호출을 판정 코어로 이동(`exec.Command(gitLsFilesArgv...)`):

```
A4 코어 블록 grep → 1  ((a) 단독 발화 — (b)는 그대로 1)
```

### 1.3 AC-VSP-015 스윕 도달 배터리 — **3/3 적발** (실트리 검사 `TestVersionStampRegistry$`)

| 뮤턴트 | `registry path missing from population` 건수 | run-phase 기록 |
|---|---|---|
| 1. 드라이버 cwd를 `internal/cli`로 좁힘 | **28** | 28 ✓ |
| 2. 제외 그룹에 `docs-site/` 한 줄 추가 | **21** | 21 ✓ |
| 3. 드라이버를 앞 50개로 절단 | **28** | 28 ✓ |

부수: 뮤턴트 1에서 `unregistered file carries` 0건 — 모집단이 좁아지면 결핍 방향만 운다(완결성 단언이 조용해지지 않고 도달 단언이 대신 우는 구조 확인).

[감사 과정 기록] 뮤턴트 2의 첫 시도에서 perl 패턴 `\Q\t"CHANGELOG.md",\E` 가 치환에 실패해 무수정 실행 → 0건이 관측됐다. 패턴을 탭 앵커 없는 형태로 고치고 치환 적용을 `grep -c` 1건으로 먼저 검증한 뒤 재실행해 21건을 얻었다. 0건은 검사 결함이 아니라 감사 도구 적용 실패였고, 기록에 남긴다.

### 1.4 §D.2 실패 문자열 판정 (AC-VSP-004, 릴리스 차단)

선명도(stamp) 판정기의 finding 방출 줄을 중화(`append` → `_ =`)한 뒤 합성 입력 테스트를 돌려 RED를 재관측:

```
$ go test ./internal/cli/ -run 'TestVersionStampSyntheticFreshness' -count=1 -v
=== RUN   TestVersionStampSyntheticFreshness/catches_a_stale_stamp
    version_stamp_registry_test.go:384: check did not emit expected failure: registered stamp does not carry the authoritative token: docs-site/hugo.toml (got 1 findings: [registered stamp missing from sweep: docs-site/hugo.toml])
--- FAIL: TestVersionStampSyntheticFreshness
    --- PASS: TestVersionStampSyntheticFreshness/exempts_a_stale_prose_entry
```

§D.2 핀 문자열 `registered stamp does not carry the authoritative token: <path>` 가 RED 실패 줄에 **그대로** 나타났다. `got 1 findings` 가 §E.2 기록의 `got 0 findings` 와 다른 이유는 명확하다: §E.2는 빈 코어(구현 전)였고 내 재관측은 완성 코어에서 스윕-상위집합 단언만 살아 있는 것이며, 같은 stale 상태에 대해 두 단언이 동반 발화한다는 것은 테스트 주석이 예고한 대로다. prose 면제 방향은 뮤턴트에서도 PASS — 양방향 유지.

### 1.5 문서 판정 — AC-VSP-011/014/007 전부 적합

```
AC-011(b) 7구절: aged-out token 1 · registered as `prose` 1 · inlined inside a file the
  exclusion set hides 1 · not a version token at all 1 · renders the version rather than
  carrying it 1 · the repository does not track 1 · this list is not exhaustive 1
AC-011(d) 닫힌-수 정규식: 실문서 0건
  양성 방향(8 뮤턴트 문장을 정규식에 투입): 8/8 적중 — 0건과 8건을 함께 봤다
AC-014 (a)1 (b)1 (c)1 (d)1 · docs-site/content 유출 0
AC-007 데이터 행 6 (`.moai/reports/`,`.moai/specs/`,`.moai/release-notes/`,`CHANGELOG.md`,
  `*_test.go`, per-locale changelog glob) · 각행 사유+감춘 수(61/62/1/1/4/0) · 트리 SHA
  `051f209b0` 표 근처 1건 · 퉁친 면제 조항 없음
한글 0건 (영문 전용 유지)
```

### 1.6 병합트리 스윕과 문서의 권위 토큰 — **27 / 0건, 사전 선언과 일치**

```
$ grep -cF 'v3.1.3' .moai/docs/version-management.md   → 0   (문서 어디에도 권위 토큰 없음 —
                                                              §E 교체 문안이 토큰을 새로 들여오지도 않음)
$ git grep -lF v3.1.3 -- . ':!.moai/reports' ':!.moai/specs' ':!.moai/release-notes' \
    ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*' | wc -l   → 27
```

27의 내역(정렬 목록 판독) = 스탬프 7 + prose 20 — 등록부 28에서 토큰을 잃은 `.moai/docs/version-management.md` (prose — 무위반) 하나만 스윕 밖. 상위집합 성립. M5 트리 `8c71cd423` 에서도 동일 명령 → **27**.

### 1.7 lint — 랜딩 변경 신규 이슈 0

```
$ golangci-lint run ./internal/cli/ --new-from-rev 051f209b0 --timeout=5m
0 issues.
```

golangci-lint v2 (활성 집합: errcheck · govet · ineffassign · staticcheck · unused, `default: none`). 두 LSP 힌트는 gopls modernize 계열 제안으로 이 집합에 없다:

- `version_stamp_registry_test.go:364` — `versionStampFindingsContain` 루프 → `slices.Contains`: **CI lint 불발 아님** (0 issues 실측).
- `version_stamp_registry_test.go:741` — `strings.Split` → `strings.SplitSeq`: go.mod `go 1.26.4` 라 적용 가능하나, 마찬가지로 활성 linter 어디에도 없어 **CI 불발 아님**.

Craft 판정: 성능 미미화 제안 수준(S3) — 루프는 명확하고, SplitSeq는 10k 경로 1회 분할이라 체감 없음. 조치 불요.

### 1.8 CHANGELOG 사실 관계 표본 검증 — 전부 적합

- 상수: `expectedRegistryEntries = 28` · `expectedStampEntries = 7` — 소스에서 직독, 등록부 리터럴 `path:` 28행·`versionStampClassStamp` 7행과 일치.
- 골든 6개 1행 변경: `git show 96bfa0c99 --stat` — 6 golden 전부 `2 +-(1+/1−)`, 테스트 파일 +22/+23. `doctor-nocolor`·`status-nocolor` diff 직독 — 정확히 버전 행 1행, doctor `ok/warn` 집계 행 불변.
- `b37e86b64`: `git log` → `b37e86b64 2026-08-24 test(cli): stamp doctor/status golden fixtures at v3.1.3` — "골든이 이미 한 번 사후 스탬프가 필요했었다"는 서술의 역사 참조로 정확.
- 제외 그룹 수치 61/62/1/1/4/0 · 129/163 · "나머지 34": 문서 표와 CHANGELOG가 일치하고, 트리 핀(`051f209b0`)이 문장 안에 명시돼 있다 — changelog-pages 그룹만 직접 재측정(0건 확인), 나머지 군수치는 재측정하지 않았다(Gaps).
- B12 자기검증 3건 재현: SPEC-ID 중복 0 · `grep -c '^### AC-VSP-' acceptance.md` = 15 · 나열 경로 전부 트리에 존재.

---

## 2. AC 매트릭스 (15/15 — 감사자 관측 근거별)

| AC | 판정 | 근거 |
|---|---|---|
| 001 | PASS | `pkg/version/version.go:8 Version = "v3.1.3"` — 단일 출처 재관측 |
| 002 | PASS | 리터럴 28항목 · exact-path 가드(모양 grep 0) · 스윕 27 ⊆ 등록부 28 · prose 상위 허용(테스트) |
| 003 | PASS | 녹색 경로(TestVersionStampRegistry ok) + §E.2의 pin 이전 6경로 RED 기록과 golden 사후 토큰 0건 |
| 004 | PASS | §1.4 RED 재관측 — §D.2 문자열 verbatim, 양방향 |
| 005 | PASS | 합성 RED 3종 녹색 경로 구현(§E.2) + 개수 상수 독립 보유 소스 확인, 스윕 개수 상수 부재 확인 |
| 006 | PASS | TestVersionStampSweepByContent 녹색 — 내용 판정, 이름 함정 표본 포함 |
| 007 | PASS | §1.5 — 표 6행·SHA 핀·군별 사유 |
| 008 | PASS | golden 토큰 0건 · pin 6함수 + defer 선례 준수 · 6경로 등록부 부재(리터럴 직독) |
| 009 | PASS | TestVersionStampRegistry 녹색(독립 문서 리터럴 대조) + 자기비교 회피 설계 소스 확인 |
| 010 | PASS | §1.2 — 배터리 4/4 |
| 011 | PASS | §1.5 — 7구절·정규식 양방향·구절 단일 줄(문서에서 전부 적중으로 간접 증명) |
| 012 | PASS | A5 walk 0건 · `gitLsFilesArgv` 단일 리터럴 · 합성 추적 양방향 녹색 |
| 013 | PASS | 유령 합성 RED(§E.2) + 녹색 경로 |
| 014 | PASS | §1.5 — 4판정 + 유출 뮤턴트 방향 |
| 015 | PASS | §1.3 — 배터리 3/3, 크기 리터럴 부재 소스 확인 |

---

## 3. 결함

**D1 — SHOULD-FIX (기록 수준, 코드 영향 0).** `progress.md` §E.2 "최종 검증 (HEAD `8c71cd423`)" 블록이 "M0 수치와 상수는 최종 트리에서 재확인: **34(불변)** · 등록부 28 · stamp 7"라고 주장한다. 실측: `git grep -lF v3.1.3 <여섯 거부 pathspec>` 이 `8c71cd423` 과 `e9704b73c` 양쪽에서 **27**이다 (M4 골든 pin −6, M5 문서 토큰 행 제거 −1). 34는 M0 트리 `051f209b0` 의 값이고, M5 단락 스스로 "M5 이후 스윕은 27"이라 정확히 적어 놓아 진행 기록 **내부에서 모순**이다. §E.4의 "스윕 34 불변" 재언급도 같은 낡은 수치다. 검사 자체는 스윕 개수 상수를 의도적으로 안 들고 있어 아무것도 깨지지 않으며, CHANGELOG는 34를 `051f209b0` 에 정확히 핀해 둔다. 이것은 verification-claim-integrity §2(귀속) 위반 형태의 **증거 기록 결함**이다 — 통합 창 전후로 progress.md 한 줄 정정 또는 errata 주기가 수리이다. 착지·병합을 막는 결함이 아니다.

**D2 — MINOR (조치 불요).** modernize 힌트 2건 (§1.7). CI 불발 실측 완료.

**D3 — MINOR (문구).** `spec.md` R-9: "그런 드라이버를 쓰려면 등록부를 드라이버 안으로 들여와야 하고, 그것은 §6이 금지한 자기 참조라 **다른 단언에 먼저 걸린다고 판단했을 뿐** 실행으로 보이지 않았다". 이 부수 판단은 의심스럽다 — 등록부 경로만 남긴 모집단은 (가)·(나)를 **모두 통과**하며, 다른 단언(유령·신선도·상위집합)도 전부 조용하다. `acceptance.md` AC-VSP-015가 정확히 그렇게 평서형으로 적어 놓았다. R-9의 겉문("손실은 남는다")은 정확하고 미관측을 "판단했을 뿐"이라 명시해 정직하나, 괄호 안 위로는 건너읽는 독자에게 거짓 안도를 줄 수 있다. 열어 둔 설계 자체는 수리 대상이 아니다 — 문구만.

---

## 4. Gaps (명시적으로 관측하지 않은 것)

- **origin CI 판정** — push된 head의 CI 완주·초록은 이 세션 밖이다(§D.2 마지막 항). 리드 몫.
- **병합 트리 전체 스위트 재측정(R-7)** — `go test ./internal/cli/... -count=1` 전체(286초급)는 재실행하지 않았다. 감사자 재실행은 영향 범위 표본(TestVersionStamp* + 6 golden 테스트 + vet + gofmt + scoped lint)이며, 전 패키지 판정은 통합 창·CI 몫(레인-로컬 검증 부하 규율).
- **제외 군수치 61/62/1/1/4(합 129/163)의 독립 재측정** — 표·CHANGELOG 일치와 changelog-pages=0만 확인했다.
- **CI lint 바이너리 v2.1.6** — 로컬 v2.10.1로 쟀다. 설정 헤더가 동일 활성 집합을 핀하고 있으나 바이너리 자체는 재현하지 않았다.
- **010 뮤턴트 2의 "첫째 grep 2건" 전체 파일 기준값** — 코어 블록 1건과 파일 전체 2건을 함께 관측해 run-phase 기록과 정합을 확인했으나, acceptance 본문이 예고한 "2건"과의 대응은 §E.2의 해석(기준값 차이)을 그대로 받아들인 것이다.

## 5. Residual-risk (관측 후에도 남을 수 있는 것)

- **R-9는 실제로 열려 있다**: 등록부 경로만 남기는 모집단은 (가)·(나)를 통과한다 — 이번 배터리로도 잡히지 않는 방향이며, SPEC이 주장하지 않은 대로다. 문안은 정직하다(괄호 안 부수 판단만 D3).
- **R-6**: `version.go`의 Version 줄 자체가 잘못 이동하면 술어 전체가 조용히 이동한다 — 선언된 잔여.
- **lint 버전 스큐**: 위 Gaps 대로.
- **감사 방법론**: 내 뮤턴트 재관측은 감사자가 직접 변형을 설계했으므로 run-phase의 것과 미묘히 다를 수 있다(015-2 첫 실패가 그 사례). 최종 수치는 변형 적용을 사전 검증한 재실행 값이다.

---

## 판정문 요약

**PASS-WITH-DEBT — 0.91** (Functionality 0.93 · Security 0.95 · Craft 0.90 · Consistency 0.86).
차단 결함 없음. 가장 영향 큰 발견 두 가지: (1) progress.md 최종 검증 블록의 "34 불변"이 최종 트리에서 거짓(실측 27, 기록 내부 모순 — D1, 기록 정정으로 수리), (2) 나머지 전부 — 배터리 4/4·3/3, §D.2 문자열, 문서 판정, CHANGELOG 사실 관계 — 감사자 독립 재실행에서 run-phase 기록과 정확히 일치했다. R-9 서술은 정직하다(괄호 부수 판단만 문구 정리 후보). 통합 창 진행을 맞을 상태다.
