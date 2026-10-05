# t438 — 원격 `WT-*` 미병합 residue 축2 대응 조사 (t264 Stage C 선행)

- 카드: t438 (조사 카드 — **원격을 하나도 지우지 않는다**)
- 브랜치: `WT-remote-residue-axis2` (워크트리 `.claude/worktrees/t438`)
- 측정 트리: `131daa2901346a7516a49554690ac30438cf3c6b` (로컬 `develop` 과 동일)

## 못박은 이동 ref (VCI §2.1 R1)

| ref | SHA (측정 시점) |
|---|---|
| `origin/main` | `7ad9f8534dc48719854c67e2b9a06db97b594eaf` |
| `origin/develop` | `f7cabfc296aecc62fc22295c94c87b894594158b` |
| 로컬 `develop` | `131daa2901346a7516a49554690ac30438cf3c6b` |

세 값이 다르다는 것 자체가 아래 판별식의 핵심 입력이다. 이 문서의 모든 판정은 이 세 SHA 에 귀속되며, 다른 시점에 재실행하려면 `git fetch origin` 후 세 값을 다시 못박고 재측정할 것.

---

## Claim

축2(저작경로 공집합)를 원격 잔여 브랜치에 적용해 삭제 후보 목록과 판별식을 세웠다. 조사 결과 **t264 가 세운 축2 판별식에는 결함이 하나 있고**, 그것을 고친 뒤 라이브 원격 13개를 삭제 가능 11 / 보존 2 로 분류했다. 아울러 이 조사 과정에서 **stale tracking 축의 위험이 원격 삭제 축과 다른 종류임**이 드러났다 — 그쪽이 오히려 조용하다.

**삭제는 한 건도 수행하지 않았다.** `git fetch --prune` 도 실행하지 않았다.

---

## 1. 전제 정정 2건

**정정 ①: "원격 19 미병합 residue" 는 재현되지 않는다.** 현재 원격의 `WT-*` 브랜치는 `git ls-remote --heads origin 'refs/heads/WT-*'` 로 **13개**다(리드 2026-09-02 실측 13과 일치). "19" 는 t264 Stage A 당시 값이며 그 사이 릴리스·병합으로 줄었다. 조사 대상 모집단은 13이다.

**정정 ②: 로컬 `refs/remotes/origin/WT-*` 는 71개다 — 그중 58개는 원격에 없다.** 즉 로컬이 들고 있는 원격 추적 ref 의 82%가 스테일이다. 이 차이를 모르고 `refs/remotes/origin/WT-*` 를 모집단으로 잡으면 이미 없는 브랜치를 삭제 후보로 세게 된다. 실제로 첫 축1 측정에서 `--merged` 목록에 `ls-remote` 에 없는 4건(`WT-glm-flash-default`·`WT-mx-fanin-perf`·`WT-statusline-cost`·`WT-t250-followup`)이 섞여 나왔다.

---

## 2. t264 축2 판별식의 결함

t264 `plan.md:41-45` 의 축2:

> 브랜치의 저작 경로(`git diff --name-only <merge-base>..<tip>`)에 대해 `git diff origin/main..<tip> -- <paths>` 가 공집합이면 squash 착지로 판정

**결함 A — 기준 ref 가 `origin/main` 이다.** `origin/main` 은 release PR 로만 전진하므로 **설계상 develop 보다 뒤처져 있다**(측정 시점 `7ad9f8534` vs `131daa290`). develop 에 이미 착지했으나 아직 릴리스되지 않은 내용은 전부 "브랜치 고유" 로 오판된다. 실측 — `WT-astgrep-16-langs` 의 저작경로 18개에 대해:

| 기준 | 비공집합 경로 수 |
|---|---|
| `origin/main` (`7ad9f8534`) | 18 / 18 |
| 로컬 `develop` (`131daa290`) | 4 / 18 |

같은 브랜치가 기준 하나 바꿨다고 "전부 고유" 에서 "대부분 착지" 로 뒤집힌다. 축2 는 **통합 프론티어**(= 로컬 develop) 를 기준으로 재야 한다.

