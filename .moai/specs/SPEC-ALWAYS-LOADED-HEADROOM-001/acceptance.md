# SPEC-ALWAYS-LOADED-HEADROOM-001 — 인수 조건

모든 AC 는 Given-When-Then 이며 이진 판정된다. 명령은 워크트리 루트에서 실행하며, `$SCRATCH` 는 커밋 트리 밖의 세션 임시 디렉터리를 가리킨다. 판정서는 `.moai/reports/t1226/verdict.md` 다.

**설계 원칙 — 뼈대만으로는 어떤 AC 도 통과하지 않는다.** plan 단계가 커밋한 판정서 뼈대에는 아래 AC 들이 요구하는 기계 줄(`total_*`·`build_head`·`hash_*`·`verdict_*` 등)이 하나도 없다. 측정이 없으면 모든 MUST-PASS 가 FAIL 한다(plan-audit D4).

---

## §D. AC 매트릭스

| AC | 대상 REQ | 판정 | 심각도 |
|---|---|---|---|
| AC-ALH-001 | REQ-ALH-001, REQ-ALH-014 | 기계적 | MUST-PASS |
| AC-ALH-002 | REQ-ALH-002 | 기계적 | MUST-PASS |
| AC-ALH-003 | REQ-ALH-003, REQ-ALH-004, REQ-ALH-005, REQ-ALH-015 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-004 | REQ-ALH-006 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-005 | REQ-ALH-007, REQ-ALH-008 | 기계적 | MUST-PASS |
| AC-ALH-006 | REQ-ALH-009 | 기계적 | MUST-PASS |
| AC-ALH-007 | REQ-ALH-010, REQ-ALH-011, REQ-ALH-012 | 기계적 + 검토 | MUST-PASS |
| AC-ALH-008 | REQ-ALH-013 | 기계적 | MUST-PASS |
| AC-ALH-009 | REQ-ALH-017 | 기계적 | MUST-PASS |
| AC-ALH-010 | REQ-ALH-016 | 기계적 + 검토 | MUST-PASS |

### 판정서 기계 줄 규약

판정서는 아래 줄을 **이 철자 그대로**, 줄머리에서 시작해 한 줄씩 적는다. 정수는 쉼표 없이 쓴다. `<s>` 는 `init` 또는 `live` 다.

| 줄 | 값 |
|---|---|
| `레인 백엔드: <관측값> (출처: <관측 방법>)` | run 레인이 관측한 서빙 모델과 그 관측 방법 |
| `charmap = <값>` | `locale charmap` 출력 |
| `build_head = <40자리 hex>` | 측정 시점 `git rev-parse HEAD` |
| `harness_sha256 = <64자리 hex>` | 격리 하네스 테스트 파일의 `shasum -a 256` |
| `total_<s> = N` | 표면의 18경로 `wc -m` 합계 |
| `hash_<s> = <64자리 hex>` | 표면의 동결 다중집합 sha256(AC-ALD2-002 파이프라인, 표면 트리에서 실행) |
| `current_<s>` · `A_adm_<s>` · `R_<s>` · `U_<s>` · `T_min_<s>` ` = N` | AC-ALH-006 |
| `P절 재조정: (a)` / `(b)` / `(공표값)` / `(기타)` | AC-ALH-005 |
| `J_includes_kanban_scope = yes` / `no` | AC-ALH-005 |
| `F(172ef22eb) = N` | AC-ALH-005 — 서술값 |
| `verdict_<s> = <토큰>` | AC-ALH-007 |
| `runtime_files_init = N` · `runtime_total_init = <런타임 표기 그대로>` · `runtime_source = <관측 방법>` | AC-ALH-010 — 관측이 없으면 세 줄 대신 `runtime_observed = no` |
| `verdict_init_18 = <토큰>` · `verdict_init_17 = <토큰>` | AC-ALH-010 — `runtime_observed = no` 일 때만 |

