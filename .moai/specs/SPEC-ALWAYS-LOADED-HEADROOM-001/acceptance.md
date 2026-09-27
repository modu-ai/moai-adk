# SPEC-ALWAYS-LOADED-HEADROOM-001 — 인수 조건

모든 AC 는 Given-When-Then 이며 이진 판정된다. 명령은 워크트리 루트에서 실행하며, `$SCRATCH` 는 커밋 트리 밖의 세션 임시 디렉터리를 가리킨다. 판정서는 `.moai/reports/t1226/verdict.md` 다. 아래 명령의 `BH` 는 판정서 `build_head` 줄의 값이다.

**설계 원칙 — 뼈대만으로는 어떤 AC 도 통과하지 않는다.** plan 단계가 커밋한 판정서 뼈대에는 아래 AC 들이 요구하는 기계 줄이 하나도 없다. 측정이 없으면 모든 MUST-PASS 가 FAIL 한다. 또한 `S_init` 의 수치는 판정서에 적힌 값을 믿지 않고 **하네스를 `BH` 에서 다시 돌려** 재현한다(AC-ALH-002·004).

---

## §D. AC 매트릭스

| AC | 대상 REQ | 판정 | 심각도 |
|---|---|---|---|
| AC-ALH-001 | REQ-ALH-001 | 기계적 | MUST-PASS |
| AC-ALH-002 | REQ-ALH-002, REQ-ALH-015 | 기계적 | MUST-PASS |
| AC-ALH-003 | REQ-ALH-003, REQ-ALH-004, REQ-ALH-005, REQ-ALH-014 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-004 | REQ-ALH-006 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-005 | REQ-ALH-007, REQ-ALH-008 | 기계적 | MUST-PASS |
| AC-ALH-006 | REQ-ALH-009, REQ-ALH-015 | 기계적 | MUST-PASS |
| AC-ALH-007 | REQ-ALH-010, REQ-ALH-011, REQ-ALH-012, REQ-ALH-015 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-008 | REQ-ALH-013 | 기계적 | MUST-PASS |
| AC-ALH-009 | REQ-ALH-016 | 기계적 | MUST-PASS |
| AC-ALH-010 | REQ-ALH-015 | 기계적 + 검토 | MUST-PASS |

### 판정서 기계 줄 규약

판정서는 아래 줄을 **이 철자 그대로**, 줄머리에서 시작해 한 줄씩 적는다. 정수는 쉼표 없이 쓴다. `<s>` 는 `init` 또는 `live` 다.

| 줄 | 값 |
|---|---|
| `레인 백엔드: <관측값> (출처: <관측 방법>)` | run 레인이 관측한 서빙 모델과 그 관측 방법 |
| `charmap = <값>` | `locale charmap` 출력 |
| `build_head = <40자리 hex>` | 측정 시점 `git rev-parse HEAD` |
| `harness_sha256 = <64자리 hex>` | 격리 하네스 테스트 파일의 `shasum -a 256` |
| `count_set_init = 18` / `17` / `observed` | `S_init` 계수 집합. 경로 목록은 `.moai/reports/t1226/count-set-init.txt`(한 줄에 하나, 프로젝트 루트 상대경로). `S_live` 는 18경로 고정이며 `.moai/reports/t1226/count-set-live.txt` |
| `total_<s> = N` | 표면 계수 집합의 `wc -m` 합계 |
| `hash_<s> = <64자리 hex>` | 표면 계수 집합 중 마크다운 파일의 동결 다중집합 sha256 — 선행 SPEC AC-ALD2-002 [REF] 의 파이프라인을 표면 트리에서 실행 |
| `current_<s>` · `A_adm_<s>` · `R_<s>` · `U_<s>` · `T_min_<s>` ` = N` | AC-ALH-006 |
| `P절 재조정: (a)` / `(b)` / `(공표값)` / `(기타)` | AC-ALH-005 |
| `J_includes_kanban_scope = yes` / `no` | AC-ALH-005 |
| `F(172ef22eb) = N` | AC-ALH-005 — 서술값 |
| `verdict_<s> = <토큰>` | AC-ALH-007 |
| 형태 1: `runtime_files_init = N` · `runtime_total_init = <런타임 표기 그대로>` · `runtime_isolated = yes` · `runtime_source = <관측 방법>` | AC-ALH-010 |
| 형태 2: `runtime_observed = no` 와 17집합 줄 `total_init_17` · `current_init_17` · `A_adm_init_17` · `R_init_17` · `U_init_17` · `T_min_init_17` · `verdict_init_17` | AC-ALH-010 — `count_set_init = 18` 일 때만 |
| `verdict_init_reason = count-set` | 형태 2 에서 `verdict_init_17` 이 18집합 규칙 토큰과 다를 때만 |
| `missing_init = <경로 …>` / `missing_init = -` | `S_init` 에 없는 18경로(§D.3). 없으면 `-` |

### 후보 표 TSV 열 규약

두 TSV(`.moai/reports/t1226/candidates-init.tsv`, `.moai/reports/t1226/candidates-live.tsv`)는 `.moai/reports/t1226/candidates.py` 가 생성하고 판정 열을 사람이 채운다. 행은 해당 표면 계수 집합의 파일만 담는다. 머리 줄 하나에 이어 후보 행이며 열은 다음 순서다.

