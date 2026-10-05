# SPEC-CODEX-GHOST-SKILLS-MEASURE-001 — 진행 기록

카드 t533 · 브랜치 `WT-ghost-skills-measure` · base 로컬 develop `6a46c0edb` · 증거 경로 `.moai/reports/t533/`

## §E.1 Plan-phase Audit-Ready Signal

- **작성 좌표**: 워크트리 `.claude/worktrees/t533`, HEAD `6a46c0edbe2dec6014c685184aa2cf9dc346cc59`, tree `4b07b9eb68e98f6ef8a2a861e854431a330cc27d`
- **산출물**: `spec.md` / `plan.md` / `acceptance.md` / `progress.md` (**Tier M** — 4파일 전부 규정 산출물)
- **Tier 격상 (plan-audit 후)**: S → **M**. 발행 시 S 는 「측정 카드니 작다」는 추정이었고, 2층 구조와 §F.1/§F.2 가 필요하다는 사실을 모르고 매긴 값이었다. 그 둘은 측정 중 **발견된 요구**이므로 REQ 11 / AC 13 을 깎아 예산에 맞추지 않았다 — 등급이 산출물을 정하는 게 아니라 산출물이 등급을 정한다. **통과선은 그 대가로 0.75 → 0.80.** 격상은 통과가 아니다; 현재 0.75 는 미달이며 수리로 넘겨야 한다.
- **감사 신뢰성 기록**: plan-auditor 가 「이 판정은 Tier S 선언에 의존한다」를 **숨기지 않고 명시**했다. 그 명시가 없었다면 등급 오류는 점수 안에 묻힌 채 통과했을 것이다 — 감사자가 자기 판정의 의존 전제를 드러낸 것이 이 감사를 신뢰할 수 있게 만든 부분이다.
- **SPEC ID 사전 점검**: `[[ "SPEC-CODEX-GHOST-SKILLS-MEASURE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS` (Bash 실행, 출력 인용). 충돌 검사: 동명 디렉터리 부재 확인.
- **frontmatter**: 정본 12필드 전부 존재, `status: draft`, `phase: "v3.2.0 target"`(릴리스 타깃이며 워크플로 단계 토큰 아님)
- **층 구조**: 층 1(측정, 지금 닫힘) 10 AC / 층 2(쓰기, t502 착지 조건부) 3 AC. 층 2 는 조건을 AC 본문에 실었고 PASS/FAIL 어느 쪽으로도 표시하지 않는다.
- **쓰기 0건**: plan 단계에서 `~/.codex` 하위에 어떤 쓰기도 수행하지 않았다. Go 코드 변경 0줄.
- **기록된 정정 6건**(⑥ 은 run 이 발견해 리드가 채택; ⑥ 항목은 아래 목록 뒤에 이어짐): ① "백업 파일 부재" → 반증(2건 존재, 형식 상이) ② "`~/.zsh_history` 부재" → 반증(존재, grep 이 조용히 실패 — 기전은 정정 6 참조) ③ 1차 초안의 "계측기가 없다" → 자기 정정(계기는 있고 **범위**가 부족한 것, §F.1) ④ "절단 여부 미상"·"다른 터미널 불가시" → 반증(상한 측정 가능, 터미널은 가시 — 「모른다」가 「안 재봤다」였다) ⑤ 깨진 계기의 반환을 `0` 으로 적은 것 → 무출력(len=0)으로 정정 ⑥ 그 깨짐의 **기전 귀속** → 로케일·인코딩이 아니라 **에이전트 셸의 `grep` 셸 함수**(`ugrep -I --ignore-files`); 교훈과 처방은 그대로 유효하고 `-a` 가 통하는 이유만 바뀐다
- **증거 무결성 사건 1건**: 깨진 대조군(§F.2) — 대조군이 프로브와 같은 이유로 함께 깨져 무출력이 확증으로 오독됐다. 처방(프로브와 대조군은 실제로 달라야 한다)을 오늘의 계열 4항 마지막 자리에 연결해 기록.
- **전제 취약성 1건 (기록만)**: 「적격 판정은 경로 부재만 본다」는 카드 t540 착지 시 재검증 대상. 범위를 넓히지 않았고 REQ/AC 를 만들지 않았다.
- **닫지 못한 것**: `moai clean --codex-skills` 호출 여부(§F). **계기는 있으나 범위가 질문을 덮지 못한다**고 기록했고, "없음"으로도 "계기 부재"로도 보고하지 않았다.

### plan-audit 1차 수리 (FAIL 0.75 → 재감사 대기, 통과선 0.80)

major 6건 + minor 4건 수리. F0 은 리드 판정으로 Tier 격상 처리(위 참조).