### 후보 표 TSV 열 규약

두 TSV(`.moai/reports/t1226/candidates-init.tsv`, `.moai/reports/t1226/candidates-live.tsv`)는 `.moai/reports/t1226/candidates.py` 가 생성하고 판정 열을 사람이 채운다. 머리 줄 하나에 이어 후보 행이며 열은 다음 순서다.

| 열 | 이름 | 값 |
|---:|---|---|
| 1 | `file` | 18경로 중 하나 |
| 2 | `section` | 절 제목. 서문은 `(서문)`, 문단 후보는 `<절 제목> ¶n` |
| 3 | `gross` | 후보 전체 자수(정수, 스크립트 산출) |
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
| `net-negative` | `REJECT` | 증거 파일에 순감 산술 |
| `rewraps-binding-line` | `REJECT` | `mech == M2`, 증거 파일에 해시 불일치 출력 |
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

### AC-ALH-002 — 두 표면을 실제로 쟀다

**Given** 판정서와 측정 시점 커밋 `build_head` 가 있을 때,
**When** `build_head` 트리에서 아래 `S_live` 블록을 다시 돌리고 판정서의 기계 줄과 대조할 때,
**Then** `total_live` 가 재실행 `total` 과 같고, `total_init` 이 정수로 있고, `charmap = UTF-8` 이며, `## S_init`·`## S_live` 절 안에 자리표시가 0건이다.

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
grep -c -E '^total_init = [0-9]+$' "$V"                               # 1
grep -E '^total_live = [0-9]+$' "$V"                                  # 재실행 total 과 같은 값
awk '/^## S_init/,/^## S_live/' "$V" | grep -c '_미측정'               # 0
awk '/^## S_live/,/^## P절 재조정/' "$V" | grep -c '_미측정'           # 0
```

`build_head` 가 `7fe658815` 과 다르고 그 사이 흡수가 18경로를 바꿨다면(`git diff --quiet 7fe658815 <build_head> -- <18경로>` 가 0 이 아님), 판정서는 그 사실과 새 `total_live` 를 적고 기준선 199,111 은 이력으로만 남긴다.

FAIL 조건: 위 기대와 다른 출력이 하나라도 있다.

### AC-ALH-003 — 후보 표의 재생성·완결성·허용 규칙

**Given** 두 TSV, `candidates.py`, `dest-sizes.tsv` 가 커밋돼 있을 때,
**When** (1) 스크립트로 키 집합을 다시 뽑아 TSV 와 비교하고, (2) 행 규칙 `awk` 를 돌리고, (3) 목적지 누적 수용량을 검사할 때,
**Then** (1) diff 가 비고, (2)·(3) 이 모두 `BAD=0` 이다.

```bash
# (1) 키 재생성 — (file, section, gross, bind). 표면 트리 루트는 판정서 S_init / S_live 절에 적힌 경로
python3 .moai/reports/t1226/candidates.py --keys "$SCRATCH/init-surface" | sort > "$SCRATCH/k-init"
tail -n +2 .moai/reports/t1226/candidates-init.tsv | cut -f1,2,3,6 | sort | diff - "$SCRATCH/k-init" && echo KEYS-INIT-OK
mkdir -p "$SCRATCH/live"
git archive build_head_value CLAUDE.md AGENTS.md .moai/config/sections .claude/rules/moai | tar -x -C "$SCRATCH/live"
python3 .moai/reports/t1226/candidates.py --keys "$SCRATCH/live" | sort > "$SCRATCH/k-live"
tail -n +2 .moai/reports/t1226/candidates-live.tsv | cut -f1,2,3,6 | sort | diff - "$SCRATCH/k-live" && echo KEYS-LIVE-OK

