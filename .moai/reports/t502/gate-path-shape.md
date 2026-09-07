# t502 판정서 — 미러된 스킬에서 codex `[[skills.config]]` 게이트가 묶이는 경로 모양

- **카드**: t502 (SPEC-CODEX-SKILL-DISABLE-001 plan-phase 차단 게이트, AC-CSD-050/051)
- **측정 일시**: 2026-09-07, 런타임 스탬프 **`codex-cli 0.153.4`** (`codex --version`, 이 실험 시작 시 직접 측정 — `lab/version.txt`)
- **바이너리**: `/Users/goos/.local/bin/codex`
- **방법**: `CODEX_HOME` 격리 (t504 선례). 실사용자 `~/.codex` **기록 없음**
- **선행**: `.moai/reports/t504/skills-config-path-shape.md` (D-시리즈 — 실디렉터리 + `$CODEX_HOME/skills/` 컨벤션)
- **랩 아카이브**: `lab/` (셀 캡처 19개 전량 + 셀별 config.toml + 요약), 재현 `harness.sh` → `harness2.sh <EXP_ROOT>`
- **계측기**: `probe.sh` — 이 매트릭스의 재생이 아니라 **파라미터화된 계측기**. `--codex-home` / `--entry-path` / `--enabled` / `--skill` 4개가 입력이며, 구현 단계가 verb가 만든 config를 그대로 겨눌 수 있다 (§ 계측기와 E2E 축)
- **원 실험 트리**: `/tmp/t502-HTmXMiC0` (세션 종료 후 OS가 지움 — 영구 사본은 `lab/`)

---

## 한 문단 요약

게이트는 **문자열이 아니라 정규화된 실파일(realpath)** 에 묶인다. 심링크 미러에서는 리터럴 경로(`.agents/skills/<n>/SKILL.md`)와 심링크 해소 경로(`.claude/skills/<n>/SKILL.md`)가 **같은 실파일로 수렴**하므로 **둘 다 차단된다**. 복사 폴백 미러에서는 둘이 **서로 다른 실파일**이 되어 **리터럴만 차단되고 `.claude/…` 표기는 무효**가 된다. 따라서 **두 미러 모양 모두에서 묶이는 표기는 리터럴 미러 경로 하나뿐**이며, 이것이 t502가 발행해야 할 표기다. 디렉터리 모양 엔트리와 상대경로 표기는 두 모양 모두에서 무효다(t504 D1d와 정합).

---

## 관측된 스킬 루트 (격리 랩 기준)

`codex debug prompt-input` 출력의 `### Skill roots` 절 (verbatim, `lab/cells/S0.out`):

```
### Skill roots
- `r0` = `/Users/goos/.agents/skills`
- `r1` = `/private/tmp/t502-HTmXMiC0/home-S0/skills/.system`
- `r2` = `/private/tmp/t502-HTmXMiC0/proj-S/.agents/skills`
```

프로브가 실제로 로드된 경로 (verbatim, 같은 파일):

```
- t502probe: T502MARKER probe skill for SPEC-CODEX-SKILL-DISABLE-001. (file: r2/t502probe/SKILL.md)
```

읽어낼 것 세 가지:

1. **cwd `.agents/skills`는 실재하는 루트다** (r2) — moai 미러가 노리는 통로가 이 버전에서 살아 있다. 표기는 **리터럴**(`proj-S/.agents/skills`)이며, 심링크 해소본(`.claude/skills`)으로 표시되지 **않는다**.
2. **`~/.agents/skills`(r0)는 `CODEX_HOME` 격리를 뚫는다.** 스크래치 홈을 줬는데도 실사용자 홈의 이 디렉터리가 루트로 들어왔다. 본 실험은 프로브 이름이 고유해 오염되지 않았지만, 격리가 완전하지 않다는 사실 자체를 기록한다.
3. **루트 집합·순서는 t504와 다르다.** t504는 실환경 셀에서 r0=`~/.codex/skills`, r1=`~/.agents/skills`를 봤고, 여기서는 격리 홈이라 `$CODEX_HOME/skills`에 `.system`만 있어 r1로 잡혔다. 즉 **루트 인덱스는 환경에 따라 재배치되는 상대 라벨**이지 고정 번호가 아니다 — SPEC이 루트를 번호로 지목하면 안 된다.

---

## 마커 규약 — "노출됨 / 차단됨"을 무엇으로 판정했는가

이 보고서의 모든 판정은 아래 한 가지 규약으로만 내려졌다. 뒤에 오는 독자가 내가 어떻게 셌는지 추론할 필요가 없도록 명시한다.