| 열 | 이름 | 값 |
|---:|---|---|
| 1 | `file` | 계수 집합의 경로 |
| 2 | `section` | 절 제목. 서문은 `(서문)`, yaml 파일은 `(전체)`, 구속 절 안의 비구속 문단은 `<절 제목> ¶n` |
| 3 | `gross` | 후보 전체 자수(정수, 스크립트 산출). 줄 끝 개행을 포함해 세며, 한 파일의 서문·절·`(전체)` 행 합이 그 파일 `wc -m` 과 같도록 분할한다(절은 겹치지 않는 평면 분할) |
| 4 | `chars` | 허용 자수(정수). `ADMIT` 이 아니면 `-` |
| 5 | `mech` | `M1` · `M1p` · `M2` · `none` |
| 6 | `bind` | 후보 안 구속 조항 줄 개수(정수, 스크립트 산출) |
| 7 | `gov` | 조건 1 단서 — `N`(해당 없음) · `a`(구속 조항이 의존) · `b`(구속 조항 범위를 좁힘) |
| 8 | `c1` | 조건 1 `Y`/`N` |
| 9 | `c2` | 조건 2 `Y`/`N`/`NA` |
| 10 | `c3` | 조건 3 `Y`/`N`/`NA` |
| 11 | `c4` | 조건 4 `Y`/`N`/`NA` |
| 12 | `dest` | M1 목적지 경로(`.claude/rules/moai/` 기준 상대경로). 그 밖은 `-` |
| 13 | `verdict` | `ADMIT` · `REJECT` · `UNTRIED` |
| 14 | `reason` | 아래 열거 |
| 15 | `evidence` | 증거 파일 경로(커밋된 `.moai/reports/t1226/` 아래) |

`reason` 열거와 대응 조건 — 밖의 값은 FAIL:

| reason | 쓰는 판정 | 반드시 성립 |
|---|---|---|
| `-` | `ADMIT` | — |
| `bind>0` | `REJECT` | `bind > 0` |
| `governs-scope` | `REJECT` | `gov == a` |
| `narrows-scope` | `REJECT` | `gov == b` |
| `citation-breaks` | `REJECT` | `c2 == N` |
| `dest-unreachable` | `REJECT` | `c3 == N` |
| `dest-over-40k` | `REJECT` | `c4 == N` |
| `agents-md-no-m1` | `REJECT` | `file` 이 `AGENTS.md`, `mech` 가 `M2` 아님 |
| `config-data` | `REJECT` | `file` 이 `.yaml` |
| `net-negative` | `REJECT` | `mech == M1`, `bind == 0`, `gov == N`, `c1~c4 == Y`, 증거 파일에 `pointer_chars = P` 줄이 있고 `P ≥ gross` |
| `rewraps-binding-line` | `REJECT` | `mech == M2`, 증거 파일의 `post_hash` 가 표면 해시와 다름(AC-ALH-004) |
| `untried:<사유>` | `UNTRIED` | `mech == M2`, 사유 비어 있지 않음 |

보조 표 둘:
- `.moai/reports/t1226/dest-sizes.tsv` — 열 `dest`, `size_live`, `size_tmpl`(각 트리에서 `wc -m`).
- `.moai/reports/t1226/pointers-<s>.tsv` — 열 `file`, `dest`, `pointer_chars`. `ADMIT` M1 행의 (file, dest) 쌍마다 정확히 한 행.

---

### AC-ALH-001 — 판정서 머리와 폐기·서술 수치의 격리

**Given** run 단계가 판정서를 채웠을 때,
**When** 아래 명령을 실행할 때,
**Then** 앞 네 명령이 각각 1 이상, 뒤 두 명령이 각각 0 을 출력한다.

```bash
head -20 .moai/reports/t1226/verdict.md | grep -c -E '^레인 백엔드: [^_ ][^(]* \(출처: [^_][^)]*\)$'
head -20 .moai/reports/t1226/verdict.md | grep -c -F '7fe658815eb0d4110b9acadad56e5a85bee3ed3f'
head -20 .moai/reports/t1226/verdict.md | grep -c -E '199,?111'
head -20 .moai/reports/t1226/verdict.md | grep -c -E '49,?111'
grep -E '197,?897|198,?361' .moai/reports/t1226/verdict.md | grep -v -c '폐기'
grep -c -E '^F = ' .moai/reports/t1226/verdict.md
```

백엔드 줄은 **형식만** 검사한다 — 값은 run 레인의 관측이며, 자리표시(`_미관측_`)로 시작하면 첫 명령이 0 을 낸다. 관측 방법의 적정성은 검토 항목이다.

FAIL 조건: 위 기대와 다른 출력이 하나라도 있다.

### AC-ALH-002 — 두 표면의 합계를 재실행으로 재현한다

**Given** 판정서의 `build_head`(`BH`)와 두 계수 집합 파일이 있을 때,
**When** `BH` 트리에서 `S_live` 를 다시 재고, AC-ALH-009 의 하네스를 `BH` 에서 다시 돌려 `$SCRATCH/init-surface` 를 만든 뒤 `S_init` 을 다시 잴 때,
**Then** 두 재측정값이 판정서의 `total_live`·`total_init` 과 같고, `charmap = UTF-8` 이며, `## S_init`·`## S_live` 절 안에 자리표시가 0건이다.

`S_live` 측정 블록(기준선 `199111 total`, `7fe658815`):

```bash
locale charmap
wc -m CLAUDE.md AGENTS.md \
  .moai/config/sections/user.yaml .moai/config/sections/language.yaml \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/moai-constitution.md \
  .claude/rules/moai/core/moai-mcp-tools.md \
  .claude/rules/moai/core/native-idiom-and-register.md \
  .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/kanban-dispatch.md \
  .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/skill-routing.md | tail -1
```

판정 명령:

```bash
V=.moai/reports/t1226/verdict.md
grep -c -E '^charmap = UTF-8$' "$V"                                   # 1
grep -c -E '^build_head = [0-9a-f]{40}$' "$V"                         # 1
grep -c -E '^count_set_init = (18|17|observed)$' "$V"                 # 1
# 18경로 목록 고정 — 줄 수가 아니라 내용을 대조한다
P18="CLAUDE.md AGENTS.md .moai/config/sections/user.yaml .moai/config/sections/language.yaml
.claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/askuser-protocol.md
.claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/moai-mcp-tools.md
.claude/rules/moai/core/native-idiom-and-register.md .claude/rules/moai/core/verification-claim-integrity.md
.claude/rules/moai/workflow/cache-aware-execution.md .claude/rules/moai/workflow/context-window-management.md
.claude/rules/moai/workflow/cross-session-messaging.md .claude/rules/moai/workflow/goal-directive.md
.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/moai/workflow/main-checkout-branch-guard.md
.claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/skill-routing.md"
printf '%s\n' $P18 | sort | diff - <(sort .moai/reports/t1226/count-set-live.txt) && echo PIN-live-OK      # PIN-live-OK
# count_set_init = 18 이면 같은 목록에서 §D.3 누락 경로(판정서 missing_init = 줄에 공백 구분)만 뺀 것과 같아야 한다
# count_set_init = 17 이면 위 목록에서 skill-routing.md 만 뺀 것과 같아야 한다
# S_live 재현 — BH 트리
mkdir -p "$SCRATCH/live" && git archive BH CLAUDE.md AGENTS.md .moai/config/sections .claude/rules/moai | tar -x -C "$SCRATCH/live"
(cd "$SCRATCH/live" && xargs wc -m < "$OLDPWD/.moai/reports/t1226/count-set-live.txt" | tail -1)   # == total_live
# S_init 재현 — BH 에서 AC-ALH-009 하네스 재실행 후
(cd "$SCRATCH/init-surface" && xargs wc -m < "$OLDPWD/.moai/reports/t1226/count-set-init.txt" | tail -1)   # == total_init
grep -E '^total_(live|init) = [0-9]+$' "$V"
awk '/^## S_init/,/^## S_live/' "$V" | grep -c '_미측정'               # 0
awk '/^## S_live/,/^## P절 재조정/' "$V" | grep -c '_미측정'           # 0
```

`BH` 가 `7fe658815` 과 다르고 그 사이 흡수가 18경로를 바꿨다면(`git diff --quiet 7fe658815 BH -- <18경로>` 가 0 이 아님), 판정서는 그 사실과 새 `total_live` 를 적고 기준선 199,111 은 이력으로만 남긴다. `count-set-init.txt` 의 내용은 위 고정 목록 대조를 따른다 — `count_set_init` 이 `18`·`17` 이면 줄 수는 그 수에서 §D.3 누락 경로 수(판정서 `missing_init = <경로 …>` 줄, 누락이 없으면 `missing_init = -`)를 뺀 값이다.

FAIL 조건: 두 재측정값 중 하나가 판정서 값과 다르거나, 위 기대와 다른 출력이 하나라도 있다.

### AC-ALH-003 — 후보 표의 재생성·분할 완결성·허용 규칙

**Given** 두 TSV, `candidates.py`, `dest-sizes.tsv`, AC-ALH-002 에서 재현한 두 표면 트리가 있을 때,
**When** (1) 스크립트로 키 집합을 다시 뽑아 TSV 와 비교하고, (2) 파일별 분할 합을 `wc -m` 과 대조하고, (3) 행 규칙 `awk` 를 돌리고, (4) 목적지 누적 수용량을 검사할 때,
**Then** (1) diff 가 비고, (2)~(4) 가 모두 0 을 보고한다.