**결함 B — `git diff A..B` 는 방향을 구분하지 않는다.** 두 점 diff 는 "브랜치에만 있는 내용" 과 "develop 에만 있는 내용" 을 같은 비공집합으로 뭉갠다. 브랜치가 단지 **오래된** 것도 비공집합으로 나와 보존 판정을 받는다. 실측 — `WT-astgrep-16-langs` 의 develop 대비 diffstat 은 `15 insertions, 728 deletions`: 삭제 728줄은 브랜치가 지운 것이 아니라 **develop 이 그만큼 앞서 있다**는 뜻이다. `coverage-matrix.md` 가 `D` 로 나온 것도 develop 에만 있는 파일이기 때문이지 브랜치 고유가 아니다.

이 결함은 조사 중 실제로 저를 한 번 속였다: 저작 경로 대신 디렉터리 접두사(`internal/cli` 등)를 넘겼더니 `D` 수백 건이 쏟아졌는데, 전부 develop 이 나중에 추가한 파일이었다.

### 축2 정련안 (제안)

브랜치 `B` 의 저작 경로 `P = git diff --name-only $(git merge-base develop B)..B`. `B` 는 다음이면 삭제 안전:

```
git diff --name-status develop..B -- <P 의 정확한 파일 목록>
```

에서 **`A` 행이 0** 이고, 남은 `M` 행 각각이 "브랜치가 더 오래된 판본" 임을 확인했을 때.

- `A` (develop 에 아예 없음) = **브랜치 고유** → 보존. 가장 강한 신호이고 기계적이다.
- `D` = develop 에만 있음 → 브랜치 고유 아님. **무시한다**(결함 B 교정).
- `M` = 양쪽에 있고 다름 → 기계적으로 결정 불가. 내용을 봐야 한다. 실무적으로는 두 가지뿐이었다: 생성물 스탬프이거나, SPEC frontmatter 의 옛 `status:` 값이거나.

[HARD] 경로는 **디렉터리 접두사가 아니라 정확한 파일 목록**으로 넘긴다. 접두사를 넘기면 결함 B 가 되살아난다.

---

## 3. 라이브 원격 13개 분류

### 축1 — `git for-each-ref --merged=develop 'refs/remotes/origin/WT-*'`

`--merged` 는 tip 이 조상인 ref 만 열거하므로 `merge-base --is-ancestor` 를 브랜치 수만큼 반복하는 것과 동치이며 명령 1회로 끝난다.

**축1 통과 9건** — 내용이 develop 에 전부 존재. 축 하나로 삭제 근거 충분:
`WT-card-landing-state` · `WT-codex-launcher` · `WT-freshness-sync` · `WT-gate-three-axes` · `WT-ossdocs-v311` · `WT-t69` · `WT-version-stamp-predicate` · `WT-wscfg-graph` · `WT-wscfg-worktree`

**축1 실패 4건** — 리드가 지목한 4건과 정확히 일치:
`WT-astgrep-16-langs` · `WT-hook-wiring-drift` · `WT-main-stamp-repair` · `WT-precommit-vet`

### 축2 — 실패 4건에 적용

| 브랜치 | `A` | `M` | 브랜치 고유 내용 | 판정 |
|---|---|---|---|---|
| `WT-main-stamp-repair` | 0 | 1 | `provenance.json` 한 파일. 브랜치판은 `commit_sha: 3abde7053` / `generated_at: 2026-08-26T19:20:39Z` / `tree_root: …/t292`, develop 판은 `ad272be20` / `2026-09-02T05:31:10Z` / `…/t432` — **더 낡은 재생성 스탬프** | **삭제 가능** |
| `WT-astgrep-16-langs` | 0 | 3 | design.md 의 "Open decision — rule-id keying (R1)" 문단 + frontmatter `status: draft` / `updated: 2026-08-25`. develop 의 같은 파일에도 "rule-id keying" 이 존재(`git grep -c` = 1)하므로 주제는 살아 있고 브랜치판은 옛 표현. develop 이 progress.md 기준 728줄 앞섬 | **삭제 가능** |
| `WT-hook-wiring-drift` | **13** | 9 | `SPEC-HOOK-WIRING-DRIFT-001/{spec,plan,acceptance,progress}.md`, `.moai/reports/t216/*` 5건, `internal/cli/doctor_hook_wiring{,_test}.go`, `internal/cli/mx_index{,_test}.go`, `internal/template/hook_entries{,_test}.go` — develop 에 **하나도 없다** | **보존** |
| `WT-precommit-vet` | 0 | 4 | `internal/cli/hook_install_precommit{,_test}.go` · `internal/template/templates/.git_hooks/pre-commit` · `CLAUDE.local.md`. tip `b6f478b1a` 는 t237 카드 본문이 명시한 **검증 완료 패치** 그 자체 | **보존** |