- **마커 토큰**: 프로브 스킬 `t502probe`의 `SKILL.md` **frontmatter `description`과 본문 양쪽**에 심은 고유 문자열. 라운드 1·2는 `T502MARKER`, 계측기(`probe.sh`)는 기본값 `T502PROBE_MARKER`(스킬명에서 유도, `--marker`로 덮어쓸 수 있음). 두 토큰 모두 codex 배포물·실사용자 스킬 어디에도 없는 값이라 우연 일치가 없다.
- **캡처**: 셀마다 `codex debug prompt-input "hi"`의 stdout을 파일로 받는다(`<cell>.out`), stderr는 별도 파일, 종료코드는 `<cell>.rc`.
- **카운트**: `grep -c '<MARKER>' <cell>.out` — **디스크 파일 대상**이다. 파이프로 세지 않는다(파이프 카운트는 상류가 실행조차 안 됐을 때도 0을 낸다). 0매치일 때 `grep`이 rc=1을 내는 것은 `|| true`로 흡수하고, 그 1은 판정에 쓰지 않는다.
- **판정 규칙** (세 갈래, 순서대로):
  - `rc ≠ 0` **또는** stderr 바이트 ≠ 0 → **ERROR**. 판정이 아니다. 마커 0이어도 "차단됨"으로 읽지 않는다 — t504 V1/V2/V3가 스키마 에러로 죽어 마커 0을 냈던 바로 그 함정이다.
  - `rc = 0` ∧ stderr 0 ∧ 마커 카운트 **≥ 1** → **노출됨(exposed)**
  - `rc = 0` ∧ stderr 0 ∧ 마커 카운트 **= 0** → **차단됨(gated)**
- **보조 신호 2개** (판정에는 안 쓰고 정합성 확인용): 스킬 이름 `t502probe` 카운트가 마커와 같은 방향으로 움직이는지, 그리고 stdout 바이트가 대조군 대비 정확히 엔트리 1줄분(약 160B)만 줄었는지. 19셀 전부에서 두 신호가 마커와 일치했다.
- **한 셀이 인쇄하는 것**: `probe.sh probe`는 `verdict=<exposed|gated|ERROR> marker=N name=N rc=N out=N err=N capture=<path>` 한 줄과, 실사용자 config의 before/after sha256 두 줄, `notouch=IDENTICAL|DIFFERS` 한 줄. `--expect`를 주면 `expect=… result=MATCH|MISMATCH`가 붙고 불일치는 exit 3이다.

## 셀 매트릭스 — 라운드 1 (미러 모양 × 표기 × enabled)

프로젝트 픽스처 2종 (`internal/template/skill_mirror.go` 실물 모양 그대로):

- **proj-S** = `MirrorModeSymlink`: `.agents/skills/t502probe` → `../../.claude/skills/t502probe` (상대 심링크). 실제 링크 확인: `lrwxr-xr-x … t502probe -> ../../.claude/skills/t502probe`
- **proj-C** = `MirrorModeCopy`: `.claude/skills/t502probe/SKILL.md`와 `.agents/skills/t502probe/SKILL.md`가 **각각 실파일**, 경로에 심링크 없음

| 셀 | 미러 | 엔트리 `path` 표기 | `enabled` | out(bytes) | err | rc | marker | 판정 |
|----|------|------------------|-----------|-----------|-----|----|--------|------|
| **S0** | symlink | (엔트리 없음) | — | 24919 | 0 | 0 | **1** | 대조군 — 미러가 노출된다 ✓ |
| **Split** | symlink | 리터럴 `.agents/…/SKILL.md` | true | 24922 | 0 | 0 | **1** | 양성 통제 ✓ |
| **Spres** | symlink | 해소 `.claude/…/SKILL.md` | true | 24922 | 0 | 0 | **1** | 양성 통제 ✓ |
| **Slit** | symlink | 리터럴 | false | 24759 | 0 | 0 | **0** | **차단됨** |
| **Sres** | symlink | 해소 | false | 24759 | 0 | 0 | **0** | **차단됨** |
| **C0** | copy | (엔트리 없음) | — | 24919 | 0 | 0 | **1** | 대조군 ✓ |
| **Cplit** | copy | 리터럴 `.agents/…/SKILL.md` | true | 24922 | 0 | 0 | **1** | 양성 통제 ✓ |
| **Cpres** | copy | 쌍둥이 `.claude/…/SKILL.md` | true | 24922 | 0 | 0 | **1** | 양성 통제 ✓ |
| **Clit** | copy | 리터럴 | false | 24759 | 0 | 0 | **0** | **차단됨** |
| **Cres** | copy | 쌍둥이 `.claude/…` | false | 24921 | 0 | 0 | **1** | **차단 안 됨 — 무효 엔트리** |

