# SPEC-FMT-GATE-001 — sync-audit verdict (card t465)

- **Verdict**: **PASS-WITH-DEBT**
- **Score**: **0.916** (4-dimension harmonic mean; must-pass 축 Functionality·Security 모두 PASS)
- **Auditor**: sync-auditor (독립 감사, fresh-judgment — 실행자 §E 보고를 신뢰하지 않고 전 AC 재실측)
- **Audit tree**: worktree `t465` @ `WT-format-gate-zero` HEAD `5a4383f45` (측정일 2026-09-03)
- **SPEC**: `.moai/specs/SPEC-FMT-GATE-001/` — status `completed` (sync commit `09c5b5431`에서 전환, backfill `5a4383f45`)

---

## Claim

SPEC-FMT-GATE-001의 인도물 — activation commit `a95939df5`(정확히 2 파일:
`.github/workflows/ci.yml` Lint 잡 format-gate 스텝 + `Makefile` `fmt-check` 타깃) — 은
6개 AC 전항목에 대해 이 트리에서 **감사자 자체 재실측**으로 이진 판정을 재현했다. 게이트는
공허 초록이 아니며(뮤턴트 왕복 RED→GREEN 관측), untracked 노이즈 엣지(acceptance §D.3)와
전수/tracked 두 판정형의 발산 지점도 실측으로 확인했다. CI 스텝은 구조적으로 건전하다
(무조건 실행 lint 잡·억제 패턴 부재·exit 전파·루트 워킹디렉터리). 유일하게 열려 있는 항목은
acceptance §D.5 세 번째 종결 게이트(develop push 후 CI Lint green 확인)로, 레인 프로토콜상
레인이 측정할 수 없는 외부 소관이다 — 이것이 PASS-WITH-DEBT의 부채 항목이다(결함 아님).

## Evidence — per-AC 재실측 (전부 이번 런·이 트리, 명령 + verbatim 출력)

| AC | 내가 실행한 명령 | verbatim 출력 | exit | 판정 |
|---|---|---|---|---|
| AC-FG-001 (더러운 트리 → 게이트 실패) | `printf ' ' >> pkg/version/version.go && make fmt-check` | `gofmt violations found (run gofmt -w or make fmt):` / `pkg/version/version.go` / `make: *** [fmt-check] Error 1` | **2** | **PASS** |
| AC-FG-001 리터럴형 동치 | `gofmt -l . > /tmp/…txt && test -s /tmp/…txt` (가드가 `test -z "$(gofmt -l .)"` 치환형을 거부 — progress.md §E.2 주와 동일 거부; `test -s` exit 0 ⇔ `test -z "$(...)"` exit 1 동치) | `test-s-equivalent-of-AC-FG-001-exit=0` (출력 파일 1행: `pkg/version/version.go`) | 0 | **PASS** (동치) |
| AC-FG-002 (녹색 트리 → 통과) | `make fmt-check` (복원 후) — 무출력 | (출력 없음 — silent-on-success) | **0** | **PASS** |
| AC-FG-003 (t457 선행) | `git merge-base --is-ancestor e1fdf00d1 a95939df5` | (무출력) | **0** | **PASS** |
| AC-FG-004 (활성 시점 녹색) | `git archive a95939df5 \| tar -x -C /tmp/t465-audit-act-tree && gofmt -l /tmp/t465-audit-act-tree \| wc -l` — HEAD 미이동, 활성 트리 직접 추출 측정 | `0` | 0 | **PASS** |
| AC-FG-005 (배포 표면 불변, 개정 판정식) | `git log 5a4383f45 --grep='card t465' --oneline -- internal/template/templates/ \| wc -l` + 대조군 `git log 5a4383f45 --grep='card t465' --oneline \| wc -l` | `0` / 대조군 `8` (커밋 집합 비어있지 않음 — 공허 0 아님) | — | **PASS** |
| AC-FG-006 (로컬 패리티) | clean: `make fmt-check` / dirty: AC-FG-001과 동일 변형 | clean 무출력 / dirty 파일명 출력 | **0 / 2** | **PASS** |

### 추가 실측 (AC 이외의 견고성 확인)