# (2) 행 규칙
for f in .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv; do
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
      if ($14=="agents-md-no-m1" && $1 ~ /AGENTS\.md/) ok=1;
      if ($14=="config-data" && $1 ~ /\.yaml$/) ok=1;
      if ($14=="net-negative" || ($14=="rewraps-binding-line" && $5=="M2")) ok=1;
      if (!ok) bad++;
    }
    if ($3 !~ /^[0-9]+$/ || $15=="") bad++;
  } END {print FILENAME" BAD="bad+0}' "$f"
done
# 모든 행의 증거 파일 존재
for f in .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv; do
  tail -n +2 "$f" | cut -f15 | sort -u | while read -r p; do [ -f "$p" ] || echo "missing $p"; done
done | wc -l | awk '{print "MISSING="$1}'                                  # MISSING=0

# (3) 목적지 누적 수용량 — 두 트리 각각, 합이 40000 미만
for f in .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv; do
  awk -F'\t' 'FNR==NR {live[$1]=$2; tmpl[$1]=$3; next}
    FNR>1 && $13=="ADMIT" && $5=="M1" {add[$12]+=$4}
    END {for (d in add) { if (!(d in live) || live[d]+add[d]>=40000 || tmpl[d]+add[d]>=40000) bad++ } print FILENAME" DEST-BAD="bad+0}' \
    .moai/reports/t1226/dest-sizes.tsv "$f"
done
```

`build_head_value` 는 판정서 `build_head` 줄의 값으로 바꿔 넣는다. 검토 항목: `SPEC-ALWAYS-LOADED-DIET-002/design.md §4.3` 기각 표의 절 중 이 트리에 남아 있는 것은 모두 `gov` 가 `a`/`b` 이거나, `N` 이면 증거 파일이 그 판단을 뒤집는 근거를 담는다.

FAIL 조건: `KEYS-INIT-OK`·`KEYS-LIVE-OK` 중 하나가 없다, `BAD`·`DEST-BAD`·`MISSING` 중 하나가 0 이 아니다.

### AC-ALH-004 — M2 자수는 실제 압축 시도로 쟀다

**Given** 두 TSV 의 `mech == M2` 이고 `ADMIT` 또는 `rewraps-binding-line` 인 행이 있을 때,
**When** 각 행의 증거 파일에서 시도 후 해시 줄을 읽어 판정서의 표면 해시와 비교할 때,
**Then** `ADMIT` 행은 모두 표면 해시와 같고, `rewraps-binding-line` 행은 모두 다르다.

```bash
V=.moai/reports/t1226/verdict.md
for s in init live; do
  H=$(grep -E "^hash_${s} = " "$V" | awk '{print $3}')
  awk -F'\t' 'NR>1 && $5=="M2" && ($13=="ADMIT" || $14=="rewraps-binding-line") {print $13"\t"$15}' \
    ".moai/reports/t1226/candidates-${s}.tsv" \
  | while IFS="$(printf '\t')" read -r v p; do
      got=$(grep -E '^post_hash = ' "$p" | awk '{print $3}')
      if [ "$v" = "ADMIT" ] && [ "$got" != "$H" ]; then echo "bad $p"; fi
      if [ "$v" != "ADMIT" ] && [ "$got" = "$H" ]; then echo "bad $p"; fi
    done