```bash
# (1) 키 재생성 — (file, section, gross, bind)
for s in init live; do
  root="$SCRATCH/init-surface"; [ "$s" = live ] && root="$SCRATCH/live"
  python3 .moai/reports/t1226/candidates.py --keys --paths ".moai/reports/t1226/count-set-$s.txt" "$root" | sort > "$SCRATCH/k-$s"
  tail -n +2 ".moai/reports/t1226/candidates-$s.tsv" | cut -f1,2,3,6 | sort | diff - "$SCRATCH/k-$s" && echo "KEYS-$s-OK"
done

# (2) 분할 완결성 — 파일마다 (서문·절·(전체) 행의 gross 합) == wc -m. 문단 행(¶)은 제외
for s in init live; do
  root="$SCRATCH/init-surface"; [ "$s" = live ] && root="$SCRATCH/live"
  awk -F'\t' 'NR>1 && $2 !~ / ¶[0-9]+$/ {g[$1]+=$3} END {for (f in g) print f"\t"g[f]}' ".moai/reports/t1226/candidates-$s.tsv" | sort > "$SCRATCH/g-$s"
  (cd "$root" && xargs wc -m < "$OLDPWD/.moai/reports/t1226/count-set-$s.txt" | grep -v ' total$' | awk '{print $2"\t"$1}' | sort) > "$SCRATCH/w-$s"
  diff "$SCRATCH/g-$s" "$SCRATCH/w-$s" > /dev/null && echo "SPLIT-$s-OK"
done

# (3) 행 규칙
for s in init live; do
  f=".moai/reports/t1226/candidates-$s.tsv"
  awk -F'\t' 'NR>1 {
    if (NF!=15) bad++;
    if ($13!="ADMIT" && $13!="REJECT" && $13!="UNTRIED") bad++;
    if ($13=="ADMIT" && ($6+0>0 || $7!="N" || $8!="Y" || $14!="-" || $4 !~ /^[0-9]+$/)) bad++;
    if ($13=="ADMIT" && $5=="M1" && ($9!="Y"||$10!="Y"||$11!="Y"||$12=="-")) bad++;
    if ($1 ~ /AGENTS\.md/ && $5!="M2" && $13=="ADMIT") bad++;
    if ($13!="ADMIT" && $4!="-") bad++;
    if ($13=="UNTRIED" && ($5!="M2" || $14 !~ /^untried:.+/)) bad++;
    if ($13=="REJECT") {
      ok=0;
      if ($14=="bind>0" && $6+0>0) ok=1;
      if ($14=="governs-scope" && $7=="a") ok=1;
      if ($14=="narrows-scope" && $7=="b") ok=1;
      if ($14=="citation-breaks" && $9=="N") ok=1;
      if ($14=="dest-unreachable" && $10=="N") ok=1;
      if ($14=="dest-over-40k" && $11=="N") ok=1;
      if ($14=="agents-md-no-m1" && $1 ~ /AGENTS\.md/ && $5!="M2") ok=1;
      if ($14=="config-data" && $1 ~ /\.yaml$/) ok=1;
      if ($14=="net-negative" && $5=="M1" && $6+0==0 && $7=="N" && $8=="Y" && $9=="Y" && $10=="Y" && $11=="Y" && $12!="-") ok=1;
      if ($14=="rewraps-binding-line" && $5=="M2") ok=1;
      if (!ok) bad++;
    }
    if ($3 !~ /^[0-9]+$/ || $15=="") bad++;
    if ($13=="ADMIT" && $4+0 > $3+0) bad++;                         # DEBT-1: chars <= gross
    if ($13=="ADMIT" && $5 ~ /^M1/ && $4+0 != $3+0) bad++;          # DEBT-1: M1·M1p 는 후보 전체를 옮기거나 지운다
  } END {print "ROWS-'"$s"' BAD="bad+0}' "$f"
  # DEBT-2: 같은 절의 절 행과 문단 행이 동시에 ADMIT 이면 이중 계상
  awk -F'\t' 'NR>1 && $13=="ADMIT" {b=$2; isp=sub(/ ¶[0-9]+$/,"",b); k=$1 SUBSEP b; if (isp) p[k]=1; else s[k]=1} END {for (k in s) if (k in p) bad++; print "OVERLAP-'"$s"' BAD="bad+0}' "$f"
  # DEBT-3: ADMIT M1 의 목적지가 그 표면 계수 집합 안에 있으면 옮긴 문자가 합계에 남는다
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1" {print ".claude/rules/moai/"$12}' "$f" | sort -u \
    | grep -x -F -f ".moai/reports/t1226/count-set-$s.txt" | wc -l | awk '{print "DESTIN-'"$s"' BAD="$1}'
  # M1p 증거는 남는 중복 원본의 경로를 dup_source = <경로> 줄로 담는다
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1p" {print $15}' "$f" | while read -r p; do grep -q -E '^dup_source = .+' "$p" || echo "m1p-bad $p"; done | wc -l | awk '{print "M1P-'"$s"' BAD="$1}'
  # net-negative: 증거의 pointer_chars >= gross
  awk -F'\t' 'NR>1 && $14=="net-negative" {print $3"\t"$15}' "$f" | while IFS="$(printf '\t')" read -r g p; do
    P=$(grep -E '^pointer_chars = [0-9]+$' "$p" | awk '{print $3}'); [ -n "$P" ] && [ "$P" -ge "$g" ] || echo "netneg-bad $p"
  done | wc -l | awk '{print "NETNEG-'"$s"' BAD="$1}'
  # 모든 행의 증거 파일 존재
  tail -n +2 "$f" | cut -f15 | sort -u | while read -r p; do [ -f "$p" ] || echo "missing $p"; done | wc -l | awk '{print "EVID-'"$s"' MISSING="$1}'
done

# (4) 목적지 누적 수용량 — 두 트리 각각, 합이 40000 미만
for s in init live; do
  awk -F'\t' 'FNR==NR {live[$1]=$2; tmpl[$1]=$3; next}
    FNR>1 && $13=="ADMIT" && $5=="M1" {add[$12]+=$4}
    END {for (d in add) { if (!(d in live) || live[d]+add[d]>=40000 || tmpl[d]+add[d]>=40000) bad++ } print "DEST-'"$s"' BAD="bad+0}' \
    .moai/reports/t1226/dest-sizes.tsv ".moai/reports/t1226/candidates-$s.tsv"
done
```

검토 항목: `SPEC-ALWAYS-LOADED-DIET-002/design.md §4.3` 기각 표의 절 중 이 트리에 남아 있는 것은 모두 `gov` 가 `a`/`b` 이거나, `N` 이면 증거 파일이 그 판단을 뒤집는 근거를 담는다.

DEBT-3 귀결: 계수 집합 안의 파일을 목적지로 삼는 M1 은 허용되지 않으므로, 17집합 재계산(AC-ALH-006 `init_17`)에서 제외 파일 `skill-routing.md` 가 목적지인 행은 18집합에서 이미 FAIL 이다 — 17집합 전용 규칙은 필요 없다.

FAIL 조건: `KEYS-*-OK`·`SPLIT-*-OK` 넷 중 하나가 없다, `BAD`·`MISSING` 값(`ROWS`·`OVERLAP`·`DESTIN`·`M1P`·`NETNEG`·`EVID`·`DEST`) 중 하나라도 0 이 아니다.

### AC-ALH-004 — 표면 해시를 재현하고, M2 자수는 실제 압축 시도로 쟀다

**Given** 판정서의 `hash_init`·`hash_live` 와 AC-ALH-002 에서 재현한 두 표면 트리가 있을 때,
**When** 동결 다중집합 파이프라인을 두 트리에서 다시 돌리고, M2 행의 증거 파일에서 시도 후 해시를 읽을 때,
**Then** 재실행 해시가 판정서 값과 같고, `ADMIT` M2 행은 모두 표면 해시와 같으며 `rewraps-binding-line` 행은 모두 다르다.