- **뮤턴트 왕복**: trailing space 주입 → exit 2 + 파일명 → `git checkout -- pkg/version/version.go` 복원 → `git status --porcelain \| wc -l` = `0`, `make fmt-check` exit 0. RED와 GREEN을 모두 관측 (verification-completeness §1.1 observed-failure).
- **untracked 노이즈 엣지 (acceptance §D.3)**: untracked `zz_audit_untracked_scratch.go`(고의 비정형) 존재 시 `make fmt-check` → exit **0** (tracked 변형은 판정 불변), 동시에 전수형 `gofmt -l . \| wc -l` → `1` (전수형은 뒤집힘). 두 판정형이 정확히 이 지점에서 발산함을 실측 — REQ-FG-006 tracked 변형 선택의 근거 확인. 스크래치 제거 후 트리 clean 복귀(porcelain 0행).
- **파스 불가 파일 스캔**: `git ls-files -z '*.go' \| xargs -0 gofmt -l > /dev/null; echo $?` → `0` — tracked .go 전수가 파스 가능+포맷 클린 (acceptance §D.3 "fixture 파스 불가 0건" 주장 실측 뒷받침).
- **경로 엣지**: tracked .go 경로 중 공백 포함 0건 (`git ls-files -z '*.go' \| tr '\0' '\n' \| grep -c ' '` → `0`). `-z`/`-0` NUL-구분 파이프라인 자체가 공백 경로에도 안전.
- **CI 스텝 구조 판독** (ci.yml L419-464): format-gate 스텝(L454-455, `run: make fmt-check`)은 `lint` 잡(L422, `if:` 없음 — 무조건 실행) 안에 위치. 스텝에 `|| true`·`continue-on-error`·`if: always()` 부재(워크플로 전체 grep의 해당 히트는 전부 타 잡/타 스텝). `working-directory:` 키 부재 → 루트에서 실행. GitHub Actions 기본 셸 `-e`에 따라 make의 non-zero exit(관측값 2)가 스텝→잡을 실패시킨다. 배치는 templ codegen drift guard 뒤·golangci-lint 앞 (progress.md §E.2 주장과 일치).
- **강제 표면 기계 검증**: `gh api repos/modu-ai/moai-adk/branches/main/protection --jq '.required_status_checks.contexts'` → `["Test (ubuntu-latest)","Lint","Build (linux/amd64)","Analyze (Go) (go)","Release PR Multi-OS Gate"]` — **`Lint`는 main의 required check다** (잡 `name: Lint`와 정확 일치). 워크플로 트리거 `on: push: branches: [main, develop]` + `pull_request: branches: [main]` — develop push마다 무조건 실행. 게이트의 기계적 강제 주장이 릴리스 경로(main PR)에서 필요조건 체크로, develop에서는 무조건 실행으로 각각 성립.
- **§D.5 나머지 종결 게이트**: 카드 커밋의 `.go` 수정 0건 (`git log 5a4383f45 --grep='card t465' --name-only … \| grep '\.go$' \| wc -l` → `0`); activation..tip `.go` diff 0건; activation 커밋 파일 목록 = `.github/workflows/ci.yml` + `Makefile` 정확히 2건 (리드 결정 D1 단일 activation 커밋 준수).
- **CHANGELOG**: `grep -c 'SPEC-FMT-GATE-001' CHANGELOG.md` → `1` (중복 없음). 항목 내용이 인도물과 일치(AC 6개 PASS·뮤턴트 기록·알려진 gap 명시). AC 수 교차검증: `grep -cE '^### AC-FG-' acceptance.md` → `6` (§E.3 `ac_pass_count: 6`·CHANGELOG와 일치).
- **기록 인스턴스 성립**: 체인 SHA 11종(`9e1b6a379`·`e00102f88`·`ce546a373`·`a95939df5`·`bafa7a5a3`·`3e98a90cf`·`350107589`·`09c5b5431`·`5a4383f45`·`e1fdf00d1`·`d592b0551`) 전부 `git cat-file -t` → `commit`. t191 패턴 준수 실측: sync commit이 spec.md frontmatter `status: in-progress → completed` 전환 + CHANGELOG + §E.4(`pending-backfill-sync`)를 3파일로 담당, backfill `5a4383f45`가 `sync_commit_sha: 09c5b5431` 완성.

## Baseline-attribution

