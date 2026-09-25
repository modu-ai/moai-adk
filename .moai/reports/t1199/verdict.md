# t1199 — deny 규칙의 `*`·`:*` 혼용 경고 제거 (차단 범위 보존)

## Claim

1. 로컬 `.claude/settings.json`(원본 `.moai/config/sections/tool-policy.yaml`)의 `rm -rf` 규칙 3건이 `*`와 `:*`를 섞어 Claude Code 2.1.283이 기동마다 경고 3건을 냈다. 수리 후 경고는 0건이다.
2. 템플릿 쪽은 t1185(`302cd9e04`)가 이미 경고를 없앴지만, 그 수리는 **차단 범위를 넓혔다**. 구 규칙은 글자 그대로의 `*`만 맞췄는데 t1185 형태(`rm -rf /*`)는 `rm -rf /<아무 경로>`·`rm -rf ~/<아무 경로>`를 모두 거부한다. 이 카드는 템플릿도 구 범위로 되돌렸다.
3. 새 형태 `Bash(rm -rf /\* *)`(JSON 원문 `"Bash(rm -rf /\\* *)"`)는 경고 0건이면서 구 규칙과 같은 거부 집합을 낸다. `~/`·`C\:/`도 같은 모양이다.

## Evidence

측정 조건: Claude Code 2.1.283, `claude -p --setting-sources user --settings <fixture> --model haiku`. 범위 측정은 `--permission-mode bypassPermissions`(deny 규칙은 bypass에서도 적용)에서 `permission_denials`와 픽스처 사후 상태를 읽었다. 진짜 `rm -rf /*`는 시험하지 않았고, 같은 모양의 `echo zz …` 대리 규칙으로 쟀다.

**경고 재현 (양성 대조)** — `origin/develop:.claude/settings.json` 으로 기동:

```
Bash(rm -rf /*:*) mixes
Bash(rm -rf C\\:/*:*) mixes
Bash(rm -rf ~/*:*) mixes
```

CC 경고 원문 발췌(구 규칙 픽스처): `mixes * with the trailing :* prefix syntax, so it is matched as a literal prefix (the * is not expanded) and matches only commands containing a literal * at that position. Use Bash(rm -rf /*) for wildcard matching.` — 권고 형태는 와일드카드이므로 의미가 다르다.

**수리 후** — 이 트리의 `.claude/settings.json` 으로 기동: `grep -c mixes` → `0`.

**거부 집합 비교** (9개 대리 명령, 각 픽스처 1회):

| 규칙 형태 | 거부된 명령 |
|---|---|
| 구 `/*:*` | `echo zz /`, `/*`, `/* x`, `~/*`, `~` |
| 새 `/\* *` | `echo zz /`, `/*`, `/* x`, `~/*`, `~` |
| `/\*:*` (탈락 후보) | `echo zz /`, `~` — 리터럴 `/*` 를 못 막음 |

두 형태 모두 `/*x`, `/etc`, `~/foo` 는 허용했다.

**범위 확대 실증 (t1185 형태)** — 실제 `rm` 픽스처:

- t1185 형태 `rm -rf /*`, `rm -rf ~/*`: `rm -rf <scratchpad>/fx1`, `rm -rf ~/.t1199probe` 둘 다 **거부**, 픽스처 잔존.
- 구 형태와 이 트리의 수리본: 둘 다 **실행**(permission_denials `[]`), 픽스처 삭제됨.

**테스트·빌드**

- `TestSettingsTemplateDenyWildcardSyntax` 기대값을 새 형태로 바꾸고, t1185 형태가 다시 들어오면 실패하는 역방향 단언을 추가했다. 수리 전 실행 → darwin/linux/windows 모두 FAIL(누락 3 + 확대 3). 수리 후 `go test ./internal/template -count=1` → `ok … 64.231s`.
- `moai tool-policy build --local-only` → `regenerated 1 target(s)`, `deny=55`. `make build` → exit 0(`tool-policy-drift-check` 통과). `go vet ./internal/template/`·`gofmt -l` → 출력 없음.

## Baseline-attribution

작업트리 `.claude/worktrees/t1199`, 브랜치 `WT-deny-wildcard-syntax`, 기준 `origin/develop` = `85d99c6b4`. 위 수치는 모두 이 트리에서 이번 실행으로 관측했다. 픽스처와 원출력은 세션 스크래치패드에 있었다(휘발).

## Gaps

- `C\:/` 규칙은 echo 대리로 재지 않았다(같은 모양이라는 추론). Windows 실기에서 측정하지 않았다.
- `\*` 이스케이프는 공식 문서에 규정이 없다. 2.1.283의 관측 동작에 기대므로, 이후 버전에서 의미가 바뀔 수 있다.
- 각 픽스처는 1회씩만 돌렸다. 거부 판정은 결정적 규칙 매칭이라 반복의 의미는 작지만 반복하지는 않았다.
- CodeRabbit·CI는 보지 않았다(리드 몫).

## Residual-risk

- 권한 규칙은 명령 문자열 패턴이다. `rm -rf "/"`·`rm -rf //` 같은 다른 철자는 구 규칙이든 새 규칙이든 막지 못한다. 보안 경계가 아니다.
- CC가 향후 `\*` 를 와일드카드로 해석하거나 경고 대상으로 바꾸면 범위가 다시 달라진다. 테스트는 문자열만 보므로 그 변화를 잡지 못한다.
- t1185(`302cd9e04`)는 `origin/main` 에 없다(`git merge-base --is-ancestor 302cd9e04 origin/main` → 거짓). 다만 develop 에서 빌드한 rc 설치본은 넓은 범위로 동작하고 있을 수 있다.