두 보존 판정은 큐 상태와도 맞는다 — t216 과 t237 은 둘 다 `picked` 로 살아 있다. 즉 이 두 브랜치는 residue 가 아니라 **아직 착수되지 않은 작업의 유일본**이다.

### 삭제 후보 (승인 대상) — 11건

축1 통과 9 + 축2 통과 2(`WT-astgrep-16-langs`, `WT-main-stamp-repair`).

우연히 t264 Stage C 가 잡았던 "원격 11건" 과 같은 수이나 **같은 목록이라는 근거는 없다** — 그쪽은 리드 지목 10 + α 였고 이쪽은 두 축의 기계 판정이다. 수가 같다는 사실을 목록이 같다는 근거로 쓰지 말 것.

---

## 4. t264 보류 근거의 현재 상태

보류 근거였던 "원격 후보의 로컬 쌍둥이가 A/B 에서 삭제돼 지금 원격을 지우면 로컬 백업이 0" 은 **사실 관계로는 여전히 참**이다. 삭제 후보 11건 중 로컬 `refs/heads/WT-*` 에 이름이 있는 것은 `WT-version-stamp-predicate` 하나뿐이고, 그마저 tip 이 다르다(로컬 `0ffb14402` vs 원격 `e6c9234a8`).

그러나 근거의 **효력**은 달라졌다. 로컬 백업이 필요한 이유는 내용 유실 대비인데, 삭제 후보 11건은 정의상 내용이 develop 에 있다(축1 9건) 또는 develop 이 앞선 판본을 갖고 있다(축2 2건). **백업의 역할은 develop 이 한다.** 보류 근거는 "내용 보존 여부가 미측정" 이던 상태에서 세워졌고, 이 조사가 그 미측정을 해소했다.

---

## 5. 축2보다 조용한 위험 — stale tracking

카드가 함께 분류하라고 한 stale tracking 축은 원격 삭제와 **위험의 종류가 다르다**.

- `git remote prune --dry-run origin` 이 예고하는 삭제: **65건** (`WT-*` 58 + 비-`WT-*` 7 — `chore/cc-update-*` 2, `dependabot/*` 1, `gate-timeout-budget` 1, `release/v3.1.{1,2,3}` 3). 카드의 "66건" 과 1 차이이며, 그 사이 원격 상태가 움직인 결과로 본다.
- 이 65건은 **원격을 건드리지 않는다** — 로컬 포인터만 지운다. 그래서 "되돌리기 가장 어려운 동작" 이라는 원격 삭제의 경고는 여기 해당하지 않는다.
- 그런데 **`WT-*` 스테일 58건 중 54건은 develop 의 조상이 아니다**(`--no-merged=develop` 58건에서 라이브 축1 실패 4건을 뺀 값). 이 ref 들을 prune 하면 그 커밋을 가리키는 마지막 포인터가 사라져 **gc 대상이 될 수 있다** — 다른 ref 가 잡고 있지 않는 한.
- 즉 위험 구조가 뒤집혀 있다: 원격 삭제는 시끄럽고 되돌리기 어렵지만 내용은 develop 이 갖고 있다. prune 은 조용하고 로컬 전용이지만 **내용을 실제로 소각할 수 있다.**

[HARD] 따라서 Stage C 를 재개하더라도 `fetch --prune` 을 원격 삭제와 같은 단계에 묶지 말 것. prune 은 54건 각각에 대해 "다른 ref 가 이 커밋을 잡고 있는가" 를 먼저 물은 뒤에 하는 별도 작업이다.

---

## 6. 재개 조건 (승인 요청 사항)

1. 삭제 후보 11건 원격 삭제 — 위 목록에 대한 운영자 승인.
2. 보존 2건(`WT-hook-wiring-drift`, `WT-precommit-vet`)은 t216·t237 이 닫힐 때까지 손대지 않는다.
3. `fetch --prune` 은 별도 카드로 분리 — 54건 orphan 위험 분류가 선행.
4. 실행 시점에 세 ref SHA 를 다시 못박고 축1·축2 를 재측정한다. 이 문서의 판정은 `131daa290` 트리에 귀속된다.

---

## Evidence