위 전 측정은 **이번 감사 런, 이 트리**(worktree `t465` @ `5a4383f45`)에서 직접 실행한
명령과 그 출력이다. AC-FG-004는 HEAD를 이동시키지 않는 `git archive` 추출로 활성 커밋
트리 그 자체를 측정했고, 활성 트리와 현재 tip 사이에 `.go` 파일이 0건 변경임을 별도
실측(`git diff --name-only a95939df5..5a4383f45 -- '*.go' \| wc -l` → `0`)해 추출 측정과
현 트리 측정의 동치도 확보했다. 실행자 §E.2 보고의 수치(gofmt 0, fmt-check 0/2, 뮤턴트
파일명 출력)는 본 감사의 재실측과 전부 일치했다 — 즉 실행자 주장은 인용이 아니라 이번
재측정의 독립 재현으로 뒷받침됐다.

## Gaps (명시적으로 관측하지 않은 것)

- **CI Lint 잡 green 판정** (acceptance §D.5 세 번째 종결 게이트): 본 레인은 push하지
  않는다(레인 프로토콜 §4 — push는 리드 일괄 소관). 리드의 develop push가 일으키는 CI
  실행이 1차 실측이 된다. **이것이 PASS-WITH-DEBT의 부채 항목이다.** 로컬 증거는 불건전하지
  않다 — 오히려 본 감사가 CI 스텝 구조·필요조건 등록·워크플로 트리거를 기계 검증해,
  남은 불확실성은 ubuntu 러너에서의 실제 실행뿐이다.
- ubuntu-latest의 `make` 가용성 — 러너 이미지 문서 근거(감사 미측정). CI 1차 실행이 측정한다.
- `main`에서의 gate 적색 거동(필요조건 위반 시 병합 차단) — 브랜치 보호 설정값
  (required contexts에 `Lint` 포함)으로만 검증했고, 실제 차단 이벤트는 관측하지 않았다.

## Residual-risk

- **CI Go 버전 드리프트**: gofmt은 `go-version-file: go.mod` 툴체인에서 온다. 버전 상승이
  포맷 규칙을 바꾸면 게이트가 붉어질 수 있다 (acceptance §D.3이 이미 명시한 알려진 위험 —
  그때의 판정은 CI 몫, 로컬 회피 금지).
- **templ 생성물**: `internal/web/*_templ.go`가 templ/go 버전 상호작용으로 gofmt-dirty가
  되면 게이트가 붉어진다 — 수리는 재생성(`templ generate`)이지 게이트 완화가 아님 (SPEC §D).
- **xargs 빈 입력 엣지 (관측 only)**: tracked `.go`가 0개인 트리에서 GNU xargs(ubuntu)는
  인수 없이 `gofmt -l`을 1회 실행해 stdin을 읽는다 — CI(stdin 닫힘)에서는 exit 0 공허
  통과, 로컬 터미널에서는 블록. 본 리포(~1500 tracked .go, 파스 스캔 exit 0)에서는
  도달 불가능한 상태라 결함이 아니다(D3 참조).
- `WT-gofmt-drift` 브랜치 tip은 핀 `e1fdf00d1` 이후 `71f2930db`로 전진했다 (감사 관측).
  조상 판정은 SHA 핀을 대상으로 하므로 무해하다 — `e1fdf00d1`은 `--no-ff` 병합
  (`d41ba7479`)으로 역사에 남아 있고 실측 exit 0. 재작성/squash 시에만 REQ-FG-003의
  재고정 절차(spec.md 재고정 노트)가 개입한다.

---

## Dimension Scores (flat weighted)

| Dimension | Score | Verdict | Evidence 요지 |
|---|---|---|---|
| Functionality (40%, must-pass) | 94/100 | PASS | AC 6/6 자체 재실측 PASS·뮤턴트 RED→GREEN 관측·untracked 엣지 실측·CI 스텝 구조+필요조건(`Lint`∈main required) 기계 검증. 감점: §D.5 CI green 게이트 미측정(외부 소관 부채) |
| Security (25%, must-pass) | 96/100 | PASS | 비밀·주입 표면 부재. NUL-안전 `-z`/`-0` 파이프라인, `printf '%s'` 안전 출력, 의존성 매니페스트 불변, 템플릿 표면 0경로 |
| Craft (20%) | 90/100 | PASS | silent-on-success·stderr 파일명 출력·동치 주석·.PHONY 등록. 단순성 사다리 준수(git+gofmt+make 재사용, 신규 의존 0). observed-failure 완수(감사자 자체 뮤턴트). 커버리지 임계값은 .go 0변경이라 구조상 비적용(문서화됨) |
| Consistency (15%) | 87/100 | PASS | Conventional Commits 8/8·카드 id 8/8 주체·t191 종결 패턴 실측 준수·Makefile/ci.yml 인접 스타일 일치. 감점: 8커밋 중 6건 `🗿 MoAI` 트레일러 부재(D1), 레인에서의 SPEC 본문 fold 커밋 재위임 경로 미기록(D2) |