stderr 10셀 전부 0바이트, rc 전부 0. `marker` = 디스크 파일 대상 `grep -c T502MARKER`(파이프 판정 아님), `name` 카운트도 동일하게 움직였다.

## 셀 매트릭스 — 라운드 2 (매칭 방식 판별 + 모양/표기 변주)

라운드 1의 「심링크에서 두 표기 모두 차단」은 두 가설이 똑같이 예측한다 — **어휘적 정규화**(`.`/`..` 접기)와 **완전 realpath 정규화**(심링크 해소). 이를 가르는 셀:

| 셀 | 프로젝트 | 엔트리 `path` | out | rc | marker | 판정 |
|----|---------|--------------|-----|----|--------|------|
| **EC** | proj-S | (없음) | 24918 | 0 | **1** | 라운드 2 자체 대조군 ✓ |
| **ECc** | proj-C | (없음) | 24919 | 0 | **1** | 라운드 2 자체 대조군 ✓ |
| **E1** | proj-S | `…/.agents/skills/./t502probe/../t502probe/SKILL.md` | 24757 | 0 | **0** | 차단 — **어휘 정규화 있음** |
| **E2** | proj-S | `…/alias-dir/SKILL.md` (무관한 별도 심링크 → 캐노니컬 디렉터리) | 24756 | 0 | **0** | 차단 — **realpath 정규화 확정** |
| **E3** | proj-S | `/private` 접두 리터럴 | 24757 | 0 | **0** | 차단 (macOS `/tmp`→`/private/tmp` 링크 통과) |
| **E4** | proj-S | 디렉터리 모양 `.agents/skills/t502probe` | 24919 | 0 | **1** | 차단 안 됨 — 디렉터리 모양 무효 |
| **E5** | proj-C | 디렉터리 모양 | 24919 | 0 | **1** | 차단 안 됨 |
| **E6** | proj-C | 상대경로 `.agents/skills/t502probe/SKILL.md` | 24919 | 0 | **1** | 차단 안 됨 — 상대 표기 무효 |
| **E8** | proj-C | `/private` 접두 리터럴 | 24754 | 0 | **0** | 차단 |

**E2가 결정적이다.** `alias-dir`는 미러와 어떤 경로 성분도 공유하지 않는 별도 심링크인데, 그 경로로 쓴 엔트리가 실제로 로드된 스킬을 껐다. 어휘 정규화로는 설명되지 않고 **엔트리 경로와 스킬 파일 경로를 각각 realpath로 정규화해 비교한다**로만 설명된다. E1·E3·E8은 그 정규화가 `.`/`..`와 상위 심링크까지 포함함을 보강한다.

---

## 연구질문 판정

### Q-a — 게이트는 리터럴 경로에 묶이는가, 심링크 해소 경로에 묶이는가?

**답: 해소(realpath) 경로에 묶인다. "리터럴"이 통하는 것은 그것이 같은 실파일로 해소되기 때문이지, 코덱스가 문자열을 그대로 비교해서가 아니다.**

근거 셀: **E2**(무관한 별칭 링크로 쓴 경로가 차단 → 문자열 비교 배제), **Slit + Sres**(심링크 미러에서 두 표기가 동일하게 차단 → 같은 실파일로 수렴), **E1/E3/E8**(비정규 철자·`/private` 접두도 차단). 반대 방향의 결정적 증거는 **Cres**: 복사 미러에서 `.claude/…` 표기는 로드된 파일과 **다른 실파일**이므로 차단이 일어나지 않았다(marker=1). 문자열 매칭이었다면 Cres와 Sres가 같은 결과여야 하는데 갈렸다.

부수 판정 2건:
- **디렉터리 모양 무효** (E4/E5, marker=1 둘 다) — t504 D1d를 미러 통로에서 재확인. 게이트는 `SKILL.md` 파일에만 묶인다.
- **상대경로 표기 무효** (E6, marker=1) — cwd 기준 상대 표기는 묶이지 않는다. 발행은 절대경로여야 한다.

### Q-b — 심링크에서 묶이는 표기가 복사 폴백에서도 묶이는가? 그 반대는?

**답: 리터럴 미러 경로는 YES(두 모양 모두 묶임). 해소 `.claude/…` 표기는 NO(복사 모양에서 무효).**

