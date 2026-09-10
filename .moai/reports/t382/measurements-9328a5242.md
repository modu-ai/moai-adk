# t382 plan-phase 측정 기록 — tree 9328a5242 (worktree `.claude/worktrees/t382`, branch `WT-era-plan-phase`)

측정 도구: `./bin/moai` (이 트리에서 `make build`한 바이너리). PATH 바이너리 미사용.

| # | 명령 | 관측값 |
|---|---|---|
| M1 | `./bin/moai spec audit` | Total SPECs 714 / Grandfathered 285 / Modern-era clean 422 / Drift findings 499 |
| M2 | `./bin/moai spec audit --json` → `EraAutoDetected` era 집계 | V2.x 144, V3R2-R4 118, V3R5 23, V3R6 205 (합 490). `EraAutoDetected`는 `era:` frontmatter가 **없는** SPEC에만 발화하므로 V3R6 실제 총계는 714−285 = **429** |
| M3 | 같은 JSON, `finding_type` 집계 | `EraAutoDetected` 490, `SyncStatusDrift` 7, `AuditError` 2. **`EraUnclassified` 0건** |
| M4 | V3R5 23건의 `created:` 판독 | `created >= 2026-04-01` **22건** / `< 2026-04-01` **0건** / `created` 부재 **1건** (`SPEC-V3R5-INIT-WIZARD-EXPANSION-001`) |
| M5 | `grep -rl '^created_at:' .moai/specs/*/spec.md \| wc -l` | 46 |
| M6 | `grep -rl '^created_at:' .moai/specs --include=spec.md` (재귀) | 47 — 초과 1건은 `.moai/specs/_archive/SPEC-DESIGN-CONST-AMEND-001/spec.md` (감사 대상 밖) |
| M7 | M5의 46건 중 V3R5 버킷 소속 | **1건** (INIT-WIZARD) |
| M8 | V3R5 23건 중 `matchesModernPhase(phase)` 참 | 5건 (모두 M4의 POST 22건 안에 포함) |
| M9 | V3R5 23건 중 progress.md `§E.4` 보유 | 21건 / `§E.3` 보유 22건 |
| M10 | V3R5 23건 중 non-terminal status | **14건** (terminal = completed/superseded/archived/rejected) |
| M11 | `./bin/moai spec lint --json <14건 spec.md>` | rc=0, findings 11 — `FrontmatterInvalid`×7 (전부 INIT-WIZARD, 메시지 말미 `[grandfathered era — downgraded to warning]`), `MovingRefUnpinned`×3, `StatusGitConsistency`×1. **11건 모두 `advisory=true`** |
| M12 | `./bin/moai spec lint --strict --json <14건>` | **rc=0** (advisory라 승격 불가) |
| M13 | `./bin/moai spec drift` 중 23건 행 | `era-exempt` 22행 + `terminal-exempt` 1행 (`SPEC-HOOK-PREEDIT-INVESTIGATE-001`, superseded) |

부속 파일: `v3r5-population.txt`(23행), `drift-before-9328a5242.txt`(23행).

## 카드 지시문 전제 중 이 트리에서 어긋난 것

1. **"V3R6 212"** — 이 트리 실측은 `EraAutoDetected` 기준 **205**, 실제 V3R6 총계 **429**. 212는 어느 단위에도 맞지 않는다.
2. **"46 SPECs carry `created_at:`"** — glob(`*/spec.md`) 기준 46은 맞으나, 재귀 grep은 47을 낸다(초과 1건은 `_archive/`). 단위 라벨이 필요하다.
3. **"두 구조 게이트가 advisory로 강등된다"** — 참이지만 **오늘 카탈로그에서는 그 강등이 실제로 걸린 SPEC이 1건뿐**이다(INIT-WIZARD의 `FrontmatterInvalid` 7건). 나머지 13건의 non-terminal V3R5 SPEC은 강등 대상 ERROR를 애초에 갖고 있지 않다(M11). 즉 lint 축의 이득은 **잠재적**이며, 오늘의 rc 델타는 0이다(M12).