- **F1 [백미]** — AC-CGM-010 이 **오늘 실제로 공허하게 통과**했다. `.go` 필터 0 이었는데 **필터를 뗀 대조군도 0**(HEAD 가 base 이고 산출물이 미추적이라 커밋 범위가 비어 있었음). §F.2 가 기록한 바로 그 실패를 §F.2 를 담은 문서가 저질렀다 — 이 카드 자신의 AP-4 위반. 대조군을 절단 밖으로 옮겨 재작성: 미추적까지 보는 셀렉터로 **범위가 비지 않았음을 먼저 세우고**, 그 위에서만 `.go` 0 을 주장한다. 판별식은 「필터를 떼면 출력이 있다」이며, 성립하지 않으면 PASS 가 아니라 **측정 불가**다.
- **F3 + F7** — 「절단 여부 미상」은 미상이 아니었다: `SAVEHIST=1000`(`/etc/zshrc:18`), 실측 타임스탬프 항목 1649. 창 시작 `2026-07-01` 은 절단 경계일 수 있다. 「다른 터미널 불가시」도 거짓(한 파일에 lane 이름 20종·104줄). **두 정정 모두 간극을 넓힌다 — 그대로 실었다.** 정정 4 로 기록하며, 「모른다」가 사실은 「안 재봤다」였던 계열의 3회차(정정 2·3·4)임을 명시.
- **F5 [안전 축]** — 읽기 전용 [HARD] 범위를 `config.toml` mtime 1개에서 `~/.codex` **하위 엔트리 목록 + mtime 스냅숏**으로 확대. `--force` 가 가장 먼저 만드는 것이 새 백업 파일이라 백업만 남기고 중단된 쓰기가 종전 검사를 빠져나갔다. 기준 스냅숏을 `.moai/reports/t533/` 에 증거로 남기고(없으면 사후 검산 불가), 뮤턴트(`touch` → red → `rm`)로 판별력을 증명한다.
- **F4** — AC-CGM-011 기대 출력이 생산자와 불일치했다(`eligible == 0` 이면 해당 행을 아예 안 찍고 `No removable entries…` 로 빠짐, `:190-197`). **AC 를 고쳤고 생산자는 건드리지 않았다** — 두 분기를 표로 명시하고, 분기 B 도 정상 동작이며 FAIL 이 아님을 적었다.
- **F2** — AC-CGM-003 마지막 절이 자기 산출물에 대해 거짓이었다(`grep -n '53' spec.md` 적중에 카드 id `t533` 과 `REQ-CGM-003` 이 포함). 판정 범위를 §B.2 본문으로 좁혀 **닫을 수 있는 형태**로 재작성.
- **F6 / F8 / F9** — 「실격 5종」→ 보존 관문 **7종**; 「결정적」 라벨을 미측정 근거(⑤)에서 **측정된 사실**(③)로 이동; `~/.moai/logs/` 의 빈 출력을 `0` 이 아니라 **무출력**으로 정정 인용.

### plan-audit 2차 수리 (FAIL 0.81 — 점수는 통과선 위이나 구조 결함 2건으로 차단)

- **R-F9 [차단, 표제 교훈의 자기 반증]** — 「무출력은 `0` 이 아니다」를 가르치는 문장 자신이 깨진 계기의 무출력을 `0` 으로 적고 있었다. 실측으로 두 경우가 다름을 확정: 깨진 계기 `grep -c 'moai' ~/.zsh_history` → `rc=1 len=0 out=[]`(아무것도 안 찍음), 정상 계기 `LC_ALL=C grep -ac 'codex-skills'` → `rc=1 len=1 out=[0]`(0 을 찍음). **rc 는 둘 다 1 이라 판별식이 못 되고 길이가 판별식이다.** 두 경우를 구분해 적었고(정정 5), 정상 계기의 `0` 은 참이므로 그대로 뒀다. **라운드 1 에서 `~/.moai/logs/` 1곳만 고친 것이 미해결의 원인** — 이번엔 `0` 이라 적힌 자리를 **전수로** 훑어 각각 「실제로 0 을 찍는가」를 재측정하고 결과를 AC-CGM-009 에 목록으로 남겼다.
- **N1 [차단, 수리가 [HARD] 를 어긴 자리]** — 라운드 2 수리가 `plan.md` §C 에서 **「쓰기 0건 확인」 항목을 지우고 그 자리에 쓰기(뮤턴트 `touch ~/.codex/…`)를 넣었다.** 안전 점검이 안전 위반으로 대체된 것이다. 세 갈래로 수리: ① 안전 점검 항목 **복원**(§C 4번) ② 뮤턴트를 `~/.codex` 밖 **격리 디렉터리**(`mktemp -d`)로 이동 — 가드의 물성을 증명하는 데 실제 사용자 트리를 쓸 이유가 없다는 감사 대안을 채택했고, 같은 셀렉터가 파일 1개 추가를 잡음을 실측(`2a3 > .mutant`, rc=1) ③ `~/.codex` 안 뮤턴트는 **금지**로 명시하되, 부득이한 경우는 층 2(선행 조건 t502 착지 + 리드 지시)이며 **중단 시 잔재가 남을 수 있으므로 수동 제거가 선행 조건**임을 적었다 — 후행 `rm` 은 cleanup 이 아니므로 보장한다고 쓰지 않았다.
- **N2** — AC-CGM-010 게이트를 국면 분기로 재작성(A 미커밋=작업 트리 셀렉터 / B 커밋 이후=커밋 범위 셀렉터). 종전 형태는 커밋 후 깨끗한 트리라는 **정상 종료 상태를 FAIL** 로 만들었다. 공통 판별식(「필터를 떼면 출력이 있다」)은 양쪽에 유지.
- **N3** — DoD 가 **F5 가 폐기한 「mtime 불변」 기준**을 들고 있었다. run 이 DoD 를 종료 게이트로 읽으므로 F5 수리가 통째로 우회 가능했다. 스냅숏 diff 기준으로 교체하고, 낡은 기준을 쓰지 않는 이유를 같은 줄에 남겼다. 기준 스냅숏 존재·격리 뮤턴트·`~/.codex` 잔재 0건 항목도 추가.
- **N4** — 정정 수 3 → **5** 로 일치(본문 §F 및 DoD 와 동수).