| 표기 | 심링크 미러 | 복사 미러 | 두 모양 공통 |
|------|------------|----------|-------------|
| 리터럴 `<proj>/.agents/skills/<n>/SKILL.md` | **차단** (Slit) | **차단** (Clit, E8) | **✓ 묶인다** |
| 해소 `<proj>/.claude/skills/<n>/SKILL.md` | 차단 (Sres) | **무효** (Cres) | ✗ |
| 디렉터리 모양 | 무효 (E4) | 무효 (E5) | ✗ |
| 상대경로 | 미측정 | 무효 (E6) | ✗ |

**따라서 t502가 발행해야 할 표기는 하나로 확정된다: 프로젝트 루트 기준 절대경로 `<projectRoot>/.agents/skills/<skill>/SKILL.md`.** 이 표기는 심링크 미러(정상 경로)에서는 캐노니컬 파일로 해소되어 묶이고, 복사 미러에서는 그 자체가 로드된 실파일이라 묶인다. `.claude/skills/…` 표기를 발행하면 **복사 폴백 사용자에게서 조용히 무효**가 된다 — 정확히 t504가 막으려던 실패 등급이다.

### 추가 질문 — `MirrorModeSkipped`가 "이름 해소 거부 경로"로 간다는 SPEC 주장

**답: 본 매트릭스가 반증한다 — C-셀이 그대로 그 모양이다.**

`MirrorModeSkipped`는 `.agents/skills/<n>`에 **심링크가 아닌 실엔트리**가 이미 있어 손대지 않은 상태다(`skill_mirror.go:mirrorOneSkill` — `os.Lstat` 후 `info.Mode()&os.ModeSymlink == 0` 분기). 그 실엔트리가 `SKILL.md`를 품은 디렉터리라면, 파일시스템 모양은 **proj-C와 구별되지 않는다** — 실디렉터리 + 실 `SKILL.md`. 그리고 C0/ECc 대조군이 보였듯 그 모양은 **정상적으로 로드되며**(marker=1), 리터럴 표기로 **정상적으로 차단된다**(Clit marker=0).

즉 "skipped ⇒ 해소 실패 ⇒ 거부 경로"는 **성립하지 않는다**. 해소는 성공한다. 단 한 가지 유보: 사용자가 만든 실엔트리가 **`SKILL.md`를 품지 않은** 경우(빈 디렉터리, 혹은 동명의 일반 파일)는 이번에 측정하지 않았다 — 그 하위 사례에서만 해소 실패가 성립할 수 있다. Gaps에 남긴다.

---

## 채택성 게이트

- **S+ 계열 양성 통제** — Split·Spres 모두 marker=1 ✓ (엔트리 존재 자체가 노출을 깨지 않음)
- **C+ 계열 양성 통제** — Cplit·Cpres 모두 marker=1 ✓
- **무엔트리 대조군** — S0·C0·EC·ECc 4셀 전부 marker=1 ✓ (두 미러 모양 모두 애초에 노출된다)
- 두 통제군이 같은 방향으로 실패한 경우 없음. **VOID 아님 — 행렬 채택 가능.**

음성 판정(marker=0) 셀들은 전부 rc=0·stderr 0바이트다. 즉 "에러로 죽어서 마커가 없다"가 아니라 **정상 실행 후 스킬 목록에서만 사라졌다**. out 크기도 대조군 대비 정확히 프로브 엔트리 1줄분(약 160B)만큼 작다(24919 → 24759).

---

## 계측기와 E2E 축 — 구현 단계가 이어받을 지점

`harness.sh`/`harness2.sh`는 이 매트릭스의 **기록**이다. `probe.sh`는 같은 랩을 **입력을 받는 도구**로 다시 만든 것으로, 나중 단계가 격리·프로브 스킬·마커 grep을 재구축하지 않고 verb가 만든 config를 겨눌 수 있게 한다.

**입력 4개** (지시받은 축 그대로): `--codex-home`(격리 루트), `--entry-path`(엔트리 `path` 값), `--enabled`(엔트리 `enabled` 값), `--skill`(프로브 스킬/마커 이름). 여기에 `--project`(cwd)와 `--expect`(단언)를 더한다.

**서브커맨드 3개:**

| 커맨드 | 역할 |
|---|---|
| `fixture --root <dir> --mirror symlink\|copy` | `skill_mirror.go`가 만드는 두 모양을 격리 트리에 재현하고 `PROJECT` / `LITERAL_PATH` / `RESOLVED_PATH` / `MARKER`를 인쇄 |
| `probe --codex-home … --project … [--entry-path … --enabled …] [--expect …]` | 한 셀 실행 후 위 규약대로 한 줄 판정. **`--entry-path`/`--enabled`를 주면 config를 쓰고, 안 주면 이미 있는 config를 그대로 읽는다** |
| `selftest` | 계측기가 **양방향을 다 검출하는지** 자체 검증 |