**Harmonic mean** = 4 / (1/0.94 + 1/0.96 + 1/0.90 + 1/0.87) = 4 / 4.366 = **0.916**

**Overall: PASS-WITH-DEBT (0.916)** — must-pass 방화벽(Functionality+Security) 통과.
부채 1건 = §D.5 세 번째 종결 게이트(CI green), 소관 리드 push 후 CI 판독.

## Findings (D1..D5 — BLOCKING 0건)

- **D1** [MINOR] [optional] `a95939df5`·`3e98a90cf`·`e00102f88` 외 — 카드 8커밋 중 6건이 `🗿 MoAI` 트레일러 없이 종결 (동봉 커밋 2건 sync·backfill만 트레일러 보유; manager-develop 위임 템플릿 Section D가 커밋 규약으로 명시하는 항목). 추적성 운반체(카드 id·SPEC-ID)는 8/8 주체에 존재하므로 종결을 막지 않는다. — Required fix: 없음(본 카드 한정). 향후 카드 커밋은 트레일러 포함.
- **D2** [MINOR] [optional] `350107589`·`ce546a373` — 레인에서 spec.md/acceptance.md **본문**을 수정하는 fold 커밋. 소유권 매트릭스는 run-phase 본문 수정을 manager-spec 재위임 경로로 규정하는데, 그 경로 이행이 커밋에 기록돼 있지 않다. 내용 자체는 plan-audit D1-D3 수리 지시와 정확 일치(본 감사 대조)라 정확성 결함은 아니다. — Required fix: 없음(본 카드 한정). 향후 fold는 재위임 경로를 커밋 body에 1줄 기록.
- **D3** [INFO] [optional] `Makefile` fmt-check — tracked `.go` 0개 트리에서 GNU xargs가 `gofmt -l`을 인수 없이 실행(CI에서 공허 통과, 로컬에서 stdin 블록). 본 리포에서 도달 불가(파스 스캔 exit 0). 결함 아님·기록 보존. — Required fix: 없음. 타 프로젝트 복사 시 `xargs -0 -r` 경화 권장.
- **D4** [INFO] [blocking=부채 항목] acceptance §D.5 세 번째 종결 게이트(CI Lint green) 미측정 — 레인은 push하지 않는다. 리드 일괄 develop push → CI 판독으로 닫는다. 본 감사가 로컬 증거의 건전성(스텝 구조·필요조건·트리거·이진 판정 전반)을 별도 검증했으므로 결함이 아니라 기록된 부채다. — Required fix: 리드 push 후 CI Lint green 판독 (card done의 전제).
- **D5** [INFO] [optional] spec.md §B — lint 잡을 "무조건 실행되는 required check"로 서술했으나, `develop` 브랜치에는 보호가 없다(`gh api` 404 "Branch not protected"). "required"는 branch-protection 의미로는 main에만 성립(실측: `Lint` ∈ main required contexts). develop에서는 무조건 실행+리드 판독으로 강제가 실질 유지되므로 위반은 아니고, 두 표면의 용어가 뒤섞인 서술. — Required fix: 없음. 차후 개정 시 "required on main; always-runs on develop" 1줄 정연 권장.

## Recommendations

- 리드는 CI green 판독(D4)을 card done의 전제로 삼는다 — 본 감사의 로컬 근거는 전항 재실측으로 건전하므로, CI 1회 green이면 부채가 소멸한다.
- 본 게이트의 tracked 변형은 로컬 표준으로 정착했다(§D.6 예측 실측 확인). 후속 카드가 `moai gate` format 스텝(배포 표면)을 다룰 때 이 tracked/전수 동치 근거를 출발점으로 쓸 수 있다.