done | wc -l | awk '{print "HASH-BAD="$1}'                                  # HASH-BAD=0
grep -c -E '^hash_(init|live) = [0-9a-f]{64}$' "$V"                          # 2
```

증거 파일은 scratch 사본에 가한 압축 diff 와 `post_hash = <sha256>` 줄을 담는다. `hash_init` 은 `S_init` 트리에서 **실측한** 값이며 라이브 해시를 옮겨 적지 않는다(D16). 검토 항목: 압축이 의무·조건·예외·수치를 지우지 않았다(REQ-ALD2-012 준용).

FAIL 조건: `HASH-BAD` 가 0 이 아니거나, 해시 줄이 2개가 아니다.

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

### AC-ALH-006 — `T_min` 산술이 표에서 재현된다

**Given** 판정서 기계 줄과 TSV·포인터 표가 있을 때,
**When** 아래 `awk` 로 각 항을 다시 계산할 때,
**Then** 표면마다 `CHECK-<s>=OK` 가 나온다.

```bash
V=.moai/reports/t1226/verdict.md
for s in init live; do
  A=$(awk -F'\t' 'NR>1 && $13=="ADMIT" {s+=$4} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  U=$(awk -F'\t' 'NR>1 && $13=="UNTRIED" {s+=$3} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  R=$(awk -F'\t' 'NR>1 {s+=$3} END {print s+0}' ".moai/reports/t1226/pointers-${s}.tsv")
  C=$(grep -E "^total_${s} = " "$V" | awk '{print $3}')
  awk -v s="$s" -v A="$A" -v U="$U" -v R="$R" -v C="$C" '
    $1=="current_"s {c=$3} $1=="A_adm_"s {a=$3} $1=="U_"s {u=$3} $1=="R_"s {r=$3} $1=="T_min_"s {t=$3}
    END { if (c==C && a==A && u==U && r==R && t==C-A+R) print "CHECK-"s"=OK"; else print "CHECK-"s"=BAD" }' "$V"
done
# 포인터 표가 ADMIT M1 (file,dest) 쌍과 일대일
for s in init live; do
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1" {print $1"\t"$12}' ".moai/reports/t1226/candidates-${s}.tsv" | sort -u > "$SCRATCH/pair-$s"
  tail -n +2 ".moai/reports/t1226/pointers-${s}.tsv" | cut -f1,2 | sort > "$SCRATCH/ptr-$s"
  diff "$SCRATCH/pair-$s" "$SCRATCH/ptr-$s" && echo "PAIRS-$s=OK"
done
```

검토 항목: 포인터 줄 원문이 증거로 커밋돼 있고 `pointer_chars` 가 그 `wc -m` 과 같다.

FAIL 조건: `CHECK-init=OK`·`CHECK-live=OK`·`PAIRS-init=OK`·`PAIRS-live=OK` 중 하나라도 없다.

### AC-ALH-007 — 판정 토큰의 규칙 재계산과 상신 절차

**Given** 판정서 기계 줄이 있을 때,
**When** 토큰을 `T_min`·`U` 에서 다시 계산하고 상신 절을 읽을 때,
**Then** 표면마다 `TOKEN-<s>=OK` 이고, `verdict_init` 이 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 또는 `UNDETERMINED` 이면 상신 절이 (a)~(d) 네 항목과 `RECOMMEND:` 문장을 담는다.

```bash
V=.moai/reports/t1226/verdict.md
for s in init live; do
  awk -v s="$s" '
    $1=="T_min_"s {t=$3} $1=="U_"s {u=$3} $1=="verdict_"s {v=$3; n++}
    END {
      if (t<150000) e="ACHIEVABLE"; else if (t-u>=150000) e="STRUCTURALLY-INFEASIBLE-UNDER-FREEZE"; else e="UNDETERMINED";
      print (n==1 && v==e) ? "TOKEN-"s"=OK" : "TOKEN-"s"=BAD("v" vs "e")" }' "$V"
done
# verdict_init 이 ACHIEVABLE 이 아닐 때
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '^### \((a|b|c|d)\)'    # 4
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '^RECOMMEND: '           # 1 이상
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '결정했다|승인했다|해제한다\.?$' # 0
```

`AC-ALH-010` 의 규칙에 따라 `runtime_observed = no` 이고 `verdict_init_18` 과 `verdict_init_17` 이 다르면, 이 재계산과 무관하게 `verdict_init = UNDETERMINED` 여야 한다(그 경우 `TOKEN-init` 은 BAD 로 나와도 `AC-ALH-010` 의 규칙이 우선한다 — 판정서에 그 사유 줄 `verdict_init_reason = count-set` 이 있어야 한다). `UNDETERMINED` 이면 상신 절이 `UNTRIED` 목록과 `U_init` 을 인용한다(검토).

FAIL 조건: 토큰 불일치(위 예외 제외), 상신 절 항목 수 4 가 아님, `RECOMMEND:` 0건, 결정 서술 존재.

### AC-ALH-008 — 카드가 작성한 커밋은 계수 파일과 미러를 건드리지 않았다

**Given** 기준 커밋 `7fe658815` 와 카드 HEAD 가 있을 때,
**When** 비병합 커밋만 대상으로 18경로와 그 미러의 이력을 볼 때,
**Then** 출력이 비어 있다.

```bash
git log --no-merges --format=%H 7fe658815..HEAD -- CLAUDE.md AGENTS.md \
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

develop 흡수로 들어온 병합 커밋은 대상이 아니다. 흡수가 18경로를 바꿨다면 AC-ALH-002 의 규정대로 재측정한다.

FAIL 조건: 출력이 0 이 아니다.

### AC-ALH-009 — `S_init` 은 격리 하네스로만 산출됐다

**Given** 격리 하네스 테스트 `TestHeadroomInitSurfaceExport` 가 `internal/cli/` 에 커밋돼 있을 때,
**When** 그 테스트를 앵커된 패턴으로 돌린 출력 파일과 판정서를 읽을 때,
**Then** 테스트가 PASS 했고(SKIP 이 아님), 하네스 해시가 판정서와 같으며, 비격리 `moai init` 실행 기록이 없다.

하네스 실행(출력은 `.moai/reports/t1226/harness-run.txt` 로 커밋):

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run '^TestHeadroomInitSurfaceExport$' -count=1 -v -args -headroom-export="$SCRATCH/init-surface"
```

판정 명령:

```bash
H=.moai/reports/t1226/harness-run.txt
grep -c -- '--- PASS: TestHeadroomInitSurfaceExport ' "$H"                    # 1
grep -c -- '--- SKIP: TestHeadroomInitSurfaceExport ' "$H"                    # 0
F=$(git ls-files 'internal/cli/*headroom*_test.go')
shasum -a 256 $F | awk '{print $1}'   # 판정서 harness_sha256 과 같음
grep -c 'prepareSafeInitHome' $F                                              # 1 이상
grep -c -E '^harness_sha256 = [0-9a-f]{64}$' .moai/reports/t1226/verdict.md   # 1
test -s .moai/reports/t1226/commands.log && echo LOG-OK                        # LOG-OK
grep -c -E '(^|[;&|] *)([^ ]*/)?moai init' .moai/reports/t1226/commands.log   # 0
```

`commands.log` 는 run 단계가 실행한 셸 명령을 한 줄씩 적은 기록이다(임시 경로는 `$SCRATCH` 표기). 감사 보고서처럼 `moai init` 을 **서술**하는 문서는 검사 대상이 아니다.

하네스 계약(검토): `prepareSafeInitHome` 를 호출한 뒤 `runInitWithFlags` 와 같은 방식으로 init 명령을 실행하고, 고정 플래그는 `--non-interactive`, `--root <t.TempDir()>`, `--name headroom-probe`, `--language go`, `--mode tdd`, `--llm claude` 이며 `--all` 을 주지 않는다(기본 slim 설치가 기본 사용자의 표면이다). 18경로를 `-headroom-export` 인자가 가리키는 디렉터리로 복사하고, 18경로의 (경로, 존재 여부, sha256) 를 `t.Log` 로 출력한다. 인자가 비면 `t.Skip` 한다 — 그래서 위 판정은 SKIP 을 FAIL 로 친다.

FAIL 조건: 위 기대와 다른 출력이 하나라도 있다. 비격리 실행(실제 홈을 쓰는 `moai init`)의 흔적은 그 자체로 FAIL 이다.

### AC-ALH-010 — 런타임 계수 집합과의 대조

**Given** `S_init` 산출 트리가 있을 때,
**When** 런타임이 그 트리에서 보고하는 always-loaded 파일 수와 합계를 관측하거나, 관측하지 못했음을 기록할 때,
**Then** 판정서가 아래 둘 중 정확히 한 형태를 갖는다.

```bash
V=.moai/reports/t1226/verdict.md
# 형태 1 — 관측됨
grep -c -E '^runtime_files_init = [0-9]+$' "$V"; grep -c -E '^runtime_total_init = .+' "$V"; grep -c -E '^runtime_source = .+' "$V"   # 각 1
# 형태 2 — 관측 안 됨
grep -c -E '^runtime_observed = no$' "$V"                                                   # 1
grep -c -E '^verdict_init_(18|17) = (ACHIEVABLE|STRUCTURALLY-INFEASIBLE-UNDER-FREEZE|UNDETERMINED)$' "$V"   # 2
```

규칙:
- 형태 1 에서 `runtime_files_init` 이 18 이 아니거나 `runtime_total_init` 이 `total_init` 과 1% 넘게 어긋나면, 판정서는 차이 나는 경로를 밝히고 `T_min_init` 을 관측된 집합으로 계산한다(검토).
- 형태 2 에서 `verdict_init_18 ≠ verdict_init_17` 이면 `verdict_init = UNDETERMINED` 와 `verdict_init_reason = count-set` 이 있어야 하고, 관측 불가 사실이 Gaps 에 있다.
- 관측 방법의 예: 리드 또는 운영자가 워크트리 밖 세션에서 `S_init` 사본 디렉터리로 `claude` 를 띄워 기동 경고 줄을 원문으로 옮기는 것. 레인이 직접 띄울 수 없으면 리드에게 관측을 요청하는 블로커를 올리고, 답이 없으면 형태 2 로 닫는다.

FAIL 조건: 두 형태 중 어느 것도 완결되지 않았거나, 두 형태가 함께 있거나, 형태 2 의 규칙이 지켜지지 않았다.

---

## §D.1 판정 게이트

- MUST-PASS 10개 — AC-ALH-001 · AC-ALH-002 · AC-ALH-003 · AC-ALH-004 · AC-ALH-005 · AC-ALH-006 · AC-ALH-007 · AC-ALH-008 · AC-ALH-009 · AC-ALH-010. 하나라도 FAIL 이면 SPEC 은 FAIL 이다.
- 품질 게이트: 이 트리에서 빌드한 바이너리로 `moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001` 가 0 error, 0 warning 으로 끝난다.
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
| REQ-ALH-014 | AC-ALH-001 |
| REQ-ALH-015 | AC-ALH-003 |
| REQ-ALH-016 | AC-ALH-010 |
| REQ-ALH-017 | AC-ALH-009 |

## §D.3 경계 사례

- **`S_init` 에 18경로 일부가 없다** — 하네스 `t.Log` 의 존재 여부 열이 근거다. 판정서는 누락 경로를 적고 실제 존재하는 경로로 `total_init` 을 잰다. 경로 목록을 조용히 바꾸지 않는다.
- **후보가 두 기제에 걸친다** — 한 절의 일부는 M1, 나머지는 M2 로 갈 수 있으면 두 행으로 나누되, 키 재생성(AC-ALH-003 (1))과 맞도록 스크립트가 분할 단위를 산출한다. 자수를 이중 계상하지 않는다.
- **포인터 재유입이 제거량보다 크다** — 순감이 음수인 후보는 `REJECT`(사유 `net-negative`)다.
- **서문** — 첫 제목 앞의 텍스트도 후보 행 `(서문)` 으로 싣는다. 파일 머리의 frontmatter 는 서문에 포함하되 `paths:` 줄을 지우는 후보는 조건 3 과 무관하게 `REJECT`(`governs-scope`)다.