**E2E 축이 재사용할 셀 모양은 S-lit / C-x다** — 즉 `--entry-path <LITERAL_PATH> --enabled false`로 차단을 확인한 그 셀. 구현 단계는 그 두 플래그를 **빼고** 같은 셀을 돌린다: verb가 쓴 config를 `probe`가 손대지 않고 읽어, verb 실행 **전** `--expect exposed`, **후** `--expect gated`로 marker 1 → 0을 끝에서 끝까지 잰다. 감사가 지적한 세 결함(엔트리가 생기는가 / 값이 맞는가 / **실제로 꺼지는가**) 중 세 번째, 어느 기준도 검사하지 않던 합성 축이 바로 이것이다.

**verb 대역(stand-in)으로 그 경로를 실제로 한 번 통과시켰다.** verb는 아직 없으므로 그 자리에 `printf`로 config를 쓰는 대역을 놓고, 계측기는 평소처럼 읽기만 했다:

```
[verb 실행 전]  verdict=exposed marker=1 name=1 rc=0 out=24917 err=0   expect=exposed result=MATCH
[대역이 config 작성: path=<LITERAL_PATH>, enabled=false]
[verb 실행 후]  verdict=gated   marker=0 name=0 rc=0 out=24754 err=0   expect=gated   result=MATCH
```

실행 후 config를 다시 읽어 대역이 쓴 4줄이 **그대로**임을 확인했다(`lab/instrument/e2e-demo-verb-config.toml`) — 계측기가 검사 대상을 덮어써 통과를 만들어내지 않는다. 즉 verb만 갈아끼우면 이 축이 그대로 닫힌다.

**계측기 자가검증 (`probe.sh selftest`) — 원 하네스와 독립적으로 4개 판정 재현:**

```
symlink	control=exposed	literal_false=gated
copy	control=exposed	literal_false=gated
copy	resolved_false=exposed	(expected exposed — the inert notation)
SELFTEST PASS
```

세 번째 줄이 특히 중요하다. 계측기가 **무엇이든 gated로 부르는 도구가 아님**을 보이는 셀이다 — 복사 모양에서 해소 표기는 여전히 노출로 나와야 하고, 그렇지 않으면 selftest가 실패한다. 즉 selftest는 「차단을 검출한다」와 「무해한 엔트리를 차단으로 오인하지 않는다」를 함께 요구한다. 이 세 판정은 라운드 1의 Slit·Clit·Cres와 **별도 트리·별도 프로브 이름 유도값**으로 다시 측정된 것이라, 원 매트릭스의 독립 재현이기도 하다.

## 노터치 증명

```
before:  9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a  /Users/goos/.codex/config.toml
after:   9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a  /Users/goos/.codex/config.toml
after2:  9f6e3a953880630afcfa6a40abe846b785e8fc0067e17b3a7ec1d513f71ca33a  /Users/goos/.codex/config.toml
```

라운드 1 후·라운드 2 후 각각 재측정, 3개 동일. 19개 셀 전부 스크래치 `CODEX_HOME`에서 실행됐고 실사용자 `~/.codex`를 읽는 셀조차 없었다(t504의 R 셀 같은 실환경 셀을 이번엔 두지 않았다).

계측기 단계(selftest + E2E 대역 시연, 총 7회 실행)의 before/after 역시 매 실행 동일하다:

```
c91a6b73e78d057f58ab460570e6c38481632fe3e6749e0b4704fba346469598  /Users/goos/.codex/config.toml
```

**단, 이 값은 라운드 1·2의 `9f6e3a95…`와 다르다 — 그리고 그 변화를 나는 귀속할 수 없다.** 파일 mtime은 `2026-09-07 17:54:55`로, 내 라운드 2 사후 해시 채취 직후이며 selftest 시작 전이다. 내가 보증하는 불변은 **각 실행 내부의 before == after**이고, 그것은 3개 하네스 실행과 7회 계측기 실행 전부에서 성립했다. 내가 보증하지 **않는** 것은 실행 **사이**의 불변이다 — 그 사이 다른 행위자(코덱스 데스크톱 앱이나 병행 세션)가 실사용자 config를 썼고, 나는 그것이 내가 아니었다는 증거를 갖고 있지 않다. 두 값을 가로질러 비교하는 독자가 이 차이를 본 작업 탓으로 읽지 않도록 적어 둔다.