```bash
V=.moai/reports/t1226/verdict.md
# 파이프라인: 표면 계수 집합 중 .md 파일에 대해 선행 SPEC AC-ALD2-002 [REF] 와 같은 grep|sed|sort|shasum
for s in init live; do
  root="$SCRATCH/init-surface"; [ "$s" = live ] && root="$SCRATCH/live"
  (cd "$root" && grep '\.md$' "$OLDPWD/.moai/reports/t1226/count-set-$s.txt" | xargs grep -hE '\[HARD\]|MUST|shall ' \
     | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | sort | shasum -a 256 | awk '{print $1}')   # == hash_<s>
done
grep -c -E '^hash_(init|live) = [0-9a-f]{64}$' "$V"                          # 2
# BH 가 18경로를 7fe658815 이후 바꾸지 않았다면 hash_live 는 d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

for s in init live; do
  H=$(grep -E "^hash_${s} = " "$V" | awk '{print $3}')
  awk -F'\t' 'NR>1 && $5=="M2" && ($13=="ADMIT" || $14=="rewraps-binding-line") {print $13"\t"$15"\t"$3"\t"$4}' \
    ".moai/reports/t1226/candidates-${s}.tsv" \
  | while IFS="$(printf '\t')" read -r v p g c; do
      got=$(grep -E '^post_hash = ' "$p" | awk '{print $3}')
      if [ "$v" = "ADMIT" ] && [ "$got" != "$H" ]; then echo "bad $p"; fi
      if [ "$v" != "ADMIT" ] && [ "$got" = "$H" ]; then echo "bad $p"; fi
      if [ "$v" = "ADMIT" ]; then   # DEBT-1: M2 허용 자수는 증거의 전후 자수 차와 같다
        pre=$(grep -E '^pre_chars = [0-9]+$' "$p" | awk '{print $3}'); post=$(grep -E '^post_chars = [0-9]+$' "$p" | awk '{print $3}')
        if [ -z "$pre" ] || [ -z "$post" ] || [ "$pre" != "$g" ] || [ "$c" != "$((pre - post))" ]; then echo "chars-bad $p"; fi
      fi
    done
done | wc -l | awk '{print "HASH-BAD="$1}'                                  # HASH-BAD=0
```

증거 파일은 scratch 사본에 가한 압축 diff 와 `post_hash = <sha256>` 줄을 담고, `ADMIT` M2 행은 추가로 `pre_chars = N`(압축 전 후보 자수, `gross` 와 같음)과 `post_chars = M`(압축 후 자수)을 담는다. 그 행의 `chars` 는 `N − M` 이어야 한다(DEBT-1). `hash_init` 은 `S_init` 트리에서 **실측한** 값이며 라이브 해시를 옮겨 적지 않는다. 검토 항목: 압축이 의무·조건·예외·수치를 지우지 않았다(REQ-ALD2-012 준용).

FAIL 조건: 재실행 해시가 판정서 값과 다르다, `HASH-BAD` 가 0 이 아니다, 해시 줄이 2개가 아니다.

### AC-ALH-005 — `sec.py` 출처와 `P절` 재조정

**Given** `.moai/reports/t1226/sec.py` 가 커밋돼 있을 때,
**When** 해시를 재고, 원 트리 `172ef22eb` 에서의 재실행 출력과 판정서 줄을 읽을 때,
**Then** 아래 기대가 모두 성립한다.

```bash
shasum -a 256 .moai/reports/t1226/sec.py | awk '{print $1}'
#   d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547
V=.moai/reports/t1226/verdict.md
grep -c -E '^P절 재조정: \((a|b|공표값|기타)\)$' "$V"                         # 1
grep -c -E '^J_includes_kanban_scope = (yes|no)$' "$V"                        # 1
grep -c -E '^F\(172ef22eb\) = [0-9]+$' "$V"                                    # 1
ls .moai/reports/t1226/sec-172ef22eb.txt                                       # 존재 — 재실행 출력 원문
cut -f15 .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv | grep -c -E '^(/tmp|/private/tmp)'   # 0
grep -c -E '^evidence: *(/tmp|/private/tmp)' "$V"                              # 0
```

`/tmp` 금지는 **증거 인용**(TSV `evidence` 열과 판정서의 `evidence:` 줄)에만 걸린다. 실행한 명령을 기록할 때 임시 디렉터리는 `$SCRATCH` 로 적는다. 판정서의 F 절은 「F 는 원 트리의 서술값이며 M2 항에 수율 외삽을 포함한다, `T_min`·판정의 입력이 아니다」를 명시한다(검토). 재조정이 `(기타)` 면 새 산식을 함께 적는다.

FAIL 조건: 해시가 다르다, 기대 개수와 다른 줄이 하나라도 있다, 재실행 출력 파일이 없다. 원본이 사라져 스크립트를 재작성한 경우는 해시 불일치로 FAIL 이며, 그때는 리드에게 블로커로 올린다.

### AC-ALH-006 — `T_min` 산술이 계수 집합 기준으로 표에서 재현된다

**Given** 판정서 기계 줄, 계수 집합 파일, TSV·포인터 표가 있을 때,
**When** 계수 집합에 속한 행만으로 각 항을 다시 계산할 때,
**Then** 표면마다 `CHECK-<s>=OK` 와 `PAIRS-<s>=OK` 가 나오고, 17집합 줄이 있으면 `CHECK-init_17=OK` 도 나온다.

```bash
V=.moai/reports/t1226/verdict.md
SR=.claude/rules/moai/workflow/skill-routing.md
calc() {  # $1=표면, $2=제외 경로(없으면 빈 문자열), $3=줄 접미사
  s=$1; x=$2; sfx=$3
  A=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x && $13=="ADMIT" {s+=$4} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  U=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x && $13=="UNTRIED" {s+=$3} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  R=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x {s+=$3} END {print s+0}' ".moai/reports/t1226/pointers-${s}.tsv")
  C=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x && $2 !~ / ¶[0-9]+$/ {s+=$3} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  awk -v k="$sfx" -v A="$A" -v U="$U" -v R="$R" -v C="$C" '
    $1=="total_"k {tt=$3} $1=="current_"k {c=$3} $1=="A_adm_"k {a=$3} $1=="U_"k {u=$3} $1=="R_"k {r=$3} $1=="T_min_"k {t=$3}
    END { if (tt==C && c==C && a==A && u==U && r==R && t==C-A+R) print "CHECK-"k"=OK"; else print "CHECK-"k"=BAD" }' "$V"
}
calc init "" init
calc live "" live
grep -q -E '^T_min_init_17 = ' "$V" && calc init "$SR" init_17
for s in init live; do
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1" {print $1"\t"$12}' ".moai/reports/t1226/candidates-${s}.tsv" | sort -u > "$SCRATCH/pair-$s"
  tail -n +2 ".moai/reports/t1226/pointers-${s}.tsv" | cut -f1,2 | sort > "$SCRATCH/ptr-$s"
  diff "$SCRATCH/pair-$s" "$SCRATCH/ptr-$s" && echo "PAIRS-$s=OK"
done
```