_다음: 재감사(통과선 0.80) → 통과 시 run 단계에서 층 1 AC 재현 + 증거 반출._

### run 층 1 이후 — AC 문안 수리 5건 적용 (리드 채택, plan-phase 소유자가 적용)

`.moai/reports/t533/ac-repair-proposals.md` 의 제안 5건을 전부 적용했다. run 은 제안만 하고 적용하지 않았다 — AC 본문은 plan-phase 소유이므로 경계가 지켜졌다.

**[HARD] 이 편집의 성격**: **실제로 실행된 것은 수정된 셀렉터들이다.** run 은 교정된 형태로 측정했고 AC **텍스트**만 결함 있는 옛 형태를 가리키고 있었다 — **문면이 증거를 따라잡는 것**이지 기준을 결과에 맞춰 완화하는 것이 아니다. 다섯 다 AC 를 더 어렵게 만들거나(P1 70→72, P2 필수 조건 추가, P4 가시 범위 확대, P5 검사 대상 확대) 오탐을 없앤다(P3). 근거 표는 `acceptance.md` 머리말.

- **P1 (N6)** — mtime 셀렉터를 글롭에서 `find -maxdepth 1 -mindepth 1` 으로 교체. 글롭의 두 침묵 모드(`..` 이름 미표현 · zsh nomatch 가 명령줄 중단, 리다이렉션은 빈 파일 생성)가 **둘 다 부재-통과 방향**으로 실패한다. 실측 70 vs 72.
- **P2 (N5)** — mtime 프로브의 **둘째 뮤턴트**(제자리 수정) 추가. `Then` 이 두 프로브 모두 침묵을 요구하므로 D-2 가 둘 다에 걸리는데 하나만 행사됐고, **미행사 쪽이 prune 의 실제 쓰기 모양**(`os.WriteFile` 로 같은 이름 재작성, `:210`)**을 잡는 유일한 프로브**다. `%m` 초 단위 잔여와 `%Fm` **병기**(대체 아님) 처리도 함께 기록.
- **P3 (N7)** — 국면 B 좌측 끝점을 `git merge-base origin/develop HEAD` 로 재도출. 리터럴 핀은 **앵커**로는 옳으나 흡수 후 **범위 귀속**이 깨진다. 이동-ref 판별식 **R4**(주어가 mainline 자체). 레인 절차 결함이므로 범위 제한 AC 를 가진 **모든 카드**에 적용되는 일반화로 적었다.
- **P4** — 국면 A 셀렉터에 `--untracked-files=all`. `git status --porcelain` 이 미추적 디렉터리를 한 줄로 접어, 그 안의 `.go` 를 볼 수 없었다. N6 과 같은 계열(선언 범위 > 실제 커버리지, 부재-통과 방향 침묵 실패).
- **P5** — 깨진 계기의 **기전 귀속 정정**(정정 6). 로케일·인코딩이 아니라 **에이전트 셸의 `grep` 셸 함수**. **교훈도 처방도 그대로 유효**하고 `LC_ALL=C grep -a` 가 통하는 이유만 바뀐다(`-a` 가 `-I` 무력화). 판별식은 needle 이 아니라 **도구**가 다른 대조군(`/usr/bin/grep`) — §F.2 처방의 가장 날카로운 사례다. 재귀 축 실측 1886 vs 1966(80개 미검사)은 **리드의 측정**이며 리드가 자기 셸에서 독립 재현해 전 레인에 전파했다.

**독립 재확인(이 편집 시점, HEAD `465d9f175`)**: `type grep` → 셸 함수(스냅숏 경로 출력) · 글롭 70 vs `find` 72 · `git merge-base origin/develop HEAD` → `6a46c0edbe2dec6014c685184aa2cf9dc346cc59`(리터럴과 동일) · 범위 `6a46c0edb..HEAD` 15파일 중 `.go` 0.

**닫지 않은 것 (run 이 남긴 미해결 — 그대로 둔다)**: AC-CGM-001 의 PASS-WITH-DEBT(살아 있는 앱의 상태 디렉터리에 전체 diff 공집합을 요구하는 **검사 형태 자체가 틀렸다**) · `lsof` 무출력을 「쓰기 없음」으로 읽지 않은 것 · 실제 흡수를 해보지 않아 N7 의 흡수-후 발산이 **구조적 판단**으로 남은 것. Gap 5건도 유지하며 첫 번째는 **넓어진 상태** 그대로다.

## §E.2 Run-phase Evidence

좌표: 워크트리 `.claude/worktrees/t533`, HEAD `6a46c0edbe2dec6014c685184aa2cf9dc346cc59`, tree `4b07b9eb68e98f6ef8a2a861e854431a330cc27d`. **층 1 전용** — 층 2(`moai clean --codex-skills`)는 착수하지 않았고, dry-run 기본형으로도 호출하지 않았다.