| # | 명령 | 결과 |
|---|---|---|
| V1 | `git rev-parse origin/main origin/develop develop HEAD` | 위 SHA 표 |
| V2 | `git ls-remote --heads origin 'refs/heads/WT-*'` | 13행 |
| V3 | `git for-each-ref --format='…' 'refs/remotes/origin/WT-*'` | 71행 |
| V4 | `git for-each-ref --merged=origin/develop 'refs/remotes/origin/WT-*'` | 13행 |
| V5 | `git for-each-ref --merged=develop 'refs/remotes/origin/WT-*'` | 13행 (V4 와 동일 목록) |
| V6 | `git for-each-ref --no-merged=develop 'refs/remotes/origin/WT-*' \| wc -l` | 58 |
| V7 | `git remote prune --dry-run origin` | 65 would-prune |
| V8 | `git diff --name-only origin/main...origin/WT-<x>` ×4 | 저작 경로 18 / 22 / 1 / 4 |
| V9 | `git diff --name-status develop..origin/WT-<x> -- <정확한 경로 목록>` ×4 | 위 A/M 표 |
| V10 | `git diff develop..origin/WT-main-stamp-repair -- .moai/project/codemaps/provenance.json` | 3 insertions / 3 deletions, 스탬프 3필드 |
| V11 | `git grep -c "rule-id keying" develop -- .moai/specs/SPEC-ASTGREP-LANG16-001/design.md` | `develop:…:1` |
| V12 | `git for-each-ref --format='…' 'refs/heads/WT-*'` | 로컬 64건, 삭제 후보와 이름 겹침 1건 |

## Baseline-attribution

전부 이 실행에서, 워크트리 `.claude/worktrees/t438`, 트리 `131daa290`, `git fetch origin` 직후에 측정했다. V4 와 V5 가 같은 목록을 낸 것은 `origin/develop`(`f7cabfc296`)과 로컬 `develop`(`131daa290`) 사이 32커밋에 `WT-*` tip 을 조상으로 만드는 병합이 없었다는 뜻이며, 두 기준의 축1 결과가 우연히 일치한 것이지 두 ref 가 같아서가 아니다.

## Gaps

명시적으로 관측하지 **않은** 것:

- 삭제 후보 11건 각각에 대한 `git log` 내용 검토. 축1 통과 9건은 조상 판정만으로 근거가 충분하다고 보아 개별 커밋을 읽지 않았다.
- 축2 `M` 행의 **일반** 판별 절차를 기계화하지 않았다. 이번 4건은 눈으로 내용을 봐서 갈랐다. 블롭 도달성(`git log --find-object`) 기반 기계 판별은 설계만 하고 실행하지 않았다 — 브랜치당 경로 수만큼 히스토리 스캔이라 비용이 크다.
- stale `WT-*` 54건 각각의 orphan 여부. "다른 ref 가 잡고 있는가" 를 건별로 재지 않았다. 54라는 수만 셌다.
- 비-`WT-*` prune 후보 7건(`release/v3.1.{1,2,3}` 포함)의 안전성. `WT-*` 축만 봤다.
- `git gc` 가 실제로 어떤 객체를 회수할지. `reflog` 만료 정책과의 상호작용을 재지 않았으므로 §5 의 소각 위험은 **구조적 논증이지 실측이 아니다.**

## Residual-risk

- **축2 `M` 판정은 사람 판단이다.** `WT-astgrep-16-langs` 를 삭제 가능으로 본 근거는 "R1 문단의 주제가 develop 에도 있고 develop 이 728줄 앞선다" 이지, 브랜치의 그 문단이 develop 어딘가에 **동일하게** 남아 있다는 증명이 아니다. 원격 삭제는 되돌리기 어렵고 이 건은 로컬 쌍둥이가 없다.
- **13이라는 모집단은 측정 시점 값이다.** 다른 레인이 push 하면 늘어난다. 실행 전 재측정이 §6.4 인 이유다.
- **보존 2건의 근거는 큐 상태에 의존한다.** t216·t237 이 DROPPED 로 닫히면 보존 근거가 사라지고 두 브랜치도 후보가 된다. 그때 이 문서를 근거로 재사용하지 말고 다시 잴 것.
- **`--merged` 는 tip 조상만 본다.** squash 로 착지한 브랜치는 tip 이 조상이 아니므로 축1 을 실패하고 축2 로 내려온다 — 설계대로다. 다만 축1 통과가 "내용 존재" 의 충분조건인 것과 달리, 축1 실패는 "내용 부재" 의 근거가 전혀 아니다. 실측 4건 중 2건이 그 경우였다.
