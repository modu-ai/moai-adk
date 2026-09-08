# t255 사전 조건 판정 (lane-15, 2026-09-02)

## Claim

t255(pre-commit.local 확장점)의 출시 조건 — "t230 착지 이후 출시" — 는 **미충족**이다.
t230의 provenance 판별기가 어떤 릴리스에도 배송되지 않았다. 배차 지시("조건 미충족이면
그 사실을 보고하고 다음 카드로 넘어가십시오")에 따라 t255 작업을 보류한다.

## Evidence

| # | 명령 | 관측 결과 |
|---|------|----------|
| 1 | `git log --oneline --all --grep=t230` | `32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)` · `539349c5b docs(t230): sync-audit PASS 95/100 (#1649)` · `12d837b9b chore(SPEC-PRECOMMIT-PRESERVE-001): backfill sync_commit_sha cd90ec40b` — t230 착지 + 3-phase close 완료 |
| 2 | `git merge-base --is-ancestor v3.1.2 32d2221fa` | 참 — t230은 v3.1.2 **이후** 착지 |
| 3 | `git tag --contains 32d2221fa` | **빈 출력** — t230을 담은 태그(배송 릴리스) 0개 |
| 4 | `git tag --list 'v3.1*'` (tail) | v3.1.0 … v3.1.2에서 종단 — v3.1.3 미태그, v3.1.4 태그 없음 |
| 5 | `git merge-base --is-ancestor 32d2221fa 10948d057` (release/v3.1.4 팁) | 참 — 판별기를 최초로 실게 될 릴리스는 **v3.1.4** (PR #1685, 미머지, t204 운영자 배포 게이트) |

## 판정

- 지금 t255의 훅 본문 변경(pre-commit.local 위임)을 develop에 착지시키면, 판별기 최초
  릴리스가 될 v3.1.4가 본문 변경을 동시에 싣는다. 카드 본문이 경고한 함정이 그대로
  발화한다: *"훅 본문 변경과 provenance 판별기가 같은 릴리스에 들어가면 예외 없이 모든
  설치 기반에서 installed != incoming이 참이 되어, 모든 사용자가 첫 업그레이드에 백업과
  경고를 받는다."*
- 조건 충족 요건: v3.1.4 배송(설치 기반 스냅샷 기록 시작) → 최소 한 릴리스 경과 →
  그 다음 릴리스에 t255 착지.
- 결론: **t255 보류** — 착지 가능 시점은 v3.1.4 배송 이후의 다음 릴리스 창이다.

## Gaps

- v3.1.4 PR(#1685)의 머지·배송 시점은 미관측 — 운영자 게이트(t204) 소관이다.

## Residual-risk

- v3.1.4가 판별기를 빼고 재편성되면(스콥 변경) "최초 판별기 릴리스" 기준점이 바뀐다 —
  그 시점에 본 판정을 재측정해야 한다.