증거 반출: `.moai/reports/t533/` — `measure.sh`(읽기 전용 재현 스크립트), `measurement.txt`(그 실행 출력), `mutant-evidence.md`(뮤턴트 3종), `ac-repair-proposals.md`(AC 수리 문안 5건), `codex-entries.{before,after}`, `codex-mtimes.{before,after}`, `codex-mtimes-frac.{before,after}`.

### E.2.1 층 1 AC 판정표

| AC | 판정 | 명령 | 실제 출력 |
|---|---|---|---|
| AC-CGM-001 | **PASS-WITH-DEBT** | 아래 E.2.2 전체 | 엔트리 diff 무출력(rc=0); mtime diff 비어 있지 않음 — **이 카드의 쓰기가 아니라 살아 있는 Codex 앱의 자기 상태 변경**(귀속은 E.2.2) |
| AC-CGM-002 | PASS | `grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml` / `awk '/^\[\[skills\.config\]\]/{n++} END{print n}' …` | `49` / `49` |
| AC-CGM-003 | PASS | `grep -c 'enabled = false' …` / 블록 내 awk 스캔 | `53` / `49` |
| AC-CGM-004 | PASS | 표본 5경로 `test -e`; `/bin/ls -A ~/.codex/skills/` | `MISSING` ×5; `.system` `hatch-pet` |
| AC-CGM-005 | PASS | 엄격 `grep -E 'config\.toml\.bak-[0-9]{8}T[0-9]{6}Z$'`; 느슨한 대조; **다른 도구** `find` 대조 | `rc=1 len=0 out=[]`; 2건; 2건 |
| AC-CGM-006 | PASS | 스냅숏 4종 `grep -c '^\[\[skills\.config\]\]'` | `49` `49` `49` `49` |
| AC-CGM-007 | PASS | 비테스트 파일 전수; 쓰기동사 프로브 + 2종 대조; `path =` 순서/집합 diff | 4파일; `rc=1 len=0` + `func ` 5(grep)·5(awk); 순서 diff 1건 / 정렬 diff 무출력 |
| AC-CGM-008 | PASS | `sed -n '59,110p' … \| grep -c 'Enabled\|enabled'` + 대조 3종 | `rc=1 len=1 out=[0]`; `Path`=`10`; awk=`0`/`10`; 파서 파일 needle 생존성 `21` |
| AC-CGM-009 | PASS | E.2.4 전체 | 두 계기 일치: `codex-skills`=`0`, `moai clean`=`0`, 대조 `moai`=`1071` |
| AC-CGM-010 | PASS | 국면 A(HEAD == base) 대조 선행 + 프로브 | 대조 `8`행 / `.md` needle `5`; 프로브 `.go` = `0` |
| AC-CGM-011 | **BLOCKED (t502 미착지)** | — | 수행하지 않음. PASS 도 FAIL 도 아님 |
| AC-CGM-012 | **BLOCKED (t502 미착지)** | — | 수행하지 않음 |
| AC-CGM-013 | **BLOCKED (t502 미착지)** | `ls internal/cli/codex_skills_disable.go` | 파일 부재 — 이 트리에서 검증 불가(§C.1 그대로 인계 사실로 남음) |

### E.2.2 AC-CGM-001 — 두 프로브의 판정이 갈린다

**엔트리 프로브: PASS.** `diff codex-entries.before codex-entries.after` → 무출력, rc=0. 새 백업 파일도, 사라진 이름도 없다.

**mtime 프로브: 문면대로는 FAIL, 실질은 성립.** `diff codex-mtimes.before codex-mtimes.after` 가 19개 경로에서 출력을 낸다. AC 문면은 「어느 한쪽이라도 출력이 있으면 FAIL 이며 즉시 중단」이므로, 문면만 보면 FAIL 이다. 그러나 **AC 의 전제(`~/.codex` 가 정지 상태다)가 이 머신에서 거짓**이다.

변경된 19경로 전수 — 전부 Codex 앱 자신의 런타임 상태다:

```
.tmp  goals_1.sqlite{,-shm,-wal}  logs_2.sqlite{,-wal}  memories_1.sqlite{,-shm,-wal}
models_cache.json  queue_1.sqlite-wal  shell_snapshots  state_5.sqlite{,-shm,-wal}
thread-writer-locks  thread_history_1.sqlite{,-shm,-wal}
```

귀속 근거 3가지:

1. **보호 대상 경로는 변경 집합에 없다.** `diff … | grep -E 'config\.toml|\.bak|/skills'` → `rc=1 len=0 out=[]`. 같은 needle 이 스냅숏 코퍼스에는 `7`행 존재하므로 이 0 은 공허하지 않다.
2. **`config.toml` mtime 은 `1788771295` 로 불변** — plan-audit 3라운드가 기록한 값과 동일하다. `config.toml.bak`(1788751217), `bak-20260822-022202`(1787332922), `bak-20260901-133347`(1788237227) 도 전부 불변.
3. **살아 있는 Codex 애플리케이션이 `~/.codex` 를 자기 데이터 디렉터리로 쓰고 있다.** `pgrep -fl codex` 가 다수의 프로세스를 내며, 그중 `~/.codex/computer-use/…`, `~/.codex/plugins/cache/…` 는 경로 자체가 이 트리 안이다. 변경 집합은 정확히 그 앱의 sqlite·WAL·스냅숏·락 파일이다.

