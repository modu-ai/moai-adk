# SPEC-FMT-GATE-001 — 구현 계획 (plan-phase, Tier M)

> 카드 t465 · 브랜치 `WT-format-gate-zero` · 기준 base `d592b0551`(=origin/develop)
> 선행 카드 t457(`WT-gofmt-drift` @ `e1fdf00d1`, ahead 6, 리드 병합 창 대기)

## §A Context — 무엇을 왜 활성화하는가

현재 포맷 게이트 0개(실측 표는 spec.md §A). 선택한 활성 표면은 **루트 `.github/workflows/ci.yml`의
Lint 잡** — 무조건 실행되는 required check이므로 로컬 훅 상태(core.hooksPath=/dev/null로 비활성)와
무관하게 기계적으로 강제된다. `moai gate` 포맷 스텝 추가는 배포 제품 코드 변경이므로 기각(별도 카드).

**결정 가역성 순서**(아래 §F 마일스톤 동일 순서): 가장 바뀔 가능성이 높은 결정(게이트가 어디서
어떤 명령으로 판정하는가)을 먼저 검토하게 배치했다.

## §B Known Issues

1. **활성 조급증 적색** — 게이트가 t457 착지 전에 develop에 들어가면 CI가 154 위반으로 전면 적색.
   M1의 전제조건 게이트(§C)가 이를 봉쇄한다.
2. **templ 생성물 상호작용** — lint 잡의 templ codegen drift guard가 `internal/web/*_templ.go`를
   재생성·검증한다. templ 표준 출력은 gofmt-clean이며 현재 154 목록에 `*_templ.go`는 0건
   (`grep -c '_templ' .moai/reports/t465/gofmt-l.txt` → `1`이나 해당 1건은 파일명에 "template"가
   들어간 테스트 파일 `runner_template_test.go`, 생성물 아님). 향후 templ/go 버전 상호작용이
   생성물을 gofmt-dirty로 만들면 게이트가 붉어진다 — 그때의 수리는 생성물 재생성이지 게이트 완화가 아니다.
3. **testdata fixture** — `internal/navigator/astx/testdata/enrich/src/` 하위 `.go` 2개가 154 목록에
   포함. 파스 가능하고 포맷만 지저분한 fixture라 t457 정리 범위에 있다. 게이트는 전수
   `.go` 파일을 검사하므로 fixture도 항상 gofmt-clean이어야 한다.
4. **untracked 노이즈** — 워크트리의 untracked `.go`가 로컬 판정을 오염할 수 있다. 로컬 측정은
   tracked-files 변형을 표준으로 한다(§D).
5. **`moai update`가 `.moai/config`를 통째 삭제** — 본 SPEC은 `gate.yaml`을 건드리지 않으므로
   무관하나, 추후 `moai gate` 포맷 스텝 논의 시 이 제약이 재등장한다(별도 카드에 인계).

## §C Pre-flight — 활성 직전 전제 게이트 (실행형)

activation commit 생성 **직전**에 다음 복합 명령을 실행하고 둘 다 exit 0임을 확인한 뒤에만
커밋한다(REQ-FG-003/REQ-FG-004의 런 페이즈 인코딩):

```bash
# (1) 브랜치가 post-t457 develop를 흡수했는가 — t457 tip이 HEAD의 조상인가
git merge-base --is-ancestor e1fdf00d1 HEAD && echo ANCESTOR-OK
# (2) 이 트리는 게이트 기준으로 녹색인가
test -z "$(gofmt -l .)" && echo FMT-CLEAN
```

(1)이 실패하면 아직 활성 창이 아니다: `git fetch origin develop` → 확인 → 본 워크트리에서
`git merge origin/develop`로 흡수 후 재측정. (2)가 실패하면 새 위반 유입이므로 원인 규명 후
해당 카드로 환수(본 카드에서 `gofmt -w` 수정 금지).

사후 감사 형태(이후 세션이 기계 판정하는 인코딩):

```bash
git merge-base --is-ancestor e1fdf00d1 <activation-sha>; echo $?      # 0 기대
git checkout <activation-sha> && gofmt -l . | wc -l                    # 0 기대
```

## §D Constraints

- 게이트 판정 명령의 표준형: CI — `test -z "$(gofmt -l .)"`(clean checkout에서 전수와 동치).
  로컬 — `git ls-files -z '*.go' | xargs -0 gofmt -l`(untracked 제외, 실측 동치 확인됨).
- `make fmt`(gofumpt)은 유효한 수정 경로로 유지 — gofumpt 출력은 gofmt-clean.
- `internal/template/templates/**` 불변(REQ-FG-005). 배포 템플릿 `.github/workflows/`는
  `label-sync.yml` 단 1개이며 본 카드가 루트 CI를 고쳐도 사용자에게 실리지 않는다.
