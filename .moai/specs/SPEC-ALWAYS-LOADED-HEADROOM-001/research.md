# SPEC-ALWAYS-LOADED-HEADROOM-001 — 조사 기록 (plan 단계)

기준 커밋 `7fe658815eb0d4110b9acadad56e5a85bee3ed3f`, 2026-09-27, manager-spec 이 이 워크트리에서 실행.

## §1. 템플릿 원본의 18경로 대응 합계

명령(워크트리 루트 기준 `internal/template/templates/` 안에서):

```bash
wc -m CLAUDE.md AGENTS.md.tmpl .moai/config/sections/user.yaml.tmpl .moai/config/sections/language.yaml.tmpl \
  .claude/rules/moai/core/{agent-common-protocol,askuser-protocol,moai-constitution,moai-mcp-tools,native-idiom-and-register,verification-claim-integrity}.md \
  .claude/rules/moai/workflow/{cache-aware-execution,context-window-management,cross-session-messaging,goal-directive,kanban-dispatch,main-checkout-branch-guard,session-handoff,skill-routing}.md | tail -1
```

관측: `203611 total`. 같은 실행에서 라이브 18경로는 `199111 total`.

해석의 한계: 이 값은 **미렌더링** 원본이다. `moai init` 이 `.tmpl` 을 렌더링하면 자수가 달라지므로 사용자 표면 `S_init` 의 값으로 쓸 수 없다. 오케스트레이터가 보고한 182,830 은 `.tmpl` 3개를 뺀 15개 파일의 합이다(203,611 − 182,830 = 20,781 이 `.tmpl` 3개 몫).

관측 사실 하나: 템플릿 `CLAUDE.md` 는 9행에 `@AGENTS.md`, 107·108행에 두 yaml 을 import 한다(`grep -n '^@'`). 즉 배포 표면도 18경로 구조를 갖는다.

## §2. 동결 다중집합 재확인

AC-ALD2-002 파이프라인 전문을 이 트리에서 재실행: `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3`. t1175 기준선과 같다.

## §3. 절 분할 스크립트의 소재

`design.md §4.0` 이 인용한 `/tmp/claude-501/sec.py` 는 plan 단계 확인 시점에 존재했다(`ls`: 1,661 bytes, 2026-09-25 13:29). 대상 16파일 목록, 판정식 `\[HARD\]|MUST|shall `, fence 추적, 최대 제목 수준 인자를 담는다.

v0.2.0(plan-audit D13): plan 단계에서 사본을 `.moai/reports/t1226/sec.py` 로 커밋했다. 해시 대조 출력:

```
$ shasum -a 256 /tmp/claude-501/sec.py .moai/reports/t1226/sec.py
d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547  /tmp/claude-501/sec.py
d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547  .moai/reports/t1226/sec.py
```

감사자 실측값과도 같다. 한계(plan-audit D19): 스크립트는 16개 마크다운만 다루고 첫 제목 앞의 서문을 절로 세지 않으며 yaml 두 개를 보지 않는다. 후보 열거는 이를 보완한 `candidates.py`(run 단계 M3)가 맡고, `sec.py` 는 `P절` 재조정 재현에만 쓴다.

## §5. `moai init` 의 홈 쓰기와 격리 경로 (plan-audit D2)

- `internal/cli/init_home_guard_test.go:6-13` — runInit 이 닿는 홈 경로: user-scope settings splice(`userHomeDirFn`), profile ledger(`profile.BaseDirOverride`), `MOAI_HOME`, 셸 rc 작성기(HOME 을 직접 해석). `prepareSafeInitHome` 이 앞 셋을 `t.TempDir()` 로 돌리고 셸 rc seam 을 스파이로 바꾼다.
- `internal/cli/init.go:923` — `userHomeDirFn()` 아래 `~/.claude/settings.json` 에 user-scope 쓰기.
- `internal/core/project/initializer.go:673` — `ConfigureShellEnvFn = defaultConfigureShellEnv`(실제 rc 파일 작성).
- `internal/cli/init_deploy_exit_test.go:50` — `runInitWithFlags(t, root, extra)` 가 `initCmd.RunE` 를 비대화형으로 구동하는 기존 헬퍼.
- `moai init --help`(이 트리에서 빌드한 바이너리): 전역 쓰기를 건너뛰는 플래그 없음. `--root`, `--name`, `--llm`, `--all`(기본은 slim) 존재.

## §4. 카드 본문과 선행 SPEC 의 조건 명칭 차이

카드 본문은 조건 1~4 를 「바인딩 조항 불변, 파일당 한도, 미러 동등성, 런타임 로드 규칙」으로 요약했다. 선행 SPEC 의 정의(`plan.md §C` M1)는 「구속 조항 줄 0, 역방향 인용, 목적지 `paths:` 도달, 목적지 40,000자 수용량」이다. 이 SPEC 은 선행 정의를 정본으로 쓰고, 카드의 「미러 동등성」은 REQ-ALD2-007 의 미러 의무로, 「런타임 로드 규칙」은 조건 3 으로 흡수한다 — 두 표면을 모두 재는 것(REQ-ALH-002)과 조건 4 를 두 트리에서 판정하는 것(REQ-ALH-015)이 그 흡수의 결과다.