**[리드 후속 측정 2026-09-07]** 원인은 여전히 귀속하지 못했다. 다만 이 카드가 걱정하는 축은 **바뀌지 않았음을 실측했다** — 귀속 불가와 위험은 다른 말이고, 여기서는 전자다.

```
$ shasum -a 256 ~/.codex/config.toml | cut -d' ' -f1
c91a6b73e78d057f58ab460570e6c38481632fe3e6749e0b4704fba346469598   ← 위 계측기 값과 동일

$ grep -c '^\[\[skills.config\]\]' ~/.codex/config.toml
49

$ grep -A3 '^\[\[skills.config\]\]' ~/.codex/config.toml | grep -c 'enabled'
49

$ ls -l ~/.codex/config.toml
-rw-------@ 1 goos  staff  17354 Sep  7 17:54 /Users/goos/.codex/config.toml
```

세 가지가 따라 나온다. **(1)** 엔트리 수 49는 t504가 실환경에서 잰 값과 같다 — 이 변화가 유령 엔트리를 새로 만들지 않았다. **(2)** 49/49가 `enabled`를 보유한다 — t504 F1의 전면 장애 위험(키 없는 엔트리 하나면 codex 시작 실패)은 이 파일에 없다. **(3)** 파일 모드가 `0600`이라, 이 사용자에 한해 prune 방식(`os.WriteFile(…, 0o600)`)으로 써도 퍼미션이 바뀌지 않는다 — 감사 D8이 가리키는 위험(0644 사용자가 조용히 0600이 되는 것)이 사라지는 것은 아니고, **이 머신에서 지금 터지는 문제는 아니라는** 사실만 붙는다. 원 모드 보존 결정은 그대로 둔다(비용이 0이고, 이 기능이 내세우는 성질이 「비파괴」이므로).

이 측정도 mtime `17:54:55` 이후 상태에 대한 것이며, 그 이전 상태와의 차분은 여전히 관측되지 않았다.

---

## downstream 함의 (t502 설계 입력)

1. **발행 표기 확정**: `<projectRoot>/.agents/skills/<skill>/SKILL.md` — 절대경로, 파일 모양. 심링크/복사 두 배포 모양 모두에서 묶이는 유일한 표기다.
2. **금지 표기 3종**: `.claude/skills/…` 해소 경로(복사 모양에서 무효), 디렉터리 모양(두 모양 모두 무효), 상대경로(무효). 셋 다 **조용히** 무효라 사용자에게 아무 신호가 없다.
3. **t504 F1 제약 유지**: `enabled` 필드는 파서가 필수로 요구한다 — 생략하면 그 사용자의 codex가 전면 하드 에러(rc=1)로 죽는다. 이번 19셀 전부 `enabled`를 명시 발행했고 스키마 에러 0건이었다.
4. **`MirrorModeSkipped` 서술 수정 필요**: SPEC의 "skipped ⇒ 이름 해소 거부" 주장은 근거가 없다. skipped 모양(실디렉터리)은 정상 로드·정상 차단된다. 다만 `SKILL.md` 부재 하위 사례는 미측정이므로, 거부 경로를 유지하려면 그 조건으로 좁혀 다시 진술해야 한다.
5. **루트 번호를 SPEC에 고정하지 말 것**: r0/r1/r2는 환경에 따라 재배치되는 상대 라벨이다(t504 실환경 관측과 본 격리 관측이 이미 다르다).
6. **격리 누수 기록**: `~/.agents/skills`는 `CODEX_HOME` 격리를 통과해 루트로 들어온다. 향후 이 계열 실험은 스킬 이름 충돌을 반드시 고유화해야 한다.

---

## Claim

codex-cli 0.153.4에서 `[[skills.config]]`의 `enabled = false` 게이트는 **엔트리 경로와 로드된 스킬 파일 경로를 각각 realpath 정규화해 비교하는 파일 단위 게이트**다. moai의 `.agents/skills` 미러를 통해 노출된 스킬에 대해, **리터럴 미러 경로 `<projectRoot>/.agents/skills/<skill>/SKILL.md`(절대·파일 모양)만이 심링크 미러와 복사 폴백 미러 양쪽에서 게이트를 묶는다.** 심링크 해소 경로(`.claude/skills/…`)는 심링크 모양에서만 묶이고 복사 모양에서는 조용히 무효이며, 디렉터리 모양과 상대경로 표기는 두 모양 모두에서 무효다. `MirrorModeSkipped`가 만드는 파일시스템 모양은 복사 모양과 구별되지 않으므로 이름 해소는 성공한다.

## Evidence

