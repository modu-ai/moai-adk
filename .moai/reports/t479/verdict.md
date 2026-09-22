# t479 — sync-phase 종결 증거

- 카드: **t479**
- SPEC: **SPEC-BINLAG-KEYGUARD-001**
- 워크트리: `.claude/worktrees/t479` · 브랜치 **`WT-binarylag-key-guard`**
- base 핀: **`93fb36344`** (로컬 `develop @ a825183dd` 흡수 병합 커밋)
- 측정 시각의 HEAD: **`559d6e26f`** → 스테이징 직전 재판독에서 **`38713c31e`**로 이동
  (아래 E1 참조. 소스 변경 폭은 새 HEAD에서 재측정했고 값이 같다)
- 측정 위치: 이 워크트리. primary 체크아웃이나 병합 트리가 아니다.

---

## Claim

1. sync 이전 브랜치 상태에서 `TestBinaryLag` 계열이 **공허하지 않게** 초록이고,
   신규 가드 `TestBinaryLag_AllowlistKeysAreLiveNames`가 선택·실행·통과했다.
2. `go vet ./internal/cli/...`가 exit 0, 출력 0바이트다.
3. base `93fb36344` 대비 `internal/` · `pkg/` · `cmd/` 소스 변경은
   `internal/cli/binary_lag_test.go` **한 파일**(+70줄)뿐이다.
4. sync 산출물 3종을 이 창에서 작성했다 — `CHANGELOG.md` `[Unreleased]` 엔트리 1건,
   `spec.md` frontmatter `in-progress → completed`, `progress.md` §E.4 신호.
5. **push하지 않았다.** 브랜치·PR·병합을 만들지 않았고 워크트리를 폐기하지 않았다.

---

## Evidence

### E1 — 브랜치·커밋 상태

    $ git rev-parse --show-toplevel
    /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t479

    $ git branch --show-current
    WT-binarylag-key-guard

    $ git log --oneline -4
    559d6e26f docs(SPEC-BINLAG-KEYGUARD-001): land plan+run artifacts, draft -> in-progress (t479)
    729060e63 test(SPEC-BINLAG-KEYGUARD-001): M1 guard allowlist keys are live check names (t479)
    93fb36344 Merge develop a825183dd into WT-binarylag-key-guard (t479): absorb t466 Hook Delivery check + t477 allowlist entry
    a825183dd Merge branch 'WT-binarylag-allowlist' into develop (t477)

**HEAD 이동 관측 — 이 창 도중 다른 세션이 이 트리에 커밋했다.** 스테이징 직전 재판독에서:

    $ git rev-parse --short HEAD
    38713c31e        (E2·E3·E4를 잰 시점에는 559d6e26f)

    $ git show --stat --oneline 38713c31e
    38713c31e docs(SPEC-BINLAG-KEYGUARD-001): record stale-read and section-E conflict check (t479)
     .moai/specs/SPEC-BINLAG-KEYGUARD-001/progress.md | 59 ++++++++++++++++++++++++
     1 file changed, 59 insertions(+)

**유실은 없다** — 그 커밋이 담은 §E.2 R14 본문은 내가 `progress.md`를 읽은 시점에 이미
워킹트리에 있었고(커밋 전 미추적 상태였다), 내 편집은 `§E.4`의 자리표시자 한 줄만 바꾼다.
편집 후 `git diff`가 `progress.md`에 대해 보이는 것은 §E.4 블록 하나뿐이다(28줄 추가, 1줄 삭제).
그럼에도 **활성 작업 중인 워크트리에 외부 커밋이 들어온 것 자체가 공정 결함**이므로
여기 기록하고 리드에 보고한다.

### E2 — 테스트 (판정은 **두 매치 수**이며 종료코드가 아니다)

    $ go test ./internal/cli/ -run TestBinaryLag -count=1 -timeout 600s -v
    (exit 0)
    === RUN   TestBinaryLag_OneSeamServesBothSurfaces
    --- PASS: TestBinaryLag_OneSeamServesBothSurfaces (0.19s)
    === RUN   TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero
    --- PASS: TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero (0.15s)
    === RUN   TestBinaryLag_AllowlistKeysAreLiveNames
    --- PASS: TestBinaryLag_AllowlistKeysAreLiveNames (0.00s)
    === RUN   TestBinaryLag_DoctorCheckNameSetIsUnchanged
    --- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.07s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.132s