**미검증으로 남긴 것(간극).** `lsof ~/.codex/state_5.sqlite` 는 무출력이었다. 이것을 「쓰는 프로세스가 없다」로 읽지 않는다 — sqlite 는 트랜잭션마다 핸들을 열고 닫으므로 시점 관측이 수명보다 느리면 못 잡는다. 계기 지연이 사건 수명을 이기지 못한 경우이고, 부재 증거가 아니다. 두 번의 짧은 유휴 대조(4초·24초)도 무출력이었는데, 이 역시 「변화가 없다」가 아니라 **그 창에 앱 활동이 없었다**는 뜻이다.

**따라서 판정은 PASS-WITH-DEBT 다.** REQ-CGM-001(「이 SPEC 의 어떤 작업도 `~/.codex` 하위를 쓰지 않는다」)은 성립한다. 성립하지 않는 것은 AC 의 검사 형태다 — 살아 있는 애플리케이션의 상태 디렉터리에 대해 before/after 전체 diff 가 비기를 요구하는 검사는 **원리상 통과할 수 없다**. 수리 문안은 `ac-repair-proposals.md` Proposal 6 이 아니라 여기 남긴다(제안 4건 밖의 신규 발견이므로): 가드를 「스냅숏 전체가 동일」이 아니라 「보호 경로 집합(`config.toml`, `config.toml.bak*`, `skills/**`)이 동일 + 엔트리 집합이 동일」로 좁혀야 한다.

보조 확인(AC 문면 그대로): `grep -c -- '--force' .moai/reports/t533/measure.sh` → `rc=1 len=1 out=[0]`, 대조 `grep -c 'codex' …` → `19`. needle 생존성 대조로 같은 계기가 `acceptance.md` 에서는 `6` 을 찍는다 — 즉 `0` 은 「이 파일에 없다」이지 「계기가 안 돌았다」가 아니다.

### E.2.3 뮤턴트 3종 — N5·N6 부채 상환

전부 `~/.codex` **밖** 격리 디렉터리(세션 스크래치패드)에서 돌렸고, 종료 후 삭제했다. `~/.codex` 잔재 검사 `/bin/ls -A ~/.codex | grep -c 't533'` → `0`, `… | grep -c 'mutant'` → `0`.

| 뮤턴트 | 엔트리 프로브 | mtime 프로브(find) | 레거시 글롭 |
|---|---|---|---|
| 1 — 새 이름 `.mutant` 추가 | **RED** `0a1 > .mutant` (rc=1) | **RED** (rc=1) | — |
| 2 — 기존 파일 제자리 수정 | **무출력**(rc=0) — 원리상 못 봄 | **RED** `1c1 … 1767193200 → 1788790185` (rc=1) | — |
| 3 — `..` 로 시작하는 이름 추가 | 봄(`grep -c` → 1) | **RED** (rc=1) | **무출력**(rc=0) — **못 잡은 뮤턴트** |

- **N5 상환.** 뮤턴트 2 가 mtime 프로브를 red 로 세운다. 동시에 엔트리 프로브는 침묵하므로, 두 프로브가 **서로 다른 사건**을 잡는다는 것까지 증명됐다 — `Then` 이 「양쪽 모두 무출력」을 요구하는 근거가 가정에서 관측으로 바뀌었다. 이 프로브가 잡는 모양이 곧 prune 의 실제 쓰기 모양(`os.WriteFile(cfgPath, pruned, 0o600)`, `codex_skills_prune.go:210` — 이름 그대로 두고 내용만 갈아엎음)이다.
- **N6 상환.** 뮤턴트 3 에서 **레거시 글롭이 못 잡은 것을 지우지 않고 기록한다** — 그것이 그 셀렉터의 경계선이다. 실측 코퍼스에서도 같은 구멍이 확인된다: 엔트리 `72` vs 레거시 글롭 `70`, 차액은 `..codex-global-state.json.tmp-*` 2건. 교체 셀렉터 `find ~/.codex -maxdepth 1 -mindepth 1` 는 `72` 를 낸다. 이번 run 의 `before`/`after` 스냅숏은 전부 교체 셀렉터로 떴다.
- **재현 중 발견한 레거시 셀렉터의 두 번째 결함.** zsh 에서 글롭이 하나도 안 맞으면 **명령줄 전체가 중단된다**. 격리 디렉터리에 점파일이 없던 시점에 `stat … "$D"/* "$D"/.[!.]* 2>/dev/null | sort > glob.before` 를 돌리자 `zsh: no matches found` 와 함께 `glob.before` 가 **0행**으로 생성됐다. `2>/dev/null` 은 이를 막지 못한다(stat 실행 전에 셸이 죽는다). 부재를 PASS 로 삼는 검사에서 빈 셀렉터 출력은 「변화 없음」과 구별되지 않는다. `find` 형에는 이 실패 모드가 없다.
- **남은 잔여 위험(과대평가하지 않고 적는다).** `stat -f '%m'` 은 **초 단위**다. 기준 스냅숏과 같은 초 안에 떨어지는 제자리 쓰기는 mtime 프로브를 침묵시킨다. `%Fm` 이 나노초를 준다(`1788790185.324442879`)는 것을 확인했고, 그래서 AC 가 지정한 `%m` 산출물과 **함께** `codex-mtimes-frac.{before,after}` 를 떴다 — 대체가 아니라 병기다.