- 라운드 1·2 셀 표 전체(위) — 각 행의 `out` 바이트, `rc`, marker 카운트는 디스크 파일 대상 `grep -c` 결과이며 파이프 판정이 아니다. 원본 요약: `lab/summary.txt`, `lab/summary2.txt`
- 루트 테이블·로드 경로 verbatim(위 인용) — 출처 `lab/cells/S0.out`, `lab/cells/C0.out`
- 셀별 발행 config 전문: `lab/configs/home-*.toml` (19개)
- 셀별 stdout/stderr/rc 원본: `lab/cells/*.{out,err,rc}` (19셀 × 3파일)
- 런타임 스탬프: `lab/version.txt` → `codex-cli 0.153.4`
- 노터치 해시 3본: `lab/hash-before.txt`, `lab/hash-after.txt`, `lab/hash-after2.txt`
- 재현: `bash .moai/reports/t502/harness.sh` → 출력의 `EXP_ROOT`를 받아 `bash .moai/reports/t502/harness2.sh <EXP_ROOT>`
- 계측기 자가검증 원본: `bash .moai/reports/t502/probe.sh selftest` (위 인용은 그 실행의 stdout), 셀별 캡처 `lab/instrument/{symlink,copy}-{ctl,lit}.txt`·`copy-res.txt`
- E2E 대역 시연: `lab/instrument/e2e-demo-verb-config.toml`(대역이 쓴 config, 실행 후 재확인본), `lab/instrument/e2e-demo-post-verb.out`(차단 상태 stdout 24754B)

## Baseline-attribution