판정식 — 종료코드는 읽지 않는다. 0매치 실행도 `ok` + exit 0을 내기 때문이다
(`spec.md` §2 REQ-BLKG-005의 실측).

    $ grep -c '\[no tests to run\]' <출력>                                  → 0   (1단: 빈 sweep 아님)
    $ grep -c '^--- PASS: TestBinaryLag_AllowlistKeysAreLiveNames ' <출력>  → 1   (2단: 뒤 공백 앵커, SKIP 차단)
    $ grep -c '^--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged '     → 1   (형제 가드도 초록)
    $ grep -c '^--- SKIP' <출력>                                            → 0
    $ grep -c '^--- FAIL' <출력>                                            → 0

`$` 앵커를 쓰지 않은 이유: go가 이름 뒤에 ` (0.06s)`를 덧붙이므로 `$` 앵커는
통과한 테스트에도 **상시 0**이 된다(`spec.md` HISTORY 0.4.0 ②).

### E3 — vet

    $ go vet ./internal/cli/...
    (exit 0)
    $ wc -c <출력파일>
    0

출력이 0바이트다 — 지적 없음.

### E4 — 변경 폭

    $ git diff --stat 93fb36344..HEAD -- internal/ pkg/ cmd/       (HEAD = 559d6e26f)
     internal/cli/binary_lag_test.go | 70 ++++++++++++++++++++++++++++++++++++++
     1 file changed, 70 insertions(+)

HEAD 이동 후 재측정 — 값이 같다(이동한 커밋이 `progress.md`만 건드렸기 때문):

    $ git diff --stat 93fb36344..38713c31e -- internal/ pkg/ cmd/
     internal/cli/binary_lag_test.go | 70 +++++++++++++++++++++++++++++++++++++++
     1 file changed, 70 insertions(+)

두 커밋 전체 범위(`559d6e26f` 시점):

    $ git diff --stat 93fb36344..HEAD
     .moai/reports/t479/plan-audit.md                   | 1020 ++++++++++++++++++++
     .moai/specs/SPEC-BINLAG-KEYGUARD-001/acceptance.md |  489 ++++++++++
     .moai/specs/SPEC-BINLAG-KEYGUARD-001/plan.md       |  178 ++++
     .moai/specs/SPEC-BINLAG-KEYGUARD-001/progress.md   |  703 ++++++++++++++
     .moai/specs/SPEC-BINLAG-KEYGUARD-001/spec.md       |  237 +++++
     internal/cli/binary_lag_test.go                    |   70 ++
     6 files changed, 2697 insertions(+)

프로덕션 코드 0, 템플릿 트리(`internal/template/templates/`) 0.

### E5 — CHANGELOG 중복 자가검사 (B12)

    $ grep -c 'SPEC-BINLAG-KEYGUARD-001' CHANGELOG.md      → 0    (append 전)
    $ grep -c 'SPEC-BINLAG-KEYGUARD-001' CHANGELOG.md      → 1    (append 후)

AC 수 일치 — `acceptance.md`가 SSOT다:

    $ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l
    9

    AC-BLKG-001 … AC-BLKG-009

0이 아니다 — 0이었다면 공허한 비교이므로 손으로 다시 봐야 했다. CHANGELOG 엔트리도 9로 적었다.

경로 주장 검증 — 엔트리가 언급한 파일이 실재한다(5개 모두 존재):

    internal/cli/binary_lag_test.go
    internal/cli/doctor.go
    .moai/specs/SPEC-BINLAG-KEYGUARD-001/spec.md
    .moai/specs/SPEC-BINLAG-KEYGUARD-001/progress.md
    .moai/specs/SPEC-BINLAG-KEYGUARD-001/acceptance.md