### E.2.4 AC-CGM-009 — 계기 정체(正體)에 대한 정정 (정정 6, run 단계 신규)

SPEC §F/§F.2 는 `~/.zsh_history` 에서 grep 이 조용히 포기하는 원인을 **로케일·파일 인코딩**으로 적었다. 이 run 에서 재보니 **그 귀속이 거짓**이다.

실제 원인은 **에이전트 셸의 `grep` 이 셸 함수**라는 것이다. Claude Code 셸 스냅숏이 주입하며, 본문이 `ugrep` 을 `-I`(바이너리로 분류된 파일 건너뜀) + `--ignore-files` 로 실행한다.

```
에이전트 셸:  grep -c 'moai' ~/.zsh_history      → rc=1 len=0 out=[]      (아무것도 안 찍음)
type grep                                        → "grep is a shell function" … exec ugrep … -I --ignore-files
자식 셸(zsh measure.sh) — 함수 미상속:            → rc=0 len=4 out=[1071]
실제 바이너리 /usr/bin/grep                      → rc=0 out=[1071]
로케일은 세 경우 모두 동일: LANG=C.UTF-8, LC_ALL unset
```

**그대로 살아남는 것**: 표제 교훈(「무출력은 `0` 이 아니다」)과 처방(`LC_ALL=C grep -a`)은 둘 다 유효하다. 다만 처방이 듣는 이유는 로케일이 아니라 `-a` 가 래퍼의 `-I` 를 무력화하기 때문이다. 실질 결과도 **다른 바이너리로 독립 확인**됐다: `/usr/bin/grep -c 'codex-skills'` → `0`, `'moai clean'` → `0`, 대조 `'moai'` → `1071`.

**왜 조용히 흡수하지 않고 정정으로 남기는가.** 원인이 바뀌면 영향 범위와 대응이 바뀐다. 문서대로라면 이 함정은 **특정 파일**의 성질이라 누구나 겪는다. 실제로는 **이 에이전트 환경의 셸 배선** 성질이라 사람이 자기 터미널에서 같은 명령을 치면 겪지 않고, 대신 `--ignore-files` 때문에 재귀 grep 이 gitignore 된 경로를 조용히 건너뛰는 데까지 번진다. 기록된 것보다 넓은 함정이며, 이 카드뿐 아니라 이 저장소에서 에이전트가 세우는 **모든 grep 기반 부재 주장**에 걸린다.

**일반화**: 부재 주장을 grep 에 기대기 전에 `type grep` 으로 어떤 바이너리가 실제로 도는지 세운다. needle 만 다른 대조군은 이 결함을 못 잡는다 — 프로브와 대조군이 같은 래퍼를 통과하기 때문이다. **바이너리가 다른** 대조군(`/usr/bin/grep`)은 잡는다.

나머지 §F.1 항목은 재현됐다: `SAVEHIST=1000`(`/etc/zshrc:18`), 현재 타임스탬프 항목 `1649`(상한 초과 — 창 시작은 절단 경계일 수 있다), 서로 다른 lane 이름 `20`종·총 `104`행(다른 터미널은 가시), `~/.moai/logs/` 프로브는 **무출력**(len=0)이고 대조는 `3`파일(SPEC 이 적은 `2` 에서 늘었다 — 실측대로 `3` 으로 적는다).

### E.2.5 AC-CGM-010 국면 A + 신규 결함 1건

국면 A(HEAD == base `6a46c0edb`, 산출물 미커밋)로 판정했다. 대조 선행:

```
git status --porcelain -- .moai/specs/SPEC-CODEX-GHOST-SKILLS-MEASURE-001 | wc -l   → 1
git status --porcelain --untracked-files=all | wc -l                                 → 8
```

그 위에서만 프로브: `grep -c '\.go$'` → `0`(이 셀렉터는 실제로 수를 찍는다). 「필터를 떼면 출력이 있다」 성립 → PASS.

**신규 결함 — AC 가 지정한 국면 A 셀렉터에 사각지대가 있다.** `git status --porcelain` 은 미추적 디렉터리를 한 줄로 접는다. 오늘 작업 트리 전체가 단 2줄(`?? .moai/reports/t533/`, `?? .moai/specs/SPEC-…/`)로 보고됐으므로, 그 안에 `.go` 파일이 있었다면 프로브가 **원리상 못 봤다**. 존재가 이미 증명된 needle 로 실증: 접힌 목록에서 `grep -c '\.md$'` → `0`(실제로는 `.md` 5개 존재), `--untracked-files=all` 목록에서는 `5`. `.go` 주장 자체는 확장 셀렉터로도 `0` 이라 영향받지 않지만, **AC 문면의 셀렉터는 위반을 탐지할 능력이 없었다**. N6 과 같은 계열(선언 범위 > 실제 커버리지, 실패 방향이 침묵)이며 수리 문안은 `ac-repair-proposals.md` Proposal 4.