- **측정 대상 바이너리**: `codex-cli 0.153.4` — 이 실행에서 `codex --version`을 직접 돌려 `lab/version.txt`에 기록. t504 보고서의 값을 반입한 것이 아니다(우연히 같은 버전이다).
- **측정 트리**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t502` @ 브랜치 `WT-codex-skillconfig`, HEAD `a70967b9d` (측정 직후 `git rev-parse --short HEAD` / `git branch --show-current`로 직접 재측정한 값. 이 세션의 시작 스냅샷은 `WT-codex-model-config` / `784912910`을 보고했으나 그 사이 같은 트리의 다른 행위자가 브랜치명과 HEAD를 옮겼다 — 시작값이 아니라 재측정값을 귀속한다. 본 작업은 측정 전용이라 커밋·스테이징 0건이다.)
  - **[리드 정정 2026-09-07]** 위 「다른 행위자」의 정체는 **이 카드의 리드 세션 자신**이다. 카드 착수 시점에 `EnterWorktree(t502)` 직후 `git branch -m WT-codex-skillconfig` 로 브랜치를 개명했고(`kanban-dispatch.md` § Isolation — 카드 워크트리 브랜치는 `WT-` 접두 + 서술 슬러그), 이어 `git merge develop` 으로 HEAD 가 `a70967b9d` 계보로 이동했다. 측정 에이전트가 읽은 시작 스냅샷은 그 개명보다 앞선 값이라 낡은 것이다 — 외부 침입도, 미귀속 동시 쓰기도 아니다. 원 서술을 지우지 않고 정정을 병기하는 이유는, 관측 자체는 정확했고 귀속만 불가능했기 때문이다(관측자에게는 자기 세션 밖의 개명을 알 방법이 없다).
- **미러 모양의 출처**: `internal/template/skill_mirror.go` — `mirrorLinkTarget` (`../../.claude/skills/<n>` 상대 링크), `mirrorOneSkill`의 symlink→copy 폴백 및 `MirrorModeSkipped` 분기. 픽스처는 이 코드가 만드는 모양을 손으로 재현한 것이며, `moai update`를 실제로 돌려 얻은 트리가 아니다.
- **판정 명령**: 각 셀 `codex debug prompt-input "hi"` (모델 호출 0회), 격리 `CODEX_HOME`, cwd = 셀별 프로젝트 픽스처
- **원 실험 트리**: `/tmp/t502-HTmXMiC0` (라운드 1·2 공유)

## Gaps

- **`moai update` 실행 미관측**: 픽스처는 `skill_mirror.go`를 읽고 손으로 재현했다. 실제 배포가 만드는 트리(다수 스킬, manifest, 기타 동거 파일)에서 같은 결과가 나오는지는 이번에 측정하지 않았다.
- **`MirrorModeSkipped`의 `SKILL.md` 부재 하위 사례 미측정**: 사용자가 만든 실엔트리가 빈 디렉터리이거나 동명의 일반 파일인 경우는 재현하지 않았다. 이 경우에만 해소 실패가 성립할 수 있다.
- **`MirrorModeFailed` 미측정**: 심링크·복사 둘 다 실패해 미러가 아예 없는 경우, 스킬은 `.claude/skills`로만 도달 가능하고 codex 루트에 없다 — 게이트할 대상 자체가 없다. 논리적으로 자명하지만 셀로 확인하지는 않았다.
- **심링크 미러 + 상대경로 표기 미측정**: E6는 복사 미러에서만 돌렸다. 상대 표기가 두 모양 모두에서 무효라고 단정하지 않는다(복사 모양에서 무효인 것만 관측).
- **단일 버전**: 0.153.4 한 버전. realpath 정규화는 구현 세부이며 버전 간 안정성은 미보증.
- **데스크톱 앱 미관측**: CLI(`debug prompt-input`) 경로만 측정. Codex 데스크톱 앱이 이 키를 다르게 해석할 가능성은 배제하지 못했다(t504와 동일한 유보).
- **`~/.agents/skills` 실내용 미격리**: r0가 격리를 뚫고 들어왔고 그 안에 항목 19개가 있다(`ls ~/.agents/skills | wc -l` → 19). 프로브 이름이 고유해 판정에 개입하지 않았지만, 그 루트를 비운 상태로는 돌리지 않았다.
- **다중 엔트리 상호작용 미측정**: 같은 파일을 가리키는 엔트리 2개(하나 true, 하나 false) 같은 충돌 케이스는 시도하지 않았다.
- **진짜 verb로는 E2E 축을 못 돌렸다**: verb가 아직 없어 `printf` 대역으로 통과시켰다. 대역이 보인 것은 「계측기가 외부가 쓴 config를 손대지 않고 읽어 marker 1 → 0을 잰다」까지이고, **실제 verb가 쓸 config가 유효한지는 측정하지 않았다** — 그건 구현 단계가 같은 셀로 닫아야 할 몫이다.
- **실사용자 config가 실행 사이에 바뀌었고 원인을 귀속하지 못했다**(위 노터치 절). 각 실행 내부의 불변만 보증한다.
- **계측기의 Windows·비-macOS 경로 미검증**: `probe.sh`는 `ln -s`·`shasum`에 의존한다. 심링크가 불가능한 환경(복사 폴백이 실제로 발생하는 그 환경)에서 `fixture --mirror symlink`가 어떻게 실패하는지는 시험하지 않았다.

## Residual-risk

- **정규화 동작은 구현 세부라 버전 간 뒤집힐 수 있다.** 본 판정이 "리터럴 미러 경로"를 안전한 표기로 지목하는 근거는 realpath 비교이며, codex가 문자열 비교로 회귀하면 심링크 미러에서 리터럴은 계속 맞지만 다른 축이 깨질 수 있다. t502가 발행 코드를 넣을 때 그 시점 버전에서 최소 2셀(심링크 차단 / 복사 차단)을 재측정해야 한다.
- **`.agents/skills` 루트가 cwd에 묶여 있다.** 사용자가 서브디렉터리에서 codex를 띄우면 r2 자체가 사라지고 스킬도 게이트도 함께 무효가 된다 — 발행한 엔트리는 그때 조용히 아무것도 하지 않는다. 이 조건은 본 실험 범위 밖이며 t502 설계가 별도로 다뤄야 한다.
- **절대경로 발행의 이식성**: 리터럴 표기는 프로젝트 절대경로를 config.toml에 굳힌다. 프로젝트를 옮기면 엔트리가 죽고, 죽은 엔트리에 대해 codex는 완전히 침묵한다(t504 V6/RQ3). t506이 정리 대상으로 삼는 "유령 엔트리"를 t502가 새로 생산하는 구조라는 점은 설계 판단이 필요하다.
- **계측기를 물려받는 쪽이 감수하는 위험**: `probe.sh`의 판정은 이 매트릭스가 채택 가능하다는 전제 위에 서 있다(양성 통제 4셀·무엔트리 대조군 4셀 전부 노출). selftest는 매 실행 그 전제를 다시 세우도록 짜여 있으므로 — 통제군이 노출을 못 내면 즉시 FAIL — 나중 단계는 **E2E 셀을 돌리기 전에 `selftest`를 먼저 통과시켜야** 한다. 공허한 매트릭스 위에 세운 계측기는 계측기가 없느니만 못하다.
- **양성 통제가 보증하는 범위**: 통제군은 "이 랩에서 미러가 노출된다"까지만 보증한다. 실제 사용자 환경에서 다른 루트(플러그인 루트 등)가 같은 이름의 스킬을 겹쳐 올리면 게이트 하나로 노출이 사라진다고 단정할 수 없다.