인용 좌표(`grep -n` on `internal/cli/binary_lag_test.go`):

    118:func checkNamesFromSource(t *testing.T, src []byte) map[string]bool {
    162:func exprSource(fset *token.FileSet, src []byte, e ast.Expr) string {
    194:var namesAddedAfterBaseline = map[string]bool{
    221:func TestBinaryLag_AllowlistKeysAreLiveNames(t *testing.T) {
    283:		if beforeNames[name] || namesAddedAfterBaseline[name] {

### E6 — frontmatter 전이

    $ sed -n '5p' .moai/specs/SPEC-BINLAG-KEYGUARD-001/spec.md
    status: completed
    $ sed -n '7p' .moai/specs/SPEC-BINLAG-KEYGUARD-001/spec.md
    updated: 2026-09-06

    $ grep -n '^status:' plan.md acceptance.md
    (매치 없음 — 두 파일은 status: 필드를 갖지 않는다)

건드린 frontmatter 필드는 `status` 하나다. `updated:`는 이미 오늘(2026-09-06)이었으므로
값이 바뀌지 않았다. 본문(§A~§H)은 어느 파일도 수정하지 않았다.

---

## Baseline-attribution

이 회차에 이 트리에서 직접 실행한 것만 적었다.

| 주장 | 명령 | 잰 트리 |
|---|---|---|
| 가드 초록 (두 매치 수) | `go test ./internal/cli/ -run TestBinaryLag -count=1 -timeout 600s -v` | `559d6e26f` (worktree t479) |
| vet 무지적 | `go vet ./internal/cli/...` | 같음 |
| 변경 폭 1파일 | `git diff --stat 93fb36344..HEAD -- internal/ pkg/ cmd/` | base `93fb36344` → HEAD `559d6e26f` |
| AC 수 9 | `grep -oE … acceptance.md \| sort -u \| wc -l` | 같음 |

**옮겨 적은 값**: 없다. run-phase(`progress.md` §E.2)의 뮤턴트 3건(m1/m2/m2′) 실행 결과는
이 회차가 재현하지 않았다 — 그 근거는 §E.2에 있고 sync가 그것을 재측정으로 주장하지 않는다.

---

## Gaps

명시적으로 **관측하지 않은** 것.

- **전체 스위트**(`go test ./...`): 돌리지 않았다. 이 머신에서 금지이며 판정은 CI 몫이다.
- **`internal/cli` 밖 패키지**: 변경이 없어 돌리지 않았다.
- **크로스플랫폼 빌드**: 안 쟀다. 테스트 파일 1개이고 플랫폼 의존 코드가 없다는 것은
  추론이지 측정이 아니다.
- **golangci-lint**: 이 회차에서는 안 돌렸다. run-phase(§E.3)가 `0 issues`를 기록했으나
  그것은 `729060e63` 시점의 값이며 sync 회차의 측정이 아니다.
- **병합 트리에서의 재측정**: develop을 재흡수하지 않았다(리드 지시로 `93fb36344` 고정).
  병합 후 상태는 미관측이며 병합 창에서 재측정이 필요하다.
- **CI 판정**: push하지 않았으므로 이 카드에 대한 CI 실행은 존재하지 않는다.
- **`sync_commit_sha` 실제 값**: sync 커밋 시점에 `pending-backfill-sync` 자리표시자이며
  후속 커밋에서 채운다.
- **run-phase 뮤턴트 재현**: 하지 않았다(위 Baseline-attribution).

---

## Residual-risk

- **다른 카드가 `doctor.go`에 체크를 더하면** 허용목록과의 관계가 달라져, 파일이 겹치지 않아도
  병합 트리에서 새 가드가 RED가 될 수 있다. 병합 창에서 재측정이 필요하다.
- **양방향 진단의 도달 가능성**은 허용목록에 두 등록 모양이 공존하는 데 기대고 있다.
  나중에 `doctor.go`의 등록 방식이 한쪽으로 통일되면 한 분기가 조용히 도달 불가가 된다 —
  가드는 고장 나도 침묵한다. 현재는 두 뮤턴트가 각 분기를 실제로 밟았다(§E.2 R6).
- **`pending-backfill-sync`가 backfill되지 않을 위험**: 자리표시자는 갚아야 할 일을
  기록하지만 스스로 갚지는 않는다. 후속 커밋이 빠지면 §E.4는 영구히 미완이 된다.
- **CHANGELOG 엔트리의 서술 정확성**은 코드 판독에 기댄다. 인용한 5개 좌표는 이 트리에서
  재측정했으나, 산문 전체가 코드와 일치하는지는 기계가 판정하지 않았다.

---

## 처분

- 커밋: `WT-binarylag-key-guard` 위 로컬 커밋. **push 없음.**
- 브랜치 생성·병합·PR 없음. 워크트리 유지.
- 공개는 리드의 일괄 행위다.
