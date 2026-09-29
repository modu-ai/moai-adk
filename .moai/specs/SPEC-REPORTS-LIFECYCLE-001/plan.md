# Plan — SPEC-REPORTS-LIFECYCLE-001

> 기준 트리: `.moai/worktrees/t1320` @ `6879cfa5e` (develop 기반 카드 워크트리).
> 증거 디렉터리: `.moai/reports/t1320/` — gitignored, 커밋 금지.

## §A Context

카드 t1320 의 6개 항목은 사실 세 부류로 묶인다: (1) **데이터 재배치** — tracked 511 파일의 `reports/` → `.moai/reports/historical/` 이동과 gitignore 상호작용(되돌리기 가장 비싼 결정), (2) **행동 변경** — skill 기본 경로, `moai worktree done` hoist, `moai clean` 아카이브(사용자 표면·CLI 동작), (3) **문서·가드** — 무시 규칙 회귀 가드, 폐기 플로우 문서, 템플릿 반영. 마일스톤 순서는 이 순서를 따른다.

핵심 발견(spec §A 참조): 루트 `reports/` 는 **tracked** 이므로 이동 동사는 `git mv` 다. 카드 본문의 "move" 를 로컬 `mv` 로 수행하면 511 파일이 원격에서 사라진다 — 이것이 본 계획의 최대 리스크이며 M1 전체가 이 축을 방어한다.

## §B Known Issues

1. `.gitignore:399` `.moai/reports/*.md` — 겉보기 중복, 실제로는 negation 블록 뒤에서 `.md` 재포함을 무조건 재배제하는 방어심층. 제거 금지, 가드 대상에 포함(spec §A.2).
2. 카드 본문 수치 3종이 이 트리 실측과 다르다(27항목/9.2MB/511 tracked · item 5 기충족 · 564MB는 머신 로컬량). REQ 는 실측에 고정됐다.
3. `done.go` 에 hoist 부재 + **L1 거부**(iter2 확인) — hoist 0건(grep, exit 1)이고 done 은 L1 트리(`.moai/worktrees/*` 포함)를 auto(`:80`)·interactive(`:284`) 양 경로에서 `--force` 로도 거부한다(SPEC-WORKTREE-DONE-TIER-001). 유일한 "hoist" 적중은 `clean.go:150` 의 무관한 주석. 즉 카드 트리는 done 이 절대 제거하지 않으므로 hoist 의 실행 메커니즘은 독립 동사여야 한다(§F.0 D2).
4. `moai update` 는 `.moai/reports` 를 지우지 않는다(관리 대상 뿌리 밖, `CleanMoaiManagedPaths` 대상 목록에 부재) — 이동·아카이브된 증거의 생존은 이 근거에 귀속된다.

## §C Pre-flight

- [x] SPEC ID 사전검증: `SPEC-REPORTS-LIFECYCLE-001` → Bash 정규식 **PASS** (본 러닝에서 실행, 출력 인용: spec 작성 전 실행 기록)
- [x] ID 중복 없음: `.moai/specs/` 내 REPORTS 접두 0건
- [x] 루트 reports/ tracked 확인: `git ls-files reports/ | wc -l` → 511
- [x] gitignore 규칙 좌표 확인: 로컬 235/306/399, 미러 267/234 (+269-270, 278)
- [x] html-report `<cwd>/reports/` 좌표 확인: 양 사본 65·74행
- [x] done.go hoist 부재 확인
- [ ] (run-phase) M1 착수 전 `git ls-files reports/` 재측정 — 다른 카드의 병합으로 집합이 늘었을 수 있다

## §D Constraints

spec §D 전항을 인용한다. 추가로:

- hoist·아카이브 구현은 **임시 프로젝트(t.TempDir)** 에서만 검증한다. dev 프로젝트의 실제 `.moai/reports/` 를 대상으로 명령을 돌리지 않는다(데이터 파괴 위험).
- 아카이브는 move 전용이다. 어떤 코드 경로도 `os.RemoveAll` 를 reports 내용에 대고 쓰지 않는다.
- 템플릿 미러 편집 커밋마다 `make build` (선행 emit-check 포함)를 돌린다.

## §E Self-Verification

run-phase 종료 시 §E.2 에 증거를 남긴다. 판정 명령 세트:

```
git ls-files reports/ | wc -l                                  # 0 기대
git ls-files .moai/reports/historical/ | wc -l                 # 511 기대 (M1 착수 시 재측정값)
git diff --stat <base>..HEAD -- .gitignore internal/template/templates/.gitignore
grep -c '<cwd>/reports/' .claude/skills/moai-domain-html-report/SKILL.md \
  internal/template/templates/.claude/skills/moai-domain-html-report/SKILL.md   # 0 0 기대
go test -timeout 30m ./internal/cli/worktree/... ./internal/cli/...
make build && make embed-check
go test -count=1 -timeout 30m ./internal/spec/... ./internal/template/...   # 가드 테스트 소속 패키지 (구현 시 확정)
```

## §F 결정 포인트 (검토 우선순위순) — 와 마일스톤

### F.0 결정 포인트 — 되돌리기 비용 내림차순

**D1. 루트 reports/ 보존 방식** (최고 변경 가능성 — 데이터 재배치) — **iter2 레인 채택 확정**
- **채택: tracked 연속성** — `git mv` 로 이동, `historical/` 재포함 negation 없음. 이미 추적된 파일은 무시 규칙의 영향을 받지 않으므로 원격 보존이 성립하고, 신규 파일은 계속 로컬 전용이다. `git status` 노이즈 0.
- 기각 기록: `!.moai/reports/historical/**` 재포함(399행 뒤 배치) — 디렉터리 전체가 tracked 후보가 되고 신규 파일이 untracked 노이즈로 뜬다. 단순성 사다리 1단계(이 동작이 필요 없다).

**D2. hoist 메커니즘 형상** — **iter2 재설계 확정 (감사 D4 structural 발견 반영)**
- **채택: (a) 독립 동사 `moai worktree hoist <tree-path>`** — hoist 루틴은 내부 공용 함수로 두고, ① 동사가 그것을 노출하고 ② `moai worktree done` 은 L2 트리 제거 직전 같은 루틴을 호출하며 ③ L1 세션 종료 폐기 플로우는 문서 의무로 동사 호출을 명시한다. 이 트리에서 기계 검증 가능한 것이 이 형상뿐이다: 동사는 t.TempDir 프로젝트에서 직접 호출·관측된다.
- 기각 (b) 세션 종료 L1 폐기 프롬프트에 내장 — 그 프롬프트는 Claude Code 런타임 표면이라 Go 검증 불가점이 없고, AC 가 조항 존재만으로 GREEN 하는 iter1 과 동일 결함을 재현한다.
- 기각 (c) REQ 를 done 전용으로 재범위 — done 은 L1 트리를 절대 제거하지 않는다(measured: done.go `:80`·`:284` 양 경로 거부, `--force` 불가). 카드 오염의 실제 모집단인 27개 L1 트리에 실행 메커니즘이 남지 않아 카드 전제가 미해결로 남는다.

**D3. skill 기본 경로** — **iter2 레인 채택 확정**
- **채택: 무조건 `.moai/reports/`** — skill 은 moai 프로젝트에만 배포되고(`moai init` 이 `.moai/` 를 만든다), 프로젝트 인지 분기는 측정 가능한 이득 없이 경로 결정 트리를 하나 늘린다.
- 기각 기록: `.moai/` 부재 시 `<cwd>/reports/` 폴백 — skill 이 존재하는데 `.moai/` 가 없는 상태는 moai 비프로젝트뿐이다.

**D4. 아카이브 정책 파라미터** — **iter2 레인 채택 확정**
- **채택: 연령 기반 90일(mtime), `<YYYY-MM>` 분할, CLI 플래그 + `defaults.go`(신설 설정 파일 없음).**
- 세부 3값(창 길이·충돌 정책 skip-and-report·1GB 경고 임계)은 열린 마커가 아니라 **보수적으로 채택된 문서화 기본값**이다 — 90일은 카드 재오픈 주기 추정치이므로, 재측정 근거가 생기면 후속 카드에서 조정한다.

### F.1 마일스톤

#### M1 (Priority High) — 루트 reports/ → `.moai/reports/historical/` (D1 확정 후 착수)

- run 전 `git ls-files reports/` 재측정(기선 갱신), `git mv` 일괄 이동, 커밋.
- 검증: `git ls-files reports/` → 0 / `git ls-files .moai/reports/historical/` → 이동 전 수와 동일 / 정렬 경로집합 diff 0 / blob SHA 총합 불변 / `git status --porcelain | grep '^ D'` → 0건. 증거를 `.moai/reports/t1320/m1-migration.txt` 로 남긴다.
- gitignore 회귀 가드 테스트(REQ-RLC-008)를 같은 마일스톤에 착지 — 대상 규칙: 로컬 235/306/399 + 미러 267/234. 변이 주입(규칙 삭제 사본)으로 붉어지는 것을 관측.

