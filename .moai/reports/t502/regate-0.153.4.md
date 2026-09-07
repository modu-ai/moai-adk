# t502 M1 재측정 게이트 — 발행 코드 착지 직전 2셀 재현

- **카드**: t502 (SPEC-CODEX-SKILL-DISABLE-001, AC-CSD-050 [HARD])
- **측정 일시**: 2026-09-07 (run-phase M1, 발행 코드 착지 **이전**)
- **런타임 스탬프**: `codex-cli 0.153.4` — 이 실행에서 직접 측정
- **바이너리**: `/Users/goos/.local/bin/codex` (`which codex`)
- **트리**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t502` @ `WT-codex-skillconfig`, HEAD `f6ed23a9c`
- **기구**: `.moai/reports/t502/probe.sh selftest` (격리 `CODEX_HOME`, 픽스처 신규 생성)

## Claim

`gate-path-shape.md`(19셀, 같은 버전)의 판정은 **이 시점에도 재현된다**. 리터럴 미러 경로 표기는 심링크 미러와 복사 폴백 미러 **양쪽에서** 스킬을 차단하고, 복사 미러의 해소(`.claude/…`) 표기는 여전히 **무효**다. 발행 표기를 바꿀 이유가 없다 — 계획대로 `<projectRoot>/.agents/skills/<skill>/SKILL.md` 를 발행한다.

## Evidence

```
$ which codex
/Users/goos/.local/bin/codex
$ codex --version
codex-cli 0.153.4

$ bash .moai/reports/t502/probe.sh selftest
symlink	control=exposed	literal_false=gated
copy	control=exposed	literal_false=gated
copy	resolved_false=exposed	(expected exposed — the inert notation)
notouch_before=c91a6b73e78d057f58ab460570e6c38481632fe3e6749e0b4704fba346469598  /Users/goos/.codex/config.toml
notouch_after=c91a6b73e78d057f58ab460570e6c38481632fe3e6749e0b4704fba346469598  /Users/goos/.codex/config.toml
ROOT=/tmp/t502-selftest-fUxG1hw0/lab
SELFTEST PASS
```

요구된 2셀: **심링크 미러 차단**(`symlink literal_false=gated`) · **복사 미러 차단**(`copy literal_false=gated`). 각각 양성 통제 동반(`control=exposed` 2건). 세 번째 줄은 계측기가 「무엇이든 gated로 부르는 도구」가 아님을 세우는 셀이다.

### [HARD] 음성 대조군 — 카운터가 망가진 것과 차단된 것을 가른다

`marker=0` 이 아무리 많아도, 카운터가 **실재 토큰에 대해 0이 아닌 값을 낼 수 있음**을 보이기 전까지 「차단됐다」와 「카운터가 죽었다」는 똑같이 생겼다. selftest 가 남긴 5개 캡처를 디스크 파일 대상으로 직접 세었다:

```
$ R=/tmp/t502-selftest-fUxG1hw0/lab
$ grep -c T502PROBE_MARKER <capture> / grep -c ZZZNOTAMARKER <capture>
symlink/ctl real=1 neg=0 bytes=24999
symlink/lit real=0 neg=0 bytes=24816
copy/ctl    real=1 neg=0 bytes=24988
copy/lit    real=0 neg=0 bytes=24807
copy/res    real=1 neg=0 bytes=24988
```

- 실재하지 않는 토큰 `ZZZNOTAMARKER` → 5셀 전부 **0** (카운터가 아무것이나 세지 않는다)
- 통제 셀의 실제 마커 → **1** (카운터가 0이 아닌 값을 낼 수 있다)
- 따라서 `lit` 두 셀의 **0** 은 「카운터 고장」이 아니라 **차단**이다

out 바이트도 통제 대비 프로브 엔트리 1줄분만큼 작다(24999 → 24816, 24988 → 24807) — 정상 실행 후 스킬 목록에서만 사라진 모양이다.

## Baseline-attribution

- **버전**: 이 실행의 `codex --version` 출력. 세션 반입값 아님. `gate-path-shape.md` 의 `0.153.4` 와 **같은 값이 재측정으로 확인됐다**(반입이 아니라 일치).
- **픽스처**: `probe.sh fixture` 가 `internal/template/skill_mirror.go` 모양으로 새로 만든 트리 (`/tmp/t502-selftest-fUxG1hw0/lab`). 원 매트릭스의 트리 재사용 아님.
- **판정 명령**: 셀마다 `codex debug prompt-input "hi"`, 격리 `CODEX_HOME`, cwd = 셀별 픽스처.
- **노터치**: 실사용자 `~/.codex/config.toml` sha256 before == after (`c91a6b73…`).

## Gaps

- 버전 드리프트를 **관측하지 못했다** — 원 측정과 같은 0.153.4 다. 다른 버전에서의 안정성은 여전히 미보증이며, 이 게이트는 「착지 시점 버전에서 재현되는가」만 답한다.
- 원 19셀 중 2셀 + 통제만 재현했다. E1/E2(정규화 방식 판별), E4/E5/E6(무효 표기)는 이번에 다시 돌리지 않았다 — AC-CSD-050 이 요구하는 범위가 2셀 + 양성 통제다.
- `moai update` 가 만드는 실제 미러 트리는 이번에도 미관측(원 측정과 같은 Gap).

## Residual-risk

- realpath 정규화는 구현 세부다. 다음 codex 릴리스에서 문자열 비교로 회귀하면 이 게이트는 다시 돌려야 한다 — 발행 표기 자체가 그 가정 위에 서 있다.
- 계측기의 판정은 원 매트릭스가 채택 가능하다는 전제 위에 있다. selftest 가 매 실행 그 전제를 다시 세우므로(통제군이 노출을 못 내면 즉시 FAIL) 이 실행에 한해 전제는 성립한다.
