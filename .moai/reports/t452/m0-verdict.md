# t452 M0 — 적재 뿌리 측정 결과

카드: t452 · SPEC-CODEX-SKILL-LOADER-001 · 브랜치 `WT-codex-skill-wiring`
BASELINE_SHA: `c529b2e4aaf5148aee7e6c67649bf392837bbb06` (트리 SHA — 여섯 회차 전부 이 값에서 실행)
codex 버전: `codex-cli 0.152.1` — **회차마다 그 회차 안에서 관측**했고, 여섯 값이 모두 같다
(`m0-runs/{A..F}-version.txt`).

## 판정 — 분기 A

`<repo>/.agents/skills/<name>/SKILL.md` 가 **적재된다.** plan.md M0 분기 기록 1 번에 따라 **분기 A**.
물려받은 전제(M1 SPEC §A)는 codex-cli 0.152.1 에서 **참**이다.

근거가 된 출력 줄 — 회차 B(`m0-runs/B-last.txt`), R1 만 채운 회차:

```
skill-installer
moai-probe-agents
agent-browser
```

같은 셀렉터가 회차 A(무표식 음성 대조)에서 0 건이므로, 이 줄은 R1 의 점유에 귀속된다.

## 회차별 표 (AC-CSL-001 의 다섯 요소)

명령은 여섯 회차 전부 동일하며(프롬프트 문자열 포함), 표식 이름만 회차마다 다르다:

```
$ cd /tmp/t452-m0/proj
$ CODEX_HOME=/tmp/t452-m0/codexhome codex --version
$ CODEX_HOME=/tmp/t452-m0/codexhome timeout 240 codex exec \
    --cd /tmp/t452-m0/proj --skip-git-repo-check -s read-only --json \
    -o <회차>-last.txt \
    "List the names of every skill available to you in this session, one per line, exactly as registered. If none are available, reply with exactly: NONE" \
    > <회차>.jsonl 2> <회차>.err
$ echo "exit=$?"
```

| 회차 | 채운 뿌리 | 표식 이름 | exit | 버전 | 표식 발화 | 판정 |
|---|---|---|---|---|---|---|
| A | 없음 (음성 대조) | — | 0 | codex-cli 0.152.1 | 0 건 | 신호 유효 |
| B | R1 `<repo>/.agents/skills` | `moai-probe-agents` | 0 | codex-cli 0.152.1 | 1 건 | **적재됨** |
| C | R2 `$CODEX_HOME/skills` | `moai-probe-codexhome` | 0 | codex-cli 0.152.1 | 1 건 | **적재됨** |
| D | R3 `<repo>/.codex/skills` | `moai-probe-dotcodex` | 0 | codex-cli 0.152.1 | 1 건 | **적재됨** |
| E | R4 `<repo>/skills` | `moai-probe-cwd` | 0 | codex-cli 0.152.1 | 0 건 | 적재되지 않음 |
| F | R5 `<repo>/.claude/skills` | `moai-probe-claude` | 0 | codex-cli 0.152.1 | 0 건 | 적재되지 않음 |

출력 전문은 `m0-runs/<회차>-last.txt`(최종 메시지) · `<회차>.jsonl`(이벤트 스트림) ·
`<회차>.err`(stderr) 에 있다. 요약이 아니라 원출력이다.

## 대조 두 방향 (plan.md M0 규칙 4)

- **음성 대조** — 회차 A 는 다섯 뿌리를 전부 비운 채 실행했고 `moai-probe` 매치가 0 건이다.
  프롬프트 문자열에 `moai-probe` 토큰이 없으므로(신호 선택 문서의 [HARD]), 다른 회차에 나타난
  표식 이름은 전부 코덱스가 공급한 것이다.
- **양성 대조** — 세 뿌리(R1 · R2 · R3)에서 발화했다. 계측기가 초록을 낼 수 있다는 것이
  직접 관측되었으므로, E · F 의 침묵을 "계측기가 죽어 있다"로 읽을 수 없다.
- `inconclusive` 조건(어느 뿌리에서도 미발화)은 **성립하지 않는다.** 신호 재선택은 0 회 —
  1 순위 S1 이 첫 회차에서 작동했다.