`C`(계수 집합 합계)를 TSV 의 분할 행에서 다시 세므로, AC-ALH-003 (2) 와 결합해 `total_<s>` 가 계수 집합 `wc -m` 과 같음을 확립한다. 계수 집합의 정의는 `count_set_init` 한 줄과 `count-set-init.txt` 가 정하고, 이 AC 는 그 집합의 TSV 행만 쓴다 — 관측 집합(`observed`)에서도 같은 명령이 성립한다. 검토 항목: 포인터 줄 원문이 증거로 커밋돼 있고 `pointer_chars` 가 그 `wc -m` 과 같다.

FAIL 조건: 위 OK 줄 중 하나라도 없거나 `BAD` 가 나온다.

### AC-ALH-007 — 판정 토큰의 규칙 재계산과 상신 절차

**Given** 판정서 기계 줄이 있을 때,
**When** 토큰을 `T_min`·`U` 에서 다시 계산하고 상신 절을 읽을 때,
**Then** 표면마다(17집합 줄이 있으면 그것도) `TOKEN-<k>=OK` 이고, `verdict_init` 이 `ACHIEVABLE` 이 아니면 상신 절이 (a)~(d) 네 항목과 `RECOMMEND:` 문장을 담는다.

```bash
V=.moai/reports/t1226/verdict.md
rule() { awk -v k="$1" '
    $1=="T_min_"k {t=$3} $1=="U_"k {u=$3} $1=="verdict_"k {v=$3; n++}
    END { if (t<150000) e="ACHIEVABLE"; else if (t-u>=150000) e="STRUCTURALLY-INFEASIBLE-UNDER-FREEZE"; else e="UNDETERMINED";
          print e"\t"v"\t"n }' "$V"; }
for k in live init_17; do
  grep -q -E "^T_min_${k} = " "$V" || continue
  rule "$k" | awk -F'\t' -v k="$k" '{print ($3==1 && $1==$2) ? "TOKEN-"k"=OK" : "TOKEN-"k"=BAD" }'
done
# init: 18집합 규칙 토큰 E18, 17집합 토큰 V17 이 있으면 둘이 다를 때 UNDETERMINED + reason
E18=$(rule init | cut -f1); VI=$(rule init | cut -f2); N=$(rule init | cut -f3)
V17=$(grep -E '^verdict_init_17 = ' "$V" | awk '{print $3}')
if [ -n "$V17" ] && [ "$V17" != "$E18" ]; then
  [ "$N" = 1 ] && [ "$VI" = UNDETERMINED ] && grep -q -E '^verdict_init_reason = count-set$' "$V" && echo TOKEN-init=OK || echo TOKEN-init=BAD
else
  [ "$N" = 1 ] && [ "$VI" = "$E18" ] && echo TOKEN-init=OK || echo TOKEN-init=BAD
fi
# verdict_init 이 ACHIEVABLE 이 아닐 때
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '^### \((a|b|c|d)\)'    # 4
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '^RECOMMEND: '           # 1 이상
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '결정했다|승인했다|해제한다\.?$' # 0
```

`init` 의 토큰은 `count_set_init` 이 가리키는 집합의 `T_min_init`·`U_init` 에서 계산한다(`init_17` 은 형태 2 에서만 존재). `UNDETERMINED` 이면 상신 절이 `UNTRIED` 목록과 `U_init` 을 인용한다(검토).

FAIL 조건: `TOKEN-*=BAD` 가 하나라도 있다, 상신 절 항목 수가 4 가 아니다, `RECOMMEND:` 0건, 결정 서술 존재.

### AC-ALH-008 — 카드 계보의 커밋은 계수 파일과 미러를 건드리지 않았다

**Given** 기준 커밋 `7fe658815` 와 카드 HEAD 가 있을 때,
**When** 제1 부모 계보의 비병합 커밋만 대상으로 18경로와 그 미러의 이력을 볼 때,
**Then** 출력이 비어 있다.

```bash
git log --first-parent --no-merges --format=%H 7fe658815..HEAD -- CLAUDE.md AGENTS.md \
  .moai/config/sections/user.yaml .moai/config/sections/language.yaml \
  .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/moai-mcp-tools.md \
  .claude/rules/moai/core/native-idiom-and-register.md .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/workflow/cache-aware-execution.md .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/moai/workflow/main-checkout-branch-guard.md \
  .claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/skill-routing.md \
  internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl \
  internal/template/templates/.moai/config/sections/user.yaml.tmpl \
  internal/template/templates/.moai/config/sections/language.yaml.tmpl \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/askuser-protocol.md \
  internal/template/templates/.claude/rules/moai/core/moai-constitution.md \
  internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md \
  internal/template/templates/.claude/rules/moai/core/native-idiom-and-register.md \
  internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md \
  internal/template/templates/.claude/rules/moai/workflow/cache-aware-execution.md \
  internal/template/templates/.claude/rules/moai/workflow/context-window-management.md \
  internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md \
  internal/template/templates/.claude/rules/moai/workflow/goal-directive.md \
  internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md \
  internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md \
  internal/template/templates/.claude/rules/moai/workflow/session-handoff.md \
  internal/template/templates/.claude/rules/moai/workflow/skill-routing.md | wc -l      # 0
```