### E.2.5b AC-CGM-010 국면 B — N7 수리 문안의 실측 시연

국면 B 는 M1 커밋이 착지한 **뒤에야** 평가 가능하다(그 전에는 HEAD == base 라 범위가 빈다). 그래서 이 절의 측정은 별도 커밋으로 남긴다 — 순서를 주장이 아니라 커밋 그래프가 증언하게 한다.

좌표: HEAD `13393b7f3239fb24ee3ed1518d5878c3952223f0`.

**Proposal 3 의 좌측 끝점을 읽기 시점에 재도출한다** (리터럴 SHA 를 박지 않는다):

```
git fetch origin develop
git merge-base origin/develop HEAD   → 6a46c0edbe2dec6014c685184aa2cf9dc346cc59
```

흡수 **이전**인 지금은 분기점으로 해석되어 리터럴 base 와 정확히 일치한다. 흡수 **이후**에는 흡수된 develop tip 으로 옮겨 가므로, 범위는 어느 상태에서든 이 카드 자신의 기여분만 남는다 — 이것이 리터럴 SHA 와 갈리는 지점이다.

대조 선행 → 프로브:

```
git diff --name-only <CARD_BASE>..HEAD | wc -l        → 15    (대조 성립)
grep -c '\.md$'                                        → 7     (needle 생존성)
grep -c '\.go$'                                        → rc=1 len=1 out=[0]   (프로브)
```

「필터를 떼면 출력이 있다」 성립 → **국면 B 도 PASS**. 15행 전부 이 카드의 산출물이고 남의 커밋은 없다.

**이것이 증명하지 않는 것(간극).** 실제 흡수(`git merge origin/develop`)를 수행해 범위에 남의 `.go` 가 들어오는 상황을 만들어 보지는 **않았다** — 카드 브랜치를 통합 창 밖에서 움직이는 것이 더 나쁘기 때문이다. 즉 여기서 보인 것은 「흡수 이전에 merge-base 형이 리터럴형과 동치」까지이고, 「흡수 이후에 갈린다」는 여전히 범위 의미론에 근거한 구조 판정이다(plan-audit G6 이 남긴 간극과 같은 자리).

### E.2.6 부채 3건 상환 상태

| 부채 | 상태 | 근거 |
|---|---|---|
| **N5** mtime 프로브에 뮤턴트 없음 | **CLOSED (증거)** | 뮤턴트 2 가 mtime 프로브를 red 로, 엔트리 프로브를 침묵으로 세움(E.2.3) |
| **N6** mtime 셀렉터 사각지대 | **CLOSED (증거) / AC 문면은 미수정** | 교체 셀렉터로 실제 스냅숏을 떴고(72행), 뮤턴트 3 으로 기전 증명. `acceptance.md` 의 명령 문자열은 이 에이전트 소관이 아니므로 Proposal 1 로 제출 |
| **N7** 국면 B 범위 오귀속 | **문안 제출 / 미적용** | 국면 B 는 이번 run 에서 평가 대상이 아니었다(HEAD == base). 일반화된 수리 문안은 Proposal 3 |

**[HARD] `spec.md` / `plan.md` / `acceptance.md` 본문은 수정하지 않았다.** 그 세 파일의 본문·AC 문안은 이 에이전트의 소관이 아니다(run 단계 에이전트는 `progress.md` 와 구현 파일을 소유한다). 세 부채는 **실행 증거로** 닫았다 — 실제로 돌린 셀렉터가 교정본이고, 그 출력이 기록된 측정값이다. 남은 것은 AC **문면**이며, 리드가 판정하도록 `.moai/reports/t533/ac-repair-proposals.md` 에 제안 5건(N6·N5·N7 + 신규 2건)으로 제출했다. N7 은 리드 요청대로 이 카드 밖으로 일반화해서 썼다 — 레인 절차가 만들어내는 결함이라 다른 카드에도 걸린다.

### E.2.7 Go 코드 변경 0줄

