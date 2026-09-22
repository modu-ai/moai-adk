# t398 — 규약 문서가 착지한 `sync_commit_sha` 룰과 어긋나던 문제

카드: t398 · 브랜치 `WT-syncsha-slot-doctrine` · 베이스 로컬 develop `b7462203a`(fast-forward)

## Claim

`sync_commit_sha` 슬롯에 무엇을 쓸지 지시하는 규약 문서가 착지한 룰과 어긋나 있었다. 두 문서를
룰에 맞추고, 빈 슬롯이 왜 더 나쁜지를 관측된 근거와 함께 적었으며, 앞으로 문서와 룰이 갈리면
무엇이 잡는지를 가드로 세웠다.

## Evidence

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t398`, HEAD `b7462203a`(편집 전).

### 카드 전제 재측정 — 참

`internal/spec/syncsha.go` 를 이 트리에서 직접 읽었다(카드가 준 좌표를 세지 않고 grep 출력으로 확인):

```
78:var syncSHAPlaceholderPattern = regexp.MustCompile(`^pending-backfill(-[A-Za-z0-9-]+)?$`)
23://	PLACEHOLDER := "pending-backfill" ( "-" [A-Za-z0-9-]+ )?
```

`isCommitSHAToken`(:107)과 `isSyncSHAPlaceholder`(:118)가 각각 SHA·플레이스홀더를 판정하고,
`internal/spec/lint_syncsha.go` 의 `SyncSHASlotFormatRule.Check` 가 둘 다 아닌 값에만 warning 을 낸다.

### 어긋난 지점 — 실제로 어느 문장인가

| 문서 | 문면 | 룰과의 관계 |
|---|---|---|
| `.claude/rules/moai/workflow/spec-workflow.md` § Sync (close) | sync 커밋이 "`sync_commit_sha` 를 populating" 한다고만 적음 | **미달** — 커밋이 자기 해시를 모른다는 사실도, 그때 무엇을 쓰는지도 말하지 않는다. 읽는 쪽이 빈 슬롯을 고른 경로가 여기다 |
| `.claude/agents/moai/manager-docs.md` | `sync_commit_sha` 를 필드 목록에 이름만 올림 | **비대칭** — 형제 문서 `manager-develop.md:178` 은 `run_commit_sha` 에 "placeholder if backfill needed" 를 이미 달고 있다 |
| `.claude/rules/moai/development/spec-frontmatter-schema.md:140` (D3) | 플레이스홀더 후 후속 커밋 백필 | **일치** — 손대지 않았다 |

즉 "두 관행이 공존"의 실체는 규약 문서 두 곳이 **아무 관행도 지정하지 않은** 것이었고,
그 공백이 CHANGELOG 의 두 관행(빈 슬롯 / 플레이스홀더)으로 갈라져 나타났다.

### 빈 슬롯이 왜 더 나쁜지 — 추론이 아니라 재현

이 트리에서 `go build -o /tmp/t398-moai ./cmd/moai` 로 세운 바이너리로,
`status: completed` 픽스처 SPEC 의 슬롯만 세 번 바꿔 가며 관측했다(`lint-repro.txt` 전문):

| 슬롯 값 | `SyncSHASlotFormat` |
|---|---|
| (빈 값) | **1건, severity=warning** |
| `pending-backfill-sync` | 0건 |
| `b7462203a` | 0건 |

그리고 `--strict` 로 같은 빈 슬롯을 다시 재니 **exit 0** 이었다 — terminal status 문서의 warning 은
advisory 로 강등되므로 어떤 게이트도 이 신호에 막히지 않는다.

여기서 나오는 결론은 "빈 슬롯이 조용하다"가 아니라 그 반대다: **신호는 뜨지만 아무도 그 신호에
행동하지 않는다.** 룰 자신의 메시지가 그 까닭을 적고 있다 — "so no close will be triggered to
repair it". SPEC 은 sync 커밋이 착지하는 순간 `completed` 가 되므로 그 뒤로 close 가 다시 돌 일이
없고, 따라서 수리가 예약되지 않는다. 플레이스홀더는 갚아야 할 것이 무엇이고 어느 단계가 지고
있는지를 이름으로 남기지만, 빈 슬롯은 아무것도 남기지 않는다.

### 수리

- `spec-workflow.md` § Sync (close) 에 두 항목 추가 — sync 커밋이 쓰는 값(정본 플레이스홀더 +
  후속 커밋 백필, D3 창 인용)과, 빈 슬롯이 대안이 아닌 이유(위 관측을 근거로).
- `manager-docs.md` 의 `sync_commit_sha` 항목에 같은 내용을 한 절로 — `manager-develop.md` 의
  기존 문면과 대칭.
- **3본 전부**: 루트 사본 · `internal/template/templates/` 미러 · 방출된
  `.codex/agents/moai/manager-docs.toml`(`AGENTEMIT_UPDATE=1` 로 재생성) + `catalog.yaml` 해시 1건 갱신.

### 회귀 — 문서와 룰이 갈리면 무엇이 잡는가

`internal/spec/syncsha_doctrine_guard_test.go` (신규, sentinel `SYNCSHA_DOCTRINE_DRIFT`). 양방향으로 건다:

1. 문서가 **처방하는** 백틱 인용 플레이스홀더 철자를 전부 걷어 `isSyncSHAPlaceholder` 에 통과시킨다.
   문법을 좁히면 문서가 거짓이 되고 가드가 붉어진다.
2. 빈 값이 두 술어 어느 쪽에도 받아들여지지 않음을 단언한다. 문서의 "비워 두지 않는다" 문장이
   서 있는 전제가 바로 그것이다.

공허 방지: 4개 문서 각각이 **읽혔고 최소 1개 철자를 담는다**는 것을 따로 단언하고,
전체 하한 6(측정값: spec-workflow 2×2 + manager-docs 1×2)을 건다 — 지침이 삭제되거나 미러 하위트리가
사라지면 스캔이 조용히 비는 대신 붉어진다.

**뮤턴트 3종을 심어 RED 를 관측하고 전부 복원했다:**

| 뮤턴트 | 관측 |
|---|---|
| M1 문법 협소화 (`^pending-backfill$`) | FAIL — 루트·미러 두 문서에서 `pending-backfill-sync` 거부 보고 |
| M2 한 문서에서 지침 삭제 | FAIL — 해당 파일 무처방 + 하한 미달(5 < 6) 두 줄 |
| M3 문법 확대해 빈 토큰 허용 | FAIL — 빈 슬롯 4형태 전부에서 전제 붕괴 보고 |

복원 후 `git diff --stat internal/spec/syncsha.go` 는 빈 출력(뮤턴트 잔재 0).

### 검증 범위

건드린 패키지 전부:

```
go test ./internal/spec/...      → ok  73.364s
go test ./internal/template/...  → ok  26.064s (template) + ok 0.289s (agentemit)
gofmt -l internal/spec/syncsha_doctrine_guard_test.go → 출력 없음
go vet ./internal/spec/          → 출력 없음
```

`go test ./...` 는 로컬에서 돌리지 않았다(전면 실행 금지 규율). 전 패키지 판정은 CI 몫이다.

## Baseline-attribution

모든 수치는 이 워크트리, 이 트리에서 이 세션이 직접 잰 값이다. 카드가 전한 t381 사고 관측
(WARNING 1096 → 1097 → 1096)은 lane-9 의 측정이라 인용하지 않았고, 대신 같은 성질을 픽스처로
새로 재현했다 — 위 3행 표가 그 재측정이다.

## Gaps

- **CI 판정 없음** — 브랜치가 미푸시라 이 변경에 대한 CI 판정이 존재하지 않는다. 위 초록은 전부
  로컬 darwin 측정이다.
- **CHANGELOG 의 과거 두 항목은 손대지 않았다.** t318 항목("no `pending-backfill-*` placeholder ever
  exists on the branch")은 그 시점에 실제로 일어난 일의 기록이고, 이력을 나중 규약으로 고쳐 쓰는 것은
  이 카드가 할 일이 아니다. 규약이 정해진 지금부터의 문서가 갈리지 않게 하는 것이 범위다.
- **`spec-frontmatter-schema.md` D3 는 미변경** — 재측정 결과 이미 룰과 일치했다. 일치하는 문서를
  건드리지 않은 것이지, 읽지 않고 넘긴 것이 아니다.
- **가드는 처방된 철자만 본다.** 문서가 룰과 어긋나는 방식은 이 축 말고도 있을 수 있고(예: 백필
  커밋의 시점을 잘못 적는 것), 그 축은 이 가드가 보지 않는다.

## Residual-risk

- 하한 6은 오늘의 문서 구성에서 나온 값이다. 나중에 어느 문서가 정당하게 통합·분할되면 하한이
  실제 모집단보다 커져 오탐이 될 수 있다 — 그때는 하한을 재측정해 갱신할 것이지, 지우지 말 것.
- `manager-docs.md` 는 미러 허용목록(`rule_template_mirror_test.go`)에 없다. 이 카드의 가드가 루트와
  미러 **양쪽**을 스캔하므로 이 축에서는 드리프트가 잡히지만, 그것은 플레이스홀더 문장에 한한 커버리지다.
  파일 전체의 바이트 동일성은 여전히 어떤 가드도 단언하지 않는다.
- 픽스처 SPEC 은 최소 형태라 lint 가 다른 warning 도 낸다(frontmatter 필수 필드 등). 위 표는
  `SyncSHASlotFormat` 코드만 필터해 센 값이고, 그 필터가 옳다는 것은 코드 문자열 대조에 의존한다.