레인은 develop 을 카드 브랜치로 흡수하므로 카드 계보가 제1 부모다. `--first-parent` 없이 `--no-merges` 만 쓰면 흡수 병합의 제2 부모 쪽 — 다른 카드의 비병합 커밋 — 까지 세어, 이 카드가 아무것도 고치지 않아도 FAIL 한다(plan-audit 2회차 D1 의 양성 대조: 흡수 병합 하나에서 53행, `--first-parent` 로 0행). 흡수가 18경로를 바꿨다면 AC-ALH-002 의 규정대로 재측정한다. 한계: 카드 브랜치를 develop 쪽으로 fast-forward 하거나 흡수 방향을 뒤집으면 제1 부모 계보가 카드 계보가 아니게 된다 — 레인 규약(`git merge develop` 흡수)을 따르는 한 성립한다.

FAIL 조건: 출력이 0 이 아니다.

### AC-ALH-009 — `S_init` 은 격리 하네스로만 산출됐다

**Given** 격리 하네스 테스트 `TestHeadroomInitSurfaceExport` 가 `internal/cli/` 에 커밋돼 있을 때,
**When** 그 테스트를 앵커된 패턴으로 돌린 출력 파일과 판정서를 읽을 때,
**Then** 테스트가 PASS 했고(SKIP 이 아님), 실행 시점 HEAD 와 하네스 해시가 판정서와 같으며, 비격리 `moai init` 실행 기록이 없다.

하네스 실행 — 출력은 `.moai/reports/t1226/harness-run.txt` 로 커밋하며, 파일 첫 줄에 실행 직전 `git rev-parse HEAD` 결과를 `harness_head = <sha>` 로 적는다:

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run '^TestHeadroomInitSurfaceExport$' -count=1 -v -args -headroom-export="$SCRATCH/init-surface"
```

판정 명령:

```bash
H=.moai/reports/t1226/harness-run.txt
V=.moai/reports/t1226/verdict.md
grep -c -- '--- PASS: TestHeadroomInitSurfaceExport ' "$H"                    # 1
grep -c -- '--- SKIP: TestHeadroomInitSurfaceExport ' "$H"                    # 0
head -1 "$H" | awk '{print $3}'                                               # == 판정서 build_head
F=$(git ls-files 'internal/cli/*headroom*_test.go')
shasum -a 256 $F | awk '{print $1}'                                           # == 판정서 harness_sha256
grep -c 'prepareSafeInitHome' $F                                              # 1 이상
grep -c -E '^harness_sha256 = [0-9a-f]{64}$' "$V"                             # 1
test -s .moai/reports/t1226/commands.log && echo LOG-OK                        # LOG-OK
grep -c -E 'moai([[:space:]]+-[^[:space:]]+)*[[:space:]]+init([[:space:]]|$)' .moai/reports/t1226/commands.log   # 0
```

하네스 계약(검토): 테스트 헬퍼 `prepareSafeInitHome` 로 실제 홈의 네 쓰기 지점을 돌리고 전후 지문을 비교한 상태에서, `runInitWithFlags` 와 같은 방식으로 init 명령을 실행한다. 고정 플래그는 `--non-interactive`, `--root <t.TempDir()>`, `--name headroom-probe`, `--language go`, `--mode tdd`, `--llm claude` 이며 `--all` 을 주지 않는다(기본 slim 설치가 기본 사용자의 표면이다). 같은 이유로 하네스는 init 실행 전에 `t.Setenv("MOAI_DISTRIBUTE_ALL", "")` 로 전체 배포 환경변수를 비운다(`internal/cli/init.go` `shouldDistributeAll` 은 이 변수로도 전체 배포를 켠다). `BH` 가 현재 HEAD 와 다르면 `BH` 를 새 격리 워크트리(`moai cc -w` 계열 또는 `EnterWorktree`)로 열어 그 안에서 같은 명령을 돌린다. init 이 만든 **프로젝트 트리 전체**를 `-headroom-export` 인자가 가리키는 디렉터리로 복사하고(`.git` 제외), 18경로의 (경로, 존재 여부, sha256) 를 `t.Log` 로 출력한다. 인자가 비면 `t.Skip` 한다 — 그래서 위 판정은 SKIP 을 FAIL 로 친다.

`commands.log` 는 run 단계가 실행한 셸 명령을 한 줄씩 적은 **자기 신고** 기록이다(임시 경로는 `$SCRATCH` 표기). 정규식은 `HOME=… moai init`, `moai --debug init`, 공백이 여럿인 형태를 잡지만, 기록에서 빠진 실행은 잡지 못한다 — 이 한계는 판정서 Residual-risk 에 적는다. 감사 보고서처럼 `moai init` 을 **서술**하는 문서는 검사 대상이 아니다.

FAIL 조건: 위 기대와 다른 출력이 하나라도 있다. 비격리 실행(실제 홈을 쓰는 `moai init`)의 흔적은 그 자체로 FAIL 이다.

### AC-ALH-010 — 런타임 계수 집합과의 대조

**Given** 하네스가 내보낸 `S_init` 프로젝트 트리 전체가 있을 때,
**When** 사용자 범위 지시문이 없는 격리 `CLAUDE_CONFIG_DIR`·`HOME` 에서 그 트리 사본으로 런타임을 띄워 보고하는 always-loaded 파일 수와 합계를 관측하거나, 관측하지 못했음을 기록할 때,
**Then** 판정서가 아래 두 형태 중 정확히 하나를 갖는다.

```bash
V=.moai/reports/t1226/verdict.md
# 형태 1 — 격리 관측됨
grep -c -E '^runtime_files_init = [0-9]+$' "$V"; grep -c -E '^runtime_total_init = .+' "$V"
grep -c -E '^runtime_isolated = yes$' "$V"; grep -c -E '^runtime_source = .+' "$V"              # 각 1
grep -c -E '^count_set_init = observed$' "$V"                                                  # 1
# 형태 2 — 관측 안 됨
grep -c -E '^runtime_observed = no$' "$V"                                                      # 1
grep -c -E '^count_set_init = 18$' "$V"                                                        # 1
grep -c -E '^(total|current|A_adm|R|U|T_min)_init_17 = [0-9]+$' "$V"                           # 6
grep -c -E '^verdict_init_17 = (ACHIEVABLE|STRUCTURALLY-INFEASIBLE-UNDER-FREEZE|UNDETERMINED)$' "$V"  # 1
```

규칙:
- 형태 1: `count-set-init.txt` 는 런타임이 보고한 파일 목록이다(18경로 밖의 파일이 있으면 포함하고, 그 파일도 TSV 행을 갖는다). `runtime_files_init` 은 그 줄 수와 같고, `runtime_total_init` 은 `total_init` 과 1% 안에서 맞아야 하며 어긋나면 원인을 적는다(검토). 격리되지 않은 관측(`runtime_isolated` 가 `yes` 가 아님)은 형태 1 로 인정하지 않는다 — 관측자의 사용자 범위 지시문이 합계에 섞일 수 있다는 가설 때문이며, 이 가설은 관측으로 확인되지 않았다.
- 형태 2: 관측 불가 사실이 Gaps 에 있고, 17집합 줄은 AC-ALH-006·007 이 재계산한다. `verdict_init_17` 이 18집합 규칙 토큰과 다르면 `verdict_init = UNDETERMINED` 와 `verdict_init_reason = count-set` 이 있어야 한다.
- 관측 방법의 예: 리드 또는 운영자가 워크트리 밖 세션에서 `CLAUDE_CONFIG_DIR`·`HOME` 을 빈 임시 디렉터리로 돌리고 `S_init` 사본 디렉터리로 `claude` 를 띄워 기동 경고 줄을 원문으로 옮기는 것. 레인이 직접 띄울 수 없으면 리드에게 관측을 요청하는 블로커를 올리고, 답이 없으면 형태 2 로 닫는다.

FAIL 조건: 두 형태 중 어느 것도 완결되지 않았거나, 두 형태가 함께 있거나, 형태의 규칙이 지켜지지 않았다.

---

## §D.1 판정 게이트

- MUST-PASS 10개 — AC-ALH-001 · AC-ALH-002 · AC-ALH-003 · AC-ALH-004 · AC-ALH-005 · AC-ALH-006 · AC-ALH-007 · AC-ALH-008 · AC-ALH-009 · AC-ALH-010. 하나라도 FAIL 이면 SPEC 은 FAIL 이다.
- 품질 게이트: 이 트리에서 빌드한 바이너리로 `moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001` 가 0 error, 0 warning 으로 끝난다. 하네스 테스트 파일에는 CI 판 `golangci-lint run ./internal/cli/...` 가 0 issues 로 끝난다.
- 이 SPEC 의 판정은 「150,000 을 달성했는가」가 아니다. 측정이 완결됐고 판정 토큰이 규칙대로 붙었는가다. `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 와 `UNDETERMINED` 도 AC 를 통과할 수 있으나, 둘 다 상신 절차를 요구한다(REQ-ALH-011) — `UNTRIED` 를 늘려 상신을 피하는 길은 없다.