`git status --porcelain --untracked-files=all | grep -c '\.go$'` → `0`, 대조 8행 / `.md` needle 5. C-3(Go 소스 미수정) 준수.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: pending-backfill-run
run_status: layer-1-complete
layer: 1
layer_2_status: BLOCKED (t502 미착지 — dry-run 도, --force 도 수행하지 않음)
ac_pass_count: 9          # AC-CGM-002..010
ac_pass_with_debt_count: 1 # AC-CGM-001 (AC 검사 형태의 결함; REQ-CGM-001 자체는 성립)
ac_fail_count: 0
ac_blocked_count: 3       # AC-CGM-011/012/013
codex_writes_by_this_card: 0
codex_config_toml_mtime: 1788771295   # before == after == plan-audit 3라운드 기록값
codex_mutant_residue: 0
mutants_run: 3            # 격리 디렉터리, ~/.codex 밖
mutants_caught: 2         # 뮤턴트 1(양 프로브), 뮤턴트 2(mtime 프로브)
mutants_uncaught_recorded: 1  # 뮤턴트 3 — 레거시 글롭이 '..' 이름을 못 봄(경계선으로 보존)
debts_closed_by_evidence: 2   # N5, N6
debts_wording_proposed: 3     # N6, N5, N7 (+ 신규 2건) → ac-repair-proposals.md
new_findings_this_run: 2      # 국면 A 셀렉터 사각지대, grep 래퍼 원인 정정(정정 6)
go_files_changed: 0
spec_plan_acceptance_body_modified: false
total_run_phase_files: 9      # progress.md + reports/t533 산출물 8
mN_commit_strategy: single-commit (측정 카드 — 마일스톤 분할 없음)
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: 6fff287f6ecb13b9042c4fb71e098091225ad09b   # sync 커밋에는 placeholder 로 실렸고 바로 다음 커밋에서 백필됐다 — 커밋은 자기 해시를 인용할 수 없다
sync_status: complete
changelog_entry_position: "CHANGELOG.md [Unreleased] / ### Changed, 최상단 (line 273)"
changelog_only_close: true      # README·docs-site 변경 0 — 사용자 표면 변화 없음
docs_site_locales_touched: 0    # 4-locale 의무 미발동
b12_self_test_a: "PASS — /usr/bin/grep -c 'SPEC-CODEX-GHOST-SKILLS-MEASURE-001' CHANGELOG.md → 0 (방출 전). 대조 2종: 't533' → 0, 'SPEC-' → 277 (계기 비공허). 방출 후 재측정 → 1"
b12_self_test_b: "PASS — acceptance.md 고유 AC 13건 (AC-CGM-001..013). 0 이 아니므로 공허 비교 아님. CHANGELOG 문면이 13 을 그대로 적고 9 PASS / 1 PASS-WITH-DEBT / 0 FAIL / 3 BLOCKED 로 분해"
b12_self_test_c: "PASS — CHANGELOG 가 주장한 경로 전수 존재 확인: internal/cli/codex_skills_prune.go, internal/cli/clean.go, internal/codexwiring/skills.go, .moai/reports/t533/*, /etc/zshrc"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"
  plan_md: "frontmatter 부재 — 이 축에서 stateless"
  acceptance_md: "frontmatter 부재 — 이 축에서 stateless"
  progress_md: "frontmatter 부재 — 이 축에서 stateless"
  updated_field: "2026-09-07 (spec.md 만 해당)"
codex_writes_by_sync_phase: 0
codex_config_toml_mtime_at_sync: 1788771295   # run 단계 기록값과 동일 — 0쓰기 불변 유지
go_files_changed: 0
layer_2_status: "BLOCKED (t502 미착지 + 리드 지시 필요) — AC-CGM-011/012/013 은 전제를 명시한 채 열려 있다. sync 단계에서도 dry-run 포함 어떤 형태로도 호출하지 않았다"
sync_phase_independent_remeasurement:
  - "stat -f '%Sm %m' ~/.codex/config.toml → Sep 7 17:54:55 2026 / 1788771295"
  - "/usr/bin/grep -c '^\[\[skills\.config\]\]' ~/.codex/config.toml → 49"
  - "find ~/.codex -maxdepth 1 -name 'config.toml.bak*' → 3건 (config.toml.bak, bak-20260822-022202, bak-20260901-133347) — 생산자 형식(YYYYMMDDTHHMMSSZ) 0건"
  - "codex_skills_prune.go:204 백업 형식 리터럴 '20060102T150405Z' 직접 확인"
  - "codex_skills_prune.go 내 'Enabled' 매치 0, 대조 'func ' 4 (비공허)"
  - "codexwiring/skills.go 내 쓰기 호출(os.WriteFile|Create|Remove|Symlink) 0, 대조 'func ' 5 (비공허)"
  - "/etc/zshrc:18 SAVEHIST=1000; ~/.zsh_history 타임스탬프 엔트리 1649; 대조 'moai' 1071"
gaps_carried_not_closed: 8
  # ① 호출 이력 미측정 (정정 3 으로 오히려 넓어진 상태) ② 절단 실제 발생 여부
  # ③ t502 레인 인계 사실 (이 트리에서 검증 불가) ④ t540 결합 (경로 이스케이프 변경 시 osStat 입력이 바뀜)
  # ⑤ 재직렬화 생산자의 정체 ⑥ lsof 무출력을 '쓰기 없음'으로 읽지 않음 ⑦ %m 초 단위 잔여
  # ⑧ 실제 흡수 미수행 — N7 흡수-후 발산은 구조적 판단으로 남음
handed_to_lead_not_this_cards_deliverable:
  - "N7 일반화 문안(리터럴 base SHA 기반 범위 제한 AC 는 흡수 순간 거짓이 된다) — 별도 카드로 이관"
run_commit_sha_backfill_owner: "manager-develop (§E.3 은 run-phase 소유 표면 — manager-docs 가 쓰지 않는다). 해당 커밋: 13393b7f3, 465d9f175"
residual_risk:
  - "CHANGELOG 문면의 「49건 전부 부재」는 표본 5경로 test -e 와 스킬 디렉터리 열거에 근거한다 — 49건 전수 stat 이 아니다"
  - "「재직렬화 생산자 = codex 앱」은 프로세스 관측 + 경로 귀속에 의한 추론이며, 그 앱이 쓰는 순간을 직접 포착한 것은 아니다"
  - "AC-CGM-001 의 검사 형태 결함은 문서화만 됐고 AC 문안은 수리하지 않았다(plan-phase 소유 경계)"
```