#### M2 (Priority High) — html-report skill 기본 경로 (D3 확정 후)

- 양 미러 SKILL.md 65행 표·74행 Output 절의 `<cwd>/reports/` → `.moai/reports/` 교체 — **같은 커밋**, 직후 `make build`(catalog 해시 재생성).
- 디렉터리 자동 생성 절차 문장 추가(REQ-RLC-004).
- 검증: grep 0/0, 미러 패리티, `make build` + emit-check 클린, 신규 프로젝트(t.TempDir)에서 렌더 1회 관측(AC-RLC-006).

#### M3 (Priority High) — hoist: 독립 동사 + done(L2) 호출 + 폐기 플로우 문서 (D2 재설계 확정 후)

- hoist 공용 루틴 구현(트리의 `.moai/reports/` → 루트 `.moai/reports/worktrees/<tree-name>/`, 상대경로 보존, 건수/바이트 출력, 충돌 skip-and-report, 프로젝트 루트 밖 경로 거부).
- `moai worktree hoist <tree-path>` 동사로 노출(`internal/cli/worktree/`), `done` 의 L2 제거 경로에 제거 전 호출 배선(같은 루틴), `--no-hoist` 플래그는 done 측에만 둔다.
- `worktree-integration.md` + 템플릿 미러에 hoist-before-dispose 의무 조항 추가 — L1 세션 종료 폐기 경로가 **동사를 호출하는 것**을 절차로 명시한다.
- 검증: t.TempDir 기반 테스트 4종(동사 정상 인출 / 충돌 미덮어쓰기 / 루트 밖 경로 거부 / done 이 루틴 호출) + `go test -timeout 30m ./internal/cli/worktree/...`.

#### M4 (Priority Medium) — moai clean 아카이브 액션 (D4 채택 확정 후)

- `moai clean` 에 reports-archive 액션 추가. **default-deny 술어**(REQ-RLC-007): 최상위 항목 ∧ 이름이 `^t[0-9]+$` 또는 `^SPEC-[A-Z0-9-]+[0-9]{3}$` ∧ mtime 90일 초과 ∧ tracked 파일 0개 — 세 조건이 모두 참일 때만 `archive/<YYYY-MM>/` 로 move. 명시 보호: `historical/`·`plan-audit/`·`worktrees/`·`archive/`·tracked 포함 항목. 건수/바이트 출력, 1GB 경고.
- 검증: t.TempDir 테스트(술어 적중 이동 / 비적합 무영향 / 보호 3종 무영향 / tracked 포함 항목 자동 보호) + `go test -timeout 30m ./internal/cli/...`.

#### M5 (Priority Medium) — 템플릿·문서 반영 및 전체 재검증

- 미러 변경분 중립성 자가점검(C1-C8 금지 클래스 grep: `t1320`, `SPEC-REPORTS-LIFECYCLE`, 본 카드 날짜, SHA) → 0건.
- `make build` + `make embed-check`, 변경 패키지 테스트 재실행, AC 매트릭스 전수 판정 결과를 `progress.md` §E.2 에 기록.

## §G Anti-Patterns

- `git mv` 대신 `mv` — 원격에서 511 파일 소실로 직행.
- skill 한쪽 미러만 수정 — 패리티 가드와 CI 동시 적색.
- `.gitignore` negation 을 399행 **앞**에 삽입 — 방어심층이 negation 을 이겨 의도 반전.
- hoist 를 세션 종료 훅에 배선 — 훅은 폐기 판단 지점이 아니며 done 의 anchor 가드와 이중 통제가 된다.
- 아카이브에 삭제 경로 동봉 — "move 전용" REQ 위반이자 데이터 파괴.
- 템플릿 미러에 카드 유래물 잔류 — AC-RLC-011 적색, sync 단계 재작업.

## §H Cross-References

- `.claude/rules/moai/workflow/worktree-integration.md` — 폐기 플로우(수정 대상)
- `.moai/docs/update-local-file-survival.md` — `moai update` 관리 뿌리 목록(생존 근거)
- `kanban-dispatch.md` § Completion is read, never trusted — 증거 인출이 필요한 이유
- AGENTS.md §3 — 트리 폐기 금지 규율의 데이터 근거
- 카드 t1305 primary hoist — 수동 hoist 의 모범 사례(D2 근거, SPEC 아님)