- 전체 테스트 스위트 로컬 실행 금지. 런 검증 = 게이트 명령 이진 판정 + develop push 후 CI 판정.

## §E Self-Verification (런 페이즈 측정 계획)

| 항목 | 명령 | 기대 |
|---|---|---|
| 활성 전제 1 | `git merge-base --is-ancestor e1fdf00d1 HEAD; echo $?` | `0` |
| 활성 전제 2 | `test -z "$(gofmt -l .)"; echo $?` | `0` |
| 명령 이진성(더러움 감지) | scratch 브랜치에서 1파일 변형 후 `test -z "$(gofmt -l .)"; echo $?` | `1` |
| 템플릿 불변 | `git diff --name-only <act-parent> <act-sha> -- internal/template/templates/ \| wc -l` | `0` |
| 로컬 패리티 | `make fmt-check; echo $?` (clean 트리 / dirty scratch 각각) | `0` / `1` |
| 최종 판정 | develop push 후 CI Lint 잡 | green |

## §F Milestones

- **M1 (High) — CI 포맷 게이트 + `make fmt-check` 단일 커밋 활성 (activation commit)**. 리드 결정
  D1(2026-09-03, BINDING — §I): 루트 `.github/workflows/ci.yml` Lint 잡의 format-gate 스텝과
  Makefile `fmt-check` 타깃(tracked-files 변형)을 **같은 activation 커밋**으로 인도한다. t457 착지
  전에 타깃만 먼저 두지 않는다 — t457 착지 전까지 non-zero를 보고하는 타깃은 주인 없는 적색이며,
  이 저장소는 이미 그 패턴("한동안 적색인 것이 정상")에서 시작한 상속 적색 46건(카드 t444)의 비용을
  치르고 있다. 그 창에서 타깃은 `gofmt -l .` 이상의 정보를 추가하지 않는다. 단일 커밋 인도는
  "언제부터 참인가"를 계보(genealogy)에서 답 가능하게 만든다. **t457 착지 후에만** §C 전제 게이트를
  통과해 커밋한다. plan-phase 문서 커밋은 본 브랜치에 이미 착지 — 활성 커밋만 대기.
  완료 신호: activation-sha + §E 표 전체 행 실측값.
- 완료 후: 리드에게 병합 SHA·증거 경로 보고 → 리드 일괄 develop push → CI 판정 확보.

## §G Anti-Patterns

- **t457 전 활성** — 전제 게이트 우회(CI 전면 적색 유발). 금지.
- **drive-by `gofmt -w`** — 154개 정리는 t457 소관. 본 카드의 커밋이 `.go` 파일을 수정하면
  REQ-FG-004 측정이 오염된다.
- **gofumpt를 게이트 기준선으로** — 154 밖의 추가 위반으로 즉시 적색. 상위집합은 수정 경로로만.
- **템플릿 동반 수정** — "템플릿에도 게이트를"은 별도 카드의 명시적 결정. REQ-FG-005 위반.
- **`.golangci.yml` enable-set 확장** — pinning 주석이 명시한 deliberate future decision이며
  본 카드 범위 밖(golangci 버전 skew 검토가 선행되어야 함).
- **게이트 완화로 적색 회피** — templ 생성물·fixture 적색은 재생성/정리로 수리한다.

## §H Cross-References

- 카드 t457 — `.claude/worktrees/t457`, 브랜치 `WT-gofmt-drift`, tip `e1fdf00d1`
- 기준선 증거: `.moai/reports/t465/gofmt-l.txt` (154 files @ `d592b0551`, 2026-09-03)
- `.golangci.yml` — enable-set pinning 근거(주석 블록)
- `internal/cli/gate.go` / `.moai/config/sections/gate.yaml` — 기각 표면의 현재 스텝 구성
- CLAUDE.local.md §4.1 — "판정은 CI" 규율(활성 표면 선택 근거)

## §I Lead Decisions (2026-09-03)

- **D1 (BINDING) — 공개 질문 ① 해소**: `make fmt-check`는 게이트 활성과 **같은 커밋**으로
  인도한다(§F M1). 타깃의 선행 착지 금지 — 착지 전 창의 non-zero 보고는 주인 없는 적색이며
  상속 적색 46건(t444)의 패턴 재현이다. 그 창에서 타깃은 `gofmt -l .` 이상의 정보를 추가하지
  않는다. 단일 커밋 인도는 "언제부터 참인가"를 계보에서 답 가능하게 한다.
- **D2 (기록 전용) — 공개 질문 ② 해소**: `.golangci.yml` gofmt linter 제외는 올바른 범위
  규율로 리드 승인. 후속 카드 후보는 리드가 등록했고 발행은 운영자 소관 — 본 레인은 발행하지
  않는다(§G anti-pattern 상동).