### 계측기가 디스크를 읽는다는 독립 증거

회차 A 실행 뒤 격리 `CODEX_HOME` 에 코덱스 자신이 `skills/.system/` 을 만들고 시스템 스킬
6 개(`imagegen` · `openai-docs` · `plugin-creator` · `review-agent` · `skill-creator` ·
`skill-installer`)를 설치했으며, 같은 회차의 나열에 그 이름들이 나타났다. 모델의 자기보고가
학습 기억이 아니라 그 실행의 디스크 상태를 반영한다는 뜻이다.

## 재현

여섯 회차를 **두 번** 독립적으로 돌렸고(1 차 `run0`~`run5`, 2 차 `A`~`F`) 뿌리별 발화 패턴이
동일했다. 2 차만 회차별 버전을 함께 떴으므로 AC-CSL-001 의 근거는 2 차다.

## AC-CSL-010 — 사용자 계층 불변

프로브가 쓴 `CODEX_HOME` 은 `/tmp/t452-m0/codexhome` 이며, 위 명령 문면에서 확인된다.
측정 대상인 개발자 계층은 `/Users/goos/.codex` 로 서로 다른 경로다.

| | config.toml sha256 | skills/ 목록 |
|---|---|---|
| pre | `8adec56f13e1cb6baafdb89a406f907f98b002d9136ca1cbf6f40a23194b4ae7` | `.system`, `hatch-pet` |
| post | `8adec56f13e1cb6baafdb89a406f907f98b002d9136ca1cbf6f40a23194b4ae7` | `.system`, `hatch-pet` |

두 측정이 동일하다 → **PASS**. 양쪽 모두 `find`(alias 가 아닌)로 떴다.

## Gaps — 관측하지 **않은** 것

- **`--json` 이벤트 스트림은 런타임의 스킬 주입 목록을 노출하지 않는다.** 신호 선택 문서가
  설계한 "두 겹 측정"은 실제로는 한 겹(모델 최종 메시지)으로 축소됐다. 이벤트 스트림에는
  `thread.started` · `turn.started` · `item.completed`(=최종 메시지) · `turn.completed` 넷만 담긴다.
  따라서 모델이 실제 등록 목록의 일부를 **누락**했을 가능성은 이 계측으로 배제되지 않는다 —
  E · F 의 0 건이 "미적재"가 아니라 "적재됐으나 모델이 안 적었다"일 수 있다는 뜻이다.
  다만 B · C · D 의 양성은 그 가능성에 영향받지 않는다(누락은 거짓 음성만 만든다).
- **규칙 2 의 문면 이탈 1 건**: R2 뿌리에는 여섯 회차 내내 코덱스가 스스로 만든
  `skills/.system/` 이 함께 있었다. 지우면 코덱스가 다음 실행에서 다시 만든다. 그 이름들이
  `moai-probe-*` 와 서로소이므로 귀속은 깨지지 않지만, "정확히 한 뿌리만 채운다"의 문면과는
  다르다 — 조용히 넘기지 않고 여기 적는다.
- 격리 `CODEX_HOME` 에 실 계층의 `auth.json` 을 복사했다. `config.toml` · `skills/` 는 복사하지
  않았으므로 적재 뿌리 측정의 격리는 유지된다.
- 스킬 **본문**(sentinel) 인용은 시험하지 않았다. S1 이 첫 회차에 작동해 S2 로 갈 이유가 없었다.

## Residual risk

- 적재는 **이 codex 버전·이 프롬프트·이 나열 방식**에서 관측됐다. 다른 프롬프트에서 모델이
  다르게 나열할 여지는 남는다. 세 양성이 서로 다른 회차에서 재현됐다는 점이 이 위험을 줄인다.
- R1 이 적재된다는 사실이 **`.agents/skills` 가 유일한 뿌리**라는 뜻은 아니다 — R2 · R3 도 적재된다.
  이 SPEC 이 묻는 것은 R1 의 적재 여부이므로 분기 판정에는 영향이 없으나, 후속 설계가
  "R1 만이 뿌리"로 읽으면 거짓이다.