## §D.2 추적성

| REQ | AC |
|---|---|
| REQ-ALH-001 | AC-ALH-001 |
| REQ-ALH-002 | AC-ALH-002 |
| REQ-ALH-003 | AC-ALH-003 |
| REQ-ALH-004 | AC-ALH-003 |
| REQ-ALH-005 | AC-ALH-003 |
| REQ-ALH-006 | AC-ALH-004 |
| REQ-ALH-007 | AC-ALH-005 |
| REQ-ALH-008 | AC-ALH-005 |
| REQ-ALH-009 | AC-ALH-006 |
| REQ-ALH-010 | AC-ALH-007 |
| REQ-ALH-011 | AC-ALH-007 |
| REQ-ALH-012 | AC-ALH-007 |
| REQ-ALH-013 | AC-ALH-008 |
| REQ-ALH-014 | AC-ALH-003 |
| REQ-ALH-015 | AC-ALH-010, AC-ALH-002, AC-ALH-006, AC-ALH-007 |
| REQ-ALH-016 | AC-ALH-009 |

## §D.3 경계 사례

- **`S_init` 에 18경로 일부가 없다** — 하네스 `t.Log` 의 존재 여부 열이 근거다. `count-set-init.txt` 에서 그 경로를 빼고 판정서 `missing_init = <경로 …>` 줄에 누락 경로를 적는다(형태 2 라도 `count_set_init = 18` 표기는 유지한다 — 줄 수 규칙은 AC-ALH-002 가 누락 수를 빼서 맞춘다). 경로 목록을 조용히 바꾸지 않는다.
- **후보가 두 기제에 걸친다** — 한 절의 일부는 M1, 나머지는 M2 로 갈 수 있으면, 스크립트가 절 행 아래 문단 행(`¶n`)을 산출하고 문단 행으로 판정한다. 분할 완결성(AC-ALH-003 (2))은 절 행으로만 세므로 이중 계상되지 않는다. 허용 자수는 절 행과 문단 행 중 한 곳에만 둔다.
- **포인터 재유입이 제거량보다 크다** — 순감이 음수인 M1 후보는 `REJECT`(사유 `net-negative`)이며 증거에 `pointer_chars` 가 있어야 한다.
- **서문** — 첫 제목 앞의 텍스트(frontmatter 포함)는 후보 행 `(서문)` 이다. `paths:` 줄을 지우는 후보는 `REJECT`(`governs-scope`)다.
